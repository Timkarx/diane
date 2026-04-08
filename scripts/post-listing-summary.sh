#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
URL="${1:-http://localhost:5555/listing-summary}"

payload="$(perl -0MJSON::PP -e 'print encode_json({listing => do { local $/; <> }})' "$ROOT/test/good_prompt.md")"

curl -sS -X POST "$URL" \
  -H 'Content-Type: application/json' \
  --data "$payload"
