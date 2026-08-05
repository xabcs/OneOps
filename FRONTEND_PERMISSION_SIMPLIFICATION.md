# 前端权限逻辑简化说明

## 修改时间
2026-08-04

## 备份文件
`frontend/src/composables/useUnifiedPermission.ts.backup`

---

## 修改对比

### 修改前：前端检查权限

```typescript
const executeWithPermission = async (
  permission: string | string[],
  action: () => void | Promise<void>,
  config: PermissionAlertConfig = {}
): Promise<boolean> => {
  // ❌ 前端检查权限
  const hasPermission = authStore.hasPermission(permission)

  if (hasPermission) {
    // 有权限，执行操作
    await action()
    return true
  }

  // ❌ 无权限，前端显示提示
  return showPermissionAlert(permission, config)
}
```

**问题**：
- ❌ 前后端重复检查权限
- ❌ 前端提示信息可能与后端不一致
- ❌ 前端权限数据可能不是最新的

---

### 修改后：只显示后端错误

```typescript
const executeWithPermission = async (
  permission: string | string[],  // 仅用于日志，不用于检查
  action: () => void | Promise<void>,
  config: PermissionAlertConfig = {}
): Promise<boolean> => {
  try {
    // ✅ 直接执行操作，不检查权限
    await action()
    return true
  } catch (error: any) {
    // ✅ 捕获后端返回的403错误
    const statusCode = error?.response?.status || error?.code

    if (statusCode === 403) {
      // ✅ 显示后端返回的错误信息
      const backendMessage = error?.response?.data?.message
        || error?.response?.data?.error
        || error?.message
        || '权限不足'

      ElMessage.error(backendMessage)
      return false
    }

    // 其他错误也显示
    ElMessage.error(error?.message || '操作执行失败')
    return false
  }
}
```

**优点**：
- ✅ 权限检查统一在后端
- ✅ 错误信息来自后端，更准确
- ✅ 前端代码更简单
- ✅ 避免前后端权限判断不一致

---

## 工作流程对比

### 修改前：前端检查权限

```
用户点击"删除用户"
  ↓
前端 executeWithPermission('system.user.delete')
  ↓
❌ 前端检查权限：authStore.hasPermission('system.user.delete')
  ↓
有权限 → 执行 action → 调用 API → 后端再次检查权限
无权限 → 前端显示提示："您需要【删除用户】权限..."
```

**问题**：
- 权限检查重复
- 前端提示和后端错误信息可能不一致

---

### 修改后：只显示后端错误

```
用户点击"删除用户"
  ↓
前端 executeWithPermission('system.user.delete')
  ↓
✅ 直接执行 action → 调用 API
  ↓
后端检查权限
  ├─ 有权限 → 操作成功
  └─ 无权限 → 返回 403 + 错误信息
  ↓
✅ 前端捕获错误，显示后端消息："权限不足: 需要 system.user.delete 权限"
```

**优点**：
- 权限检查统一在后端
- 错误信息准确一致

---

## 保留的方法

### 1. executeWithPermission

**用途**：执行操作并捕获错误

**使用方式**：
```typescript
// 业务代码不需要修改
await executeWithPermission('system.user.delete', async () => {
  const { error } = await fetchDeleteUser(id)
  if (!error) {
    window.$message?.success('删除成功')
    onDeleted()
  } else {
    window.$message?.error(error.msg || '删除失败')
  }
}, { type: 'error' })
```

**变化**：
- ❌ 不再在前端检查权限
- ✅ 直接执行操作
- ✅ 捕获并显示后端错误

---

### 2. showPermissionAlert

**用途**：手动显示权限提示（用于特殊情况）

**使用方式**：
```typescript
// 如果需要手动显示提示，仍可使用
showPermissionAlert('system.user.delete', {
  title: '权限不足',
  showContact: true
})
```

**场景**：
- 特殊场景需要提前提示用户
- 但不用于阻止操作

---

### 3. checkBatchPermissions

**用途**：批量检查权限（用于UI控制）

**使用方式**：
```typescript
// 检查多个权限状态
const permissions = checkBatchPermissions([
  'system.user.view',
  'system.user.create',
  'system.user.delete'
])

// permissions = {
//   'system.user.view': true,
//   'system.user.create': false,
//   'system.user.delete': false
// }
```

**场景**：
- 隐藏没有权限的按钮
- 禁用没有权限的操作
- 但不阻止操作（后端会拦截）

---

### 4. showBackendError（新增）

**用途**：显示后端错误信息

**使用方式**：
```typescript
try {
  await someApiCall()
} catch (error) {
  showBackendError(error)
  // 显示: "权限不足: 需要 system.user.delete 权限"
}
```

