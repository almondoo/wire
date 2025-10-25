#!/usr/bin/env bash
# 各Goバージョンでテストを実行します

set -euo pipefail

VERSIONS=("1.19" "1.20" "1.21" "1.22" "1.23" "1.24" "1.25")

echo "========================================"
echo "Testing Wire with multiple Go versions"
echo "========================================"
echo

for version in "${VERSIONS[@]}"; do
    echo "----------------------------------------"
    echo "Testing with Go ${version}..."
    echo "----------------------------------------"

    docker run --rm \
        -v "/Users/tm/development/wire:/work" \
        -w /work/internal/wire \
        golang:${version} \
        go test -mod=readonly ./... 2>&1 | grep -E "(PASS|FAIL|ok|---)" | tail -5

    if [ $? -eq 0 ]; then
        echo "✓ Go ${version}: PASSED"
    else
        echo "✗ Go ${version}: FAILED"
    fi
    echo
done

echo "========================================"
echo "All tests completed!"
echo "========================================"
