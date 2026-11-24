
#!/usr/bin/env bash

set -Eeuo pipefail

INCLUDE_MAJOR=false
SKIP_TESTS=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --include-major) INCLUDE_MAJOR=true ;;
        --skip-tests) SKIP_TESTS=true ;;
        --dry-run) DRY_RUN=true ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
    shift
done

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
NC='\033[0m'

info() { echo -e "${CYAN}$1${NC}"; }
success() { echo -e "${GREEN}$1${NC}"; }
error() { echo -e "${RED}$1${NC}"; }

command -v go >/dev/null || {
    error "Go is not installed."
    exit 1
}

# Find all Go modules recursively.
mapfile -d '' MODULES < <(
    find . \
        \( -type d \( -name .git -o -name vendor -o -name node_modules \) -prune \) -o \
        \( -type f -name go.mod -print0 \)
)

if [[ ${#MODULES[@]} -eq 0 ]]; then
    error "No Go modules found."
    exit 1
fi

info "Found ${#MODULES[@]} Go module(s)."

FAILED=0
UPDATED=0

for MOD_FILE in "${MODULES[@]}"; do
    MODULE_DIR=$(dirname "$MOD_FILE")

    info "\n========================================"
    info "Processing: $MODULE_DIR"
    info "========================================"

    (
        cd "$MODULE_DIR"

        BACKUP_DIR=$(mktemp -d)

        cp go.mod "$BACKUP_DIR/go.mod"

        HAD_GO_SUM=false
        if [[ -f go.sum ]]; then
            cp go.sum "$BACKUP_DIR/go.sum"
            HAD_GO_SUM=true
        fi

        restore() {
            cp "$BACKUP_DIR/go.mod" go.mod

            if [[ "$HAD_GO_SUM" == true ]]; then
                cp "$BACKUP_DIR/go.sum" go.sum
            else
                rm -f go.sum
            fi
        }

        cleanup() {
            rm -rf "$BACKUP_DIR"
        }

        trap cleanup EXIT
        trap 'restore; error "Update failed in $MODULE_DIR"; exit 1' ERR

        info "Checking outdated dependencies..."

        go list -u -m -json all > "$BACKUP_DIR/outdated.json"

        python3 - "$BACKUP_DIR/outdated.json" "$INCLUDE_MAJOR" <<'PY'
import json
import sys

path, include_major = sys.argv[1], sys.argv[2] == "true"

with open(path) as f:
    content = f.read()

decoder = json.JSONDecoder()
index = 0
count = 0

while index < len(content):
    while index < len(content) and content[index].isspace():
        index += 1

    if index >= len(content):
        break

    item, index = decoder.raw_decode(content, index)
    update = item.get("Update")

    if not update:
        continue

    current = item.get("Version", "")
    target = update.get("Version", "")

    if not include_major:
        current_major = current.split(".")[0].lstrip("v")
        target_major = target.split(".")[0].lstrip("v")

        if current_major != target_major:
            continue

    print(f'{item["Path"]}: {current} -> {target}')
    count += 1

if count == 0:
    print("No compatible updates available.")
PY

        if [[ "$DRY_RUN" == true ]]; then
            success "Dry run completed."
            exit 0
        fi

        info "Updating dependencies..."
        go get -u ./...

        info "Tidying modules..."
        go mod tidy

        if [[ "$SKIP_TESTS" == false ]]; then
            info "Running tests..."
            go test ./...

            info "Running go vet..."
            go vet ./...
        fi

        success "Successfully updated $MODULE_DIR"

    ) || {
        error "Failed: $MODULE_DIR"
        FAILED=$((FAILED + 1))
        continue
    }

    UPDATED=$((UPDATED + 1))
done

info "\n========================================"
info "Dependency update summary"
info "========================================"

success "Successful: $UPDATED"
if [[ $FAILED -gt 0 ]]; then
    error "Failed: $FAILED"
else
    success "Failed: 0"
fi

[[ $FAILED -eq 0 ]]
