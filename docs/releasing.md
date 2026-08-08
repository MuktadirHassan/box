# Releasing

This guide covers the release contract for the CLI and Desktop alpha. Keep both products on one version stream: the Release Please manifest, changelog, Git tag, CLI archives, and `box-desktop` package metadata must all describe the same release version.

## Release flow

1. Merge conventional commits to `main` and let Release Please prepare and create the release.
2. Use the exact release SHA supplied by Release Please for every release check. Run `go test -tags webkit2_41 ./...`, `go vet -tags webkit2_41 ./...`, `go build -tags webkit2_41 ./...`, and the Podman-backed end-to-end suite (`scripts/e2e.sh`) against that SHA.
3. Only after those checks pass, run GoReleaser from the same SHA and tag to publish the release assets. Do not rebuild from a later `main` commit.

## Required assets and checks

Publish the CLI archives and checksum file, plus Linux amd64 Desktop alpha native packages named for `box-desktop`:

- Debian/Ubuntu: `box-desktop_*.deb`
- Fedora/RPM: `box-desktop_*.rpm`
- Arch/CachyOS: `box-desktop_*.pkg.tar.zst`

Automated release checks run the tagged test, vet, and build commands, smoke-test the unpackaged desktop binary, and run the Podman-backed end-to-end suite. The desktop smoke test does not install native packages or verify rootless Podman on target distributions.

Before publishing, verify that each expected asset is present, its package metadata names it `box-desktop`, and the package version matches the tag without its `v` prefix. Also verify the CLI archive checksums, CLI version output, release notes, and that no asset was built from a different commit.

For optional/manual target-distro verification, install each native package in its matching distribution image or runner. Confirm the desktop starts with rootless Podman available and the GTK3/WebKitGTK 4.1 runtime dependencies resolved by the package manager.

If a release fails after Release Please has made a draft or tag, inspect and clean the failed staging release, then start a new release attempt according to the workflow semantics. Do not mix assets from separate attempts.
