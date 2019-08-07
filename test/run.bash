#!/bin/bash

set -e

# Export variables
CONF=test/conf/env.sh
if [ -f "$CONF" ]; then
    . "$CONF"
fi

# Run tests
for TEST in "$@"; do
    bats "test/$TEST.bats"
done