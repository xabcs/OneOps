# 按钮级权限系统 - 完整实现指南

## 概述

这是一个**按钮级权限系统**，提供最细粒度的权限控制，适合企业级应用的复杂权限需求。

## 核心优势

1. **最细粒度控制**：精确到每个按钮、每个API接口
2. **灵活组合**：支持任意权限组合
3. **前后端一致**：统一的权限编码规则
4. **安全可靠**：后端API级别权限检查
5. **易于扩展**：新增权限无需修改核心代码

## 权限编码规范

```
格式：{module}.{resource}.{action}
示例：system.user.create
```

### 权限层级

```
module (模块)      → system, business, auth
resource (资源)    → user, role, permission, order
action (操作)      → view, create, update, delete, export, etc
```

## 数据库设计

### 1. 权限表结构

```sql
CREATE TABLE permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(100) UNIQUE NOT NULL COMMENT '权限编码',
  name VARCHAR(50) NOT NULL COMMENT '权限名称',
  description VARCHAR(200) COMMENT '权限描述',
  module VARCHAR(50) NOT NULL COMMENT '模块',
  resource VARCHAR(50) NOT NULL COMMENT '资源',
  action VARCHAR(50) NOT NULL COMMENT '操作',
  level ENUM('module', 'page', 'button', 'api') DEFAULT 'button',
  parent_id BIGINT DEFAULT NULL,
  sort_order INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 2. API权限映射表

```sql
CREATE TABLE api_permission_mappings (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  method VARCHAR(10) NOT NULL COMMENT 'GET, POST, PUT, DELETE',
  path VARCHAR(200) NOT NULL COMMENT 'API路径',
  permission_code VARCHAR(100) NOT NULL COMMENT '权限编码',
  description VARCHAR(200) COMMENT '描述',
  UNIQUE KEY uk_api (method, path)
);
```

## 前端实现

### 1. 权限指令使用

```vue
<template>
  <!-- 单个权限 -->
  <el-button v-permission="'system.user.create'" type="primary">
    新增用户
  </el-button>

  <!-- 多个权限（满足任一） -->
  <el-button v-permission="['system.user.update', 'system.user.delete']">
    操作
  </el-button>

  <!-- 权限或逻辑 -->
  <el-button v-permission-or="['system.user.create', 'system.user.update']">
    用户操作
  </el-button>

  <!-- 权限与逻辑 -->
  <el-button v-permission-and="['system.user.view', 'system.user.update']">
    编辑用户
  </el-button>
</template>
```

### 2. 权限Hook使用

```vue
<script setup lang="ts">
import { useUserPermissions } from '@/composables/useButtonPermissions'

const {
  canViewUser,
  canCreateUser,
  canUpdateUser,
  canDeleteUser,
  canResetPassword,
  canExportUser,
  canImportUser,
  isUserReadOnly
} = useUserPermissions()

// 组合权限检查
const canEditOrDelete = computed(() =>
  canUpdateUser.value || canDeleteUser.value
)
</script>

<template>
  <div class="user-management">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button
        v-if="canCreateUser"
        type="primary"
        @click="handleCreate"
      >
        新增用户
      </el-button>

      <el-button
        v-if="canDeleteUser"
        type="danger"
        :disabled="!selectedUsers.length"
        @click="handleBatchDelete"
      >
        批量删除
      </el-button>

      <el-button
        v-if="canExportUser"
        @click="handleExport"
      >
        导出用户
      </el-button>

      <el-button
        v-if="canImportUser"
        @click="handleImport"
      >
        导入用户
      </el-button>
    </div>

    <!-- 用户表格 -->
    <el-table :data="users">
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="email" label="邮箱" />

      <el-table-column label="操作" width="300">
        <template #default="{ row }">
          <el-button
            v-if="canViewUser"
            type="info"
            size="small"
            @click="handleView(row)"
          >
            查看
          </el-button>

          <el-button
            v-if="canUpdateUser"
            type="primary"
            size="small"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>

          <el-button
            v-if="canDeleteUser"
            type="danger"
            size="small"
            @click="handleDelete(row)"
          >
            删除
          </el-button>

          <el-dropdown
            v-if="canResetPassword || canLockUser"
            @command="handleMoreCommand($event, row)"
          >
            <el-button type="info" size="small">
              更多
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-if="canResetPassword"
                  command="reset_password"
                >
                  重置密码
                </el-dropdown-item>
                <el-dropdown-item
                  v-if="canLockUser"
                  command="lock"
                >
                  锁定用户
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
```

## 后端实现

### 1. 路由权限配置

```go
package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, userController *controllers.UserController) {
    users := api.Group("/users")
    // 应用全局API权限中间件（可选）
    // users.Use(middlewares.APIPermissionMiddleware())

    {
        // 查看权限
        users.GET("",
            middlewares.RequireView("user"),
            userController.GetUsers,
        )
        users.GET("/:id",
            middlewares.RequireView("user"),
            userController.GetUserById,
        )

        // 创建权限
        users.POST("",
            middlewares.RequireCreate("user"),
            userController.CreateUser,
        )

        // 更新权限
        users.PUT("/:id",
            middlewares.RequireUpdate("user"),
            userController.UpdateUser,
        )

        // 删除权限
        users.DELETE("/:id",
            middlewares.RequireDelete("user"),
            userController.DeleteUser,
        )

        // 特殊操作权限
        users.PUT("/:id/password",
            middlewares.PermissionMiddleware("system.user.reset_password"),
            userController.ResetPassword,
        )

        users.POST("/:id/roles",
            middlewares.PermissionMiddleware("system.user.assign_role"),
            userController.AssignRoles,
        )

        users.DELETE("/:id/roles/:roleId",
            middlewares.PermissionMiddleware("system.user.remove_role"),
            userController.RemoveRole,
        )

        // 批量删除
        users.DELETE("/batch",
            middlewares.PermissionMiddleware("system.user.batch_delete"),
            userController.BatchDelete,
        )

        // 导入导出
        users.GET("/export",
            middlewares.PermissionMiddleware("system.user.export"),
            userController.ExportUsers,
        )

        users.POST("/import",
            middlewares.PermissionMiddleware("system.user.import"),
            userController.ImportUsers,
        )
    }
}
```

### 2. 控制器内部权限检查

```go
package controllers

