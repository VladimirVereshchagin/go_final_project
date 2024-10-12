#!/bin/bash
set -euo pipefail

export TOKEN=$(curl -X POST http://localhost:7540/api/signin -H "Content-Type: application/json" -d "{\"password\":\"$TODO_PASSWORD\"}" --silent -c cookies.txt \
                | jq --raw-output .token)
go test ./tests -count=1
