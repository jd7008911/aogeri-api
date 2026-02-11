#!/usr/bin/env bash
set -euo pipefail

BASE=http://localhost:8080

echo "Running smoke tests against $BASE"

# register (idempotent)
curl -s -X POST $BASE/api/v1/register -H 'Content-Type: application/json' -d '{"email":"smoke+1@example.com","password":"Password1!","confirm_password":"Password1!"}' || true

# login
RESP=$(curl -s -X POST $BASE/api/v1/login -H 'Content-Type: application/json' -d '{"email":"smoke+1@example.com","password":"Password1!"}')
TOKEN=$(echo "$RESP" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Failed to obtain access token" >&2
  echo "$RESP" >&2
  exit 1
fi

echo "Got token, creating stake..."
curl -s -X POST $BASE/api/v1/stakes -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"token_symbol":"AOG","amount":"1","duration_days":30}'

echo "Smoke tests completed"
