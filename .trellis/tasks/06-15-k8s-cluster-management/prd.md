# K8s 集群管理功能

## Goal

为 OneOps 平台添加 Kubernetes 集群管理能力，使用户能够通过统一界面管理多个 K8s 集群，执行常见的运维操作。

## What I already know

* 用户请求："开发k8s集群管理功能"
* OneOps 是一个全栈运维管理平台，现有功能包括：
  - 主机资产管理 (CMDB)
  - 堡垒机 (SSH 连接)
  - 监控管理
  - 审计日志
* 技术栈：
  - 后端：Go + Gin + MySQL
  - 前端：Vue 3 + Element Plus (soybean-admin-element-plus)

## Assumptions (temporary)

* 需要支持多个 K8s 集群的连接和管理
* 需要与现有的 OneOps 权限系统集成
* 需要审计日志记录

## Open Questions

* 核心功能范围是什么？
* 如何与现有系统集成？
* 技术实现方式？

## Requirements (evolving)

### 核心功能

**1. 基础集群管理**
- 添加/删除 K8s 集群配置
- 查看集群信息（节点列表、命名空间、版本、状态）
- 集群连接测试和健康检查

**2. 资源管理**（带安全保护）
- 查看资源：Deployment、Pod、Service、ConfigMap、Secret 等
- 更新/删除资源：**需要安全确认机制**
- 资源详情查看（YAML/JSON）
- 实时状态监控

**3. 终端访问**（Pod 级别）
- Pod 管理页面 → 选择 Pod → 点击"进入终端"按钮
- 直接 exec 到该 Pod 的容器中（基于 K8s exec API）
- 支持多容器 Pod（选择进入哪个容器）
- 权限检查：用户需要有该 Pod 所在命名空间的访问权限
- 审计日志：记录终端操作和会话信息
- 会话管理：在线会话列表、强制断开连接

### 前端组织

- **独立顶层模块**：创建"K8s 管理"菜单（类似 CMDB、监控）
- **菜单分组设计**：按照 Kubernetes 资源的功能特性进行逻辑分组
  - **工作负载**：
    - 容器组
    - 无状态
    - 有状态
    - 守护进程集
    - 定时任务
    - 任务
  - **网络**：
    - 服务
    - 路由
  - **配置管理**：
    - 配置项
    - 保密字典
- **其他子菜单**：
  - 集群列表
  - 终端
- 路由前缀：`/k8s/`

**菜单命名规范**：
- 使用中文名称，符合 Kubernetes 资源的行业标准翻译
- Deployments → 无状态（或 工作负载）
- Pods → 容器组
- Services → 服务
- Ingress → 路由
- ConfigMaps → 配置项
- Secrets → 保密字典
- StatefulSets → 有状态
- DaemonSets → 守护进程集
- Jobs → 任务
- CronJobs → 定时任务

### 权限控制（资源级）

- **超出现有 RBAC 范围**：需要设计新的资源级权限系统
- **权限粒度**：
  - 集群级：查看/管理特定集群
  - 资源类型级：查看/管理 Deployment/Pod/Service 等
  - 命名空间级：限制特定命名空间
  - 操作级：只读/更新/删除权限
- **与现有 RBAC 集成**：扩展权限系统，保持向后兼容

**授权操作流程**：

**管理员授权流程**：
1. 登录 OneOps → K8s 管理 → 集群列表
2. 选择目标集群 → 点击"权限管理"
3. 添加用户 → 选择角色（管理员/运维/开发/只读）
4. 确认 → 后台自动创建 K8s RoleBinding

**用户使用流程**：
1. 登录 OneOps → K8s 管理 → 集群列表（只显示有权限的集群）
2. 选择集群 → 查看资源（根据角色显示/隐藏操作按钮）
3. 执行操作 → OneOps 代理转发 → K8s RBAC 验证 → 返回结果

**预设角色映射**：
- 管理员 → `cluster-admin`（完全控制）
- 运维人员 → `admin`（命名空间级管理）
- 开发人员 → `edit`（修改资源，除 RBAC）
- 只读人员 → `view`（只读访问）

### 安全要求（重要）

**危险操作确认机制**（D 方案：二次确认 + 审计日志强化）

**操作分级**：
- 🔴 **高危操作**：删除集群、删除命名空间、删除核心 Deployment
  - 二次确认弹窗：需要输入资源名称确认
  - 审计日志：完整记录操作前后状态
  
- 🟠 **中危操作**：重启 Pod、更新 Deployment、删除 Pod
  - 简单确认弹窗：点击"确定"确认操作
  - 审计日志：记录操作详情
  
