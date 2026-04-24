#!/bin/bash
# 审计中间件修复脚本

echo "=== 操作审计功能修复 ==="
echo ""

# 1. 备份原文件
echo "1. 备份原文件..."
cp /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go.backup
echo "   已备份到 audit.go.backup"
echo ""

# 2. 应用修复
echo "2. 应用修复..."
cat > /tmp/new_getModule.txt << 'EOF'
// getModuleFromPath 从路径解析模块名称（智能识别）
func (m *AuditMiddleware) getModuleFromPath(path string) string {
	// 如果不是API路径，直接返回
	if !strings.HasPrefix(path, "/api/") {
		return "其他"
	}

	// 去掉/api前缀
	apiPath := strings.TrimPrefix(path, "/api")

	// 基于 API 路径前缀智能识别模块
	if strings.HasPrefix(apiPath, "/login") || strings.HasPrefix(apiPath, "/logout") || strings.HasPrefix(apiPath, "/user") {
		return "认证管理"
	}
	if strings.HasPrefix(apiPath, "/system/menus") {
		return "菜单管理"
	}
	if strings.HasPrefix(apiPath, "/system/roles") {
		return "角色管理"
	}
	if strings.HasPrefix(apiPath, "/system/users") {
		return "用户管理"
	}
	if strings.HasPrefix(apiPath, "/audit") {
		return "审计管理"
	}
	if strings.HasPrefix(apiPath, "/monitoring") {
		return "监控中心"
	}
	if strings.HasPrefix(apiPath, "/tasks") {
		return "任务管理"
	}
	if strings.HasPrefix(apiPath, "/servers") {
		return "服务器管理"
	}
	if strings.HasPrefix(apiPath, "/containers") {
		return "容器管理"
	}
	if strings.HasPrefix(apiPath, "/certificates") {
		return "证书管理"
	}
	if strings.HasPrefix(apiPath, "/system") {
		return "系统管理"
	}

	// 从路径中提取第一级作为模块名
	parts := strings.Split(strings.Trim(apiPath, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return "其他"
}
EOF

# 删除旧函数（171-216行）
sed -i.tmp '171,216d' /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go

# 在171行插入新函数
sed -i.tmp '170r /tmp/new_getModule.txt' /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go

# 清理临时文件
rm -f /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go.tmp

echo "   已应用修复"
echo ""

# 3. 验证修改
echo "3. 验证修改..."
if grep -q "监控中心" /Users/mpm/doc/go/OneOpsV2/OneOps/backend/middlewares/audit.go; then
    echo "   ✓ 修复成功！"
else
    echo "   ✗ 修复失败，请手动修改"
    exit 1
fi
echo ""

# 4. 生成 SQL 清理脚本
echo "4. 生成历史数据清理SQL..."
cat > /tmp/cleanup_audit_modules.sql << 'SQLEOF'
-- 更新历史数据中的 module 字段
UPDATE operation_logs
SET module = CASE
    WHEN path LIKE '/api/monitoring%' THEN '监控中心'
    WHEN path LIKE '/api/audit%' THEN '审计管理'
    WHEN path LIKE '/api/system/menus%' THEN '菜单管理'
    WHEN path LIKE '/api/system/roles%' THEN '角色管理'
    WHEN path LIKE '/api/system/users%' THEN '用户管理'
    WHEN path LIKE '/api/user%' THEN '用户管理'
    WHEN path LIKE '/api/login%' THEN '认证管理'
    WHEN path LIKE '/api/logout%' THEN '认证管理'
    ELSE '其他'
END
WHERE module = '未知模块' OR module = '';
SQLEOF

echo "   SQL 脚本已生成: /tmp/cleanup_audit_modules.sql"
echo ""

# 5. 重启后端提示
echo "5. 重启后端服务..."
echo "   请重启后端服务以应用修改："
echo "   cd /Users/mpm/doc/go/OneOpsV2/OneOps/backend"
echo "   air  # 或 go run main.go"
echo ""

echo "=== 修复完成 ==="
echo ""
echo "新的模块映射规则："
echo "  /api/monitoring*  → 监控中心"
echo "  /api/audit*       → 审计管理"
echo "  /api/system/menus* → 菜单管理"
echo "  /api/system/roles* → 角色管理"
echo "  /api/system/users* → 用户管理"
echo "  /api/user*        → 用户管理"
echo "  /api/login*       → 认证管理"
echo "  /api/logout*      → 认证管理"
