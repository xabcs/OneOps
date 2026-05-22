#!/bin/bash
# 构建 OneOps Agent 二进制文件
# 输出到 ../agent-binaries/ 目录（供 OneOps 后端 SCP 部署使用）

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/../agent-binaries"

mkdir -p "$OUTPUT_DIR"

echo "Building OneOps Agent..."
cd "$SCRIPT_DIR"

echo "  -> linux/amd64"
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/oneops-agent-linux-amd64" .

echo "  -> linux/arm64"
GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/oneops-agent-linux-arm64" .

echo "Build complete:"
ls -lh "$OUTPUT_DIR/"
