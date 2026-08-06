# Box

Box creates persistent, rootless Podman development environments on Linux.

## Requirements

- Go 1.26.5 or later to build Box
- Podman available to your user

## Build

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

See [architecture.md](architecture.md) for design details and
[internal/templates/README.md](internal/templates/README.md) for environment
templates.
