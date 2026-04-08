#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
URL="${1:-http://localhost:5555/listing-summary}"
LISTING_LINK="${2:-https://example.com/listings/good-prompt}"

payload="$(perl -0MJSON::PP -e 'my $link = shift @ARGV; print encode_json({listing => do { local $/; <> }, link => $link})' "$LISTING_LINK" "$ROOT/test/good_prompt.md")"

curl -sS -X POST "$URL" \
  -H 'Content-Type: application/json' \
  --data "$payload"
