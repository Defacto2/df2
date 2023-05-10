# df2

[![Go Report Card](https://goreportcard.com/badge/github.com/Defacto2/df2)](https://goreportcard.com/report/github.com/Defacto2/df2)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/Defacto2/df2)](https://github.com/Defacto2/df2/releases)
![GitHub](https://img.shields.io/github/license/Defacto2/df2?style=flat)

The `df2` program is a terminal tool for managing plus optimising the files and database of [defacto2.net](https://defacto2.net) that is broken down into multiple parts.

```
The tool to optimise and manage defacto2.net
Copyright © 2020-23 Ben Garrett
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

## Install or update

`df2` is built on [Go](https://golang.org/doc/install) and is packaged for [Debian Linux](https://www.debian.org/intro/index).

```bash
# download the package
wget https://github.com/Defacto2/df2/releases/latest/download/df2.deb

# install or update the package
dpkg -i df2.deb

# test the new install
df2 --version
```

### Dependencies

- The `df2 fix text` command requires the installation of [AnsiLove/C](https://github.com/ansilove/ansilove) in the system `PATH`.
- [WebP support](https://en.wikipedia.org/wiki/WebP) image conversion needs [libwebp](https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html). 
- PNG image compression relies on [pngquant](https://pngquant.org). 
- Image conversion needs both [imagemagick](https://imagemagick.org) and [netpbm](http://netpbm.sourceforge.net/).

#### Dependency installation on Ubuntu

```bash
# required dependencies
sudo apt install -y ansilove imagemagick netpbm pngquant webp

# optional file archivers
sudo apt install -y arj lhasa unrar unzip
```

### Database dependancy

The `df2` program expects local access to the [Defacto2 database](https://github.com/Defacto2/database).

## Configuration

To view and test the [database](https://github.com/Defacto2/database) and directory configurations.

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

## Use the code on your own system

Confirm the installation of [Go](https://golang.org/doc/install) on Linux, WSL or macOS.

```sh
$ go version

go version go1.19.5 linux/amd64
```

Clone the `Defacto2/df2` repository.

```sh
$ git clone git@github.com:Defacto2/df2.git

Cloning into 'df2'...
```

Run the `df2` tool. It may take a moment on the first run as it downloads dependencies.

```sh
$ cd df2
$ go run . --version

  version  0.0.0 (developer build)
     path  /tmp/go-build3011510136/b001/exe/df2
   commit  unknown
     date  unknown
       go  v1.19.5 linux/amd64
```

Build the `df2` tool (not usually required).

```sh
$ cd df2
$ go build
$ ./df2 --version

  version  0.0.0 (developer build)
     path  /home/ben/df2/df2
   commit  3352b5393353d4e09cf19f636d4be593d961cece
     date  2023 Jan 3, 02:01 UTC
       go  v1.19.5 linux/amd64
```

## Releasing

[GitHub Actions](https://github.com/features/actions) combined with [GoReleaser](https://goreleaser.com/) handles the building process when new release tags are created.

All changes should be tested with the `golangci-lint` [Go linters aggregator](https://golangci-lint.run/).

## Usage tips

### `import`

Use [rsync](https://rsync.samba.org/) to upload to the remote server.

```bash
# rsync -a  optional archive mode
# rsync -P  optional progress bar and keeps the partially transferred files
# some_archive.rar is a local file to upload
# user@[ip address] is the remote destination and user account with SSH access.
# :~/downloads will place the some_archive.rar to the user account downloads directory.
rsync -aP some_archive.rar user@[ip address]:~/downloads
```

On the remote server.

```
df2 import ~/downloads/some_archive.rar --limit=10
```
