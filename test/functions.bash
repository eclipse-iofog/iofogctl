#!/usr/bin/env bash

function test(){
    result=$("$@")
    [[ $? == 0 ]]
}