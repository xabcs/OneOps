# 研究：阿里云 ACK 用户权限控制实现方式

- **查询**：研究阿里云 ACK（Alibaba Cloud Kubernetes）的用户权限控制实现方式，重点关注 RAM 集成、权限模型、RBAC 策略、多租户隔离和审计日志
- **范围**：外部研究（阿里云官方文档、技术博客、最佳实践）
- **日期**：2025-06-15

## 核心架构概述

阿里云 ACK 的权限控制采用**双层权限架构**：

```
┌─────────────────────────────────────────────────────────────┐
│                     阿里云 ACK 权限控制架构                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  阿里云 RAM 层    │  ←→    │  K8s RBAC 层     │         │
│  │  (云平台级权限)    │         │  (集群级权限)      │         │
│  └──────────────────┘         └──────────────────┘         │
│           ↓                            ↓                    │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  RAM 权限策略     │         │  Role/ClusterRole│         │
│  │  - 管理集群       │         │  - RoleBinding   │         │
│  │  - 查看集群       │         │  - 命名空间隔离   │         │
│  │  - 绑定 RBAC      │         │  - 资源级控制     │         │
│  └──────────────────┘         └──────────────────┘         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 一、RAM 集成方式

### 1.1 RAM 用户/角色访问 ACK

**授权模式**：RAM 子账号或 RAM 角色（如 ECS 实例角色）通过**权限策略**获得对 ACK 集群的访问权限。

**权限策略类型**：

| 权限类型 | 权限描述 | 使用场景 |
|---------|---------|---------|
| **AliyunCSFullAccess** | ACK 完全管理权限 | 集群管理员、DevOps 工程师 |
| **AliyunCSReadOnlyAccess** | ACK 只读权限 | 审计人员、监控查看者 |
| **自定义权限策略** | 细粒度权限控制 | 特定场景的权限限制 |

**权限策略示例**：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cs:DescribeClusters",
        "cs:DescribeClusterNodes",
        "cs:GetClusterCredentials"
      ],
      "Resource": [
        "acs:cs:*:*:cluster/<cluster-id>"
      ]
    }
  ]
}
```

**关键权限动作**：

- `cs:CreateCluster` - 创建集群
- `cs:DeleteCluster` - 删除集群
- `cs:DescribeClusters` - 查看集群列表
- `cs:GetClusterCredentials` - **关键**：获取集群 KubeConfig（包含 K8s API 访问凭证）
- `cs:ScaleOutCluster` - 扩容集群节点
- `cs:AttachClusterSSG` - 绑定安全组

### 1.2 授权流程

```
步骤 1: 创建 RAM 用户/角色
   ↓
步骤 2: 分配 RAM 权限策略
   ↓
步骤 3: 用户通过 RAM 控制台或 API 获取集群访问凭证
   ↓
步骤 4: 凭证包含临时 Token（STS Token）或 KubeConfig
   ↓
步骤 5: 使用凭证访问 K8s API（受 RBAC 控制）
```

**关键点**：
- RAM 权限**不直接控制** K8s 资源（Pod、Service 等）
- RAM 权限控制**能否拿到** K8s API 访问凭证
- 拿到凭证后的**细粒度权限**由 K8s RBAC 控制

## 二、K8s RBAC 权限模型

### 2.1 RBAC 核心概念

ACK 完全使用 Kubernetes 原生 RBAC，包含四个核心对象：

| 对象 | 作用 | 作用域 |
|------|------|--------|
| **Role** | 定义命名空间内权限 | 单命名空间 |
| **ClusterRole** | 定义集群级权限 | 全集群 |
| **RoleBinding** | 绑定 Role/ClusterRole 到用户/组 | 单命名空间 |
| **ClusterRoleBinding** | 绑定 ClusterRole 到用户/组 | 全集群 |

### 2.2 常见授权场景

#### 场景 1：只读访问某命名空间

```yaml
# Role: namespace-readonly
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: production
  name: namespace-readonly
rules:
- apiGroups: ["*"]
  resources: ["pods", "services", "deployments"]
  verbs: ["get", "list", "watch"]

---
# RoleBinding: 绑定到 RAM 用户
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: readonly-binding-prod
  namespace: production
subjects:
- kind: User
  name: <RAM_USER_ARN>  # 格式: acs:ram::<account-id>:user/<username>
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: namespace-readonly
  apiGroup: rbac.authorization.k8s.io
```

