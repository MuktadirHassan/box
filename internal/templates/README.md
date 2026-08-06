# Environment templates

Each template is a self-contained directory that Box embeds into the CLI binary.

```text
<template-name>/
└── ubuntu/
    ├── template.toml
    ├── Containerfile
    ├── initialize-home
    └── dotfiles/
```

`template.toml` supplies a name, version, description, and base-image family. The
`Containerfile` receives the chosen base image through the `BASE_IMAGE` build
argument. Its `initialize-home` entrypoint must copy dotfiles non-destructively:
persistent home volumes survive runtime recreation and remain owned by the user.

To add a built-in template, add a directory with this layout. The setup interface
discovers it automatically; no template registry or CLI changes are required.
