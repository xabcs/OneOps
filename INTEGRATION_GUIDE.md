# OneOps K8s诊断功能集成指南

## 🎯 集成概述

将Arthas诊断功能集成到OneOps的K8s集群管理中，用户可以通过Web界面直接对Java应用进行实时诊断。

## 📋 集成步骤

### 1. 后端集成

#### 1.1 安装依赖
```bash
cd backend
npm install axios @types/axios
```

#### 1.2 添加诊断路由
```typescript
// backend/server.ts
import diagnosticRoutes from './routes/diagnosticRoutes';

// 在现有路由后添加
app.use('/api/v1/diagnostic', diagnosticRoutes);
```

#### 1.3 注册权限
```sql
-- 在权限表中添加诊断相关权限
INSERT INTO permissions (name, description, category) VALUES
('k8s:diagnostic:execute', '执行K8s诊断', 'k8s'),
('k8s:diagnostic:view', '查看K8s诊断结果', 'k8s');
```

### 2. 前端集成

#### 2.1 添加诊断菜单
```vue
<!-- frontend/src/views/k8s/layout.vue -->
<template>
  <div class="k8s-layout">
    <el-menu>
      <!-- 现有菜单项 -->
      <el-menu-item index="/k8s/clusters">
        <el-icon><Monitor /></el-icon>
        <span>集群管理</span>
      </el-menu-item>

      <el-menu-item index="/k8s/pods">
        <el-icon><Box /></el-icon>
        <span>Pod管理</span>
      </el-menu-item>

      <!-- 新增诊断菜单 -->
      <el-menu-item index="/k8s/diagnostic">
        <el-icon><Operation /></el-icon>
        <span>诊断中心</span>
      </el-menu-item>
    </el-menu>

    <router-view />
  </div>
</template>
```

#### 2.2 在Pod列表中添加诊断按钮
```vue
<!-- frontend/src/views/k8s/pods/index.vue -->
<template>
  <el-table :data="pods">
    <!-- 现有列 -->
    <el-table-column label="操作" width="200">
      <template #default="{ row }">
        <el-button
          v-if="row.diagnostic.enabled"
          type="primary"
          size="small"
          @click="goToDiagnostic(row)"
        >
          <el-icon><Operation /></el-icon>
          诊断
        </el-button>
        <!-- 其他操作按钮 -->
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
const goToDiagnostic = (pod) => {
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

### 3. 部署Agent

#### 3.1 选择部署方式

**方式A：DaemonSet方式（推荐用于开发/测试）**
```bash
# 构建DaemonSet Agent
cd agent/diagnostic/daemonset-improved
docker build -t daemonset-diagnostic-agent:latest .

# 部署到K8s
kubectl apply -f k8s-daemonset.yaml
```

**方式B：Sidecar方式（推荐用于生产）**
```bash
# 构建Sidecar Agent
cd agent/diagnostic/arthas-sidecar
go build -o arthas-sidecar main.go
docker build -t arthas-sidecar:latest .

# 修改应用部署配置
# 在需要诊断的Java应用Deployment中添加：
spec:
  template:
    spec:
      shareProcessNamespace: true
      containers:
      - name: java-app
        image: myapp:latest
      # 添加以下内容
      - name: arthas-agent
        image: arthas-sidecar:latest
        ports:
        - containerPort: 8563
        env:
        - name: TARGET_PID
          value: ""  # 自动检测Java进程
```

### 4. 数据库配置

#### 4.1 创建诊断历史表
```sql
CREATE TABLE IF NOT EXISTS diagnostic_history (
  id INT AUTO_INCREMENT PRIMARY KEY,
  cluster_id VARCHAR(50) NOT NULL,
  namespace VARCHAR(100) NOT NULL,
  pod_name VARCHAR(100) NOT NULL,
  command VARCHAR(50) NOT NULL,
  args TEXT,
  result TEXT,
  user VARCHAR(50),
  timestamp DATETIME NOT NULL,
  INDEX idx_cluster (cluster_id),
  INDEX idx_pod (namespace, pod_name),
  INDEX idx_timestamp (timestamp)
);
```

### 5. 测试集成

#### 5.1 后端测试
```bash
# 测试获取诊断命令
curl http://localhost:3000/api/v1/diagnostic/commands

# 测试获取可诊断Pods
curl http://localhost:3000/api/v1/diagnostic/pods/cluster-1/default

# 测试执行诊断
curl -X POST http://localhost:3000/api/v1/diagnostic/execute/cluster-1/default/myapp \
  -H "Content-Type: application/json" \
  -d '{
    "command": "thread",
    "args": ["-n", "10"],
    "timeout": 60
  }'
