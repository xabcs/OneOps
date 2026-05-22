# 资产管理与堡垒机能力集成设计

## 背景与结论

OneOps 当前具备 CMDB 资产管理基础能力（主机、分组、业务系统、机房机柜、标签、SSH 凭证、资产变更）。主机资产天然延伸出"连接、授权、审计"的堡垒机场景。

**结论：堡垒机功能作为"资产访问能力"集成到资产管理中，不单独成为割裂系统。**

```
资产管理 = 资产台账 + 访问入口 + 访问审计
堡垒机   = 资产访问能力，不单独成为一级模块
```

**与现有能力的关系：**
- `servers` 表继续作为资产主表，无需为堡垒机重复建主机表
- `ssh_credentials` 继续作为凭证库，补强加密存储和可用性检测
- `asset_changes` 继续记录资产字段变更；会话、命令、文件传输另建专用审计表

---

## 产品与工程原则

- **工作流优先**：围绕"找资产 → 看资产 → 连资产 → 查审计"设计页面和接口
- **渐进交付**：每个阶段都必须交付前后端联调完成、可演示、可被真实用户试用的闭环系统；阶段之间只做能力增强
- **前后端同阶段**：每个阶段同时定义前端目标、后端目标、联调目标和验收标准
- **安全默认**：前端不接触明文凭证，后端统一做鉴权、凭证解密、连接代理和审计写入
- **避免过早复杂化**：不为早期阶段引入 Casbin、RDP、K8s、审批流等重量级组件

---

## 整体架构

### 目标菜单结构

```
资产管理
├─ 资产总览              /cmdb/dashboard               阶段二
├─ 主机资产              /cmdb/servers                 阶段一
├─ 访问控制              /cmdb/access
│  ├─ 访问策略           /cmdb/access/policies         阶段二
│  └─ 凭证库             /cmdb/access/credentials      阶段一
├─ 会话审计              /cmdb/audit
│  ├─ 在线会话           /cmdb/audit/online            阶段三
│  ├─ 历史会话           /cmdb/audit/sessions          阶段一
│  ├─ 命令审计           /cmdb/audit/commands          阶段一
│  └─ 文件传输           /cmdb/audit/files             阶段三
├─ 资产配置              /cmdb/config
│  ├─ 业务系统           /cmdb/config/business         阶段一
│  ├─ 机房机柜           /cmdb/config/rooms            阶段一
│  ├─ 标签管理           /cmdb/config/tags             阶段一
│  └─ Agent 管理         /cmdb/config/agents           阶段二
└─ 资产变更              /cmdb/changes                 已有

隐藏路由（hideInMenu）
└─ SSH 终端             /cmdb/terminal/:sessionId      阶段一（全屏终端页）
```

菜单原则：
- "主机分组"不再作为独立菜单，分组能力收敛到"主机资产"页左侧资产树
- "凭证库"放在"访问控制"下，强调凭证是连接能力的一部分
- "业务系统、机房机柜、标签"归入"资产配置"，避免与日常主机操作混在一起

### 资产组织关系

资产配置不是孤立的字典页面，这三类数据最终都作为主机的组织、定位和授权维度：

| 维度 | 表达含义 | 在堡垒机中的作用 |
|------|----------|-----------------|
| 业务系统 | 这台主机服务于哪个业务/项目 | 按业务系统授权；审计时反查业务影响 |
| 机房机柜 | 这台主机物理位置在哪 | 按机房限制物理机访问；故障时快速定位影响范围 |
| 标签 | 主机的灵活特征（生产/核心链路/公网暴露） | 按标签设置差异化策略；审计高风险资产操作 |

在主机新增/编辑表单中提供业务系统、机房机柜、标签的快捷创建入口，避免用户为补一个标签离开主机表单。

### 技术选型

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端终端 | xterm.js + @xterm/addon-fit | Web SSH 终端显示与自适应尺寸 |
| 前端框架 | Vue 3 + Element Plus + Pinia | 复用现有技术栈 |
| 后端框架 | Go + Gin | 复用现有后端 |
| WebSocket | github.com/coder/websocket | 浏览器与后端双向流转发 |
| SSH 客户端 | golang.org/x/crypto/ssh | 后端建立 SSH 连接 |
| 凭证加密 | crypto/aes + GCM | 凭证加密存储在 MySQL，后续可迁移到 Vault/KMS |
| 文件传输 | github.com/pkg/sftp | 阶段三引入 |
| 数据存储 | MySQL | 复用现有数据库，新增审计和策略表 |

---

## 权限设计

### 两层权限模型

**第一层：功能权限（菜单 RBAC）**

控制用户能看到哪些页面入口，通过角色绑定菜单 ID 实现，现有机制已够用：

```
超级管理员  → 全部菜单
运维工程师  → 资产管理 + 会话审计
审计员      → 只读：会话审计、命令审计
普通用户    → 首页
```

**第二层：连接权限（访问策略表）**

控制用户能 SSH 进哪台服务器，通过 `asset_access_policies` 表按角色 + 资产范围匹配：

- 非生产服务器：有 CMDB 菜单权限的用户默认可连
- 生产服务器：必须匹配访问策略；管理员直接放行
- 临时授权：管理员设置策略的 `time_window` 字段，到期自动失效（替代审批流）

