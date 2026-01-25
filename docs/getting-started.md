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
