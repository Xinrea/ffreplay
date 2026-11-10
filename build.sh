#!/bin/sh

set -e

if [ -z "${CREDENTIAL:-}" ]; then
  echo "Error: CREDENTIAL is required. Set it to client_id:secret before building." >&2
  exit 1
fi

# update_actions.sh
sh update_actions.sh

mkdir -p public
# Clean generated files without removing asset directories.
find public -maxdepth 1 -type f -delete

# Publish the asset tree as individual static files. The wasm build fetches
# them at runtime, so browsers and CDNs can cache them across releases
# instead of re-downloading them inside the bundled wasm.
rm -rf public/asset
cp -R asset public/asset
find public/asset \( -name '.DS_Store' -o -name '*.psd' -o -name '*.go' -o -name '.gitkeep' \) -delete
find public/asset -type d -empty -delete

# Pack the ability icons into a texture atlas: the browser build loads one
# image and slices each icon out of it instead of requesting them one by one.
rm -rf public/asset/abilities
go run ./cmd/abilityatlas -in asset/abilities -out public/asset/abilities

if command -v sha256sum >/dev/null 2>&1; then
  HASH_CMD="sha256sum"
else
  HASH_CMD="shasum -a 256"
fi

# Hash the published asset tree: the wasm appends this version to every asset
# URL, so long-lived caches are invalidated exactly when the files change.
ASSET_VERSION=$(
  find public/asset -type f -print | LC_ALL=C sort | while IFS= read -r file; do
    printf '%s\n' "$file"
    cat "$file"
  done | $HASH_CMD | cut -c1-12
)
echo "Asset version: $ASSET_VERSION"

# Get currrent commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)

# Build ffreplay to wasm named public/ffreplay-$commithash.wasm
# Build with credential that read from environment variable
GOOS=js GOARCH=wasm go build \
  -ldflags "-X main.credential=$CREDENTIAL -X github.com/Xinrea/ffreplay.assetVersion=$ASSET_VERSION" \
  -o public/ffreplay-$COMMIT_HASH.wasm ./cmd/ffreplay/ffreplay.go
WASM_SIZE=$(wc -c < public/ffreplay-$COMMIT_HASH.wasm | tr -d '[:space:]')

# Include the wasm content hash in the release name: builds from the same
# commit with local changes then get new chunk URLs, so browsers and CDNs
# never serve a stale wasm for a new release. Unchanged builds keep their URLs.
WASM_DIGEST=$($HASH_CMD public/ffreplay-$COMMIT_HASH.wasm | cut -c1-12)
WASM_RELEASE="$COMMIT_HASH-$WASM_DIGEST"
echo "WASM release: $WASM_RELEASE ($WASM_SIZE bytes)"

# split the wasm file into 10 chunks
split -n 10 -d public/ffreplay-$COMMIT_HASH.wasm public/ffreplay-$WASM_RELEASE.wasm.chunk
# remove the original wasm file
rm public/ffreplay-$COMMIT_HASH.wasm

cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./public

cp index.html ./public/index.html
cp ffreplay.html ./public/ffreplay.html

# Replace $WASM_RELEASE in ffreplay.html with the commit + content hash
case "$(uname -s)" in
  Darwin)
    # Require gnu-sed.
    if ! [ -x "$(command -v gsed)" ]; then
      echo "Error: 'gsed' is not istalled." >&2
      echo "If you are using Homebrew, install with 'brew install gnu-sed'." >&2
      exit 1
    fi
    SED_CMD=gsed
    ;;
  *)
    SED_CMD=sed
    ;;
esac

${SED_CMD} -i "s/\$WASM_RELEASE/$WASM_RELEASE/g" ./public/ffreplay.html
${SED_CMD} -i "s/\$WASM_SIZE/$WASM_SIZE/g" ./public/ffreplay.html

# get client_id from env CREDENTIAL, example: client_id:secret

CLIENT_ID=$(printf '%s' "$CREDENTIAL" | cut -d: -f1)
${SED_CMD} -i "s/\$CLIENT_ID/$CLIENT_ID/g" ./public/index.html
