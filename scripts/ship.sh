#!/usr/bin/env bash
# Build the stack's images LOCALLY for linux/amd64 and ship them to the remote,
# which then only loads + restarts — it never compiles. Use this instead of
# scripts/deploy.sh on the reduced-spec VPS, where building 4 Go services at
# once OOM-kills the box.
#
# Usage:
#   bash scripts/ship.sh                  # uses scripts/targets/staging.env
#   bash scripts/ship.sh --target prod
set -euo pipefail

TARGET="${ARBC_DEPLOY_TARGET:-staging}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --target) TARGET="$2"; shift 2 ;;
    -h|--help) sed -n '2,12p' "$0"; exit 0 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TARGET_FILE="$REPO_ROOT/scripts/targets/${TARGET}.env"
[[ -f "$TARGET_FILE" ]] || { echo "ERROR: no target file $TARGET_FILE" >&2; exit 1; }
# shellcheck disable=SC1090
source "$TARGET_FILE"
: "${REMOTE:?target must set REMOTE}"
: "${REMOTE_DIR:?target must set REMOTE_DIR}"

PLATFORM="linux/amd64"
PROJECT="arbc-membership"   # compose project → image prefix (arbc-membership-<svc>)
cd "$REPO_ROOT"

echo "→ staging KreaZcy into .kzcy/"
bash "$REPO_ROOT/scripts/stage.sh"

# service:dockerfile pairs (mongo is a public image, not built)
SERVICES=(
  "konsumzcy:konsumzcy.Dockerfile"
  "promozcy:promozcy.Dockerfile"
  "agregazcy:agregazcy.Dockerfile"
  "backend:backend.Dockerfile"
  "frontend:frontend.Dockerfile"
)

IMAGES=()
for entry in "${SERVICES[@]}"; do
  svc="${entry%%:*}"; dockerfile="${entry##*:}"
  img="${PROJECT}-${svc}:latest"
  echo "→ building $img ($PLATFORM) from $dockerfile"
  docker buildx build --platform "$PLATFORM" --load \
    -t "$img" -f "$dockerfile" "$REPO_ROOT"
  IMAGES+=("$img")
done

echo "→ ensuring remote dir + compose files are present"
# Ship just the files the remote needs to RUN (compose + nginx); no source.
ssh "$REMOTE" "mkdir -p $REMOTE_DIR"
scp -q "$REPO_ROOT/docker-compose.yml" "$REMOTE:$REMOTE_DIR/docker-compose.yml"
[[ -f "$REPO_ROOT/nginx.conf" ]] && scp -q "$REPO_ROOT/nginx.conf" "$REMOTE:$REMOTE_DIR/nginx.conf"

echo "→ saving + streaming ${#IMAGES[@]} images to $REMOTE (gzip over ssh)"
docker save "${IMAGES[@]}" | gzip | ssh "$REMOTE" "gunzip | docker load"

echo "→ recreating containers from loaded images (no build on remote)"
ssh "$REMOTE" "cd $REMOTE_DIR && docker compose up -d --no-build && sleep 6 && docker compose ps"

echo "✓ shipped. Front the published 127.0.0.1:3071 with your host nginx + TLS."
