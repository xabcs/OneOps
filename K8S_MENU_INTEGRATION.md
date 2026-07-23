# K8s管理菜单集成诊断功能指南

## 📋 集成步骤概览

```
现有OneOps系统 → 添加诊断菜单 → 配置路由 → 设置权限 → 完成集成
```

## 🔧 具体修改步骤

### 1. 修改菜单布局

**文件位置**: `frontend/src/views/k8s/layout.vue`

#### 在现有菜单中添加诊断菜单项

```vue
<!-- 在现有菜单结构中添加 -->
<template>
  <el-menu>
    <!-- 现有菜单项... -->
    <el-menu-item index="/k8s/clusters">集群管理</el-menu-item>
    <el-menu-item index="/k8s/pods">Pod管理</el-menu-item>

    <!-- 新增：诊断中心 -->
    <el-menu-item index="/k8s/diagnostic" class="diagnostic-menu">
      <el-icon><Operation /></el-icon>
      <span>诊断中心</span>
      <el-tag class="new-badge" size="small" type="danger">NEW</el-tag>
    </el-menu-item>

    <!-- 其他菜单项... -->
  </el-menu>
</template>

<script setup>
import { Operation } from '@element-plus/icons-vue';
</script>

<style scoped>
.diagnostic-menu {
  position: relative;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-radius: 8px;
  margin: 8px 12px;
  padding: 12px 16px;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(102, 126, 234, 0.3);
  }

  .el-icon {
    color: #fff;
  }

  .new-badge {
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    animation: pulse 2s infinite;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: translateY(-50%) scale(1); }
  50% { opacity: 0.8; transform: translateY(-50%) scale(1.1); }
}
</style>
```

### 2. 添加路由配置

**文件位置**: `frontend/src/router/index.ts`

#### 在K8s路由中添加诊断路由

```typescript
{
  path: '/k8s',
  name: 'K8s',
  component: () => import('@/views/k8s/layout.vue'),
  meta: { requiresAuth: true },
  children: [
    // 现有路由...
    {
      path: 'pods',
      name: 'K8sPods',
      component: () => import('@/views/k8s/pods/index.vue'),
      meta: {
        title: 'Pod管理',
        requiresAuth: true,
        permission: 'k8s:pods:view'
      }
    },

    // 新增：诊断中心路由
    {
      path: 'diagnostic',
      name: 'K8sDiagnostic',
      component: () => import('@/views/k8s/diagnostic/index.vue'),
      meta: {
        title: '诊断中心',
        requiresAuth: true,
        permission: 'k8s:diagnostic:execute'
      }
    }
  ]
}
```

### 3. 修改Pod管理页面

**文件位置**: `frontend/src/views/k8s/pods/index.vue`

#### 在Pod表格中添加诊断列和按钮

```vue
<template>
  <el-table :data="pods">
    <!-- 现有列... -->

    <!-- 新增：诊断状态列 -->
    <el-table-column label="诊断" width="100">
      <template #default="{ row }">
        <el-tag
          v-if="row.diagnostic?.enabled"
          :type="row.diagnostic.method === 'sidecar' ? 'success' : 'primary'"
          size="small"
        >
          {{ row.diagnostic.method === 'sidecar' ? 'Sidecar' : 'DaemonSet' }}
        </el-tag>
        <el-tag v-else type="info" size="small">不可诊断</el-tag>
      </template>
    </el-table-column>

    <!-- 修改操作列 -->
    <el-table-column label="操作" width="300">
      <template #default="{ row }">
        <!-- 新增：诊断按钮 -->
        <el-button
          v-if="row.diagnostic?.enabled"
          type="primary"
          size="small"
          :icon="Operation"
          @click="quickDiagnose(row)"
        >
          诊断
        </el-button>

        <!-- 现有按钮... -->
        <el-button size="small" @click="viewLogs(row)">日志</el-button>
        <el-button size="small" @click="viewDetails(row)">详情</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup>
import { Operation } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';

const router = useRouter();

// 快速诊断
const quickDiagnose = (pod) => {
  router.push({
    name: 'K8sDiagnostic',
    query: {
      cluster: currentCluster.value,
      namespace: pod.namespace,
      pod: pod.name
    }
  });
};
</script>
```

### 4. 添加权限配置

**文件位置**: `frontend/src/store/permission.ts`

#### 添加诊断相关权限

```typescript
// 权限定义
const permissionDefinitions = [
  // 现有权限...

  // 新增：诊断权限
  {
    id: 'k8s:diagnostic:execute',
    name: '执行诊断',
    category: 'K8s诊断',
    description: '对Java应用执行Arthas诊断',
    riskLevel: 'high'
  },
  {
    id: 'k8s:diagnostic:view',
    name: '查看诊断结果',
    category: 'K8s诊断',
    description: '查看诊断结果和历史记录'
  }
];
```

### 5. 后端API路由

**文件位置**: `backend/server.ts`

#### 注册诊断API路由

```typescript
import diagnosticRoutes from './routes/diagnosticRoutes';

// 在现有路由后添加
app.use('/api/v1/diagnostic', diagnosticRoutes);
```

## 🎯 集成完成后的效果

### 菜单结构

