# Box

Box provides a Linux CLI and Desktop alpha for creating persistent, rootless Podman development environments.

> **Alpha:** Box is pre-1.0 software. Interfaces, configuration, and behavior may change without migration support. Do not rely on it for critical workloads.

## Requirements

### CLI

- Linux
- [Podman](https://podman.io/) available to your user

### Desktop alpha

Desktop packages support Linux amd64 with glibc 2.39: Ubuntu 24.04+, Fedora 40+ with WebKitGTK 4.1, and current Arch Linux/CachyOS. They need rootless Podman plus GTK3 and WebKitGTK 4.1 at runtime; native packages normally install the GTK and WebKitGTK dependencies through the system package manager.

## CLI installation

Install the latest CLI release for your architecture:

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

### Manual CLI installation

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

## Desktop app alpha

The Desktop alpha is a graphical view of the same Box data and Podman environments used by the CLI. It lists each box with its state and image, lets you create a named box or leave the name blank, and shows command output for every action.

Select a listed box to:

- **Inspect** its definition and runtime details.
- **Set up with defaults**, which confirms and runs the noninteractive CLI setup defaults (`box setup <name> --yes`).
- **Stop** its runtime.
- **Delete**, which requires typing the box name and permanently purges its runtime, managed persistent data, and definition.

The Desktop alpha does not replace CLI-only workflows such as entering a box or supplying custom setup flags.

### Install the Desktop app

Download the native package for Linux amd64 from the [GitHub Releases](https://github.com/MuktadirHassan/box/releases) page. Release assets are named for the `box-desktop` package; use the matching downloaded asset rather than substituting a version number:

```bash
# Ubuntu 24.04+
sudo apt install ./box-desktop_*.deb

# Fedora 40+ with WebKitGTK 4.1
sudo dnf install ./box-desktop_*.rpm

# Current Arch Linux / CachyOS
sudo pacman -U ./box-desktop_*.pkg.tar.zst
```

### Uninstall the Desktop app

```bash
# Ubuntu 24.04+
sudo apt remove box-desktop

# Fedora 40+ with WebKitGTK 4.1
sudo dnf remove box-desktop

# Current Arch Linux / CachyOS
sudo pacman -R box-desktop
```

Uninstalling the desktop package does **not** remove `~/.local/share/box`, Podman containers, or Podman volumes. Use [Remove an environment](#remove-an-environment) or [Remove all Box data](#remove-all-box-data) when you intend to permanently delete that data.

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
  --template ubuntu-24.04-terminal-tools \
  --shell fish \
  --prompt starship \
  --yes
```

`ubuntu-24.04-terminal-tools` supports Ubuntu 24.04 images. It installs Bash, Fish, or Zsh as selected, and can add a Starship prompt without changing existing shell configuration. Use `box setup work --refresh-template --yes` after a template update to add new default files without overwriting files already in the persistent home.

Run a command or manage environments:

```bash
box exec work -- go test ./...
box list
box inspect work
box stop work
box delete work --purge
```

## Environment templates

Use the canonical template ID when configuring an environment:

```bash
box setup work --template ubuntu-24.04-terminal-tools --image ubuntu:24.04 --yes
```

Template IDs are backed by catalog manifests that declare image release, supported shells, and prompts. Ubuntu 24.04 templates require `ubuntu:24.04` (a digest-qualified 24.04 reference is also accepted); `latest` and other releases are rejected.

## Upgrade

Run the CLI installer again to replace the CLI binary. Definitions remain in `~/.local/share/box/boxes/`. Back up `~/.local/share/box/` before upgrading across minor alpha versions. Upgrade the desktop app by installing the newer native package from the release page.

## CLI uninstall

Remove only the CLI binary and completions:

```bash
rm -f "${BOX_INSTALL_DIR:-$HOME/.local/bin}/box"
rm -f "${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions/box.fish"
rm -f "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/box"
rm -f "${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions/_box"
```

This removes the CLI binary and installed completions, while keeping environments and their data.

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

Requires Go 1.26.5 or later. Native desktop builds also require a C compiler, `pkg-config`, and GTK3 and WebKitGTK 4.1 development headers.

```bash
git clone https://github.com/MuktadirHassan/box.git
cd box

# CLI
go build -o box .

# Desktop alpha: the frontend is plain static HTML, CSS, and JavaScript; no Node build is required.
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2
export PATH="$(go env GOPATH)/bin:$PATH"
wails doctor
cd desktop
wails build -tags webkit2_41
```

Use the pinned Wails v2.10.2 command above rather than an arbitrary global Wails version. The `webkit2_41` build tag selects the WebKitGTK 4.1 binding used by the desktop app.

## Versioning and documentation

Box follows [Semantic Versioning](https://semver.org/). `v0.1.0` is the first release; before `v1.0.0`, minor versions may contain breaking changes and patch versions contain compatible fixes.

- [Releases](https://github.com/MuktadirHassan/box/releases)
- [Release guidelines](docs/releasing.md)
- [Architecture](docs/architecture.md)
- [Environment templates](templates/README.md)
- `box --help`
