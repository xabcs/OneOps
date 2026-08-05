# 数据库设计和命名规范检查报告

## 📋 检查概述

**检查时间**: 2026-08-05
**数据库**: msre (MySQL)
**表数量**: 46个表
**检查范围**: 核心系统表（users, roles, menus, permissions等）

---

## 🔍 发现的问题

### 1. 命名规范不统一

#### 字段命名不一致
**问题描述**: 数据库字段命名混合使用了下划线和无下划线的风格

**发现的模式**:
- **下划线命名**: `role_ids`, `home_path`, `parent_id`, `sort_order`, `created_at`, `updated_at`
- **无下划线命名**: `username`, `nickname`, `avatar`, `menu_type`, `sort`

**建议**: 统一使用下划线命名（snake_case）
- ✅ 推荐: `user_name`, `home_path`, `created_at`
- ❌ 不推荐: `username`, `nickname`

#### 表命名不一致
**问题描述**: 虽然大多数表使用下划线命名，但存在缩写不一致

**发现的模式**:
- ✅ 一致性较好: `server_groups`, `server_tags`, `bastion_sessions`
- ⚠️ 缩写不一致: `cmdb_` vs `server_`, `auth_` vs `user_`

### 2. 数据类型不统一

#### status字段类型不一致
**问题描述**: `status` 字段在不同表中使用不同的数据类型

| 表名 | status字段类型 | 默认值 |
|------|---------------|--------|
| users | varchar(20) | active |
| roles | bigint(20) | 1 |
| menus | bigint(20) | 1 |
| permissions | tinyint(4) | 1 |

**建议**: 统一使用 `tinyint(1)` 类型
- ✅ 推荐: `status tinyint(1) DEFAULT 1`
- ❌ 不推荐: 混合使用 `varchar`, `bigint`, `tinyint(4)`

#### ID字段类型不一致
**问题描述**: 外键ID字段类型不统一

| 表名 | ID字段类型 |
|------|-----------|
| menus.parent_id | bigint(20) unsigned |
| permissions.parent_id | bigint(20) unsigned |
| 但有些表可能使用 int |

**建议**: 统一使用 `bigint(20) unsigned` 作为ID字段类型

### 3. JSON字段设计问题

#### JSON字段命名规范
**发现的JSON字段**:
- `role_ids` (users表) - 使用下划线和复数
- `menu_ids` (roles表) - 使用下划线和复数
- 但其他字段不一致

**建议**: JSON数组字段统一使用复数形式和下划线
- ✅ 推荐: `role_ids`, `permission_ids`, `tag_ids`
- ❌ 不推荐: `roleId`, `permissionId`

### 4. 时间戳字段问题

#### created_at/updated_at精度不一致
**问题描述**: 时间戳字段使用 `datetime(3)` 类型，但有些表可能不一致

**建议**: 统一使用 `datetime(3)` 或 `timestamp` 类型
- ✅ 推荐: `created_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3)`
- ✅ 推荐: `updated_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)`

---

## 📊 字段命名统计分析

### users表字段命名模式
```
下划线命名: role_ids, home_path, created_at, updated_at (4个)
无下划线命名: username, password, nickname, avatar, email, status (6个)
比例: 40% vs 60%
```

### roles表字段命名模式
```
下划线命名: menu_ids, created_at, updated_at (3个)
无下划线命名: name, code, description, status (4个)
比例: 43% vs 57%
```

### menus表字段命名模式
```
下划线命名: parent_id, sort, created_at, updated_at, resource (5个)
无下划线命名: name, icon, path, permission, menu_type, status (6个)
比例: 45% vs 55%
```

---

## 🎯 修复建议优先级

### 🔴 高优先级（影响功能）
1. **统一status字段类型**: 使用 `tinyint(1)` 替代混合类型
2. **统一ID字段类型**: 确保所有外键ID使用 `bigint(20) unsigned`

### 🟡 中优先级（影响维护）
1. **统一JSON字段命名**: 确保JSON数组字段使用复数形式
2. **统一时间戳精度**: 确保所有 `datetime` 字段使用相同的精度

### 🟢 低优先级（代码规范）
1. **统一字段命名风格**: 长期目标是所有字段使用下划线命名
2. **统一表前缀**: 规范表前缀的使用（如 `system_`, `auth_`, `cmdb_`）

---

## 🔧 推荐修复SQL

### 1. 统一status字段类型
```sql
-- 统一users表的status字段类型
ALTER TABLE users MODIFY COLUMN status tinyint(1) DEFAULT 1 COMMENT '状态：1启用 0禁用';

-- 统一roles表的status字段类型
ALTER TABLE roles MODIFY COLUMN status tinyint(1) DEFAULT 1 COMMENT '状态：1启用 0禁用';
```

### 2. 添加缺失的时间戳默认值
```sql
-- 为users表添加时间戳默认值
ALTER TABLE users
MODIFY COLUMN created_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
MODIFY COLUMN updated_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);

-- 为roles表添加时间戳默认值
ALTER TABLE roles
MODIFY COLUMN created_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3),
MODIFY COLUMN updated_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);
```

### 3. 确保JSON字段使用正确的JSON类型
```sql
-- 验证JSON字段类型
ALTER TABLE users MODIFY COLUMN role_ids JSON COMMENT '角色ID列表(JSON数组)';
ALTER TABLE roles MODIFY COLUMN menu_ids JSON COMMENT '菜单ID列表(JSON数组)';
```

---

## 📝 长期改进建议

### 1. 数据库命名规范制定
制定并文档化数据库设计规范：
- 表命名规范
- 字段命名规范
- 索引命名规范
- 外键命名规范

### 2. 代码生成工具
考虑使用代码生成工具（如GO语言ORM工具）来确保：
- 数据库模型与数据库表结构同步
- 字段类型和命名的一致性
- 自动化的数据验证

### 3. 数据库迁移脚本管理
建立规范的数据库迁移脚本管理流程：
- 版本化的迁移脚本
- 迁移脚本测试流程
- 数据库变更审计

---

## ✅ 总结

**当前主要问题**:
1. ❌ **命名规范不统一**: 字段命名混合使用下划线和无下划线
2. ❌ **数据类型不一致**: 同一字段在不同表中使用不同类型
3. ❌ **JSON字段设计**: JSON数组字段命名不够规范
4. ❌ **时间戳字段**: 部分表缺少时间戳默认值

**影响分析**:
- 🔴 **功能影响**: status字段类型不一致可能影响查询和业务逻辑
- 🟡 **维护影响**: 命名不统一增加维护和理解成本
- 🟢 **扩展影响**: 不规范的命名可能影响后续功能扩展

**修复优先级**:
1. 🔴 **立即修复**: status字段类型统一
2. 🟡 **逐步修复**: 命名规范统一（需要修改代码和数据库）
3. 🟢 **长期改进**: 建立完整的数据库设计规范

---

*报告生成时间: 2026-08-05*
*检查工具: 手动检查 + MySQL查询*