#!/usr/bin/env bash
set -euo pipefail

go build -o requiredcheck tools/requiredcheck.go

# run vet
go vet ./sandwich/... &&
go vet ./welcomer-backend/... &&
go vet ./welcomer-gateway/... &&
go vet ./welcomer-images-next/... &&
go vet ./welcomer-images/... &&
go vet ./welcomer-interactions/... &&

go vet -vettool=./tools/requiredcheck ./sandwich/... &&
go vet -vettool=./tools/requiredcheck ./welcomer-backend/... &&
go vet -vettool=./tools/requiredcheck ./welcomer-gateway/... &&
go vet -vettool=./tools/requiredcheck ./welcomer-images-next/... &&
go vet -vettool=./tools/requiredcheck ./welcomer-images/... &&
go vet -vettool=./tools/requiredcheck ./welcomer-interactions/... &&

# run tests only if there are *_test.go files in the folder
dirs=(
    welcomer-core
    welcomer-backend
    welcomer-gateway
    welcomer-images
    welcomer-images-next
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