```

#### 5.2 前端测试
```bash
# 启动前端开发服务器
cd frontend
npm run dev

# 访问诊断页面
http://localhost:5173/k8s/diagnostic
```

## 🎯 功能验证

### 验证清单

1. **界面显示**
   - [ ] 诊断页面正常显示
   - [ ] 集群和命名空间选择器工作正常
   - [ ] Pod列表正确显示Java应用
   - [ ] 诊断状态标识正确

2. **诊断功能**
   - [ ] 选择Pod后可以查看详情
   - [ ] 诊断命令列表正确显示
   - [ ] 参数配置界面正常工作
   - [ ] 执行诊断按钮可用性正确

3. **结果展示**
   - [ ] 诊断结果正确显示
   - [ ] 错误信息友好展示
   - [ ] 结果可以复制和下载
   - [ ] 历史记录正常保存

4. **权限控制**
   - [ ] 无权限用户无法访问
   - [ ] 权限验证正确工作
   - [ ] 敏感操作有确认提示

## 🔧 故障排查

### 常见问题

**Q: 诊断页面显示"暂无可诊断的Java应用"**
- 检查Agent是否正确部署
- 验证Java应用Pod是否正常运行
- 检查网络连接和权限

**Q: 执行诊断时提示"Pod诊断Agent不可用"**
- 检查Sidecar/DaemonSet Agent状态
- 验证Agent健康检查是否正常
- 查看Agent日志排查问题

**Q: 诊断执行超时**
- 检查网络连接
- 增加超时时间配置
- 验证Agent负载情况

**Q: 历史记录不显示**
- 检查数据库连接
- 验证权限配置
- 查看后端日志

## 📊 监控和日志

### 日志配置

```typescript
// backend/middleware/diagnosticLogger.ts
export const diagnosticLogger = (req, res, next) => {
  const originalSend = res.send;
  res.send = function(data) {
    console.log(`[诊断日志] ${req.method} ${req.url} - ${res.statusCode}`);

    if (req.url.includes('/execute')) {
      console.log(`[诊断执行] 用户: ${req.user?.username}, ` +
                 `集群: ${req.params.clusterId}, ` +
                 `Pod: ${req.params.namespace}/${req.params.podName}, ` +
                 `命令: ${req.body.command}`);
    }

    originalSend.call(this, data);
  };
  next();
};
```

### 监控指标

```typescript
// 添加Prometheus监控
import { Counter, Histogram } from 'prom-client';

export const diagnosticMetrics = {
  executions: new Counter({
    name: 'diagnostic_executions_total',
    help: '诊断执行总次数',
    labelNames: ['cluster', 'command', 'status']
  }),

  duration: new Histogram({
    name: 'diagnostic_duration_seconds',
    help: '诊断执行耗时',
    labelNames: ['cluster', 'command']
  }),

  errors: new Counter({
    name: 'diagnostic_errors_total',
    help: '诊断错误总次数',
    labelNames: ['cluster', 'error_type']
  })
};
```

## 🎓 使用指南

### 用户操作流程

1. **访问诊断页面**
   - 导航到：K8s管理 → 诊断中心
   - 或从Pod列表点击"诊断"按钮

2. **选择目标Pod**
   - 选择集群和命名空间
   - 在Pod列表中选择目标Java应用
   - 查看Pod详情和诊断方式

3. **配置诊断命令**
   - 选择诊断命令类型
   - 配置相关参数
   - 查看命令说明

4. **执行诊断**
   - 点击"执行诊断"按钮
   - 等待诊断结果
   - 查看结果输出

5. **结果管理**
   - 复制结果到剪贴板
   - 下载结果为文件
   - 查看历史记录

### 高级功能

**实时监控**：
- 选择"实时仪表板"命令
- 设置自动刷新间隔
- 持续监控应用状态

**批量诊断**：
- 在Pod列表中选择多个Pod
- 批量执行相同诊断命令
- 对比分析结果

**历史分析**：
- 查看历史诊断记录
- 分析问题趋势
- 导出诊断报告

## 🚀 性能优化

### 前端优化
- 使用虚拟滚动处理大量Pod
- 实现结果分页显示
- 添加缓存机制

### 后端优化
- 实现连接池复用
- 添加结果缓存
- 异步处理长时间诊断

### Agent优化
- 调整资源限制
- 优化命令执行效率
- 实现连接复用

## ✅ 集成完成

完成以上步骤后，OneOps系统将具备完整的K8s Java应用诊断功能，用户可以通过友好的Web界面进行实时诊断分析。

**下一步建议**：
1. 根据实际使用情况调整UI布局
2. 添加更多诊断命令类型
3. 实现结果分析功能
4. 添加告警和通知机制