# Dotfiles

A small, dependency-free dotfiles setup for CachyOS/Arch Linux and Ubuntu/WSL2.
It uses plain Bash scripts and symbolic links—no dotfile manager required.

## Layout

```text
.
├── install.sh              # Install packages, then link configs
├── link.sh                 # Safely link configs only
├── arch.sh                 # pacman package installation
├── ubuntu.sh               # apt package installation
├── common.sh               # Per-distribution package lists
├── docker/                  # Optional Docker host configuration and guide
├── podman/                  # Optional rootless Podman configuration and guide
├── containers/              # Transitional Box source; pending extraction
└── config/
    ├── environment.d/       # Global user-session environment variables
    ├── fastfetch/
    ├── fish/
    ├── git/
    ├── nvim/
    ├── starship.toml
    └── tmux/
        └── tmux.conf
```

## Box extraction

Managed development containers are becoming **Box**, a separate
application. The existing `containers/` directory is transitional source kept
only while that extraction is in progress; it is not a supported part of this
dotfiles setup and should not receive new dotfiles-specific workflows.

This repository continues to own its optional Docker and rootless Podman host
configuration. Box will own its container definitions, lifecycle commands,
bootstrap logic, desktop launchers, and container-specific documentation.

## Install

Clone the repository anywhere, then run:

```bash
./install.sh
```

The installer detects `pacman` or `apt-get`, installs available packages, and
runs the linker. Packages missing from the configured repositories are reported
and skipped rather than making the whole setup fail. The deliberately small
package list lives in `common.sh`; add tools only when they become useful.

### Included packages

| Package | Purpose |
| --- | --- |
| `git` | Track source code and collaborate through Git repositories |
| `fish` | Interactive shell with friendly defaults and completions |
| `neovim` | Terminal text editor |
| `ripgrep` | Search text recursively with the `rg` command |
| `starship` | Consistent, informative shell prompt |
| `tmux` | Terminal multiplexer with sessions, windows, and pane management |
| `fastfetch` | Display a concise summary of the current system |

Tmux is now included in the default package list; once installed, run
`./link.sh` again to materialize `~/.config/tmux/tmux.conf`.

To create links without installing packages:

```bash
./install.sh --links-only
```

To preview link operations without changing anything:

```bash
./link.sh --dry-run
```

## Linking and backups

The linker creates these user-level links under `${XDG_CONFIG_HOME:-~/.config}`:

| Repository path | Config path |
| --- | --- |
| `config/environment.d/10-base.conf` | `environment.d/10-base.conf` |
| `config/environment.d/20-development-caches.conf` | `environment.d/20-development-caches.conf` |
| `config/fish/config.fish` | `fish/config.fish` |
| `config/fish/functions/mkcd.fish` | `fish/functions/mkcd.fish` |
| `config/ghostty` | `ghostty` |
| `config/hypr` | `hypr` |
| `config/git/config` | `git/config` |
| `config/nvim` | `nvim` |
| `config/fastfetch` | `fastfetch` |
| `config/starship.toml` | `starship.toml` |
| `config/tmux/tmux.conf` | `tmux/tmux.conf` |

Existing files, directories, and symbolic links are preserved as
`<name>.bak.YYYYMMDD-HHMMSS`. Running the linker again is safe: links already
pointing at this checkout are left alone. Fish files are linked individually so
machine-local generated state such as `fish_variables` remains untracked.

The `environment.d` files provide the editor, locale, and development-cache
variables to the user session rather than only to Fish. Log out and back in
after changing them so the desktop session receives the updated environment.

## Git identity

On a clean machine, Git reads the linked XDG config directly. If
`~/.gitconfig` already exists (for example, with GitHub CLI credentials), the
linker adds the tracked config to that file's `include.path` without replacing
its existing settings. Machine- or employer-specific Git identity does not
belong in this repository.
Put it in `~/.gitconfig.local`, which the tracked Git config includes:

```gitconfig
[user]
    name = Your Name
    email = you@example.com
```

## Commands

The initial setup includes:

- `mkcd DIRECTORY` — Fish function that creates a directory and enters it
- tmux setup available at `~/.config/tmux/tmux.conf`

## Shared work drive

The repository does not need to live at a fixed path. `link.sh` records the
checkout's absolute path each time it runs, so WSL2 and CachyOS may mount the
shared drive at different locations. Run `./link.sh` once from each operating
system after cloning or moving the checkout.

Keep secrets, tokens, SSH keys, and machine-specific credentials outside this
repository. Avoid committing generated package stores such as `node_modules`,
the pnpm store, Cargo caches, or npm caches.

