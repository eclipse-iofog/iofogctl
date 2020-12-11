#!/bin/sh
set -x
set -e

CONF_FOLDER=/opt/iofog/config/controller
SOURCE_FILE_NAME=env.sh # Used to source env variables
ENV_FILE_NAME=env.env # Used as an env file in systemd

SOURCE_FILE="$CONF_FOLDER/$SOURCE_FILE_NAME"
ENV_FILE="$CONF_FOLDER/$ENV_FILE_NAME"

# Create folder
mkdir -p "$CONF_FOLDER"

# Source file
echo "#!/bin/sh" > "$SOURCE_FILE"

# Env file (for systemd)
rm -f "$ENV_FILE"
touch "$ENV_FILE"

for var in "$@"
do
  echo "export $var" >> "$SOURCE_FILE"
  echo "$var" >> "$ENV_FILE"
done