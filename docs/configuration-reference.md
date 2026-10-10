# Configuration Reference

## Table of Contents

- [Global Options](#global-options)
- [Example Config](#example-config)

Here is a breakdown of all configuration options:

## Global Options

- **`install`** (Array)
  - Installation steps to execute.

  - See [Installer Configuration](./installer-configuration.md) for supported types and options that
    you can provide.

- **`debug`** (Boolean)
  - Enable or disable debug mode.
  - Default: `false`.

- **`check_updates`** (Boolean)
  - Enable or disable checking for updates before running operations.
  - Default: `false`.

- **`repo_update`** (Object)
  - Controls how repository index updates (e.g. `apt update`, `brew update`) are handled per
    installer type. Keys are installer types, values are one of:
    - `once` — Run the repo update at most once per sofmani run (default).
    - `always` — Run the repo update before every install/update operation.
    - `never` — Skip the repo update entirely.
  - Supported types: `brew`, `apt`, `apk`.
  - Default: `once` for all supported types.
  - Example:
    ```yaml
    repo_update:
      brew: once
      apt: always
      apk: never
    ```

- **`summary`** (Boolean)
  - Enable or disable the installation summary at the end.
  - The summary shows newly installed and upgraded software in a hierarchical format.
  - Default: `true`.

- **`category_display`** (String)
  - Controls how category headers are rendered in the output.
  - Values:
    - `border` — Full border with spacing before and after (default).
    - `border-compact` — Border without spacing before and after.
    - `minimal` — Plain text without border or spacing.
  - Default: `border`.

- **`defaults`** (Object)
  - Defaults to apply to all installer types, such as specifying supported platforms or commonly
    used flags.

  - **`defaults.type`**

    A mapping between each type (key) and their default options (value).
    - See [Installer Configuration](./installer-configuration.md) for supported types and options
      that you can override.

- **`env`** (Object)
  - Environment variables that will be set for the context of the installer.
  - OS environment variables are passed and may be overridden for this config and all of its
    installers here.

- **`env_command`** (String or Object)
  - A shell command that runs once at startup. The variables it **exports** are passed to every
    installer and command. Use it for secrets that aren't exported in the shell you run sofmani
    from.
  - The command runs in the shell process itself, so a shell function that exports variables works.
    For a loader that prints `export KEY=VALUE` lines, use `eval "$(loader)"`. Variables that are
    set without `export` are not picked up.
  - Values are passed on exactly as exported: unlike `env`, `~` and `$VARS` in them are not
    expanded, so secrets containing `$` or `=` arrive intact.
  - Anything the command prints goes to the terminal, and standard input is passed through, so the
    command can prompt to unlock a vault.
  - Variables already set by `env` or `platform_env` keep their configured value.
  - Values are never written to the log; only the variable names are.
  - Like `env`, the loaded variables reach a loaded manifest only when it inherits `env`.
  - Only read from the main config file, not from manifests it loads. Not supported on Windows.
  - As an object:
    - `command` (String, required): the command to run.
    - `interactive` (Boolean): start the shell with `-i`, so its rc file (`.zshrc`, `.bashrc`) loads
      first — plugins, functions and aliases are available to the command. Variables the rc file
      itself exports are not loaded, only those the command exports. Default: `false`.
    - `shell` (String): the shell to use. Default: `$SHELL`. It must be POSIX-compatible (`sh`,
      `bash`, `zsh`).
  - Examples:

    ```yaml
    # A loader that prints export lines
    env_command: eval "$(op inject -i ~/.config/sofmani/secrets.env.tpl)"

    # A function defined by your zsh config
    env_command:
      command: load_secrets
      interactive: true
    ```

- **`machine_aliases`** (Object)
  - A mapping of friendly names to machine IDs.
  - Use `sofmani --machine-id` to get the machine ID for each of your machines.
  - These aliases can then be used in installer `machines.only` and `machines.except` fields instead
    of the raw machine IDs.
  - The alias for the current machine is also available as the `{{ .DeviceIDAlias }}` template
    variable and the `$DEVICE_ID_ALIAS` environment variable in all commands.
  - Example:
    ```yaml
    machine_aliases:
      work-laptop: 5fa2a8e8193868df
      home-desktop: a1b2c3d4e5f67890
      home-server: fedcba0987654321
    ```

## Example Config

```yaml
debug: false
check_updates: true
summary: true
category_display: border
repo_update:
  brew: once
  apt: once
defaults:
  type:
    brew:
      platforms:
        only: ['macos']
install:
  - name: jq
    type: brew
```
