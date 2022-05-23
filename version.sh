#!/bin/sh

LATEST_TAG=$(git describe --tags $(git rev-list --tags --max-count=1))
MAJOR=$(echo "$LATEST_TAG" | tr -d "v" | sed "s|-.*||" | sed -E "s|(.)\..\..|\1|g")
MINOR=$(echo "$LATEST_TAG" | tr -d "v" | sed "s|-.*||" | sed -E "s|.\.(.)\..|\1|g")
PATCH=$(echo "$LATEST_TAG" | tr -d "v" | sed "s|-.*||" | sed -E "s|.\..\.(.)|\1|g")
SUFFIX_TAG=$(echo "$LATEST_TAG" | sed "s|v$MAJOR\.$MINOR\.$PATCH||")
SUFFIX_DEV="-dev"
SUFFIX="$([ -z "$(git tag --points-at HEAD)" ] && echo "$SUFFIX_DEV" || echo "$SUFFIX_TAG")"
