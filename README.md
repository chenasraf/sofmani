# sofmani

**`sofmani`** stands for **[Sof]tware [Mani]fest**. It is a robust and flexible provisioning tool
written in **Go**, designed to simplify software installations, configuration syncing, and system
provisioning for both personal and work computers. With a single config file, `sofmani` automates
locating, installing, or updating software and configurations, making system setup quick and
reproducible.

![Downloads](https://img.shields.io/github/downloads/chenasraf/sofmani/total?style=flat-square)
![Go Version](https://img.shields.io/github/go-mod/go-version/chenasraf/sofmani)
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

That's it! You're now ready to use `sofmani`.

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

### Command-Line Flags

The following flags are supported to customize behavior:

| Flag                | Description                        |
| ------------------- | ---------------------------------- |
| `-d`, `--debug`     | Enable debug mode.                 |
| `-D`, `--no-debug`  | Disable debug mode (default).      |
| `-u`, `--update`    | Enable update checking.            |
| `-U`, `--no-update` | Disable update checking (default). |
| `-h`, `--help`      | Display help information and exit. |

If a configuration file is not explicitly provided, `sofmani` attempts to locate one automatically
in the current directory.

If a configuration file argument is not present, sofmani will try to find a `sofmani.yaml` or
`sofmani.json` in the following directories, ordered by priority:

1. Current directory
1. `$HOME/.config` directory
1. Home directory

If no file is found, sofmani will fail to start.

---

## 📚 Configuration Reference

Here is a detailed breakdown of all configuration options:

### Global Options

| Field           | Type    | Description                                                                                              |
| --------------- | ------- | -------------------------------------------------------------------------------------------------------- |
| `debug`         | Boolean | Enable or disable debug mode. Default: `false`.                                                          |
| `check_updates` | Boolean | Enable or disable checking for updates before running operations. Default: `false`.                      |
| `defaults`      | Object  | Defaults to apply to all installer types, such as specifying supported platforms or commonly used flags. |

### `install` Node

The `install` field describes the steps to execute. Each step represents an action or group of
actions. Steps can be of **several types**, such as `brew`, `rsync`, `shell`, and more.

| Field              | Type                  | Description                                                                                                                                                                                                                       |
| ------------------ | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`             | String                | Identifier for the step. It does not have to be unique, but is usually used to check for the app's existence, if applicable (can be overridden using `bin_name`)                                                                  |
| `type`             | String (required)     | Type of the step. Currently supported: `group`, `brew`, `apt`, `rsync`, `shell`, `npm`/`pnpm`/`yarn`.                                                                                                                             |
| `platforms`        | Object (optional)     | Platform-specific execution controls. See `platforms` subfields below.                                                                                                                                                            |
| `platforms.only`   | Array of Strings      | Platforms where the step should execute (e.g., `['macos', 'linux']`). Supercedes `platforms.except`.                                                                                                                              |
| `platforms.except` | Array of Strings      | Platforms where the step should **not** execute; replaces `platforms.only`.                                                                                                                                                       |
| `steps`            | Array of Installers   | Sub-steps for `group` type. Allows nesting multiple steps together.                                                                                                                                                               |
| `opts`             | Object (optional)     | Step-specific options and configurations. Content varies depending on the `type`.                                                                                                                                                 |
| `bin_name`         | String (optional)     | Binary name for the installed software, used instead of `name` when checking for app's existence.                                                                                                                                 |
| `check_has_update` | String (shell script) | Shell command to check whether an update is available for the installed software. This will override the default binary name check for `name` or `bin_name`. The check **must succeed** (return exit code 0) to prompt an update. |
| `check_installed`  | String (shell script) | Shell command to check if the step has already been installed. Skips the install step if the check succeeds.                                                                                                                      |
| `pre_install`      | String (shell script) | Shell script to execute _before_ the step is installed.                                                                                                                                                                           |
| `post_install`     | String (shell script) | Shell script to execute _after_ the step is installed.                                                                                                                                                                            |
| `pre_update`       | String (shell script) | Shell script to execute _before_ the step is updated (if applicable).                                                                                                                                                             |
| `post_update`      | String (shell script) | Shell script to execute _after_ the step is updated (if applicable).                                                                                                                                                              |
| `env_shell`        | Object (optional)     | Shell to use for command executions. See `env_shell` subfields below.                                                                                                                                                             |
| `env_shell.macos`  | String (optional)     | Shell to use for macOS command executions. If not specified, the default shell will be used.                                                                                                                                      |
| `env_shell.linux`  | String (optional)     | Shell to use for Linux command executions. If not specified, the default shell will be used.                                                                                                                                      |

### Supported `type` Installers

1. **`rsync`**

   - Copy files from `source` to `destination` using rsync.
   - **Options**:
     - `opts.source`: Source directory/file.
     - `opts.destination`: Destination directory/file.
     - `opts.flags`: Additional rsync flags (e.g., `--delete`, `--exclude`).

2. **`group`**

   - Executes a logical group of steps in sequence.
   - Allows nesting multiple steps together.
   - **Options**:
     - `steps`: List of nested steps.

3. **`brew`**

   - Installs packages using Homebrew.
   - **Options**:
     - `opts.tap`: Name of the tap to install the package from.

4. **`shell`**

   - Executes arbitrary shell commands.
   - **Options**:
     - `opts.command`: The command to execute for installing.
     - `opts.update_command`: The command to execute for updating.

5. **npm/pnpm/yarn**

   - Installs packages using npm/pnpm/yarn.
   - Use `type: npm` for `npm install`, `type: pnpm` for `pnpm install`, and `type: yarn` for
     `yarn install`.

6. **`apt`**

   - Installs packages using apt install.

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

`sofmani` is licensed under the
[CC0-1.0 License](https://github.com/chenasraf/sofmani/blob/main/LICENSE).

---

Happy provisioning! 🎉
