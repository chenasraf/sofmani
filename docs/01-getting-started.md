# Getting Started

## Installation

### Download Precompiled Binaries

Precompiled binaries for `sofmani` are available for **Linux**, **macOS**, and **Windows**:

- Visit the [Releases Page](https://github.com/chenasraf/sofmani/releases/latest) to download the
  latest version for your platform.

### Homebrew (macOS/Linux only)

Install from a custom tap:

```bash
brew install chenasraf/tap/sofmani
```

---

### Linux

You can install `sofmani` by downloading the release tar, and extracting it to your preferred
location.

- You can see an example script for install here: [install.sh](/install.sh)
- The example script can be used for actual install, use this command to download and execute the
  file (use at your own discretion):

  ```sh
  curl https://raw.githubusercontent.com/chenasraf/sofmani/master/install.sh | sh
  ```

## Config file location

The config file can be in YAML or JSON format.

You can place the config file anywhere, and provide the path to sofmani CLI to load. If you don't
give it an explicit path, the CLI will attempt to find a `sofmani.yml` or `sofmani.json` file ine
the following directories (ordered by priority):

1. Current working directory
1. `$HOME/.config` directory
1. Home directory

## Using sofmani

The sofmani CLI will iterate through each of your install steps (called "Installers") and execute
them in sequence.

Installers can be grouped and nested arbitrarily, and loaded either directly from your config, or
loaded from an external additional config, either a local file or a remote file hosted on a git
repository.

### CLI Flags

You can call `sofmani` with the following flags to alter the behavior for the current run:

| Flag                | Description                                             |
| ------------------- | ------------------------------------------------------- |
| `-d`, `--debug`     | Enable debug mode.                                      |
| `-D`, `--no-debug`  | Disable debug mode (default).                           |
| `-u`, `--update`    | Enable update checking.                                 |
| `-U`, `--no-update` | Disable update checking (default).                      |
| `-f`, `--filter`    | Filter by installer name (can be used multiple times)\* |
| `-h`, `--help`      | Display help information and exit.                      |
| `-v`, `--version`   | Display version information and exit.                   |

Each of these flags overrides the loaded config file, so while your default config can choose not to
check for updates by default, you or another user can add the `--update` flag to override this
behavior for a single run of the CLI.

\* The filter argument accepts multiple values.

- To only run installers that contain "sofmani" in their name, use `-f sofmani`.
- To run all installers except those that contain "sofmani", use `-f "!sofmani"`.
- To only installers that contain "sofmani", but exclude "sofmani-config", use
  `-f sofmani -f "!sofmani-config"`.

#### Examples

Search for the config in one of the default directories, and enable update checking:

```sh
sofmani -u
```

Load a specific config file, and enable debug mode:

```sh
sofmani -d sofmani.yml
```

Load a config file, and only run installers matching `brew` in their name:

```sh
sofmani -f brew sofmani.yml
```
