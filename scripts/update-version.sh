#!/usr/bin/env bash
# scripts/update-version.sh
# Usage:
#   ./scripts/update-version.sh v0.2.5
#   ./scripts/update-version.sh 0.2.5

set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "❌ Error: Version argument is required."
  echo "Usage: $0 <version> (e.g. $0 v0.2.5)"
  exit 1
fi

RAW_VERSION="$1"
# Ensure format has leading 'v'
if [[ ! "$RAW_VERSION" =~ ^v ]]; then
  NEW_VERSION="v${RAW_VERSION}"
else
  NEW_VERSION="${RAW_VERSION}"
fi

# Validate semver pattern (e.g. v1.2.3, v1.2.3-beta.1)
if [[ ! "$NEW_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "❌ Error: Invalid semantic version format: '$RAW_VERSION'"
  echo "Expected format: vMAJOR.MINOR.PATCH (e.g. v0.2.5)"
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "🔄 Updating RouteWarden version to: ${NEW_VERSION}"

UPDATED_COUNT=0

# List of files to update
FILES=(
  "README.md"
  "examples/01-basic-sensitive-files/docker-compose.yml"
  "examples/02-global-entrypoint-shield/docker-compose.yml"
  "examples/03-ip-whitelist-vpn/docker-compose.yml"
  "examples/04-captcha-challenge/docker-compose.yml"
  "examples/05-kubernetes-ingressroute/README.md"
)

for REL_PATH in "${FILES[@]}"; do
  FILE_PATH="${ROOT_DIR}/${REL_PATH}"
  if [ -f "$FILE_PATH" ]; then
    # Replace CLI flag: --experimental.plugins.routewarden.version=vX.Y.Z
    sed -i '' -E "s|(--experimental\.plugins\.routewarden\.version=)v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?|\1${NEW_VERSION}|g" "$FILE_PATH"

    # Replace YAML version property: version: vX.Y.Z
    sed -i '' -E "s|(version:[[:space:]]+)v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?|\1${NEW_VERSION}|g" "$FILE_PATH"

    echo "  ✓ Updated ${REL_PATH}"
    UPDATED_COUNT=$((UPDATED_COUNT + 1))
  fi
done

echo ""
echo "✨ Successfully updated version to ${NEW_VERSION} in ${UPDATED_COUNT} file(s)!"
echo "👉 Next steps:"
echo "   git diff"
echo "   git commit -am 'chore: release ${NEW_VERSION}'"
echo "   git tag ${NEW_VERSION}"
echo "   git push origin main --tags"
