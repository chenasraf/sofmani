# sofmani

**`sofmani`** stands for **[Sof]tware [Mani]fest**. It is a robust and flexible provisioning tool
written in **Go**, designed to simplify software installations, configuration syncing, and system
provisioning for both personal and work computers. With a single config file, `sofmani` automates
locating, installing, or updating software and configurations, making system setup quick and
reproducible.

![Release](https://img.shields.io/github/v/release/chenasraf/sofmani)
![Downloads](https://img.shields.io/github/downloads/chenasraf/sofmani/total)
![License](https://img.shields.io/github/license/chenasraf/sofmani)

---

## 🚀 Features

- Install and provision software using a **declarative YAML/JSON configuration**.
- Multi-platform support: macOS, Linux, or Windows.
- Modular and extendable **installer types**: shell scripts, rsync, Homebrew taps, and more.
- Configurable **platform-specific behaviors**.
- Automatic software updates using custom logic.
- Group software installations into logical "steps" with sophisticated orchestration.

---

## 🎯 Installation

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

  To change the install location, provide an env variable `$INSTALL_DIR` to the script:

  ```sh
  # below is the default value, change as needed:
  curl https://raw.githubusercontent.com/chenasraf/sofmani/master/install.sh | INSTALL_DIR=~/.local/bin sh
  ```

## ✨ Getting Started

`sofmani` works based on a configuration file written in **YAML** or **JSON**. Below is an annotated
example configuration to demonstrate most of its options.

```yaml
debug: true # Global debug mode (optional).
check_updates: true # Enable update checking (optional).
defaults: # Define default behaviors for installer types.
  type:
    brew:
      platforms:
        only: ['macos'] # Only run this installer type on macOS.

install: # Declare installation steps:
  - name: nvim # Identifier for this step.
    type: rsync
    opts:
      source: ~/.dotfiles/.config/nvim/
      destination: ~/.config/nvim/
      flags: --delete --exclude .git --exclude .DS_Store

  - name: lazygit
    type: group # Logical group of steps.
    steps:
      - name: lazygit
        type: brew
        opts:
          tap: jesseduffield/lazygit
      - name: lazygit # Additional step for Linux systems only.
        type: shell
        platforms:
          only: ['linux']
        opts:
          command: |
            cd $(mktemp -d)
            latest_version=$(curl -s https://... )
            ...
```

---

## 🔧 Usage

Run `sofmani` with an optional configuration file or flags. Example:

```bash
sofmani my-config.yaml
```

See [the documentation](/docs) for more information and examples.

### Command-Line Flags

The following flags are supported to customize behavior:

| Flag                | Description                                           |
| ------------------- | ----------------------------------------------------- |
| `-d`, `--debug`     | Enable debug mode.                                    |
| `-D`, `--no-debug`  | Disable debug mode (default).                         |
| `-u`, `--update`    | Enable update checking.                               |
| `-U`, `--no-update` | Disable update checking (default).                    |
| `-f`, `--filter`    | Filter by installer name (can be used multiple times) |
| `-h`, `--help`      | Display help information and exit.                    |
| `-v`, `--version`   | Display version information and exit.                 |

If a configuration file is not explicitly provided, `sofmani` attempts to locate a `sofmani.yaml`,
`sofmani.yml` or `sofmani.json` in the following directories, in this order (first match is used):

1. Current directory
1. `$HOME/.config` directory
1. Home directory

If no file is found or provided, sofmani will fail to start.

For more information, see [Configuration Reference](./docs/configuration-reference.md)

---

## 📚 Configuration Reference

Here is a quick breakdown of all configuration options.

For a full breakdown with all the supported options, see [the docs](./docs/installer-types.md).

### Global Options

| Field           | Type    | Description                                                                                                                                                            |
| --------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `debug`         | Boolean | Enable or disable debug mode. Default: `false`.                                                                                                                        |
| `check_updates` | Boolean | Enable or disable checking for updates before running operations. Default: `false`.                                                                                    |
| `defaults`      | Object  | Defaults to apply to all installer types, such as specifying supported platforms or commonly used flags.                                                               |
| `env`           | Object  | Environment variables that will be set for the context of the installer. OS env vars are passed, and may be overridden for this config and all of its installers here. |
| `install`       | Array   | Installation steps to execute.                                                                                                                                         |

### `install` Node

The `install` field describes the steps to execute. Each step represents an action or group of
actions. Steps can be of **several types**, such as `brew`, `rsync`, `shell`, and more.

| Field              | Type                  | Description                                                                                                                                                                                                                                                                                   |
| ------------------ | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`             | String (required)     | Identifier for the step. It does not have to be unique, but is usually used to check for the app's existence, if applicable (can be overridden using `bin_name`)                                                                                                                              |
| `type`             | String (required)     | Type of the step. See [supported types](#supported-type-of-installers) for a comprehensive list of supported values.                                                                                                                                                                          |
| `platforms`        | Object (optional)     | Platform-specific execution controls. See `platforms` subfields below.                                                                                                                                                                                                                        |
| `platforms.only`   | Array of Strings      | Platforms where the step should execute (e.g., `['macos', 'linux']`). Supercedes `platforms.except`.                                                                                                                                                                                          |
| `platforms.except` | Array of Strings      | Platforms where the step should **not** execute; replaces `platforms.only`.                                                                                                                                                                                                                   |
| `steps`            | Array of Installers   | Sub-steps for `group` type. Allows nesting multiple steps together.                                                                                                                                                                                                                           |
| `opts`             | Object (optional)     | Step-specific options and configurations. Content varies depending on the `type`. See [supported types](#supported-type-of-installers) for a comprehensive list of supported values.                                                                                                          |
| `bin_name`         | String (optional)     | Binary name for the installed software, used instead of `name` when checking for app's existence.                                                                                                                                                                                             |
| `check_has_update` | String (shell script) | Shell command to check whether an update is available for the installed software. This will override the default check provided by the corresponding `type`. The check **must succeed** (return exit code 0) if the app has an update, or fail (other status codes) if the app is up to date. |
| `check_installed`  | String (shell script) | Shell command to check if the step has already been installed. If the check succeeds (exits with status 0), it means the app is already installed and can be skipped if not checking for updates.                                                                                             |
| `pre_install`      | String (shell script) | Shell script to execute _before_ the step is installed.                                                                                                                                                                                                                                       |
| `post_install`     | String (shell script) | Shell script to execute _after_ the step is installed.                                                                                                                                                                                                                                        |
| `pre_update`       | String (shell script) | Shell script to execute _before_ the step is updated (if applicable).                                                                                                                                                                                                                         |
| `post_update`      | String (shell script) | Shell script to execute _after_ the step is updated (if applicable).                                                                                                                                                                                                                          |
| `env_shell`        | Object (optional)     | Shell to use for command executions. See `env_shell` subfields below.                                                                                                                                                                                                                         |
| `env_shell.macos`  | String (optional)     | Shell to use for macOS command executions. If not specified, the default shell will be used.                                                                                                                                                                                                  |
| `env_shell.linux`  | String (optional)     | Shell to use for Linux command executions. If not specified, the default shell will be used.                                                                                                                                                                                                  |

### Supported `type` of Installers

For a full list with all the supported options, see [the docs](./docs/installer-types.md).

- **`shell`**

  - Executes arbitrary shell commands.

- **`group`**

  - Executes a logical group of steps in sequence.
  - Allows nesting multiple steps together.

- **`git`**

  - Clones a git repository to a local directory.
  - If `name` is a full git URL (https or SSH), the repository is cloned directly. If it is a
    repository path, e.g. `chenasraf/sofmani`, GitHub is assumed.

- **`manifest`**

  - Installs an entire manifest from a local or remote file.
  - Every entry in the `install` array will be run, similar to how `steps` are run for `group`
    installers.
  - `debug` and `check_updates` will be inherited by the loaded config.
  - `env` and `defaults` will be merged into the loaded config, overriding any existing values.

- **`rsync`**

  - Copy files from `source` to `destination` using rsync.

- **`brew`**

  - Installs packages using Homebrew.

- **`npm`/`pnpm`/`yarn`**

  - Installs packages using npm/pnpm/yarn.
  - Use `type: npm` for `npm install`, `type: pnpm` for `pnpm install`, and `type: yarn` for
    `yarn install`.

- **`apt`/`apk`**

  - Installs packages using apt/apk install.
  - Use `type: apt` for `apt install`, and `type: apk` for `apk add`.

- **`pipx`**

  - Installs packages using pipx.

---

## 📂 Example Workflow

Here’s how you might configure `sofmani` to provision a new system:

1. Create the YAML configuration file `sofmani.yaml`:

   ```yaml
   debug: true
   check_updates: true
   install:
     - name: jq
       type: brew
     - name: yq
       type: shell
       opts:
         command: pipx install yq
   ```

2. Run `sofmani` with your config file:

   ```bash
   sofmani sofmani.yaml
   ```

3. Let `sofmani` handle the installation, configuration syncing, and software updates automatically!
   🎉

---

## 💡 Tips and Tricks

1. Use `platforms.only` and `platforms.except` to fine-tune actions for platform-specific
   environments.
2. Use `check_installed` to skip steps if a specific condition or software is already installed, if
   the check is not a simple binary existence check on `name` or `bin_name`.
3. You can use groups to group together a set of installation steps to run in order (for example,
   for using different installers per-OS), or to simply exclude several steps together if an app is
   already installed with a single check.

---

## 🛠️ Contributing

I am developing this package on my free time, so any support, whether code, issues, or just stars is
very helpful to sustaining its life. If you are feeling incredibly generous and would like to donate
just a small amount to help sustain this project, I would be very very thankful!

<a href='https://ko-fi.com/casraf' target='_blank'>
  <img height='36' style='border:0px;height:36px;'
    src='https://cdn.ko-fi.com/cdn/kofi1.png?v=3'
    alt='Buy Me a Coffee at ko-fi.com' />
</a>

I welcome any issues or pull requests on GitHub. If you find a bug, or would like a new feature,
don't hesitate to open an appropriate issue and I will do my best to reply promptly.

---

## 📜 License

`sofmani` is licensed under the [CC0-1.0 License](/LICENSE).

---

Happy provisioning! 🎉
