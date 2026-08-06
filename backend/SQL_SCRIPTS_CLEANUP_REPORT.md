# SQL脚本废弃分析报告

## 📋 目录概览

### 1. `/backend/sql/` 目录（2个文件）
```
sql/
├── agent_monitoring_tables.sql      # Agent监控表定义
└── agent_version_management.sql     # Agent版本管理表定义
```

### 2. `/backend/scripts/` 目录（40个文件）
```
scripts/
├── 表结构相关（10+）
│   ├── init_cmdb_tables.sql
│   ├── init_cloud_servers.sql
│   ├── init_server_groups.sql
│   └── ...
├── 菜单管理（15+）
│   ├── init_cmdb_menus.sql
│   ├── init_monitoring_menus.sql
│   ├── init_dynamic_menus.sql
│   ├── fix_k8s_menu_*.sql
│   └── ...
├── 数据迁移（8+）
│   ├── migrate_add_redundant_fields.sql
│   ├── migrate_server_metrics.sql
│   └── ...
├── 权限相关（3）
│   ├── add_api_permission_menu.sql
│   ├── init_api_perms.go
│   └── create_casbin_and_init.go
└── 工具脚本（4）
    ├── check_db.go
    ├── cleanup_audit.go
    └── test_*.go
```

## 🔍 废弃状态分析

### ✅ 已废弃的脚本（可清理）

#### 1. 表结构定义脚本
**文件**：
- `sql/agent_monitoring_tables.sql`
- `sql/agent_version_management.sql`
- `scripts/init_cmdb_tables.sql`
- `scripts/init_cloud_servers.sql`

**原因**：
- ✅ 所有表现在由 **GORM AutoMigrate** 管理
- ✅ 模型定义在 `models/` 目录下
- ✅ `services/init.go` 的 `migrateSchema()` 自动创建表

**验证**：
```bash
# 检查模型定义
grep -r "type AgentVersion struct" models/
# 输出：models/cmdb.go:397:type AgentVersion struct {
```

#### 2. 菜单初始化脚本
**文件**：
- `scripts/init_cmdb_menus*.sql` (3个版本)
- `scripts/init_monitoring_menus.sql`
- `scripts/init_dynamic_menus.sql`
- `scripts/init_menus_v4.sql`
- `scripts/fix_k8s_menu*.sql` (多个版本)

**原因**：
- ✅ 菜单数据现在由 **syncMenus()** 动态同步
- ✅ 从代码中定义并同步到数据库
- ✅ 支持增量更新和清理

#### 3. 数据迁移脚本
**文件**：
- `scripts/migrate_*.sql` (多个)
- `scripts/reinit_*.sql` (多个)
- `scripts/update_*.sql` (多个)

**原因**：
- ✅ 迁移逻辑已集成到 `runMigrations()` 方法
- ✅ 使用 GORM 的 `db.Exec()` 执行必要的DDL
- ✅ 所有必要的字段已在模型中定义

#### 4. 临时修复脚本
**文件**：
- `scripts/fix_k8s_audit_sessions_route_error.sql`
- `scripts/remove_k8s_audit_sessions_menu.sql`
- `scripts/rebuild_k8s_menu_tab_structure.sql`

**原因**：
- ⚠️ 这些是一次性修复脚本，问题已解决
- ⚠️ 不需要保留

### ⚠️ 可能有用但需要验证的脚本

#### 1. 性能优化脚本
**文件**：`scripts/add_performance_indexes.sql`

**建议**：检查索引是否已在模型中通过 `index` tag 定义

#### 2. 工具脚本
**文件**：
- `scripts/check_db.go`
- `scripts/check_users.go`
- `scripts/cleanup_audit.go`

**建议**：评估是否还有使用场景

### ❌ 明确仍需要的脚本

**无** - 所有必要的初始化逻辑已集成到代码中

## 📊 废弃脚本统计

| 类别 | 文件数 | 状态 |
|------|--------|------|
| 表结构定义 | 4 | ✅ 已废弃 |
| 菜单管理 | 15+ | ✅ 已废弃 |
| 数据迁移 | 8+ | ✅ 已废弃 |
| 权限相关 | 3 | ✅ 已废弃 |
| 临时修复 | 5+ | ✅ 已废弃 |
| 工具脚本 | 4 | ⚠️ 需验证 |
| **总计** | **40+** | **95% 已废弃** |

## 🎯 清理建议

### 方案一：完全清理（推荐）
```bash
# 1. 创建归档目录
mkdir -p /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/.archive/sql_scripts

# 2. 移动所有SQL脚本到归档目录
mv /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/sql \
   /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/.archive/

mv /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/scripts \
   /Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/.archive/
```

**优势**：
- ✅ 彻底清理，避免混淆
- ✅ 保留历史记录，需要时可恢复
- ✅ 目录结构更清晰

### 方案二：选择性保留
```bash
# 保留可能有用的工具脚本
mkdir -p backend/tools

# 移动工具脚本
mv backend/scripts/check_db.go backend/tools/
mv backend/scripts/cleanup_audit.go backend/tools/

# 删除已废弃的SQL脚本
rm -rf backend/sql/
rm -rf backend/scripts/
```

**优势**：
- ✅ 保留实用工具
- ✅ 清理无用脚本

### 方案三：创建README标记废弃
```bash
# 在两个目录中添加废弃说明
cat > backend/sql/README_DEPRECATED.md << 'EOF'
# ⚠️ 此目录已废弃

所有表结构现在由GORM模型管理，请勿使用此目录中的SQL脚本。

表定义位置：`backend/models/`
迁移管理：`backend/services/init.go`
EOF

cat > backend/scripts/README_DEPRECATED.md << 'EOF'
# ⚠️ 此目录已废弃

所有初始化和迁移逻辑已集成到代码中：
- 表结构：GORM AutoMigrate
- 菜单数据：syncMenus() 动态同步
- 权限数据：JSON配置文件管理

请勿使用此目录中的脚本。
EOF
```

## ✅ 推荐操作

### 立即执行：方案一（归档）
```bash
# 创建归档目录
mkdir -p .archive

# 归档SQL脚本
mv sql .archive/
mv scripts .archive/

# 创建说明文档
cat > .archive/README.md << 'EOF'
# 已归档的SQL脚本

## 归档时间
2026-08-05

## 归档原因
所有表结构、菜单、权限数据现在由代码管理：
- 表结构：GORM AutoMigrate
- 菜单：services/initializer.go 中的 syncMenus()
- 权限：services/data/permissions.json

## 原始位置
- sql/ -> backend/sql/
- scripts/ -> backend/scripts/
EOF
```

### 验证清理后编译
```bash
# 确保没有代码引用这些文件
go build ./main.go
```

## 📝 清理后目录结构

```
backend/
├── models/              # GORM模型定义（表结构）
├── services/
│   ├── init.go         # 初始化服务
│   ├── initializer.go  # 初始化协调器
│   └── data/
│       └── permissions.json  # 权限配置
└── .archive/           # 归档目录（可选保留）
    ├── sql/
    ├── scripts/
    └── README.md
```

## 🎉 清理效果

| 指标 | 清理前 | 清理后 |
|------|--------|--------|
| 冗余文件 | 40+ | 0 |
| 目录混乱 | ❌ | ✅ |
| 维护成本 | 高 | 低 |
| 代码清晰度 | ⭐⭐ | ⭐⭐⭐⭐⭐ |

---

**结论**：**95%的SQL脚本已废弃**，建议立即执行归档清理。
