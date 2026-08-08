#!/usr/bin/env bash
set -Eeuo pipefail

fail() {
  printf 'e2e: %s\n' "$*" >&2
  exit 1
}

expect_contains() {
  local output=$1
  local expected=$2

  [[ "$output" == *"$expected"* ]] || fail "expected output to contain '$expected', got: $output"
}

if [[ ${RUN_E2E:-} != "1" ]]; then
  printf 'e2e: skipped (set RUN_E2E=1 to run Podman acceptance tests)\n'
  exit 0
fi

workdir=$(mktemp -d)
test_home="$workdir/home"
box_binary="$workdir/box"
run_suffix=$(basename "$workdir" | tr -cd '[:alnum:]')
box_name="e2e-${GITHUB_RUN_ID:-local}-${GITHUB_RUN_ATTEMPT:-1}-${run_suffix}"
container_name="box-$box_name"
home_volume="$container_name-home"
cache_volume="$container_name-cache"

cleanup() {
  local status=$?
  trap - EXIT

  if [[ -x "$box_binary" && -d "$test_home/.local/share/box/boxes/$box_name" ]]; then
    HOME="$test_home" "$box_binary" delete "$box_name" --purge >/dev/null 2>&1 || true
  fi
  podman rm --force "$container_name" >/dev/null 2>&1 || true
  podman volume rm --force "$home_volume" "$cache_volume" >/dev/null 2>&1 || true
  podman image rm "$container_name-template" >/dev/null 2>&1 || true
  rm -rf "$workdir" >/dev/null 2>&1 || true
  exit "$status"
}
trap cleanup EXIT

mkdir -p "$test_home"
export HOME="$test_home"
go build -o "$box_binary" .

[[ $(podman info --format '{{.Host.Security.Rootless}}') == "true" ]] || fail "Podman must run rootlessly"

version_output=$($box_binary --version)
expect_contains "$version_output" "box version"
for shell in bash fish zsh; do
  $box_binary completion "$shell" >"$workdir/$shell.completion"
  [[ -s "$workdir/$shell.completion" ]] || fail "$shell completion is empty"
done

create_output=$($box_binary create "$box_name")
expect_contains "$create_output" "Created box \"$box_name\""
list_output=$($box_binary list)
expect_contains "$list_output" "$box_name"
inspect_output=$($box_binary inspect "$box_name")
expect_contains "$inspect_output" "created"
expect_contains "$inspect_output" "missing"

setup_output=$($box_binary setup "$box_name" --image ubuntu:24.04 --user boxuser --yes)
expect_contains "$setup_output" "Configured and created box \"$box_name\""
[[ $(podman inspect --format '{{.HostConfig.ReadonlyRootfs}}' "$container_name") == "false" ]] || fail "container root filesystem is read-only"
[[ $(podman inspect --format '{{.HostConfig.Privileged}}' "$container_name") == "false" ]] || fail "container unexpectedly runs privileged"

podman start "$container_name" >/dev/null
identity_output=$($box_binary exec "$box_name" -- sh -c 'printf "%s:%s:%s" "$(id -un)" "$(id -u)" "$HOME"')
IFS=: read -r identity_name identity_uid identity_home <<<"$identity_output"
[[ "$identity_name" == "boxuser" && "$identity_uid" != "0" && "$identity_home" == "/home/boxuser" ]] || fail "unexpected container identity: $identity_output"
[[ $($box_binary exec "$box_name" -- sudo -n id -u) == "0" ]] || fail "passwordless in-container sudo is unavailable"
$box_binary exec "$box_name" -- sh -c 'command -v curl git ip ping ps sudo >/dev/null'
$box_binary exec "$box_name" -- sudo -n apt-get update >/dev/null
$box_binary exec "$box_name" -- sudo -n apt-get install --yes --no-install-recommends bc >/dev/null
$box_binary exec "$box_name" -- sh -c 'command -v bc >/dev/null; printf home > "$HOME/e2e-home"; printf cache > "$HOME/.cache/e2e-cache"'

$box_binary stop "$box_name"
inspect_output=$($box_binary inspect "$box_name")
expect_contains "$inspect_output" "stopped"

recreate_output=$($box_binary setup "$box_name" --network none --yes)
expect_contains "$recreate_output" "Recreated box \"$box_name\""
[[ $(podman inspect --format '{{.HostConfig.NetworkMode}}' "$container_name") == "none" ]] || fail "network-none configuration was not applied"
podman start "$container_name" >/dev/null
persistent_output=$($box_binary exec "$box_name" -- sh -c 'printf "%s:%s" "$(cat "$HOME/e2e-home")" "$(cat "$HOME/.cache/e2e-cache")"')
[[ "$persistent_output" == "home:cache" ]] || fail "persistent data did not survive recreation: $persistent_output"

$box_binary stop "$box_name"
template_output=$($box_binary setup "$box_name" --network outbound --template ubuntu-24.04-terminal-tools --shell bash --yes)
expect_contains "$template_output" "ubuntu-24.04-terminal-tools"
expect_contains "$template_output" "Recreated box \"$box_name\""
podman start "$container_name" >/dev/null
$box_binary exec "$box_name" -- sh -c 'command -v bash jq nvim tmux rg >/dev/null'
persistent_output=$($box_binary exec "$box_name" -- bash -c 'printf "%s:%s" "$(cat "$HOME/e2e-home")" "$(cat "$HOME/.cache/e2e-cache")"')
[[ "$persistent_output" == "home:cache" ]] || fail "template recreation lost persistent data: $persistent_output"

if delete_output=$($box_binary delete "$box_name" 2>&1); then
  fail "delete without --purge unexpectedly succeeded"
fi
expect_contains "$delete_output" "refusing to delete box without --purge"
podman container exists "$container_name" || fail "failed delete removed the container"

$box_binary delete "$box_name" --purge
[[ ! -e "$HOME/.local/share/box/boxes/$box_name" ]] || fail "definition remains after purge"
if podman container exists "$container_name"; then
  fail "container remains after purge"
fi
if podman volume exists "$home_volume" || podman volume exists "$cache_volume"; then
  fail "managed volume remains after purge"
fi

printf 'e2e: all lifecycle and template checks passed\n'
