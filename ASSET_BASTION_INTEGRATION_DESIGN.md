# 资产管理与堡垒机能力集成设计

## 背景

OneOps 当前已经具备 CMDB 资产管理的基础能力，包括主机资产、主机分组、业务系统、机房机柜、标签、SSH 凭证和资产变更记录。与此同时，主机资产天然会延伸出“连接主机、控制权限、审计操作”的堡垒机场景。

因此，堡垒机功能不建议作为一个完全独立、割裂的一级模块，而应作为“资产访问能力”集成到资产管理中。用户的操作路径应当是：先找到资产，再连接资产，再审计资产访问。

## 设计结论

建议采用以下定位：

```text
资产管理 = 资产台账 + 访问入口 + 访问审计
堡垒机 = 资产访问能力，不单独成为割裂系统
```

这样做的好处是：

- 用户心智更简单：围绕“资产”完成查看、连接、授权和审计。
- 能复用现有 CMDB 数据：主机、分组、业务、标签、凭证、变更记录都可以直接作为堡垒机上下文。
- 权限模型更自然：可以按主机、分组、业务系统、标签授予访问权限。
- 后续扩展更清晰：SSH、SFTP、RDP、Kubernetes、数据库代理都可以作为资产访问协议逐步扩展。

## 推荐信息架构

```text
资产管理
├─ 资产总览
├─ 主机资产
│  ├─ 主机列表
│  ├─ 主机详情
│  ├─ 分组/标签/业务归属
│  ├─ 分组树管理
│  └─ 连接入口
├─ 资产组织
│  ├─ 业务系统
│  ├─ 机房机柜
│  └─ 标签管理
├─ 访问管理
│  ├─ SSH 凭证库
│  ├─ 授权策略
│  ├─ 连接账号映射
│  └─ 连通性检测
├─ 会话审计
│  ├─ 在线会话
│  ├─ 历史会话
│  ├─ 命令审计
│  └─ 文件传输记录
└─ 资产变更
```

菜单层级上，建议仍保留“资产管理”作为一级目录。主机分组不再作为独立菜单，而是作为“主机资产”页面内的资产树、筛选维度和分组管理能力。堡垒机相关能力放在“访问管理”和“会话审计”下，而不是新增一个独立的“堡垒机”一级菜单。

## 主机资产页设计

当前 `cmdb_servers` 页面可以继续作为主机资产的主入口。建议保持左侧资产树、右侧主机列表的基本布局，在主机列表操作列增加连接能力。

独立的 `cmdb_groups` 页面不再保留。分组新增、编辑、删除、主机挂载分组等能力统一收敛到 `cmdb_servers` 页面，避免用户在“主机资产”和“主机分组”两个菜单之间来回切换。

推荐操作：

```text
查看 | 编辑 | 连接 | 更多
```

“连接”按钮点击后打开连接面板：

```text
连接主机：web-01
连接方式：SSH 终端 / SFTP / 复制连接命令
登录账号：root / appuser / readonly
凭证来源：绑定凭证 / 临时授权 / 手动输入
有效策略：生产环境需审批 / 非生产可直连
```

异常状态要直接反馈给用户：

```text
无连接权限
未绑定 SSH 凭证
凭证不可用
主机离线
当前时间段不可连接
生产环境连接需审批
```

## 主机详情页设计

主机详情建议使用抽屉或独立详情页，避免用户离开资产上下文。详情页可以按 Tab 组织：

```text
基础信息
连接信息
分组与标签
凭证绑定
会话记录
变更记录
监控状态
```

其中：

- 基础信息：主机名、IP、系统、CPU、内存、磁盘、环境、状态等。
- 连接信息：SSH 端口、登录账号、凭证绑定状态、最近连通性检测结果。
- 凭证绑定：选择 SSH 凭证、测试连接、查看凭证可用状态，但不展示敏感内容。
- 会话记录：查看该主机的连接历史、操作人、时间、持续时长、命令记录。
- 变更记录：复用现有资产变更记录。

## 技术选型总览

整体技术路线应尽量贴合 OneOps 当前架构，不为 P0 引入过重的平台级组件。

