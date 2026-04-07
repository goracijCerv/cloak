#!/bin/bash
set -e

#1. check operative system
OS = "$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then 
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

echo "Detected OS=$OS, Architecture=$ARCH"

#2. DEFINE THE URL AND DOWNLOAD THE LAST REALSE ON GITHUB
BINARY_NAME="cloak_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/goracijCerv/cloak/releases"

#DONWLOAD FILE
echo "Downloading Cloak from GitHub..."
curl -sl -o cloak "$DOWNLOAD_URL"

#EXECUTION PERMISIONS
chmod +x cloak

#Moving to a global folder will ask for password if it is needed
echo "Installing cloak in /usr/local/bin/ (it could ask for the password)"
sudo mv cloak /usr/local/bin/cloak

echo "Cloak installed successfully!"
echo "Try runing: cloak --help"