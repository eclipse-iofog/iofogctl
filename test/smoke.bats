#!/usr/bin/env bash

. test/functions.bash

@test "Help" {
  test iofogctl --help
}

@test "Help w/o Flag" {
  test iofogctl help
}

@test "Create Help" {
  test iofogctl create --help
}

@test "Delete Help" {
  test iofogctl delete --help
}

@test "Deploy Help" {
  test iofogctl deploy --help
}

@test "Describe Help" {
  test iofogctl describe --help
}

@test "Connect Help" {
  test iofogctl connect --help
}

@test "Disconnect Help" {
  test iofogctl disconnect --help
}

@test "Legacy Help" {
  test iofogctl legacy --help
}

@test "Logs Help" {
  test iofogctl logs --help
}

@test "Get Help" {
  test iofogctl get --help
}

@test "Version" {
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

@test "Get Microservices" {
  test iofogctl get microservices
}

@test "Get Applications" {
  test iofogctl get applications
}

@test "Create Namespace" {
  test iofogctl create namespace smoketestsnamespace1234
}

@test "Set Default Namespace" {
  test iofogctl configure default-namespace smoketestsnamespace1234
  test iofogctl get all
}

@test "Delete Namespace" {
  test iofogctl delete namespace smoketestsnamespace1234
  test iofogctl get all
  test iofogctl get namespaces
}