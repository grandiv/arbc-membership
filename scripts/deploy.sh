#!/usr/bin/env bash
# Deploy the arbc-membership stack to a remote host via tar-over-ssh.
#
# Usage:
#   bash scripts/deploy.sh                     # uses scripts/targets/staging.env
#   bash scripts/deploy.sh --target prod       # uses scripts/targets/prod.env
#
# A target = scripts/targets/<name>.env setting:
#   REMOTE          ssh host alias (e.g. izcy-engine)
#   REMOTE_DIR      deploy root on the remote (e.g. /home/AgentZcy/arbc-membership)
#
# Because .kzcy/ is staged locally first, the build context is self-contained —
# we only push THIS repo (no separate KreaZcy push needed).
set -euo pipefail

TARGET="${ARBC_DEPLOY_TARGET:-staging}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --target) TARGET="$2"; shift 2 ;;
    -h|--help) sed -n '2,14p' "$0"; exit 0 ;;
    *) echo "unknown arg: $1" >&2; exit 2 ;;
  esac
done

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TARGET_FILE="$REPO_ROOT/scripts/targets/${TARGET}.env"
[[ -f "$TARGET_FILE" ]] || { echo "ERROR: no target file $TARGET_FILE (copy targets/staging.env)" >&2; exit 1; }
# shellcheck disable=SC1090
source "$TARGET_FILE"
: "${REMOTE:?target must set REMOTE}"
: "${REMOTE_DIR:?target must set REMOTE_DIR}"

echo "→ staging KreaZcy into .kzcy/"
bash "$REPO_ROOT/scripts/stage.sh"

echo "→ deploying target=$TARGET → $REMOTE:$REMOTE_DIR"
ssh "$REMOTE" "mkdir -p $REMOTE_DIR && rm -rf $REMOTE_DIR/backend $REMOTE_DIR/frontend $REMOTE_DIR/.kzcy"

# tar the project (includes the freshly-staged .kzcy/), excluding heavy junk.
# COPYFILE_DISABLE stops macOS from injecting ._* AppleDouble files.
COPYFILE_DISABLE=1 tar \
  --exclude='**/node_modules' --exclude='**/.next' --exclude='**/out' \
  --exclude='**/.git' --exclude='**/.DS_Store' \
  -czf - -C "$REPO_ROOT" . | ssh "$REMOTE" "tar -xzf - -C $REMOTE_DIR"

echo "→ building + starting on remote"
ssh "$REMOTE" "cd $REMOTE_DIR && docker compose build && docker compose up -d"

echo "→ waiting for health"
ssh "$REMOTE" "cd $REMOTE_DIR && sleep 8 && docker compose ps"
echo "✓ deployed. Front the published 127.0.0.1:3070 with your host nginx + TLS."
