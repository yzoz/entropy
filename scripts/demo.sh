#!/usr/bin/env bash
# End-to-end demonstration of Entropy Money.
# Creates three participants, issues an IOU, countersigns it, transfers credit,
# settles an obligation, shows balances, the debt graph and CSV export.
# Runs entirely in a throwaway directory.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export ENTROPY_HOME="${ENTROPY_HOME:-$(mktemp -d)}"
PASS="${ENTROPY_DEMO_PASS:-demo}"

run() { go run "$ROOT/cmd/entropy" "$@"; }
step() { printf '\n\033[1;36m== %s ==\033[0m\n' "$1"; }

step "1. Generate keys for alice, bob and carol"
for p in alice bob carol; do
	run key gen -n "$p" --passphrase "$PASS" | sed -n '1,2p'
done

step "2. Register aliases"
for p in alice bob carol; do
	run id set --name "$p" --key "$p"
done
run id list

step "3. alice issues a 12.50 USD IOU to bob"
IOU=$(run issue --to bob --amount 12.50 --currency USD --key alice --passphrase "$PASS" | head -1 | awk '{print $1}')
echo "IOU id: $IOU"

step "4. Balances while the IOU is only a draft (unaffected)"
run balance

step "5. bob countersigns -> contract becomes signed"
run sign --id "$IOU" --key bob --passphrase "$PASS"
run balance
run verify

step "6. bob transfers 4.00 USD of credit to carol (TRANSFER)"
TR=$(run send --to carol --amount 4.00 --currency USD --key bob --passphrase "$PASS" | head -1 | awk '{print $1}')
run sign --id "$TR" --key carol --passphrase "$PASS"
run balance

step "7. Debt graph (Mermaid and DOT) — after transfer"
run graph
run graph -f dot

step "8. alice partially settles her IOU with bob (8.50 of 12.50)"
SETTLE=$(run settle --iou "$IOU" --amount 8.50 --key alice --passphrase "$PASS" | head -1 | awk '{print $1}')
echo "SETTLE id: $SETTLE"
run sign --id "$SETTLE" --key bob --passphrase "$PASS"
run balance
run verify
echo "-- remaining debt graph: 4.00 USD still owed --"
run graph

step "9. CSV report"
run export -f csv -o "$ENTROPY_HOME/report.csv"
cat "$ENTROPY_HOME/report.csv"

step "10. Export the ledger to a bundle and import it into a fresh ledger"
BUNDLE="$ENTROPY_HOME/bundle.json"
run export -o "$BUNDLE"
FRESH="$(mktemp -d)"
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" import "$BUNDLE"
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" verify
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" balance

step "Done. Demo data left in $ENTROPY_HOME"
