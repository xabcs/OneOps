#!/bin/bash

# OneOps 侧边栏样式验证脚本
# 用于快速验证 SxDevOps 样式是否正确应用

echo "🔍 开始验证 OneOps 侧边栏样式实施..."
echo ""

# 检查关键文件是否存在
echo "📋 检查关键样式文件..."

files=(
  "src/styles/scss/sxdevops-theme.scss"
  "src/styles/scss/layout-theme.scss"
  "src/styles/scss/interaction-states.scss"
  "src/styles/scss/sidebar-enhanced.scss"
)

for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    echo "✅ $file - 存在"
  else
    echo "❌ $file - 缺失"
  fi
done

echo ""

# 检查 global.scss 导入是否正确
echo "🔍 验证全局样式导入..."
if grep -q "sidebar-enhanced.scss" src/styles/scss/global.scss; then
  echo "✅ 全局样式导入正确"
else
  echo "❌ 全局样式导入缺失"
fi

echo ""

# 检查组件样式类是否应用
echo "🎨 验证组件样式类..."

if grep -q "global-sider-enhanced" src/layouts/modules/global-sider/index.vue; then
  echo "✅ 侧边栏组件样式类已应用"
else
  echo "❌ 侧边栏组件样式类缺失"
fi

if grep -q "global-logo" src/layouts/modules/global-logo/index.vue; then
  echo "✅ Logo 组件类已应用"
else
  echo "❌ Logo 组件类缺失"
fi

echo ""

# 统计样式文件
echo "📊 样式文件统计..."
style_files=$(find src/styles/scss -name "*.scss" | wc -l)
echo "总样式文件数: $style_files"

echo ""

# 检查是否有构建输出
echo "🏗️  检查构建状态..."
if [ -d "dist" ]; then
  echo "✅ 构建输出目录存在"
  echo "📦 构建产物:"
  ls -lh dist/ | head -10
else
  echo "ℹ️  构建输出目录不存在（需要运行构建命令）"
fi

echo ""
echo "🚀 启动验证..."
echo ""
echo "运行以下命令启动开发服务器："
echo "  cd /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/frontend"
echo "  npm run dev"
echo ""
echo "然后在浏览器中检查："
echo "  ✅ 侧边栏渐变背景"
echo "  ✅ Logo 区域新样式"
echo "  ✅ 菜单项悬停效果"
echo "  ✅ 菜单激活状态"
echo "  ✅ 侧边栏折叠功能"
echo ""
echo "🎉 验证完成！"