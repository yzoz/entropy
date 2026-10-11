# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-11

First public release of **Entropy Money** — a CLI for partner mutual credit
(IOU) built on signed contracts.

### Added

- **Keys & identities**
  - `key gen` creates an Ed25519 key pair protected by a passphrase.
  - Private keys are encrypted with Argon2id (KDF) + AES-256-GCM.
  - Base58 participant addresses derived from public keys (`ed25519:...`).
  - `id set` / `id list` manage human-readable aliases.
- **Contracts**
  - Contract types: `IOU` (debt note), `TRANSFER` (credit transfer),
    `SETTLE` (repayment / netting).
  - Lifecycle: `draft → signed → settled → expired`, effective only after
    both parties countersign.
  - Amounts stored as integer minor units (no floating point); currencies
    `USD`, `EUR`, `RUR`, `GLD`, `SLV`, `OIL` with per-currency scales.
- **Ledger**
  - bbolt-backed local store for contracts, aliases and public keys.
  - Balances and the debt graph are computed projections of the ledger;
    balances can go negative (mutual credit, no built-in limits).
- **Commands**: `key gen`, `id set|list`, `issue`, `send`, `sign`, `settle`,
  `view`, `balance`, `verify`, `graph`, `export`, `import`, `version`.
- **Audit**: `verify` re-checks every signature, status and reference and
  confirms that balances net to zero per currency.
- **Exchange**: `export`/`import` of portable JSON bundles; imports verify
  everything before writing and are idempotent. `export -f csv` produces a
  human-readable report.
- **Output**: human-readable by default, `--json` for scripting; global
  `--home` flag and `$ENTROPY_HOME` support.
- **Tooling**: Makefile, end-to-end demo script, GitHub Actions CI
  (fmt/vet/test/build + golangci-lint), GoReleaser release pipeline.

[Unreleased]: https://github.com/yzoz/entropy/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yzoz/entropy/releases/tag/v1.0.0
