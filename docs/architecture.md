# Box architecture

## Goal

Box is a Linux CLI for persistent development environments.

The CLI stays the same regardless of how a Box runs:

```text
box create <name>
box setup <name>
box enter <name>
box exec <name> -- <command>
box stop <name>
box delete <name>
```

Box owns the user experience, saved configuration, confirmations, and lifecycle.
Users should not need to use a backend's commands directly.

## Template catalogs

Templates are selected by opaque canonical IDs such as
`ubuntu-24.04-terminal-tools`. The `templates.Catalog` port lists descriptors
and resolves validated templates; the embedded-files adapter owns manifests,
flat filesystem layout, and build-context materialization. Podman, CLI, and
terminal presentation receive the port by constructor injection and do not
depend on embedded assets or image-specific UI rules. Manifest compatibility
pins an image family and release (including digest-qualified references), while
manifest version is schema compatibility and `TemplateRevision` controls home
refreshes.

Built-in template assets live under the repository-root `templates/<template-id>/` directories, separate from Go source. The root `templates` package owns only the `go:embed` filesystem; `internal/templates` receives that filesystem through `NewEmbeddedCatalog`, preserving the catalog port without a parent-path embed or import cycle. Adding a top-level template directory is therefore automatically included by the root package's `all:` embed pattern.

## Model

A Box has three parts:

- **Definition**: the saved configuration: backend, base environment, mounts,
  environment, limits, network policy, and enabled integrations.
- **Runtime**: the replaceable backend-specific instance created from the
  definition. A Podman runtime has a writable root filesystem, so packages and
  system customization survive stop/start operations.
- **Persistent state**: the home, caches, and backend metadata that should
  survive a runtime recreation when possible. Runtime root-filesystem changes
  are intentionally discarded when the runtime is recreated.

The definition is human-readable, versioned, and independent of an existing
runtime.

```text
~/.local/share/box/
└── boxes/<name>/
    ├── box.toml       # Definition
    └── metadata.json  # Backend identity and lifecycle metadata
```

Backend-owned data, such as Podman volumes or Lima VM disks, remains managed by
the backend.

## Backends

Box supports backends behind one internal interface. The initial backends are:

- **Podman**: writable rootless containers; lightweight and suitable for
  trusted local development. The configured user can use passwordless `sudo`
  as container root inside the rootless user namespace, not as host root.
- **Lima**: persistent Linux VMs; a separate guest kernel and normal Docker or
  Podman workflows inside the guest.

A backend must be able to:

- validate prerequisites;
- create, start, stop, inspect, and delete a runtime;
- open a shell and run a command;
- report its state and storage details.

A future backend should implement this interface without changing Box commands
or portable definition fields.

Backend-specific details, such as a container or VM name, belong in metadata,
not in the portable definition.

## Safety

Host exposure is always explicit and visible in `box inspect`.

Never implicitly provide host container-engine sockets, SSH keys, cloud
configuration, credential directories, writable host mounts, host namespaces,
or privileged containers. The sole container-engine socket exception is the
persisted, inspected `--insecure-mode` opt-in, which exposes only the host user's
rootless Podman API. It does not enable privileged containers or host
namespaces.

- Podman uses `keep-id` in a rootless user namespace. The normal Box user may
  elevate inside the container, while container root remains mapped to an
  unprivileged subordinate host ID.
- Podman shares the host kernel and is not a boundary for hostile code.
- Explicit writable mounts can modify their corresponding host files.
- Lima provides a guest-kernel boundary, but a writable workspace mount can
  still modify host files.

Changing the backend, base environment, mounts, or network policy requires an
explained runtime recreation. Destructive actions require confirmation.

## Implementation order

1. Finish the backend-neutral definition and lifecycle states.
2. Add the internal backend interface.
3. Implement rootless Podman.
4. Implement Lima.
5. Add `enter`, `exec`, safe edits, and deletion.
6. Add future backends only after the core lifecycle is stable.
