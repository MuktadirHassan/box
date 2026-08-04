#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
readonly ROOT
readonly CONTAINER_HOME_BASE="${CONTAINER_HOME_BASE:-$HOME/.local/share/box-homes}"
readonly CONTAINER_PREFIX="${CONTAINER_PREFIX:-box}"
readonly CONTAINER_USER="developer"

require_command() {
    local command="$1"

    if ! command -v "$command" >/dev/null 2>&1; then
        printf 'Required command is not installed: %s\n' "$command" >&2
        exit 1
    fi
}

run() {
    if [[ "$dry_run" == true ]]; then
        printf 'Would run:'
        printf ' %q' "$@"
        printf '\n'
    else
        "$@"
    fi
}

container_name() {
    printf '%s-%s' "$CONTAINER_PREFIX" "$1"
}

image_name() {
    printf 'local/box-%s:latest' "$1"
}

container_exists() {
    podman container exists "$(container_name "$1")"
}

build_image() {
    local name="$1"

    require_command podman
    run podman build \
        --build-arg "USER_UID=$(id -u)" \
        --build-arg "USER_GID=$(id -g)" \
        --tag "$(image_name "$name")" \
        --file "$ROOT/containers/$name/Containerfile" \
        "$ROOT/containers/$name"
}

container_options() {
    local name="$1"
    local home="$2"

    printf '%s\0' \
        --name "$(container_name "$name")" \
        --userns keep-id \
        --user "$(id -u):$(id -g)" \
        --env HOME="/home/$CONTAINER_USER" \
        --workdir "/home/$CONTAINER_USER" \
        --hostname "$name" \
        --network pasta \
        --read-only \
        --cap-drop ALL \
        --security-opt no-new-privileges \
        --pids-limit "${PODMAN_PIDS_LIMIT:-512}" \
        --memory "${PODMAN_MEMORY_LIMIT:-8g}" \
        --cpus "${PODMAN_CPU_LIMIT:-4}" \
        --mount "type=bind,src=$home,dst=/home/$CONTAINER_USER,rw" \
        --tmpfs "/tmp:rw,nosuid,nodev,size=${PODMAN_TMP_SIZE:-2g}" \
        --tmpfs /run:rw,nosuid,nodev,size=16m
}

read_container_options() {
    local name="$1"
    local home="$2"

    mapfile -d '' -t container_args < <(container_options "$name" "$home")
}

prepare_home() {
    run mkdir -p -- "$1"
}

remove_container() {
    local name="$1"

    if container_exists "$name"; then
        run podman rm --force "$(container_name "$name")"
    fi
}

create_container() {
    local name="$1"
    local home="$2"
    shift 2

    if container_exists "$name"; then
        printf 'Container %q already exists. Use --recreate to replace it.\n' "$(container_name "$name")" >&2
        exit 1
    fi

    read_container_options "$name" "$home"
    run podman create --tty "${container_args[@]}" "$@" "$(image_name "$name")" fish --login
}
