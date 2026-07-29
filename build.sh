#!/bin/bash
set -euo pipefail

BINARY_NAME="dianshu-mcp"
OUTPUT_DIR="bin"

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "=== 交叉编译 ${BINARY_NAME} ==="

# macOS Apple Silicon (arm64)
echo "[1/5] darwin/arm64 ..."
GOOS=darwin GOARCH=arm64 go build -o "${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64" .
echo "  -> ${OUTPUT_DIR}/${BINARY_NAME}-darwin-arm64"

# macOS Intel (amd64)
echo "[2/5] darwin/amd64 ..."
GOOS=darwin GOARCH=amd64 go build -o "${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64" .
echo "  -> ${OUTPUT_DIR}/${BINARY_NAME}-darwin-amd64"

# Linux amd64
echo "[3/5] linux/amd64 ..."
GOOS=linux GOARCH=amd64 go build -o "${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64" .
echo "  -> ${OUTPUT_DIR}/${BINARY_NAME}-linux-amd64"

# Linux arm64
echo "[4/5] linux/arm64 ..."
GOOS=linux GOARCH=arm64 go build -o "${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64" .
echo "  -> ${OUTPUT_DIR}/${BINARY_NAME}-linux-arm64"

# Windows amd64
echo "[5/5] windows/amd64 ..."
GOOS=windows GOARCH=amd64 go build -o "${OUTPUT_DIR}/${BINARY_NAME}-windows-amd64.exe" .
echo "  -> ${OUTPUT_DIR}/${BINARY_NAME}-windows-amd64.exe"

echo ""
echo "编译完成，产物在 ${OUTPUT_DIR}/"
ls -lh "${OUTPUT_DIR}/"
