# Environment templates

Each template is a self-contained directory under the repository-root `templates/` directory. Box embeds these assets into the CLI binary through a small root-level asset package; catalog behavior remains in `internal/templates`.

```text
<template-id>/
├── template.toml
├── Containerfile
├── initialize-home
└── dotfiles/
```

`template.toml` supplies the canonical ID, manifest version, display metadata,
image compatibility, and supported shells/prompts. IDs are opaque and immutable.
The `Containerfile` receives the chosen base image through `BASE_IMAGE`.
The catalog port separates consumers from this embedded adapter, so future local
or downloaded providers can be registered without UI or backend changes.

Built-in IDs currently include `ubuntu-24.04-terminal-tools`; the legacy
`terminal-tools` alias remains accepted and always means Ubuntu 24.04. Explicit
setup updates persist the canonical ID only after successful setup.
