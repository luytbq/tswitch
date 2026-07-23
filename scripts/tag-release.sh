#!/usr/bin/env bash
# Tag the current commit as a new version and push the tag to origin.
#
# Usage: scripts/tag-release.sh v0.3.6 ["release message"]

set -euo pipefail

VERSION="${1:?Usage: $0 vX.Y.Z [message]}"
MESSAGE="${2:-Release $VERSION}"

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: version must look like vX.Y.Z (got: $VERSION)" >&2
  exit 1
fi

if git rev-parse "$VERSION" >/dev/null 2>&1; then
  echo "error: tag $VERSION already exists" >&2
  exit 1
fi

git tag -a "$VERSION" -m "$MESSAGE"
git push origin "$VERSION"

echo "Tagged and pushed $VERSION"
git describe --tags --always --dirty