前端"连接"按钮的状态由后端 `CheckConnectPermission` 接口返回值驱动，不由前端自行判断权限字符串。

**审计作为核心安全机制**

内部运维平台的安全依赖"知道自己被审计"的约束，每次连接、每条命令全部关联用户 + 主机 + 会话 ID，出问题可以完整追溯。

### 暂不规划（后期 P2+）

- **审批流**（`bastion_approvals` 表）：访问策略 `time_window` 已能替代临时授权场景，审批流开发成本高；触发时机：团队规模扩大、有合规审批链路要求时
- **权限点体系**（`cmdb:server:connect` 等）：需要后端 buttons 数据源 + 前端 hasAuth() + 后端接口校验同时配合，当前未实现；触发时机：平台用户超 500 人、有等保/SOC2 合规要求时
- **Casbin**：策略条件变复杂（ABAC、多租户隔离）时再评估引入

---

## 数据模型

### 会话表（阶段一建表）

```
bastion_sessions
- id
- server_id / user_id / username / login_account
- client_ip / protocol
- started_at / ended_at / duration
- status          active | closed | terminated | error
- close_reason
- created_at
```

### 命令审计表（阶段一建表）

```
bastion_commands
- id / session_id
- command / executed_at / exit_code
- risk_level      low | medium | high
- blocked         是否被拦截
- output_summary
```

### 访问策略表（阶段二建表）

```
asset_access_policies
- id / name
- subject_type    user | role
- subject_id
- asset_scope_type  server | group | business | tag
- asset_scope_id
- login_accounts  允许的登录账号（JSON 数组）
- protocols       ssh | sftp
- allow_file_transfer / allow_sudo
- require_approval  暂不使用，保留字段，后期规划
- time_window     允许连接的时间段（JSON）
- status / created_at / updated_at
```

### 文件传输审计表（阶段三建表）

```
bastion_file_transfers
- id / session_id
- direction       upload | download
- path / size / status
- created_at
```

### ~~审批表~~（暂不规划，后期 P2+）

```
bastion_approvals — 暂不建表，用访问策略 time_window 代替临时授权
```

---

## 凭证分层设计

### 问题背景

现有 `ssh_credentials` 表被两类操作混用：

| 操作类型 | 执行者 | 典型权限需求 |
|----------|--------|-------------|
| **用户堡垒连接** | 人工，通过 Web Terminal | 受限账号（`ops`/`deploy`），禁止 rm -rf 等高危命令，受访问策略约束 |
| **系统自动化运维** | OneOps 后端，无人工干预 | 高权限账号（`root` 或具备 sudo），需要安装软件、写 systemd 配置、重启服务 |

混用同一凭证会导致两难：
- 用凭证满足系统操作 → 用户堡垒连接时账号权限过大，增加审计风险
- 用用户凭证做系统操作 → 权限不够，Agent 部署失败（无法 `systemctl enable`、无法写 `/etc/systemd/`）

**当前 Agent 部署失败的根本原因之一：** `loadServerWithCredential` 拿用户绑定凭证（可能是 `ops` 用户），SSH 过去执行 `systemctl enable oneops-agent` 失败，但错误被静默丢弃。

### 设计方案

**在现有 `SSHCredential` 表增加 `credential_type` 字段**（扩展现有表，不新建表），并在 `Server` 表增加 `system_credential_id` 字段：

```
ssh_credentials 表新增字段：
  credential_type  ENUM('user', 'system')  DEFAULT 'user'
    - user：用于用户堡垒连接，在连接弹窗凭证选择器中可见
    - system：用于系统自动化运维，在连接弹窗中隐藏，仅用于 Agent 部署/重启/卸载及 SSH 采集

servers 表新增字段：
  system_credential_id  BIGINT UNSIGNED  NULL（外键指向 ssh_credentials.id）
    - 专供 Agent 部署、systemctl 操作、SSH 指标采集使用
    - 为空时 Agent 部署直接返回错误，不 fallback 到用户凭证
```

### 两类凭证的使用边界

```
                    ┌────────────────────────────────────────┐
                    │          ssh_credentials               │
                    │  credential_type = 'user'              │
                    │  典型：ops / deploy / 业务账号         │
                    └──────────────┬─────────────────────────┘
                                   │  用于堡垒连接
                     ┌─────────────▼──────────────┐
                     │  bastion_sessions（用户会话） │
                     │  CheckConnectPermission 校验 │
                     │  全程审计命令 + 会话           │
                     └────────────────────────────┘

                    ┌────────────────────────────────────────┐
                    │          ssh_credentials               │
                    │  credential_type = 'system'            │
                    │  典型：root / ansible / oneops-agent   │
                    └──────────────┬─────────────────────────┘
                                   │  用于系统操作
                     ┌─────────────▼──────────────┐
                     │  Agent 部署 / 重启 / 卸载    │
                     │  SSH 指标采集                 │
                     │  硬件配置采集                 │
                     └────────────────────────────┘
```

系统凭证**永远不会出现在**以下场景：
- 连接弹窗的凭证选择列表
- `CheckConnectPermission` 返回的可用凭证
- `bastion_sessions` 会话记录
- 用户视角的凭证管理（`/cmdb/access/credentials`）

