# df2

![Go](https://github.com/Defacto2/df2/workflows/Go/badge.svg)

df2 is a command-line tool for managing plus optimising the files and database of defacto2.net. It is broken down into five parts.

**approve** all validated file records that are ready to go live.

**clean** discover and remove orphan files that exist on the server but have no matching database entries.

**config** adjust the default settings for this `df2` tool.

**demozoo** interacts with the [Demozoo.org API](http://demozoo.org/api/v1/) to synchronise data and fetch linked download files.

**fix** malformed data and generates missing assets from distinct sources.

**lookup** a website record URL by its database ID or UUID.

**new** checks Demozoo and _#releaseproof_ submissions, fetches downloads, generate previews and repairs any malformed data.

**output** generates webpages for groups, peoples and the site map.

**proof** automates the parsing of files tagged as _#releaseproof_.

**shrink** moves old SQL backups and incoming uploads to the user home directory ready for download and removal.

---

```bash
A tool to optimise and manage defacto2.net
Copyright © 2020-21 Ben Garrett
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
  new         Handler for files flagged as waiting to go live
  output      JSON, HTML, SQL and sitemap generator
  proof       Handler for files tagged as #releaseproof
  shrink      Reduces the space used in directories

Flags:
      --config string   config file (default is config.yaml)
  -h, --help            help for df2
  -q, --quiet           suspend feedback to the terminal
      --version         version for df2

Use "df2 [command] --help" for more information about a command.
```

## Install

Is built on [Go v1.16](https://golang.org/doc/install) and is packaged for the Ubuntu Linux platform.

```bash
cd ~/downloads
wget https://github.com/Defacto2/df2/releases/latest/download/df2.deb
dpkg -i df2.deb # this also works for updates
df2 --version
```

### Dependencies

The `df2 fix text` command requires the installation of [AnsiLove/C](https://github.com/ansilove/ansilove) in the system `PATH`.

[WebP support](https://en.wikipedia.org/wiki/WebP) image conversion needs [libwebp](https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html). PNG image compression relies on [pngquant](https://pngquant.org).


#### Ubuntu installation

```bash
sudo apt install -y ansilove pngquant webp
```

## Configuration

To view and test the database and directory configurations.

```bash
df2 config info
```

To change the configuration.

```bash
df2 config edit
```

## Builds

All changes should be tested with the `golangci-lint` [Go linters aggregator](https://golangci-lint.run/).

**Local builds** can be done using the standard tools.

```bash
git clone git@github.com:Defacto2/df2.git
cd df2
go build .
```

**GitHub releasing** requires [GoReleaser](https://goreleaser.com) and [upx](https://upx.github.io/).

**Test** the building of the release.

```bash
cd df2/internal
./test.bash
```

**Build** the release. This will request a __semantic version__ and a release __comment__.
```bash
cd df2/internal
./release.bash
```

**Deploy** the release to GitHub.
```bash
cd df2/internal
./deploy.bash
```

#### Troubleshoot tags

A common goreleaser error requires the removal of unreleased tags.

> goreleaser git tag _v1.0.0_ was not made against commit

```bash
cd df2/internal

# list tags
git tag -l

# delete local tag
git tag -d v1.0.0

# delete github tag
cd df2
git push --delete origin v1.0.0
```