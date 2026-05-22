#!/bin/sh
set -eu

# Server-side game registration at container start.
# Tries POST /api/games on backend (default http://backend:8080) with retries.

BACKEND_URL=${GAME_HUB_URL:-http://backend:8080}
ENDPOINT="$BACKEND_URL/api/games"

payload='{"id":"tictactoe","name":"Tic Tac Toe","target":"http://tictactoe:80","status":"online","icon":"❌","author":"Jakub Šefl a Claude Sonnet","description":"Local 2-player Tic Tac Toe. No AI, no network — just two players on one screen battling it out.","image":""}'

max_retries=30
sleep_seconds=2
attempt=1

while [ $attempt -le $max_retries ]; do
  echo "[register-and-run] Attempt $attempt: POST $ENDPOINT"

  http_code=$(curl -s -o /tmp/register_resp -w "%{http_code}" -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" -d "$payload" || true)

  if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
    echo "[register-and-run] Registration succeeded (HTTP $http_code)"
    cat /tmp/register_resp
    break
  else
    echo "[register-and-run] Registration failed (HTTP $http_code):"
    cat /tmp/register_resp || true
    attempt=$((attempt+1))
    sleep $sleep_seconds
  fi

done

if [ $attempt -gt $max_retries ]; then
  echo "[register-and-run] Registration did not succeed after $max_retries attempts; continuing to start nginx."
fi

# Exec nginx as PID 1
exec nginx -g "daemon off;"
