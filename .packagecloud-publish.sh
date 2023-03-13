#!/usr/bin/env bash

# This script publishes ".deb" and ".rpm" files in the ./dist
# dir to packagecloud. This script relies on the existence
# of the "packagecloud" binary: https://github.com/edgeworx/packagecloud
# The binary can be installed via: go install github.com/edgeworx/packagecloud@v0.1.0

set -e

echo ""
echo "*************** Publish to packagecloud.io ***************"

if [[ -z "$PACKAGECLOUD_TOKEN" ]]; then
    echo "Must provide PACKAGECLOUD_TOKEN envar" 1>&2
    exit 1
fi

repo="${PACKAGECLOUD_REPO}"
echo "Using packagecloud repo: $repo"

pushd ./dist > /dev/null
echo "Using dist dir: $PWD"


failed_push_file='./failed_packagecloud_push'
echo -n "" > $failed_push_file

# add some stutter to avoid overloading packagecloud API
function sleepStutter() {
    sleep "0.$((50 + RANDOM % 300))s"
}

function deb() {
  packages=$(ls | grep .deb)

  echo ""
  echo "*************** Publish .deb ***************"
  echo "deb packages to publish..."
  echo "$packages"
  echo ""
  declare -a distro_versions=(
    "ubuntu/focal" "ubuntu/xenial" "ubuntu/bionic" "ubuntu/trusty"
    "debian/stretch" "debian/buster" "debian/bullseye"
    "raspbian/stretch" "raspbian/buster" "raspbian/bullseye"
    "any/any"
  )

  for package in $packages; do
    for distro_version in "${distro_versions[@]}"; do
      sleepStutter
      repo_full_path="$repo/$distro_version"
      {
        packagecloud push --overwrite "${repo_full_path}" "${package}" 2> >(tee -a $failed_push_file >&2)
      } &
    done
  done

  wait
}

function rpm() {
  packages=$(ls | grep .rpm)

  echo ""
  echo "*************** Publish .rpm ***************"
  echo "rpm packages to publish..."
  echo "$packages"
  echo ""


  declare -a distro_versions=(
    "fedora/23" "fedora/24" "fedora/30" "fedora/31"
    "el/6" "el/7" "el/8"
    "rpm_any/rpm_any"
  )

  for package in $packages; do
    for distro_version in "${distro_versions[@]}"; do
      sleepStutter
      repo_full_path="$repo/$distro_version"
      {
        packagecloud push --overwrite "${repo_full_path}" "${package}" 2> >(tee -a $failed_push_file >&2)
      } &
    done
  done

  wait
}

deb &
sleep 0.1s # give the deb func time to do its output
rpm &

wait

echo ""
if [ -s $failed_push_file ]; then
  # There's content in #failed_push_file... so we need to output it and exit 1.
  echo "*************** Failures (from $failed_push_file) ***************"
  echo ""
  cat $failed_push_file
  echo ""
  exit 1
else
  echo "*************** SUCCESS ***************"
fi