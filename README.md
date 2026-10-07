# hexgen

A single-binary CLI that writes a new Go API project in the shape of
[go-api-project-template](https://github.com/residwi/go-api-project-template):
a modular monolith of hexagonal feature modules, with layer rules enforced by
go-arch-lint. The project skeleton is embedded in the binary, so nothing is
fetched when you generate.

## What you get

- **Default (platform-only):** `cmd/api`, `internal/platform` (database,
  cache, jobqueue, web, ...), `internal/config`, an empty composition root
  (`internal/app`, `internal/server/router.go`), `internal/testutil`, goose
  migrations for the base schema and River's job tables, and Makefile, Docker,
  compose, CI, arch-lint and mockery config. No feature modules and no example
  domain.
- **With `--auth`:** the `auth` and `user` features (register/login/refresh,
  `/users/me`, admin user routes), the users migration and a dev admin seed.

The module path is rewritten to yours. The project name becomes the default
`APP_NAME`, `DB_NAME` and JWT issuer, and names the project's test containers.
`go mod tidy` runs on the output, and every Go file is gofmt-clean.

## Install

```bash
go install github.com/residwi/hexgen/cmd/hexgen@latest
```

Or build from source with `make build`, which outputs `bin/hexgen`.

## Usage

```bash
hexgen new <name> --module <path> [flags]
hexgen version

hexgen new myapp --module github.com/me/myapp
hexgen new myapp --module github.com/me/myapp --auth
```

`<name>` must be lowercase letters, digits, `-` or `_`, start with a letter and
end with a letter or digit (at most 63 characters): it becomes the database
name and the Docker container and image names. Run `hexgen new` with no name or
module in a terminal to be prompted for them.

### Flags for `new`

| Flag       | Description                                                                          |
| ---------- | ------------------------------------------------------------------------------------ |
| `--module` | Go module path (required), e.g. `github.com/me/myapp`                                |
| `--auth`   | include the auth and user features (register/login/refresh, /users/me, admin routes) |
| `--output` | output directory (default `./<name>`)                                                |
| `--force`  | write into a non-empty directory                                                     |
| `--git`    | run `git init` in the generated project                                              |
| `--check`  | run `go build ./...` in the output after generating                                  |

## How it works

1. Walk the embedded skeleton (`internal/template/testdata/skeleton`) and, with
   `--auth`, the auth overlay (`internal/template/testdata/auth`). A path in the
   overlay replaces the same path in the skeleton.
2. Render each `.tmpl` file with `text/template` (data: `Module`,
   `ProjectName`, `Auth`) and strip the suffix. A file that renders to only
   whitespace is omitted.
3. Replace the module placeholder `__MODULE__` with `--module` and
   `__PROJECT_NAME__` with the name, then gofmt every `.go` file.
4. Write the project, run `go mod tidy`, then `git init` and `go build ./...`
   if `--git` and `--check` are set.

## The skeleton

The skeleton started as a copy of go-api-project-template, and hexgen owns it:
its files carry no template names or shop words, and some (migrations, README,
Makefile) are reshaped. Edit the files under `internal/template/testdata/`
directly. `internal/template/UPSTREAM` records the template commit the skeleton
was last synced to.

### Syncing from the template

`make sync-template` (or `make sync-template TEMPLATE=<checkout>`, default
`../go-api-project-template`) lists the template commits since `UPSTREAM` and
applies each changed file's diff to the skeleton's copy, with the template's
names neutralized. It copies `go.mod` and `go.sum`, and moves `UPSTREAM` to the
template's HEAD. It reports files that hexgen has no copy of (shop modules,
docs) as skipped. It refuses to run while hexgen has uncommitted changes.

After a sync:

1. Port each reported `.rej` hunk by hand, then delete the `.rej` file. A hunk
   that does not apply to the skeleton (such as a README row it lacks) needs
   no port.
2. Run `make test`. The guard test fails on leftover `.rej` files or template
   names, and the e2e builds and tests both generated variants.
3. Commit, e.g. `chore(skeleton): sync go-api-project-template to <sha>`.

## Development

```bash
make test          # all tests, including the e2e (-race)
make test-short    # unit tests only (skips the e2e)
make e2e           # generate both variants and check them
make lint          # golangci-lint (standard linter set)
make build         # build bin/hexgen
make help          # list all targets
```

The e2e generates the platform-only and `--auth` projects. It checks each for
dropped paths and leftover template names, then runs `go mod tidy`,
`go build`, `go vet`, `gofmt -l` and go-arch-lint inside it. When Docker is
available it also runs the generated project's own `go test ./...`. It needs
GOPROXY access for the generated projects' dependencies.
