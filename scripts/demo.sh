#!/usr/bin/env bash
# End-to-end demonstration of Entropy Money.
# Creates two participants, issues an IOU, countersigns it, settles it,
# shows balances and the debt graph. Runs entirely in a throwaway directory.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export ENTROPY_HOME="${ENTROPY_HOME:-$(mktemp -d)}"
PASS="${ENTROPY_DEMO_PASS:-demo}"

run() { go run "$ROOT/cmd/entropy" "$@"; }
step() { printf '\n\033[1;36m== %s ==\033[0m\n' "$1"; }

step "1. Generate keys for alice and bob"
run key gen -n alice --passphrase "$PASS"
run key gen -n bob --passphrase "$PASS"

step "2. Register aliases"
run id set --name alice --key alice
run id set --name bob --key bob
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

step "6. Debt graph (Mermaid) — bob is owed 12.50"
run graph
run graph -f dot

step "7. alice settles the IOU"
SETTLE=$(run settle --iou "$IOU" --key alice --passphrase "$PASS" | head -1 | awk '{print $1}')
echo "SETTLE id: $SETTLE"
run sign --id "$SETTLE" --key bob --passphrase "$PASS"
run balance
run verify
echo "-- after settlement the debt graph is empty --"
run graph

step "8. Export the ledger to a bundle"
BUNDLE="$ENTROPY_HOME/bundle.json"
run export -o "$BUNDLE"
echo "bundle written to $BUNDLE"

step "9. Import the bundle into a fresh ledger"
FRESH="$(mktemp -d)"
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" import "$BUNDLE"
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" verify
ENTROPY_HOME="$FRESH" go run "$ROOT/cmd/entropy" balance

step "Done. Demo data left in $ENTROPY_HOME"