import (
    "oneops/backend/middlewares"
    "github.com/gin-gonic/gin"
)

type UserController struct {
    // 依赖注入
}

// 获取用户列表
func (ctrl *UserController) GetUsers(c *gin.Context) {
    // 中间件已检查权限，直接处理业务逻辑
    // 但也可以进行额外的业务权限检查
    users, err := ctrl.userService.GetUsers()
    if err != nil {
        c.JSON(500, gin.H{"error": "获取用户列表失败"})
        return
    }

    c.JSON(200, utils.SuccessWithData(users))
}

// 分配角色
func (ctrl *UserController) AssignRoles(c *gin.Context) {
    userID := c.Param("id")

    // 检查是否有分配角色权限
    if !ctrl.hasPermission(c, "system.user.assign_role") {
        c.JSON(403, utils.ErrorForbidden("无权限分配角色"))
        return
    }

    // 业务逻辑...
}

// 辅助方法：检查用户权限
func (ctrl *UserController) hasPermission(c *gin.Context, permission string) bool {
    userID, _ := c.Get("user_id")
    hasPerm, _ := ctrl.permissionService.HasPermission(userID.(uint), permission)
    return hasPerm
}
```

## 角色权限配置示例

### 1. 数据库角色配置

```sql
-- 用户查看员（只能查看，不能操作）
INSERT INTO roles (name, code, description) VALUES
('用户查看员', 'user_viewer', '只能查看用户');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'user_viewer'
AND p.code IN (
    'system',                    -- 系统管理模块
    'system.user',              -- 用户管理页面
    'system.user.view',         -- 查看用户
    'system.user.view_roles',   -- 查看用户角色
    'system.user.view_history' -- 查看用户历史
);

-- 用户管理员（完整用户管理权限）
INSERT INTO roles (name, code, description) VALUES
('用户管理员', 'user_manager', '完整的用户管理权限');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'user_manager'
AND p.resource = 'user'  -- 所有用户相关权限
AND p.level = 'button';

-- 数据导入导出专员
INSERT INTO roles (name, code, description) VALUES
('数据专员', 'data_specialist', '负责数据导入导出');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'data_specialist'
AND p.code IN (
    'system', 'system.user', 'system.user.view',
    'system.user.export', 'system.user.import',
    'system.user.download_template'
);
```

### 2. 角色权限组合建议

```javascript
// 常用角色权限配置模板
const roleTemplates = {
  // 超级管理员
  super_admin: {
    permissions: ['*:*:*'], // 所有权限
    description: '系统超级管理员，拥有所有权限'
  },

  // 模块管理员（系统管理员）
  system_admin: {
    permissions: [
      'system.*.*' // 系统管理模块所有权限
    ],
    description: '系统管理员，管理系统用户、角色、权限'
  },

  // 资源管理员（用户管理员）
  user_admin: {
    permissions: [
      'system.user',
      'system.user.*'
    ],
    description: '用户管理员，管理用户所有操作'
  },

  // 只读用户
  user_viewer: {
    permissions: [
      'system.user.view',
      'system.user.view_roles',
      'system.user.view_history'
    ],
    description: '只能查看用户，不能操作'
  },

  // 特定操作用户
  user_operator: {
    permissions: [
      'system.user.view',
      'system.user.update',
      'system.user.reset_password' // 只能查看、编辑、重置密码
    ],
    description: '用户操作员，有限的用户管理权限'
  },

  // 导入导出专员
  data_operator: {
    permissions: [
      'system.user.view',
      'system.user.export',
      'system.user.import',
      'system.user.download_template'
    ],
    description: '数据导入导出专员'
  }
}
```

## 权限调试工具

### 1. 前端调试

```javascript
// 在浏览器控制台使用
// 获取用户权限信息
window.permissionDebug.showUserInfo()

