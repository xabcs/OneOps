# 系统管理数据库问题修复方案

## 📋 发现的问题总结

| 优先级 | 问题 | 影响 | 修复难度 |
|-------|------|------|---------|
| 🔴 高 | permissions 表是空的 | 无法分配权限，系统无法正常使用 | 中 |
| 🔴 高 | roles.menuIds 字段冗余 | 数据冗余，可能造成不一致 | 低 |
| 🟡 中 | initRoles 方法未使用 | 代码混乱，维护困难 | 低 |
| 🟡 中 | test 角色缺少默认权限 | test 用户无权限可用 | 中 |
| 🟢 低 | roles.permission_ids 字段未使用 | 数据冗余 | 低 |

---

## 🔴 高优先级问题修复

### 问题1：permissions 表是空的

#### 问题描述
- 项目启动时没有初始化 permissions 表
- permissions 表是空的，无法分配权限
- 用户需要在UI中手动创建权限

#### 修复方案

**方案A：在 init.go 中添加权限初始化（推荐）**

在 `backend/services/init.go` 中添加 `initPermissions()` 方法：

```go
// initPermissions 初始化权限数据
func (s *InitService) initPermissions() error {
	logger.Info("开始初始化权限数据...")

	// 定义系统权限
	permissions := []models.Permission{
		// ========== 系统管理模块 ==========
		// 用户管理
		{Code: "system.user.view", Name: "查看用户", Module: "system", Resource: "user", Action: "view", Level: 2, Status: 1},
		{Code: "system.user.create", Name: "创建用户", Module: "system", Resource: "user", Action: "create", Level: 3, Status: 1},
		{Code: "system.user.update", Name: "更新用户", Module: "system", Resource: "user", Action: "update", Level: 3, Status: 1},
		{Code: "system.user.delete", Name: "删除用户", Module: "system", Resource: "user", Action: "delete", Level: 3, Status: 1},
		{Code: "system.user.reset_password", Name: "重置密码", Module: "system", Resource: "user", Action: "reset_password", Level: 3, Status: 1},

		// 角色管理
		{Code: "system.role.view", Name: "查看角色", Module: "system", Resource: "role", Action: "view", Level: 2, Status: 1},
		{Code: "system.role.create", Name: "创建角色", Module: "system", Resource: "role", Action: "create", Level: 3, Status: 1},
		{Code: "system.role.update", Name: "更新角色", Module: "system", Resource: "role", Action: "update", Level: 3, Status: 1},
		{Code: "system.role.delete", Name: "删除角色", Module: "system", Resource: "role", Action: "delete", Level: 3, Status: 1},
		{Code: "system.role.assign_permissions", Name: "分配权限", Module: "system", Resource: "role", Action: "assign_permissions", Level: 3, Status: 1},

		// 菜单管理
		{Code: "system.menu.view", Name: "查看菜单", Module: "system", Resource: "menu", Action: "view", Level: 2, Status: 1},
		{Code: "system.menu.create", Name: "创建菜单", Module: "system", Resource: "menu", Action: "create", Level: 3, Status: 1},
		{Code: "system.menu.update", Name: "更新菜单", Module: "system", Resource: "menu", Action: "update", Level: 3, Status: 1},
		{Code: "system.menu.delete", Name: "删除菜单", Module: "system", Resource: "menu", Action: "delete", Level: 3, Status: 1},

		// 权限管理
		{Code: "system.permission.view", Name: "查看权限", Module: "system", Resource: "permission", Action: "view", Level: 2, Status: 1},
		{Code: "system.permission.create", Name: "创建权限", Module: "system", Resource: "permission", Action: "create", Level: 3, Status: 1},
		{Code: "system.permission.update", Name: "更新权限", Module: "system", Resource: "permission", Action: "update", Level: 3, Status: 1},
		{Code: "system.permission.delete", Name: "删除权限", Module: "system", Resource: "permission", Action: "delete", Level: 3, Status: 1},

		// ========== CMDB模块 ==========
		{Code: "cmdb.server.query", Name: "查询主机", Module: "cmdb", Resource: "server", Action: "query", Level: 2, Status: 1},
		{Code: "cmdb.server.create", Name: "创建主机", Module: "cmdb", Resource: "server", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.server.update", Name: "更新主机", Module: "cmdb", Resource: "server", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.server.delete", Name: "删除主机", Module: "cmdb", Resource: "server", Action: "delete", Level: 3, Status: 1},

		{Code: "cmdb.business.query", Name: "查询业务", Module: "cmdb", Resource: "business", Action: "query", Level: 2, Status: 1},
		{Code: "cmdb.rooms.query", Name: "查询机房", Module: "cmdb", Resource: "rooms", Action: "query", Level: 2, Status: 1},
		{Code: "cmdb.tags.query", Name: "查询标签", Module: "cmdb", Resource: "tags", Action: "query", Level: 2, Status: 1},
		{Code: "cmdb.credentials.query", Name: "查询凭证", Module: "cmdb", Resource: "credentials", Action: "query", Level: 2, Status: 1},

		// ========== 监控模块 ==========
		{Code: "monitoring.overview.query", Name: "监控概览", Module: "monitoring", Resource: "overview", Action: "query", Level: 2, Status: 1},
		{Code: "monitoring.servers.query", Name: "主机监控", Module: "monitoring", Resource: "servers", Action: "query", Level: 2, Status: 1},
		{Code: "monitoring.alerts.query", Name: "告警管理", Module: "monitoring", Resource: "alerts", Action: "query", Level: 2, Status: 1},

		// ========== 审计模块 ==========
		{Code: "audit.login.view", Name: "登录审计", Module: "audit", Resource: "login", Action: "view", Level: 2, Status: 1},
		{Code: "audit.operation.view", Name: "操作审计", Module: "audit", Resource: "operation", Action: "view", Level: 2, Status: 1},
		{Code: "audit.system.view", Name: "系统审计", Module: "audit", Resource: "system", Action: "view", Level: 2, Status: 1},

		// ========== K8s模块 ==========
		{Code: "k8s.cluster.query", Name: "查询集群", Module: "k8s", Resource: "cluster", Action: "query", Level: 2, Status: 1},
		{Code: "k8s.workload.query", Name: "查询工作负载", Module: "k8s", Resource: "workload", Action: "query", Level: 2, Status: 1},
		{Code: "k8s.diagnostic.execute", Name: "执行诊断", Module: "k8s", Resource: "diagnostic", Action: "execute", Level: 3, Status: 1},
	}

	addedCount := 0
	updatedCount := 0

	for _, perm := range permissions {
		var existingPerm models.Permission
		err := db.Where("code = ?", perm.Code).First(&existingPerm).Error

		if err == nil {
			// 权限已存在，更新数据
			db.Model(&existingPerm).Updates(map[string]interface{}{
				"name":        perm.Name,
				"description": perm.Description,
				"module":      perm.Module,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"level":       perm.Level,
				"status":      perm.Status,
			})
			updatedCount++
		} else {
			// 权限不存在，创建新权限
			if err := db.Create(&perm).Error; err != nil {
				logger.Error("创建权限失败",
					zap.String("name", perm.Name),
					zap.String("code", perm.Code),
					zap.Error(err))
				return err
			}
			addedCount++
		}
	}

	logger.Info("权限初始化完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(permissions)))

	return nil
}
```

