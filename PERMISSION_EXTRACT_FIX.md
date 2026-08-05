# 紧急修复：权限码提取逻辑错误

## 问题根源

**当前代码（错误）：**

```go
// 提取权限列表
permissions := s.extractPermissions(menuTree)
```

从**菜单树的 permission 字段**提取权限码，而不是从 `role_permissions` 表！

**导致的问题：**

1. 用户绑定 ops 角色，权限正确（system.user, system.role等）
2. 但 `extractPermissions` 从菜单树提取权限
3. 菜单树包含未授权的菜单
4. 用户得到错误的权限码列表

## 正确的实现

**需要同时做两件事：**

### 1. 从权限码推导菜单

```go
// 获取用户的所有权限码
var userPermissions []string
for _, role := range roles {
    var rolePerms []models.RolePermission
    db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
    for _, rp := range rolePerms {
        if rp.Permission.Code != "" {
            userPermissions = append(userPermissions, rp.Permission.Code)
        }
    }
}

// 从权限码提取 resource
allowedResources := make(map[string]bool)
for _, permCode := range userPermissions {
    parts := strings.Split(permCode, ":")
    if len(parts) >= 2 {
        resource := parts[1]
        allowedResources[resource] = true
    }
}

// 查找 resource 匹配的菜单
for _, menu := range allMenus {
    if menu.Resource != "" && allowedResources[menu.Resource] {
        menuIDs[menu.ID] = true
        // 标记父菜单
        for _, m := range allMenus {
            if m.ID == menu.ParentID {
                menuIDs[m.ID] = true
            }
        }
    }
}
```

### 2. 从 role_permissions 表获取权限码

```go
// ✅ 正确：从 role_permissions 表获取
permissions := make([]string, 0)
if isAdmin {
    permissions = append(permissions, "*:*:*")
} else {
    for _, role := range roles {
        var rolePerms []models.RolePermission
        db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
        for _, rp := range rolePerms {
            if rp.Permission.Code != "" {
                permissions = append(permissions, rp.Permission.Code)
            }
        }
    }
}

// ❌ 错误：从菜单树提取
// permissions := s.extractPermissions(menuTree)
```

## 手动修复步骤

由于时间关系，建议手动修改 `backend/services/rbac.go`：

### 步骤1：找到第 95-120 行，替换为：

```go
	// 获取所有菜单
	var allMenus []*models.Menu
	err = db.Where("status = 1").Order("sort ASC").Find(&allMenus).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// 🚀 新功能：从权限码自动推导菜单权限
	// 同时获取用户的真实权限码列表
	permissions := make([]string, 0)
	if isAdmin {
		// 管理员：拥有所有菜单和通配符权限
		menuIDs = make(map[uint]bool)
		for _, menu := range allMenus {
			menuIDs[menu.ID] = true
		}
		permissions = append(permissions, "*:*:*")
	} else {
		// 1. 获取用户的所有权限码（从 role_permissions 表）
		for _, role := range roles {
			var rolePerms []models.RolePermission
			db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
			for _, rp := range rolePerms {
				if rp.Permission.Code != "" {
					permissions = append(permissions, rp.Permission.Code)
				}
			}
		}

		// 2. 从权限码提取 resource 列表
		allowedResources := make(map[string]bool)
		for _, permCode := range permissions {
			parts := strings.Split(permCode, ":")
			if len(parts) >= 2 {
				resource := parts[1]
				allowedResources[resource] = true
			}
		}

		// 3. 根据 resource 匹配菜单
		for _, menu := range allMenus {
			if menu.Resource != "" && allowedResources[menu.Resource] {
				menuIDs[menu.ID] = true
				// 标记父菜单
				for _, m := range allMenus {
					if m.ID == menu.ParentID {
						menuIDs[m.ID] = true
					}
				}
			}
		}
	}
```

### 步骤2：删除第 124-128 行

删除这几行（因为权限已经在上面正确获取了）：

```go
// 删除这些行：
	// 提取权限列表
	permissions := s.extractPermissions(menuTree)
	if isAdmin {
		permissions = append(permissions, "*:*:*")
	}
```

### 步骤3：编译和重启

```bash
cd backend
go build
pkill -f "go run main.go"
go run main.go
```

## 数据库验证

```sql
-- 确认 test 用户绑定 ops 角色
SELECT username, role_ids FROM users WHERE username = 'test';
-- 应该返回：test | [2]

-- 确认 ops 角色权限正确
SELECT p.code FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = 2;
-- 应该返回：system, system.user, system.role, system.menu, system.user.view, system.role.view, system.menu.view

-- 确认 ops 角色的 menu_ids 为空
SELECT menu_ids FROM roles WHERE code = 'ops';
-- 应该返回：[]
```

## 预期结果

修复后，test 用户登录应该看到：

**后端日志：**
```
[登录调试-RBAC] 最终权限码 数量=7 ["system", "system.user", "system.role", ...]
```

**前端日志：**
```javascript
🐛 [路由Store] 后端返回的路由数据: [
  { id: "1", path: "/home", name: "home" },
  { id: "6", path: "/manage", name: "manage", children: [...] }
]
```

**左侧菜单：**
- 首页
- 系统管理
  - 用户管理
  - 角色管理
  - 菜单管理
