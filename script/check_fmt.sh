#!/bin/sh

FMT_FILES=$(gofmt -s -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./client/*"))
if [ -z "$FMT_FILES" ]; then
    echo "Files are formatted correctly. Thanks!"
    exit 0
else
    echo "Files are not formatted correctly:"
    echo "$FMT_FILES"
    exit 1
fi