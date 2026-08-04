# Box

Box provides managed, rootless Podman environments for local development and
operations.

## Create and enter

Create and enter the daily development environment:

```bash
./containers/dev/create
./containers/dev/enter
```

The smaller operations environment is created and entered separately:

```bash
./containers/ops/create
./containers/ops/enter
```

Creation builds a local image and creates a persistent container with its own
home. `dev` uses Arch Linux; `ops` uses Ubuntu 24.04. Rebuild a container after
its image or mounts change:

```bash
./containers/dev/create --recreate
```

`--recreate` removes only the container, not its home at
`~/.local/share/box-homes/<role>`. Use `--dry-run` to print state-changing
commands without running them.

## Host exposure

Each container has a dedicated writable home. `dev` additionally receives these
explicit read/write mounts:

- `~/DeveloperCache` at `/home/developer/DeveloperCache`
- `/ssd2/projects` at `/home/developer/projects`
- the active Wayland socket at `/tmp/$WAYLAND_DISPLAY`, with `WAYLAND_DISPLAY`
  and `XDG_RUNTIME_DIR=/tmp` forwarded so graphical clipboard clients work

Wayland clients can inspect and interact with the graphical session, including
its clipboard and input. Treat `dev` as trusted local code; do not run untrusted
images or code there. Creation requires an active Wayland session and fails
clearly otherwise.

The wrappers do not mount the rest of the host home, `/run/host`, host SSH keys,
cloud configuration, kubeconfig, or a Podman/Docker socket. They run rootlessly
with `--userns keep-id`, no Linux capabilities, no new privileges, a read-only
container filesystem, explicit writable home/cache mounts, and tmpfs mounts for
`/tmp` and `/run`.

Networking uses `pasta`: containers have outbound access but receive no host
ports unless a future wrapper explicitly publishes one. Defaults limit
containers to 512 processes, 8 GiB memory, and four CPUs. Override these
per-invocation when needed, for example:

```bash
PODMAN_MEMORY_LIMIT=16g PODMAN_CPU_LIMIT=8 ./containers/dev/create --recreate
```

## Roles and boundary

- `dev` is the daily development baseline. Define language versions in each
  project with `mise.toml`.
- `ops` is a small terminal baseline. Add operational tools and credentials only
  for a deliberate workflow.

This is a strong workflow and filesystem-isolation boundary, not a virtual
machine. Rootless containers still share the host kernel; use a VM or separate
machine for untrusted or hostile code.
