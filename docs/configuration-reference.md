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
