#!/bin/bash

APP_NAME="enraijin"
VERSION_FILE="./pkg/app/app.go"
VERSION=$(grep -oP 'const Version string = "\K[^"]+' $VERSION_FILE)
SAMPLE_CONFIG="config.sample.yml"

RELEASE_DIR="releases"
VERSION_DIR=$RELEASE_DIR/$APP_NAME/$VERSION

export $RELEASE_DIR

# Check if there is a version in $VERSION
if [ -z "$VERSION" ]; then
    echo "Version not found in $VERSION_FILE"
    exit 1
fi

# Creating relases folder if doesn't exist
if [ ! -d "$RELEASE_DIR" ]; then
    mkdir "$RELEASE_DIR"
fi

# Creating version folder if doesn't exist
if [ ! -d "$VERSION_DIR" ]; then
    mkdir -p "$VERSION_DIR"
fi

# Printing a message that binaries are being compiled
echo "[-] Compiling binaries... (This might take a while)"

cp $SAMPLE_CONFIG $VERSION_DIR

WINDOWS_BIN="$APP_NAME.exe"
LINUX_BIN=$APP_NAME
MAC_BIN=$APP_NAME

# --- Windows ---

GOOS=windows GOARCH=amd64 go build -o $VERSION_DIR/$WINDOWS_BIN main.go
WINDOWS_ARCHIVE=$VERSION_DIR/$APP_NAME-$VERSION-windows-amd64.tar.gz
tar zcvf $WINDOWS_ARCHIVE -C $VERSION_DIR $WINDOWS_BIN $SAMPLE_CONFIG > /dev/null
echo "+ Windows binary compiled to $WINDOWS_ARCHIVE"

# --- Linux ---

LINUX_ARCHIVE=$VERSION_DIR/$APP_NAME-$VERSION-linux-amd64.tar.gz
GOOS=linux GOARCH=amd64 go build -o $VERSION_DIR/$LINUX_BIN main.go
tar zcvf $LINUX_ARCHIVE -C $VERSION_DIR $LINUX_BIN $SAMPLE_CONFIG > /dev/null
echo "+ Linux binary compiled to $LINUX_ARCHIVE"

# --- MacOS ---

MAC_ARCHIVE=$VERSION_DIR/$APP_NAME-$VERSION-macos-arm64.tar.gz
GOOS=darwin GOARCH=arm64 go build -o $VERSION_DIR/$MAC_BIN main.go
tar zcvf $MAC_ARCHIVE -C $VERSION_DIR $MAC_BIN $SAMPLE_CONFIG > /dev/null
echo "+ MacOS binary compiled to $MAC_ARCHIVE"

# Deleting files that are not necessary anymore
rm -f $VERSION_DIR/$WINDOWS_BIN
rm -f $VERSION_DIR/$LINUX_BIN
rm -f $VERSION_DIR/$MAC_BIN
rm -f $VERSION_DIR/$SAMPLE_CONFIG