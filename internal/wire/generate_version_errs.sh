#!/usr/bin/env bash
# このスクリプトは各Goバージョンでwire_errs.txtを生成します

set -euo pipefail

VERSIONS=("1.19" "1.20" "1.21" "1.22" "1.23" "1.24" "1.25")

for version in "${VERSIONS[@]}"; do
    echo "========================================"
    echo "Generating error files for Go ${version}..."
    echo "========================================"

    docker run --rm \
        -v "$(cd ../.. && pwd):/work" \
        -w /work/internal/wire \
        golang:${version} \
        bash -c "go test -run TestWire -record 2>&1"

    echo "Done with Go ${version}"
    echo
done

echo "========================================"
echo "All version-specific error files generated!"
echo "========================================"
find testdata -name "wire_errs_go*.txt" | sort
