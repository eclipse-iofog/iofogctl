#!/usr/bin/env bash

. test/func/include.bash

@test "Help" {
  iofogctl --help
}

@test "Help w/o Flag" {
  iofogctl help
}

@test "Create Help" {
  iofogctl create --help
}

@test "Delete Help" {
  iofogctl delete --help
}

@test "Deploy Help" {
  iofogctl deploy --help
}

@test "Describe Help" {
  iofogctl describe --help
}

@test "Connect Help" {
  iofogctl connect --help
}

@test "Disconnect Help" {
  iofogctl disconnect --help
}

@test "Legacy Help" {
  iofogctl legacy --help
}

@test "Logs Help" {
  iofogctl logs --help
}

@test "Get Help" {
  iofogctl get --help
}

@test "Version" {
  iofogctl version
}

@test "Get All" {
  iofogctl get all
}

@test "Get Namespaces" {
  iofogctl get namespaces
}

@test "Get Controllers" {
  iofogctl get controllers
}

@test "Get Agents" {
  iofogctl get agents
}

@test "Get Microservices" {
  iofogctl get microservices
}

@test "Get Applications" {
  iofogctl get applications
}

@test "Create Namespace" {
  iofogctl create namespace smoketestsnamespace1234
}

@test "Set Default Namespace" {
  iofogctl configure current-namespace smoketestsnamespace1234
  iofogctl get all
}

@test "Delete Namespace" {
  iofogctl delete namespace smoketestsnamespace1234
  iofogctl get all
  iofogctl get namespaces
}