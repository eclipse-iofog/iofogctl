#!/bin/sh
# Script to detect Linux distribution and version
# Used as a precursor for system-specific installations

# Exit on error and print commands for debugging
set -e
set -x

# Define user variable
user="$(id -un 2>/dev/null || true)"

# Check if a command exists
command_exists() {
    command -v "$@" > /dev/null 2>&1
}

# Detect the Linux distribution
get_distribution() {
    lsb_dist=""
    dist_version=""
    
    # Every system that we officially support has /etc/os-release
    if [ -r /etc/os-release ]; then
        
        lsb_dist="$(. /etc/os-release && echo "$ID")"
        
        dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
        lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
    else
        echo "Error: Unsupported Linux distribution! /etc/os-release not found."
        exit 1
    fi
    
    echo "# Detected distribution: $lsb_dist (version: $dist_version)"
}

# Check if this is a forked Linux distro
check_forked() {
    # Skip if lsb_release doesn't exist
    if ! command_exists lsb_release; then
        return
    fi
    
    # Check if the `-u` option is supported
    set +e
    lsb_release -a > /dev/null 2>&1
    lsb_release_exit_code=$?
    set -e

    # Check if the command has exited successfully, it means we're in a forked distro
    if [ "$lsb_release_exit_code" = "0" ]; then
        # Get the upstream release info
        current_lsb_dist=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'id' | cut -d ':' -f 2 | tr -d '[:space:]')
        current_dist_version=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'codename' | cut -d ':' -f 2 | tr -d '[:space:]')

        # Print info about current distro
        echo "You're using '$current_lsb_dist' version '$current_dist_version'."
        
        # Check if current is different from detected (indicating a fork)
        if [ "$current_lsb_dist" != "$lsb_dist" ] || [ "$current_dist_version" != "$dist_version" ]; then
            echo "Upstream release is '$lsb_dist' version '$dist_version'."
        fi
    else
        # Additional checks for specific distros that might not be properly detected
        if [ -r /etc/debian_version ] && [ "$lsb_dist" != "ubuntu" ] && [ "$lsb_dist" != "raspbian" ]; then
            if [ "$lsb_dist" = "osmc" ]; then
                # OSMC runs Raspbian
                lsb_dist=raspbian
            else
                # We're Debian and don't even know it!
                lsb_dist=debian
            fi
            # Get Debian version and map it to codename
            dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
            case "$dist_version" in
                14)
                    dist_version="forky"
                ;;
                13)
                    dist_version="trixie"
                ;;
                12)
                    dist_version="bookworm"
                ;;
                11)
                    dist_version="bullseye"
                ;;
                10)
                    dist_version="buster"
                ;;
                9)
                    dist_version="stretch"
                ;;
                8|'Kali Linux 2')
                    dist_version="jessie"
                ;;
                7)
                    dist_version="wheezy"
                ;;
            esac
        elif [ -r /etc/redhat-release ] && [ -z "$lsb_dist" ]; then
            lsb_dist=redhat
            # Extract version from redhat-release file
            dist_version="$(sed 's/.*release \([0-9.]*\).*/\1/' /etc/redhat-release)"
        fi
    fi
}

# Set up sudo command if necessary
setup_sudo() {
    sh_c='sh -c'
    if [ "$user" != 'root' ]; then
        if command_exists sudo; then
            sh_c='sudo -E sh -c'
        elif command_exists su; then
            sh_c='su -c'
        else
            echo "Error: this installer needs the ability to run commands as root."
            echo "We are unable to find either 'sudo' or 'su' available to make this happen."
            exit 1
        fi
    fi
    echo "# Using command executor: $sh_c"
}

