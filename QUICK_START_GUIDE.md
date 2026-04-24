# 代码简化 - 快速使用指南

## 🚀 新工具快速上手

### 1. 存储管理 (`utils/storage.js`)

#### 认证存储
```javascript
import { authStorage } from '@/utils/storage'

// 保存认证信息
authStorage.saveAuth(token, user, permissions, menuTree)

// 清除认证信息
authStorage.clearAuth()

// 获取认证数据
const { token, user, permissions } = authStorage.getAuth()

// 单独获取
const token = authStorage.getToken()
const user = authStorage.getUser()
```

#### 后端状态存储
```javascript
import { backendStatusStorage } from '@/utils/storage'

// 设置后端不可用
backendStatusStorage.setUnavailable()

// 清除不可用状态
backendStatusStorage.clearUnavailable()

// 检查状态
if (backendStatusStorage.isUnavailable()) {
  // 处理不可用情况
}
```

### 2. Composables

#### useAuth - 认证管理
```javascript
import { useAuth } from '@/composables/useAuth'

export default {
  setup() {
    const {
      isAuthenticated,  // 是否已登录
      user,           // 用户信息
      permissions,    // 权限列表
      login,          // 登录方法
      logout,         // 登出方法
      fetchUserInfo,  // 获取用户信息
      hasPermission   // 权限检查
    } = useAuth()

    // 登录示例
    const handleLogin = async () => {
      const result = await login({
        username: 'admin',
        password: '123456'
      })

      if (result.success) {
        console.log('登录成功', result.user)
      }
    }

    // 权限检查示例
    const canEdit = hasPermission('menu:system:users:edit')

    return { isAuthenticated, user, login, logout, canEdit }
  }
}
```

#### useBackendStatus - 后端状态管理
```javascript
import { useBackendStatus } from '@/composables/useBackendStatus'

export default {
  setup() {
    const {
      isBackendAvailable,    // 后端是否可用
      setBackendAvailable,   // 设置为可用
      setBackendUnavailable, // 设置为不可用
      retryConnection        // 重试连接
    } = useBackendStatus()

    // 重试连接示例
    const handleRetry = async () => {
      const success = await retryConnection()
      if (success) {
        console.log('连接成功')
      }
    }

    return { isBackendAvailable, handleRetry }
  }
}
```

#### useTable - 表格数据管理
```javascript
import { useTable } from '@/composables/useTable'
import { systemApi } from '@/api'

export default {
  setup() {
    const {
      loading,      // 加载状态
      data,         // 数据列表
      total,        // 总数
      queryParams,  // 查询参数
      currentPage,  // 当前页
      fetchData,    // 获取数据
      search,       // 搜索
      reset         // 重置
    } = useTable(systemApi.getUsers, {
      pageSize: 20,
      defaultQuery: { status: 'active' }
    })

    return {
      loading,
      users: data,
      total,
      currentPage,
      fetchData,
      search
    }
  }
}
```

### 3. 常量使用 (`constants/index.js`)

```javascript
import { STORAGE_KEYS, ROUTES, BACKEND_STATUS } from '@/constants'

// 存储键名
localStorage.setItem(STORAGE_KEYS.TOKEN, token)
localStorage.getItem(STORAGE_KEYS.USER)

// 路由路径
router.push(ROUTES.LOGIN)
router.push(ROUTES.HOME)

// 配置值
const checker = new HealthChecker({
  interval: BACKEND_STATUS.CHECK_INTERVAL,
  timeout: BACKEND_STATUS.API_TIMEOUT
})
```

### 4. 可重用组件

#### PageCard - 页面卡片
```vue
<template>
  <PageCard
    title="用户管理"
    subtitle="管理系统用户及权限"
    :loading="loading"
  >
    <template #actions>
      <el-input v-model="search" placeholder="搜索..." />
      <el-button type="primary">新增</el-button>
    </template>

    <!-- 表格内容 -->
    <el-table :data="users">
      <!-- ... -->
    </el-table>
  </PageCard>
</template>

<script setup>
import { ref } from 'vue'
import PageCard from '@/components/PageCard.vue'

const loading = ref(false)
const search = ref('')
const users = ref([])
</script>
```

