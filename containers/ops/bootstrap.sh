#!/usr/bin/env bash
set -Eeuo pipefail

readonly DOTFILES_ROOT="${DOTFILES_ROOT:-/home/developer/dotfiles}"
readonly CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"

link_if_missing() {
    local source="$1"
    local target="$2"

    mkdir -p -- "$(dirname -- "$target")"
    if [[ -e "$target" || -L "$target" ]]; then
        printf 'Keeping existing configuration: %s\n' "$target"
        return
    fi

    ln -s -- "$source" "$target"
    printf 'Linked %s -> %s\n' "$target" "$source"
}

if [[ ! -d "$DOTFILES_ROOT/config" ]]; then
    printf 'Dotfiles checkout is unavailable at %s\n' "$DOTFILES_ROOT" >&2
    exit 1
fi

link_if_missing "$DOTFILES_ROOT/config/fish/config.fish" "$CONFIG_HOME/fish/config.fish"
link_if_missing "$DOTFILES_ROOT/config/fish/functions/mkcd.fish" "$CONFIG_HOME/fish/functions/mkcd.fish"
link_if_missing "$DOTFILES_ROOT/config/git/config" "$CONFIG_HOME/git/config"
link_if_missing "$DOTFILES_ROOT/config/starship.toml" "$CONFIG_HOME/starship.toml"
link_if_missing "$DOTFILES_ROOT/config/tmux/tmux.conf" "$CONFIG_HOME/tmux/tmux.conf"

printf '%s\n' 'ops bootstrap complete. Install only the operational tools and credentials required for this machine.'
