#!/bin/bash

set -e

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT_SHA=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

CLEAN_VERSION=$(echo "$VERSION" | sed 's/-dirty$//')

mkdir -p dist

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

LDFLAGS="-X github.com/benwsapp/rlgl/cmd.Version=${VERSION} -X github.com/benwsapp/rlgl/cmd.CommitSHA=${COMMIT_SHA} -X github.com/benwsapp/rlgl/cmd.BuildTime=${BUILD_TIME}"

echo "Building rlgl version ${VERSION}"
echo "Commit: ${COMMIT_SHA}"
echo "Build time: ${BUILD_TIME}"
echo "Clean version: ${CLEAN_VERSION}"
echo "Targets: ${#TARGETS[@]}"
echo "Cosign signing: keyless (GitHub OIDC)"
echo ""

for target in "${TARGETS[@]}"; do
    IFS='/' read -r os arch <<< "$target"

    echo "Building for ${os}/${arch}..."

    release_dir="dist/rlgl-${CLEAN_VERSION}-${os}-${arch}"
    mkdir -p "$release_dir"

    binary_name="rlgl-${CLEAN_VERSION}"

    export GOOS="$os"
    export GOARCH="$arch"

    # Build directly into the release directory
    go build -ldflags "$LDFLAGS" -o "$release_dir/${binary_name}" main.go

    cd dist
    tar -czf "rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz" "rlgl-${CLEAN_VERSION}-${os}-${arch}/"
    cd ..

    # Keyless signing with GitHub OIDC
    echo "  Signing with cosign (keyless)..."
    cosign sign-blob --yes \
        --bundle "dist/rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz.bundle" \
        "dist/rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz"

    if [ $? -eq 0 ]; then
        echo "  ✓ Signed: rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz"
    else
        echo "  ✗ Failed to sign: rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz"
    fi

    size=$(stat -f%z "$release_dir/${binary_name}" 2>/dev/null || stat -c%s "$release_dir/${binary_name}" 2>/dev/null || echo "unknown")

    rm -rf "$release_dir"

    echo "  ✓ Built: rlgl-${CLEAN_VERSION}-${os}-${arch}.tar.gz (${size} bytes)"
done

echo ""
echo "Scanning for vulnerabilities..."
trivy fs --severity HIGH,CRITICAL --exit-code 1 .
if [ $? -eq 0 ]; then
    echo "  ✓ No HIGH or CRITICAL vulnerabilities found!"
else
    echo "  ✗ Vulnerabilities detected! Check output above."
    exit 1
fi

echo ""
echo "Generating Software Bill of Materials (SBOM)..."

echo "  Generating source code SBOMs..."
trivy fs --format spdx-json --output "dist/sbom.spdx.json" .
trivy fs --format cyclonedx --output "dist/sbom.cyclonedx.json" .
echo "  ✓ SBOM generation complete!"

echo ""
echo "Generating checksums..."
cd dist
sha256sum rlgl-*.tar.gz > checksums.txt 2>/dev/null || shasum -a 256 rlgl-*.tar.gz > checksums.txt
cd ..
echo "  ✓ Checksums saved to dist/checksums.txt"

echo ""
echo "Build complete!"
