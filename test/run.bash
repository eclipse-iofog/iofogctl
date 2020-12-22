#!/bin/bash

set -e

# Export variables
CONF=test/conf/env.sh
if [ -f "$CONF" ]; then
    . "$CONF"
fi

function preprocess() {
    TO_LOAD=$((cat $1 | grep "# LOAD: ") || (echo -n ""))
    if [[ -z $TO_LOAD ]]; then
        return 0
    fi
    while IFS= read -r GREP_LINE; do
        grep -B 999999 "${GREP_LINE}" $1 > ${1}_temp
        local FILE_TO_LOAD=$(echo $GREP_LINE | awk '{print $3}')
        preprocess $FILE_TO_LOAD
        echo "" >> ${1}_temp
        cat $FILE_TO_LOAD >> ${1}_temp
        echo "" >> ${1}_temp
        grep -A 999999 "${GREP_LINE}" $1 >> ${1}_temp
        mv ${1}_temp $1
    done <<< "$TO_LOAD"
}

# Run tests
for TEST in "$@"; do
    TEST_FILE="/tmp/tmp.bats"
    cp "test/bats/$TEST.bats" $TEST_FILE
    preprocess $TEST_FILE
    bats $TEST_FILE
done