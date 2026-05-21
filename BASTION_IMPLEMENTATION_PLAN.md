# 堡垒机能力集成实施计划

## 总体目标

将堡垒机能力集成到现有 CMDB 资产管理中，实现"资产管理 = 资产台账 + 访问入口 + 访问审计"的定位。

## 实施原则

1. **不创建独立堡垒机模块**：将访问能力作为资产管理的一部分
2. **前端不接触敏感凭证**：密码和私钥只在后端使用
3. **所有操作可审计**：连接、命令、文件传输都留痕
4. **分阶段实施**：P0 最小可用版本 → P1 增强版本 → P2 扩展

## 数据库设计

### 1. 访问策略表 (asset_access_policies)

```sql
CREATE TABLE asset_access_policies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '策略名称',
    subject_type ENUM('user', 'role', 'user_group') NOT NULL COMMENT '授权对象类型',
    subject_id BIGINT UNSIGNED NOT NULL COMMENT '授权对象ID',
    asset_scope_type ENUM('server', 'group', 'business', 'tag') NOT NULL COMMENT '资产范围类型',
    asset_scope_id BIGINT UNSIGNED NOT NULL COMMENT '资产范围ID，all表示全部',
    login_accounts JSON COMMENT '允许使用的登录账号 ["root", "appuser"]',
    protocols JSON COMMENT '允许的协议 ["ssh", "sftp"]',
    allow_file_transfer BOOLEAN DEFAULT TRUE COMMENT '是否允许文件传输',
    allow_sudo BOOLEAN DEFAULT FALSE COMMENT '是否允许sudo',
    require_approval BOOLEAN DEFAULT FALSE COMMENT '是否需要审批',
    time_window JSON COMMENT '时间窗口 {"start": "09:00", "end": "18:00", "days": [1,2,3,4,5]}',
    high_risk_commands JSON COMMENT '高危命令列表 ["rm -rf", "mkfs", "shutdown"]',
    status TINYINT DEFAULT 1 COMMENT '状态 1启用 0禁用',
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    INDEX idx_subject (subject_type, subject_id),
    INDEX idx_asset_scope (asset_scope_type, asset_scope_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资产访问策略表';
```

### 2. 会话表 (bastion_sessions)

```sql
CREATE TABLE bastion_sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    login_account VARCHAR(50) NOT NULL COMMENT '登录账号',
    client_ip VARCHAR(50) COMMENT '客户端IP',
    protocol ENUM('ssh', 'sftp') DEFAULT 'ssh' COMMENT '连接协议',
    ssh_credential_id BIGINT UNSIGNED COMMENT '使用的SSH凭证ID',
    started_at DATETIME(3) COMMENT '开始时间',
    ended_at DATETIME(3) COMMENT '结束时间',
    duration INT COMMENT '持续时长(秒)',
    status ENUM('active', 'closed', 'error', 'terminated') DEFAULT 'active' COMMENT '会话状态',
    close_reason VARCHAR(200) COMMENT '关闭原因',
    created_at DATETIME(3) NULL,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_server_user (server_id, user_id),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='堡垒机会话表';
```

### 3. 命令审计表 (bastion_commands)

```sql
CREATE TABLE bastion_commands (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
    command TEXT NOT NULL COMMENT '执行的命令',
    executed_at DATETIME(3) COMMENT '执行时间',
    exit_code INT COMMENT '退出码',
    risk_level ENUM('safe', 'low', 'medium', 'high', 'critical') DEFAULT 'safe' COMMENT '风险等级',
    blocked BOOLEAN DEFAULT FALSE COMMENT '是否被拦截',
    output_summary TEXT COMMENT '输出摘要',
    created_at DATETIME(3) NULL,
    FOREIGN KEY (session_id) REFERENCES bastion_sessions(id) ON DELETE CASCADE,
    INDEX idx_session (session_id),
    INDEX idx_risk_level (risk_level),
    INDEX idx_executed_at (executed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='命令审计表';
```

