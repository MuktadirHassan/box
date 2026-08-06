# Box

A Linux CLI for creating persistent, rootless Podman development environments.

## Requirements

- Linux
- [Podman](https://podman.io/) available to your user

## Install

Download the archive for your architecture from [Releases](https://github.com/MuktadirHassan/box/releases), verify it, and install it on your `PATH`:

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

## Build from source

Requires Go 1.26.5 or later.

```bash
git clone https://github.com/MuktadirHassan/box.git
cd box
go build -o box .
```

## Usage

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

`--purge` is required for deletion. Box keeps definitions in `~/.local/share/box/boxes/`.

## Documentation

- [Architecture](architecture.md)
- [Environment templates](internal/templates/README.md)
- `box --help`
