# df2

df2 is a command-line tool for managing and optimising the files and database of defacto2.net. It is broken down into five parts.

**clean** used to discover and remove orphan files that exist on the server but have no matching database entries.

**demozoo** interacts with the Demozoo.org API to synchronise data and fetch linked download files.

**fix** repairs malformed data and generates missing assets from distinct sources plus batch-approve pre-screened files.

**html** generates HTML pages for groups, peoples and the site map.

**proof** automates the parsing of files tagged as *release-proof*.

---

```bash
A tool to optimise and manage defacto2.net
Copyright © 2020 Ben Garrett
https://github.com/Defacto2/df2

Usage:
  df2 [command]

Available Commands:
  clean       Discover or clean orphan files
  config      Configure the settings for this tool
  demozoo     Interact with Demozoo.org upload submissions
  fix         Fixes database entries and records
  help        Help about any command
  html        HTML and sitemap generator
  proof       Batch handler files tagged as #release-proof

Flags:
      --config string   config file (default is $HOME/.df2.yaml)
  -h, --help            help for df2
  -q, --quiet           suspend feedback to the terminal
      --version         version for df2

Use "df2 [command] --help" for more information about a command.
```

### Compile and install

```
git clone git@github.com:Defacto2/df2.git
cd df2
go install
df2 --version
```

### Dependencies

The `df2 fix text` command requires the installation of [AnsiLove/C](https://github.com/ansilove/ansilove) to the system PATH.

### Configuration

To view and test the database and directory configurations.

```
df2 config info
```

To change the configuration.

```
df2 config edit
```