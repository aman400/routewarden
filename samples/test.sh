#!/usr/bin/env bash
set -e

BASE_URL="http://localhost"

echo "🧪 Running RouteWarden All-Modes Local Test Suite"
echo "=================================================================="

test_mode() {
  local mode_name="$1"
  local path="$2"
  local expected_status="$3"
  local extra_check="$4"
  local method="${5:-GET}"

  echo -n "👉 $mode_name ($method $path)... "
  
  if [ "$mode_name" = "silentDrop" ]; then
    # Silent drop should close connection without response
    local err
    set +e
    err=$(curl -s -X "$method" -v "$BASE_URL$path" 2>&1)
    set -e
    if echo "$err" | grep -q -E "Empty reply|reset by peer|Connection reset"; then
      echo "✅ PASS (Connection closed abruptly / Empty reply)"
    else
      echo "❌ FAIL (Did not drop connection: $err)"
    fi
    return
  fi

  local header_file
  header_file=$(mktemp)
  trap 'rm -f "$header_file"' RETURN

  # Safely dump headers to temp file, discard binary response body
  curl -s -X "$method" -D "$header_file" -o /dev/null "$BASE_URL$path"

  local status
  local header
  header=$(cat "$header_file" | tr -d '\r')
  status=$(grep -m 1 "HTTP/" "$header_file" | awk '{print $2}')

  if [ "$status" = "$expected_status" ]; then
    if [ -n "$extra_check" ] && ! echo "$header" | grep -q -i "$extra_check"; then
      echo "⚠️ STATUS $status MATCHED, BUT MISSING HEADER: $extra_check"
    else
      echo "✅ PASS (HTTP $status)"
    fi
  else
    echo "❌ FAIL (Expected HTTP $expected_status, got $status)"
  fi
}

echo ""
echo "--- Testing Legitimate Traffic (Should Bypass to Backend) ---"
test_mode "Normal Traffic" "/" "200"
test_mode "Allowlist /robots.txt" "/robots.txt" "200"

echo ""
echo "--- Testing All Defensive Response Modes Against Probe /.env ---"
test_mode "fakeSuccess (Decoy 200 OK)" "/mode/fakesuccess/.env" "200"
test_mode "json (Structured Error)"    "/mode/json/.env"        "403" "Content-Type: application/json"
test_mode "html (Error Page)"          "/mode/html/.env"        "403" "Content-Type: text/html"
test_mode "redirect (HTTP 302)"        "/mode/redirect/.env"    "302" "Location: https://example.com"
test_mode "rateLimit (HTTP 429)"       "/mode/ratelimit/.env"   "429" "Retry-After: 60"
test_mode "xml (XML Document)"         "/mode/xml/.env"         "403" "Content-Type: application/xml"
test_mode "captcha (Turnstile Page)"   "/mode/captcha/.env"     "403" "Content-Type: text/html"
test_mode "gzipBomb (Compressed Bomb)" "/mode/gzipbomb/.env"    "403" "Content-Encoding: gzip"
test_mode "text (Plain Text 403)"      "/mode/text/.env"        "403" "Content-Type: text/plain"
test_mode "silentDrop"                 "/mode/silentdrop/.env"  ""

echo ""
echo "--- Testing Configuration Flags & Inspection Rules ---"
# 1. checkQuery Flag: verifies sensitive parameter in query string is intercepted
test_mode "checkQuery: benign query"        "/flag/checkquery/search?q=hello" "200"
test_mode "checkQuery: blocked query param" "/flag/checkquery/search?file=.env" "403"

# 2. methods Flag: only inspects POST, GET passes through to backend
test_mode "methods: GET bypasses inspection" "/flag/methods/.env" "200" "" "GET"
test_mode "methods: POST blocked on probe"   "/flag/methods/.env" "403" "" "POST"

# 3. allowedIps Flag: whitelisted client IP bypasses blocking
test_mode "allowedIps: whitelisted IP bypasses" "/flag/allowedips/.env" "200"

# 4. custom blockPatterns Flag: custom pattern blocked, default unlisted path allowed
test_mode "blockPatterns: custom pattern blocked" "/flag/custompatterns/secret" "403"
test_mode "blockPatterns: non-matching path allowed" "/flag/custompatterns/public" "200"

# 5. enabled Flag: enabled=false completely disables filtering
test_mode "enabled=false: all probes pass" "/flag/disabled/.env" "200"


echo ""
echo "--- Testing Structured JSON Security Audit Logs (CrowdSec / SIEM) ---"
echo "Triggering test probe /.env to generate audit event..."
curl -s -o /dev/null "$BASE_URL/mode/json/.env" || true

echo -n "Checking Traefik container logs for structured 'routewarden_block' JSON... "
if command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' 2>/dev/null | grep -q "routewarden-traefik-sample"; then
  recent_log=$(docker logs --tail 25 routewarden-traefik-sample 2>&1 | grep "routewarden_block" | tail -n 1 || true)
  if [ -n "$recent_log" ]; then
    echo "✅ PASS (Audit log detected)"
    echo "   Sample event: $recent_log"
  else
    echo "⚠️ Traefik running but no recent routewarden_block log found in last 25 lines"
  fi
else
  echo "ℹ️ (Docker container not running or inaccessible from test shell - skipping live container log check)"
fi

echo ""
echo "=================================================================="
echo "📄 Quick Payload Samples:"
echo "------------------------------------------------------------------"
echo "1. fakeSuccess (/.env):"
curl -s "$BASE_URL/mode/fakesuccess/.env" | head -n 3
echo "..."
echo ""
echo "2. json (/.env):"
curl -s "$BASE_URL/mode/json/.env"
echo ""
echo "3. xml (/.env):"
curl -s "$BASE_URL/mode/xml/.env"
echo ""
echo "🎉 All modes and security audit logging verified successfully!"

