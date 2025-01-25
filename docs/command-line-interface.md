# Command Line Interface (CLI)

The sofmani CLI will iterate through each of your install steps (called "Installers") and execute
them in sequence.

Installers can be grouped and nested arbitrarily, and loaded either directly from your config, or
loaded from an external additional config, either a local file or a remote file hosted on a git
repository.

## CLI Flags

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

### Installer Filters

The filter argument accepts multiple values.

The following filter types are available:

- `-f <name>` - filter by name
- `-f tag:<tag>` - filter by tag name
- `-f type:<type>` - filter by type (brew, shell, etc)

Each of the above filters can be negated by prefixing with `!`. For example, to exclude installers
containing the tag `"system"`, use `-f "!tag:system"`. See more information about tags in the
documentation for (Installer Configuration)[./installer-configuration.md#fields].

If there are no filters in the command flags, then all the installers will run.

If there are filters, each installer will have to match an inclusion filter. Exclusion filters can
be combined to then remove from the filtered installers, even if they already matched to be
included.

- To only run installers that contain "sofmani" in their name, use `-f sofmani`.
- To run all installers except those that contain "sofmani", use `-f "!sofmani"`.
- To only installers that contain "sofmani", but exclude ones tagged "config", use
  `-f sofmani -f "!tag:config"`.

## Examples

Search for the config in one of the default directories, and enable update checking:

```sh
sofmani -u
```

Load a specific config file, and enable debug mode:

```sh
sofmani -d sofmani.yml
```

Load a config file, and only run installers of type `brew`:

```sh
sofmani -f type:brew sofmani.yml
```
