# Testing Guide — Aogeri API

This file provides concise, actionable test cases you can run manually or include in CI.

## Quick smoke (manual)

1. Start stack:

```bash
make docker-up
```

2. Health check:

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health
# expect: 200
```

3. Register / login / create stake (example):

```bash
curl -s -X POST http://localhost:8080/api/v1/register -H 'Content-Type: application/json' -d '{"email":"test+1@example.com","password":"Password1!","confirm_password":"Password1!"}' || true
RESP=$(curl -s -X POST http://localhost:8080/api/v1/login -H 'Content-Type: application/json' -d '{"email":"test+1@example.com","password":"Password1!"}')
TOKEN=$(echo "$RESP" | jq -r '.access_token')
curl -s -X POST http://localhost:8080/api/v1/stakes -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"token_symbol":"AOG","amount":"1","duration_days":30}'
```

## Unit test examples

- Test APY and reward calculation: write table-driven tests asserting expected floats.
- Run:

```bash
go test ./internal/services -v
```

## Integration / E2E test scenarios

1. Register -> Login -> Create stake -> List stakes
   - Assert: create returns 201 and list contains created stake

2. Create stake -> Unstake
   - Assert: unstake returns 200 and DB row `status` updated to `unstaked`

3. Dashboard and assets
   - Assert: seeded assets appear in GET /api/v1/assets

## Edge cases to include

- invalid amount (non-numeric) -> 400
- amount 0 or negative -> 400
- unsupported token_symbol -> 400
- duration < 30 -> 400
- unauthorized access -> 401

## Automation suggestions

- Add `scripts/smoke.sh` to run the quick smoke flow.
- Add GitHub Actions job to run `gofmt` and `go test ./...` on PRs.

If you want, I can add `scripts/smoke.sh` and a minimal GitHub Actions workflow now.
