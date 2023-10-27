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

## Docker
`docker compose` is used to manage services.

<table>
  <tr>
    <th>Profile</th>
    <th>Files</th>
    <th>Description</th>
    <th>Running</th>
  </tr>

  <tr>
    <td><code>dev</code></td>
    <td>
      <code>compose.yaml</code><br>
      <code>compose.override.yaml</code>
    </td>
    <td>Development profile</td>
    <td><code>docker compose --profile dev up</code></td>
  </tr>

  <tr>
    <td><code>db-test</code></td>
    <td>
      <code>compose.yaml</code><br>
      <code>compose.db-test.yaml</code>
    </td>
    <td>DB testing profile</td>
    <td>
      <code>test-database</code> (nix shell)<br><br>
      <code>docker compose --file compose.yaml --file compose.db-test.yaml --profile db-test --env-file .env run db --rm --build</code>
    </td>
  </tr>
</table>

## PostgreSQL
`PostgreSQL` 16 with `PL/pgSQL` is used.

Migrations are managed with [`migrate`](https://github.com/golang-migrate/migrate)
and stored in [postgres/migrations](/postgres/migrations).
They include both main DB migrations and test functions migrations.

[`pgTAP`](https://github.com/theory/pgtap) `v1.3.1` is used to write and run `postgres` tests.

To run DB tests start the `db-test` profile (the command is listed in `docker` section).

### Warning
Building the image for `db-test` takes a long time (~11 minutes on my laptop)
and might fail in unexpected ways. This seems to be due to how unstable `cpan`
(and Perl infrastructure in general) is.

Some possible problems (that I've encountered so far):
- Image is built, but `pg_prove` is missing when running tests
- Obscure errors (segfault or network-related) with `cpan` on `pg_prove` installation

If something happens, you can just rebuild the image from scratch:
```bash
docker compose --file compose.yaml --file compose.db-test.yaml \
  --profile db-test --env-file .env build --no-cache
```

## Go
`go` `1.21` is used.

All source files are in the [`src`](/src) directory (including `go.mod` and
`go.sum`).

Tests can be run with `test-server` command (nix) or 
```bash
CGO_ENABLED=1 go test -race -vet="" -coverpkg=./... \
  -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
```
Source code is linted with [`golangci-lint`](https://golangci-lint.run/).
[`.githooks/pre-commit`](/.githooks/pre-commit) contains a git hook to
run linters before commit. If you add it to your git hooks (done
automatically for nix shell) you can run it with
```bash
git hook run pre-commit
```

## Integration testing
To test integration of server and DB:
1. Run `docker compose --profile dev up` to start dev DB instance
2. Wait for migrations to finish
2. Run `test-server-database-integration` from nix shell (or run equivalent commands)
3. Stop the DB and remove the volumes: `docker compose --profile dev down -v`
