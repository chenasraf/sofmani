# Installer Configuration

The `install` field describes the steps to execute. Each step represents an action or group of
actions. Steps can be of **several types**, such as `brew`, `rsync`, `shell`, and more.

## Table of Contents

- [Categories](#categories)
- [Fields](#fields)
- [Template Variables](#template-variables)
- [Supported `type` of Installers](#supported-type-of-installers)
  - [shell](#shell)
  - [group](#group)
  - [git](#git)
  - [github-release](#github-release)
  - [manifest](#manifest)
  - [rsync](#rsync)
  - [brew](#brew)
  - [npm / pnpm / yarn](#npm--pnpm--yarn)
  - [apt / apk](#apt--apk)
  - [pacman / yay](#pacman--yay)
  - [pipx](#pipx)
  - [cargo](#cargo)
  - [docker](#docker)
- [Installer Examples](#installer-examples)
  - [group](#group-1)
  - [Machine-specific installers](#machine-specific-installers)
  - [manifest](#manifest-1)
  - [git](#git-1)
  - [github-release](#github-release-1)
  - [shell](#shell-1)
  - [rsync](#rsync-1)
  - [brew](#brew-1)
  - [npm/pnpm/yarn](#npmpnpmyarn)
  - [apt](#apt)
  - [pacman/yay](#pacmanyay)
  - [cargo](#cargo-1)
  - [docker](#docker-1)

## Categories

You can add **category headers** to visually organize your installers list. Categories are special
entries that display a bordered header in the output but don't perform any installation.

### Fields

- **`category`**
  - **Type**: String (required for category entries)
  - **Description**: The category name to display. When this field is present, the entry is treated
    as a category header, not an installer.

- **`desc`**
  - **Type**: String (optional)
  - **Description**: An optional description shown below the category name. Supports multi-line text
    with automatic word wrapping. Existing line breaks are preserved.

### Example

```yaml
install:
  - category: Development Tools

  - name: neovim
    type: brew

  - name: lazygit
    type: brew

  - category: System Utilities
    desc: Tools for system maintenance and monitoring.

  - name: htop
    type: brew

  - category: Configuration
    desc: |
      These installers sync configuration files from dotfiles.
      They run on every execution to keep configs up to date.

  - name: nvim-config
    type: rsync
    opts:
      source: ~/.dotfiles/.config/nvim
      destination: ~/.config/nvim
```

### Output

The appearance of category headers is controlled by the top-level `category_display` option.

#### `border` (default)

Categories are displayed with a bordered header and spacing before/after:

```
┌──────────────────────────────────────────────────────────┐
│ Development Tools                                        │
└──────────────────────────────────────────────────────────┘
```

With a description:

```
┌──────────────────────────────────────────────────────────┐
│ System Utilities                                         │
├──────────────────────────────────────────────────────────┤
│ Tools for system maintenance and monitoring.             │
└──────────────────────────────────────────────────────────┘
```

#### `border-compact`

Same as `border`, but without the empty lines before and after the box.

#### `minimal`

Categories are displayed as plain text without any border or spacing:

```
Development Tools
```

With a description:

```
System Utilities
Tools for system maintenance and monitoring.
```

The box width (for `border` and `border-compact`) adapts to narrower terminals (minimum of terminal
width or 60 characters).

---

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

- **`enabled`**
  - **Type**: String or Boolean (optional)
  - **Description**: Enable or disable the step. Disabled steps are not run. This can either be a
    static boolean (`true` or `false`), or a command that returns a success status code for true, or
    a failure for false. Commands support [template variables](#template-variables).

- **`tags`**
  - **Type** String (optional)
  - **Description**: Arbitrary tags to attach to an installer. These can later be used to filter
    this installer in or out when running sofmani. This should be a string containing
    space-separated tags.

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

- **`machines`**
  - **Type**: Object (optional)
  - **Description**: Machine-specific execution controls. Use this to run installers only on
    specific machines. Get the machine ID by running `sofmani --machine-id`. You can use either raw
    machine IDs or aliases defined in the top-level `machine_aliases` configuration. See `machines`
    subfields below.
  - **Subfields**:
    - **`machines.only`**
      - **Type**: Array of Strings
      - **Description**: Machine IDs or aliases where the step should execute. Supercedes
        `machines.except`.
    - **`machines.except`**
      - **Type**: Array of Strings
      - **Description**: Machine IDs or aliases where the step should **not** execute.

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
    the app is up to date. Supports [template variables](#template-variables).

- **`check_installed`**
  - **Type**: String (shell script)
  - **Description**: Shell command to check if the step has already been installed. If the check
    succeeds (exits with status 0), it means the app is already installed and can be skipped if not
    checking for updates. Supports [template variables](#template-variables).

- **`pre_install`**
  - **Type**: String (shell script)
  - **Description**: Shell script to execute _before_ the step is installed. Supports
    [template variables](#template-variables).

- **`post_install`**
  - **Type**: String (shell script)
  - **Description**: Shell script to execute _after_ the step is installed. Supports
    [template variables](#template-variables).

- **`pre_update`**
  - **Type**: String (shell script)
  - **Description**: Shell script to execute _before_ the step is updated (if applicable). Supports
    [template variables](#template-variables).

- **`post_update`**
  - **Type**: String (shell script)
  - **Description**: Shell script to execute _after_ the step is updated (if applicable). Supports
    [template variables](#template-variables).

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

- **`verbose`**
  - **Type**: Boolean (optional)
  - **Description**: Enable verbose output for the installer's native commands. When set to `true`,
    the installer will pass verbose flags to the underlying tool, producing more detailed output
    during installation and updates. The specific flag used depends on the installer type. Can also
    be set via [defaults](#) per installer type.
  - **Default**: `false` (not set)
  - **Verbose flags per installer type**:

    | Type                | Verbose flag |
    | ------------------- | ------------ |
    | `rsync`             | `-v`         |
    | `brew`              | `--verbose`  |
    | `git`               | `--verbose`  |
    | `npm`/`pnpm`/`yarn` | `--verbose`  |
    | `pipx`              | `--verbose`  |
    | `cargo`             | `--verbose`  |
    | `pacman`/`yay`      | `--verbose`  |
    | `apk`               | `--verbose`  |
    | `apt`               | _(no-op)_    |
    | `docker`            | _(no-op)_    |
    | `shell`             | _(no-op)_    |
    | `github-release`    | _(no-op)_    |
    | `manifest`          | _(no-op)_    |
    | `group`             | _(no-op)_    |

  - **Examples**:

    ```yaml
    # Enable verbose for a single installer
    - name: xdg-config
      type: rsync
      verbose: true
      opts:
        source: ~/.dotfiles/.config
        destination: ~/.config

    # Enable verbose for all brew installers via defaults
    defaults:
      type:
        brew:
          verbose: true
    ```

- **`frequency`**
  - **Type**: String (optional)
  - **Description**: Limits how often the installer runs. After a successful install or update, the
    next run will be skipped until the specified duration has elapsed. The timestamp of the last
    successful run is stored in the sofmani cache directory.
  - **Format**: A prettified duration string. Multiple components can be combined. Supported units:
    - `s` — seconds (e.g., `60s`)
    - `m` — minutes (e.g., `30m`)
    - `h` — hours (e.g., `12h`)
    - `d` — days (e.g., `1d`)
    - `w` — weeks (e.g., `1w`)
    - Combined: `1d12h`, `1w2d`
  - **Default**: Not set (installer runs every time).
  - **Note**: Use the `--ignore-frequency` CLI flag to bypass frequency checks for all installers.
  - **Examples**:

    ```yaml
    # Only check for updates once a day
    - name: neovim
      type: brew
      frequency: 1d

    # Only run once a week
    - name: sync-dotfiles
      type: rsync
      frequency: 1w
      opts:
        source: ~/.dotfiles/.config
        destination: ~/.config

    # Run at most every 12 hours
    - name: my-tool
      type: shell
      frequency: 12h
      opts:
        command: ./install.sh
    ```

- **`skip_summary`**
  - **Type**: Boolean or Object (optional)
  - **Description**: Exclude this installer from the installation summary. Useful for installers
    that always run (like config sync scripts) and would clutter the summary output.
  - **Values**:
    - **Boolean**: When set to `true`, the installer is excluded from both install and update
      summaries. When set to `false` (default), the installer appears in summaries normally.
    - **Object**: For granular control, specify which summaries to skip:
      - **`skip_summary.install`**: Boolean - exclude from the "Installed" section of the summary.
      - **`skip_summary.update`**: Boolean - exclude from the "Upgraded" section of the summary.
  - **Examples**:

    ```yaml
    # Skip from both install and update summaries
    - name: sync-dotfiles
      type: rsync
      skip_summary: true
      opts:
        source: ~/.dotfiles/.config
        destination: ~/.config

    # Skip only from install summary (still shows in upgrade summary)
    - name: config-setup
      type: shell
      skip_summary:
        install: true
      opts:
        command: ./setup.sh

    # Skip only from upgrade summary (still shows in install summary)
    - name: my-tool
      type: brew
      skip_summary:
        update: true
    ```

## Template Variables

All shell commands across installers support **Go template syntax** for dynamic value insertion.
This includes `opts.command`, `opts.update_command`, `pre_install`, `post_install`, `pre_update`,
`post_update`, `check_installed`, `check_has_update`, and `enabled` (when it's a shell command).

Available variables:

| Variable               | Description                                                                          | Example                     |
| ---------------------- | ------------------------------------------------------------------------------------ | --------------------------- |
| `{{ .Arch }}`          | System architecture in Go format                                                     | `amd64`, `arm64`            |
| `{{ .ArchAlias }}`     | Architecture in common alias format                                                  | `x86_64`, `arm64`           |
| `{{ .ArchGnu }}`       | Architecture in GNU/Linux format                                                     | `x86_64`, `aarch64`         |
| `{{ .OS }}`            | Current operating system                                                             | `macos`, `linux`, `windows` |
| `{{ .DeviceID }}`      | Unique machine identifier (truncated SHA-256 hash)                                   | `5fa2a8e8193868df`          |
| `{{ .DeviceIDAlias }}` | Friendly alias for the current machine, if defined in `machine_aliases`              | `work-laptop`               |
| `{{ .Tag }}`           | Full tag name (only available in `github-release` `download_filename`)               | `v1.0.0`                    |
| `{{ .Version }}`       | Version without leading "v" (only available in `github-release` `download_filename`) | `1.0.0`                     |
| `{{ .DownloadFile }}`  | Absolute path to the downloaded asset (only in `github-release` `extract_command`)   | `/tmp/sofmani.../app.download` |
| `{{ .ExtractDir }}`    | Temp directory to extract into (only in `github-release` `extract_command`)          | `/tmp/sofmani...`           |
| `{{ .Destination }}`   | Final destination directory (only in `github-release` `extract_command`)             | `~/.local/bin`              |
| `{{ .BinName }}`       | Expected output binary name (only in `github-release` `extract_command`)             | `my-tool`                   |
| `{{ .ArchiveBinName }}`| Filename sofmani copies from `ExtractDir` → `Destination` (only in `extract_command`)| `my-tool`                   |

In addition, `DEVICE_ID` and `DEVICE_ID_ALIAS` are injected as **environment variables** into all
command executions, so they can also be referenced as `$DEVICE_ID` and `$DEVICE_ID_ALIAS` in shell
commands.

### Example

```yaml
machine_aliases:
  work-laptop: 5fa2a8e8193868df
  home-desktop: a1b2c3d4e5f67890

install:
  - name: sync-config
    type: shell
    opts:
      command: cp ~/dotfiles/config-{{ .DeviceIDAlias }}.yaml ~/.config/myapp/config.yaml

  - name: setup-tool
    type: shell
    opts:
      command: |
        echo "Setting up on device $DEVICE_ID ({{ .DeviceIDAlias }})"
        ./setup.sh --arch {{ .Arch }} --os {{ .OS }}
```

## Supported `type` of Installers

### `shell`

Executes arbitrary shell commands. Commands support [template variables](#template-variables).

**Options**:

- `opts.command`: The command to execute for installing.
- `opts.update_command`: The command to execute for updating.

### `group`

Executes a logical group of steps in sequence. Allows nesting multiple steps together.

**Options**:

- `steps`: List of nested steps.

### `git`

Clones a git repository to a local directory.

- If `name` is a full git URL (https or SSH), the repository is cloned directly.
- If it is a repository path, e.g. `chenasraf/sofmani`, GitHub is assumed.

**Options**:

- `opts.destination`: The local directory to clone the repository to.
- `opts.ref`: The branch, tag, or commit to checkout after cloning.
- `opts.flags`: Additional flags to pass to git commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only to `git clone`.
- `opts.update_flags`: Additional flags to pass only to `git pull`.

### `github-release`

Downloads a GitHub release asset. Optionally untar/unzip the downloaded file.

**Options**:

- `opts.repository`: The repository to download from. Should be in the format:
  `user/repository-name`
- `opts.destination`: The target directory to extract the files to.
- `opts.strategy`: The download strategy. Can be one of: `tar`, `zip`, `gzip`, `custom`, `none`
  (default)
  - `none` - the release file is not compressed, and should be copied directly
  - `tar` - the release file is a tar file, and should be extracted
  - `zip` - the release file is a zip file, and should be extracted
  - `gzip` (alias: `gz`) - the release file is a single gzip-compressed file (not a tar
    archive). It is decompressed with Go's `compress/gzip` and written to
    `destination/bin_name` with the executable bit set. Use this for projects like
    `tree-sitter` that publish each binary as a plain `.gz` file:

    ```yaml
    - name: tree-sitter
      type: github-release
      opts:
        repository: tree-sitter/tree-sitter
        destination: ~/.local/bin
        strategy: gzip
        download_filename: tree-sitter-{{ .OS }}-{{ .ArchAlias }}.gz
    ```

  - `custom` - run a user-provided shell hook (`opts.extract_command`) to extract the
    downloaded asset yourself. After the command finishes, sofmani copies
    `{{ .ExtractDir }}/{{ .ArchiveBinName }}` to `{{ .Destination }}/{{ .BinName }}` and
    sets the executable bit — exactly like the `tar` and `zip` strategies do. Use this
    for unusual archive formats (7-Zip, xz, self-extracting installers, ...).

- `opts.extract_command`: The shell command to run when `strategy: custom`. It goes through
  the same Go template substitution as other sofmani shell hooks, with extra variables
  specific to the extract context:

  | Variable               | Description                                                             |
  | ---------------------- | ----------------------------------------------------------------------- |
  | `{{ .DownloadFile }}`  | Absolute path to the downloaded asset                                   |
  | `{{ .ExtractDir }}`    | Temp directory — your command should place extracted files here        |
  | `{{ .Destination }}`   | Final destination directory (from `opts.destination`)                   |
  | `{{ .BinName }}`       | The expected output binary name (`bin_name` or installer name)          |
  | `{{ .ArchiveBinName }}`| Filename sofmani will copy from `ExtractDir` → `Destination` afterwards |

  All the usual template variables (`{{ .OS }}`, `{{ .Arch }}`, `{{ .Tag }}`, ...) are also
  available. `extract_command` is required when `strategy: custom`, and is not allowed
  with any other strategy.

  Example — extracting a `.tar.xz` asset by shelling out to `tar`:

  ```yaml
  - name: my-tool
    type: github-release
    opts:
      repository: example/my-tool
      destination: ~/.local/bin
      strategy: custom
      download_filename: my-tool-{{ .Version }}-{{ .OS }}.tar.xz
      extract_command: tar -xJf {{ .DownloadFile }} -C {{ .ExtractDir }}
  ```

  Example — extracting a 7-Zip asset:

  ```yaml
  - name: weird-tool
    type: github-release
    opts:
      repository: example/weird-tool
      destination: ~/.local/bin
      strategy: custom
      download_filename: weird-tool-{{ .Version }}.7z
      extract_command: 7z x {{ .DownloadFile }} -o{{ .ExtractDir }}
  ```
- `opts.download_filename`: The filename of the release asset to download.

  This should either be a string, or a map of platforms to filenames.

  You can use Go template syntax to insert dynamic values into the filename:
  - `{{ .Tag }}` - the full tag name, e.g. `v1.0.0`
  - `{{ .Version }}` - the version without the leading "v", e.g. `1.0.0`
  - `{{ .Arch }}` - the system architecture in Go format, e.g. `amd64`, `arm64`
  - `{{ .ArchAlias }}` - the architecture in common alias format, e.g. `x86_64`, `arm64`
  - `{{ .ArchGnu }}` - the architecture in GNU/Linux format, e.g. `x86_64`, `aarch64`
  - `{{ .OS }}` - the current operating system, e.g. `macos`, `linux`, `windows`
  - `{{ .DeviceID }}` - the unique machine identifier (truncated SHA-256 hash)
  - `{{ .DeviceIDAlias }}` - the friendly alias for the current machine, if defined in
    `machine_aliases`

  **Legacy syntax (deprecated):** The old `{tag}`, `{version}`, `{arch}`, `{arch_alias}`,
  `{arch_gnu}`, `{os}`, `{device_id}`, and `{device_id_alias}` tokens are still supported but
  deprecated. A deprecation warning will be logged at DEBUG level when they are used.

  Examples:

  ```yaml
  # Using Go template syntax (recommended)
  download_filename: myapp_{{ .Tag }}_linux_{{ .ArchAlias }}.tar.gz # outputs: myapp_v1.0.0_linux_x86_64.tar.gz
  download_filename: myapp_{{ .Version }}_{{ .OS }}.tar.gz # outputs: myapp_1.0.0_linux.tar.gz

  # Platform-specific filenames
  download_filename:
    macos: myapp_{{ .Tag }}_darwin_{{ .ArchAlias }}.tar.gz
    linux: myapp_{{ .Tag }}_linux_{{ .ArchAlias }}.tar.gz
    windows: myapp_{{ .Tag }}_windows_{{ .ArchAlias }}.zip

  # Legacy syntax (deprecated, still works)
  download_filename: myapp_{tag}_linux.tar.gz # outputs: myapp_v1.0.0_linux.tar.gz
  ```

- `opts.archive_bin_name`: The name of the binary file inside the archive (tar/zip). Use this when
  the filename inside the archive differs from the desired output `bin_name`. If not set, falls back
  to `bin_name` (or the installer name).

  ```yaml
  - name: cospend-cli
    bin_name: cospend
    type: github-release
    opts:
      repository: chenasraf/cospend-cli
      destination: ~/.local/bin
      strategy: tar
      download_filename: cospend-cli-linux-{{ .Arch }}.tar.gz
      archive_bin_name: cospend-cli # file inside the tar is "cospend-cli", output will be "cospend"
  ```

- `opts.extract_to`: Enables **tree mode**. When set, the full archive contents are extracted into
  this directory, preserving sibling files (`lib/`, `share/`, `libexec/`, etc.). Use this for
  toolchains that ship as a pre-built directory tree where the binary resolves paths relative to its
  own location (Neovim, Go, Node, Flutter, Zig, many language servers). In tree mode, `destination`
  and `archive_bin_name` are ignored; use `bin_links` to expose binaries on your `$PATH`. Requires
  `strategy` to be `tar` or `zip`.

- `opts.strip_components`: Drops this many leading path components from each archive entry,
  equivalent to `tar --strip-components=N`. Release tarballs almost always wrap their contents in a
  single versioned directory (e.g. `nvim-linux-x86_64/`), so `strip_components: 1` is the common
  value. Only meaningful with `extract_to`.

- `opts.bin_links`: A list of binaries to expose from inside the extracted tree. Each entry has a
  `source` (relative to `extract_to`, or an absolute path) and a required `target` (absolute path
  where the symlink is placed). On unix, each entry becomes a symlink, which is essential so the
  binary resolves its sibling files via its real location. On Windows, where creating symlinks
  requires elevated privileges, the file is copied instead. Only meaningful with `extract_to`.

  Example — installing Neovim as a full tree with `nvim` on `$PATH`:

  ```yaml
  - name: neovim
    type: github-release
    platforms: { only: ['linux'] }
    opts:
      repository: neovim/neovim
      strategy: tar
      download_filename: nvim-linux-{{ .Arch }}.tar.gz
      extract_to: ~/.local/share/neovim
      strip_components: 1
      bin_links:
        - source: bin/nvim
          target: ~/.local/bin/nvim
  ```

  On update, the extracted tree is replaced atomically (extracted to a sibling staging directory and
  renamed into place), so files removed in a new release do not linger from the old version.

- `opts.github_token`: GitHub personal access token for authenticated API requests. Authenticated
  requests have a much higher rate limit (5,000/hour vs 60/hour for unauthenticated).

  Supports environment variable expansion, so you don't need to hard-code credentials:

  ```yaml
  # Using environment variables (recommended)
  github_token: $GITHUB_TOKEN
  github_token: ${GITHUB_TOKEN}

  # Can also be set as a default for all github-release installers
  defaults:
    type:
      github-release:
        opts:
          github_token: $GITHUB_TOKEN
  ```

### `manifest`

Installs an entire manifest from a local or remote file.

- Every entry in the `install` array will be run, similar to how `steps` are run for `group`
  installers.
- `debug` and `check_updates` will be inherited by the loaded config.
- `env` and `defaults` will be merged into the loaded config, overriding any existing values.
- Remote manifests are fetched directly via HTTP (no git clone required).

**Options**:

- `opts.source`: The source of the manifest file. Supports:
  - Local file paths (e.g., `~/.dotfiles/manifest.yml`)
  - Git repository URLs (SSH or HTTPS) - GitHub, GitLab, Bitbucket, and self-hosted instances
  - Raw HTTP URLs (e.g., `https://raw.githubusercontent.com/user/repo/master/manifest.yml`)
- `opts.path`: The path to the manifest file within the repository. Required for git URLs, optional
  for local files (will be appended to source). Ignored for raw HTTP URLs.
- `opts.ref`: The branch, tag, or commit to use if `opts.source` is a git URL. Defaults to `master`.
  Ignored for local files and raw HTTP URLs.

### `rsync`

Copy files from `source` to `destination` using rsync.

**Options**:

- `opts.source`: Source directory/file.
- `opts.destination`: Destination directory/file.
- `opts.flags`: Additional rsync flags (e.g., `--delete`, `--exclude`).

### `brew`

Installs packages using Homebrew.

**Repo update**: Brew auto-updates its index on each command. By default, sofmani lets the first
brew command auto-update normally and suppresses it for subsequent ones (`once` mode). Configure via
the top-level [`repo_update`](./configuration-reference.md#global-options) option.

**Options**:

- `opts.tap`: Name of the tap to install the package from.
- `opts.cask`: Install as a cask instead of a formula.
- `opts.flags`: Additional flags to pass to brew commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only to `brew install`.
- `opts.update_flags`: Additional flags to pass only to `brew upgrade`.

### `npm` / `pnpm` / `yarn`

Installs packages using npm/pnpm/yarn.

- Use `type: npm` for `npm install`, `type: pnpm` for `pnpm install`, and `type: yarn` for
  `yarn install`.

**Options**:

- `opts.flags`: Additional flags to pass to commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only during install.
- `opts.update_flags`: Additional flags to pass only during update.

### `apt` / `apk`

Installs packages using apt install or apk add.

- Use `type: apt` for `apt install`, and `type: apk` for `apk add`.

**Repo update**: Runs `apt update` or `apk update` before installing. By default, the update runs at
most once per sofmani run (`once` mode). Configure via the top-level
[`repo_update`](./configuration-reference.md#global-options) option.

**Options**:

- `opts.flags`: Additional flags to pass to commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only during install.
- `opts.update_flags`: Additional flags to pass only during update.

### `pacman` / `yay`

Installs packages using pacman or yay (Arch Linux).

- Use `type: pacman` for official Arch repository packages.
- Use `type: yay` for AUR (Arch User Repository) packages.
- Both use `--noconfirm` for non-interactive installation.

**Options**:

- `opts.needed`: Skip reinstalling up-to-date packages (`--needed` flag).
- `opts.flags`: Additional flags to pass to commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only during install.
- `opts.update_flags`: Additional flags to pass only during update.

### `pipx`

Installs packages using pipx.

**Options**:

- `opts.flags`: Additional flags to pass to commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only to `pipx install`.
- `opts.update_flags`: Additional flags to pass only to `pipx upgrade`.

### `cargo`

Installs packages using Rust's cargo. Uses `cargo install` for both installation and updates.
`cargo install` will automatically skip rebuilding if the package is already up-to-date.

**Options**:

- `opts.flags`: Additional flags to pass to commands (fallback for install/update).
- `opts.install_flags`: Additional flags to pass only during install.
- `opts.update_flags`: Additional flags to pass only during update.

### `docker`

Pulls and runs Docker containers using `docker run`. Also supports update checks by comparing image
digests.

- The image is pulled from the registry (e.g., Docker Hub or GHCR) and started with the provided
  options.
- If the container already exists, it will be started instead of run again.
- Updates are detected by comparing the image digest before and after a pull.
- The container is always run with `--restart always -d`, unless overridden in a custom shell.

**Required**:

- `name`: The full Docker image name, including tag (e.g., `ghcr.io/open-webui/open-webui:main`).
- `bin_name`: The container name to assign to the running instance (used in install and update
  checks).

**Options**:

- `opts.flags`: A string of flags to pass to `docker run` (e.g., ports, volumes, extra args). These
  are appended after the default flags and before the image name.

  Example:

  ```yaml
  opts:
    flags: >
      -p 3300:8080 -v data-volume:/app/data --add-host=host.docker.internal:host-gateway
  ```

- `opts.platform`: Override the platform used when checking the image manifest for updates. Accepts
  a per-OS map with values in `os/arch` format (e.g., `linux/amd64`).

  This is useful if you're running on a platform like `darwin/arm64`, but want to compare digests
  for a different image target (e.g., `linux/amd64`).

  Example:

  ```yaml
  opts:
    platform:
      macos: linux/amd64
      linux: linux/amd64
  ```

- `opts.skip_if_unavailable`: Whether to skip the installation/update if the Docker daemon is not
  running. Defaults to false (so it will fail the installer)

## Installer Examples

All of these examples should be usable, but don't count on them being maintained. Why not look at
the [Recipes](./recipes)?

### group

```yaml
install:
  - name: pyenv
    type: group
    tags: python
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

### Machine-specific installers

```yaml
# Define friendly names for your machines (get IDs with `sofmani --machine-id`)
machine_aliases:
  work-laptop: a1b2c3d4e5f67890
  home-desktop: 5fa2a8e8193868df
  home-server: fedcba0987654321

install:
  # Only install on specific machines using aliases
  - name: work-tools
    type: group
    machines:
      only: ['work-laptop']
    steps:
      - name: slack
        type: brew
        opts:
          cask: true
      - name: zoom
        type: brew
        opts:
          cask: true

  # Install everywhere except the home server
  - name: desktop-apps
    type: group
    machines:
      except: ['home-server']
    steps:
      - name: firefox
        type: brew
        opts:
          cask: true

  # You can also use raw machine IDs directly
  - name: special-tool
    type: brew
    machines:
      only: ['a1b2c3d4e5f67890'] # Raw machine ID also works
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

### github-release

```yaml
install:
  - name: lazygit
    type: github-release
    opts:
      repository: jesseduffield/lazygit
      strategy: tar
      destination: /usr/local/bin
      download_filename: lazygit_{{ .Version }}_Linux_{{ .ArchAlias }}.tar.gz
      github_token: $GITHUB_TOKEN # optional, for higher rate limits
```

### shell

```yaml
install:
  - name: fnm
    type: shell
    tags: node
    post_install: |
      fnm install --lts
      fnm use lts-latest
    opts:
      command: curl -fsSL https://fnm.vercel.app/install | bash
```

### rsync

```yaml
install:
  - name: xdg-config
    type: rsync
    tags: config
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
    tags: node
```

### apt

```yaml
install:
  - name: pipx
    type: apt
    tags: python
    platforms:
      only: ['linux']
```

### pacman/yay

```yaml
install:
  # Install from official Arch repositories
  - name: neovim
    type: pacman
    bin_name: nvim
    opts:
      needed: true # Skip if already up-to-date

  # Install from AUR using yay
  - name: visual-studio-code-bin
    type: yay
    bin_name: code
```

### cargo

```yaml
install:
  - name: ripgrep
    type: cargo
    bin_name: rg
```

### docker

```yaml
- name: ghcr.io/open-webui/open-webui:main
  bin_name: open-webui
  type: docker
  opts:
    flags: >
      -p 3300:8080 --add-host=host.docker.internal:host-gateway -v open-webui:/app/backend/data
```
