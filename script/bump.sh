#!/bin/sh

# Bump project version
if [ -z "$1" ]; then
    echo "Provide a version argument e.g. 1.2.3-beta"
fi
if [ ! -f "version" ]; then
    echo "File node found: $(pwd)/version"
    exit 1
fi

# Extract version numbers and suffix
major=$1
minor=$2
patch=$3
suffix=$4
version=$1.$2.$3$4

# Update version file
sed -i.bkp "s/MAJOR=.*/MAJOR=$major/g" "version"
sed -i.bkp "s/MINOR=.*/MINOR=$minor/g" "version"
sed -i.bkp "s/PATCH=.*/PATCH=$patch/g" "version"
sed -i.bkp "s/SUFFIX=.*/SUFFIX=$suffix/g" "version"
rm "version.bkp"

# Update Makefile
sed -i.bkp -E "s/(.*iofog-go-sdk\/v2@).*/\1v$version/g" Makefile
sed -i.bkp -E "s/(.*iofog-operator\/v2@).*/\1v$version/g" Makefile
sed -i.bkp -E "s/(.*-X.*Tag=).*/\1$version/g" Makefile
sed -i.bkp -E "s/(.*-X.*Version=).*/\1$version/g" Makefile
sed -i.bkp -E "s/(.*-X.*repo=).*/\1iofog/g" Makefile
rm Makefile.bkp

# Update pipeline
for file in azure-pipelines.yaml test/env.sh; do
    sed -i.bkp -E "s/(gcr\.io\/focal-freedom.*:).*/\1$version'/g" $file
    sed -i.bkp -E "s/(_version: ).*/\1'$version'/g" $file
    sed -i.bkp -E "s/(_VERSION=').*/\1$version'/g" $file
    rm $file.bkp
done

# Pull modules
make modules
