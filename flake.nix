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
              ${pkgs.python311Packages.grip}/bin/grip README.md
            '';
            render-api-docs = ''
              redocly lint openapi.yaml                          && \
              redocly build-docs openapi.yaml -o docs/api.html   && \
              ${formatted-echo fmt.bold_green "Generated docs!"} || \
              ${formatted-echo fmt.bold_red "Invalid API spec"}
            '';
            # if you want to reinstall the tools that are not
            # managed by nix, run this and re-enter nix shell
            remove-non-nix-tools = ''
              sudo rm -rf .nix-node .nix-go
            '';
            test-server = ''
              pushd src
              CGO_ENABLED=1 go test -race -vet="" -coverpkg=./... \
                -coverprofile=cover.out ./...
              go tool cover -html=cover.out -o cover.html
              popd
            '';
          });
        tools = with pkgs; [
          git
          go_1_21
          golangci-lint
          gotools
          go-tools
          impl

          # for redocly
          nodejs_latest
        ];
      in {
        devShells.default = pkgs.mkShell rec {
          name = "test-project";
          buildInputs = tools ++ scripts;
          shellHook = ''
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
