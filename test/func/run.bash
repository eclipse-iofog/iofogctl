#!/usr/bin/env bash

# These functions are designed to be used with the bats `run` command

function curlMsvc(){
    IP="$1"
    curl -s --max-time 120 http://${IP}:5000/api/raw
}

function jqMsvcArray(){
    ARR="$1"
    echo "$ARR" | jq '. | length'
}