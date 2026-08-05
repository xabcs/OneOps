# 权限名称优化完成

## 修改内容

### 1. 修改 auth store
**文件：** `frontend/src/store/modules/auth/index.ts`

#### 新增内容：

**1.1 导入权限列表API**
```typescript
import { fetchGetPermissionList } from '@/service/api/system-manage';
```

**1.2 添加权限名称映射**
```typescript
const permissionNameMap = ref<Record<string, string>>({});
```

**1.3 添加获取权限名称方法**
```typescript
function getPermissionName(code: string): string {
  if (permissionNameMap.value[code]) {
    return permissionNameMap.value[code];
  }
  return code;
}
```

**1.4 添加加载权限映射方法**
```typescript
async function loadPermissionNameMap() {
  if (isSuperAdmin.value) return;
  
  try {
    const { data } = await fetchGetPermissionList({ page: 1, size: 1000 });
    if (data?.list) {
      const map: Record<string, string> = {};
      data.list.forEach((perm: any) => {
        map[perm.code] = perm.name;
      });
      permissionNameMap.value = map;
      console.log('[权限映射] 加载成功，共', Object.keys(map).length, '个权限');
    }
  } catch (error) {
    console.error('[权限映射] 加载失败:', error);
  }
}
```

**1.5 在登录成功后加载权限映射**
```typescript
async function loginByToken(loginToken: Api.Auth.LoginToken) {
  localStg.set('token', loginToken.token);
  const pass = await handleUserInfo(loginToken.user);
  
  if (pass) {
    token.value = loginToken.token;
    await loadPermissionNameMap();  // 新增
    return true;
  }
  return false;
}
```

**1.6 导出新方法**
```typescript
return {
  // ... 其他导出 ...
  getPermissionName
};
```

---

### 2. 修改权限检查 Composable
**文件：** `frontend/src/composables/useUnifiedPermission.ts`

#### 删除内容：

**2.1 删除写死的权限名称映射函数**
```typescript
// ❌ 已删除这个函数（约128-151行）
const getPermissionName = (permission: string): string => {
  const permissionNames: Record<string, string> = {
    'system.user.view': '查看用户',
    'system.user.create': '创建用户',
    // ... 更多映射
  }
  return permissionNames[permission] || permission
}
```

#### 修改内容：

**2.2 使用 authStore 的 getPermissionName**
```typescript
// 修改前
: `【${getPermissionName(permission)}】权限 (${permission})`

// 修改后
: `【${authStore.getPermissionName(permission)}】权限 (${permission})`
```

**2.3 从 return 中移除 getPermissionName**
```typescript
return {
  executeWithPermission,
  createPermissionExecutor,
  showPermissionAlert,
  checkBatchPermissions
  // ❌ 已移除 getPermissionName
}
```

---

## 优化效果

### 优化前
- ❌ 权限名称写死在代码中
- ❌ 每次新增权限都要修改前端代码
- ❌ 需要维护权限码 → 名称的映射表
- ❌ 容易出现映射不一致的问题

### 优化后
- ✅ 权限名称从数据库动态获取
- ✅ 新增权限无需修改前端代码
- ✅ 权限名称统一管理（在数据库中）
- ✅ 保证数据一致性

---

## 使用方式

### 前端开发
```typescript
// 在需要显示权限名称的地方
const authStore = useAuthStore()
const permName = authStore.getPermissionName('system.user.create')
// 返回：'创建用户'
```

### 权限提示
```typescript
// 权限不足时自动显示友好提示
执行操作 → 无权限 → 显示："您需要【创建用户】权限才能新增用户"
```

---

## 测试步骤

1. **启动后端服务**
2. **test 用户重新登录**
3. **查看控制台日志**：
   ```
   [权限映射] 加载成功，共 XX 个权限
   ```
4. **测试权限提示**：
   - 点击"新增用户"按钮
   - 应该看到："您需要【创建用户】权限才能新增用户"
   - 权限名称来自数据库，不是写死的

---

## 注意事项

1. **登录时额外请求一次**：登录后会调用 `fetchGetPermissionList` 获取所有权限
2. **超级管理员跳过**：admin 用户或拥有 `*.*:*` 权限的用户不会加载权限映射
3. **向后兼容**：如果权限映射中没有找到，会回退到显示权限码
4. **性能考虑**：只请求一次，缓存在内存中（页面刷新会重新加载）

---

## 完成状态

✅ **已全部完成**
- auth store 已修改
- useUnifiedPermission.ts 已修改
- 删除了所有写死的权限映射
- 权限名称从数据库动态获取
