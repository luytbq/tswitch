#!/usr/bin/env bash
# Release a new version of this Go module to the public module proxy.
#
# Publishing a Go module is just pushing a semver tag: proxy.golang.org fetches
# and stores it the first time anyone asks for it. This script runs the checks
# that are worth running before a tag becomes permanent, pushes it, then asks
# the proxy for it so the version is live rather than waiting for a stranger.
#
# Usage:
#   scripts/go-release.sh v1.2.3            explicit version
#   scripts/go-release.sh patch             bump the last tag's patch field
#   scripts/go-release.sh minor "message"   bump minor, custom tag message
#   scripts/go-release.sh major --dry-run   run every check, tag nothing

set -euo pipefail

usage() {
  cat >&2 <<USAGE
Usage: $0 <vX.Y.Z|patch|minor|major> [message] [--dry-run] [--skip-install-check]

  vX.Y.Z    release exactly this version
  patch     x.y.Z+1   bug fixes only
  minor     x.Y+1.0   new features, backwards compatible
  major     X+1.0.0   breaking changes
USAGE
  exit 2
}

[[ $# -ge 1 ]] || usage

BUMP=""
VERSION=""
MESSAGE=""
DRY_RUN=0
INSTALL_CHECK=1

case "$1" in
  patch|minor|major) BUMP="$1" ;;
  v[0-9]*)           VERSION="$1" ;;
  *)                 usage ;;
esac
shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)            DRY_RUN=1 ;;
    --skip-install-check) INSTALL_CHECK=0 ;;
    -*)                   usage ;;
    *)                    MESSAGE="$1" ;;
  esac
  shift
done

die() { echo "error: $*" >&2; exit 1; }
step() { echo; echo "==> $*"; }

cd "$(git rev-parse --show-toplevel 2>/dev/null)" || die "not inside a git repository"

MODULE="$(awk '/^module /{print $2; exit}' go.mod)"
[[ -n "$MODULE" ]] || die "no module path in go.mod"

# ---------------------------------------------------------------------------
# Work out the version
# ---------------------------------------------------------------------------

LATEST="$(git tag --list 'v[0-9]*' --sort=-v:refname | head -1)"

if [[ -n "$BUMP" ]]; then
  [[ -n "$LATEST" ]] || die "no existing vX.Y.Z tag to bump from; pass a version"
  IFS=. read -r major minor patch <<<"${LATEST#v}"
  case "$BUMP" in
    patch) patch=$((patch + 1)) ;;
    minor) minor=$((minor + 1)); patch=0 ;;
    major) major=$((major + 1)); minor=0; patch=0 ;;
  esac
  VERSION="v${major}.${minor}.${patch}"
fi

[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "version must look like vX.Y.Z (got: $VERSION)"
git rev-parse "$VERSION" >/dev/null 2>&1 && die "tag $VERSION already exists"

MESSAGE="${MESSAGE:-Release $VERSION}"

echo "module:   $MODULE"
echo "previous: ${LATEST:-none}"
echo "new:      $VERSION"
[[ $DRY_RUN -eq 1 ]] && echo "mode:     dry run, nothing will be tagged or pushed"

# ---------------------------------------------------------------------------
# Pre-flight
# ---------------------------------------------------------------------------

step "Checking the working tree"
if ! git diff --quiet || ! git diff --cached --quiet; then
  git status --short >&2
  die "tracked files have uncommitted changes; a tag must point at what you ship"
fi
UNTRACKED="$(git ls-files --others --exclude-standard)"
[[ -n "$UNTRACKED" ]] && echo "note: untracked files present, they are not part of the release:" && echo "$UNTRACKED" | sed 's/^/  /'

step "Checking the branch is pushed"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
git rev-parse --verify --quiet "@{upstream}" >/dev/null || die "branch $BRANCH has no upstream; push it first"
git fetch --quiet origin "$BRANCH"
AHEAD="$(git rev-list --count '@{upstream}..HEAD')"
BEHIND="$(git rev-list --count 'HEAD..@{upstream}')"
[[ "$AHEAD"  -eq 0 ]] || die "$AHEAD local commit(s) not pushed; the proxy fetches from the remote, so push first"
[[ "$BEHIND" -eq 0 ]] || die "$BEHIND commit(s) on the remote are not merged locally"
echo "$BRANCH is level with its upstream"

step "Building and vetting"
go build ./...
go vet ./...

step "Checking go.mod and go.sum are tidy"
if ! go mod tidy -diff >/dev/null 2>&1; then
  die "go.mod/go.sum are not tidy; run: go mod tidy"
fi
git ls-files --error-unmatch go.sum >/dev/null 2>&1 \
  || echo "warning: go.sum is not tracked, so a clone cannot reproduce this build"

UNFORMATTED="$(gofmt -l . | grep -v '^vendor/' || true)"
[[ -n "$UNFORMATTED" ]] && echo "warning: not gofmt-clean (not blocking the release):" && echo "$UNFORMATTED" | sed 's/^/  /'

if [[ $DRY_RUN -eq 1 ]]; then
  step "Dry run finished; every check passed"
  exit 0
fi

# ---------------------------------------------------------------------------
# Publish
# ---------------------------------------------------------------------------

step "Tagging $VERSION"
git tag -a "$VERSION" -m "$MESSAGE"
git push origin "$VERSION"

step "Asking proxy.golang.org to fetch it"
# Upper-case letters in a module path are escaped as !lower on the proxy.
ESCAPED="$(echo "$MODULE" | sed 's/\([A-Z]\)/!\L\1/g')"
for attempt in $(seq 1 10); do
  if curl -fsS "https://proxy.golang.org/${ESCAPED}/@v/${VERSION}.info"; then
    echo
    break
  fi
  [[ $attempt -eq 10 ]] && die "the proxy did not serve $VERSION; check the tag reached GitHub"
  sleep 6
done

if [[ $INSTALL_CHECK -eq 1 ]]; then
  step "Installing $MODULE@$VERSION from the public proxy"
  BINDIR="$(mktemp -d)"
  trap 'rm -rf "$BINDIR"' EXIT
  GOBIN="$BINDIR" GOPROXY=proxy.golang.org GOFLAGS= go install "$MODULE@$VERSION"
  echo "installed: $(ls "$BINDIR")"
fi

step "Released $VERSION"
cat <<DONE
  install:  go install $MODULE@$VERSION
  docs:     https://pkg.go.dev/$MODULE@$VERSION

pkg.go.dev builds its page from the proxy on its own schedule and can lag by
tens of minutes. The install command above works immediately regardless.
DONE