### 4. 文件传输审计表 (bastion_file_transfers)

```sql
CREATE TABLE bastion_file_transfers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
    direction ENUM('upload', 'download') NOT NULL COMMENT '传输方向',
    remote_path VARCHAR(500) NOT NULL COMMENT '远程路径',
    local_path VARCHAR(500) COMMENT '本地路径',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    status ENUM('pending', 'transferring', 'success', 'failed') DEFAULT 'pending' COMMENT '传输状态',
    error_message VARCHAR(500) COMMENT '错误信息',
    started_at DATETIME(3) COMMENT '开始时间',
    completed_at DATETIME(3) COMMENT '完成时间',
    created_at DATETIME(3) NULL,
    FOREIGN KEY (session_id) REFERENCES bastion_sessions(id) ON DELETE CASCADE,
    INDEX idx_session (session_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件传输审计表';
```

### 5. 审批表 (bastion_approvals)

```sql
CREATE TABLE bastion_approvals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    requester_id BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
    approver_id BIGINT UNSIGNED COMMENT '审批人ID',
    server_id BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    login_account VARCHAR(50) NOT NULL COMMENT '登录账号',
    reason TEXT COMMENT '申请原因',
    status ENUM('pending', 'approved', 'rejected', 'expired') DEFAULT 'pending' COMMENT '审批状态',
    requested_at DATETIME(3) COMMENT '申请时间',
    approved_at DATETIME(3) COMMENT '审批时间',
    expired_at DATETIME(3) COMMENT '过期时间',
    comment TEXT COMMENT '审批意见',
    created_at DATETIME(3) NULL,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    INDEX idx_requester (requester_id),
    INDEX idx_status (status),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='连接审批表';
```

### 6. 服务器表扩展

在 servers 表增加字段：

```sql
ALTER TABLE servers ADD COLUMN ssh_credential_id BIGINT UNSIGNED NULL COMMENT '绑定的SSH凭证ID';
ALTER TABLE servers ADD COLUMN last_connect_time DATETIME(3) COMMENT '最后连接时间';
ALTER TABLE servers ADD COLUMN connectivity_status ENUM('online', 'offline', 'unknown') DEFAULT 'unknown' COMMENT '连通性状态';
ALTER TABLE servers ADD COLUMN last_check_time DATETIME(3) COMMENT '最后连通性检查时间';
```

## 后端实施计划

### 阶段 P0：最小可用版本

#### 1. 模型层 (models/)

创建以下模型文件：

- `models/bastion.go` - 堡垒机相关模型
  - AssetAccessPolicy
  - BastionSession
  - BastionCommand
  - BastionFileTransfer
  - BastionApproval

#### 2. 服务层 (services/)

创建 `services/bastion.go`：

**连接服务**
- `CheckConnectPermission(userID, serverID) (bool, error)` - 检查连接权限
- `CreateSSHSession(userID, serverID, loginAccount, clientIP) (*BastionSession, error)` - 创建SSH会话
- `CloseSession(sessionID, reason) error` - 关闭会话
- `GetActiveSessions() ([]BastionSession, error)` - 获取活跃会话

**命令审计服务**
- `RecordCommand(sessionID, command, exitCode) error` - 记录命令
- `AnalyzeCommandRisk(command) string` - 分析命令风险等级
- `IsCommandBlocked(command) bool` - 检查命令是否被拦截

**访问策略服务**
- `GetUserPolicies(userID) ([]AssetAccessPolicy, error)` - 获取用户策略
- `CheckAccessPolicy(userID, serverID) (bool, []string, error)` - 检查访问策略，返回允许的账号列表

**会话审计服务**
- `GetSessionList(page, pageSize) ([]BastionSession, int64, error)` - 获取会话列表
- `GetSessionCommands(sessionID) ([]BastionCommand, error)` - 获取会话命令
- `GetSessionFileTransfers(sessionID) ([]BastionFileTransfer, error)` - 获取文件传输记录

