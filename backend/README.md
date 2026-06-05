# arbc-membership — BFF

Per-client Backend-for-Frontend for **tana arabica**'s membership / data-capture
system. Wires the generic KreaZcy engines behind one API the frontend talks to.
Same pattern as `to-ijul` and `truck-maintenance`.

> Full design: [`docs/Projects/arbc-membership/TECHNICAL_PLAN.md`](../../../KreaZcy/docs/Projects/arbc-membership/TECHNICAL_PLAN.md) in the KreaZcy workspace.

## Engines wired

| Engine | Role | Default URL |
|---|---|---|
| KonsumZcy | member store (`Customer`) | `:8084` |
| PromoZcy | vouchers / campaigns | `:8082` |
| NotifikaZcy | voucher delivery (email) | `:8086` |
| AgregaZcy | analytics timeline + forecasting | `:5900` |

## API

| Method | Route | Purpose |
|---|---|---|
| `GET`  | `/health` | health check |
| `POST` | `/api/register` | capture member data → issue voucher → deliver → log events |
| `POST` | `/api/lookup` | barista: find member by phone |
| `POST` | `/api/redeem` | barista: validate + apply a voucher code |
| `GET`  | `/api/admin/members` | list members (proxies KonsumZcy) |
| `GET`  | `/api/admin/campaigns` | list campaigns (proxies PromoZcy) |
| `POST` | `/api/admin/campaigns` | create a campaign (e.g. "200 free coffee") |

All brand vocabulary lives here; engines stay neutral. Upstream errors are
sanitized (4xx passes through, 5xx/transport → `502`). AgregaZcy events are
best-effort (never block a request).

## Run locally

```bash
cp .env.example .env
# Bring up the engines (each needs Mongo). For a quick local spin:
#   docker run -d --rm -p 27021:27017 mongo:7
#   KonsumZcy:  AUTH_METHOD=none MONGO_URI=mongodb://localhost:27021 go run ./cmd/server   (:8084)
#   PromoZcy:   AUTH_METHOD=none MONGO_URI=mongodb://localhost:27021 go run ./cmd/server   (:8082)
#   AgregaZcy:  AUTH_METHOD=none MONGO_URI=mongodb://localhost:27021 go run ./cmd/server   (:5900)
GOWORK=off go run ./cmd/server   # :8080
```

`go.mod` references KreaZcy libs via local `replace` directives
(`../../../KreaZcy/libs/*`); build with `GOWORK=off`.

## Status

Phase 2 complete: builds, and the full **register → redeem → re-redeem-blocked**
loop plus AgregaZcy analytics events are verified end-to-end against live engines.
NotifikaZcy delivery is stubbed (`NOTIFY_ENABLED=false` returns the code in-band).
Frontend (Next.js) is Phase 3.
