#!/usr/bin/env bash
# Stage the KreaZcy libs + engine sources this project depends on into ./.kzcy/
# so the Docker build context is self-contained (the Go replace dirs point
# outside the repo otherwise). Run before `docker compose build`.
#
# Idempotent. .kzcy/ is gitignored. Override the KreaZcy location with KZCY_ROOT.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KZCY_ROOT="${KZCY_ROOT:-$HERE/../../KreaZcy}"
DEST="$HERE/.kzcy"

if [ ! -d "$KZCY_ROOT/libs" ]; then
  echo "ERROR: KreaZcy not found at $KZCY_ROOT (set KZCY_ROOT)" >&2
  exit 1
fi

echo "Staging KreaZcy from $KZCY_ROOT → $DEST"
rm -rf "$DEST"
mkdir -p "$DEST/libs" "$DEST/services"

# Shared libs (all of them — engines reference several).
cp -R "$KZCY_ROOT/libs/." "$DEST/libs/"

# Engine sources used by this product.
for svc in \
  "KonsumZcy/KonsumZcy" \
  "PromoZcy/PromoZcy" \
  "AgregaZcy/AgregaZcy" \
  "AgregaZcy/AgregaZcy-BI-Go" ; do
  mkdir -p "$DEST/services/$(dirname "$svc")"
  cp -R "$KZCY_ROOT/services/$svc" "$DEST/services/$(dirname "$svc")/"
done

# Drop heavy/irrelevant subtrees to keep the context lean.
find "$DEST" -type d \( -name node_modules -o -name .git -o -name dist -o -name .next \) -prune -exec rm -rf {} + 2>/dev/null || true

echo "Staged. Now: docker compose build"