#### 3. WebSocket 处理器 (handlers/)

创建 `handlers/websocket.go`：

- `SSHWebSocketHandler` - SSH WebSocket 处理器
  - 升级 HTTP 连接到 WebSocket
  - 建立到目标主机的 SSH 连接
  - 双向转发数据（终端 ↔ SSH）
  - 记录所有命令
  - 处理会话关闭

#### 4. 控制器层 (controllers/)

在 `controllers/cmdb.go` 中增加：

**连接相关**
- `ConnectServer(ctx)` - 创建连接会话
- `GetServerSessions(ctx)` - 获取服务器会话列表

**会话管理**
- `GetSessions(ctx)` - 获取会话列表
- `GetSessionByID(ctx)` - 获取会话详情
- `TerminateSession(ctx)` - 强制断开会话
- `GetSessionCommands(ctx)` - 获取会话命令
- `GetSessionFileTransfers(ctx)` - 获取会话文件传输

**访问策略**
- `GetAccessPolicies(ctx)` - 获取访问策略列表
- `CreateAccessPolicy(ctx)` - 创建访问策略
- `UpdateAccessPolicy(ctx)` - 更新访问策略
- `DeleteAccessPolicy(ctx)` - 删除访问策略

**审批管理**
- `GetApprovals(ctx)` - 获取审批列表
- `CreateApproval(ctx)` - 创建审批申请
- `ApproveRequest(ctx)` - 审批通过
- `RejectRequest(ctx)` - 审批拒绝

#### 5. 路由层 (routes/)

在 `routes/routes.go` 中增加：

```go
// 连接相关
cmdbGroup.POST("/servers/:id/connect", cmdbController.ConnectServer)
cmdbGroup.GET("/servers/:id/sessions", cmdbController.GetServerSessions)

// WebSocket
cmdbGroup.GET("/sessions/:id/ws", gin.WrapH(handlers.SSHWebSocketHandler()))

// 会话管理
cmdbGroup.GET("/sessions", cmdbController.GetSessions)
cmdbGroup.GET("/sessions/:id", cmdbController.GetSessionByID)
cmdbGroup.POST("/sessions/:id/terminate", cmdbController.TerminateSession)
cmdbGroup.GET("/sessions/:id/commands", cmdbController.GetSessionCommands)
cmdbGroup.GET("/sessions/:id/file-transfers", cmdbController.GetSessionFileTransfers)

// 访问策略
cmdbGroup.GET("/access-policies", cmdbController.GetAccessPolicies)
cmdbGroup.POST("/access-policies", cmdbController.CreateAccessPolicy)
cmdbGroup.PUT("/access-policies/:id", cmdbController.UpdateAccessPolicy)
cmdbGroup.DELETE("/access-policies/:id", cmdbController.DeleteAccessPolicy)

// 审批管理
cmdbGroup.GET("/approvals", cmdbController.GetApprovals)
cmdbGroup.POST("/approvals", cmdbController.CreateApproval)
cmdbGroup.POST("/approvals/:id/approve", cmdbController.ApproveRequest)
cmdbGroup.POST("/approvals/:id/reject", cmdbController.RejectRequest)
```

### 阶段 P1：增强版本

- 在线会话监控仪表板
- SFTP 文件传输审计
- 高危命令实时拦截
- 会话回放功能
- 生产环境审批流程
- 批量授权功能

### 阶段 P2：扩展版本

- RDP/VNC 协议支持
- Kubernetes exec 支持
- 数据库代理
- 外部堡垒机对接
- 零信任访问策略

## 前端实施计划

### 菜单结构调整

删除 `/cmdb/groups`，新增：

```text
资产管理
├─ 资产总览 [新增]
├─ 主机资产
│  ├─ 主机列表（含连接按钮）
│  ├─ 主机详情抽屉 [增强]
│  └─ 分组管理（内嵌）
├─ 访问管理 [新增]
│  ├─ 访问策略 [新增]
│  ├─ 凭证库 [重命名自SSH凭证]
│  └─ 连接审批 [新增]
└─ 会话审计 [新增]
   ├─ 在线会话 [新增]
   ├─ 历史会话 [新增]
   ├─ 命令审计 [新增]
   └─ 文件传输 [新增]
```

