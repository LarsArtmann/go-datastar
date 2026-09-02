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
          # TODO(go-1.26.7): drop this override when nixpkgs ships go_1_26 >= 1.26.7.
          # The go.mod directives pin 1.26.7 to clear stdlib CVEs
          # (GO-2026-5972/6089/6090/6218) and GOTOOLCHAIN=local forbids
          # auto-downloading a newer toolchain in hermetic builds.
          goPkg = pkgs.go_1_26.overrideAttrs (_old: rec {
            version = "1.26.7";
            src = pkgs.fetchurl {
              url = "https://go.dev/dl/go${version}.src.tar.gz";
              hash = "sha256-DtJOrHVRBQhbif6cq8J0K5GgrXuUtZ0602SRjryJVq0=";
            };
          });
          buildGoModule = pkgs.buildGoModule.override { go = goPkg; };
          version = self.rev or self.dirtyRev or "dev";
          # Root vendorHash moves ONLY on requires changes (go.mod/go.sum) or
          # toolchain modules.txt format changes (e.g. go1.26.5 -> 1.26.6) —
          # verified 2026-09-02 (ADR 004 correction): root imports no
          # directory-replaced package, so repo source never enters its
          # vendor tree.
          vendorHash = "sha256-dgqHjh3F0QFtRwgFD+2ntKmdfJqs/uCd8EZhJxg+7EQ=";
          # datastartest vendors root + static through its directory replaces:
          # `go mod vendor` copies the replaced directories ENTIRELY (docs
          # included), so this hash moves on ANY edit to any tracked file
          # under the repo root or static/ (plus requires/toolchain changes)
          # — verified 2026-09-02 (ADR 004 correction, evidence matrix).
          datastartestVendorHash = "sha256-KmTASe3SCWWBFTrzeo1+O8p7ZPdIzHMLS2GRe8NQKas=";

          maintainer = {
            name = "Lars Artmann";
            github = "LarsArtmann";
          };

          # One hermetic buildGoModule derivation per Go module — the Nix
          # analog of the CI GOWORK=off per-module legs. All run in module
          # mode: GOWORK=off and go.work excluded from the source fileset.
          hermeticCheck = buildGoModule {
            pname = "go-datastar";
            inherit version vendorHash;
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.difference (lib.fileset.gitTracked ./.) ./go.work;
            };
            subPackages = [ "." ];
            proxyVendor = true;
            doCheck = true;
            env = {
              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";
            };

            meta = {
              description = "DataStar protocol library for Go";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
            };
          };

          # static module: stdlib-only, nothing to vendor.
          hermeticCheckStatic = buildGoModule {
            pname = "go-datastar-static";
            inherit version;
            src = lib.fileset.toSource {
              root = ./static;
              fileset = lib.fileset.gitTracked ./static;
            };
            vendorHash = null;
            subPackages = [ "." ];
            doCheck = true;
            env.GOWORK = "off";

            meta = {
              description = "Embedded DataStar JS client bundle";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
            };
          };

          # datastartest module: modRoot points into the repo source so the
          # sibling replaces (=> .., => ../static) resolve inside the sandbox.
          # The src fileset is deliberately MINIMAL: datastartest itself, the
          # replaced modules' package sources (root *.go + go.mod, static/).
          # Root-level metadata (flake.nix, *.md, CI config) MUST be excluded:
          # `go mod vendor` copies replaced module directories ENTIRELY, so
          # with flake.nix inside the fileset the vendorHash constant would
          # sit inside its own FOD input — an unsolvable self-reference (the
          # hash could never converge; verified 2026-09-02, ADR 004).
          datastartestSrc = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              (lib.fileset.gitTracked ./datastartest)
              (lib.fileset.gitTracked ./static)
              ./go.mod
              (lib.fileset.fileFilter (file: lib.hasSuffix ".go" file.name) ./.)
            ];
          };
          hermeticCheckDatastartest = buildGoModule {
            pname = "go-datastar-datastartest";
            inherit version;
            vendorHash = datastartestVendorHash;
            src = datastartestSrc;
            modRoot = "datastartest";
            subPackages = [ "." ];
            doCheck = true;
            env = {
              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";
            };

            meta = {
              description = "Consumer E2E test helpers for go-datastar";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
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
            # treefmt-nix's built-in flakeCheck would add an UNGUARDED
            # checks.treefmt without goPkg on PATH; checks.format below is the
            # guarded equivalent (goimports shells out to `go env` per file).
            flakeCheck = false;
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
          checks.buildStatic = hermeticCheckStatic;
          checks.buildDatastartest = hermeticCheckDatastartest;

          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.actionlint
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

            # EXACT CI parity: pins the same golangci-lint version the CI lint
            # job installs (nixpkgs' golangci-lint drifts between versions,
            # which caused "green locally, red in CI" masters). THE pre-push
            # gate; first run compiles the linter from source (module cache).
            docspec = mkApp "docspec" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go test -tags docspec -run TestDocspec -count=1 ./... ./datastartest/...
            '';

            lint-ci = mkApp "lint-ci" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./... ./datastartest/... ./static/... --timeout 5m
            '';

            # erraudit is NOT hermetically buildable (its dependency tree
            # contains private modules, e.g. go-finding), so this app
            # go-installs it and requires local GitHub credentials. One
            # directory per run — the tool rejects package patterns.
            erraudit = mkApp "erraudit" [ goPkg ] ''
              export GOEXPERIMENT=jsonv2
              go install github.com/larsartmann/erraudit/cmd/erraudit@v0.3.0
              for mod in . ./datastartest ./static; do
                echo "== erraudit $mod"
                (cd "$mod" && "$HOME/go/bin/erraudit" . --type-aware --enforce-go-error-family --severity-threshold error)
              done
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

            bench = mkApp "bench" [ goPkg ] ''
              echo "== bench: $(go version) / $(uname -s) $(uname -m) / $(nproc) cores"
              export GOEXPERIMENT=jsonv2
              go test -run '^$' -bench . -benchmem "$@"
              (cd datastartest && go test -run '^$' -bench . -benchmem "$@")
            '';
          };
        };
    };
}
