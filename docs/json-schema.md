# JSON Schema

`sofmani` ships a [JSON Schema](https://json-schema.org/) describing the full configuration format.
Pointing your editor at it gives you **autocompletion**, **inline documentation**, and
**validation** while editing your `sofmani.yaml` or `sofmani.json` files.

The schema lives in the repo at [`schema/sofmani.schema.json`](../schema/sofmani.schema.json) and is
published on the `master` branch at:

```
https://raw.githubusercontent.com/chenasraf/sofmani/master/schema/sofmani.schema.json
```

## Using the schema with YAML

Most editors use the
[YAML Language Server](https://github.com/redhat-developer/yaml-language-server) (bundled with VS
Code's Red Hat YAML extension, Neovim's `yamlls`, Zed, and others). You can enable the schema in
either of two ways.

### 1. Inline comment (per file)

Add a modeline as the **first line** of the YAML file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/chenasraf/sofmani/master/schema/sofmani.schema.json

debug: true
install:
  - name: neovim
    type: brew
```

### 2. Editor-wide association

In VS Code, add this to your `settings.json`:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/chenasraf/sofmani/master/schema/sofmani.schema.json": [
      "sofmani.yaml",
      "sofmani.yml",
      "**/sofmani/*.yml",
      "**/recipes/*.yml"
    ]
  }
}
```

Adjust the glob patterns to match where you keep your manifests.

### Using a local copy

If you have `sofmani` checked out locally, or you vendor the schema, point at the file on disk:

```yaml
# yaml-language-server: $schema=./schema/sofmani.schema.json
```

## Using the schema with JSON

In a JSON config, set the `$schema` key at the top of the document:

```json
{
  "$schema": "https://raw.githubusercontent.com/chenasraf/sofmani/master/schema/sofmani.schema.json",
  "debug": true,
  "install": [
    {
      "name": "neovim",
      "type": "brew"
    }
  ]
}
```

VS Code picks this up automatically. Most other JSON-aware editors do too.

Alternatively, you can configure `json.schemas` in VS Code's `settings.json`:

```json
{
  "json.schemas": [
    {
      "fileMatch": ["sofmani.json", "**/sofmani/*.json"],
      "url": "https://raw.githubusercontent.com/chenasraf/sofmani/master/schema/sofmani.schema.json"
    }
  ]
}
```

## Validating from the command line

You can validate a config file against the schema using any JSON Schema validator. Two convenient
options:

### `check-jsonschema` (Python)

```bash
pipx install check-jsonschema
check-jsonschema \
  --schemafile schema/sofmani.schema.json \
  sofmani.yaml
```

`check-jsonschema` supports both JSON and YAML input files out of the box.

### `ajv` (Node)

For JSON files:

```bash
npx ajv-cli validate \
  -s schema/sofmani.schema.json \
  -d sofmani.json
```

For YAML files, convert on the fly with `yq`:

```bash
yq -o=json sofmani.yaml | npx ajv-cli validate -s schema/sofmani.schema.json -d /dev/stdin
```

## What the schema covers

- All top-level options (`debug`, `check_updates`, `summary`, `category_display`, `repo_update`,
  `defaults`, `env`, `platform_env`, `machine_aliases`, `install`).
- All supported installer types and their type-specific `opts`.
- Enums for `category_display`, `repo_update` modes, installer `type`, and platform names.
- The `frequency` duration pattern (`1d`, `12h`, `1w2d`, ...).
- Dual shapes for fields like `skip_summary` (bool or object), `enabled` (bool or shell string), and
  `github-release` `download_filename` (string or per-platform map).
- Per-type narrowing of `opts`: typos like `tap: foo` vs. `tapp: foo` are flagged, and `group`
  installers cannot accidentally set `opts`.
- The shell-script fields (`check_has_update`, `check_installed`, `pre_install`, `post_install`,
  `pre_update`, `post_update`) accept either a string or a boolean. Booleans are a shorthand that
  YAML coerces to the literal `"true"`/`"false"` — handy for forcing `check_has_update: true` to
  mean "always treat as having an update".
