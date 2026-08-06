# 权限初始化迁移说明

## 📝 变更概述

将权限初始化从 JSON 文件方式改为 SQL 迁移方式，提高可读性和可维护性。

## 🔄 变更对比

### 之前的方式（JSON）

```go
// ❌ 旧方式：解析 JSON 文件
permissions, err := i.dataLoader.LoadPermissions()
for _, perm := range permissions {
    // 逐条插入或更新
    db.Create(&perm)
}
```

**缺点：**
- ❌ JSON 文件可读性差
- ❌ 需要解析和转换代码
- ❌ 维护困难，容易出错
- ❌ 不支持 SQL 特性（如事务、批量操作）

### 现在的方式（SQL）

```go
// ✅ 新方式：执行 SQL 文件
sqlFile := "migrations/v2_permissions.sql"
content, err := os.ReadFile(sqlFile)
db.Exec(string(content))
```

**优点：**
- ✅ SQL 文件清晰易读
- ✅ 支持语法高亮和 IDE 工具
- ✅ 使用 UPSERT 语法，可安全重复执行
- ✅ 批量操作，性能更好
- ✅ 支持事务和复杂 SQL 特性

## 📂 文件变更

### 新增文件

```
backend/migrations/v2_permissions.sql  # ✅ 完整的权限初始化 SQL（129个权限）
```

### 删除文件

```
backend/services/data/permissions.json  # ❌ 已删除
```

### 修改文件

```
backend/services/initializer.go  # 🔧 改用 SQL 方式初始化
```

### 备份文件

```
backend/database/init_permissions.sql.deprecated  # 📦 旧 SQL 文件备份
```

## 🚀 使用方法

### 自动初始化（推荐）

启动应用时自动执行：

```bash
./oneops

# 或开发模式
go run main.go
```

初始化日志：
```
2026-08-06 开始数据库初始化流程...
2026-08-06 阶段1：开始数据库模式迁移...
2026-08-06 阶段2：执行 SQL 迁移...
2026-08-06 阶段3：初始化基础数据...
2026-08-06 阶段4：初始化模块数据...
2026-08-06 从 SQL 文件初始化权限数据...
2026-08-06 权限数据初始化完成 total_permissions=129
```

### 手动执行（可选）

如需手动初始化权限：

```bash
# 方式1：MySQL 客户端
mysql -u root -p ops < backend/migrations/v2_permissions.sql

# 方式2：登录后执行
mysql -u root -p ops
source /path/to/OneOps/backend/migrations/v2_permissions.sql;
```

## 📊 权限统计

### 各模块权限数量

| 模块 | 权限数量 | 说明 |
|------|---------|------|
| 系统管理 | 27 | 菜单、角色、用户、权限、属性、路由 |
| 审计中心 | 6 | 登录日志、操作日志、系统事件、统计 |
| 授权中心 | 28 | 应用、用户、用户组、绑定、映射 |
| 资产管理 | 33 | 服务器、Agent、分组、业务、机房、标签、会话、策略 |
| 监控中心 | 16 | 数据、告警、任务、报告 |
| K8s管理 | 15 | 集群、权限、资源、诊断 |
| **总计** | **129** | 完整的系统权限定义 |

### SQL 文件结构

```sql
-- 使用 UPSERT 语法
INSERT INTO permissions (code, name, description, ...) VALUES
('system.user.list', '用户列表', '查看用户列表', ...)
ON DUPLICATE KEY UPDATE
  name=VALUES(name),
  description=VALUES(description),
  updated_at=NOW();
```

**特点：**
- ✅ 幂等性：可重复执行，不会重复插入
- ✅ 更新策略：已存在的权限会更新名称和描述
- ✅ 自动时间戳：使用 NOW() 函数

## 🔧 技术细节

### SQL 语法说明

#### ON DUPLICATE KEY UPDATE

```sql
INSERT INTO permissions (code, name, ...) VALUES (...)
ON DUPLICATE KEY UPDATE
  name=VALUES(name),
  description=VALUES(description);
```

**作用：**
- 如果 `code` 不存在 → 插入新记录
- 如果 `code` 已存在 → 更新指定字段

### 批量插入优化

```sql
-- ✅ 推荐：批量插入
INSERT INTO permissions (...) VALUES
('system.user.list', ...),
('system.user.create', ...),
('system.user.update', ...);
```

vs

```go
// ❌ 旧方式：逐条插入
for _, perm := range permissions {
    db.Create(&perm)  // N 次数据库操作
}
```

**性能提升：**
- 批量插入：1 次数据库操作
- 逐条插入：N 次数据库操作

## 📚 维护指南

### 添加新权限

1. 编辑 SQL 文件：

```sql
-- 在对应模块下添加
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.new_feature.view', '查看新功能', '查看新功能', 'system', 'new_feature', 'view', 3, 130, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), updated_at=NOW();
```

2. 重启应用，自动初始化

### 修改权限

直接修改 SQL 文件中的权限定义，重启应用即可更新。

### 删除权限

**注意：** 删除权限需要谨慎操作：

```sql
-- 1. 先删除关联关系
DELETE FROM role_permissions WHERE permission_id IN (
  SELECT id FROM permissions WHERE code = 'system.old_permission.view'
);

-- 2. 再删除权限
DELETE FROM permissions WHERE code = 'system.old_permission.view';
```

## ✅ 验证方法

### 查询权限总数

```sql
SELECT COUNT(*) as total FROM permissions;
-- 预期结果: 129
```

### 按模块统计

```sql
SELECT module, COUNT(*) as count
FROM permissions
GROUP BY module
ORDER BY module;
```

预期结果：
```
+--------+-------+
| module | count |
+--------+-------+
| audit  |     6 |
| auth   |    28 |
| cmdb   |    33 |
| k8s    |    15 |
| monitor|    16 |
| system |    27 |
+--------+-------+
```

### 查看权限详情

```sql
SELECT code, name, module, resource, action, level
FROM permissions
WHERE module = 'system' AND resource = 'user'
ORDER BY sort_order;
```

## 🎯 优势总结

| 维度 | JSON 方式 | SQL 方式 |
|------|----------|---------|
| **可读性** | ⭐⭐ 需要解析 | ⭐⭐⭐⭐⭐ 原生 SQL |
| **可维护性** | ⭐⭐ 需懂 Go | ⭐⭐⭐⭐⭐ 标准 SQL |
| **性能** | ⭐⭐⭐ 逐条插入 | ⭐⭐⭐⭐⭐ 批量操作 |
| **幂等性** | ⭐⭐⭐ 需手动实现 | ⭐⭐⭐⭐⭐ UPSERT |
| **工具支持** | ⭐⭐⭐ JSON 工具 | ⭐⭐⭐⭐⭐ IDE/SQL 客户端 |
| **事务支持** | ⭐⭐ 应用层 | ⭐⭐⭐⭐⭐ 数据库原生 |

## 🔗 相关文档

- [权限系统设计文档](../docs/permission-design.md)
- [路由权限修复报告](../docs/ROUTE_PERMISSION_FIX_REPORT.md)
- [数据库迁移指南](../migrations/README.md)

---

**更新时间**: 2026-08-06
**版本**: v2.0.0
**影响范围**: 权限初始化逻辑
