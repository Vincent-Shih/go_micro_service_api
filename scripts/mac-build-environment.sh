#!/bin/sh

# Install Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install GVM
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)

source ~/.gvm/scripts/gvm

gvm version

# Install Go
gvm install go1.23.0

gvm use go1.23.0 --default

# Install golang protoc plugin
chmod +x ./scripts/mac-go-grpc-gen.sh