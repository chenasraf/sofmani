# Getting Started

To install sofmani, refer to the [Readme](/README.md) for instructions.

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

| Flag                | Description                        |
| ------------------- | ---------------------------------- |
| `-d`, `--debug`     | Enable debug mode.                 |
| `-D`, `--no-debug`  | Disable debug mode (default).      |
| `-u`, `--update`    | Enable update checking.            |
| `-U`, `--no-update` | Disable update checking (default). |
| `-h`, `--help`      | Display help information and exit. |

Each of these flags overrides the loaded config file, so while your default config can choose not to
check for updates by default, you or another user can add the `--update` flag to override this
behavior for a single run of the CLI.
