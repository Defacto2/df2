# df2

![Go](https://github.com/Defacto2/df2/workflows/Go/badge.svg)

df2 is a terminal tool for managing plus optimising the files and database of [defacto2.net](https://defacto2.net). 
It is broken down into multiple parts.

```
The tool to optimise and manage defacto2.net
Copyright © 2020-22 Ben Garrett
https://github.com/Defacto2/df2

Usage:
  df2 [flags]
  df2 [command]

Admin:
  approve     Approve the records that are ready to go live.
  fix         Fixes database entries and records.
  new         Manage files marked as waiting to go live (default).
  output      Generators for JSON, HTML, SQL and sitemap documents.
  proof       Manage records tagged as #releaseproof.

Drive:
  clean       Discover or clean orphan files.
  shrink      Reduces the space used in directories.

Remote:
  apis        Batch data synchronization with remote APIs.
  demozoo     Interact with Demozoo submissions.
  lookup      Lookup the file URL of a record's ID or UUID.
  test        Test various features of the website or database that cannot be fixed with automation.

Additional Commands:
  config      Configure the settings for this tool.
  help        Help about any command

Flags:
      --ascii     suppress all ANSI color feedback
  -h, --help      help for df2
      --quiet     suppress all feedback except for errors
  -v, --version   version and information for this program

Use "df2 [command] --help" for more information about a command.
```

## Install

df2 is built on [Go](https://golang.org/doc/install) and is packaged for [Debian](https://www.debian.org/intro/index) Linux.

```bash
wget https://github.com/Defacto2/df2/releases/latest/download/df2.deb
dpkg -i df2.deb # also works for updating
df2 --version
```

### Dependencies

The `df2 fix text` command requires the installation of [AnsiLove/C](https://github.com/ansilove/ansilove) in the system `PATH`.

[WebP support](https://en.wikipedia.org/wiki/WebP) image conversion needs [libwebp](https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html). PNG image compression relies on [pngquant](https://pngquant.org). Image conversion needs both [imagemagick](https://imagemagick.org) and [netpbm](http://netpbm.sourceforge.net/).

#### Ubuntu installation

```bash
sudo apt install -y ansilove imagemagick netpbm pngquant webp
# optional file archivers
sudo apt install -y arj lhasa unrar unzip
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

## Docker container

The Docker container runs on a [Go container](https://hub.docker.com/_/golang) built in Debian Linux. 
The main purpose is unit testing and compiling of the Go source code in a Linux environment.

```sh
# change directory to the local repository
cd df2
# synchronize any remote repository tags
git pull 
# build the current directory as an image tagged as 'df2'
docker build --tag df2 . 
# run the image tagged as 'df2' with the container name 'df2-test'
docker run -it --name df2-test df2
```

## Source code and building

[GitHub Actions](https://github.com/features/actions) combined with [GoReleaser](https://goreleaser.com/) handles the building process when new release tags are created.

All changes should be tested with the `golangci-lint` [Go linters aggregator](https://golangci-lint.run/).


