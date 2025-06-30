# Changelog

## [0.3.2](https://github.com/elct9620/ccmon/compare/v0.3.1...v0.3.2) (2025-06-30)


### Bug Fixes

* add pull-requests permission for Homebrew PR creation ([de210a1](https://github.com/elct9620/ccmon/commit/de210a151e3d9eea14eb11e2e50a885a033ce6e0))

## [0.3.1](https://github.com/elct9620/ccmon/compare/v0.3.0...v0.3.1) (2025-06-30)


### Bug Fixes

* replace direct Homebrew push with pull request workflow ([ce8cde8](https://github.com/elct9620/ccmon/commit/ce8cde82e6cd8b2847b947061955c240be361544))

## [0.3.0](https://github.com/elct9620/ccmon/compare/v0.2.1...v0.3.0) (2025-06-30)


### Features

* homebrew formula ([#10](https://github.com/elct9620/ccmon/issues/10)) ([950546a](https://github.com/elct9620/ccmon/commit/950546af66e8665710d4f10703bb3dae57e0ab13))
* implement auto-width table columns for better model name display ([9ba25bd](https://github.com/elct9620/ccmon/commit/9ba25bdb974017505f95d92ff53df97aba1945d1))

## [0.2.1](https://github.com/elct9620/ccmon/compare/v0.2.0...v0.2.1) (2025-06-30)


### Bug Fixes

* simplify time block logic with automatic advancement ([8ae4efc](https://github.com/elct9620/ccmon/commit/8ae4efcebce9b9fa19f846d7e2f2521e5443c1f1))

## [0.2.0](https://github.com/elct9620/ccmon/compare/v0.1.4...v0.2.0) (2025-06-29)


### Features

* add logging for received API requests in OTLP receiver ([319af85](https://github.com/elct9620/ccmon/commit/319af85174df7959158dba4581d97cf440174b73))

## [0.1.4](https://github.com/elct9620/ccmon/compare/v0.1.3...v0.1.4) (2025-06-29)


### Bug Fixes

* grant contents write permission to build-and-push job for GitHub releases ([d7e0982](https://github.com/elct9620/ccmon/commit/d7e0982e63be3713fc6be8d9ca6815366e0111fe))

## [0.1.3](https://github.com/elct9620/ccmon/compare/v0.1.2...v0.1.3) (2025-06-29)


### Bug Fixes

* add required permissions for artifact attestation in release workflow ([93b2ec1](https://github.com/elct9620/ccmon/commit/93b2ec19bff806139e24581d6768dc2e83ba4104))

## [0.1.2](https://github.com/elct9620/ccmon/compare/v0.1.1...v0.1.2) (2025-06-29)


### Bug Fixes

* add GoReleaser-specific Dockerfile for pre-built binary Docker builds ([20a269f](https://github.com/elct9620/ccmon/commit/20a269f9cfcdbb8aed2af8554fed420b73480a56))
* update GoReleaser config to v2 format and remove deprecated options ([631697b](https://github.com/elct9620/ccmon/commit/631697b644f2fec5bca013b886f20339fff35016))

## [0.1.1](https://github.com/elct9620/ccmon/compare/v0.1.0...v0.1.1) (2025-06-29)


### Bug Fixes

* install protoc in release workflow to fix GoReleaser build failure ([085f66b](https://github.com/elct9620/ccmon/commit/085f66b146356b39ce74c3cb29eaf5b05ad9e67d))

## 0.1.0 (2025-06-29)


### Features

* add Claude token limit progress bar with 5-hour block tracking ([caf70eb](https://github.com/elct9620/ccmon/commit/caf70eb190d6714ea5e5f4e49d75093f7b591d4e))
* add configurable monitor refresh interval ([6bdd27c](https://github.com/elct9620/ccmon/commit/6bdd27ccb2dcf20f9e365f462afdf05ec898825f))
* add configuration system with Viper ([add8cf7](https://github.com/elct9620/ccmon/commit/add8cf73c5a94a778f5e330479fe1e6f230570d8))
* add efficient query limiting to reduce server load and improve TUI performance ([48a9fd8](https://github.com/elct9620/ccmon/commit/48a9fd879a13e6fc6a114c9eda8f951f1174676e))
* add gRPC query service to enable distributed monitor/server architecture ([e9a77e4](https://github.com/elct9620/ccmon/commit/e9a77e42d0ec0986f727859b97dadd6517f4841d))
* add local data persistence with time-based filtering ([d26f116](https://github.com/elct9620/ccmon/commit/d26f116fda2b2f1a86814cce84cb403b1a003ce0))
* add minimal OTEL gRPC receiver for Claude Code telemetry ([a133ba0](https://github.com/elct9620/ccmon/commit/a133ba052062bf0b9b8e0beb857dd86b00dea4ff))
* add pflag support for command-line configuration overrides ([41be9d4](https://github.com/elct9620/ccmon/commit/41be9d44a3b414364d7d21958ea95b5a8f4ba04b))
* add production Dockerfile with multi-stage build ([6f72f06](https://github.com/elct9620/ccmon/commit/6f72f06b89d474e045ec2d6f522517fff05217ef))
* add release automation with release-please and goreleaser ([a236f4e](https://github.com/elct9620/ccmon/commit/a236f4e27d44a9c30bbbf2b29ca2665f7e73326a))
* add separate usage tracking for base (Haiku) vs premium models ([f5d198f](https://github.com/elct9620/ccmon/commit/f5d198ff19d78e2f83fee8f85d6eba7ddb9bf4c5))
* add timezone support and migrate from TimeFilter to Period-based queries ([2651177](https://github.com/elct9620/ccmon/commit/265117789ded83eb8da6d749637044c8548d21fb))
* add token usage tracking for Claude Code API requests ([88f53cd](https://github.com/elct9620/ccmon/commit/88f53cd1b2557530bbf7de0b61c3fd2eab9dd79f))
* display usage statistics in table format with token categorization ([6c61f1f](https://github.com/elct9620/ccmon/commit/6c61f1f96e57934d0d5fab6425c38276a6640a70))
* implement dual-mode architecture with read-only database access ([8255eee](https://github.com/elct9620/ccmon/commit/8255eee2ffac7928723fc99ead3beb17d1a27db2))


### Bug Fixes

* add issues write permission and fix file ending for release workflow ([ba3cc01](https://github.com/elct9620/ccmon/commit/ba3cc01d740ee3cdc1c33900d79116e2b627f9a6))
* bind Docker server to 0.0.0.0:4317 for external access ([d7f3ad0](https://github.com/elct9620/ccmon/commit/d7f3ad02c4f4eaf1daa7307e7ff5f68d84fecb3f))
* correct license description in README and add LICENSE file ([5af2183](https://github.com/elct9620/ccmon/commit/5af2183462bdff6f4da130ea86c365593549a6d9))
* handle error return values in gRPC test cleanup ([34ed38c](https://github.com/elct9620/ccmon/commit/34ed38ca714c8e9a985d893cdc2fbfbc30729d78))
* improve monitor progress bar accuracy and timezone handling ([1fcbdd3](https://github.com/elct9620/ccmon/commit/1fcbdd3f348393f5bdf6070b017fc3e2d3ae7e27))
* prevent TUI height overflow with conservative height calculations ([f741cdf](https://github.com/elct9620/ccmon/commit/f741cdf6817004907a38f28594ed70b8259e4067))
* resolve all golangci-lint issues with proper error handling ([4a53007](https://github.com/elct9620/ccmon/commit/4a530072ca9a87323d6376584e466cb08acf0ccd))
* separate statistics display from block progress calculation ([a024404](https://github.com/elct9620/ccmon/commit/a0244049d39ed417df457de24f63fa5ebb126aad))


### Miscellaneous Chores

* release as 0.1.0 ([bc9fb2a](https://github.com/elct9620/ccmon/commit/bc9fb2a82b49261799288c093dbe35ca6133b0b0))