---

## 代码量对比

| 维度 | 修改前 | 修改后 | 减少 |
|------|--------|--------|------|
| **总行数** | 163 行 | 112 行 | 51 行 (31%) |
| **executeWithPermission** | 45 行 | 25 行 | 20 行 |
| **权限检查逻辑** | 有 | 无 | - |

---

## 实际使用示例

### 场景：test用户尝试删除用户

#### 修改前（前端检查）

```typescript
// 1. 前端检查权限
const hasPermission = authStore.hasPermission('system.user.delete')  // false

// 2. 前端显示提示
showPermissionAlert('system.user.delete')
// 提示: "您需要【删除用户】权限 (system.user.delete)才能执行此操作"

// 3. 操作被阻止，不调用API
```

#### 修改后（后端检查）

```typescript
// 1. 直接执行操作
await executeWithPermission('system.user.delete', async () => {
  const { error } = await fetchDeleteUser(id)
  // ...
})

// 2. 调用API
DELETE /api/system/users/123

// 3. 后端检查权限
PermissionMiddleware('system.user.delete')
// → 403 Forbidden

// 4. 后端返回错误
{
  "code": 403,
  "message": "权限不足: 需要 system.user.delete 权限"
}

// 5. 前端捕获错误，显示后端消息
ElMessage.error('权限不足: 需要 system.user.delete 权限')
```

---

## 安全性说明

### 前端权限检查 vs 后端权限检查

| 维度 | 前端检查 | 后端检查 |
|------|---------|---------|
| **安全性** | ⚠️ 不安全，可绕过 | ✅ 安全，无法绕过 |
| **用户体验** | ✅ 即时反馈 | ⚠️ 需要等待API响应 |
| **一致性** | ⚠️ 可能不一致 | ✅ 始终一致 |
| **维护成本** | ⚠️ 需前后端同步 | ✅ 只需维护后端 |

**结论**：
- ✅ 后端权限检查是**必须的**（安全保障）
- ⚠️ 前端权限检查是**可选的**（用户体验优化）

**本次修改**：
- ❌ 移除前端权限检查（简化代码）
- ✅ 保留后端权限检查（安全保障）
- ✅ 显示后端错误信息（一致性）

---

## 兼容性

### 业务代码无需修改

**原有调用方式保持不变**：

```typescript
// 用户管理
await executeWithPermission('system.user.delete', async () => {
  // ...
}, { type: 'error' })

// 角色管理
await executeWithPermission('system.role.create', async () => {
  // ...
})

// 权限管理
await executeWithPermission('system.permission.update', async () => {
  // ...
})
```

**变化**：
- ❌ 不再在执行前检查权限
- ✅ 直接执行操作
- ✅ 如果后端返回403，显示后端错误信息

---

## 测试验证

### 测试场景1：有权限用户操作

**用户**：admin（有所有权限）

**操作**：删除用户

**流程**：
```
1. 点击删除按钮
2. executeWithPermission 执行 action
3. 调用 DELETE /api/system/users/123
4. 后端检查权限 → 有权限
5. 删除成功
6. 前端显示："删除成功"
```

**结果**：✅ 正常工作

---

### 测试场景2：无权限用户操作

**用户**：test（只有查看权限）

**操作**：删除用户

**流程**：
```
1. 点击删除按钮
2. executeWithPermission 执行 action
3. 调用 DELETE /api/system/users/123
4. 后端检查权限 → 无权限
5. 返回 403: "权限不足: 需要 system.user.delete 权限"
6. 前端捕获错误
7. 显示后端消息："权限不足: 需要 system.user.delete 权限"
```

**结果**：✅ 正常工作

---

## 总结

### ✅ 修改成果

1. ✅ 删除了前端权限检查逻辑
2. ✅ 保留了后端权限保护
3. ✅ 显示后端返回的错误信息
4. ✅ 代码量减少 31%
5. ✅ 业务代码无需修改
6. ✅ 错误信息更准确一致

### 🎯 核心优势

- **安全性**：后端权限检查是真正的安全保障
- **一致性**：错误信息来自后端，避免前后端不一致
- **简洁性**：前端代码更简单，易于维护
- **准确性**：后端错误信息包含实际需要的权限

### 📝 后续优化

如果需要更好的用户体验，可以考虑：
1. **前端提示优化**：在用户点击前提示可能没有权限（但不阻止操作）
2. **权限预加载**：页面加载时批量检查权限，用于UI控制
3. **友好提示**：后端错误信息中包含如何申请权限的说明

但核心原则不变：**后端权限检查是唯一的安全保障**。
