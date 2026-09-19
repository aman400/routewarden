#!/usr/bin/env bash
set -euo pipefail

echo "========================================================"
echo " RouteWarden Local Docker Verification"
echo "========================================================"

BASE_URL="http://localhost"

echo ""
echo "[1] Testing normal request (should return 200 OK from whoami)..."
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "${BASE_URL}/"

echo ""
echo "[2] Testing whitelisted endpoint /robots.txt (should pass through)..."
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "${BASE_URL}/robots.txt"

echo ""
echo "[3] Testing sensitive endpoint /.env (should trigger RouteWarden honeypot decoy)..."
echo "--- Response Headers & Body ---"
curl -i -s "${BASE_URL}/.env"

echo ""
echo "[4] Testing sensitive endpoint /phpinfo.php..."
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "${BASE_URL}/phpinfo.php"

echo ""
echo "Done! Check Traefik debug logs with:"
echo "  docker logs routewarden-traefik-dev 2>&1 | grep '\\[DEBUG\\]'"