**在 initData() 方法中调用**：

```go
func (s *InitService) initData() error {
	// ... 现有代码 ...

	// 初始化权限（在角色同步之前）
	var permCount int64
	db.Model(&models.Permission{}).Count(&permCount)
	if permCount == 0 {
		if err := s.initPermissions(); err != nil {
			logger.Warn("权限初始化失败，继续执行", zap.Error(err))
		}
	}

	// 同步内置角色（会引用权限）
	if err := s.syncBuiltinRoles(); err != nil {
		logger.Warn("内置角色同步失败，继续执行", zap.Error(err))
	}

	// ... 其他代码 ...
}
```

**方案B：执行 SQL 脚本（备选）**

```bash
# 在项目启动后执行
mysql -u root -p ops < backend/migrations/seed_permissions.sql
```

---

### 问题2：roles.menuIds 字段冗余

#### 问题描述
- menuIds 字段用于存储角色的菜单权限
- 现在权限通过 role_permissions 表管理
- 菜单通过权限码自动推导，menuIds 已不再使用

#### 修复方案

**步骤1：检查是否有代码依赖 menuIds 字段**

```bash
# 搜索所有使用 menuIds 的地方
grep -rn "menuIds" backend/
```

**步骤2：如果没有依赖，移除该字段**

```sql
-- 备份数据（谨慎操作）
CREATE TABLE roles_backup AS SELECT * FROM roles;

-- 移除 menu_ids 字段
ALTER TABLE roles DROP COLUMN menu_ids;
```