```text
前端：Vue 3 + Element Plus + xterm.js
后端：Go + Gin + WebSocket + golang.org/x/crypto/ssh
数据库：MySQL
权限：现有 JWT + RBAC + 权限点，复杂策略阶段再评估 Casbin
审计：MySQL 记录会话、命令、文件传输和审批记录
```

Web SSH 的基础链路：

```text
浏览器 xterm.js
  ↓ WebSocket
Go 后端连接代理
  ↓ SSH
目标主机
```

前端只负责终端展示、键盘输入和屏幕输出，不接触 SSH 密码、私钥、临时凭证等敏感信息。真正的 SSH 连接、凭证读取、权限校验和审计写入都在 Go 后端完成。

## 功能分期

### P0：最小可用版本

P0 目标是让用户能安全地从资产管理里连接主机，并留下基础审计记录。

- 主机绑定 SSH 凭证。
- 主机列表支持一键 Web SSH。
- 连接前进行权限校验。
- 记录会话：用户、主机、登录账号、客户端 IP、开始时间、结束时间、持续时长、状态。
- 记录命令：命令内容、执行时间、退出状态、是否被拦截。
- 前端不接触明文密码或私钥。
- 凭证只在后端连接代理中使用。

P0 技术栈：

| 层级 | 技术选择 | 说明 |
| --- | --- | --- |
| 前端终端 | `xterm.js`、`@xterm/addon-fit`、`@xterm/addon-web-links` | 用于 Web SSH 终端显示、自适应尺寸和链接识别。 |
| 前端框架 | 现有 `Vue 3`、`Element Plus`、Pinia | 复用现有前端技术栈，不新增 UI 框架。 |
| 后端框架 | 现有 Go + Gin | 复用当前后端服务，不单独引入 Node.js 终端代理。 |
| WebSocket | `github.com/coder/websocket` | 用于浏览器和后端连接代理之间的双向流转发。 |
| SSH 客户端 | `golang.org/x/crypto/ssh` | 后端通过 SSH 协议连接目标主机。 |
| 凭证加密 | Go 标准库 `crypto/aes` + GCM 模式 | 凭证先加密存储在 MySQL，后续可迁移到 Vault/KMS。 |
| 数据存储 | MySQL | 复用现有数据库，新增会话和命令审计表。 |
| 权限校验 | 现有 JWT + RBAC + `cmdb:server:connect` 权限点 | P0 不引入 Casbin，只校验用户是否具备连接主机的操作权限。 |
| 日志审计 | 现有 Zap 日志 + MySQL 审计表 | 记录连接会话、命令输入和连接异常。 |

### P1：增强版本

P1 目标是让堡垒机能力更接近生产可用。

- 在线会话管理，管理员可强制断开。
- SFTP 文件上传下载审计。
- 生产环境连接审批。
- 高危命令拦截，例如 `rm -rf /`、`mkfs`、`shutdown` 等。
- 会话回放。
- 按分组、业务系统、标签批量授权。

P1 技术栈：

| 层级 | 技术选择 | 说明 |
| --- | --- | --- |
| 前端终端 | 继续使用 `xterm.js` | 增加会话状态、只读回放、断开提示等交互。 |
| 文件传输 | `github.com/pkg/sftp` + 后端流式传输接口 | 用于 SFTP 上传下载，并写入文件传输审计。 |
| 在线会话 | Go 内存会话管理 + MySQL 会话状态 | 管理在线会话、断开会话、会话心跳。 |
| 访问策略 | MySQL `asset_access_policies` 表 + Go 策略服务 | 先用自研策略表支持分组、标签、业务系统、时间窗口和登录账号限制。 |
| 权限引擎 | 默认继续使用现有 RBAC；策略复杂后评估 Casbin | 如果只是权限点和资产范围校验，不需要 Casbin；如果策略表达越来越复杂，再引入。 |
| 审批流 | MySQL `bastion_approvals` 表 + 现有用户/角色体系 | 支持生产环境连接审批和临时授权。 |
| 命令风控 | Go 命令解析与规则匹配 | 支持高危命令识别、拦截和审计。 |