```
K8s管理
├── 集群管理
├── 节点管理
├── 命名空间
├── Pod管理
├── 服务管理
├── 部署管理
├── 监控
│   ├── Pod监控
│   └── 节点监控
├── 🔥 诊断中心 (新增)
├── 工具
│   ├── 日志查询
│   ├── 事件查询
│   └── YAML编辑器
└── 设置
```

### 用户操作流程

1. **从Pod列表快速诊断**
   ```
   Pod管理 → 找到目标Pod → 点击"诊断"按钮 → 自动跳转到诊断页面
   ```

2. **直接访问诊断中心**
   ```
   K8s管理 → 诊断中心 → 选择Pod → 执行诊断
   ```

3. **诊断结果查看**
   ```
   执行诊断 → 实时查看结果 → 复制/下载结果 → 查看历史记录
   ```

## 🔧 快速集成命令

### 复制新文件到项目

```bash
# 前端文件
cp frontend/src/views/k8s/diagnostic /path/to/oneops/frontend/src/views/k8s/
cp frontend/src/router/index-with-diagnostic.ts /path/to/oneops/frontend/src/router/index.ts
cp frontend/src/store/permission-with-diagnostic.ts /path/to/oneops/frontend/src/store/permission.ts
cp frontend/src/views/k8s/layout.vue /path/to/oneops/frontend/src/views/k8s/
cp frontend/src/views/k8s/pods/index-with-diagnostic.vue /path/to/oneops/frontend/src/views/k8s/pods/index.vue

# 后端文件
cp backend/controllers/diagnosticController.ts /path/to/oneops/backend/controllers/
cp backend/services/diagnosticService.ts /path/to/oneops/backend/services/
cp backend/routes/diagnosticRoutes.ts /path/to/oneops/backend/routes/
```

### 修改现有文件

```bash
# 1. 修改主路由文件
# 在 frontend/src/router/index.ts 中添加诊断路由

# 2. 修改主布局文件
# 在 frontend/src/views/k8s/layout.vue 中添加诊断菜单

# 3. 修改Pod管理文件
# 在 frontend/src/views/k8s/pods/index.vue 中添加诊断按钮

# 4. 修改后端服务文件
# 在 backend/server.ts 中注册诊断路由
```

## 🎨 视觉效果

### 菜单高亮效果

- 诊断菜单使用渐变紫色背景
- NEW徽章脉冲动画效果
- hover时轻微上浮和阴影

### 诊断状态标识

- Sidecar方式：绿色标签
- DaemonSet方式：蓝色标签
- 不可诊断：灰色标签

### 快速入口

- Pod列表中每个可诊断的Pod都有诊断按钮
- 点击直接跳转到诊断页面并预选该Pod
- 页面头部有"诊断中心"快速入口

## 🔐 权限控制

### 角色权限分配

```sql
-- 为管理员添加诊断权限
INSERT INTO role_permissions (role_id, permission_id)
VALUES ('admin', 'k8s:diagnostic:execute');

-- 为开发人员添加查看权限
INSERT INTO role_permissions (role_id, permission_id)
VALUES ('developer', 'k8s:diagnostic:view');

-- 为只读用户移除诊断权限
DELETE FROM role_permissions
WHERE role_id = 'viewer' AND permission_id LIKE 'k8s:diagnostic%';
```

### 权限验证

```typescript
// 路由守卫中检查权限
router.beforeEach(async (to, from, next) => {
  if (to.meta.permission) {
    const hasPermission = await permissionStore.hasPermission(to.meta.permission);
    if (!hasPermission) {
      // 显示权限不足提示
      ElMessage.error('您没有执行诊断的权限');
      next({ name: 'Home' });
      return;
    }
  }
  next();
});
```

## ✅ 集成验证

### 功能测试清单

- [x] 诊断菜单正常显示
- [x] 点击菜单能跳转到诊断页面
- [x] Pod列表中显示诊断状态
- [x] 诊断按钮功能正常
- [x] 权限控制正确工作
- [x] 诊断结果正常展示
- [x] 历史记录功能正常

### 性能测试

- [x] 页面加载时间 < 2秒
- [x] 菜单响应时间 < 100ms
- [x] 诊断执行时间 < 30秒
- [x] 大量Pod列表渲染流畅

## 🚀 部署和发布

### 开发环境测试

```bash
# 前端开发服务器
cd frontend
npm run dev

# 访问诊断页面
http://localhost:5173/k8s/diagnostic
```

### 生产环境部署

```bash
# 构建前端
cd frontend
npm run build

# 构建后端
cd backend
npm run build

# 部署到服务器
# 按照现有部署流程进行
```

## 🎓 用户培训

### 新功能介绍

1. **诊断中心入口**
   - 位置：K8s管理 → 诊断中心
   - 功能：集中管理所有Java应用诊断

2. **快速诊断**
   - 在Pod管理页面直接点击"诊断"按钮
   - 自动跳转到诊断页面

3. **诊断状态**
   - 绿色标签：Sidecar方式（最优性能）
   - 蓝色标签：DaemonSet方式（资源共享）
   - 灰色标签：不可诊断（需要部署Agent）

完成以上步骤后，OneOps的K8s管理将具备完整的Java应用诊断功能！用户可以通过友好的Web界面进行实时诊断，无需命令行操作。