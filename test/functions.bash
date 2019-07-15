#!/usr/bin/env bash

function test(){
    eval "$@"
    [[ $? == 0 ]]
}