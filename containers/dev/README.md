# Box development environment

This directory defines Box's `box-dev` Podman development container.

## Launch from the application menu

`dev.desktop` creates a **Dev Container** launcher that opens an interactive
terminal attached to the container through `enter`.

Install the launcher for the current user:

```sh
mkdir -p ~/.local/share/applications
cp /path/to/box/containers/dev/dev.desktop ~/.local/share/applications/
```

Update the `Exec` and `TryExec` paths in `dev.desktop` to this checkout's
absolute path before copying it. Close and reopen the application launcher after
installing it.

## Command-line use

Start or attach to the container directly:

```sh
./containers/dev/enter
```

Create the container first if it does not exist:

```sh
./containers/dev/create
```
