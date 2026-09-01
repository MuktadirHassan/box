# Changelog

## Unreleased

* Make the default Ubuntu environment writable and add passwordless in-container sudo with baseline development and networking tools.
* Keep development containers rootless and non-privileged while allowing package installation and system customization.
* Add catalog-driven, flat canonical template IDs.
* Enforce manifest-declared image, shell, and prompt compatibility and inject catalogs into consumers.
* Move built-in template assets to the repository-root `templates/<template-id>/` layout and embed them through the root asset package.

## [0.5.0](https://github.com/MuktadirHassan/box/compare/v0.4.0...v0.5.0) (2026-09-01)


### Features

* enhance tmux configuration with improved defaults and navigation shortcuts ([64d343a](https://github.com/MuktadirHassan/box/commit/64d343a5e4aca64c42d85549694614149c65dc66))

## [0.4.0](https://github.com/MuktadirHassan/box/compare/v0.3.0...v0.4.0) (2026-08-08)


### Features

* improve Ubuntu template and Podman E2E isolation ([#28](https://github.com/MuktadirHassan/box/issues/28)) ([6f72a1e](https://github.com/MuktadirHassan/box/commit/6f72a1eeeaf45a9e78e4eccdc64b302d03fce464))


### Bug Fixes

* add local bin path to fish config ([4e87ef3](https://github.com/MuktadirHassan/box/commit/4e87ef339a1635593f5749bd2950b88719671531))

## [0.3.0](https://github.com/MuktadirHassan/box/compare/v0.2.1...v0.3.0) (2026-08-08)


### Features

* add interactive writable mounts ([#26](https://github.com/MuktadirHassan/box/issues/26)) ([e780816](https://github.com/MuktadirHassan/box/commit/e780816a9047a7363cd36c209b6967ab295419f0))

## [0.2.1](https://github.com/MuktadirHassan/box/compare/v0.2.0...v0.2.1) (2026-08-08)


### Bug Fixes

* make Ubuntu template package installation work ([#24](https://github.com/MuktadirHassan/box/issues/24)) ([5dcce81](https://github.com/MuktadirHassan/box/commit/5dcce81ee67d06c83c1cc1d6f618668d2fbf8434))

## [0.2.0](https://github.com/MuktadirHassan/box/compare/v0.1.5...v0.2.0) (2026-08-08)


### Features

* make Ubuntu boxes mutable dev environments ([#23](https://github.com/MuktadirHassan/box/issues/23)) ([b67a491](https://github.com/MuktadirHassan/box/commit/b67a4911ad1c648c0970733e128f7d206dc41f96))


### Bug Fixes

* update go.mod and go.sum to remove indirect dependencies ([d8af9d8](https://github.com/MuktadirHassan/box/commit/d8af9d80e34df7abb3d3d21e888c5b175480ecfd))

## [0.1.5](https://github.com/MuktadirHassan/box/compare/v0.1.4...v0.1.5) (2026-08-08)


### Bug Fixes

* **refactor:** remove legacy template alias ([#18](https://github.com/MuktadirHassan/box/issues/18)) ([5296413](https://github.com/MuktadirHassan/box/commit/52964131073865b60f433984d298f1da66d3c6ce))

## [0.1.4](https://github.com/MuktadirHassan/box/compare/v0.1.3...v0.1.4) (2026-08-08)


### Bug Fixes

* allow periods in template IDs ([#15](https://github.com/MuktadirHassan/box/issues/15)) ([3ad97bb](https://github.com/MuktadirHassan/box/commit/3ad97bbed62c85751a8a4e0050f7edc9a2013d13))
* clean up release staging drafts ([88e7aaa](https://github.com/MuktadirHassan/box/commit/88e7aaafdefc535049614081c076842b450e6b51))

## [0.1.3](https://github.com/MuktadirHassan/box/compare/v0.1.2...v0.1.3) (2026-08-08)


### Bug Fixes

* allow release workflow retries ([2b0b48a](https://github.com/MuktadirHassan/box/commit/2b0b48ab3a42c942d80e0b250c56b983a6abbb29))
* build assets before publishing immutable releases ([3470cd6](https://github.com/MuktadirHassan/box/commit/3470cd6e76a04432fef5ee503a486bd1218a5d46))
* preserve release boundary without prior tag ([cb5b4eb](https://github.com/MuktadirHassan/box/commit/cb5b4eb0a2f4ee9c4c776bc6c2f9b3bb58864f53))

## [0.1.2](https://github.com/MuktadirHassan/box/compare/v0.1.1...v0.1.2) (2026-08-07)


### Bug Fixes

* install shell completions ([#7](https://github.com/MuktadirHassan/box/issues/7)) ([b5c7ef5](https://github.com/MuktadirHassan/box/commit/b5c7ef567575bc42af7dfc99fa72e7d1cb1e60ca))
* validate shell and template setup consistently ([#9](https://github.com/MuktadirHassan/box/issues/9)) ([652dc9f](https://github.com/MuktadirHassan/box/commit/652dc9f8ec0bbfe646bd482080642ab7319d1eb3))

## [0.1.1](https://github.com/MuktadirHassan/box/compare/v0.1.0...v0.1.1) (2026-08-06)


### docs

* reorganize README ([9cc679d](https://github.com/MuktadirHassan/box/commit/9cc679d881a914e5636a5a42405070eaedb2108a))


### Features

* add verified release installer ([#5](https://github.com/MuktadirHassan/box/issues/5)) ([887fe1f](https://github.com/MuktadirHassan/box/commit/887fe1f30e49f170172c41744c4b8f0a78de4c6d))

## 0.1.0 (2026-08-06)


### Features

* add .gitignore to exclude containers directory ([70dbb9d](https://github.com/MuktadirHassan/box/commit/70dbb9d10f79960bafdd4121f2d49148397bac29))
* add CI and release workflows for automated testing and deployment ([1fdb10f](https://github.com/MuktadirHassan/box/commit/1fdb10fce92bc2b9cff197a7749dae66ca6b8bad))
* add environment template option to configuration and update tests for output ([62090c1](https://github.com/MuktadirHassan/box/commit/62090c1059f3986a8081b571d17601b4d5c41079))
* add environment template options to configuration setup ([357ae7d](https://github.com/MuktadirHassan/box/commit/357ae7d58baff99580eb123009d7cc2416036655))
* add GoReleaser configuration, version management, and update README with installation instructions ([f229d10](https://github.com/MuktadirHassan/box/commit/f229d10d3adc76a50a664976db63bedd9a7a644f))
* add initial README with setup instructions and usage examples ([56ed7db](https://github.com/MuktadirHassan/box/commit/56ed7db521b673a3d9eb4788a07c927867db653e))
* add template configuration option and validation for CLI setup ([9d9d0c5](https://github.com/MuktadirHassan/box/commit/9d9d0c5c6b5da093612f0b3fb4542b94966039f0))
* add terminal-tools template with configuration and initialization scripts ([c987d1a](https://github.com/MuktadirHassan/box/commit/c987d1a5ea351f5c20c31ba22f16f8c4faa47172))
* enhance Delete method to support options for persistent data removal ([5c8d5d2](https://github.com/MuktadirHassan/box/commit/5c8d5d2a5accf1aa969c66dc9d7917a9879ecfe6))
* enhance template support for Arch Linux in CLI setup and update validation ([aedcd43](https://github.com/MuktadirHassan/box/commit/aedcd43f4f7f05061b456ddf8cfda0e203f8cea2))
* implement ShowRuntime and ShowList methods in Presenter, add tests for output alignment ([3cf5825](https://github.com/MuktadirHassan/box/commit/3cf5825a259dcdea7d41681e0f172c484c2f6e6e))
* refactor template validation and remove unused TemplatePackages function ([737131a](https://github.com/MuktadirHassan/box/commit/737131a6fac470d0284cc622151d3a5013ee76a9))
* remove obsolete container files and scripts for dev and ops environments ([9bf943f](https://github.com/MuktadirHassan/box/commit/9bf943f05646b3b4ce9045e3e05d0a088cb0a9d2))
* update CLI commands to support presenter for enhanced output ([fa36943](https://github.com/MuktadirHassan/box/commit/fa36943b7125be38168a6bc59702307e46f663ee))
* update go.mod and go.sum with additional indirect dependencies ([bb3493e](https://github.com/MuktadirHassan/box/commit/bb3493e1cde1b234265d8de2769b9f87542c5e42))
* update README for clarity and consistency in installation and usage instructions ([d189377](https://github.com/MuktadirHassan/box/commit/d1893774a12b5b4f53247d3d22cf49c84747a820))
* update setup command logic ([3ce7b75](https://github.com/MuktadirHassan/box/commit/3ce7b75fa268c8d51dc344f1784c27d662825541))
* update template handling in build process and remove deprecated functions ([168552b](https://github.com/MuktadirHassan/box/commit/168552bfd8b1583f52a853004809cf1e394847e7))
* update template validation to use string literal and improve command flag description ([764ef1a](https://github.com/MuktadirHassan/box/commit/764ef1a0feefa3f00d5e248d4f1e342735bc6a09))
* validate configuration and build template in Create method, add tests for template usage ([95ca45c](https://github.com/MuktadirHassan/box/commit/95ca45cfbf89a2c0178d16622ddc60bbcd6398af))


### Bug Fixes

* configure Release Please manifest packages ([#2](https://github.com/MuktadirHassan/box/issues/2)) ([7b16f2f](https://github.com/MuktadirHassan/box/commit/7b16f2fb1b30c533eb8e79263a80c733b6a4e1bb))
* reorganize go.mod dependencies for clarity ([ef97c55](https://github.com/MuktadirHassan/box/commit/ef97c55906a912363df368a0808877cdb84fb5dd))
* start releases at v0.1.0 ([#4](https://github.com/MuktadirHassan/box/issues/4)) ([c9ea46c](https://github.com/MuktadirHassan/box/commit/c9ea46cd8e057c33f764a19e035c7c13bd7c9f53))
* update Delete method signature to include box.Definition parameter ([fa7513c](https://github.com/MuktadirHassan/box/commit/fa7513c4797cd3a3414e356b2a25ccd29c610867))
