# Test project

## Nix
I use [`nix` package manager](https://nixos.org/) with 
[flakes](https://nixos.wiki/wiki/Flakes#Enable_flakes) to manage the development
environment. This includes all the dev dependencies and convenience scripts.
The environment is defined in [`flake.nix`](flake.nix).

## OpenAPI
This project uses [`OpenAPI 3.0`](https://swagger.io/specification/v3/)
(formerly known as `Swagger` for old versions) definition
([openapi.yaml](openapi.yaml)) to generate [API docs](docs/api.html)
via [`redocly`](https://github.com/Redocly/redocly-cli/) and server boilerplate
code via [`oapi-codegen`](https://github.com/deepmap/oapi-codegen).