#### 场景 2：开发者管理某命名空间

```yaml
# Role: namespace-admin
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: development
  name: namespace-admin
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]

---
# RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developer-binding-dev
  namespace: development
subjects:
- kind: User
  name: <RAM_USER_ARN>
roleRef:
  kind: Role
  name: namespace-admin
  apiGroup: rbac.authorization.k8s.io
```

#### 场景 3：集群管理员（谨慎使用）

```yaml
# ClusterRoleBinding: 集群管理员
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-admin-binding
subjects:
- kind: User
  name: <RAM_USER_ARN>
roleRef:
  kind: ClusterRole
  name: cluster-admin  # 内置超级管理员角色
  apiGroup: rbac.authorization.k8s.io
```

### 2.3 RAM 用户身份映射

在 RoleBinding/ClusterRoleBinding 中，RAM 用户的身份格式：

```
acs:ram::<account-id>:user/<username>

示例：
acs:ram::1234567890123456:user/zhangsan
acs:ram::1234567890123456:role/application
```

**验证命令**：
```bash
# 查看 RAM 用户的 ARN
aliyun ram GetUser --UserName zhangsan

# 验证权限
kubectl auth can-i get pods --namespace=production --as=acs:ram::1234567890123456:user/zhangsan
```

## 三、多租户隔离机制

### 3.1 命名空间隔离（推荐方案）

**适用场景**：不同团队/项目在同一集群

```yaml
# 命名空间创建
apiVersion: v1
kind: Namespace
metadata:
  name: team-a
  labels:
    team: team-a
    environment: production

---
# Namespace 配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-a-quota
  namespace: team-a
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    persistentvolumeclaims: "5"

---
# LimitRange（限制单资源）
apiVersion: v1
kind: LimitRange
metadata:
  name: team-a-limits
  namespace: team-a
spec:
  limits:
  - max:
      cpu: "2"
      memory: 4Gi
    min:
      cpu: 100m
      memory: 128Mi
    default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 200m
      memory: 256Mi
    type: Container
```

**多租户最佳实践**：

1. **每个团队一个命名空间**：资源隔离 + RBAC 隔离
2. **NetworkPolicy**：网络策略隔离（可选，但推荐）
3. **ResourceQuota**：资源配额防止单租户耗尽集群资源
4. **LimitRange**：限制单 Pod 资源上下限

### 3.2 集群物理隔离（高安全要求）

**适用场景**：不同业务线/安全级别要求高的场景

- 不同集群完全隔离（API Server、etcd、节点）
- 通过 RAM 权限策略控制用户能看到哪些集群
- **成本更高**，但安全性更强

### 3.3 NetworkPolicy 网络隔离

```yaml
# 拒绝团队间访问
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-cross-team
  namespace: team-a
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          team: team-a  # 只允许同命名空间流量
```

## 四、托管版 vs 专有版权限差异

### 4.1 托管版 ACK（ACK Managed）

**特点**：
- 阿里云托管 K8s 控制平面（Master 节点）
- 用户只需管理 Worker 节点

**权限控制差异**：

| 维度 | 托管版 | 专有版 |
|------|--------|--------|
| **RAM 集成** | ✅ 原生支持，通过 ACK API 授权 | ⚠️ 需自行集成 |
| **KubeConfig 获取** | 通过 `cs:GetClusterCredentials` API | 直接访问 Master 节点 |
| **RBAC 创建** | 通过 ACK 控制台/kubectl | 仅通过 kubectl |
| **审计日志** | 集成阿里云 ActionTrail | 自建审计系统 |
| **权限边界** | RAM（集群级）+ RBAC（资源级） | 仅 RBAC |

**托管版优势**：
- RAM 用户/角色可直接通过阿里云控制台获取 KubeConfig
- 无需维护 Master 节点安全和证书管理
- 操作审计集成到 ActionTrail

