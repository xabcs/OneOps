# Diagnostic Agent for Kubernetes

一个用于Kubernetes集群的Java应用诊断Agent，基于DaemonSet模式部署，支持Arthas诊断功能。

## 功能特性

- **DaemonSet部署**: 自动在每个K8s节点上运行
- **Arthas集成**: 支持thread、heap、jvm等诊断命令
- **RESTful API**: 提供HTTP接口进行诊断操作
- **节点级过滤**: 只处理当前节点上的Pod
- **健康检查**: 内置健康检查和监控功能
- **RBAC支持**: 包含完整的权限配置

## 快速开始

### 1. 构建镜像

```bash
cd agent/diagnostic
chmod +x build.sh deploy.sh

# 构建Docker镜像
./build.sh
```

### 2. 部署到K8s

```bash
# 部署到test命名空间
./deploy.sh

# 或者手动部署
kubectl apply -f k8s-deployment.yaml
```

### 3. 验证部署

```bash
# 查看Pod状态
kubectl get pods -l app=diagnostic-agent -n test

# 查看日志
kubectl logs -l app=diagnostic-agent -n test

# 测试健康检查
kubectl exec -it <POD_NAME> -n test -- wget -q -O- http://localhost:8888/health
```

## API接口

### 健康检查
```bash
GET /health
```

**响应示例:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "node_name": "node-1",
  "namespace": "test",
  "pod_ip": "10.244.1.10",
  "capabilities": ["thread", "heap", "jvm", "dashboard", "logger"]
}
```

### 列出Pod
```bash
GET /pods?namespace=test
```

### 获取Pod详情
```bash
GET /pods/{namespace}/{name}
```

### 执行诊断
```bash
POST /diagnostic
Content-Type: application/json

{
  "pod_name": "myapp-pod-xxx",
  "namespace": "test",
  "container": "myapp",
  "command": "thread",
  "args": ["-n", "20"]
}
```

**支持的诊断命令:**
- `thread` - 线程分析
- `heap` - 堆内存分析  
- `jvm` - JVM信息
- `dashboard` - 仪表板数据
- `logger` - 日志管理

**响应示例:**
```json
{
  "status": "success",
  "output": "诊断输出...",
  "timestamp": 1698765432,
  "duration_ms": 1234,
  "metadata": {
    "pod_name": "myapp-pod-xxx",
    "namespace": "test",
    "container_name": "myapp",
    "node_name": "node-1",
    "pod_ip": "10.244.2.15"
  }
}
```

## 架构说明

### DaemonSet模式
Agent以DaemonSet模式部署，确保每个K8s节点都运行一个Agent实例。这样可以:

1. **节点本地化**: Agent只诊断同节点上的Pod，减少网络开销
2. **高可用性**: 每个节点独立运行，单点故障不影响其他节点
3. **资源优化**: 共享节点资源，避免过度分配

### 网络架构
```
┌─────────────┐
│   用户Web   │
└──────┬──────┘
       │ HTTP
       ▼
┌──────────────┐
│  OneOps后端  │
└──────┬───────┘
       │ K8s Service
       ▼
┌──────────────┐
│ Diagnostic   │ ← DaemonSet (每个节点)
│   Service    │
└──────┬───────┘
       │ Pod-to-Pod
       ▼
┌──────────────┐
│  目标Java Pod │
└──────────────┘
```

### 安全特性
- **RBAC权限**: 只授予必要的Pod操作权限
- **命名空间隔离**: 默认只在test命名空间工作
- **只读操作**: 不修改目标Pod的状态
- **非root用户**: 容器以普通用户运行

## 使用场景

### 1. CPU异常诊断
```bash
curl -X POST http://diagnostic-service:8888/diagnostic \
  -H 'Content-Type: application/json' \
  -d '{
    "pod_name": "myapp-xxx",
    "namespace": "test", 
    "command": "thread",
    "args": ["-n", "10"]
  }'
```

### 2. 内存泄漏分析
```bash
curl -X POST http://diagnostic-service:8888/diagnostic \
  -H 'Content-Type: application/json' \
  -d '{
    "pod_name": "myapp-xxx",
    "namespace": "test",
    "command": "heap",
    "args": ["--gc"]
  }'
```

### 3. JVM信息查看
```bash
curl -X POST http://diagnostic-service:8888/diagnostic \
  -H 'Content-Type: application/json' \
  -d '{
    "pod_name": "myapp-xxx",
    "namespace": "test",
    "command": "jvm"
  }'
```

## 监控和日志

### 查看Agent日志
```bash
# 所有节点日志
kubectl logs -l app=diagnostic-agent -n test

# 特定节点日志
kubectl logs <POD_NAME> -n test
```

### 监控Agent状态
```bash
# DaemonSet状态
kubectl get daemonset diagnostic-agent -n test

# Pod详情
kubectl describe pod <POD_NAME> -n test

# 资源使用
kubectl top pods -l app=diagnostic-agent -n test
```

## 故障排查

### Agent无法启动
```bash
# 检查镜像是否存在
kubectl describe pod <POD_NAME> -n test | grep Image

# 查看启动日志
kubectl logs <POD_NAME> -n test --previous
```

### 诊断失败
```bash
# 检查RBAC权限
kubectl auth can-i get pods --as=system:serviceaccount:test:diagnostic-agent

# 检查网络连接
kubectl exec <POD_NAME> -n test -- wget -q -O- http://localhost:8888/health
```

### Pod节点不匹配
Agent只处理同节点上的Pod，确保目标Pod与Agent在同一节点:
```bash
# 检查Pod分布
kubectl get pods -n test -o wide

# 查看Agent节点
kubectl get pods -l app=diagnostic-agent -n test -o wide
```

## 扩展和定制

### 添加新的诊断命令
在`main.go`的`buildArthasCommand`函数中添加新的命令支持:

```go
case "mycommand":
    return []string{"sh", "-c", "custom-diagnostic-script"}
```

### 修改资源限制
编辑`k8s-deployment.yaml`:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "1000m"
```

### 多命名空间支持
修改RBAC配置，添加更多命名空间:
```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: diagnostic-agent
  namespace: production  # 添加其他命名空间
rules:
# ... 相同的规则
```

## 与OneOps集成

这个Agent可以与OneOps平台无缝集成:

1. **后端集成**: 在OneOps后端调用Agent的API
2. **前端展示**: 在Web界面显示诊断结果
3. **自动化触发**: 基于监控指标自动执行诊断

## 维护和升级

### 升级Agent
```bash
# 构建新版本
docker build -t diagnostic-agent:v2.0 .

# 更新部署
kubectl set image daemonset/diagnostic-agent \
  diagnostic-agent=diagnostic-agent:v2.0 -n test
```

### 回滚
```bash
kubectl rollout undo daemonset/diagnostic-agent -n test
```

### 卸载
```bash
kubectl delete -f k8s-deployment.yaml
```

## 许可证
MIT License

## 支持
如有问题，请提Issue或联系开发团队。