### P2：未来扩展

P2 目标是扩展资产访问协议和复杂场景。

- RDP / VNC。
- Kubernetes exec。
- 数据库登录代理。
- 多人协同会话。
- 外部堡垒机系统对接，例如 JumpServer。
- 零信任访问策略和临时凭证。

P2 技术栈：

| 层级 | 技术选择 | 说明 |
| --- | --- | --- |
| 复杂权限 | Casbin 或独立策略服务 | 当需要 ABAC、资源属性、环境、时间、审批状态组合判断时再引入。 |
| 远程桌面 | Apache Guacamole 或独立协议网关 | 用于 RDP/VNC 等图形协议，避免在现有 Go 服务里硬实现远程桌面协议。 |
| Kubernetes 连接 | Kubernetes client-go + WebSocket/stream | 支持 Pod exec、日志查看等容器场景。 |
| 数据库代理 | 独立数据库连接代理服务 | 支持 MySQL/PostgreSQL 等数据库登录审计，和 SSH 堡垒机解耦。 |
| 凭证系统 | Vault、云 KMS 或企业密钥管理系统 | 当凭证规模和合规要求提高时，从 MySQL 加密存储迁移出去。 |
| 会话回放存储 | 对象存储或归档存储 | 大规模保存终端录屏、回放事件流和文件传输证据。 |
| 外部集成 | JumpServer API、LDAP/AD、企业审批系统 | 支持和既有身份源、审批系统、堡垒机系统集成。 |

## 权限设计

连接权限不应复用普通资产编辑权限，需要单独拆分。

建议权限点：

```text
cmdb:server:query
cmdb:server:create
cmdb:server:update
cmdb:server:delete
cmdb:server:connect

cmdb:credential:query
cmdb:credential:create
cmdb:credential:update
cmdb:credential:delete
cmdb:credential:test

cmdb:access-policy:query
cmdb:access-policy:create
cmdb:access-policy:update
cmdb:access-policy:delete

cmdb:session:query
cmdb:session:terminate
cmdb:session:replay

cmdb:command:query
cmdb:file-transfer:query
```

授权策略建议支持以下维度：

- 授权对象：用户、角色、用户组。
- 资产范围：单台主机、主机分组、业务系统、标签。
- 登录账号：允许使用哪些系统账号。
- 连接协议：SSH、SFTP，未来可扩展 RDP、VNC 等。
- 时间窗口：允许连接的时间段。
- 审批要求：是否需要审批。
- 操作限制：是否允许文件传输、是否允许 sudo、是否启用高危命令拦截。

### 是否需要 Casbin

P0 不需要 Casbin。`cmdb:server:connect` 这类权限点本质是判断用户是否拥有某个操作许可，现有 RBAC 就能覆盖。

```text
用户 → 角色 → 权限点 → 是否允许连接
```

例如：

```text
cmdb:server:query    可以查看主机
cmdb:server:connect  可以连接主机
```

P1 可以先用自研访问策略表实现资源范围校验：

```text
用户/角色 + 主机/分组/标签/业务 + 登录账号 + 时间窗口 + 审批状态
```

当策略条件变得更复杂，例如需要 ABAC、资源属性判断、优先级策略、拒绝优先、多租户隔离等，再引入 Casbin。这样可以避免 P0 阶段过早增加权限模型和调试成本。

## 后端模型建议

当前已有核心表：

- `servers`：主机资产。
- `ssh_credentials`：SSH 凭证。
- `server_groups`：主机分组。
- `server_tags`：主机标签。
- `asset_changes`：资产变更记录。

建议新增以下表。

### 访问策略表

```text
asset_access_policies
- id
- name
- subject_type        用户/角色/用户组
- subject_id
- asset_scope_type    主机/分组/业务/标签
- asset_scope_id
- login_accounts
- protocols
- allow_file_transfer
- allow_sudo
- require_approval
- time_window
- status
- created_at
- updated_at
```

### 会话表

```text
bastion_sessions
- id
- server_id
- user_id
- username
- login_account
- client_ip
- protocol
- started_at
- ended_at
- duration
- status
- close_reason
- created_at
```

