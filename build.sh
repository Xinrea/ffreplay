#!/bin/bash

# update_actions.sh
sh update_actions.sh

# check go installed
if ! [ -x "$(command -v go)" ]; then
  echo "Warn: 'go' is not installed."
  wget https://git.io/go-installer.sh && bash go-installer.sh
fi

mkdir public

# Get currrent commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)

# Build ffreplay to wasm named public/ffreplay-$commithash.wasm
# Build with credential that read from environment variable
GOOS=js GOARCH=wasm go build -ldflags "-X main.credential=$CREDENTIAL" -o public/ffreplay-$COMMIT_HASH.wasm ./cmd/ffreplay/ffreplay.go
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./public

cp index.html ./public/index.html
cp ffreplay.html ./public/ffreplay.html

# Replace $WASM_RELEASE in ffreplay.html with the current commit hash
if [[ "$OSTYPE" == "darwin"* ]]; then
  # Require gnu-sed.
  if ! [ -x "$(command -v gsed)" ]; then
    echo "Error: 'gsed' is not istalled." >&2
    echo "If you are using Homebrew, install with 'brew install gnu-sed'." >&2
    exit 1
  fi
  SED_CMD=gsed
else
  SED_CMD=sed
fi
${SED_CMD} -i "s/\$WASM_RELEASE/$COMMIT_HASH/g" ./public/ffreplay.html
