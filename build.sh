#!/bin/bash

mkdir public

# Get currrent commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)

# Build ffreplay to wasm named public/ffreplay-$commithash.wasm
# Build with credential that read from environment variable
GOOS=js GOARCH=wasm go build -ldflags "-X main.credential=$CREDENTIAL" -o public/ffreplay-$COMMIT_HASH.wasm ./cmd/ffreplay/ffreplay.go
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./public

# Copy asset folder into public
cp -r ./asset ./public

cp index.html ./public/index.html
cp ffreplay.html ./public/ffreplay.html

# Replace $WASM_RELEASE in ffreplay.html with the current commit hash
sed -i '' "s/\$WASM_RELEASE/$COMMIT_HASH/g" ./public/ffreplay.html