#!/usr/bin/env bash
set -euo pipefail

FEATURE_BRANCH="${FEATURE_BRANCH:-account-service-v1}"
BASE_BRANCH="${BASE_BRANCH:-master}"
REMOTE="${REMOTE:-origin}"
PUSH_AFTER_MERGE="false"

if [[ "${1:-}" == "--push" ]]; then
  PUSH_AFTER_MERGE="true"
elif [[ "${1:-}" != "" ]]; then
  echo "Usage: $0 [--push]"
  echo
  echo "Environment overrides:"
  echo "  FEATURE_BRANCH=account-service-v1"
  echo "  BASE_BRANCH=master"
  echo "  REMOTE=origin"
  exit 2
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "error: this script must be run inside a git work tree" >&2
  exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is not clean. Commit, stash, or discard changes first." >&2
  git status --short
  exit 1
fi

echo "Fetching $REMOTE/$BASE_BRANCH and $REMOTE/$FEATURE_BRANCH..."
git fetch "$REMOTE" "$BASE_BRANCH" "$FEATURE_BRANCH"

echo "Checking out $BASE_BRANCH..."
git checkout "$BASE_BRANCH"

echo "Updating $BASE_BRANCH from $REMOTE/$BASE_BRANCH..."
git pull --ff-only "$REMOTE" "$BASE_BRANCH"

echo "Merging $FEATURE_BRANCH into $BASE_BRANCH..."
git merge "$FEATURE_BRANCH"

echo
echo "Merge completed locally."
echo "Current HEAD:"
git log --oneline -1

if [[ "$PUSH_AFTER_MERGE" == "true" ]]; then
  echo
  echo "Pushing $BASE_BRANCH to $REMOTE..."
  git push "$REMOTE" "$BASE_BRANCH"
  echo "Pushed $BASE_BRANCH to $REMOTE."
else
  echo
  echo "Next recommended commands:"
  echo "  cd \"$REPO_ROOT/service\" && go test ./..."
  echo "  cd \"$REPO_ROOT/web\" && npm test -- --run && npm run build"
  echo "  git push $REMOTE $BASE_BRANCH"
  echo
  echo "Tip: run $0 --push to push automatically after a successful merge."
fi