- 🟢 **低危操作**：查看资源、查看日志、进入终端
  - 无需确认，直接执行
  - 审计日志：记录访问记录

**审计强化**：
- 所有危险操作在审计日志中标记为"危险操作"
- 记录操作前后状态（如删除前的资源 YAML）
- 可配置告警：某些操作触发钉钉/企业微信通知

**其他安全机制**：
- Kubeconfig 文件加密存储到数据库（参考现有 SSH 凭证加密机制）
- 用户不直接持有 Kubeconfig（OneOps 中转访问）

## Acceptance Criteria

### Phase 1：MVP（基础功能）

**集群管理**：
- [ ] 能够添加 K8s 集群（输入名称、endpoint、上传 kubeconfig）
- [ ] 能够删除集群（需要二次确认）
- [ ] 能够查看集群列表（只显示有权限的集群）
- [ ] 能够查看集群详情（节点列表、命名空间、版本、状态）
- [ ] 集群连接测试和健康检查

**资源查看**：
- [ ] 能够查看 Pod 列表（支持命名空间过滤）
- [ ] 能够查看 Deployment 列表
- [ ] 能够查看 Service 列表
- [ ] 能够查看 ConfigMap/Secret 列表（敏感数据脱敏）
- [ ] 能够查看资源详情（YAML/JSON 格式）
- [ ] 能够查看实时状态（Pod 状态、副本数、镜像版本）
- [ ] **菜单分组实现**：按功能分组显示（工作负载、网络、配置管理）
- [ ] **资源详情页**：Deployment 详情页显示关联的 Pods（使用 Tab 切换）

**权限控制**：
- [ ] 集群级权限检查（无权限用户看不到集群）
- [ ] 权限分配功能（管理员给用户分配集群访问权限）
- [ ] 权限角色映射（管理员、运维、开发、只读）
- [ ] 用户只能看到有权限的集群

**终端访问**：
- [ ] Pod 管理页面可以点击"进入终端"按钮
- [ ] 支持多容器 Pod（选择进入哪个容器）
- [ ] 终端连接建立（基于 K8s exec API）
- [ ] 终端会话管理（查看在线会话、强制断开）
- [ ] 命令审计记录（记录所有执行的命令）
- [ ] WebSocket 连接稳定性（断线重连）

**安全要求**：
- [ ] kubeconfig 加密存储到数据库
- [ ] 用户不直接持有 Kubeconfig（OneOps 中转访问）
- [ ] 危险操作二次确认（删除集群、删除资源）
- [ ] 完整审计日志（权限操作、资源操作、终端操作）
- [ ] 用户身份映射到 K8s User（`oneops:user:<id>`）

**集成要求**：
- [ ] 集成现有 OneOps RBAC 系统
- [ ] 集成现有审计日志系统
- [ ] 集成现有用户管理
- [ ] 复用 WebSocket 处理逻辑
- [ ] 复用在线会话列表组件

### Phase 2：增强功能（后续阶段）

**新增资源类型支持**：
- [ ] StatefulSet（有状态）列表和详情页
- [ ] DaemonSet（守护进程集）列表和详情页
- [ ] Job（任务）列表和详情页
- [ ] CronJob（定时任务）列表和详情页
- [ ] Ingress（路由）列表和详情页

**增强功能**：
- [ ] 资源操作（更新 Deployment、重启 Pod、删除 Pod）
- [ ] 实时日志查看
- [ ] 事件监控查看
- [ ] 多集群统一视图
- [ ] 命名空间级权限可视化配置
- [ ] 资源使用率监控

## Definition of Done (team quality bar)

## Definition of Done (team quality bar)

* 功能代码完成并通过测试
* 前后端集成测试通过
* 文档更新
* 符合 OneOps 编码规范

## Decision (ADR-lite)

### 权限架构决策

**最终选择：方案 C - ACK 式双层架构**

**背景**：需要为 OneOps K8s 管理功能设计资源级权限控制系统

**决策**：采用阿里云 ACK 的双层权限架构模式

**理由**：
1. ✅ 符合主流云平台最佳实践（阿里云 ACK、AWS EKS）
2. ✅ 简化实现：避免维护复杂的权限同步机制
3. ✅ 安全性高：用户不直接持有 Kubeconfig
4. ✅ 审计完整：所有操作都经过 OneOps 便于审计

**架构**：
```
┌─────────────────────────────────────────┐
│       OneOps RBAC（平台级）            │
│  控制"能看到哪些集群"                   │
│  控制"能否访问集群"                     │
│  用户不直接持有 KubeConfig             │
└─────────────────────────────────────────┘
                    │
                    │ OneOps 中转访问
                    ↓
┌─────────────────────────────────────────┐
│        K8s RBAC（集群级）               │
│  使用 oneops:user:123 身份             │
│  控制具体资源操作权限                    │
│  命名空间级权限由 K8s RBAC 处理         │
└─────────────────────────────────────────┘
```