用户凭证**永远不会被用于**以下场景：
- Agent 部署 / 重启 / 卸载
- SSH 指标采集
- 任何由 OneOps 后端自动发起的 SSH 操作

### 数据模型变更

**`SSHCredential` 新增字段：**

```go
// CredentialType 凭证用途类型
type CredentialType string

const (
    CredentialTypeUser   CredentialType = "user"   // 用于用户堡垒连接
    CredentialTypeSystem CredentialType = "system" // 用于系统自动化运维
)

type SSHCredential struct {
    // ...现有字段...
    CredentialType CredentialType `json:"credentialType" gorm:"type:enum('user','system');default:'user'"`
}
```

**`Server` 新增字段：**

```go
type Server struct {
    // ...现有字段...
    SystemCredentialID uint           `json:"systemCredentialId" gorm:"index"`
    SystemCredential   *SSHCredential `json:"systemCredential,omitempty" gorm:"foreignKey:SystemCredentialID;constraint:OnDelete:SET NULL"`
}
```

### 后端逻辑变更

**`AgentService.loadServerWithCredential` 重命名为 `loadServerForAgent`，强制使用系统凭证：**

```
loadServerForAgent(serverID) 查找逻辑：
  1. server.SystemCredentialID → db.First(&cred, server.SystemCredentialID)
  2. 为空 → 直接返回错误"主机未配置系统运维凭证，请在主机编辑页绑定 credential_type=system 的凭证"
     （不 fallback 到用户凭证，两类凭证完全隔离）
```

**`dialSSH` 函数**无需改动，入参 `*models.Server` 已包含解析出的凭证。

**`bastion.go` 的 `CheckConnectPermission`**：查询 `server.Credentials` 时过滤 `credential_type = 'user'`，确保系统凭证不出现在连接弹窗。

### API 变更

```
# 凭证 CRUD 接口新增 type 筛选参数
GET /api/cmdb/credentials?type=user      仅返回用户凭证（连接弹窗使用）
GET /api/cmdb/credentials?type=system    仅返回系统凭证（主机编辑绑定使用）
GET /api/cmdb/credentials               返回全部（凭证库管理页使用）

# 创建/编辑凭证时新增 credentialType 字段
POST /api/cmdb/credentials              body 新增 credentialType: 'user'|'system'
PUT  /api/cmdb/credentials/:id          body 新增 credentialType: 'user'|'system'

# 主机接口无变化，systemCredentialId 随 Server CRUD 正常读写
```

### 前端页面设计

#### 凭证库页面（/cmdb/access/credentials）

**改动：顶部增加凭证类型切换 Tab**

```
[全部凭证] [用户凭证] [系统凭证]

表格列：凭证名称 | 类型标签 | 认证方式 | 用户名 | 已绑定主机数 | 创建时间 | 操作

类型标签：
  - 用户凭证：蓝色 tag "用户连接"
  - 系统凭证：橙色 tag "系统运维"
```

**新增/编辑凭证弹窗**：增加「凭证用途」单选：
```
凭证用途 *
  ● 用户连接    → credential_type = 'user'
               （用于用户堡垒 SSH，受访问策略约束）
  ○ 系统运维    → credential_type = 'system'
               （仅供 OneOps 后端 Agent 部署、采集使用，不出现在用户连接列表）
```

#### 主机编辑弹窗（ServerEditDialog）

**改动：凭证部分拆分为两个字段**

```
用户连接凭证    [下拉选择 type=user 的凭证 ▼]
               ℹ 用于用户通过堡垒机建立 SSH 会话，可绑定多个

系统运维凭证    [下拉选择 type=system 的凭证 ▼]
               ℹ 用于 Agent 部署、systemctl 操作、指标采集，需具备 root 或 sudo 权限
               △ 未配置时 Agent 部署将使用用户连接凭证（可能失败）
```

后端：主机 `PATCH /api/cmdb/servers/:id` 接受 `systemCredentialId` 字段更新。

#### Agent 管理页（/cmdb/config/agents）

**改动：表格新增"系统凭证"列**

```
主机名 | IP | Agent状态 | 版本 | 系统凭证 | 最近心跳 | 操作

系统凭证列显示：
  - 已配置：绿色文字 "root@cred-name"（凭证名称+用户名）
  - 未配置：红色警告 "✗ 未配置"，鼠标 hover 显示"Agent 部署需要系统运维凭证，请先在主机编辑页绑定"
```

批量部署前置校验：若勾选主机中有未配置系统凭证的主机，**阻断部署并提示**：

```
以下 N 台主机未配置系统运维凭证，无法部署：
  - server-01 (192.168.1.1)
  - server-02 (192.168.1.2)

请先在主机编辑页为这些主机绑定 credential_type=system 的凭证，再执行部署。

[关闭]   [去配置 →]
```

（不提供"仍然部署"选项，两类凭证完全隔离，无 fallback 路径）

### 实现顺序

1. **数据模型**：`SSHCredential` 加 `credential_type`，`Server` 加 `system_credential_id`，`AutoMigrate` 自动建列
2. **后端逻辑**：`AgentService` 的凭证加载逻辑修改，`CheckConnectPermission` 过滤系统凭证
3. **凭证 CRUD**：接口支持 `type` 筛选和 `credentialType` 字段读写
4. **前端凭证库**：增加类型 Tab 和创建/编辑弹窗的凭证用途字段
5. **前端主机编辑**：拆分凭证字段
6. **前端 Agent 管理**：增加系统凭证列和批量部署预检

