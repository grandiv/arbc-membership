# arbc-membership — Frontend

Next.js 16 + React 19 + Tailwind v4 app for **tana arabica**'s membership system.
Reuses the brand system (gold/cocoa/caramel, Montserrat + Inter) from
`arbc-landing-page`. Talks only to the BFF.

## Pages

| Route | Audience | Purpose |
|---|---|---|
| `/` | all | landing — links to the three flows |
| `/join` | consumer (QR at the stall) | data-capture form → issues + shows the voucher code |
| `/redeem` | barista | look up a member by phone, and redeem a voucher code |
| `/admin` | owner | stats, members, campaigns, create a campaign |

No consumer login — phone is the member key. `/join` is the only consumer
surface (reach it via a QR at the booth).

## Run

```bash
cp .env.example .env.local      # set NEXT_PUBLIC_API_URL to the BFF
pnpm install
pnpm dev                        # http://localhost:3000
```

> **Build-time env:** `NEXT_PUBLIC_API_URL` is inlined at `pnpm build`. For a
> deploy, set it at build time to the BFF's public URL (default
> `http://localhost:8080`). Changing it only at `pnpm start` has no effect on
> the client bundle.

## Status

Phase 3 complete: `pnpm build` is green (all routes static), and all pages serve
against the live BFF + engines. Real form submissions exercise the BFF's
`/api/register`, `/api/lookup`, `/api/redeem`, and `/api/admin/*` endpoints
(verified at the API layer). Wiring a QR poster → `/join` and deployment
(Docker/compose) are the remaining steps.
