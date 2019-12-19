#!/bin/sh
set -x
set -e

CONNECTOR_DIR="/opt/iofog/connector/"
CONNECTOR_CONFIG_FOLDER="/etc/iofog-connector/"
CONNECTOR_LOG_FOLDER="/etc/iofog-connector/"

do_uninstall_connector() {
  # Remove folders
  sudo rm -rf $CONNECTOR_DIR
  sudo rm -rf $CONNECTOR_CONFIG_FOLDER
  sudo rm -rf $CONNECTOR_LOG_FOLDER

  # Remove symbolic links
  sudo rm -f /usr/local/bin/iofog-connector
  # Connector is hard coded to look into /usr/bin for .jar and .jard
  sudo rm -f /usr/bin/iofog-connectord.jar
  sudo rm -f /usr/bin/iofog-connector.jar
}

do_uninstall_connector