---

## 安全设计原则

- 前端永远不获取明文密码和私钥，凭证只在后端连接代理中读取和使用
- 每次连接必须生成会话 ID，与用户、主机、登录账号绑定
- 所有连接、命令、文件传输都要写入审计，关联用户 + 主机 + 会话 ID
- 生产环境主机默认更严格，必须匹配访问策略才能连接；临时授权通过 `time_window` 实现
- 高危命令支持拦截和告警
- 凭证加密存储（AES-GCM），后续可迁移到 Vault/KMS

---

## 阶段一：资产整理 + SSH 连接闭环

**目标**：整理资产管理菜单和信息架构，完成从主机资产页发起 SSH 连接、到全屏终端可用、到会话和命令可查询的完整闭环。

**可运行范围**：
- 用户进入资产管理，完成主机 CRUD、分组筛选、凭证绑定等现有能力（菜单结构已调整）
- 用户在主机列表点击"连接"，通过后端 SSH 代理打开全屏 Web SSH 终端
- 连接完成后，系统记录会话和命令，可在审计页面查询

### 页面设计

#### 主机资产页（/cmdb/servers）

主机资产和堡垒机能力融合的核心页面，左侧资产树 + 右侧主机工作区。

```
左侧：资产树
├─ 全部主机
├─ 生产 / 测试 / 开发
├─ 按业务系统
├─ 按主机分组
└─ 按标签

右侧：主机工作区
├─ 筛选栏：主机名/IP、环境、状态、业务系统、标签、凭证状态
├─ 主机表格
│  ├─ 主机名（点击打开详情抽屉）
│  ├─ 连接 IP / 环境 / 操作系统
│  ├─ 硬件配置（CPU核 / 内存GB / 磁盘GB）
│  ├─ 状态 / 凭证状态 / 最近连接时间
│  └─ 操作：[连接] [详情] [编辑] [更多▼ → 测试连接 / 删除]
└─ 分页
```

**"连接"按钮状态**（由后端 `CheckConnectPermission` 返回值驱动）：

| 状态 | 按钮表现 | 说明 |
|------|----------|------|
| 可连接 | 主按钮（蓝色） | 有可用凭证且策略允许直连 |
| 无权限 | 禁用 + Tooltip | 策略不允许，显示后端返回的拒绝原因 |
| 无凭证 | 警告按钮（橙色） | 点击后引导绑定凭证 |
| 离线/异常 | 禁用 | 提示最近连通检测失败 |
| ~~需审批~~ | ~~次按钮~~ | **暂不规划**，后期 P2+ |

**资产树交互**：
- 单击节点过滤右侧主机列表
- 右键分组节点：新增子分组、重命名、删除
- 主机分组管理能力收敛到本页，不再保留独立"主机分组"菜单

**硬件配置采集**：主机新增/编辑保存后，后端异步通过 SSH 采集 CPU 核数、内存容量、磁盘容量，写入 `servers.cpu`/`memory`/`disk`。采集失败不影响保存，仅记录告警日志。

#### 主机详情抽屉

触发方式：点击主机名或"详情"按钮，右侧抽屉宽 720~900px，不离开主机资产页。

```
头部：主机名 | 环境标签 | 在线/离线状态 | [连接] [编辑]

Tab 内容
├─ 概览：主机名、IP、内网IP、系统、CPU/内存/磁盘配置、环境、业务系统、机房机柜
├─ 连接：SSH 端口、可用登录账号、绑定凭证状态、最近连通检测时间
├─ 分组与标签：所属分组、标签、业务归属（可直接编辑）
├─ 会话记录：该主机最近10条会话，支持跳转历史会话页
└─ 变更记录：复用资产变更数据
```

#### 连接弹窗

触发方式：主机表格"连接"按钮或详情抽屉"连接"按钮。

```
标题：连接主机 {hostname}

左侧（连接配置）              右侧（安全上下文）
├─ 协议：SSH / SFTP          ├─ 主机环境（生产醒目标红）
├─ 登录账号（来自后端）       ├─ 命中策略（无策略显示"默认允许直连"）
└─ 连接原因（生产环境必填）   ├─ 最近连接人（最近一次会话用户）
                              └─ 审计提示

底部：[取消]  [连接]
（~~[申请访问]~~ 暂不规划）
```

无凭证时：底部变为 [取消] [去绑定凭证]，引导用户进入凭证库。

#### Web SSH 终端页（/cmdb/terminal/:sessionId）

全屏独立路由（hideInMenu），不内嵌在弹窗中。

```
顶部会话栏（48px，深色背景）
├─ [SSH] 协议标签
├─ 主机名 / IP / 登录账号
├─ 会话时长（倒计时格式 HH:MM:SS）
└─ [断开连接]（二次确认）

主体：xterm.js 全屏终端
```

关键交互：
- 窗口 resize 时同步发送 resize 消息到后端
- 断线时显示重连提示条，不清空已有终端内容
- 生产环境顶部栏增加醒目的"[生产]"标识，防误操作
- 断开后写入会话关闭原因，跳转回主机资产页

#### 历史会话页（/cmdb/audit/sessions）

