#!/usr/bin/env bash
set -euo pipefail

# run vet
go vet welcomer-backend/cmd/main.go &&
go vet welcomer-gateway/cmd/main.go &&
go vet welcomer-images/cmd/main.go &&
go vet welcomer-interactions/cmd/main.go &&
go vet sandwich/sandwich.go

# run tests only if there are *_test.go files in the folder
dirs=(
    welcomer-core
    welcomer-backend
    welcomer-gateway
    welcomer-images
    welcomer-interactions
    sandwich
)

for d in "${dirs[@]}"; do
    if compgen -G "$d"/*_test.go > /dev/null; then
        (cd "$d" && go test .)
    else
        echo "Skipping $d (no _test.go files)"
    fi
done