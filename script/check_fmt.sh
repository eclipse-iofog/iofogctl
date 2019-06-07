#!/bin/sh

# Import our helper functions
. script/utils.sh

prettyTitle "Checking src file formatting"

FMT_FILES=$(gofmt -s -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./client/*"))

if [ -z "$FMT_FILES" ]; then
    echoInfo "Files are formatted correctly. Thanks!"
    exit 0
else
    echoError "Bad '${USER}'! Files are not formatted correctly:"
    echoInfo "$FMT_FILES"
    exit 1
fi