```
筛选栏：用户、主机/IP、环境、登录账号、状态、时间范围

表格：会话ID | 用户 | 主机 | 登录账号 | 开始时间 | 持续时长 | 状态 | [详情]

详情抽屉：会话摘要 + 命令时间线
```

#### 命令审计页（/cmdb/audit/commands）

```
筛选栏：命令关键词、风险等级、用户、主机、时间范围、是否拦截

表格：命令 | 用户 | 主机 | 执行时间 | 退出码 | 风险等级 | 是否拦截
```

点击命令跳转至所在会话详情。高危命令使用醒目风险标签。

#### 凭证库页（/cmdb/access/credentials）

```
筛选栏：凭证名称、用户名、认证方式、状态

表格：凭证名称 | 登录用户 | 认证方式 | 绑定主机数 | 最近检测时间 | 状态 | 操作
```

关键：密码和私钥永远不回显；编辑时只允许覆盖更新；"测试连接"需选择目标主机。

#### 资产配置页组（/cmdb/config/*）

业务系统、机房机柜、标签管理三个标准 CRUD 页面，保留基本增删改查。主机新增/编辑表单中提供快捷创建入口，避免用户为补一个标签离开主机表单。

### 后端接口

```
# 连接权限校验
GET  /api/cmdb/servers/:id/check-connect-permission
     → {hasPermission, allowedAccounts, denyReason}

# 发起连接（创建会话）
POST /api/cmdb/servers/:id/connect
     body: {protocol, loginAccount}
     → {sessionId, websocketUrl}

# WebSocket 终端代理
GET  /api/bastion/sessions/:id/ws        升级为 WebSocket，双向转发 SSH 流

# 会话查询
GET  /api/cmdb/sessions                  分页查询（支持多条件筛选）
GET  /api/cmdb/sessions/:id              会话详情
GET  /api/cmdb/sessions/:id/commands     会话内命令列表

# 命令审计
GET  /api/cmdb/commands                  跨会话命令检索

# 凭证管理
GET    /api/cmdb/credentials
POST   /api/cmdb/credentials
PUT    /api/cmdb/credentials/:id
DELETE /api/cmdb/credentials/:id
POST   /api/cmdb/credentials/:id/test   指定主机测试连接

# 资产配置 CRUD（业务系统/机房机柜/标签）
GET/POST/PUT/DELETE /api/cmdb/business-units
GET/POST/PUT/DELETE /api/cmdb/rooms
GET/POST/PUT/DELETE /api/cmdb/tags
```

### 前后端联调目标

- 登录后资产管理菜单结构正确（无独立"主机分组"菜单）
- 主机列表按分组、业务系统、标签、环境筛选正常
- 有权限的用户可从主机列表发起 SSH 连接，打开全屏终端
- 无凭证时前端展示明确原因，引导绑定凭证
- 终端输入输出正常，窗口 resize 同步到后端
- 主动断开、网络断开、SSH 失败都能正确结束会话并更新状态
- 历史会话和命令审计能实时查询刚产生的数据
- 主机新增/编辑后，CPU/内存/磁盘字段异步回填

### 验收标准

- 完整演示路径：登录 → 主机资产 → 点击连接 → 全屏 Web SSH → 执行命令 → 断开 → 历史会话和命令审计可查
- 不使用 Mock 终端，不伪造审计数据；后端接口、WebSocket、数据库记录全部联调完成
- 资产管理核心 CRUD（主机、凭证、业务系统、机房机柜、标签）无回归

---

## 阶段二：访问策略与资产总览

**目标**：从"有连接权限即可连接"升级为"按资产范围、登录账号、时间窗口控制连接"；同时落地资产总览页和主机使用率采集。

**可运行范围**：
- 管理员在访问策略页配置策略后，用户连接行为立即受策略控制
- 资产总览页展示主机数量、会话统计、环境分布等汇总数据
- 主机表格新增 CPU%/内存%/磁盘% 使用率列，每 5 分钟自动更新

### 页面设计

#### 资产总览页（/cmdb/dashboard）

资产管理的入口页，快速了解资产规模和连接概况。

```
顶部指标区：主机总数 | 在线主机 | 今日会话次数 | 活跃会话数

中部分析区
├─ 资产按环境分布（生产/测试/开发，进度条展示）
└─ 最近 7 天连接趋势（折线图）

底部列表区
├─ 最近 5 条会话
└─ 最近 5 条命令
```

点击指标卡跳转对应列表页并自动带入筛选条件。

#### 访问策略页（/cmdb/access/policies）

```
左侧：策略分类（全部 / 临时授权 / 生产环境 / 按角色 / 按用户）

右侧：策略列表
├─ 策略名称 | 授权对象（用户/角色）| 资产范围 | 登录账号 | 时间窗口 | 状态 | 操作

策略编辑抽屉
├─ 基础信息（名称、描述）
├─ 授权对象（用户 / 角色）
├─ 资产范围（单台主机 / 主机分组 / 业务系统 / 标签）
├─ 连接限制（允许的登录账号、协议、是否允许文件传输、是否允许 sudo）
└─ 生效时间窗口（time_window，实现临时限时授权）
```

**时间窗口为临时授权的核心机制**：管理员通过设置 `time_window` 实现临时限时授权，到期后策略自动不命中，不需要审批流。

