#!/bin/sh
set -x
set -e

CONTROLLER_DIR="/opt/iofog/controller/"
CONTROLLER_LOG_DIR="/var/log/iofog/"

do_uninstall_controller() {
  # Remove folders
  sudo rm -rf $CONTROLLER_DIR
  sudo rm -rf $CONTROLLER_LOG_DIR

  # Remove symbolic links
  rm -f /usr/local/bin/iofog-controller

  # Remove service files
  USE_SYSTEMD=`grep -m1 -c systemd /proc/1/comm`
  USE_INITCTL=`which initctl | wc -l`
  USE_SERVICE=`which service | wc -l`

  if [ $USE_SYSTEMD -eq 1 ]; then
    systemctl stop iofog-controller.service
    rm -f /etc/systemd/system/iofog-controller.service
  elif [ $USE_INITCTL -eq 1 ]; then
    rm -f /etc/init/iofog-controller.conf
  elif [ $USE_SERVICE -eq 1 ]; then
    rm -f /etc/init.d/iofog-controller
  else
    echo "Unable to setup Controller startup script."
  fi
}

do_uninstall_controller