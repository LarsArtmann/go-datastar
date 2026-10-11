{
  description = "DataStar protocol library for Go (patches as values on top of go-sse)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      treefmt-nix,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Inline systems: nixpkgs 26.11 dropped x86_64-darwin, which github:nix-systems/default still lists.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

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
          goPkg = pkgs.go_1_27;
          buildGoModule = pkgs.buildGoModule.override { go = goPkg; };
          version = self.rev or self.dirtyRev or "dev";
          # Root vendorHash moves on requires changes (go.mod/go.sum), on
          # toolchain modules.txt format changes (e.g. go1.26.5 -> 1.26.6),
          # AND on any edit to the directory-replaced ./static sources:
          # root imports github.com/larsartmann/go-datastar/static
          # (script_handler.go) and `go mod vendor` copies the replaced
          # directory ENTIRELY, so repo source DOES enter the vendor tree
          # (corrected 2026-10-11; the earlier "root imports no
          # directory-replaced package" claim predated the static import).
          vendorHash = "sha256-iQzWQbeGUGgNv/lEuSVvzfTo2t2we/9Jt/XUQReeCS4=";
          # datastartest vendors root + static through its directory replaces:
          # `go mod vendor` copies the replaced directories ENTIRELY (docs
          # included), so this hash moves on ANY edit to any tracked file
          # under the repo root or static/ (plus requires/toolchain changes)
          # — verified 2026-09-02 (ADR 004 correction, evidence matrix).
          datastartestVendorHash = "sha256-dvw0guGGGL9cGSe9G2Oo7cGg/Wf4tm1JHC0/1bWsjRk=";
          # broadcast vendors root + static through its directory replaces —
          # same movement rules as datastartestVendorHash above.
          broadcastVendorHash = "sha256-5b6VtYAOZF2BunD4HBvVhPmY6VNlpmBhiTKY2hQ7drQ=";

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
            };

            meta = {
              description = "DataStar protocol library for Go";
              homepage = "https://github.com/larsartmann/go-datastar";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
              mainProgram = "go-datastar";
              platforms = lib.platforms.all;
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
              homepage = "https://github.com/larsartmann/go-datastar";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
              mainProgram = "go-datastar-static";
              platforms = lib.platforms.all;
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
            };

            meta = {
              description = "Consumer E2E test helpers for go-datastar";
              homepage = "https://github.com/larsartmann/go-datastar";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
              mainProgram = "go-datastar-datastartest";
              platforms = lib.platforms.all;
            };
          };

          # broadcast module: same minimal-fileset pattern as datastartest
          # (see the ADR 004 note above — root-level metadata stays out so
          # the vendorHash constant never lands inside its own FOD input).
          broadcastSrc = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              (lib.fileset.gitTracked ./broadcast)
              (lib.fileset.gitTracked ./static)
              ./go.mod
              (lib.fileset.fileFilter (file: lib.hasSuffix ".go" file.name) ./.)
            ];
          };
          hermeticCheckBroadcast = buildGoModule {
            pname = "go-datastar-broadcast";
            inherit version;
            vendorHash = broadcastVendorHash;
            src = broadcastSrc;
            modRoot = "broadcast";
            subPackages = [ "." ];
            doCheck = true;
            env = {
              GOWORK = "off";
            };

            meta = {
              description = "Connection lifecycle for DataStar patches: fan-out, replay, hub sharing";
              homepage = "https://github.com/LarsArtmann/go-datastar";
              license = lib.licenses.mit;
              maintainers = [ maintainer ];
              mainProgram = "go-datastar-broadcast";
              platforms = lib.platforms.all;
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

          checks = {
            # goimports shells out to `go env` per file; the `go` on its PATH
            # must satisfy the go.mod directive or the sandbox tries a
            # toolchain download (no network). The gotools wrapper only
            # APPENDS its build-time go, so a goPkg first on PATH wins.
            format = (config.treefmt.build.check self).overrideAttrs (old: {
              buildInputs = old.buildInputs ++ [ goPkg ];
            });
            build = hermeticCheck;
            buildStatic = hermeticCheckStatic;
            buildDatastartest = hermeticCheckDatastartest;
            buildBroadcast = hermeticCheckBroadcast;
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.actionlint
              pkgs.go-licenses
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.govulncheck
              pkgs.templ
            ];

            GOTOOLCHAIN = "local";

            shellHook = ''
              echo "go-datastar dev shell: $(go version)"
            '';
          };

          apps = {
            test = mkApp "test" [ goPkg ] ''
              go test ./... ./broadcast/... ./datastartest/... ./static/... -count=1 "$@"
            '';

            test-race = mkApp "test-race" [ goPkg ] ''
              go test ./... ./broadcast/... ./datastartest/... ./static/... -race -count=1 "$@"
            '';

            build = mkApp "build" [ goPkg ] ''
              go build ./... ./broadcast/... ./datastartest/... ./static/...
            '';

            vet = mkApp "vet" [ goPkg ] ''
              go vet ./... ./broadcast/... ./datastartest/... ./static/...
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              golangci-lint run ./... ./broadcast/... ./datastartest/... ./static/...
            '';

            # EXACT CI parity: pins the same golangci-lint version the CI lint
            # job installs (nixpkgs' golangci-lint drifts between versions,
            # which caused "green locally, red in CI" masters). THE pre-push
            # gate; first run compiles the linter from source (module cache).
            docspec = mkApp "docspec" [ goPkg ] ''
              go test -tags docspec -run TestDocspec -count=1 ./... ./broadcast/... ./datastartest/...
            '';

            # GOEXPERIMENT=jsonv2 removed 2026-10-11: json/v2 is a
            # go1.27-gated stable API, the flag is a no-op. Per-module loop:
            # the sibling modules are separate Go modules, so one root-level
            # `golangci-lint run ./static/...` typechecks against the wrong
            # module graph under GOWORK=off (exit 7).
            lint-ci = mkApp "lint-ci" [ goPkg ] ''
              for mod in . ./broadcast ./datastartest ./static; do
                echo "== lint-ci: $mod"
                (cd "$mod" && GOWORK=off go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./... --timeout 5m) || exit 1
              done
            '';

            # erraudit is NOT hermetically buildable (its dependency tree
            # contains private modules, e.g. go-finding), so this app
            # go-installs it and requires local GitHub credentials. One
            # directory per run — the tool rejects package patterns.
            erraudit = mkApp "erraudit" [ goPkg ] ''
              go install github.com/larsartmann/erraudit/cmd/erraudit@v0.3.0
              for mod in . ./broadcast ./datastartest ./static; do
                echo "== erraudit $mod"
                (cd "$mod" && "$HOME/go/bin/erraudit" . --type-aware --enforce-go-error-family --severity-threshold error)
              done
            '';

            govulncheck = mkApp "govulncheck" [ pkgs.govulncheck goPkg ] ''
              govulncheck ./... ./broadcast/... ./datastartest/... ./static/...
            '';

            coverage = mkApp "coverage" [ goPkg ] ''
              go test ./... ./broadcast/... ./datastartest/... ./static/... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            bench = mkApp "bench" [ goPkg ] ''
              echo "== bench: $(go version) / $(uname -s) $(uname -m) / $(nproc) cores"
              go test -run '^$' -bench . -benchmem "$@"
              (cd datastartest && go test -run '^$' -bench . -benchmem "$@")
            '';
          };
        };
    };
}