## 📝 迁移现有代码

### 从旧代码迁移到新工具

#### 场景1: 替换重复的 localStorage 操作

**旧代码**:
```javascript
localStorage.removeItem('token')
localStorage.removeItem('user')
localStorage.removeItem('permissions')
localStorage.removeItem('menuTree')
```

**新代码**:
```javascript
import { authStorage } from '@/utils/storage'
authStorage.clearAuth()
```

#### 场景2: 替换复杂的登录逻辑

**旧代码**:
```javascript
const handleLogin = async () => {
  const response = await loginApi.login(form)
  if (response.success) {
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('user', JSON.stringify(response.data.user))
    // ... 更多存储操作
    store.commit('SET_TOKEN', response.data.token)
    // ... 更多 mutations
    ElMessage.success('欢迎回来')
    router.push('/')
  }
}
```

**新代码**:
```javascript
import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()

const handleLogin = async () => {
  const result = await login(form)
  if (result.success) {
    window.location.href = result.user?.homePath || '/'
  }
}
```

#### 场景3: 替换手动检查后端状态

**旧代码**:
```javascript
const isUnavailable = !!sessionStorage.getItem('backendUnavailable')
const unavailableTime = sessionStorage.getItem('backendUnavailableTime')
```

**新代码**:
```javascript
import { backendStatusStorage } from '@/utils/storage'

const isUnavailable = backendStatusStorage.isUnavailable()
const unavailableTime = backendStatusStorage.getUnavailableTime()
```

## 🎯 最佳实践

### 1. 优先使用 Composables
```javascript
// ✅ 推荐
import { useAuth } from '@/composables/useAuth'
const { user, hasPermission } = useAuth()

// ❌ 不推荐
import { useStore } from 'vuex'
const store = useStore()
const user = computed(() => store.state.user)
```

### 2. 使用统一的存储接口
```javascript
// ✅ 推荐
import { authStorage } from '@/utils/storage'
authStorage.clearAuth()

// ❌ 不推荐
localStorage.removeItem('token')
localStorage.removeItem('user')
// ... 分散的操作
```

### 3. 使用常量避免魔法字符串
```javascript
// ✅ 推荐
import { ROUTES } from '@/constants'
router.push(ROUTES.LOGIN)

// ❌ 不推荐
router.push('/login')
```

### 4. 利用 useTable 简化表格逻辑
```javascript
// ✅ 推荐
import { useTable } from '@/composables/useTable'
const { data, loading, fetchData } = useTable(api.getData)

// ❌ 不推荐
const data = ref([])
const loading = ref(false)
const fetchData = async () => {
  loading.value = true
  // ... 手动实现
}
```

## 🐛 常见问题

### Q: 如何在现有组件中逐步采用新工具？
A: 可以逐步替换。先从存储操作开始，然后是认证逻辑，最后是表格管理。

### Q: 新工具是否向后兼容？
A: 是的，所有新工具都保持向后兼容。你可以混合使用新旧代码。

### Q: 性能是否受影响？
A: 不会。新工具实际上提升了性能，减少了重复操作。

### Q: 如何调试新工具？
A: 所有工具都有清晰的错误处理和日志输出。可以在控制台查看详细日志。

## 📚 相关资源

- **完整报告**: `CODE_SIMPLIFICATION_REPORT.md`
- **API文档**: 查看各文件的 JSDoc 注释
- **示例代码**: `frontend/src/composables/` 中的示例

## 💡 提示

1. **从简单开始**: 先使用存储工具，再使用 composables
2. **渐进式迁移**: 不需要一次性重写所有代码
3. **保持一致**: 团队统一使用新工具和模式
4. **分享经验**: 将好的实践分享给团队成员

祝编码愉快！🎉