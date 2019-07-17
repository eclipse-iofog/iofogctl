#!/bin/bash

set -e

# Export variables
. test/conf/env.sh

# Run smoke tests
bats test/smoke.bats

# Run functional tests
bats test/functional.bats