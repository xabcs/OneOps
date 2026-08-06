#!/bin/bash

# 测试权限初始化脚本

echo "🔍 检查权限 SQL 文件..."
if [ -f "migrations/v2_permissions.sql" ]; then
    echo "✅ SQL 文件存在: migrations/v2_permissions.sql"
    
    # 统计权限数量
    count=$(grep -c "INSERT INTO permissions" migrations/v2_permissions.sql)
    echo "✅ 发现 $count 组 INSERT 语句"
    
    # 检查文件大小
    size=$(wc -l < migrations/v2_permissions.sql)
    echo "✅ 文件总行数: $size"
else
    echo "❌ SQL 文件不存在"
    exit 1
fi

echo ""
echo "📊 验证 SQL 语法..."
# 检查是否有语法错误的明显标志
if grep -q "ON DUPLICATE KEY UPDATE" migrations/v2_permissions.sql; then
    echo "✅ 使用 UPSERT 语法，支持幂等性"
else
    echo "⚠️ 未发现 UPSERT 语法"
fi

echo ""
echo "🎉 检查完成！"
echo ""
echo "📝 下一步："
echo "  1. 启动应用：./oneops 或 go run main.go"
echo "  2. 检查日志：权限数据初始化完成 total_permissions=129"
echo "  3. 验证数据库：SELECT COUNT(*) FROM permissions;"