### 命令审计表

```text
bastion_commands
- id
- session_id
- command
- executed_at
- exit_code
- risk_level
- blocked
- output_summary
```

### 文件传输审计表

```text
bastion_file_transfers
- id
- session_id
- direction
- path
- size
- status
- created_at
```

### 审批表

```text
bastion_approvals
- id
- requester_id
- approver_id
- server_id
- login_account
- reason
- status
- requested_at
- approved_at
- expired_at
```

## 后端接口建议

### 主机连接

```text
POST /api/cmdb/servers/:id/connect
```

用途：创建 Web SSH 会话。

前端只提交连接意图，不提交敏感凭证。

```json
{
  "protocol": "ssh",
  "loginAccount": "root"
}
```

后端返回会话信息和 WebSocket 地址。

```json
{
  "sessionId": 10001,
  "websocketUrl": "/api/bastion/sessions/10001/ws"
}
```

### 会话管理

```text
GET    /api/cmdb/sessions
GET    /api/cmdb/sessions/:id
POST   /api/cmdb/sessions/:id/terminate
GET    /api/cmdb/sessions/:id/commands
GET    /api/cmdb/sessions/:id/file-transfers
```

### 访问策略

```text
GET    /api/cmdb/access-policies
POST   /api/cmdb/access-policies
PUT    /api/cmdb/access-policies/:id
DELETE /api/cmdb/access-policies/:id
```

## 前端页面建议

建议第一阶段新增或调整以下页面：

```text
/cmdb/servers              主机资产
/cmdb/access-policies      访问策略
/cmdb/sessions             会话审计
/cmdb/commands             命令审计
/cmdb/ssh-credentials      凭证库
```

其中 `/cmdb/ssh-credentials` 的菜单名建议从“SSH 凭证”调整为“凭证库”，未来可以容纳密码、私钥、云厂商密钥等类型。

`/cmdb/servers` 是核心入口，需要补充：

- 主机连接按钮。
- 主机详情抽屉。
- 连接弹窗。
- 凭证绑定状态。
- 最近会话记录。
- 权限不足、凭证缺失、连接失败等明确状态提示。

## 安全设计原则

- 前端永远不获取明文密码和私钥。
- 敏感凭证只在后端连接代理中使用。
- 每次连接必须生成会话 ID。
- 所有连接、命令、文件传输都要关联用户、主机和会话。
- 生产环境主机默认更严格，可以要求审批或二次确认。
- 高危命令需要支持拦截和审计。
- 会话日志与操作日志要能互相追溯。

## 与现有能力的关系

现有 `servers` 表继续作为资产主表，不需要为堡垒机重复维护一套主机资产。

现有 `ssh_credentials` 继续作为凭证库，但需要补强：

- 凭证加密存储。
- 凭证可用性检测。
- 凭证使用审计。
- 凭证权限范围。

现有 `asset_changes` 继续记录资产字段变更；堡垒机会话和命令审计单独建表，避免和资产变更混在一起。

## 推荐落地路径

### 第一阶段：连接闭环

- 主机资产页增加连接按钮。
- 后端实现 SSH WebSocket 代理。
- 使用主机绑定的 SSH 凭证建立连接。
- 写入会话记录和命令审计。
- 增加 `cmdb:server:connect` 权限。

### 第二阶段：策略闭环

- 增加访问策略页面。
- 支持按用户、角色、分组、标签授权。
- 支持连接时间窗口。
- 支持生产环境审批。

### 第三阶段：审计闭环

- 增加在线会话列表。
- 增加命令检索。
- 增加文件传输记录。
- 增加会话回放。

## 最终建议

堡垒机能力应该集成到资产管理中，但要保持清晰边界：

- 资产管理负责“资产是谁、在哪里、属于谁、怎么连接”。
- 访问管理负责“谁可以连、什么时候连、用什么账号连”。
- 会话审计负责“谁连过、做了什么、是否合规”。

这种设计既贴合 OneOps 当前 CMDB 架构，也能为后续堡垒机能力扩展留下空间。
