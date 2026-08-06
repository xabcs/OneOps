#!/bin/bash

# 测试权限非硬编码方案

echo "🔍 测试权限非硬编码方案"
echo "========================"
echo ""

echo "1️⃣ 检查数据库迁移文件..."
if [ -f "migrations/add_route_fields_to_permissions.sql" ]; then
    echo "✅ SQL迁移文件存在"

    # 统计路由配置数量
    count=$(grep -c "UPDATE permissions SET route_method" migrations/add_route_fields_to_permissions.sql)
    echo "✅ 配置了 $count 个路由-权限映射"
else
    echo "❌ SQL迁移文件不存在"
    exit 1
fi

echo ""
echo "2️⃣ 检查模型更新..."
if grep -q "RouteMethod" models/permission.go; then
    echo "✅ Permission 模型已添加 RouteMethod 字段"
else
    echo "❌ Permission 模型未更新"
    exit 1
fi

if grep -q "RoutePath" models/permission.go; then
    echo "✅ Permission 模型已添加 RoutePath 字段"
else
    echo "❌ Permission 模型未更新"
    exit 1
fi

echo ""
echo "3️⃣ 检查中间件..."
if [ -f "middleware/permission_db.go" ]; then
    echo "✅ 新的权限中间件已创建"
else
    echo "❌ 权限中间件不存在"
    exit 1
fi

echo ""
echo "4️⃣ 检查路由文件..."
routes=(
    "routes/audit_routes.go"
    "routes/system_routes.go"
    "routes/auth_routes.go"
    "routes/monitoring_routes.go"
    "routes/cmdb_routes.go"
    "routes/k8s_routes.go"
)

for route in "${routes[@]}"; do
    if grep -q "RequirePermissionFromDB" "$route"; then
        echo "✅ $route 已使用统一权限中间件"
    else
        echo "❌ $route 未使用统一权限中间件"
    fi
done

echo ""
echo "5️⃣ 检查权限常量文件..."
if [ -f "routes/permissions.go" ]; then
    echo "⚠️  permissions.go 文件仍存在，应该删除"
else
    echo "✅ permissions.go 文件已删除"
fi

echo ""
echo "========================"
echo "🎉 测试完成！"
echo ""
echo "📝 下一步："
echo "  1. 运行数据库迁移："
echo "     mysql -u root -p ops < migrations/add_route_fields_to_permissions.sql"
echo ""
echo "  2. 重启应用："
echo "     ./oneops 或 go run main.go"
echo ""
echo "  3. 验证权限："
echo "     访问任意 API，权限应该从数据库加载"
