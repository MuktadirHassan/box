# Box (transitional source)

Box is the standalone application for managed, rootless Podman development
environments. This directory is its temporary in-repository source while the
application is extracted from dotfiles. It is no longer part of the supported
dotfiles surface; its documentation remains here only to describe the existing
implementation during the transition.

The current wrappers provide daily development environments with a dedicated
container home and no implicit host-home integration. Complete the rootless
Podman host setup in [`../podman/`](../podman/) first.

> **Migration status:** Do not add new dotfiles-specific container workflows
> here. Future development, installation, and user documentation belong in the
> standalone Box application.

## Create and enter

```bash
./containers/dev/create
./containers/dev/enter
```

The smaller operations environment is created and entered separately:

```bash
./containers/ops/create
./containers/ops/enter
```

Creation builds a local image, bootstraps the dedicated home, and creates a
persistent container. `dev` uses Arch Linux; `ops` uses Ubuntu 24.04. Rebuild a
container intentionally after its image changes:

```bash
./containers/dev/create --recreate
```

Recreate `dev` after changes to its image or mounts, including the Wayland
clipboard integration. `--recreate` removes only the container, not its home at
`~/.local/share/podman-homes/<role>`. Use `--dry-run` to print all
state-changing commands.

## Host exposure

Both containers mount this checkout read-only at `/home/developer/dotfiles` and
use their own writable home directory. `dev` additionally has these explicit
read/write mounts:

- `~/DeveloperCache` at `/home/developer/DeveloperCache`
- `/ssd2/projects` at `/home/developer/projects`
- `~/.claude` at `/home/developer/.claude`
- the active Wayland socket at `/tmp/$WAYLAND_DISPLAY`, with `WAYLAND_DISPLAY`
  and `XDG_RUNTIME_DIR=/tmp` forwarded so graphical clipboard clients work

Wayland clients can inspect and interact with the graphical session, including
its clipboard and input. Treat `dev` as trusted local code; do not run untrusted
images or code there. Creation requires an active Wayland session and fails
clearly otherwise.

The wrappers do not mount the rest of the host home, `/run/host`, host SSH keys,
cloud configuration, kubeconfig, or a Podman/Docker socket. They run
rootlessly with `--userns keep-id`, no Linux capabilities, no new privileges, a
read-only container filesystem, explicit writable home/cache mounts, and tmpfs
mounts for `/tmp` and `/run`.

Networking uses `pasta`: containers have outbound access but receive no host
ports unless a future wrapper explicitly publishes one. Defaults limit
containers to 512 processes, 8 GiB memory, and four CPUs. Override these
per-invocation when needed, for example:

```bash
PODMAN_MEMORY_LIMIT=16g PODMAN_CPU_LIMIT=8 ./containers/dev/create --recreate
```

## Roles and boundary

- `dev` is the daily development baseline. It installs mise into the isolated
  home; define language versions in each repository's `mise.toml`.
- `ops` is a small terminal baseline. Add operational tools and credentials only
  for a deliberate workflow.

This is a strong workflow and filesystem-isolation boundary, not a virtual
machine. Rootless containers still share the host kernel; use a VM or separate
machine for untrusted or hostile code.
