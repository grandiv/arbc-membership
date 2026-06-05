# arbc-membership — Deploy

Mirrors the truck-maintenance / to-ijul deploy model: build the whole stack with
Docker Compose on the target host, publish only the frontend container to
`127.0.0.1`, and front it with the host's nginx + Certbot for TLS.

> **Live:** https://arbc-membership.izcy.tech (izcy-engine, frontend on
> `127.0.0.1:3071`, Mongo `127.0.0.1:27049`, rest internal; Cloudflare edge +
> Let's Encrypt origin cert, HTTP→HTTPS). Deploy dir `/home/AgentZcy/arbc-membership`.
> Redeploy: `bash scripts/deploy.sh --target staging`.

## Prerequisites (remote host)

- Docker + Docker Compose v2
- An ssh alias to the host (e.g. `izcy-engine`)
- A host-level nginx + Certbot for the public domain
- This repo lives next to a `KreaZcy/` checkout **only on your dev machine** —
  the deploy bundles the needed KreaZcy bits into `.kzcy/`, so the remote does
  **not** need a separate KreaZcy checkout.

## One-command deploy

```bash
# 1. Point a target at your host:
cp scripts/targets/staging.env scripts/targets/prod.env   # edit REMOTE, REMOTE_DIR

# 2. Ship it (stages .kzcy → tar-over-ssh → compose build + up):
bash scripts/deploy.sh --target prod
```

`deploy.sh` runs `stage.sh` for you, so `.kzcy/` is fresh on every deploy.

## What the deploy does

1. `scripts/stage.sh` copies `KreaZcy/libs` + the four engine sources into
   `.kzcy/` (build context becomes self-contained).
2. tars this repo (incl. `.kzcy/`, excluding `node_modules`/`.next`/`.git`) and
   pipes it over ssh to `$REMOTE_DIR`. `COPYFILE_DISABLE=1` avoids macOS `._*`
   files that have crash-looped containers before.
3. `docker compose build && docker compose up -d` on the remote.
4. Prints `docker compose ps` so you can confirm health.

## Host nginx (TLS termination)

The frontend container listens on `127.0.0.1:3071`. Front it:

```nginx
server {
    server_name membership.tanaarabica.com;   # your domain
    location / {
        proxy_pass http://127.0.0.1:3071;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Then `certbot --nginx -d membership.tanaarabica.com` for HTTPS + HTTP→HTTPS.

## izcy-engine gotcha (Go proxy)

The VPS blocks Google's `proxy.golang.org` / `sum.golang.org`. Every Go
Dockerfile here already sets `GOPROXY=https://goproxy.io,direct` +
`GOSUMDB=off`, so `go mod download` won't hang. (Documented in
`izcy_engine_goproxy_block` memory.)

## Ports (host-bound, 127.0.0.1 only)

| Container | Host port | Notes |
|---|---|---|
| frontend (nginx) |  3071 | the only public-facing one |
| mongo | 27049 | for debugging/backup only |

Engines (8084/8082/5900) + BFF (8080) have **no** host port — internal network only.

## Backups

The member data lives in the `mongo_data` volume. Back it up:

```bash
ssh $REMOTE "docker exec arbc-mongo mongodump --archive" > backup-$(date +%F).archive
```

## Enabling email (NotifikaZcy) later

Currently `NOTIFY_ENABLED=false` (voucher code is returned in the API response).
To turn on email: add a `notifikazcy` service to the compose (Docker-only build —
its go.mod uses container `/libs` paths), set a Resend API key + verified sender,
and flip `NOTIFY_ENABLED=true` on the backend service.
