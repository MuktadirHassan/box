#!/usr/bin/env bash
set -Eeuo pipefail

readonly MISE_VERSION="${MISE_VERSION:-v2026.7.18}"
readonly DOTFILES_ROOT="${DOTFILES_ROOT:-/home/developer/dotfiles}"
readonly CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
readonly LOCAL_BIN_DIR="$HOME/.local/bin"

usage() {
    cat <<'EOF'
Usage: bootstrap.sh [--skip-mise]

Install the user-level development baseline for the dev container.
Set MISE_VERSION to deliberately choose a different mise release.
EOF
}

skip_mise=false
while (($#)); do
    case "$1" in
        --skip-mise) skip_mise=true ;;
        -h|--help) usage; exit 0 ;;
        *) printf 'Unknown option: %s\n' "$1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

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

mkdir -p -- "$HOME/DeveloperCache/pnpm-store" "$HOME/DeveloperCache/go/pkg/mod" "$HOME/DeveloperCache/go/build" "$HOME/DeveloperCache/cargo" "$HOME/DeveloperCache/uv" "$HOME/DeveloperCache/npm" "$HOME/DeveloperCache/bun"

if [[ "$skip_mise" == false ]] && ! command -v mise >/dev/null 2>&1; then
    case "$(uname -m)" in
        x86_64|amd64) mise_architecture=x64 ;;
        aarch64|arm64) mise_architecture=arm64 ;;
        *) printf 'Unsupported architecture for mise: %s\n' "$(uname -m)" >&2; exit 1 ;;
    esac

    mise_archive="mise-${MISE_VERSION}-linux-${mise_architecture}.tar.gz"
    mise_url="https://github.com/jdx/mise/releases/download/${MISE_VERSION}/${mise_archive}"
    mise_tmpdir="$(mktemp -d)"
    trap 'rm -rf -- "$mise_tmpdir"' EXIT

    printf 'Installing mise %s\n' "$MISE_VERSION"
    curl --fail --location --proto '=https' --tlsv1.2 --retry 2 --retry-delay 2 \
        "$mise_url" -o "$mise_tmpdir/$mise_archive"
    tar -xzf "$mise_tmpdir/$mise_archive" -C "$mise_tmpdir"
    install -Dm0755 "$mise_tmpdir/mise/bin/mise" "$LOCAL_BIN_DIR/mise"
fi

printf 'dev bootstrap complete. Define project tool versions in mise.toml.\n'
