# Box

A Linux CLI for creating persistent, rootless Podman development environments.

> **Alpha:** Box is pre-1.0 software. Interfaces, configuration, and behavior may change without migration support. Do not rely on it for critical workloads.

## Requirements

- Linux
- [Podman](https://podman.io/) available to your user

## Install

Install the latest release for your architecture:

```bash
curl -fsSL https://raw.githubusercontent.com/MuktadirHassan/box/main/install.sh | sh
```

The script downloads the archive and `checksums.txt`, verifies the archive, and installs `box` to `~/.local/bin`. By default it generates Fish, Bash, and Zsh completions; pass `--no-completions` to skip them, or use `--shell bash`, `--shell fish`, or `--shell zsh` to install a single completion. Fish discovers its completion automatically. Bash requires the usual `bash-completion` user-completion support.

To install a specific version or use another destination:

```bash
curl -fsSL https://raw.githubusercontent.com/MuktadirHassan/box/main/install.sh | sh -s -- \
  --version 0.1.0 \
  --install-dir /usr/local/bin
```

### Zsh setup

Fish loads its completion automatically. For Zsh, add this before `compinit` in `~/.zshrc`:

```zsh
fpath=("${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions" $fpath)
autoload -Uz compinit && compinit
```

### Manual installation

The installer performs these steps. Use them when you want to inspect or control each step yourself:

```bash
VERSION=0.1.0 # replace with the release version, without "v"
ARCH=amd64    # amd64 or arm64

curl -LO "https://github.com/MuktadirHassan/box/releases/download/v${VERSION}/box_${VERSION}_linux_${ARCH}.tar.gz"
curl -LO "https://github.com/MuktadirHassan/box/releases/download/v${VERSION}/checksums.txt"
sha256sum --ignore-missing --check checksums.txt
tar -xzf "box_${VERSION}_linux_${ARCH}.tar.gz"
install -Dm755 box "$HOME/.local/bin/box"

box --version
```

## Quick start

Create, configure, and enter a writable Ubuntu development environment:

```bash
box create work
box setup work
box enter work
```

The default environment runs as a normal Linux user and provides passwordless
`sudo` inside the rootless container. Its baseline tools include CA
certificates, Curl, Git, networking and process utilities, Neovim, tmux, and
ripgrep. Install anything else normally:

```bash
sudo apt update
sudo apt install build-essential
```

Container root is scoped to Podman's rootless user namespace; it is not host
root. Box does not enable privileged mode, host namespaces, host
container-engine sockets, or implicit host mounts. Explicit writable mounts can
still modify the corresponding host files, and containers share the host
kernel, so Box is intended for trusted local development rather than hostile
code isolation.

`setup` is interactive by default. Use flags and `--yes` for scripts:

```bash
box setup work \
  --image ubuntu:24.04 \
  --mount "$HOME/projects:/workspace" \
  --mount "$HOME/.claude:~/.claude" \
  --cpus 4 \
  --memory 8g \
  --shell fish \
  --prompt starship \
  --yes
```

Interactive setup can add any number of writable host mounts. Repeat `--mount`
for scripted setup. Destinations may be absolute paths or `~` / `~/path`; Box
resolves the latter to the selected container user's home (for example,
`/home/amber-otter/.claude`), not the host user's home. User mounts are
read-write, and every `--mount` supplied to a later setup replaces the saved
mount list.

The built-in `ubuntu-24.04-terminal-tools` development template is selected by
default. It supports Ubuntu 24.04 images, installs Bash, Fish, or Zsh as
selected, and can add a Starship prompt without changing existing shell
configuration. A digest-qualified Ubuntu 24.04 image is also accepted; `latest`
and other releases are rejected.

Packages and other root-filesystem changes survive normal stop/start operations,
but recreating a Box during `box setup` rebuilds that filesystem. Keep durable
user configuration and project data in the persistent home or explicit mounts.
Use `box setup work --refresh-template --yes` to add new default files without
overwriting files already in the persistent home.

Run a command or manage environments:

```bash
box exec work -- go test ./...
box list
box inspect work
box stop work
box delete work --purge
```

## Environment templates

Template IDs are backed by catalog manifests that declare image release,
supported shells, and prompts. The current built-in template is
`ubuntu-24.04-terminal-tools`.

## Upgrade

Run the installer again to replace the binary. Definitions remain in `~/.local/share/box/boxes/`. Back up `~/.local/share/box/` before upgrading across minor alpha versions.

## Uninstall

Remove only the Box binary:

```bash
rm -f "${BOX_INSTALL_DIR:-$HOME/.local/bin}/box"
rm -f "${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions/box.fish"
rm -f "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/box"
rm -f "${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions/_box"
```

This removes the binary and installed completions, while keeping environments and their data.

### Remove an environment

Permanently remove one environment, including its runtime and managed persistent data:

```bash
box delete <name> --purge
```

### Remove all Box data

> **Warning:** This permanently deletes every Box environment, its managed persistent data, and saved definitions.

```bash
set -e
for definition in "$HOME"/.local/share/box/boxes/*; do
  [ -d "$definition" ] || continue
  box delete "$(basename "$definition")" --purge
done
rm -rf "$HOME/.local/share/box"
```

## Build from source

Requires Go 1.26.5 or later.

```bash
git clone https://github.com/MuktadirHassan/box.git
cd box
go build -o box .
```

### Test local changes with an existing Box

Rebuild and replace the local binary after changing Box or an embedded template, then recreate the environment with the updated template:

```bash
go build -o box .
install -Dm755 box "$HOME/.local/bin/box"
box setup <box-name> --refresh-template --yes
```

Recreating a Box replaces its container root filesystem. The persistent home directory and explicit mounts remain, but packages and other root-filesystem changes made inside the existing container are removed. Use a test Box when developing template changes.

## Versioning and documentation

Box follows [Semantic Versioning](https://semver.org/). `v0.1.0` is the first release; before `v1.0.0`, minor versions may contain breaking changes and patch versions contain compatible fixes.

- [Releases](https://github.com/MuktadirHassan/box/releases)
- [Architecture](docs/architecture.md)
- [Environment templates](templates/README.md)
- `box --help`