**托管版限制**：
- 无法直接访问 K8s API Server（通过 ACK API 代理）
- 部分高级 K8s 特性可能受限

### 4.2 专有版 ACK（ACK Dedicated）

**特点**：
- 用户自建 K8s 集群（包括 Master 节点）
- 完全控制 K8s 控制平面

**权限控制差异**：
- **RAM 集成弱**：需自行开发 RAM → K8s 权限映射
- **RBAC 原生**：与标准 K8s 一致
- **自定义身份验证**：可集成企业 LDAP/OAuth

**适用场景**：
- 对 K8s 控制平面有深度定制需求
- 需要与企业 IAM 系统深度集成
- 合规要求必须自建控制平面

## 五、审计日志

### 5.1 托管版审计能力

**阿里云 ActionTrail 集成**：

| 事件类型 | 记录内容 | 审计字段 |
|---------|---------|---------|
| **RAM 操作** | 谁创建了/修改了 RAM 权限策略 | userId、eventName、requestParameters |
| **ACK 集群操作** | 集群创建/删除/扩容 | clusterId、operationType |
| **RBAC 变更** | Role/RoleBinding 创建/修改 | resourceName、requestBody |
| **KubeConfig 获取** | 谁获取了集群访问凭证 | requestId、sourceIp |

**ActionTrail 查询示例**：

```sql
-- 查询所有 KubeConfig 获取事件
SELECT
  userId,
  eventTime,
  sourceIp,
  userAgent
FROM
  actiontrail_events
WHERE
  eventName = 'GetClusterCredentials'
  AND serviceName = 'ContainerService'
ORDER BY
  eventTime DESC;

-- 查询 RBAC 变更事件
SELECT
  userId,
  eventName,
  requestBody,
  eventTime
FROM
  actiontrail_events
WHERE
  serviceName = 'ContainerService'
  AND eventName IN ('CreateRole', 'DeleteRoleBinding')
  AND resourceType = 'cluster';
```

### 5.2 K8s Audit Logs（集群内审计）

**启用审计日志**：
```yaml
# Audit Policy 示例
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
  verbs: ["get", "list", "watch"]
- level: Request
  verbs: ["create", "update", "delete"]
- level: RequestResponse
  namespaces: ["kube-system", "production"]
```

**日志内容示例**：
```json
{
  "kind": "Event",
  "auditID": "12345",
  "stage": "ResponseComplete",
  "requestURI": "/api/v1/namespaces/production/pods",
  "verb": "create",
  "user": {
    "username": "acs:ram::1234567890123456:user/zhangsan",
    "groups": ["system:authenticated"]
  },
  "sourceIPs": ["192.168.1.100"],
  "responseStatus": {
    "code": 201
  },
  "responseObject": {
    "kind": "Pod",
    "name": "nginx-deployment-xxx"
  }
}
```

### 5.3 审计日志最佳实践

1. **集中存储**：日志投递到 SLS（日志服务）或 OSS
2. **告警规则**：
   - 监控 `GetClusterCredentials` 事件（异常访问凭证获取）
   - 监控 ClusterRoleBinding 创建（权限提升）
   - 监控 Pod exec/attach（敏感操作）
3. **合规留存**：审计日志至少保留 6 个月（等保要求）

## 六、授权场景实操

### 6.1 场景：新开发者加入团队

**目标**：让新开发者 `lisi` 能管理 `development` 命名空间

**步骤**：

1. **创建 RAM 用户**（如果不存在）
   ```bash
   aliyun ram CreateUser --UserName lisi
   ```

2. **分配 RAM 基础权限**（至少能查看集群）
   ```json
   {
     "Version": "1",
     "Statement": [{
       "Effect": "Allow",
       "Action": ["cs:DescribeClusters", "cs:GetClusterCredentials"],
       "Resource": ["acs:cs:*:*:cluster/my-cluster"]
     }]
   }
   ```

3. **创建 K8s Role**
   ```yaml
   # developer-role.yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     namespace: development
     name: developer-role
   rules:
   - apiGroups: ["apps"]
     resources: ["deployments", "replicasets"]
     verbs: ["get", "list", "create", "update", "delete"]
   - apiGroups: [""]
     resources: ["pods", "services"]
     verbs: ["get", "list", "create", "update", "delete"]
   ```

