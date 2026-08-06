#!/usr/bin/env sh
set -eu

repo="MuktadirHassan/box"
version="latest"
arch=""
install_dir="${BOX_INSTALL_DIR:-$HOME/.local/bin}"
install_completions=true

usage() {
	cat <<'EOF'
Usage: install.sh [--version VERSION] [--arch ARCH] [--install-dir DIR]

Install the latest Box release for Linux amd64 or arm64.

Options:
  --version VERSION    Release version, with or without a leading v
  --arch ARCH          amd64 or arm64; defaults to the current machine
  --install-dir DIR    Destination directory (default: ~/.local/bin)
  --no-completions     Do not install Fish and Bash completions
  -h, --help           Show this help message
EOF
}

die() {
	printf '%s\n' "error: $*" >&2
	exit 1
}

require() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--version)
			[ "$#" -ge 2 ] || die "--version requires a value"
			version=$2
			shift 2
			;;
		--arch)
			[ "$#" -ge 2 ] || die "--arch requires a value"
			arch=$2
			shift 2
			;;
		--install-dir)
			[ "$#" -ge 2 ] || die "--install-dir requires a value"
			install_dir=$2
			shift 2
			;;
		--no-completions)
			install_completions=false
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			die "unknown option: $1"
			;;
	esac
done

require curl
require sha256sum
require tar
require install

if [ -z "$arch" ]; then
	case "$(uname -m)" in
		x86_64) arch="amd64" ;;
		aarch64|arm64) arch="arm64" ;;
		*) die "unsupported architecture: $(uname -m)" ;;
	esac
fi

case "$arch" in
	amd64|arm64) ;;
	*) die "unsupported architecture: $arch (supported: amd64, arm64)" ;;
esac

if [ "$version" = "latest" ]; then
	version=$(curl --fail --location --silent --show-error "https://api.github.com/repos/$repo/releases/latest" |
		sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' |
		head -n 1)
	[ -n "$version" ] || die "could not determine the latest release version"
fi

case "$version" in
	v*) tag="$version" ;;
	*) tag="v$version" ;;
esac
release_version=${tag#v}
archive="box_${release_version}_linux_${arch}.tar.gz"
base_url="https://github.com/$repo/releases/download/$tag"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

printf 'Installing Box %s for linux/%s...\n' "$tag" "$arch"
curl --fail --location --retry 3 --silent --show-error --output "$tmp_dir/$archive" "$base_url/$archive"
curl --fail --location --retry 3 --silent --show-error --output "$tmp_dir/checksums.txt" "$base_url/checksums.txt"

checksum=$(awk -v archive="$archive" '$2 == archive || $2 == "*" archive { print }' "$tmp_dir/checksums.txt")
[ -n "$checksum" ] || die "checksum for $archive was not found"
printf '%s\n' "$checksum" | (cd "$tmp_dir" && sha256sum --check --status -) || die "checksum verification failed"

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"
[ -f "$tmp_dir/box" ] || die "release archive does not contain box"
install -Dm755 "$tmp_dir/box" "$install_dir/box"

if [ "$install_completions" = true ]; then
	fish_completions_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions"
	bash_completions_dir="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions"
	zsh_completions_dir="${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions"
	mkdir -p "$fish_completions_dir" "$bash_completions_dir" "$zsh_completions_dir"
	"$install_dir/box" completion fish > "$fish_completions_dir/box.fish"
	"$install_dir/box" completion bash > "$bash_completions_dir/box"
	"$install_dir/box" completion zsh > "$zsh_completions_dir/_box"
	printf 'Installed Fish, Bash, and Zsh completions.\n'
fi

printf 'Installed Box %s to %s/box\n' "$tag" "$install_dir"
case ":$PATH:" in
	*":$install_dir:"*) ;;
	*) printf 'Add %s to PATH to run box from any directory.\n' "$install_dir" ;;
esac
