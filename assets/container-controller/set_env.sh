#!/bin/sh
set -x
set -e

ETC_DIR="/etc/iofog/controller"
ENV_FILE_NAME=iofog-controller.env # Used as an env file in systemd

ENV_FILE="$ETC_DIR/$ENV_FILE_NAME"

# Create folder
mkdir -p "$ETC_DIR"

# Env file (for systemd)
rm -f "$ENV_FILE"
touch "$ENV_FILE"

for var in "$@"
do
  echo "$var" >> "$ENV_FILE"
done