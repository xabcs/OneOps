# 用户管理API权限保护清单

## 📊 当前状态分析

### ✅ 已实现并已保护的API（5个）

| HTTP方法 | 路径 | Controller方法 | 权限码 | 状态 |
|---------|------|---------------|--------|------|
| GET | `/api/system/users` | `GetUsers` | `system.user.view` | ✅ 已保护 |
| POST | `/api/system/users` | `CreateUser` | `system.user.create` | ✅ 已保护 |
| PUT | `/api/system/users/:id` | `UpdateUser` | `system.user.update` | ✅ 已保护 |
| DELETE | `/api/system/users/:id` | `DeleteUser` | `system.user.delete` | ✅ 已保护 |
| PUT | `/api/system/users/:id/password` | `ResetPassword` | `system.user.reset_password` | ✅ 已保护 |

---

### ❌ 已定义权限但未实现的API（17个）

#### 1. 批量操作类（2个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.batch_delete` | 批量删除用户 | `DELETE /api/system/users/batch` | 批量删除多个用户 |
| `system.user.copy` | 复制用户 | `POST /api/system/users/:id/copy` | 复制用户信息创建新用户 |

**实现优先级**：⭐⭐⭐⭐⭐ 高

---

#### 2. 用户状态管理类（4个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.lock` | 锁定用户 | `PUT /api/system/users/:id/lock` | 锁定用户账号 |
| `system.user.unlock` | 解锁用户 | `PUT /api/system/users/:id/unlock` | 解锁被锁定的用户 |
| `system.user.enable` | 启用用户 | `PUT /api/system/users/:id/enable` | 启用用户账号 |
| `system.user.disable` | 禁用用户 | `PUT /api/system/users/:id/disable` | 禁用用户账号 |

**实现优先级**：⭐⭐⭐⭐ 高

---

#### 3. 角色管理类（3个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.view_roles` | 查看用户角色 | `GET /api/system/users/:id/roles` | 查看用户的角色列表 |
| `system.user.assign_role` | 分配角色 | `POST /api/system/users/:id/roles` | 为用户分配角色 |
| `system.user.remove_role` | 移除角色 | `DELETE /api/system/users/:id/roles/:roleId` | 移除用户角色 |

**实现优先级**：⭐⭐⭐⭐⭐ 高

---

#### 4. 数据导入导出类（3个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.export` | 导出用户 | `GET /api/system/users/export` | 导出用户数据 |
| `system.user.import` | 导入用户 | `POST /api/system/users/import` | 导入用户数据 |
| `system.user.download_template` | 下载导入模板 | `GET /api/system/users/import/template` | 下载用户导入模板 |

**实现优先级**：⭐⭐⭐ 中

---

#### 5. 用户信息管理类（2个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.update_profile` | 修改个人信息 | `PUT /api/system/users/:id/profile` | 用户修改个人资料 |
| `system.user.update_avatar` | 修改头像 | `PUT /api/system/users/:id/avatar` | 修改用户头像 |

**实现优先级**：⭐⭐ 低（通常只允许修改自己的信息）

---

#### 6. 高级功能类（3个）

| 权限码 | 权限名称 | 建议API | 说明 |
|--------|---------|---------|------|
| `system.user.view_history` | 查看操作历史 | `GET /api/system/users/:id/history` | 查看用户操作历史 |
| `system.user.approve` | 审批用户 | `POST /api/system/users/:id/approve` | 审批用户注册或变更 |
| `system.user.reject` | 拒绝用户 | `POST /api/system/users/:id/reject` | 拒绝用户注册或变更 |

**实现优先级**：⭐⭐ 低（特殊场景才需要）

---

## 🎯 推荐的实施优先级

### 第一优先级：核心功能（立即实施）⭐⭐⭐⭐⭐

这些是用户管理最常用的功能，**必须实现**：

#### 1. 批量删除用户

**API**: `DELETE /api/system/users/batch`

**权限**: `system.user.batch_delete`

**Controller方法**:
```go
func (ctrl *UserController) BatchDeleteUsers(c *gin.Context) {
    var req struct {
        IDs []uint `json:"ids" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, utils.ErrorBadRequest("参数错误"))
        return
    }

    // 批量删除逻辑
    // ...
}
```

**路由注册**:
```go
system.DELETE("/users/batch",
    middlewares.PermissionMiddleware("system.user.batch_delete"),
    userController.BatchDeleteUsers)
```

---

#### 2. 角色管理功能

**API**:
- `GET /api/system/users/:id/roles` - 查看用户角色
- `POST /api/system/users/:id/roles` - 分配角色
- `DELETE /api/system/users/:id/roles/:roleId` - 移除角色

**权限**:
- `system.user.view_roles`
- `system.user.assign_role`
- `system.user.remove_role`

**Controller方法**:
```go
func (ctrl *UserController) GetUserRoles(c *gin.Context) {
    // 获取用户角色列表
}

func (ctrl *UserController) AssignRole(c *gin.Context) {
    // 为用户分配角色
}

