#!/bin/bash

# 前端路由错误修复脚本
# 错误：Error: No match for {"name":"manage_apipermission","params":{}}
# 正确：manage_api-permission

echo "======================================"
echo "前端路由错误修复"
echo "======================================"
echo ""

echo "问题分析："
echo "  错误的路由名称: manage_apipermission (缺少连字符)"
echo "  正确的路由名称: manage_api-permission (保留连字符)"
echo ""

echo "检查步骤："
echo ""

# 1. 检查前端路由配置
echo "1. 检查前端路由配置..."
echo "--------------------------------------"
grep "manage.*api.*permission" /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/frontend/src/router/elegant/transform.ts
echo ""

# 2. 检查数据库菜单配置
echo "2. 检查数据库菜单配置..."
echo "--------------------------------------"
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre -e "
SELECT
    id,
    name,
    path,
    CASE
        WHEN path = '/manage/api-permission' THEN '✅ 正确'
        ELSE '❌ 需要检查'
    END AS status
FROM menus
WHERE name = 'API权限管理';
" 2>/dev/null
echo ""

echo "======================================"
echo "解决方案"
echo "======================================"
echo ""

echo "方案1：清除浏览器缓存（最有效）"
echo "--------------------------------------"
echo "1. Chrome/Edge: Cmd+Shift+Delete (Mac) 或 Ctrl+Shift+Delete (Windows)"
echo "2. 选择: '缓存图片和文件' 和 'Cookie及其他网站数据'"
echo "3. 点击: 清除数据"
echo "4. 重新登录系统"
echo ""

echo "方案2：清除LocalStorage"
echo "--------------------------------------"
echo "1. 打开浏览器开发者工具 (F12)"
echo "2. 切换到 Console 标签"
echo "3. 执行以下命令："
echo ""
echo "   localStorage.clear();"
echo "   sessionStorage.clear();"
echo "   location.reload();"
echo ""

echo "方案3：使用无痕模式"
echo "--------------------------------------"
echo "1. Chrome: Cmd+Shift+N (Mac) 或 Ctrl+Shift+N (Windows)"
echo "2. 访问系统并登录"
echo ""

echo "======================================"
echo "验证修复"
echo "======================================"
echo ""
echo "修复后，点击菜单中的'API权限管理'，应该能正常跳转"
echo ""

echo "如果问题仍然存在："
echo "1. 检查浏览器控制台的具体错误信息"
echo "2. 检查Network标签中的API响应"
echo "3. 联系开发团队"
