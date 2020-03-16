#!/bin/sh

# Bump project version
if [ -z "$1" ]; then
    echo "Provide a version argument e.g. 1.2.3-beta"
fi
if [ ! -f "version" ]; then
    echo "File node found: $(pwd)/version"
    exit 1
fi

version="$1"

# Extract version numbers and suffix
major=$(echo $version | sed 's/\..*\..*-.*//g')
minor=$(echo $version | sed -E 's/.*\.(.*)\..*/\1/g')
patch=$(echo $version | sed -E 's/.*\..*\.(.*)-.*/\1/g')
suffix=$(echo $version | sed 's/.*-//g')

# Update version file
sed -i.bkp "s/MAJOR=.*/MAJOR=$major/g" "version"
sed -i.bkp "s/MINOR=.*/MINOR=$minor/g" "version"
sed -i.bkp "s/PATCH=.*/PATCH=$patch/g" "version"
sed -i.bkp "s/SUFFIX=.*/SUFFIX=-$suffix/g" "version"
rm "version.bkp"

# Pull modules
make modules

# Update Makefile
sed -i.bkp -E "s/(.*-X.*Tag=).*/\1$version/g" Makefile
sed -i.bkp -E "s/(.*-X.*Version=).*/\1$version/g" Makefile
rm Makefile.bkp
# Update pipeline
sed -i.bkp -E "s/(_image:.*focal-freedom.*:).*/\1$version'/g" azure-pipelines.yaml
sed -i.bkp -E "s/(_version: ).*/\1'$version'/g" azure-pipelines.yaml
rm azure-pipelines.yaml.bkp