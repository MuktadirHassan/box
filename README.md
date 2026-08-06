# Box

Box creates persistent, rootless Podman development environments on Linux.

## Requirements

- Go 1.26.5 or later to build Box
- Podman available to your user

## Install

Download a Linux archive for your architecture from the [GitHub releases page](https://github.com/MuktadirHassan/box/releases), verify it against `checksums.txt`, and place `box` on your `PATH`:

```bash
curl -LO https://github.com/MuktadirHassan/box/releases/download/v<VERSION>/box_<VERSION>_linux_amd64.tar.gz
curl -LO https://github.com/MuktadirHassan/box/releases/download/v<VERSION>/checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf box_<VERSION>_linux_amd64.tar.gz
install -Dm755 box "$HOME/.local/bin/box"
```

Replace `<VERSION>` with a released version without the leading `v`. Releases are available for Linux `amd64` and `arm64`; substitute `arm64` in the archive name where appropriate.

Confirm the installed release:

```bash
box --version
```

## Build from source

```bash
go build -o box .
```

## Usage

Create a box, configure it, and enter its shell:

```bash
./box create work
./box setup work
./box enter work
```

`setup` opens an interactive configuration on first use. For scripting, provide
configuration flags and confirm with `--yes`:

```bash
./box setup work \
  --image ubuntu:24.04 \
  --mount "$HOME/projects:/workspace" \
  --cpus 4 \
  --memory 8g \
  --yes
```

Run a command without opening a shell:

```bash
./box exec work -- go test ./...
```

Manage boxes:

```bash
./box list
./box inspect work
./box stop work
./box delete work --purge
```

`--purge` is required to delete a box. It removes its runtime, managed persistent
data, and saved definition. Configuration changes made with `setup` recreate the
runtime while preserving managed home and cache data.

Use `./box <command> --help` to see all options.

Box stores definitions in `~/.local/share/box/boxes/`.

## Releases

Releases use [Semantic Versioning](https://semver.org/) and [Conventional Commits](https://www.conventionalcommits.org/). After a Conventional Commit reaches `main`, Release Please opens or updates a release pull request with the calculated version and changelog. Merging that pull request creates the `vMAJOR.MINOR.PATCH` tag and GitHub release, then GoReleaser builds and attaches Linux `amd64` and `arm64` archives plus `checksums.txt`.

Use `feat:` for a minor release, `fix:` for a patch release, and add `!` or a `BREAKING CHANGE:` footer for a major release. Commits such as `docs:`, `test:`, `chore:`, `ci:`, and `refactor:` do not appear in release notes.

See [architecture.md](architecture.md) for design details and
[internal/templates/README.md](internal/templates/README.md) for environment
templates.