#### 主机使用率列（主机资产页增强）

**数据模型新增字段**：

`servers` 表新增：`cpu_usage` / `memory_usage` / `disk_usage` / `metrics_updated_at` / `agent_status` / `agent_port` / `agent_version`。

- `agent_status`：`uninstalled`（未安装）/ `running`（运行中）/ `offline`（离线/心跳超时）
- `agent_port`：Agent HTTP 监听端口，默认 9100，支持按主机覆盖
- `agent_version`：已部署的 Agent 版本号

**Agent 方案（替代 SSH shell 采集）**：

采集方式从"每次 SSH 执行 top/free/df 解析"改为"在目标主机部署自研轻量 Agent，后端定时 HTTP 拉取"。

```
目标主机 Agent (gopsutil)
  └─ 暴露 HTTP :9100/metrics  ←── 每5分钟 HTTP GET ── OneOps 调度器 ──→ MySQL

Agent 部署（一次性）：
  OneOps 后端 ──SSH──→ 上传预编译二进制 ──→ 启动 systemd 服务
  （部署完成后不再依赖 SSH 凭证做采集）
```

**Agent 实现**：

- 使用 `gopsutil` 库直接读取 `/proc`，精度高且跨发行版兼容
- 暴露 HTTP JSON 接口（`GET /metrics`），返回 CPU%、内存%、磁盘%、load5、进程总数
- 定期向 OneOps 后端发送心跳（`POST /api/cmdb/agent/heartbeat`），携带 hostname / IP / PID / 版本
- 以 systemd 服务运行，`Restart=always`，崩溃自动重启
- 预编译各平台二进制（linux/amd64、linux/arm64），存放在 OneOps 服务器，部署时通过 SSH SCP 传输 + 启动，不在运行时动态生成代码编译

**Agent 部署流程**：

1. 用户在主机表格或 Agent 管理界面选择目标主机，点击"部署 Agent"
2. 后端使用主机绑定的**系统运维凭证**（`system_credential_id`）将预编译二进制 SCP 到目标主机 `/opt/oneops-agent/`；若未配置系统凭证，直接返回错误，不 fallback
3. 通过 SSH 执行 systemd 注册和启动命令（需 root 或 sudo 权限）
4. 前端轮询 Agent 状态直到变为 `running` 或 `failed`
5. Agent 启动后开始发送心跳，后端收到心跳后更新 `agent_status=running`

**后端调度器改动**：

- `StartMetricsScheduler` 每 5 分钟对 `agent_status=running` 的主机发起 HTTP GET `http://{innerIP}:{agentPort}/metrics`
- 解析 JSON 响应写入 `cpu_usage` / `memory_usage` / `disk_usage` / `metrics_updated_at`
- 心跳超时检测（> 3 分钟无心跳）将 `agent_status` 置为 `offline`，停止采集

**前端展示**：

主机表格新增两列：
- "使用率"（宽 160px）：Agent 运行且有采集数据时，显示 CPU/MEM/DSK 三条进度条，颜色按阈值（绿 <70%、黄 70-90%、红 >90%）；`uninstalled` 时显示灰色"未安装"标签；`offline` 时显示橙色"Agent 离线"标签；`running` 首次采集中显示绿色"采集中"标签
- "Agent"（宽 130px）：状态徽标 + 操作入口
  - `running`：绿色"运行中"徽标，hover 时 tooltip 显示版本号（如 `v1.0.0`）
  - `offline`：红色"离线"徽标 + "重启"链接按钮（直接触发重启，无需打开更多菜单）
  - `uninstalled`：灰色"未安装"徽标 + "部署"链接按钮（直接触发部署）

主机"更多"下拉菜单 Agent 相关选项（根据当前 agent_status 动态显示）：

| agent_status | 显示选项 |
|---|---|
| `running` | 刷新指标 / 重启 Agent / 卸载 Agent |
| `offline` | 重启 Agent / 卸载 Agent |
| `uninstalled` | 部署 Agent |

**轮询终态规则**：部署、重启操作提交后轮询预期终态 `running`；卸载操作提交后轮询预期终态 `uninstalled`。轮询每 3 秒一次，最多 20 次，达到预期终态或超时后均停止并刷新列表。

#### 独立 Agent 管理页面（/cmdb/config/agents）

在"资产配置"目录下设置独立 Agent 管理页面，补充主机表格内嵌操作无法覆盖的批量场景。

**页面布局**：

```
顶部工具栏
  [搜索框: 主机名/IP]  [状态筛选: 全部/运行中/离线/未安装]  [刷新]  [批量部署]  [批量卸载]

Agent 列表（el-table）
  □  主机名称  IP地址  内网IP  版本  状态  监听端口  最近心跳  操作
  □  web-01   1.2.3.4  -      v1.0  运行中  9100   2分钟前   [重启] [卸载] [删除记录]
  □  db-01    1.2.3.5  -      -     未安装   -      -        [部署]
  □  redis-01 1.2.3.6  -      v1.0  离线    9100   8分钟前   [重启] [卸载] [删除记录]
```

**状态徽标颜色**：
- `running`：绿色"运行中"
- `offline`：红色"离线"（tooltip 显示最近心跳时间）
- `uninstalled`：灰色"未安装"

