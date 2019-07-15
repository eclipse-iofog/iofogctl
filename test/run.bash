#!/bin/bash

export NAMESPACE=testing
bats test/smoke.bats
# bats test/functional.bats