#!/usr/bin/env bash
# Copyright 2026 The Wire Authors
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

# Invoked by .github/workflows/go-compat.yml's report-failure job after at
# least one (os, go-version) leg of the check matrix has failed. Summarizes
# per-leg pass/fail via the Actions API, then either files a new Issue or
# comments on the existing open one (deduplicated via the "go-version-check"
# label) so a recurring weekly failure doesn't spam new Issues.
#
# Required env: GH_TOKEN, GITHUB_REPOSITORY, GITHUB_RUN_ID
# (All three are provided by Actions automatically; GITHUB_SERVER_URL falls
# back to github.com when run manually outside Actions.)

LABEL="go-version-check"
# Assign the repo owner so the notification reaches them regardless of their
# Watch settings (assignees are always notified as participants). Empty when
# run outside Actions, where GITHUB_REPOSITORY_OWNER is not set. Assumes a
# user-owned repo; an org owner would not be assignable and would fail the
# issue creation.
ASSIGNEE="${GITHUB_REPOSITORY_OWNER:-}"
SERVER_URL="${GITHUB_SERVER_URL:-https://github.com}"
RUN_URL="${SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"

echo "=== Collecting matrix job results ==="
# The run has only 7 jobs (6 matrix legs + this one), so a single page is
# enough; per_page=100 leaves ample headroom without pagination handling.
jobs_json="$(gh api "repos/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}/jobs?per_page=100")"

table="$(echo "$jobs_json" | jq -r '
  ["| Job | Result | Log |", "|---|---|---|"] + [
    .jobs[]
    | select(.name | startswith("check"))
    | "| " + .name + " | " +
      (if .conclusion == "success" then "OK"
       elif .conclusion == "failure" then "**FAIL**"
       else (.conclusion // "unknown") end) +
      " | [log](" + .html_url + ") |"
  ] | join("\n")
')"

# Assigned via `read` rather than body="$(cat <<EOF ...)" because bash 3.2
# (macOS /bin/bash) mis-parses quotes inside heredocs nested in command
# substitution. read -d '' exits nonzero on EOF, so guard it under set -e.
read -r -d '' body <<EOF || true
Automated weekly check against \`go-version: stable\` / \`oldstable\`
(see \`.github/workflows/go-compat.yml\`) found at least one failing leg.

**Run:** ${RUN_URL}

${table}

Notes:
- Only the \`ubuntu-latest\` legs run \`go vet\`, the \`gofmt -s\` check, and the
  \`internal/listdeps.sh\` dependency diff (see \`internal/runtests.sh\`); the
  \`macos-latest\` / \`windows-latest\` legs only run \`go test\`. Open the failing
  job's log above to see which specific step failed.
- This issue is filed/updated automatically and deduplicated only by the
  \`${LABEL}\` label (not by title). Closing it lets a future failure open a
  fresh issue; a recurring failure while this issue stays open results in a
  new comment here each week instead of a new issue.
EOF

echo "=== Ensuring label exists ==="
gh label create "$LABEL" \
  --repo "$GITHUB_REPOSITORY" \
  --color FBCA04 \
  --description "Automated Go stable/oldstable compatibility check failures" \
  --force >/dev/null 2>&1 || echo "warning: could not create/update label '$LABEL'; continuing anyway"

echo "=== Looking for an existing open issue ==="
existing="$(gh issue list --repo "$GITHUB_REPOSITORY" --state open --label "$LABEL" --json number --jq '.[0].number // empty')"

if [[ -n "$existing" ]]; then
  echo "=== Commenting on existing issue #$existing ==="
  gh issue comment --repo "$GITHUB_REPOSITORY" "$existing" --body "$body"
else
  echo "=== Filing new issue ==="
  gh issue create --repo "$GITHUB_REPOSITORY" \
    --title "Go version compatibility check failed ($(date -u +%F))" \
    --label "$LABEL" \
    ${ASSIGNEE:+--assignee "$ASSIGNEE"} \
    --body "$body"
fi
