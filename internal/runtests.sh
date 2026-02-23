#!/usr/bin/env bash
# Copyright 2019 The Wire Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

if [[ $# -gt 0 ]]; then
  echo "usage: runtests.sh" 1>&2
  exit 64
fi

result=0

run_step() {
  echo
  echo "=== $1 ==="
}

# --- Tests (all OS) ---

run_step "Running Go tests"
go test -mod=readonly -race ./... || result=1

# No need to run lint/dep checks on OSs other than Linux.
# We default RUNNER_OS to "Linux" so that we don't abort here when run locally.
if [[ "${RUNNER_OS:-Linux}" != "Linux" ]]; then
  exit $result
fi

# --- Lint (Linux only) ---

run_step "Running go vet"
go vet ./... || result=1

run_step "Checking gofmt -s"
UNFORMATTED="$(gofmt -s -l . | grep -v testdata || true)"
if [[ -n "$UNFORMATTED" ]]; then
  echo "FAIL: please run 'gofmt -s -w .' and commit the result"
  echo "$UNFORMATTED"
  result=1
else
  echo "OK"
fi

# --- Dependency check (Linux only) ---

run_step "Checking dependencies against ./internal/alldeps"
(./internal/listdeps.sh | diff ./internal/alldeps - && echo "OK") || {
  echo "FAIL: dependencies changed; run: internal/listdeps.sh > internal/alldeps"
  echo "using the latest go version."
  result=1
}

exit $result