4. **创建 RoleBinding**
   ```yaml
   # developer-binding.yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: developer-binding
     namespace: development
   subjects:
   - kind: User
     name: acs:ram::<ACCOUNT_ID>:user/lisi
     apiGroup: rbac.authorization.k8s.io
   roleRef:
     kind: Role
     name: developer-role
     apiGroup: rbac.authorization.k8s.io
   ```

5. **应用配置**
   ```bash
   kubectl apply -f developer-role.yaml
   kubectl apply -f developer-binding.yaml
   ```

6. **验证权限**
   ```bash
   # 验证是否能查看 Pod
   kubectl auth can-i get pods --namespace=development --as=acs:ram::<ACCOUNT_ID>:user/lisi
   
   # 验证是否能删除 Deployment（应该拒绝）
   kubectl auth can-i delete deployment --namespace=production --as=acs:ram::<ACCOUNT_ID>:user/lisi
   ```

### 6.2 场景：临时授予集群管理员权限

**目标**：运维人员 `admin_user` 需要临时处理集群问题（1小时）

**步骤**：

1. **创建临时 ClusterRoleBinding**
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: temp-admin-binding
   subjects:
   - kind: User
     name: acs:ram::<ACCOUNT_ID>:user/admin_user
   roleRef:
     kind: ClusterRole
     name: cluster-admin
   apiGroup: rbac.authorization.k8s.io
   ```

2. **设置自动清理任务**
   ```bash
   # 使用 kubectl delete --wait=false 立即删除（不等待资源清理完成）
   echo "kubectl delete clusterrolebinding temp-admin-binding" | at now + 1 hour
   ```

3. **记录审计事件**
   - ActionTrail 会记录 ClusterRoleBinding 创建和删除
   - 可通过 SLS 告警监控临时权限未及时回收

## 七、与 OneOps 的设计启示

### 7.1 双层权限架构借鉴

**ACK 模式** → **OneOps 建议**：

```
ACK: RAM（云平台级）+ RBAC（集群级）
OneOps: OneOps RBAC（平台级）+ K8s RBAC（集群级）
```

**设计建议**：

1. **OneOps 平台权限**（类比 ACK RAM）
   - 控制"谁能看到哪些 K8s 集群"
   - 控制"能否获取 KubeConfig"
   - 存储：`users` 表 + `k8s_clusters` 表 + 关联表 `user_cluster_permissions`

2. **K8s 集群权限**（类比 ACK RBAC）
   - 控制"获取 KubeConfig 后能在集群内做什么"
   - 保持 K8s 原生 RBAC，OneOps 仅负责 RoleBinding 创建

### 7.2 RAM 用户映射方案

**ACK 做法**：RoleBinding 的 Subject 使用 `acs:ram::123456:user/zhangsan`

**OneOps 建议**：

| 方案 | Subject 格式 | 优点 | 缺点 |
|------|-------------|------|------|
| **方案 A** | `oneops:user:1` | 简单直观 | 需自定义 Webhook 认证 |
| **方案 B** | `oidc:oneops:user-zhangsan` | 兼容 OIDC | 需搭建 OIDC Provider |
| **方案 C** | 直接使用用户邮箱 | 可读性强 | 需处理邮箱冲突 |

**推荐方案 A**（实现成本最低）：

```go
// OneOps 后端在创建 RoleBinding 时：
subject := rbacv1.Subject{
    Kind:      "User",
    Name:      fmt.Sprintf("oneops:user:%d", userID),
    APIGroup:  "rbac.authorization.k8s.io",
}
```

配套 K8s Webhook 认证：
```go
// 当用户通过 OneOps 访问 K8s API 时：
func authenticateUser(token string) (*UserInfo, error) {
    // 1. 验证 token 是 OneOps 颁发的
    session := validateOneOpsToken(token)
    
    // 2. 返回 K8s 用户格式
    return &UserInfo{
        Username: fmt.Sprintf("oneops:user:%d", session.UserID),
        UID:      session.UserUUID,
        Groups:   []string{"system:authenticated"},
    }, nil
}
```

### 7.3 多租户隔离方案

**ACK 做法**：命名空间隔离 + NetworkPolicy

**OneOps 建议增强**：

1. **自动命名空间创建**
   - 当用户在 OneOps 创建"项目"时，自动在 K8s 创建同名命名空间
   - 自动创建 ResourceQuota 和 LimitRange

2. **自动 RoleBinding 创建**
   - 项目成员加入时，自动创建对应 RoleBinding
   - 项目成员移除时，自动删除 RoleBinding

3. **可视化 NetworkPolicy**
   - OneOps 提供 UI 配置网络策略
   - 默认策略：同项目可互通，跨项目隔离

### 7.4 审计日志设计

**ACK 做法**：ActionTrail + K8s Audit Logs

**OneOps 建议**：

| 日志类型 | 存储位置 | 保留期限 | 查询方式 |
|---------|---------|---------|---------|
| **平台权限操作** | `audit_logs` 表 | 180 天 | UI 查询 |
| **KubeConfig 获取** | `k8s_access_events` 表 | 90 天 | UI 告警 |
| **K8s 集群内操作** | K8s Audit Logs | 30 天（热存储） | ELK/SLS |
| **RBAC 变更** | `k8s_rbac_events` 表 | 180 天 | 审计报告 |

**告警规则建议**：

1. **异常 KubeConfig 获取**：同一用户 1 小时内获取 3 次以上不同集群的凭证
2. **权限提升告警**：非管理员用户创建了 ClusterRoleBinding
3. **跨命名空间访问**：用户访问了无权限的命名空间

## 八、常见问题

### Q1: RAM 用户获取 KubeConfig 后能否绕过 RAM 权限？

**答**：部分可以。RAM 权限只控制"能否获取凭证"，一旦拿到 KubeConfig，后续的 K8s 操作由 RBAC 控制。因此：
- ⚠️ **不要分配** `cs:GetClusterCredentials` 给所有用户
- ✅ **推荐做法**：通过 OneOps 中转，用户不直接持有 KubeConfig

### Q2: 托管版和专有版如何选择？

| 维度 | 托管版 | 专有版 |
|------|--------|--------|
| **权限管理复杂度** | 低（RAM 集成） | 高（需自建） |
| **成本** | 低（省 Master 节点费用） | 高 |
| **控制力** | 中等 | 完全控制 |
| **适用场景** | 中小规模、快速上线 | 大规模、深度定制 |

### Q3: 如何实现"最小权限原则"？

**分层应用最小权限**：

1. **RAM 层**：只授予必要权限（如只读集群、获取凭证）
2. **RBAC 层**：Role 只授予必要资源+动作（如只读 Pod，不能删除 Service）
3. **NetworkPolicy 层**：只允许必要流量（如只允许同命名流量）

### Q4: RBAC 权限变更何时生效？

**答**：立即生效（无缓存）。K8s API Server 实时评估 RBAC，但需注意：
- 客户端缓存 Token（如 kubectl config）
- 旧 RoleBinding 删除后，需重新获取凭证

## 九、推荐阅读

- [阿里云 RAM 权限策略文档](https://help.aliyun.com/document_detail/28647.html)
- [Kubernetes RBAC 官方文档](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [ACK 托管版权限控制](https://help.aliyun.com/document_detail/180983.html)
- [K8s Audit Logs 配置](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/)

## 十、总结

阿里云 ACK 的权限控制采用 **RAM + RBAC 双层架构**：

- **RAM 层**：云平台级权限，控制"能否访问集群"
- **RBAC 层**：集群级权限，控制"能在集群内做什么"

**关键启示**：
1. 双层权限分离是云平台最佳实践（OneOps 可借鉴）
2. 命名空间是 K8s 多租户的基础（应结合 NetworkPolicy）
3. 审计日志是权限安全的关键（应分级存储、设置告警）
4. 托管版大幅降低权限管理复杂度（中小团队首选）

---

**下一步研究建议**：
- 研究其他云厂商的权限实现（AWS EKS、Azure AKS）
- 对比 K8s 多租户增强方案（如 vcluster、kiosk）
- 设计 OneOps 的 K8s 集群权限管理架构
