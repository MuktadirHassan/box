# Box development environment (transitional source)

This directory is transitional Box source. It currently defines the
`dotfiles-dev` Podman development container, but the container application is
being extracted from this dotfiles repository. New documentation and workflows
should target Box rather than this path.

## Launch from the application menu

`dev.desktop` creates a **Dev Container** launcher that opens an interactive terminal attached to the container through `enter`.

Install the launcher for the current user:

```sh
mkdir -p ~/.local/share/applications
cp ~/dotfiles/containers/dev/dev.desktop ~/.local/share/applications/
```

Close and reopen the application launcher after installing it. If the launcher is already open, restarting the desktop session may be necessary for it to refresh.

The desktop file uses the absolute path `/home/tamim/dotfiles/containers/dev/enter`. If the dotfiles repository lives somewhere else, update the `Exec` and `TryExec` paths in `dev.desktop` before copying it.

## Command-line use

Start or attach to the container directly:

```sh
~/dotfiles/containers/dev/enter
```

Create the container first if it does not exist:

```sh
~/dotfiles/containers/dev/create
```
