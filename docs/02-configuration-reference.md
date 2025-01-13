# Configuration Reference

Here is a breakdown of all configuration options:

## Global Options

- **`install`** (Array)

  - Installation steps to execute.

  - See [Installer Types](./03-installer-types.md) for supported types and options that you can
    provide.

- **`debug`** (Boolean)

  - Enable or disable debug mode.
  - Default: `false`.

- **`check_updates`** (Boolean)

  - Enable or disable checking for updates before running operations.
  - Default: `false`.

- **`defaults`** (Object)

  - Defaults to apply to all installer types, such as specifying supported platforms or commonly
    used flags.

  - **`defaults.type`**

    A mapping between each type (key) and their default options (value).

    - See [Installer Types](./03-installer-types.md) for supported types and options that you can
      override.

- **`env`** (Object)
  - Environment variables that will be set for the context of the installer.
  - OS environment variables are passed and may be overridden for this config and all of its
    installers here.

## Example config base

```yaml
debug: false
check_updates: true
defaults:
  type:
    brew:
      platforms:
        only: ['macos']
install:
  - name: jq
    type: brew
```
