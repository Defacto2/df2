# df2

![Go](https://github.com/Defacto2/df2/workflows/Go/badge.svg)

df2 is a command-line tool for managing and optimising the files and database of defacto2.net. It is broken down into five parts.

**approve** approves all validated file records that are ready to go live.

**clean** used to discover and remove orphan files that exist on the server but have no matching database entries.

**demozoo** interacts with the Demozoo.org API to synchronise data and fetch linked download files.

**fix** repairs malformed data and generates missing assets from distinct sources plus batch-approve pre-screened files.

**html** generates HTML pages for groups, peoples and the site map.

**proof** automates the parsing of files tagged as _releaseproof_.

---

```
A tool to optimise and manage defacto2.net
Copyright © 2020 Ben Garrett
https://github.com/Defacto2/df2

Usage:
  df2 [command]

Available Commands:
  approve     Approve the file records that are ready to go live
  clean       Discover or clean orphan files
  config      Configure the settings for this tool
  demozoo     Interact with Demozoo.org upload submissions
  fix         Fixes database entries and records
  help        Help about any command
  lookup      Lookup the file URL of a database ID or UUID
  output      JSON, HTML, SQL and sitemap generator
  proof       Handler for files tagged as #releaseproof
  waiting     Handler for files flagged as waiting to go live

Flags:
      --config string   config file (default is config.yaml)
  -h, --help            help for df2
  -q, --quiet           suspend feedback to the terminal
      --version         version for df2

Use "df2 [command] --help" for more information about a command.
```

### Compile and install

```bash
git clone git@github.com:Defacto2/df2.git
cd df2
go install
df2 --version
```

### Dependencies

The `df2 fix text` command requires the installation of [AnsiLove/C](https://github.com/ansilove/ansilove) to the system PATH.

### Configuration

To view and test the database and directory configurations.

```bash
df2 config info
```

To change the configuration.

```bash
df2 config edit
```

### Development notes

df2 uses [Go modules](https://github.com/golang/go/wiki/Modules) and Go v1.14.

It requires the use of the `go xxx -mod=vendor` argument which is set as default from v.1.14 onwards.

`go env` should return either `GO111MODULE=on` or `GO111MODULE=auto`.

#### Updating dependencies

```bash
go list -m all # List all direct and indirect dependencies

go list -u -m all # List all possible upgrades

go get -u ./... # Update all (major)

go get -u=patch ./... # Patch update all (minor)

go build ./... # Build all

go test ./... # Test all

go mod tidy # clean go.mod by removing all unused dependencies
```
