# executeWithPermission 方案实施完成

## 改进总结

### 删除的代码

**useUnifiedPermission.ts**：
- ❌ 删除了 131 行预定义执行器代码
- ❌ 删除了 `useUserActionExecutors()`
- ❌ 删除了 `useRoleActionExecutors()`
- ❌ 删除了 `usePermissionActionExecutors()`
- 文件从 294 行精简到 163 行

### 修改的文件

**业务页面（4个）**：
1. ✅ `views/manage/user/index.vue` - 用户管理
2. ✅ `views/manage/role/index.vue` - 角色管理
3. ✅ `views/manage/permission/index.vue` - 权限管理
4. ✅ `views/manage/permission/index_new.vue` - 权限管理新版本

---

## 修改对比

### 修改前（预定义执行器方案）

```typescript
// 1. composable 中预定义执行器（硬编码）
export const useUserActionExecutors = () => {
  return {
    executeCreate: createPermissionExecutor('system.user.create'),
    executeDelete: createPermissionExecutor('system.user.delete')
  }
}

// 2. 业务代码中使用预定义执行器
const { executeCreate, executeDelete } = useUserActionExecutors()

await executeCreate(async () => {
  handleAdd()
})
```

**问题**：
- ❌ 权限码硬编码在 composable 中
- ❌ 执行器名称硬编码
- ❌ 新增权限需要修改 composable

---

### 修改后（executeWithPermission 方案）

```typescript
// 1. composable 只提供核心方法（无硬编码）
export const useUnifiedPermission = () => {
  const executeWithPermission = async (
    permission: string | string[],
    action: () => void | Promise<void>,
    config: PermissionAlertConfig = {}
  ): Promise<boolean> => {
    // 检查权限并执行操作
  }

  return { executeWithPermission, showPermissionAlert }
}

// 2. 业务代码中动态使用
const { executeWithPermission } = useUnifiedPermission()

await executeWithPermission('system.user.create', async () => {
  handleAdd()
})

await executeWithPermission('system.user.delete', async () => {
  await fetchDeleteUser(id)
}, { type: 'error' })
```

**优点**：
- ✅ 权限码在业务代码中定义，清晰明了
- ✅ 配置灵活，每个操作可以不同
- ✅ 新增权限不需要修改 composable
- ✅ 完全无硬编码

---

## 实际使用示例

### 用户管理页面

```typescript
import { useUnifiedPermission } from '@/composables/useUnifiedPermission'

const { executeWithPermission } = useUnifiedPermission()

// 新增用户
async function handleAddClick() {
  await executeWithPermission('system.user.create', async () => {
    handleAdd()
  })
}

// 删除用户
async function handleDelete(id: number) {
  await executeWithPermission('system.user.delete', async () => {
    const { error } = await fetchDeleteUser(id)
    if (!error) {
      window.$message?.success($t('common.deleteSuccess'))
      onDeleted()
    }
  }, { type: 'error' })
}

// 重置密码
async function openResetPassword(row: Api.SystemManage.User) {
  await executeWithPermission('system.user.reset_password', async () => {
    resetPasswordUserId.value = row.id
    resetPasswordVisible.value = true
  })
}

// 批量删除
async function handleBatchDelete() {
  await executeWithPermission('system.user.batch_delete', async () => {
    // 批量删除逻辑...
  })
}
```

---

## 核心方法 API

### executeWithPermission

```typescript
/**
 * 执行权限检查和操作
 * @param permission 权限编码（单个或多个）
 * @param action 要执行的操作
 * @param config 提醒配置（可选）
 * @returns Promise<boolean> 是否成功执行
 */

// 单个权限
await executeWithPermission('system.user.create', async () => {
  // 操作逻辑
})

// 多个权限（满足任意一个即可）
await executeWithPermission(['system.user.view', 'system.user.update'], async () => {
  // 操作逻辑
})

// 带配置
await executeWithPermission('system.user.delete', async () => {
  // 操作逻辑
}, {
  type: 'error',          // 提醒类型
  useNotification: true   // 使用通知而非弹窗
})
```

### showPermissionAlert

```typescript
/**
 * 显示权限提醒
 * @param permission 权限编码
 * @param config 提醒配置
 */

// 使用示例
showPermissionAlert('system.user.create', {
  title: '权限不足',
  showContact: true,
  contactInfo: 'admin@company.com'
})
```

---

## 权限提示效果

**动态生成的提示消息**：

```
权限不足

您需要【创建用户】权限 (system.user.create)才能执行此操作

如需使用此功能，请联系管理员申请权限
📧 联系方式：admin@company.com
```

**关键特性**：
- ✅ 权限名称来自数据库（`authStore.getPermissionName()`）
- ✅ 显示权限码，方便沟通
- ✅ 包含联系方式
- ✅ 完全动态，无硬编码

---

## 方案优势

| 维度 | 改进前 | 改进后 |
|------|--------|--------|
| **代码量** | 294 行 | 163 行（减少 45%） |
| **维护性** | 新增权限需修改 composable | 新增权限只需在业务代码添加 |
| **灵活性** | 配置固定 | 每个操作可自定义配置 |
| **可读性** | 执行器名称暗示权限 | 权限码直接可见 |
| **硬编码** | ❌ 权限码、执行器名都硬编码 | ✅ 只有业务代码中有权限码 |
| **类比** | 硬编码的路由表 | 动态路由配置 |

---

## 文件结构

```
frontend/src/
├── composables/
│   └── useUnifiedPermission.ts  (163行，核心方法)
└── views/manage/
    ├── user/index.vue           (使用 executeWithPermission)
    ├── role/index.vue           (使用 executeWithPermission)
    └── permission/
        ├── index.vue            (使用 executeWithPermission)
        └── index_new.vue        (使用 executeWithPermission)
```

---

## 总结

**executeWithPermission 方案的本质**：
- 权限检查是业务逻辑的一部分，应该在业务代码中定义
- composable 只提供工具方法，不预设业务规则
- 类似于路由的 `meta.permission` 定义在路由配置中

**核心优势**：
- ✅ 完全消除硬编码
- ✅ 代码更简洁（减少 45%）
- ✅ 更灵活、更易维护
- ✅ 权限码在业务代码中可见，便于理解和调试