**关键设计点**：
- 权限存储：OneOps 只存储集群级访问权限（`cluster_role_bindings` 表）
- 权限检查：集群级 OneOps 检查，资源级委托 K8s RBAC
- 访问模式：OneOps 中转访问，用户不直接持有 Kubeconfig
- 用户映射：`oneops:user:<user_id>` 作为 K8s User 身份

**权衡**：
- 优势：简化实现、安全性高、符合主流实践
- 劣势：命名空间级权限需要在 K8s 中手动配置 RoleBinding（但符合 K8s 标准做法）

**后果**：
- 需要 OneOps 实现 K8s API 代理层
- 需要实现用户身份映射机制（`oneops:user:<id>`）
- 需要集成 K8s SubjectAccessReview API

## Out of Scope (explicit)

* 待确定...

## Technical Approach

### 数据库设计

**新增表结构**：

**1. k8s_clusters（集群表）**
```sql
CREATE TABLE k8s_clusters (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  endpoint VARCHAR(512) NOT NULL,
  kubeconfig TEXT NOT NULL,             -- 加密存储的 kubeconfig
  cluster_type VARCHAR(50) DEFAULT 'standard',
  region VARCHAR(100),
  version VARCHAR(50),
  node_count INT DEFAULT 0,
  status INT DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**2. cluster_role_bindings（集群权限绑定表）**
```sql
CREATE TABLE cluster_role_bindings (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (cluster_id) REFERENCES k8s_clusters(id),
  FOREIGN KEY (role_id) REFERENCES roles(id),
  UNIQUE KEY uk_user_cluster (user_id, cluster_id)
);
```

**3. k8s_sessions（终端会话表）**
```sql
CREATE TABLE k8s_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  cluster_id BIGINT UNSIGNED NOT NULL,
  pod_name VARCHAR(255) NOT NULL,
  container_name VARCHAR(255),          -- NULL 表示默认容器
  namespace VARCHAR(255) NOT NULL,
  client_ip VARCHAR(50),
  status VARCHAR(20) DEFAULT 'active',
  started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ended_at TIMESTAMP NULL,
  duration INT DEFAULT 0,
  close_reason VARCHAR(200),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (cluster_id) REFERENCES k8s_clusters(id)
);
```

**4. k8s_commands（命令审计表，复用 BastionCommand 逻辑）**
```sql
CREATE TABLE k8s_commands (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT UNSIGNED NOT NULL,
  command TEXT NOT NULL,
  executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  exit_code INT,
  risk_level VARCHAR(20) DEFAULT 'safe',
  output_summary TEXT,
  FOREIGN KEY (session_id) REFERENCES k8s_sessions(id)
);
```

**复用现有机制**：
- ✅ WebSocket 处理逻辑（复用堡垒机 WebSocket）
- ✅ 在线会话列表前端组件
- ✅ 审计日志中间件
- ✅ 命令审计展示页面

### 实施策略（分阶段）

**Phase 1：MVP（基础功能）** - 4-6 周
- 集群管理（添加/删除/查看集群信息）
- 基础权限控制（集群级访问）
- 资源查看（Pod、Deployment、Service、ConfigMap 只读）
- Web 终端连接（kubectl shell）

**Phase 2：增强功能** - 2-3 周
- 资源操作（更新 Deployment、重启 Pod、删除 Pod - 带确认）
- 实时日志查看
- 事件监控查看

**Phase 3：高级功能** - 2-3 周
- 多集群统一视图
- 命名空间级权限可视化配置
- 资源使用率监控

### 后端技术栈
- **K8s 客户端**：`client-go`（官方 Kubernetes Go 客户端）
- **Kubeconfig 加载**：`clientcmd` 包
- **加密存储**：复用现有 `utils.Encryption` 模块

### 前端组织

**现有架构模式**：
- 后端分层：Controller → Service → Model
- API 路由：`/api/<module>/<resource>` （需要 Auth 中间件保护）
- 数据库：MySQL + GORM
- 前端：Vue 3 + Element Plus，按模块组织在 `src/views/`
- 权限：基于 RBAC 的菜单级权限控制
- 审计：全局操作日志中间件自动记录

**相关模块**：
- CMDB 模块 (`backend/cmdb`, `src/views/cmdb/`) - 可参考的资产管理模式
- 监控模块 (`backend/monitoring`, `src/views/monitoring/`) - 可参考的监控集成模式