func (ctrl *UserController) RemoveRole(c *gin.Context) {
    // 移除用户角色
}
```

---

#### 3. 用户状态管理

**API**:
- `PUT /api/system/users/:id/lock` - 锁定用户
- `PUT /api/system/users/:id/unlock` - 解锁用户
- `PUT /api/system/users/:id/enable` - 启用用户
- `PUT /api/system/users/:id/disable` - 禁用用户

**权限**:
- `system.user.lock`
- `system.user.unlock`
- `system.user.enable`
- `system.user.disable`

---

### 第二优先级：常用功能（近期实施）⭐⭐⭐

这些功能提升用户体验，建议实现：

#### 1. 数据导入导出

**API**:
- `GET /api/system/users/export` - 导出用户
- `POST /api/system/users/import` - 导入用户
- `GET /api/system/users/import/template` - 下载导入模板

**权限**:
- `system.user.export`
- `system.user.import`
- `system.user.download_template`

---

### 第三优先级：增强功能（可选实施）⭐⭐

这些功能用于特殊场景，可根据需求实现：

#### 1. 用户信息管理

**API**:
- `PUT /api/system/users/:id/profile` - 修改个人信息
- `PUT /api/system/users/:id/avatar` - 修改头像

**权限**:
- `system.user.update_profile`
- `system.user.update_avatar`

**注意**：通常只允许用户修改自己的信息，需要额外权限检查。

---

#### 2. 高级功能

**API**:
- `GET /api/system/users/:id/history` - 查看操作历史
- `POST /api/system/users/:id/approve` - 审批用户
- `POST /api/system/users/:id/reject` - 拒绝用户
- `POST /api/system/users/:id/copy` - 复制用户

**权限**:
- `system.user.view_history`
- `system.user.approve`
- `system.user.reject`
- `system.user.copy`

---

## 📋 完整API清单（按优先级）

### 第一优先级（10个API）

| 序号 | HTTP方法 | 路径 | 权限码 | 说明 |
|-----|---------|------|--------|------|
| 1 | DELETE | `/users/batch` | `system.user.batch_delete` | 批量删除 |
| 2 | GET | `/users/:id/roles` | `system.user.view_roles` | 查看用户角色 |
| 3 | POST | `/users/:id/roles` | `system.user.assign_role` | 分配角色 |
| 4 | DELETE | `/users/:id/roles/:roleId` | `system.user.remove_role` | 移除角色 |
| 5 | PUT | `/users/:id/lock` | `system.user.lock` | 锁定用户 |
| 6 | PUT | `/users/:id/unlock` | `system.user.unlock` | 解锁用户 |
| 7 | PUT | `/users/:id/enable` | `system.user.enable` | 启用用户 |
| 8 | PUT | `/users/:id/disable` | `system.user.disable` | 禁用用户 |
| 9 | GET | `/users/:id` | `system.user.view` | 查看用户详情 |
| 10 | GET | `/users` | `system.user.view` | 查看用户列表 |

**状态**：
- ✅ 1-2 已实现（基础CRUD）
- ❌ 3-10 未实现

---

### 第二优先级（3个API）

| 序号 | HTTP方法 | 路径 | 权限码 | 说明 |
|-----|---------|------|--------|------|
| 11 | GET | `/users/export` | `system.user.export` | 导出用户 |
| 12 | POST | `/users/import` | `system.user.import` | 导入用户 |
| 13 | GET | `/users/import/template` | `system.user.download_template` | 下载导入模板 |

**状态**：❌ 未实现

---

### 第三优先级（7个API）

| 序号 | HTTP方法 | 路径 | 权限码 | 说明 |
|-----|---------|------|--------|------|
| 14 | PUT | `/users/:id/profile` | `system.user.update_profile` | 修改个人信息 |
| 15 | PUT | `/users/:id/avatar` | `system.user.update_avatar` | 修改头像 |
| 16 | GET | `/users/:id/history` | `system.user.view_history` | 查看操作历史 |
| 17 | POST | `/users/:id/approve` | `system.user.approve` | 审批用户 |
| 18 | POST | `/users/:id/reject` | `system.user.reject` | 拒绝用户 |
| 19 | POST | `/users/:id/copy` | `system.user.copy` | 复制用户 |
| 20 | GET | `/users/:id/roles` | `system.user.view_roles` | 查看用户角色（已列） |

**状态**：❌ 未实现

---

## 🛡️ 权限保护原则

### Casbin管理的范围

**应该被Casbin管理的API**：
1. ✅ 所有涉及数据修改的API（POST、PUT、DELETE）
2. ✅ 所有涉及敏感数据查看的API（GET）
3. ✅ 所有批量操作API
4. ✅ 所有导入导出API
5. ✅ 所有状态管理API

**不应该被Casbin管理的API**：
1. ❌ 公开API（如登录、注册）
2. ❌ 用户自己的信息查询（如获取当前用户信息）
3. ❌ 健康检查API
4. ❌ 静态资源API

---

## 📝 实施建议

### 阶段1：核心功能（立即）

**目标**：实现最常用的用户管理功能

**API数量**：10个

**预计工时**：2-3天

**关键API**：
- ✅ 批量删除
- ✅ 角色管理（查看、分配、移除）
- ✅ 状态管理（锁定、解锁、启用、禁用）
- ✅ 用户详情查询

---

### 阶段2：增强功能（近期）

**目标**：提升用户体验

**API数量**：3个

**预计工时**：1-2天

**关键API**：
- ✅ 导入导出
- ✅ 下载导入模板

---

### 阶段3：高级功能（可选）

**目标**：满足特殊场景需求

**API数量**：7个

**预计工时**：2-3天

**关键API**：
- ✅ 用户信息修改
- ✅ 操作历史查看
- ✅ 审批流程
- ✅ 用户复制

---

## ✅ 实施检查清单

### 每个API实施时需要：

- [ ] **Controller方法实现**
- [ ] **路由注册**
- [ ] **权限中间件配置**
- [ ] **权限数据库记录**
- [ ] **前端调用代码**
- [ ] **API文档更新**
- [ ] **测试用例编写**

---

## 📊 总结

| 统计项 | 数量 |
|--------|------|
| **已实现API** | 5个 |
| **待实现API** | 15个 |
| **总计API** | 20个 |
| **第一优先级** | 10个 |
| **第二优先级** | 3个 |
| **第三优先级** | 7个 |

**当前覆盖率**：25%（5/20）

**建议**：优先实现第一优先级的10个核心API，使覆盖率达到50%。
