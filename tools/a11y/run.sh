#!/usr/bin/env bash
# Build the showcase, serve it on a free port, run the axe scan, stop the server by PID.
# Run from the repository root (task a11y does). Extra arguments go to run.mjs.
set -u
root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

[ -d tools/a11y/node_modules ] || (cd tools/a11y && npm ci) || exit 2

bin="$(mktemp -d)/showcase"
go build -o "$bin" ./cmd/showcase || exit 2

port="$(node -e 'const s=require("net").createServer().listen(0,"127.0.0.1",()=>{console.log(s.address().port);s.close()})')"
"$bin" -addr "127.0.0.1:$port" >"$(dirname "$bin")/server.log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null; wait "$pid" 2>/dev/null; rm -rf "$(dirname "$bin")"' EXIT

for _ in $(seq 1 50); do
  curl -fsS "http://127.0.0.1:$port/" >/dev/null 2>&1 && break
  kill -0 "$pid" 2>/dev/null || { echo "showcase exited early:"; cat "$(dirname "$bin")/server.log"; exit 2; }
  sleep 0.2
done

cd tools/a11y
BASE_URL="http://127.0.0.1:$port" node run.mjs "$@"
