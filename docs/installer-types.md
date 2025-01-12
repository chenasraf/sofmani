# Installer Types

The `install` field describes the steps to execute. Each step represents an action or group of
actions. Steps can be of **several types**, such as `brew`, `rsync`, `shell`, and more.

## Fields

These fields are shared by all installer types. Some fields may vary in behavior depending on the
`type`.

- **`name`**

  - **Type**: String (required)
  - **Description**: Identifier for the step. It does not have to be unique, but is usually used to
    check for the app's existence (can be overridden using `bin_name`).

- **`type`**

  - **Type**: String (required)
  - **Description**: Type of the step. See [supported types](#supported-type-of-installers) for a
    comprehensive list of supported values.

- **`platforms`**

  - **Type**: Object (optional)
  - **Description**: Platform-specific execution controls. See `platforms` subfields below.
  - **Subfields**:
    - **`platforms.only`**
      - **Type**: Array of Strings
      - **Description**: Platforms where the step should execute (e.g., `['macos', 'linux']`).
        Supercedes `platforms.except`.
    - **`platforms.except`**
      - **Type**: Array of Strings
      - **Description**: Platforms where the step should **not** execute; replaces `platforms.only`.

- **`steps`**

  - **Type**: Array of Installers
  - **Description**: Sub-steps for `group` type. Allows nesting multiple steps together. Ignored for
    all other types.

- **`opts`**

  - **Type**: Object (optional)
  - **Description**: Step-specific options and configurations. Content varies depending on the
    `type`. See [supported types](#supported-type-of-installers) for a comprehensive list of
    supported values.

- **`bin_name`**

  - **Type**: String (optional)
  - **Description**: Binary name for the installed software, used instead of `name` when checking
    for app's existence.

- **`check_has_update`**

  - **Type**: String (shell script)
  - **Description**: Shell command to check whether an update is available for the installed
    software. This will override the default check provided by the corresponding `type`. The check
    **must succeed** (return exit code 0) if the app has an update, or fail (other status codes) if
    the app is up to date.

- **`check_installed`**

  - **Type**: String (shell script)
  - **Description**: Shell command to check if the step has already been installed. If the check
    succeeds (exits with status 0), it means the app is already installed and can be skipped if not
    checking for updates.

- **`pre_install`**

  - **Type**: String (shell script)
  - **Description**: Shell script to execute _before_ the step is installed.

- **`post_install`**

  - **Type**: String (shell script)
  - **Description**: Shell script to execute _after_ the step is installed.

- **`pre_update`**

  - **Type**: String (shell script)
  - **Description**: Shell script to execute _before_ the step is updated (if applicable).

- **`post_update`**

  - **Type**: String (shell script)
  - **Description**: Shell script to execute _after_ the step is updated (if applicable).

- **`env_shell`**
  - **Type**: Object (optional)
  - **Description**: Shell to use for command executions. See `env_shell` subfields below. Windows
    always uses `cmd`.
  - **Subfields**:
    - **`env_shell.macos`**
      - **Type**: String (optional)
      - **Description**: Shell to use for macOS command executions. If not specified, the default
        shell will be used.
    - **`env_shell.linux`**
      - **Type**: String (optional)
      - **Description**: Shell to use for Linux command executions. If not specified, the default
        shell will be used.

## Supported `type` of Installers

- **`rsync`**

  - **Description**: Copy files from `source` to `destination` using rsync.
  - **Options**:
    - `opts.source`: Source directory/file.
    - `opts.destination`: Destination directory/file.
    - `opts.flags`: Additional rsync flags (e.g., `--delete`, `--exclude`).

- **`group`**

  - **Description**: Executes a logical group of steps in sequence.
    - Allows nesting multiple steps together.
  - **Options**:
    - `steps`: List of nested steps.

- **`brew`**

  - **Description**: Installs packages using Homebrew.
  - **Options**:
    - `opts.tap`: Name of the tap to install the package from.

- **`shell`**

  - **Description**: Executes arbitrary shell commands.
  - **Options**:
    - `opts.command`: The command to execute for installing.
    - `opts.update_command`: The command to execute for updating.

- **`npm`/`pnpm`/`yarn`**

  - **Description**: Installs packages using npm/pnpm/yarn.
    - Use `type: npm` for `npm install`, `type: pnpm` for `pnpm install`, and `type: yarn` for
      `yarn install`.

- **`git`**

  - **Description**: Clones a git repository to a local directory.
    - If `name` is a full git URL (https or SSH), the repository is cloned directly.
    - If it is a repository path, e.g. `chenasraf/sofmani`, GitHub is assumed.
  - **Options**:
    - `opts.destination`: The local directory to clone the repository to.
    - `opts.ref`: The branch, tag, or commit to checkout after cloning.

- **`manifest`**

  - **Description**: Installs an entire manifest from a local or remote file.
    - Every entry in the `install` array will be run, similar to how `steps` are run for `group`
      installers.
    - `debug` and `check_updates` will be inherited by the loaded config.
    - `env` and `defaults` will be merged into the loaded config, overriding any existing values.
  - **Options**:
    - `opts.source`: The local file, or remote git URL (https or SSH) containing the manifest.
    - `opts.path`: The path to the manifest file within the repository. If `opts.source` is a local
      file, `opts.path` will be appended to it.
    - `opts.ref`: The branch, tag, or commit to checkout after cloning if `opts.source` is a git
      URL. For local manifests, this value will be ignored.

- **`apt`**
  - **Description**: Installs packages using apt install.

## Installer Examples

All of these examples should be usable, but don't count on them being maintained. Why not look at
the [Recipes](./recipes)?

### group

```yaml
install:
  - name: pyenv
    type: group
    steps:
      - name: pyenv
        type: brew
        platforms:
          only: ['macos']
      - name: pyenv
        type: shell
        platforms:
          only: ['linux']
        opts:
          command: 'curl https://pyenv.run | bash'
```

### manifest

```yaml
install:
  - name: lazygit
    type: manifest
    opts:
      source: git@github.com:chenasraf/sofmani.git
      path: docs/recipes/lazygit.yml
```

### git

```yaml
install:
  - name: github/gitignore
    type: git
    opts:
      destination: ~/.gitignore-templates
```

### shell

```yaml
install:
  - name: fnm
    type: shell
    post_install: |
      fnm install --lts
      fnm use lts-latest
    opts:
      command: curl -fsSL https://fnm.vercel.app/install | bash
```

### rsync

```yaml
install:
  - name: config
    type: rsync
    opts:
      source: ~/.dotfiles/.config
      destination: ~/.config
```

### brew

```yaml
install:
  - name: sofmani
    type: brew
    opts:
      tap: chenasraf/tap
```

### npm/pnpm/yarn

```yaml
install:
  - name: prettier
    type: pnpm
```

### apt

```yaml
install:
  - name: pipx
    type: apt
    platforms:
      only: ['linux']
```
