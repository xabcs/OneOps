#!/bin/bash

# API权限管理菜单修复脚本
# 解决路由名称不匹配问题

echo "=== API权限管理菜单修复 ==="
echo ""

# 数据库连接信息
DB_HOST="60.191.116.75"
DB_PORT="38089"
DB_USER="root"
DB_PASS="YourPassword"  # 请修改为实际密码
DB_NAME="ops"

echo "问题："
echo "  - 前端路由名称：manage_api-permission (带连字符)"
echo "  - 数据库菜单名称：manage_apipermission (无连字符)"
echo ""

echo "修复步骤："
echo ""

# 方法1：使用MySQL命令行
echo "方法1：执行SQL修复"
echo "----------------------------------------"
cat <<'EOF'
-- 连接数据库后执行以下SQL：

-- 1. 查看当前配置
SELECT id, name, route_name, path
FROM menus
WHERE name LIKE '%API权限%' OR route_name LIKE '%apipermission%';

-- 2. 修复路由名称
UPDATE menus
SET route_name = 'manage_api-permission'
WHERE route_name = 'manage_apipermission';

-- 3. 验证修复结果
SELECT id, name, route_name, path, parent_id, sort
FROM menus
WHERE route_name = 'manage_api-permission';
EOF

echo ""
echo "----------------------------------------"
echo ""

# 方法2：提供连接命令
echo "方法2：使用MySQL客户端连接"
echo "----------------------------------------"
echo "执行命令："
echo "  mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASS $DB_NAME < fix_api_permission_menu.sql"
echo ""

echo "修复完成后："
echo "1. 清除浏览器缓存或重新登录"
echo "2. 菜单应该可以正常跳转到API权限管理页面"
echo ""

echo "=== 完成 ==="
