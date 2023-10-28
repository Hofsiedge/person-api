{
  description = "EffectiveMobile test project";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
        fmt = {
          reset = ''\033[0m'';
          red = ''\033[0;31m'';
          green = ''\033[0;32m'';
          bold_red = ''\033[1;31m'';
          bold_green = ''\033[1;32m'';
        };
        formatted-echo = format: msg: ''echo -e "${format}${msg}${fmt.reset}"'';

        # executable binary shell scripts
        scripts = with builtins;
          attrValues (mapAttrs pkgs.writeShellScriptBin {
            preview-readme = ''
              ${pkgs.python311Packages.grip}/bin/grip $ROOT/README.md
            '';
            generate-from-openapi = ''
              {
                redocly lint $ROOT/openapi.yaml                    && \
                redocly build-docs $ROOT/openapi.yaml \
                        -o $ROOT/docs/api.html                     && \
                ${formatted-echo fmt.bold_green "Generated docs!"} || \
                ${formatted-echo fmt.bold_red "Invalid API spec"}
              } && {
                pushd $ROOT/src > /dev/null
                go generate ./...
                popd > /dev/null
              } &&
              ${formatted-echo fmt.bold_green "Generated Go files!"}   || \
              ${formatted-echo fmt.bold_red "Error generating Go files"}
            '';
            # if you want to reinstall the tools that are not
            # managed by nix, run this and re-enter nix shell
            remove-non-nix-tools = ''
              sudo rm -rf $ROOT/.nix-node $ROOT/.nix-go
            '';
            test-server = ''
              pushd $ROOT/src > /dev/null
              CGO_ENABLED=1 go test -race -vet="" -coverpkg=./... \
                -coverprofile=cover.out ./...
              go tool cover -html=cover.out -o cover.html
              popd > /dev/null
            '';
            test-database = ''
              docker compose \
                --file $ROOT/compose.yaml \
                --file $ROOT/compose.db-test.yaml \
                --profile db-test \
                --env-file $ROOT/.env \
                run db --rm --build
            '';
            wrapped-migrate = ''
              set -a
              source $ROOT/.env
              set +a
              migrate \
                -source file://postgres/migrations \
                -database "postgres://$DB_USERNAME:$DB_PASSWORD@localhost:5432/$DB_NAME?sslmode=disable" \
                $@
            '';
            test-server-database-integration = ''
              set -a
              source $ROOT/.env
              set +a
              pushd $ROOT/src
              export DB_CONN="postgres://$DB_USERNAME:$DB_PASSWORD@localhost:5432/$DB_NAME"
              go test internal/repo/postgres/postgres_integration_test.go -integration true
              popd
            '';
          });
        tools = with pkgs; [
          git

          # go
          go_1_21
          golangci-lint
          gotools
          go-tools

          #sql
          go-migrate
          sqlfluff

          # for redocly
          nodejs_latest
        ];
      in {
        devShells.default = pkgs.mkShell rec {
          name = "test-project";
          buildInputs = tools ++ scripts;
          shellHook = ''
            export ROOT=$PWD

            # git setup
            git config core.hooksPath .githooks

            # custom prompt
            export PS1="\n(${name}) \[${fmt.bold_green}\][\[\e]0;\u@\h: \w\a\]\u@\h:\w]\n\$\[${fmt.reset}\] ";

            # -- node.js setup --
            mkdir -p .nix-node
            # make npm local to not fill the whole system with garbage
            export NODE_PATH=$PWD/.nix-node
            export NPM_CONFIG_PREFIX=$NODE_PATH
            # make executables available
            export PATH=$NODE_PATH/bin:$PATH

            # -- go setup --
            mkdir -p .nix-go
            # make go use a local directory
            export GOPATH=$PWD/.nix-go
            # make executables available
            export PATH=$GOPATH/bin:$PATH
            # -- end of go setup --

            # -- installing additional tools --
            # install redocly if not installed already
            command -v redocly >/dev/null 2>&1 || \
              npm install @redocly/cli --global
            # install oapi-codegen if not installed already
            command -v oapi-codegen >/dev/null 2>&1 || \
              go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
          '';
          env = {
            CGO_ENABLED = 0; # pkgs.delve does not work otherwise
          };
        };
      }
    );
}
