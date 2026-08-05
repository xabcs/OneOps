# 权限授权完整流程

## 📋 概述

OneOps 系统实现了完整的 **RBAC（基于角色的访问控制）** 权限管理体系，支持**菜单级、页面级、按钮级**的多层级权限控制。

## 🔄 完整的授权流程

### 1️⃣ 权限字典管理（系统管理员）

**位置：** 系统管理 → 权限管理
**URL：** `/manage/permission`

**功能：** 定义系统中有哪些权限

**权限级别：**
- **模块级** (level=1) - 控制整个模块的访问
- **页面级** (level=2) - 控制页面的访问权限
- **按钮级** (level=3) - 控制按钮操作的权限
- **API级** (level=4) - 控制API接口的访问

**权限编码规范：**
```
格式：module.resource.action
示例：
- system.user.view      - 查看用户
- system.user.create    - 创建用户
- system.user.update    - 更新用户
- system.user.delete    - 删除用户
```

### 2️⃣ 角色权限分配（系统管理员）

**位置：** 系统管理 → 角色管理
**URL：** `/manage/role`

**操作步骤：**

#### A. 分配菜单权限
1. 在角色列表中找到目标角色
2. 点击 **"菜单权限"** 按钮
3. 在树形结构中勾选该角色可访问的菜单
4. 点击确定保存

**效果：** 角色拥有菜单权限后，用户登录后只能看到被授权的菜单项

#### B. 分配按钮权限
1. 在角色列表中找到目标角色
2. 点击 **"按钮权限"** 按钮
3. 在权限树中勾选该角色可执行的操作
4. 点击确定保存

**效果：** 用户在页面中只能看到被授权的按钮

**权限示例：**
```
✅ 查看 - system.user.view      → 用户可以看到"查看"按钮
✅ 创建 - system.user.create    → 用户可以看到"新增"按钮
✅ 更新 - system.user.update    → 用户可以看到"编辑"按钮
✅ 删除 - system.user.delete    → 用户可以看到"删除"按钮
```

### 3️⃣ 用户角色分配（系统管理员）

**位置：** 系统管理 → 用户管理
**URL：** `/manage/user`

**操作步骤：**
1. 在用户列表中找到目标用户
2. 点击 **"编辑"** 按钮
3. 在"用户角色"下拉框中选择角色（支持多选）
4. 点击确定保存

**效果：** 用户继承所选角色的所有权限

## 🎯 权限检查流程

### 用户登录后的权限检查：

```
用户登录
   ↓
获取用户角色列表
   ↓
合并所有角色的权限
   ↓
权限生效：
  - 菜单：只显示有权限的菜单项
  - 页面：只显示有权限的页面内容
  - 按钮：只显示有权限的操作按钮
  - API：后端拦截未授权的API请求
```

## 🔐 四层权限防护

### 1. 菜单层（导航）
- 前端路由守卫检查菜单权限
- 无权限的菜单不显示在侧边栏
- 直接访问URL也会被拦截

### 2. 页面层（页面访问）
- 页面组件加载时检查权限
- 无权限时显示403提示或重定向

### 3. 按钮层（操作控制）
- 使用 `v-permission` 指令控制按钮显示
- 使用 `PermissionButton` 组件实现多模式控制
- 使用 `hasPermission()` 函数进行逻辑判断

### 4. API层（接口保护）
- 后端中间件拦截未授权请求
- 返回403错误码
- 前端全局拦截器统一处理

## 💡 使用示例

### 前端权限控制示例

```vue
<template>
  <!-- 方式1：使用v-permission指令 -->
  <el-button v-permission="'system.user.create'" type="primary">
    新增用户
  </el-button>

  <!-- 方式2：使用PermissionButton组件 -->
  <PermissionButton
    permission="system.user.delete"
    mode="disabled"
    type="danger"
    @click="handleDelete">
    删除用户
  </PermissionButton>

  <!-- 方式3：使用hasPermission函数 -->
  <el-button
    v-if="hasPermission('system.user.update')"
    type="primary"
    @click="handleEdit">
    编辑用户
  </el-button>

  <!-- 方式4：组合使用 -->
  <el-button
    v-if="hasAnyPermission(['system.user.view', 'system.user.update'])"
    type="info"
    @click="handleDetail">
    查看详情
  </el-button>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission';

const { hasPermission, hasAnyPermission } = usePermission();
</script>
```

### 后端权限检查示例

```go
// 在路由中使用权限中间件
router.GET("/users",
  middleware.AuthMiddleware(),           // 认证中间件
  middleware.PermissionMiddleware("system.user.view"),  // 权限中间件
  userController.GetUsers
)

// 在控制器中检查权限
func (c *UserController) DeleteUser(ctx *gin.Context) {
    userID := ctx.GetUint("user_id")

    // 检查权限
    hasPermission := permissionService.HasPermission(userID, "system.user.delete")
    if !hasPermission {
        ctx.JSON(403, gin.H{"error": "权限不足"})
        return
    }

    // 执行删除操作
    // ...
}
```

## 📊 权限数据结构

### 数据库表结构

```sql
-- 权限表
CREATE TABLE permissions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    code VARCHAR(150) NOT NULL UNIQUE COMMENT '权限编码',
    module VARCHAR(50) NOT NULL COMMENT '所属模块',
    resource VARCHAR(50) NOT NULL COMMENT '资源名称',
    action VARCHAR(50) NOT NULL COMMENT '操作名称',
    level TINYINT NOT NULL COMMENT '权限级别',
    status TINYINT DEFAULT 1 COMMENT '状态'
);

-- 角色权限关联表
CREATE TABLE role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- 用户角色关联
-- 在users表中存储role_ids字段（JSON数组）
```

## 🎨 最佳实践

### 1. 权限命名规范
- 使用 **小写字母** 和 **下划线**
- 遵循 `module.resource.action` 格式
- 动作使用标准动词：view, create, update, delete, manage, export, import

### 2. 权限分配原则
- **最小权限原则**：只分配必要的权限
- **角色继承**：用户通过角色获得权限，不直接给用户分配权限
- **职责分离**：不同职责的用户分配不同角色

### 3. 特殊角色
- **超级管理员**：拥有所有权限（代码层面判断，不在数据库中存储）
- **审计员**：只读权限，不能修改数据
- **查看者**：只能查看，不能操作

## 🔧 系统内置权限

系统已内置以下模块的完整权限：

| 模块 | 页面级权限 | 按钮级权限 |
|------|-----------|-----------|
| 用户管理 | `system.user` | view, create, update, delete |
| 角色管理 | `system.role` | view, create, update, delete |
| 菜单管理 | `system.menu` | view, create, update, delete |
| 权限管理 | `system.permission` | view, create, update, delete |

## 🚀 快速开始

### 完整授权流程示例

1. **创建角色** "运维工程师"
   - 菜单权限：用户管理、角色管理（只读）
   - 按钮权限：查看用户、查看角色

2. **创建角色** "系统管理员"
   - 菜单权限：所有菜单
   - 按钮权限：所有按钮

3. **分配角色**
   - 给用户A分配 "运维工程师" 角色
   - 给用户B分配 "系统管理员" 角色

4. **效果验证**
   - 用户A登录：只能看到用户管理和角色管理菜单，且只有查看按钮
   - 用户B登录：可以看到所有菜单和所有操作按钮

---

**权限管理是系统安全的核心，请谨慎操作！**
