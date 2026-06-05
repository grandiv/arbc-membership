# arbc-membership

Membership / data-capture system for **tana arabica** — a per-client product
that wires generic [KreaZcy](https://github.com/KreaZcy) engines behind a BFF +
Next.js frontend. Same pattern as `to-ijul` and `truck-maintenance`.

**The loop:** member gives data → gets a voucher (e.g. "200 free coffee") →
redemptions deepen the profile for loyalty. The voucher is gated on the data.

> Design docs: `KreaZcy/docs/Projects/arbc-membership/` (`TECHNICAL_PLAN.md`,
> `CHECKLIST.md`, `ENGINE_STANDARDS_AUDIT.md`).

## Stack

```
            Browser (Next.js static, served by nginx)
                          │  /api/* (same-origin)
                    ┌─────▼─────┐
                    │  BFF (Go) │  :8080  brand vocabulary lives here
                    └─┬───┬───┬─┘
            ┌─────────┘   │   └─────────┐
       ┌────▼────┐  ┌─────▼────┐  ┌─────▼─────┐
       │KonsumZcy│  │ PromoZcy │  │ AgregaZcy │   (NotifikaZcy: stubbed)
       │ members │  │ vouchers │  │ analytics │
       │  :8084  │  │  :8082   │  │   :5900   │
       └────┬────┘  └─────┬────┘  └─────┬─────┘
            └───────── Mongo ───────────┘
```

Only the frontend container is published (`127.0.0.1:3070`); everything else is
internal to the docker network.

| Dir | What |
|---|---|
| `backend/` | Go + Gin BFF — see `backend/README.md` |
| `frontend/` | Next.js 16 + React 19 + Tailwind v4 — see `frontend/README.md` |
| `*.Dockerfile` | one per service |
| `docker-compose.yml` | full stack |
| `nginx.conf` | frontend nginx (serves static + proxies `/api`) |
| `scripts/stage.sh` | copies KreaZcy libs + engine sources into `.kzcy/` for the build |
| `scripts/deploy.sh` | tar-over-ssh deploy |

## Run the whole thing locally (Docker)

```bash
bash scripts/stage.sh        # populate .kzcy/ from ../../KreaZcy
docker compose build
docker compose up -d
open http://localhost:3070
docker compose down -v       # stop + wipe data
```

The engines' Go modules `replace` to `../../../libs/*`, which live outside the
repo; `stage.sh` copies them into `.kzcy/` so the Docker build context is
self-contained. **Re-run `stage.sh` after pulling KreaZcy changes.**

## Develop without Docker

Run Mongo + the three engines + the BFF natively (see `backend/README.md`),
then `cd frontend && pnpm dev`.

## Deploy

See [DEPLOY.md](./DEPLOY.md). On the izcy-engine VPS, builds set
`GOPROXY=goproxy.io` + `GOSUMDB=off` (Google's Go proxy is blocked there) —
already baked into every Dockerfile.

## Status

Phases 1–4 done: engines (incl. a KonsumZcy `POST /api/customers` + PromoZcy
standards pass), BFF, frontend, and Docker compose — **all six containers build,
boot healthy, and the full register→redeem loop + analytics events are verified
end-to-end through nginx**. NotifikaZcy email delivery is stubbed
(`NOTIFY_ENABLED=false`, code returned in-band). Remaining: real email, a host
nginx + TLS in front, and the `/admin/analytics` deep-dive.
