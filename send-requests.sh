#!/usr/bin/env bash
# Sends the notes API a set of calls covering every endpoint, so a recording
# captures test cases and their Mongo mocks.
# Run after port-forwarding: kubectl port-forward svc/notes-app -n bugtest 8080:8080

set -euo pipefail

BASE="${BASE:-http://localhost:8080}"

echo "==> health"
curl -sf "$BASE/health"; echo

create() {
  curl -sf -X POST "$BASE/notes" -H "Content-Type: application/json" -d "$1" |
    python3 -c "import sys,json; print(json.load(sys.stdin)['id'])"
}

echo "==> create 3 notes"
ID1=$(create '{"title":"Buy groceries","body":"Milk, eggs, bread"}')
ID2=$(create '{"title":"Meeting notes","body":"Discuss the release"}')
ID3=$(create '{"title":"Read later","body":"Keploy docs"}')
echo "created: $ID1 $ID2 $ID3"

echo "==> list notes"
curl -sf "$BASE/notes"; echo

echo "==> get notes 1 and 2"
curl -sf "$BASE/notes/$ID1"; echo
curl -sf "$BASE/notes/$ID2"; echo

echo "==> delete note 3"
curl -s -X DELETE "$BASE/notes/$ID3" -o /dev/null -w "status: %{http_code}\n"

echo "==> list after delete"
curl -sf "$BASE/notes"; echo

echo "==> deleted note answers 404"
curl -s "$BASE/notes/$ID3"; echo

echo "==> bad requests"
curl -s -X POST "$BASE/notes" -H "Content-Type: application/json" -d '{"body":"no title"}'; echo
curl -s "$BASE/notes/not-an-id"; echo

echo "Done."
