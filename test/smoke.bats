#!/usr/bin/env bash

# Required environment variables
# NAMESPACE

. test/functions.bash

@test "Help" {
  test iofogctl --help
}

@test "Help w/o flag" {
  test iofogctl help
}

@test "create Help" {
  test iofogctl create --help
}

@test "delete Help" {
  test iofogctl delete --help
}

@test "deploy Help" {
  test iofogctl deploy --help
}

@test "describe Help" {
  test iofogctl describe --help
}

@test "connect Help" {
  test iofogctl connect --help
}

@test "disconnect Help" {
  test iofogctl disconnect --help
}

@test "legacy Help" {
  test iofogctl legacy --help
}

@test "logs Help" {
  test iofogctl logs --help
}

@test "get Help" {
  test iofogctl get --help
}

@test "version" {
  test iofogctl version
}

@test "Get All" {
  test iofogctl get all
}

@test "Get Namespaces" {
  test iofogctl get namespaces
}

@test "Get Controllers" {
  test iofogctl get controllers
}

@test "Get Agents" {
  test iofogctl get agents
}

#@test "Get Microservices" {
#  test iofogctl get microservices
#}

@test "create namespace" {
  test iofogctl create -n $NAMESPACE
}

@test "delete namespace" {
  test iofogctl delete -n $NAMESPACE
}