**搜索与筛选**：
- 主机名 / IP 模糊搜索
- Agent 状态单选（全部 / 运行中 / 离线 / 未安装）
- 分页（pageSize 默认 20）

**操作说明**：

| 操作 | 条件 | 行为 |
|------|------|------|
| 部署 | `uninstalled` | 调用 deploy 接口，轮询至 `running` |
| 重启 | `offline` 或 `running` | 调用 restart 接口，轮询至 `running` |
| 卸载 | `running` 或 `offline` | 二次确认 → 调用 uninstall 接口，轮询至 `uninstalled` |
| 删除记录 | `offline` 或 `uninstalled` | 二次确认 → 清除该主机的 agent 字段（不 SSH，仅更新数据库），立即刷新列表 |
| 批量部署 | 勾选若干 `uninstalled` 主机 | 逐台串行调用 deploy，工具栏显示进度（X/Y 台完成） |
| 批量卸载 | 勾选若干 `running`/`offline` 主机 | 二次确认 → 逐台串行调用 uninstall |

**轮询策略**：操作后同样采用 3 秒/次、最多 20 次的轮询，达到预期终态停止；批量操作时各主机独立轮询互不阻塞。

**与主机表格内嵌操作的关系**：

两者并存、互补：
- 主机资产表格的 Agent 列提供单台快速操作（部署/重启），适合偶发场景
- 独立 Agent 管理页面提供批量视角和完整状态总览，适合批量部署或集中排查离线 Agent

### 后端接口

```
# 访问策略 CRUD
GET    /api/cmdb/access-policies
POST   /api/cmdb/access-policies
PUT    /api/cmdb/access-policies/:id
DELETE /api/cmdb/access-policies/:id

# 资产总览统计
GET /api/cmdb/stats/servers     主机统计（按环境、状态分布）
GET /api/cmdb/stats/sessions    会话统计（今日/活跃/近7天趋势）

# Agent 管理
POST /api/cmdb/servers/:id/agent/deploy     部署 Agent（异步，优先用系统凭证 SCP + 启动，见"凭证分层设计"）
POST /api/cmdb/servers/:id/agent/restart    重启 Agent（SSH 执行 systemctl restart，用系统凭证）
POST /api/cmdb/servers/:id/agent/uninstall  卸载 Agent（SSH 执行 systemctl stop + 文件清理，用系统凭证）
GET  /api/cmdb/servers/:id/agent/status     查询 Agent 当前状态

# Agent 管理页面专用接口
GET  /api/cmdb/agents                       Agent 列表（支持 hostname/ip/status 筛选 + 分页）
POST /api/cmdb/agents/batch-deploy          批量部署（body: { serverIds: [1,2,3] }）
POST /api/cmdb/agents/batch-uninstall       批量卸载（body: { serverIds: [1,2,3] }）
DELETE /api/cmdb/agents/:id                 删除 Agent 记录（仅清空数据库 agent 字段，不 SSH）

# Agent 心跳接收（由 Agent 主动上报，不需要 Auth 中间件）
POST /api/cmdb/agent/heartbeat              接收 Agent 心跳，更新 agent_status 和 last_heartbeat_at

# 手动触发单台指标采集（智能分发）
POST /api/cmdb/servers/:id/sync-metrics
     当 agent_status=running → 触发 Agent HTTP GET /metrics 拉取
     否则 → fallback 到 SSH 采集（top/free/df）
```

### 前后端联调目标

- 管理员创建策略后，用户连接行为立即受影响（策略命中 → 允许；无匹配 → 拒绝 + 返回原因）
- 策略停用后连接判断实时变化
- 设置 `time_window` 的临时策略在窗口外不命中
- 资产总览页数据来自后端真实接口
- Agent 部署流程完整可用：部署 → 轮询状态 → 变为"运行中" → 使用率列有数据
- Agent 心跳超时后 `agent_status` 自动变为 `offline`，使用率列显示"Agent 离线"
- 手动触发"刷新指标"3 秒后表格自动刷新使用率数据
- Agent 管理页面可查看全量主机的 Agent 状态，支持按状态筛选
- 批量部署：勾选多台未安装主机 → 点击批量部署 → 进度条展示各台完成情况
- 离线 Agent 可在管理页面删除记录（仅清库，不 SSH），记录删除后该主机 Agent 列显示"未安装"

### 验收标准

- 演示路径：创建策略 → 用户连接被允许 → 停用策略 → 用户无法连接 → 启用临时时间窗口 → 窗口内可连接
- 策略支持主机、分组、标签、业务系统四种资产范围中的至少两种
- 资产总览数据正确，不依赖 Mock
- 演示 Agent 部署：选择一台有 SSH 凭证的主机 → 部署 Agent → 状态变为"运行中" → 使用率列展示真实数据
- 阶段一的 SSH 连接能力不回归

---

## 阶段三：在线会话与审计增强

**目标**：增强堡垒机管控能力，让管理员可以实时查看和强制断开会话，并对 SFTP 文件传输进行审计。

**可运行范围**：
- 管理员在在线会话页查看活跃会话，可强制断开
- 用户按策略进行 SFTP 文件传输，系统记录所有上传/下载操作
- 命令审计支持风险等级标记，高危命令标记告警

### 页面设计

#### 在线会话页（/cmdb/audit/online）