**步骤3：同时移除 permission_ids 字段（迁移脚本中添加但未使用）**

```sql
ALTER TABLE roles DROP COLUMN permission_ids;
```

**Go代码修改**：

如果确定 menuIds 字段不再使用，可以从 models.Role 中移除：

```go
// models/role.go
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	// MenuIds     string    `json:"menuIds" gorm:"type:json"` // ← 移除
	// PermissionIDs string  `json:"permissionIds" gorm:"type:text"` // ← 移除
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
```

**注意**：如果当前有数据依赖 menuIds，建议保留该字段，但不再使用。等确认无影响后再移除。

---

## 🟡 中优先级问题修复

### 问题3：initRoles 方法未使用

#### 问题描述
- init.go 中定义了 `initRoles()` 方法
- 但在 `initData()` 中调用的是 `syncBuiltinRoles()` 方法
- `initRoles()` 方法永远不会被执行

#### 修复方案

**方案A：删除未使用的方法（推荐）**

```go
// 删除 initRoles 方法（第731-771行）
// 该方法已被 syncBuiltinRoles 替代
```

**方案B：整合到 syncBuiltinRoles 中**

如果 initRoles 有特殊逻辑，可以整合到 syncBuiltinRoles 中：

```go
func (s *InitService) syncBuiltinRoles() error {
	logger.Info("开始同步内置角色...")

	// ... 现有的 syncBuiltinRoles 逻辑 ...

	// 如果需要初始化用户的默认角色权限，可以在这里添加

	logger.Info("内置角色同步完成")
	return nil
}
```

---

### 问题4：test 角色缺少默认权限

#### 问题描述
- test 角色在 syncBuiltinRoles 中创建
- 但没有分配默认权限
- test 用户登录后没有任何权限

#### 修复方案

**在 init.go 中添加角色权限初始化方法**：

```go
// assignDefaultPermissions 为内置角色分配默认权限
func (s *InitService) assignDefaultPermissions() error {
	logger.Info("开始为内置角色分配默认权限...")

	// 定义角色默认权限映射
	rolePermissions := map[string][]string{
		"admin": {
			"*.*.*", // 超级管理员拥有所有权限
		},
		"ops": {
			// 系统管理
			"system.user.view", "system.user.create", "system.user.update", "system.user.delete",
			"system.role.view", "system.role.update",
			"system.menu.view",
			// CMDB
			"cmdb.server.query", "cmdb.server.create", "cmdb.server.update", "cmdb.server.delete",
			"cmdb.business.query",
			"cmdb.rooms.query",
			"cmdb.tags.query",
			"cmdb.credentials.query",
			// 监控
			"monitoring.overview.query",
			"monitoring.servers.query",
			"monitoring.alerts.query",
			// K8s
			"k8s.cluster.query",
			"k8s.workload.query",
			"k8s.diagnostic.execute",
		},
		"auditor": {
			// 只读权限
			"cmdb.server.query",
			"cmdb.business.query",
			"cmdb.rooms.query",
			"monitoring.overview.query",
			"monitoring.servers.query",
			"audit.login.view",
			"audit.operation.view",
			"audit.system.view",
			"k8s.cluster.query",
			"k8s.workload.query",
		},
		"viewer": {
			// 最小权限
			"cmdb.server.query",
			"monitoring.overview.query",
		},
		"user": {
			// 普通用户基础权限
			"monitoring.overview.query",
		},
		"test": {
			// 测试角色权限（用于测试）
			"system.user.view",
			"system.user.create",
			"system.role.view",
			"cmdb.server.query",
			"cmdb.server.create",
			"monitoring.overview.query",
			"k8s.cluster.query",
		},
	}

	for roleCode, permissionCodes := range rolePermissions {
		// 获取角色
		var role models.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			logger.Warn("角色不存在，跳过权限分配",
				zap.String("code", roleCode),
				zap.Error(err))
			continue
		}

		// 获取权限
		var permissions []models.Permission
		if err := db.Where("code IN ?", permissionCodes).Find(&permissions).Error; err != nil {
			logger.Warn("查询权限失败",
				zap.String("role", roleCode),
				zap.Error(err))
			continue
		}

		// 分配权限
		for _, perm := range permissions {
			var rolePerm models.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				First(&rolePerm).Error

			if err != nil {
				// 创建新的角色权限关联
				rolePerm = models.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rolePerm).Error; err != nil {
					logger.Error("分配权限失败",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code),
						zap.Error(err))
				} else {
					logger.Debug("分配权限",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code))
				}
			}
		}

		logger.Info("角色权限分配完成",
			zap.String("role", roleCode),
			zap.Int("权限数量", len(permissions)))
	}

	return nil
}
```