### 页面实施列表

#### 1. 主机资产页增强 (`/cmdb/servers`)

**增加功能**：
- 操作列增加"连接"按钮
- 主机详情抽屉组件
- 连接弹窗组件
- 凭证绑定状态显示
- 连通性状态显示

#### 2. 主机详情抽屉组件

**Tab 结构**：
- 基础信息
- 连接信息（SSH端口、凭证绑定、连通性）
- 分组与标签
- 凭证绑定
- 会话记录
- 变更记录

#### 3. 连接弹窗组件

**功能**：
- 选择登录账号
- 显示凭证来源
- 连接前检查（权限、凭证、连通性）
- 错误提示（无权限、无凭证、离线、需审批等）
- 打开 Web SSH 终端

#### 4. SSH 终端组件

**功能**：
- 使用 xterm.js 渲染终端
- WebSocket 连接到后端 SSH 代理
- 支持终端大小调整
- 支持复制粘贴
- 断线重连

#### 5. 访问策略页 (`/cmdb/access-policies`)

**功能**：
- 策略列表（名称、授权对象、资产范围、状态）
- 创建/编辑策略表单
- 策略删除确认
- 策略启用/禁用

**策略表单字段**：
- 策略名称
- 授权对象类型（用户/角色）
- 授权对象（多选）
- 资产范围类型（主机/分组/业务/标签）
- 资产范围（多选）
- 允许的登录账号
- 允许的协议
- 文件传输权限
- sudo 权限
- 是否需要审批
- 时间窗口
- 高危命令列表

#### 6. 会话审计页 (`/cmdb/sessions`)

**功能**：
- 会话列表（用户、主机、IP、状态、时长）
- 状态筛选（活跃/已关闭）
- 查看会话详情
- 强制断开会话
- 查看会话命令
- 查看文件传输

#### 7. 命令审计页 (`/cmdb/commands`)

**功能**：
- 命令列表（会话、命令、执行时间、退出码、风险等级）
- 按时间筛选
- 按风险等级筛选
- 按用户/主机筛选
- 高危命令高亮显示

#### 8. 在线会话页 (`/cmdb/sessions/active`)

**功能**：
- 实时显示活跃会话
- 支持监控会话
- 支持强制断开
- 显示会话时长

#### 9. 连接审批页 (`/cmdb/approvals`)

**功能**：
- 审批申请列表
- 审批操作（通过/拒绝）
- 审批意见填写
- 申请详情查看

### 权限定义

在数据库中新增权限：

```sql
-- 服务器连接权限
INSERT INTO permissions (code, name, description) VALUES
('cmdb:server:connect', '连接服务器', '允许通过堡垒机连接服务器'),
('cmdb:session:query', '查看会话', '允许查看会话列表和详情'),
('cmdb:session:terminate', '断开会话', '允许强制断开会话'),
('cmdb:session:replay', '回放会话', '允许回放会话记录'),
('cmdb:command:query', '查看命令', '允许查看命令审计'),
('cmdb:file-transfer:query', '查看文件传输', '允许查看文件传输记录'),
('cmdb:access-policy:query', '查看访问策略', '允许查看访问策略'),
('cmdb:access-policy:create', '创建访问策略', '允许创建访问策略'),
('cmdb:access-policy:update', '更新访问策略', '允许更新访问策略'),
('cmdb:access-policy:delete', '删除访问策略', '允许删除访问策略'),
('cmdb:approval:query', '查看审批', '允许查看审批申请'),
('cmdb:approval:approve', '审批申请', '允许审批连接申请');
```

## 安全设计

### 1. 凭证安全

- SSH 凭证加密存储（使用 AES-256-GCM）
- 密钥使用环境变量或密钥管理系统
- 定期轮换加密密钥
- 凭证使用日志记录

