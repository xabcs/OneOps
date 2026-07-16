# K8s Pod 终端集成方案分析

## 方案对比

### 方案一：集成到CMDB终端（堡垒机）

#### 优点
- ✅ 统一的终端入口
- ✅ 用户不需要记忆多个入口
- ✅ 可以复用终端UI组件
- ✅ 统一的审计日志查询

#### 缺点
- ❌ **技术栈完全不同**
  - CMDB终端：SSH协议 → 服务器操作系统
  - K8s终端：WebSocket + K8s API → 容器进程

- ❌ **权限模型不同**
  - CMDB：服务器级权限（用户→服务器）
  - K8s：集群级权限（用户→集群→命名空间）

- ❌ **会话管理不同**
  - CMDB：SSH会话，长期运维场景
  - K8s：临时调试场景，Pod销毁会话即失效

- ❌ **审计需求不同**
  - CMDB：操作命令审计（rm、chmod等）
  - K8s：容器日志、调试命令审计

- ❌ **业务定位冲突**
  - CMDB终端：生产环境运维，需要严格审批
  - K8s终端：开发调试，需要快速访问

---

### 方案二：独立的K8s终端（推荐）✅

#### 优点
- ✅ **职责清晰**
  - K8s有独立的工作负载管理页面
  - 终端与Pod详情页在同一模块，用户体验更好

- ✅ **技术实现独立**
  - SSH终端：使用SSH库（golang.org/x/crypto/ssh）
  - K8s终端：使用K8s client-go + WebSocket

- ✅ **权限模型独立**
  - 可以实现K8s特有的RBAC权限控制
  - 支持命名空间级别的权限隔离

- ✅ **功能优化空间大**
  - 多容器选择
  - 命名空间快速切换
  - Pod日志集成
  - 容器重启、删除等操作

- ✅ **会话管理独立**
  - K8s会话表：k8s_sessions
  - CMDB会话表：bastion_sessions
  - 审计日志不会混淆

#### 缺点
- ⚠️ 用户需要记住两个入口（但这是合理的分离）

---

## 推荐方案：独立设计 🎯

### 架构设计

```
前端结构：
├── /cmdb/servers          # CMDB服务器管理
│   ├── 列表页
│   └── 详情页
│       └── SSH终端按钮 → 跳转到 /terminal?serverId=xxx
│
├── /terminal              # 独立的SSH终端页面（现有）
│   └── 堡垒机会话管理
│
├── /k8s/workloads         # K8s工作负载管理
│   ├── Pod列表
│   └── Pod详情页
│       └── 容器终端按钮 → 在新标签页打开Pod终端
│
└── /k8s/terminal          # 独立的K8s终端页面（新增）
    └── K8s会话管理
```

### 为什么选择独立设计？

#### 1. 业务场景本质不同

| 维度 | CMDB终端 | K8s终端 |
|------|----------|---------|
| **目标对象** | 物理服务器/虚拟机 | 容器Pod |
| **连接方式** | SSH协议 | K8s exec API |
| **会话生命周期** | 长期（小时到天） | 短期（分钟到小时） |
| **主要用途** | 生产运维、配置管理 | 问题排查、调试 |
| **权限要求** | 服务器访问审批 | 集群/命名空间权限 |
| **审计重点** | 系统命令、文件操作 | 容器日志、应用命令 |

#### 2. 技术实现差异巨大

**CMDB终端（SSH）：**
```go
// 使用SSH协议
import "golang.org/x/crypto/ssh"

session, _ := client.NewSession()
session.Stdout = writer
session.Stdin = reader
session.Shell()
```

**K8s终端（WebSocket + K8s API）：**
```go
// 使用K8s exec API
import "k8s.io/client-go/tools/remotecommand"

req := client.CoreV1().RESTClient().Post().
    Resource("pods").Name(podName).
    SubResource("exec")

executor, _ := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
executor.StreamWithContext(ctx, options)
```

两种协议完全不同，强行集成会增加复杂度。

#### 3. 用户体验更好

**当前用户体验：**
```
用户进入K8s管理 → 查看Pod列表 → 点击"终端" →
→ 在新标签页打开独立的K8s终端
→ 可以同时查看Pod详情、日志、事件
```

**如果集成到CMDB终端：**
```
用户进入K8s管理 → 查看Pod列表 → 点击"终端" →
→ 跳转到CMDB终端页面
→ 需要重新选择连接类型（SSH vs K8s）
→ 操作步骤增加，体验割裂
```

#### 4. 安全隔离

**CMDB终端：**
- 通常需要更严格的审批流程
- 可能涉及生产服务器
- 操作影响范围大

**K8s终端：**
- 通常是开发/测试环境
- Pod销毁可以快速恢复
- 影响范围可控

分开设计可以实施不同的安全策略。

---

## 实现建议

### 1. 前端实现

**K8s终端入口：**
```vue
<!-- frontend/src/views/k8s/resources/pods/detail.vue -->
<template>
  <ElButton
    v-for="container in pod.containers"
    :key="container.name"
    @click="handleTerminal(container.name)"
  >
    终端: {{ container.name }}
  </ElButton>
</template>

<script setup>
function handleTerminal(containerName) {
  // 在新标签页打开
  const url = `/k8s/terminal?clusterId=${clusterId}&namespace=${namespace}&podName=${podName}&container=${containerName}`
  window.open(url, '_blank')
}
</script>
```

### 2. 后端实现

**独立的数据表：**
```sql
-- CMDB SSH会话
CREATE TABLE bastion_sessions (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  server_id BIGINT,  -- 服务器ID
  ...
);

-- K8s Pod会话
CREATE TABLE k8s_sessions (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  cluster_id BIGINT,  -- 集群ID
  pod_name VARCHAR(255),
  namespace VARCHAR(100),
  ...
);
```

**独立的权限控制：**
```go
// CMDB权限
func CheckServerAccess(userID, serverID) bool {
    // 检查用户是否有服务器访问权限
}

// K8s权限
func CheckClusterAccess(userID, clusterID) bool {
    // 检查用户是否有集群访问权限
}
```

### 3. 审计隔离

**CMDB审计：**
- 操作命令：rm -rf /data
- 风险等级：HIGH
- 审批流程：需要

**K8s审计：**
- 操作命令：kubectl logs xxx
- 风险等级：LOW
- 审批流程：无需

---

## 结论

### ✅ 推荐：独立设计

**理由总结：**
1. **业务场景不同**：生产运维 vs 开发调试
2. **技术栈不同**：SSH vs K8s API
3. **权限模型不同**：服务器级 vs 集群级
4. **用户体验更好**：功能内聚，操作路径短
5. **安全隔离**：不同安全等级的实施不同策略
6. **代码维护性**：职责单一，易于理解和维护

**实施建议：**
- K8s终端放在 `/k8s/terminal` 路径下
- 从Pod详情页点击"终端"在新标签页打开
- 使用独立的会话管理表（k8s_sessions）
- 实现独立的权限控制和审计日志

这样的设计既符合微服务的单一职责原则，又能为用户提供更好的使用体验。
