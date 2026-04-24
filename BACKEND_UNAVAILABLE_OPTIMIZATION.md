# 后端服务不可用处理 - 优化版本

## 核心改进

现在的实现严格遵循"只显示登录页面"的原则：
- ✅ 后端不可用时，完全隐藏主界面布局（侧边栏、顶部菜单等）
- ✅ 只显示纯净的登录页面
- ✅ 登录时智能检测后端状态并给出相应提示
- ✅ 主界面不显示后端状态监控（避免混淆）

## 行为变化

### 🔴 后端服务未启动时

**访问任何页面**：
1. 自动重定向到 `/login`
2. 显示纯净登录页面（无侧边栏、无顶部菜单）
3. 显示友好的后端不可用警告信息
4. 提供详细的检查清单

**尝试登录时**：
- 显示"正在尝试重新连接..."的提示
- 如果连接失败，显示详细的错误信息和检查步骤
- 如果连接成功，正常进入系统

### 🟢 后端服务正常时

**登录页面**：
- 不显示后端状态警告
- 提供正常的登录表单
- 登录成功后进入主界面

**主界面**：
- 显示完整的布局（侧边栏、顶部菜单、标签页等）
- 不显示后端状态监控组件
- 正常的功能访问

## 代码变更说明

### 1. App.vue - 布局控制

```javascript
// 新增计算属性：决定是否显示主界面布局
const shouldShowMainLayout = computed(() => {
    return isAuthenticated.value && isBackendAvailable.value && route.path !== '/login'
})
```

**模板更新**：
```vue
<!-- 之前：只要认证就显示主界面 -->
<div v-if="isAuthenticated" class="app-wrapper">

<!-- 现在：认证且后端可用才显示主界面 -->
<div v-if="shouldShowMainLayout" class="app-wrapper">
```

### 2. router/index.js - 严格路由守卫

**关键变更**：
```javascript
// 之前：只有访问需要认证的页面时才检查
if (isBackendUnavailable && requiresAuth) {

// 现在：任何非登录页面都会被重定向
if (isBackendUnavailable && to.path !== '/login') {
```

### 3. Login.vue - 智能登录检测

**新增后端可用性预检**：
```javascript
// 在登录前检查后端状态
const isBackendUnavailable = !!sessionStorage.getItem('backendUnavailable')

if (isBackendUnavailable) {
  ElMessage.warning('后端服务当前不可用，正在尝试重新连接...')
  // 尝试连接，成功则继续，失败则给出详细提示
}
```

### 4. BackendStatusMonitor.vue - 智能隐藏

**新增登录页面检测**：
```javascript
const isOnLoginPage = computed(() => route.path === '/login')

// 在登录页面时不显示状态监控
v-if="!isBackendAvailable && showNotification && !isOnLoginPage"
```

## 用户体验流程

### 情况1：后端服务未启动

1. **用户访问任何URL** → 自动跳转到登录页
2. **显示登录页面**（无主界面布局）
3. **页面顶部显示警告**：
   ```
   ⚠️ 后端服务不可用
   无法连接到服务器，请确认：
   • 后端服务已启动（运行在端口 8082）
   • 网络连接正常
   • 防火墙设置正确
   登录后系统会自动重试连接
   ```
4. **用户尝试登录** → 显示"正在尝试重新连接..."
5. **连接失败** → 显示详细错误信息
6. **用户启动后端** → 刷新页面或重新登录即可

### 情况2：后端服务运行中

1. **访问登录页** → 显示正常登录表单，无警告
2. **登录成功** → 进入完整的主界面
3. **主界面功能** → 全部正常可用
4. **后端突然不可用** → 所有操作被重定向到登录页

### 情况3：后端恢复服务

1. **健康检查自动检测** → 清除不可用标记
2. **用户在登录页** → 警告自动消失，可正常登录
3. **用户已登录** → 正常使用，无感知切换

## 技术细节

### 状态管理

**sessionStorage 标记**：
- `backendUnavailable`: 后端不可用标记
- `backendUnavailableTime`: 标记时间戳

**Vuex Store 状态**：
- `isBackendAvailable`: 后端可用性状态
- 自动同步到组件和路由守卫

### 自动恢复机制

1. **健康检查服务**：每30秒自动检查
2. **成功API调用**：自动清除不可用标记
3. **登录成功**：立即清除所有错误状态

### 错误分类处理

1. **网络错误**（ERR_CONNECTION_REFUSED）：
   - 标记后端不可用
   - 强制重定向到登录页
   - 显示详细检查清单

2. **认证错误**（401）：
   - 清除过期凭证
   - 重定向到登录页
   - 提示重新登录

3. **其他错误**：
   - 不影响主界面显示
   - 显示具体错误信息
   - 保持用户会话

## 测试场景

### 测试1：后端未启动时访问应用

```bash
# 确保后端未运行
pkill -f "go run main.go"

# 访问前端应用
open http://localhost:5173/

# 预期：显示纯净登录页面 + 后端不可用警告
```

### 测试2：运行中启动后端

```bash
# 在登录页状态下启动后端
cd backend && go run main.go &

# 刷新页面或尝试登录
# 预期：警告消失，可正常登录
```

### 测试3：已登录用户遇到后端故障

```bash
# 正常使用时停止后端
pkill -f "go run main.go"

# 尝试任何操作
# 预期：立即重定向到登录页
```

## 文件变更总结

### 修改的文件
1. ✅ `frontend/src/App.vue` - 布局显示逻辑
2. ✅ `frontend/src/router/index.js` - 严格路由守卫
3. ✅ `frontend/src/views/Login.vue` - 智能登录检测
4. ✅ `frontend/src/components/BackendStatusMonitor.vue` - 登录页隐藏
5. ✅ `frontend/src/services/healthCheck.js` - 自动状态清理

### 无需修改的文件
- `frontend/src/store/index.js` - 状态管理已完善
- `frontend/src/api/index.js` - 错误处理已完善
- `frontend/src/utils/errorHandler.js` - 错误工具已完善

## 用户反馈改进

### 之前的问题
- ❌ 在主界面框架内显示登录页
- ❌ 侧边栏和菜单仍然显示
- ❌ 用户界面混乱
- ❌ 功能状态不清晰

### 现在的改进
- ✅ 只显示纯净登录页面
- ✅ 完全隐藏主界面布局
- ✅ 界面清晰简洁
- ✅ 功能状态一目了然

## 后续优化建议

1. **添加连接状态指示器**：在登录页添加实时连接状态
2. **自动重试机制**：登录页自动定期尝试连接后端
3. **离线模式**：支持离线查看缓存数据
4. **友好错误页面**：针对不同错误类型显示不同页面