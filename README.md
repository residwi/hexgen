# go-project-generator (`gen`)

A CLI that bootstraps a fresh Go API project from
[go-api-project-template](https://github.com/residwi/go-api-project-template),
carving it to a skeleton (auth + user, tests included) and rewriting the module path.

## Install

    go install github.com/residwi/go-project-generator@latest

## Usage

    gen new myapp --module github.com/me/myapp

Flags: `--ref` (template ref, default `main`), `--output`, `--force`, `--git`, `--check`.
Run `gen new` with no flags in a terminal to be prompted for name and module.

## Development

    go test ./...                 # unit tests (no network)
    go test -run TestEndToEnd .   # end-to-end: fetch template + build generated project (needs network)
