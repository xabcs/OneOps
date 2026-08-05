# 前后端权限检查架构设计

## 权限检查的两层架构

### 1. 前端权限检查（按钮级别）
**目的：** 用户体验优化，不是安全控制

**数据来源：**
```typescript
// 用户登录时从后端API获取权限列表
const authStore = useAuthStore()

// 登录响应中包含权限信息
interface LoginResponse {
  token: string
  user: UserInfo
  permissions: string[]  // 权限列表
  roles: string[]
}

// 权限检查基于本地权限列表
function hasPermission(permission: string): boolean {
  return authStore.permissions.includes(permission)
}
```

**不连接Casbin的原因：**
- 前端不应该直接访问业务逻辑层
- Casbin是后端的权限引擎，不应该被前端直接调用
- 前端只需要知道"有/无权限"，不需要知道如何验证
- 安全性：前端权限检查可以被绕过，所以它不是安全措施

### 2. 后端权限检查（API级别）
**目的：** 真正的安全控制，防止恶意API调用

**数据来源：**
```go
// 连接到Casbin进行权限验证
func (s *PermissionService) HasPermission(userID uint, permissionCode string) (bool, error) {
    // 1. 获取用户角色
    roles, err := s.getUserRoles(userID)

    // 2. 使用Casbin检查权限
    for _, role := range roles {
        allowed, _ := s.enforcer.Enforce(role.Code, permissionCode, "*")
        if allowed {
            return true, nil
        }
    }
    return false, nil
}
```

**连接Casbin的原因：**
- 后端拥有权限引擎的访问权限
- Casbin提供权威的权限判断
- 这是防止恶意调用的最后一道防线
- 可以根据业务规则灵活配置权限策略

## 正确的权限检查流程

### 完整流程：

```
1. 用户登录
   ↓
2. 后端验证身份 + 查询权限 + Casbin验证
   ↓
3. 返回用户信息 + 权限列表给前端
   ↓
4. 前端存储权限列表到本地状态
   ↓
5. 用户操作时前端检查本地权限（优化体验）
   ↓
6. 发送API请求到后端
   ↓
7. 后端再次连接Casbin验证权限（安全控制）
   ↓
8. 执行/拒绝操作
```

### 前端权限检查：
```typescript
// 快速检查，优化用户体验
const handleDelete = () => {
  // 前端检查（体验优化）
  if (!hasPermission('system.user.delete')) {
    showPermissionAlert()  // 显示权限提醒
    return
  }

  // 发送API请求
  deleteUserAPI(id)
}
```

### 后端权限检查：
```go
// 每个API都检查权限（真正的安全控制）
permGroup.DELETE("/:id",
    PermissionMiddleware("system.user.delete"),  // 连接Casbin验证
    permController.DeleteUser)
```

## 为什么这样设计？

### 1. 安全性考虑
**前端权限检查 ≠ 安全控制**
- 前端权限检查只是UI优化
- 用户可以打开开发者工具，直接调用API
- 必须在后端进行真正的权限验证

### 2. 性能考虑
**前端本地检查 = 快速响应**
- 检查本地数组：O(n)复杂度
- 无需网络请求，即时响应
- 提供流畅的用户体验

**后端Casbin检查 = 权威验证**
- 查询权限引擎，可能涉及复杂规则
- 可以根据业务需求灵活配置
- 提供最终的安全保障

### 3. 职责分离
**前端负责：用户体验**
- 权限按钮显示/隐藏
- 操作前权限提醒
- 界面交互优化

**后端负责：安全控制**
- API访问权限验证
- Casbin权限策略执行
- 防止恶意攻击

## 实际实现示例

### 前端（不直接连接Casbin）：
```typescript
// 权限来自登录API
const { permissions } = authStore

// 本地检查权限
function hasPermission(permission: string): boolean {
  return permissions.includes(permission) || permissions.includes('*:*:*')
}

// 使用权限检查
const handleDelete = () => {
  if (!hasPermission('system.user.delete')) {
    showNoPermissionAlert()
    return
  }
  // 发送API请求
}
```

### 后端（连接Casbin）：
```go
// 权限中间件连接Casbin
func PermissionMiddleware(permissionCode string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取权限服务（已连接Casbin）
        permService := container.PermissionService()

        // Casbin权限验证
        hasPermission, _ := permService.HasPermission(userID, permissionCode)

        if !hasPermission {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 总结

**按钮级别权限检查：**
- ✅ 前端：基于本地权限列表检查（优化体验）
- ❌ 不直接连接Casbin（不需要也不应该）

**API级别权限检查：**
- ✅ 后端：连接Casbin进行权限验证（真正的安全控制）
- ✅ 防止恶意API调用（最终安全防线）

**前后端配合：**
- 前端：优化用户体验，快速响应
- 后端：真正安全验证，防止攻击
- 分层架构：体验+安全

这样的设计既保证了安全性，又提供了良好的用户体验！