// 测试用户管理权限
window.permissionDebug.testUserPermissions()

// 检查特定权限
window.permissionDebug.check('system.user.create')

// 模拟权限（开发测试用）
window.permissionDebug.mockPermission('system.user.delete', true)
```

### 2. 后端调试

```go
// 权限调试接口（开发环境）
func (ctrl *DebugController) GetUserPermissions(c *gin.Context) {
    userID, _ := c.Get("user_id")

    permissions, err := ctrl.permissionService.GetUserPermissions(userID.(uint))
    if err != nil {
        c.JSON(500, gin.H{"error": "获取权限失败"})
        return
    }

    c.JSON(200, gin.H{
        "user_id": userID,
        "permissions": permissions,
        "total": len(permissions),
    })
}

// 权限检查测试接口
func (ctrl *DebugController) TestPermission(c *gin.Context) {
    var req struct {
        Permission string `json:"permission" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "参数错误"})
        return
    }

    userID, _ := c.Get("user_id")
    hasPermission, _ := ctrl.permissionService.HasPermission(userID.(uint), req.Permission)

    c.JSON(200, gin.H{
        "user_id": userID,
        "permission": req.Permission,
        "has_permission": hasPermission,
    })
}
```

## 最佳实践

### 1. 权限设计原则

```javascript
// ✅ 推荐：细粒度按钮权限
const goodPermissions = [
  'user.create',           // 创建用户
  'user.update',           // 更新用户
  'user.delete',           // 删除用户
  'user.reset_password',   // 重置密码
  'user.export',           // 导出数据
  'user.import'            // 导入数据
]

// ❌ 不推荐：过于笼统
const badPermissions = [
  'user.manage',           // 用户管理（太笼统）
  'user.basic',            // 基础操作（不明确）
  'user.all'               // 所有操作（不安全）
]
```

### 2. 权限命名规范

```
// 格式：{module}.{resource}.{specific_action}
system.user.create              ✅ 清晰
system.user.reset_password      ✅ 具体
system.user.password_management ❌ 模糊
system.user.all                ❌ 过于宽泛
```

### 3. 前后端权限一致性

```javascript
// ❌ 错误：前端和后端权限不一致
前端: v-permission="'user.create'"
后端: requirePermission('user.new')

// ✅ 正确：统一的权限编码
前端: v-permission="'system.user.create'"
后端: PermissionMiddleware("system.user.create")
```

### 4. 权限检查策略

```go
// ✅ 推荐：中间件检查（统一处理）
func SetupRoutes(r *gin.Engine) {
    users := r.Group("/users")
    users.POST("", PermissionMiddleware("system.user.create"), createUser)
    users.PUT("/:id", PermissionMiddleware("system.user.update"), updateUser)
}

// ⚠️ 可接受：控制器内检查（业务相关）
func (ctrl *UserController) AssignRoles(c *gin.Context) {
    if !ctrl.hasPermission(c, "system.user.assign_role") {
        c.JSON(403, gin.H{"error": "无权限"})
        return
    }
    // 业务逻辑...
}
```

## 维护和扩展

### 1. 新增功能权限

```sql
-- 1. 添加新权限到权限表
INSERT INTO permissions (code, name, description, module, resource, action, level, parent_id, sort_order)
VALUES ('system.user.approve', '审批用户', '审批用户注册', 'system', 'user', 'approve', 'button',
        (SELECT id FROM permissions WHERE code = 'system.user'), 23);

-- 2. 添加API映射
INSERT INTO api_permission_mappings (method, path, permission_code, description)
VALUES ('POST', '/api/users/:id/approve', 'system.user.approve', '审批用户API');

-- 3. 为角色分配权限（如果需要）
INSERT INTO role_permissions (role_id, permission_id)
VALUES (role_id, LAST_INSERT_ID());
```

### 2. 权限清理脚本

```sql
-- 查找未使用的权限
SELECT p.code, p.name
FROM permissions p
LEFT JOIN role_permissions rp ON p.id = rp.permission_id
LEFT JOIN user_permissions up ON p.id = up.permission_id
WHERE rp.id IS NULL AND up.id IS NULL
AND p.created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 清理无效权限（谨慎使用）
DELETE FROM permissions
WHERE id IN (
    SELECT p.id
    FROM permissions p
    LEFT JOIN role_permissions rp ON p.id = rp.permission_id
    LEFT JOIN user_permissions up ON p.id = up.permission_id
    WHERE rp.id IS NULL AND up.id IS NULL
);
```

## 总结

按钮级权限系统提供了最细粒度的权限控制，虽然维护成本相对较高，但对于企业级应用来说是值得的投资。通过合理的设计和工具支持，可以有效降低维护复杂度，提供强大的权限控制能力。

**关键优势：**
- 最细粒度控制
- 高度灵活可组合
- 前后端一致
- 安全可靠
- 易于扩展