### 2. 连接安全

- WebSocket 使用 WSS（加密传输）
- 会话超时自动断开
- 异常连接检测（短时间频繁连接）
- IP 白名单/黑名单

### 3. 命令安全

- 高危命令黑名单
- 命令风险等级评估
- 违规命令拦截
- 命令审计日志

### 4. 审计安全

- 所有操作留痕
- 审计日志不可删除
- 审计日志定期归档
- 审计日志访问权限控制

## 技术选型

### 后端依赖

```go
// SSH 库
github.com/gliderlabs/ssh

// WebSocket
github.com/gorilla/websocket

// 加密
crypto/aes
crypto/cipher
```

### 前端依赖

```json
{
  "xterm": "^5.3.0",
  "xterm-addon-fit": "^0.8.0",
  "xterm-addon-web-links": "^0.9.0"
}
```

## 实施步骤

### 第一步：数据库准备（1小时）

1. 创建新表结构
2. 修改 servers 表增加字段
3. 插入测试数据
4. 添加权限记录

### 第二步：后端模型层（1小时）

1. 创建 bastion.go 模型文件
2. 定义所有模型结构
3. 添加 GORM 关联关系

### 第三步：后端服务层（3小时）

1. 实现权限检查服务
2. 实现 SSH 会话管理服务
3. 实现命令审计服务
4. 实现访问策略服务
5. 实现审批服务

### 第四步：WebSocket SSH 代理（4小时）

1. 实现 SSH 连接建立
2. 实现 WebSocket 升级
3. 实现双向数据转发
4. 实现命令记录
5. 实现会话管理

### 第五步：后端 API 接口（2小时）

1. 实现连接接口
2. 实现会话管理接口
3. 实现访问策略接口
4. 实现审批接口
5. 添加路由注册

### 第六步：前端组件开发（6小时）

1. 开发 SSH 终端组件
2. 开发连接弹窗组件
3. 开发主机详情抽屉
4. 开发访问策略页面
5. 开发会话审计页面
6. 开发命令审计页面
7. 开发审批页面

### 第七步：集成测试（2小时）

1. 功能测试
2. 权限测试
3. 安全测试
4. 性能测试

## 预计工期

- P0 最小可用版本：约 19 小时
- P1 增强版本：约 16 小时
- P2 扩展版本：待定

## 验收标准

### P0 阶段验收标准

1. ✅ 用户可以从主机列表点击"连接"按钮
2. ✅ 系统能检查用户是否有连接权限
3. ✅ 系统能检查主机是否绑定了 SSH 凭证
4. ✅ 成功建立 Web SSH 连接
5. ✅ 所有执行的命令都被记录
6. ✅ 会话信息（用户、主机、时间、状态）被正确记录
7. ✅ 可以在会话审计页面查看历史会话
8. ✅ 可以在命令审计页面查看命令记录
9. ✅ 前端无法获取明文密码或私钥

### P1 阶段验收标准

1. ✅ 管理员可以查看和强制断开活跃会话
2. ✅ SFTP 文件传输被记录和审计
3. ✅ 生产环境连接需要审批
4. ✅ 高危命令可以被拦截
5. ✅ 支持会话回放
6. ✅ 支持按分组/标签批量授权

## 风险与注意事项

1. **SSH 连接稳定性**：需要处理网络中断、主机重启等异常情况
2. **并发连接限制**：需要限制单用户和单主机的并发连接数
3. **性能问题**：命令记录、会话监控可能产生大量数据，需要定期归档
4. **安全风险**：WebSocket 需要使用 WSS，凭证需要加密存储
5. **兼容性**：需要测试不同操作系统的 SSH 兼容性

## 后续优化方向

1. 实现会话录像和回放
2. 实现智能风险评估（基于历史行为）
3. 实现自动化审批（基于信任等级）
4. 实现 MFA 多因素认证
5. 实现会话共享和协作
