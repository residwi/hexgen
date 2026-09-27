# go-project-generator (`hexgen`)

A single-binary CLI that bootstraps a fresh Go API project from
[go-api-project-template](https://github.com/residwi/go-api-project-template) —
the `rails new` of that template. It fetches the template, carves it down to a
working **auth + user** skeleton, overlays curated override files, rewrites the
module path, and writes a ready-to-run project.

## What you get

- A compiling project with the `auth` and `user` features, the shared `core` /
  `platform` / `middleware` infrastructure, tests, and the `users` migration.
- The template's e-commerce features (cart, order, payment, …) and their
  platform packages, wiring, mocks, and migrations are dropped.
- Config, router, error mapping, seeds, Docker, and CI come pre-trimmed for the
  skeleton; the module path is rewritten to yours and every Go file is gofmt-clean.
- With `--worker`: a generic background worker (`cmd/worker` built on
  `internal/platform/jobs.Runner`) plus its config and tooling.

## Install

```bash
go install github.com/residwi/go-project-generator/cmd/hexgen@latest
```

Or build from source with `make build`, which outputs `bin/hexgen`.

## Usage

```bash
hexgen new <name> --module <path> [flags]
hexgen version

hexgen new myapp --module github.com/me/myapp
```

Run `hexgen new` with no name or module in a terminal to be prompted for them.

### Flags for `new`

| Flag       | Description                                                             |
| ---------- | ----------------------------------------------------------------------- |
| `--module` | Go module path (required), e.g. `github.com/me/myapp`                   |
| `--ref`    | template ref to fetch (default `main`)                                  |
| `--output` | output directory (default `./<name>`)                                   |
| `--force`  | write into a non-empty directory                                        |
| `--git`    | run `git init` in the generated project                                 |
| `--check`  | run `go build ./...` in the output after generating                     |
| `--worker` | include a background worker (cmd/worker + jobs runner + config/tooling) |

## How it works

1. Download the template tarball from GitHub (`codeload`) at `--ref`.
2. Carve: drop e-commerce features, the payment/email/storage platform packages,
   cross-feature wiring, and their migrations/mocks (see
   `internal/scaffold/rules.go`).
3. Overlay the embedded override tree (`internal/assets/files/`), rendered with
   `text/template` so `--worker`-only content is gated behind `{{if .Worker}}`.
4. Rewrite the template module path to `--module`, substitute the project name,
   and re-format every generated `.go` file.

## Development

```bash
make test          # all tests, including the network e2e (-race)
make test-short    # unit tests only (no network)
make e2e           # end-to-end: fetch template, generate, build both variants
make lint          # golangci-lint (standard linter set)
make build         # build bin/hexgen
make help          # list all targets
```
