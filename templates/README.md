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
The `Containerfile` receives the chosen base image, configured user, numeric
UID/GID, shell, prompt, and revision as build arguments. The catalog port
separates consumers from this embedded adapter, so future local or downloaded
providers can be registered without UI or backend changes.

The default `ubuntu-24.04-terminal-tools` template builds a writable development
image with baseline tools and passwordless in-container sudo. The runtime stays
rootless and starts as the configured non-root user.
