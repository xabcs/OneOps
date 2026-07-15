#!/bin/bash
# 自动清理Air缓存并重启

echo "🧹 清理Air缓存..."

# 查找并杀死所有Air进程
pkill -f "air" || true

# 清理tmp目录
if [ -d "backend/tmp" ]; then
    rm -rf backend/tmp
    echo "✅ 已清理 backend/tmp/"
fi

# 清理Go编译缓存（可选，耗时较长）
# go clean -cache

echo "🚀 启动Air..."
air
