# Entropy Money

[![CI](https://github.com/yzoz/entropy/actions/workflows/ci.yml/badge.svg)](https://github.com/yzoz/entropy/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yzoz/entropy)](https://goreportcard.com/report/github.com/yzoz/entropy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Signed mutual-credit contracts from the command line.**

Entropy Money is a small, self-contained Go CLI for **partner mutual credit**.
Participants issue, countersign and settle debt obligations (IOU) as portable
signed documents, and every node keeps its own local ledger. Balances are not
stored: they are *computed* from the set of signed contracts, which is the
single source of truth.

> This is **not** a blockchain. There is no mining, no consensus and no central
> server. Participants exchange contracts as explicit export/import bundles.

---

## Why it is interesting

- **Real cryptography**: Ed25519 signatures over a canonical contract payload;
  private keys are stored encrypted with **Argon2id + AES-256-GCM**.
- **A proper lifecycle**: `draft → signed → settled → expired`, where a contract
  only becomes effective once *both* parties sign it.
- **Derived state**: balances and the debt graph are projections of the ledger,
  recomputed from scratch on every run.
- **Auditable**: `entropy verify` re-checks every signature, status and reference
  and confirms that balances net to zero.
- **No float money**: amounts are integers in minor units (cents, etc.).

## Features

- Passphrase-protected Ed25519 key management.
- Human-readable **aliases** for participant addresses.
- Contract types: **IOU** (debt note), **TRANSFER** (credit transfer),
  **SETTLE** (repay/nets an IOU).
- Currencies: `USD`, `EUR`, `RUR`, `GLD`, `SLV`, `OIL` (each with its own scale).
- Portable **export/import** bundles that are verified before being trusted, and
  idempotent on re-import.
- Signed-balance **audit** (`verify`) and **debt-graph** rendering
  (Mermaid / DOT).
- Human-readable output by default, `--json` for scripting.

## Install

### Prebuilt binaries

Grab the archive for your OS/arch from the
[releases page](https://github.com/yzoz/entropy/releases), unpack and move
`entropy` onto your `PATH`. See [INSTALL.md](./INSTALL.md) for details.

### With Go

```bash
go install github.com/yzoz/entropy/cmd/entropy@v1.0.0
```

### From source

```bash
git clone https://github.com/yzoz/entropy
cd entropy
make build          # -> ./bin/entropy
```

Requires Go 1.18+.

## Quick start

Data lives in `~/.entropy` (override with `--home` or `$ENTROPY_HOME`).

```bash
# 1. create two identities
entropy key gen -n alice --passphrase secret
entropy key gen -n bob   --passphrase secret
entropy id set --name alice --key alice
entropy id set --name bob   --key bob

# 2. alice issues a 12.50 USD IOU to bob (a draft)
entropy issue --to bob --amount 12.50 --currency USD --key alice --passphrase secret

# 3. bob countersigns -> the contract becomes effective
entropy sign --id <IOU_ID> --key bob --passphrase secret

# 4. inspect
entropy balance
entropy view
entropy verify
entropy graph

# 5. alice repays the IOU
entropy settle --iou <IOU_ID> --key alice --passphrase secret
entropy sign --id <SETTLE_ID> --key bob --passphrase secret

# 6. move the ledger elsewhere
entropy export -o ledger.bundle.json
entropy --home /path/to/peer import ledger.bundle.json

# 7. produce a human-readable report
entropy export -f csv -o ledger.csv
```

Run the whole story at once:

```bash
make demo
```

## How balances work

Each contract contributes a **delta** to its parties' balances:

| Type       | From      | To        |
|------------|-----------|-----------|
| `IOU`      | `−amount` | `+amount` |
| `TRANSFER` | `−amount` | `+amount` |
| `SETTLE`   | `+amount` | `−amount` |

Only contracts in the `signed` state count. Because every obligation creates an
equal and opposite delta, a consistent ledger always nets to **zero** per
currency — which is exactly what `entropy verify` checks. A negative balance
simply means the participant owes value; the credit model has no built-in limit,
the risk is taken by whoever signs.

## Command reference

| Command | Description |
|---------|-------------|
| `key gen -n NAME` | Create a passphrase-protected Ed25519 key |
| `id set --name ALIAS --key KEY` | Bind an alias to an address |
| `id list` | List aliases |
| `issue --to PARTY --amount N -c CUR` | Issue an IOU |
| `send --to PARTY --amount N -c CUR` | Transfer credit (TRANSFER) |
| `sign --id ID` | Countersign a received contract |
| `settle --iou ID` | Create a SETTLE for an IOU |
| `view [ID]` | List or inspect contracts |
| `balance` | Show computed net balances |
| `verify` | Audit the ledger |
| `graph [-f mermaid\|dot]` | Render the debt graph |
| `export -o FILE [-f json\|csv]` | Export a verifiable bundle (json) or a report (csv) |
| `import FILE` | Verify and import a bundle |
| `version` | Print build version information |

Global flags: `--home DIR`, `--json`.

## Architecture

```
cmd/entropy/          entry point
internal/domain/      contract model, currencies, canonical signing payload
internal/crypto/      keys, addresses, signing, encrypted keystore
internal/store/       bbolt persistence (contracts, aliases, public keys)
internal/sign/        applies signatures and drives the lifecycle
internal/ledger/      balance and debt-graph projections
internal/bundle/      verified export/import
internal/verify/      ledger audit
internal/graph/       Mermaid / DOT rendering
internal/cli/         cobra command tree
```

## Development

```bash
make test        # go test ./...
make test-race   # go test -race ./...
make lint        # gofmt + go vet
make demo        # end-to-end scenario in a temp dir
make snapshot    # build release artifacts locally (GoReleaser)
```

CI runs formatting, vet, tests and build on every push, plus `golangci-lint`
(see `.github/workflows/ci.yml`). Pushing a `v*` tag triggers the release
workflow, which builds cross-platform binaries with GoReleaser and publishes
them together with `checksums.txt` (see `.github/workflows/release.yml`).

## Releasing

Maintainers publish a new version by tagging:

```bash
git tag -a v1.0.0 -m "Entropy Money v1.0.0"
git push origin v1.0.0
```

The tag starts `.github/workflows/release.yml`, which cross-compiles for
Linux/macOS/Windows (amd64/arm64), embeds version metadata, and creates the
GitHub release. See [CHANGELOG.md](./CHANGELOG.md) for the version history.

## Проект по-русски

**Entropy Money** — консольная утилита на Go для **партнёрских взаимозачётов
на подписанных контрактах (IOU / mutual credit)**. Участники выпускают,
встречно подписывают и погашают долговые обязательства, обмениваясь ими через
файлы-пакеты (`export`/`import`). Каждый узел хранит собственный реестр
(bbolt), а балансы и граф долгов не хранятся, а **вычисляются** из реестра.

Это **не блокчейн**: без майнинга, консенсуса и центрального сервера.
Ключи Ed25519 защищены паролем (Argon2id + AES-256-GCM), суммы — целые
минорные единицы, контракт вступает в силу только после подписи обеих сторон.
`entropy verify` независимо проверяет все подписи, статусы и ссылки и
подтверждает, что балансы сходятся в ноль.