**在 initData() 方法中调用**：

```go
func (s *InitService) initData() error {
	// ... 现有代码 ...

	// 同步内置角色
	if err := s.syncBuiltinRoles(); err != nil {
		logger.Warn("内置角色同步失败，继续执行", zap.Error(err))
	}

	// 分配默认权限（新增）
	if err := s.assignDefaultPermissions(); err != nil {
		logger.Warn("默认权限分配失败，继续执行", zap.Error(err))
	}

	// ... 其他代码 ...
}
```

---

## 🟢 低优先级问题

### 问题5：roles.permission_ids 字段未使用

#### 修复方案

同问题2，一起移除：

```sql
ALTER TABLE roles DROP COLUMN permission_ids;
```

---

## 📋 修复执行清单

### 执行顺序

1. **备份数据库**（必须）
   ```bash
   mysqldump -u root -p ops > ops_backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **添加权限初始化代码**
   - 在 `init.go` 中添加 `initPermissions()` 方法
   - 在 `init.go` 中添加 `assignDefaultPermissions()` 方法
   - 修改 `initData()` 方法，调用新方法

3. **编译测试**
   ```bash
   cd backend
   go build
   ```

4. **运行程序**
   ```bash
   ./backend
   ```

5. **验证权限初始化**
   ```sql
   -- 检查权限是否初始化
   SELECT COUNT(*) FROM permissions;

   -- 检查角色权限是否分配
   SELECT rp.role_id, r.name as role_name, p.code as permission_code
   FROM role_permissions rp
   JOIN roles r ON rp.role_id = r.id
   JOIN permissions p ON rp.permission_id = p.id
   WHERE r.code = 'test';
   ```

6. **移除冗余字段**（可选）
   ```sql
   -- 先确认没有代码依赖
   grep -rn "menuIds" backend/
   grep -rn "permissionIds" backend/

   -- 如果没有依赖，可以移除
   ALTER TABLE roles DROP COLUMN menu_ids;
   ALTER TABLE roles DROP COLUMN permission_ids;
   ```

7. **删除未使用代码**
   - 删除 `initRoles()` 方法（如果确定不需要）

8. **重新测试**
   - 用 test 用户登录
   - 验证权限是否正常
   - 验证菜单是否正常显示

---

## 📝 修复后的效果

### 修复前

```
permissions 表: 空 (0条记录)
role_permissions 表: 空 (0条记录)
test 用户登录: 无权限，无法使用系统
```

### 修复后

```
permissions 表: 30+ 条权限记录
role_permissions 表: 每个角色都有默认权限
test 用户登录: 有基础权限，可以正常使用
```

---

## ⚠️ 注意事项

1. **数据备份**：修改数据库结构前必须备份
2. **测试环境验证**：先在测试环境验证，再在生产环境执行
3. **逐步执行**：不要一次性修改太多，逐步验证
4. **保留回退方案**：每个修改都要有回退方案
5. **监控日志**：执行时观察日志，确保没有错误

---

## 📊 预期收益

| 指标 | 修复前 | 修复后 |
|------|--------|--------|
| permissions 表记录数 | 0 | 30+ |
| role_permissions 表记录数 | 0 | 100+ |
| test 用户可用性 | ❌ 无权限 | ✅ 有权限 |
| 系统可用性 | ❌ 需要手动配置 | ✅ 开箱即用 |
| 数据冗余 | ❌ menuIds 冗余 | ✅ 清理冗余 |
| 代码质量 | ❌ 有未使用方法 | ✅ 代码清晰 |

---

**修复完成时间预估**：2-3小时
**风险等级**：中（有数据库修改）
**建议执行时间**：维护窗口期