# Refine distribution version detection based on the distro
refine_distribution_version() {
    case "$lsb_dist" in
        ubuntu)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/lsb-release ]; then
                
                dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
            fi
        ;;

        debian|raspbian)
            # If we only have a number, map it to a codename for better recognition
            if echo "$dist_version" | grep -qE '^[0-9]+$'; then
                case "$dist_version" in
                    14)
                        dist_version="forky"
                    ;;
                    13)
                        dist_version="trixie"
                    ;;
                    12)
                        dist_version="bookworm"
                    ;;
                    11)
                        dist_version="bullseye"
                    ;;
                    10)
                        # Handle special case for Buster
                        dist_version="buster"
                        if [ "$user" = 'root' ]; then
                            apt-get update --allow-releaseinfo-change || true
                        elif command_exists sudo; then
                            sudo apt-get update --allow-releaseinfo-change || true
                        fi
                    ;;
                    9)
                        dist_version="stretch"
                    ;;
                    8)
                        dist_version="jessie"
                    ;;
                    7)
                        dist_version="wheezy"
                    ;;
                esac
            fi
        ;;

        centos|rhel|fedora|ol)
            # Make sure we have a version number
            if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
                
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/redhat-release ]; then
                dist_version="$(sed 's/.*release \([0-9.]*\).*/\1/' /etc/redhat-release)"
            fi
        ;;

        sles|opensuse)
            if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
            # Fallback for older versions
            if [ -z "$dist_version" ] && [ -r /etc/SuSE-release ]; then
                dist_version="$(grep VERSION /etc/SuSE-release | sed 's/^VERSION = //')"
            fi
            # Ensure version is in the correct format (e.g., 15.4 for SLES 15 SP4)
            if [ -n "$dist_version" ]; then
                # Remove any non-numeric characters except dots
                dist_version="$(echo "$dist_version" | sed 's/[^0-9.]//g')"
            fi
            # Normalize distribution name
            if [ "$lsb_dist" = "sles" ]; then
                lsb_dist="sles"
            elif [ "$lsb_dist" = "opensuse" ]; then
                lsb_dist="opensuse"
            fi
        ;;

        *)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --release | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
                
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
        ;;
    esac
}

# Detect init system 
detect_init_system() {
    if command -v systemctl >/dev/null 2>&1 && [ -d /etc/systemd/system ]; then
        INIT_SYSTEM="systemd"
    elif [ -f /sbin/init ] && /sbin/init --version 2>/dev/null | grep -q upstart; then
        INIT_SYSTEM="upstart"
    elif command -v openrc >/dev/null 2>&1 || [ -f /sbin/openrc ]; then
        INIT_SYSTEM="openrc"
    elif [ -d /etc/s6 ] || command -v s6-svc >/dev/null 2>&1; then
        INIT_SYSTEM="s6"
    elif command -v runit >/dev/null 2>&1 || [ -d /etc/runit ]; then
        INIT_SYSTEM="runit"
    elif [ -d /etc/init.d ]; then
        INIT_SYSTEM="sysvinit"
    else
        INIT_SYSTEM="unknown"
    fi
    export INIT_SYSTEM
    echo "# Detected init system: $INIT_SYSTEM"
}

# Detect package type (deb, rpm, or other)
detect_package_type() {
    case "$lsb_dist" in
        debian|ubuntu|raspbian|mendel)
            PACKAGE_TYPE="deb"
            ;;
        fedora|centos|rhel|ol|sles|opensuse*)
            PACKAGE_TYPE="rpm"
            ;;
        alpine)
            PACKAGE_TYPE="apk"
            ;;
        *)
            if command_exists apt-get || command_exists dpkg; then
                PACKAGE_TYPE="deb"
            elif command_exists yum || command_exists dnf || command_exists zypper; then
                PACKAGE_TYPE="rpm"
            elif command_exists apk; then
                PACKAGE_TYPE="apk"
            else
                PACKAGE_TYPE="other"
            fi
            ;;
    esac
    export PACKAGE_TYPE
    echo "# Detected package type: $PACKAGE_TYPE"
}

# Init function
init() {
    # Detect basic distribution info
    get_distribution
    
    # Set up sudo for privileged commands
    setup_sudo
    
    # Refine version information
    refine_distribution_version
    
    # Check if this is a forked distro
    check_forked
    
    # Detect init system and package type (for universal OS/init support)
    detect_init_system
    detect_package_type
    
    # Print final distribution information
    echo "----------------------------------------"
    echo "Linux Distribution: $lsb_dist"
    echo "Version: $dist_version"
    echo "Init system: $INIT_SYSTEM"
    echo "Package type: $PACKAGE_TYPE"
    echo "----------------------------------------"
    
}
