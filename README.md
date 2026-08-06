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

The script downloads the archive and `checksums.txt`, verifies the archive, and installs `box` to `~/.local/bin`.

To install a specific version or use another destination:

```bash
curl -fsSL https://raw.githubusercontent.com/MuktadirHassan/box/main/install.sh | sh -s -- \
  --version 0.1.0 \
  --install-dir /usr/local/bin
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

Create, configure, and enter an environment:

```bash
box create work
box setup work
box enter work
```

`setup` is interactive by default. Use flags and `--yes` for scripts:

```bash
box setup work \
  --image ubuntu:24.04 \
  --mount "$HOME/projects:/workspace" \
  --cpus 4 \
  --memory 8g \
  --yes
```

Run a command or manage environments:

```bash
box exec work -- go test ./...
box list
box inspect work
box stop work
box delete work --purge
```

## Upgrade

Run the installer again to replace the binary. Definitions remain in `~/.local/share/box/boxes/`. Back up `~/.local/share/box/` before upgrading across minor alpha versions.

## Uninstall

Remove only the Box binary:

```bash
rm -f "${BOX_INSTALL_DIR:-$HOME/.local/bin}/box"
```

This keeps environments and their data.

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

## Versioning and documentation

Box follows [Semantic Versioning](https://semver.org/). `v0.1.0` is the first release; before `v1.0.0`, minor versions may contain breaking changes and patch versions contain compatible fixes.

- [Releases](https://github.com/MuktadirHassan/box/releases)
- [Architecture](architecture.md)
- [Environment templates](internal/templates/README.md)
- `box --help`
