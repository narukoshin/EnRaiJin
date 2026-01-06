#!/bin/bash

PLUGIN_NAME="agentix"
PLUGIN_FILE="./plugins/$PLUGIN_NAME/$PLUGIN_NAME.go"
VERSION=$(grep -oP 'Version string = "\K[^"]+' $PLUGIN_FILE)

VERSION_DIR=$RELEASE_DIR/$PLUGIN_NAME

# Creating version dir
if [ ! -d "$VERSION_DIR" ]; then
    mkdir -p "$VERSION_DIR"
fi

echo "[-] Compiling plugin '$PLUGIN_NAME'..."

go build -buildmode=plugin -o $VERSION_DIR/$PLUGIN_NAME.so $PLUGIN_FILE

PLUGIN_ARCHIVE=$VERSION_DIR/$PLUGIN_NAME-$VERSION-universal-amd64.tar.gz
tar zcvf $PLUGIN_ARCHIVE -C $VERSION_DIR $PLUGIN_NAME.so > /dev/null
echo "+ Plugin '$PLUGIN_NAME' compiled to $PLUGIN_ARCHIVE"

# Deleting files that are not necessary anymore
rm -f $VERSION_DIR/$PLUGIN_NAME.so