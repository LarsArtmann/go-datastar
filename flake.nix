{
  description = "DataStar protocol library for Go (patches as values on top of go-sse)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      systems,
      treefmt-nix,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          # TODO(go-1.26.6): drop this override when nixpkgs ships go_1_26 >= 1.26.6.
          # The go.mod directives pin 1.26.6 to clear stdlib CVEs
          # (GO-2026-5972/6089/6090/6218) and GOTOOLCHAIN=local forbids
          # auto-downloading a newer toolchain in hermetic builds.
          goPkg = pkgs.go_1_26.overrideAttrs (_old: rec {
            version = "1.26.6";
            src = pkgs.fetchurl {
              url = "https://go.dev/dl/go${version}.src.tar.gz";
              hash = "sha256-oHIcVMaIkBRI13rZs+x+p8R0cwdV/4kTgukuy5P/LLE=";
            };
          });
          buildGoModule = pkgs.buildGoModule.override { go = goPkg; };
          version = self.rev or self.dirtyRev or "dev";
          # go1.26.6's `go mod vendor` output differs from 1.26.5's (modules.txt
          # format), so this hash moved with the toolchain bump.
          vendorHash = "sha256-ExmfW1vWz+7w8j4kuBsPwvaWqwci1DBFawnfk13XasE=";

          # TODO: add hermeticCheckStatic and hermeticCheckDatastartest
          # buildGoModule derivations for full multi-module Nix CI.
          # The GitHub Actions CI already covers all three modules.
          hermeticCheck = buildGoModule {
            pname = "go-datastar";
            inherit version vendorHash;
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.gitTracked ./.;
            };
            subPackages = [ "." ];
            proxyVendor = true;
            doCheck = true;
            env.GOEXPERIMENT = "jsonv2";

            meta = {
              description = "DataStar protocol library for Go";
              license = lib.licenses.mit;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };

          mkApp =
            name: runtimeInputs: text:
            let
              script = pkgs.writeShellApplication {
                inherit name runtimeInputs text;
              };
            in
            {
              type = "app";
              program = lib.getExe script;
              meta.description = "go-datastar: ${name}";
            };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              golines.enable = true;
              nixfmt.enable = true;
            };
            settings = {
              formatter = {
                gofumpt.excludes = [ "*_templ.go" ];
                goimports.excludes = [ "*_templ.go" ];
                golines.excludes = [ "*_templ.go" ];
              };
            };
          };

          # goimports shells out to `go env` per file; the `go` on its PATH
          # must satisfy the go.mod directive or the sandbox tries a
          # toolchain download (no network). The gotools wrapper only
          # APPENDS its build-time go, so a goPkg first on PATH wins.
          checks.format = (config.treefmt.build.check self).overrideAttrs (old: {
            buildInputs = old.buildInputs ++ [ goPkg ];
          });
          checks.build = hermeticCheck;

          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.govulncheck
              pkgs.templ
            ];

            GOTOOLCHAIN = "local";
            GOEXPERIMENT = "jsonv2";

            shellHook = ''
              echo "go-datastar dev shell: $(go version)"
              echo "GOEXPERIMENT=$GOEXPERIMENT"
            '';
          };

          apps = {
            test = mkApp "test" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... ./datastartest/... ./static/... -count=1 "$@"
            '';

            test-race = mkApp "test-race" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... ./datastartest/... ./static/... -race -count=1 "$@"
            '';

            build = mkApp "build" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go build ./... ./datastartest/... ./static/...
            '';

            vet = mkApp "vet" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go vet ./... ./datastartest/... ./static/...
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              export GOEXPERIMENT=jsonv2
              golangci-lint run ./... ./datastartest/... ./static/...
            '';

            erraudit = mkApp "erraudit" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go install github.com/larsartmann/erraudit/cmd/erraudit@v0.3.0
              "$HOME/go/bin/erraudit" ./... ./datastartest/... ./static/... --type-aware --enforce-go-error-family --severity-threshold error
            '';

            govulncheck = mkApp "govulncheck" [ pkgs.govulncheck goPkg ] ''
              export GOEXPERIMENT=jsonv2
              govulncheck ./... ./datastartest/... ./static/...
            '';

            coverage = mkApp "coverage" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test ./... ./datastartest/... ./static/... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';
          };
        };
    };
}
