#!/usr/bin/env sh
set -eu

repo="MuktadirHassan/box"
version="latest"
arch=""
install_dir="${BOX_INSTALL_DIR:-$HOME/.local/bin}"
completion_shell="all"

usage() {
	cat <<'EOF'
Usage: install.sh [--version VERSION] [--arch ARCH] [--install-dir DIR] [--shell SHELL]

Install the latest Box release for Linux amd64 or arm64.

Options:
  --version VERSION    Release version, with or without a leading v
  --arch ARCH          amd64 or arm64; defaults to the current machine
  --install-dir DIR    Destination directory (default: ~/.local/bin)
  --shell SHELL        Install completions for bash, fish, zsh, all, or none (default: all)
  --no-completions     Do not install shell completions
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
		--shell)
			[ "$#" -ge 2 ] || die "--shell requires a value"
			completion_shell=$2
			shift 2
			;;
		--no-completions)
			completion_shell="none"
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

case "$completion_shell" in
	bash|fish|zsh|all|none) ;;
	*) die "unsupported shell: $completion_shell (supported: bash, fish, zsh, all, none)" ;;
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
completion_tmp=""
trap 'rm -rf "$tmp_dir"; [ -z "$completion_tmp" ] || rm -f "$completion_tmp"' EXIT HUP INT TERM

printf 'Installing Box %s for linux/%s...\n' "$tag" "$arch"
curl --fail --location --retry 3 --silent --show-error --output "$tmp_dir/$archive" "$base_url/$archive"
curl --fail --location --retry 3 --silent --show-error --output "$tmp_dir/checksums.txt" "$base_url/checksums.txt"

checksum=$(awk -v archive="$archive" '$2 == archive || $2 == "*" archive { print }' "$tmp_dir/checksums.txt")
[ -n "$checksum" ] || die "checksum for $archive was not found"
printf '%s\n' "$checksum" | (cd "$tmp_dir" && sha256sum --check --status -) || die "checksum verification failed"

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"
[ -f "$tmp_dir/box" ] || die "release archive does not contain box"
install -Dm755 "$tmp_dir/box" "$install_dir/box"

install_completion() {
	shell=$1
	case "$shell" in
		fish) destination="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions/box.fish" ;;
		bash) destination="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/box" ;;
		zsh) destination="${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions/_box" ;;
	esac
	completion_dir=$(dirname "$destination")
	mkdir -p "$completion_dir" || return 1
	completion_tmp=$(mktemp "$completion_dir/.box-completion.XXXXXX") || return 1
	if "$install_dir/box" completion "$shell" > "$completion_tmp" && mv -f "$completion_tmp" "$destination"; then
		completion_tmp=""
		printf 'Installed %s completions to %s\n' "$shell" "$destination"
		if [ "$shell" = "zsh" ]; then
			printf 'Add %s to fpath before running compinit; see the README for details.\n' "$completion_dir"
		fi
		return 0
	fi
	rm -f "$completion_tmp"
	completion_tmp=""
	return 1
}

case "$completion_shell" in
	bash|fish|zsh) install_completion "$completion_shell" || die "could not install $completion_shell completions" ;;
	all)
		install_completion bash || die "could not install bash completions"
		install_completion fish || die "could not install fish completions"
		install_completion zsh || die "could not install zsh completions"
		;;
esac

printf 'Installed Box %s to %s/box\n' "$tag" "$install_dir"
case ":$PATH:" in
	*":$install_dir:"*) ;;
	*) printf 'Add %s to PATH to run box from any directory.\n' "$install_dir" ;;
esac