```
顶部统计：在线会话数 | 生产环境会话数 | 异常会话数（每30秒自动刷新）

表格：用户 | 主机 | 登录账号 | 来源IP | 开始时间 | 持续时长 | 状态 | [强制断开]
```

强制断开后：用户终端收到"会话已被管理员终止"提示，会话状态更新为 terminated。

#### 历史会话详情增强

会话详情抽屉新增：文件传输记录 Tab、风险事件列表。

#### 文件传输页（/cmdb/audit/files）

```
筛选栏：用户、主机、方向（上传/下载）、文件路径、时间范围、状态

表格：文件路径 | 方向 | 大小 | 用户 | 主机 | 会话 | 时间 | 状态
```

策略禁止文件传输时，后端拒绝并返回明确原因，前端入口同时禁用。

#### 命令审计增强

在已有基础上新增：风险等级（low/medium/high）自动标记；高危命令规则（`rm -rf`、`mkfs`、`shutdown` 等）写入 `risk_level` 字段；支持是否拦截筛选。

### 后端接口

```
# 在线会话管理
GET  /api/cmdb/sessions/online              当前活跃会话列表
POST /api/cmdb/sessions/:id/terminate       强制断开会话

# 文件传输审计
GET  /api/cmdb/file-transfers               跨会话文件传输检索
GET  /api/cmdb/sessions/:id/file-transfers  单会话文件传输记录
```

后端使用 `github.com/pkg/sftp` 实现文件传输代理，传输开始/完成/失败写入 `bastion_file_transfers` 表，并对接访问策略的 `allow_file_transfer` 字段做控制。

### 前后端联调目标

- 用户建立连接后，管理员在在线会话页立即看到该会话
- 管理员强制断开后，用户终端收到断开提示，会话状态实时更新
- 用户上传/下载文件后，文件传输页能查询到记录
- 策略设置 `allow_file_transfer=false` 时，后端拒绝传输，前端 SFTP 入口禁用
- 高危命令被标记，可在命令审计页用风险等级筛选

### 验收标准

- 演示路径：用户连接 → 管理员看到在线会话 → 强制断开 → 用户终端收到提示 → 会话审计状态更新
- 演示文件上传/下载，文件传输审计中查询到记录
- 策略控制文件传输权限，前端无法绕过后端校验
- 阶段一、阶段二的能力不回归

---

## 后期规划（P2+）

以下能力纳入后期规划，待阶段一至三稳定运行后评估引入。每项能力引入时必须形成独立可运行闭环，不以破坏现有 SSH 能力为代价。

### 审批流

`bastion_approvals` 表 + 审批记录页（`/cmdb/access/approvals`）+ 连接弹窗"申请访问"按钮 + 审批状态校验。

当前替代：访问策略 `time_window` + 管理员直接调整策略。

触发时机：团队规模扩大（管理员无法及时手动调整策略）；有合规要求，需要留存完整申请/审批/过期审计链路。

### 权限点体系

`cmdb:server:connect` 等按钮级权限码，需要：
1. 后端维护独立 buttons 数据源
2. 登录接口返回 buttons 数组
3. 前端 `hasAuth()` 控制按钮显示
4. 后端接口中间件校验权限码

触发时机：平台用户超 500 人；需要按角色差异化按钮操作能力；有等保/SOC2 合规要求。

### 会话回放

终端录屏存储（ASCII cast 格式）+ 回放页面，存储迁移到对象存储。

### 多协议扩展

- **RDP/VNC**：通过 Apache Guacamole 或独立协议网关接入，避免在现有 Go 服务中实现桌面协议
- **Kubernetes exec**：client-go + WebSocket/stream，支持 Pod exec、日志查看
- **数据库登录代理**：独立代理服务，支持 MySQL/PostgreSQL 登录审计，与 SSH 代理解耦

### 监控采集架构演进（Pushgateway）

当前 Agent 方案为 OneOps 主动 HTTP 拉取（pull 模式），要求 OneOps 服务器能访问目标主机 agentPort（默认 9100）。

面对内网主机或防火墙隔离场景，可引入 Prometheus Pushgateway 升级为推送模式：

```
目标主机 Agent ──每30秒 PUSH──→ Pushgateway :9091
Prometheus :9090 ──scrape──→ Pushgateway
OneOps 后端 ──PromQL──→ Prometheus → 写 MySQL / 返回前端
```

升级收益：
- 解决 OneOps 无法主动访问内网主机的网络问题（Agent 只需能访问 Pushgateway）
- Prometheus 存储时序数据，可做历史趋势图和告警规则
- 支持进程监控、HTTP 探活、Ping 等扩展业务监控

升级前置条件：当前 pull 模式稳定运行；需要时序历史数据；存在防火墙隔离场景。

触发时机：被管主机数量超过 200 台，或出现大量防火墙后主机，或需要监控历史趋势功能。

**注意**：引入 Pushgateway 需同步实现心跳超时触发清理（`DELETE /metrics/job/{hostname}`），防止离线主机数据在 Pushgateway 中永久残留。

从 MySQL AES-GCM 加密存储迁移到 Vault、云 KMS 或企业密钥管理系统，满足更高合规要求。

### 复杂权限引擎（Casbin）

当策略条件变得复杂（ABAC、资源属性、优先级策略、多租户隔离）时再评估引入。
