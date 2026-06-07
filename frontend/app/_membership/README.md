# Membership screens (HIDDEN — preserved for the future)

This folder is a **Next.js private folder** (the `_` prefix excludes it from routing),
so these screens are **in the repo but NOT accessible by URL**. They are the
full-membership UIs, parked while the product runs in **campaign-only mode**
(the "200 Kopi Gratis" launch).

The membership **capability is fully alive at the backend** — the BFF still exposes
`/api/register`, `/api/redeem`, `/api/lookup`, and the full `/api/admin/*`. The
campaign's `/api/claim` is just a thin composition over `register` + `redeem`.
These components are alternate front-ends over those same endpoints.

## What's here
- `JoinScreen.tsx` — consumer self-signup (name/phone/email/IG/DOB), decoupled
  from the campaign (registration is its own concern; the free-cup perk is shown
  separately via `GET /api/campaign`).
- `RedeemScreen.tsx` — barista flow: look up a member by phone, redeem by phone
  (auto-applies the active campaign) or by explicit code; price-aware.
- `MembersDashboard.tsx` — full admin: member list (with email/visits/spend),
  campaign management with configurable discounts (free / percent / fixed Rp).

## To re-enable membership mode
1. Move the screen(s) back into routed folders, e.g.
   `app/(member)/join/page.tsx`, `app/redeem/page.tsx`, and restore the richer
   `app/admin/page.tsx`.
2. Re-add the consumer nav links in `components/Header.tsx`
   (Daftar / Tukar / Admin).
3. The BFF needs no change — the endpoints are already there.

> Campaign-vs-membership is currently a hard fork (this folder). If you later want
> a runtime switch, promote it to a `NEXT_PUBLIC_MODE` feature flag.
