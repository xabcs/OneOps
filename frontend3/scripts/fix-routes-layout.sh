#!/bin/bash

# 修复生成路由中的 layout 配置
# 确保 login、403、404、500 页面使用 blank 布局

ROUTES_FILE="src/router/elegant/routes.ts"

if [ -f "$ROUTES_FILE" ]; then
  # 修复 component 字段
  sed -i '' "s/layout\.base\$view\.login/layout.blank\$view.login/g" "$ROUTES_FILE"
  sed -i '' "s/layout\.base\$view\.403/layout.blank\$view.403/g" "$ROUTES_FILE"
  sed -i '' "s/layout\.base\$view\.404/layout.blank\$view.404/g" "$ROUTES_FILE"
  sed -i '' "s/layout\.base\$view\.500/layout.blank\$view.500/g" "$ROUTES_FILE"

  echo "✅ 路由布局已修复"
else
  echo "❌ 路由文件不存在: $ROUTES_FILE"
  exit 1
fi
