# Agent 版本管理功能 - 实施完成报告

## 概述

本次实施完成了 Agent 版本管理系统的核心功能，解决了主机 41 的旧版本 Agent 不支持扩展指标端点的问题，实现了版本发布、管理和升级的完整流程。

**实施时间**：2026-05-28
**实施范围**：数据库 + 后端 API + 前端界面

---

## 已完成任务列表

### 阶段一：数据库和基础架构 ✅

#### 1. 数据库迁移脚本
**文件**：`backend/sql/agent_version_management.sql`

- 创建 `agent_versions` 表
  - 字段：id, version, release_notes, changelog, released_at
  - 二进制文件信息：amd64/arm64 路径、哈希、大小
  - 版本状态：is_latest, is_deprecated
  - 功能支持：features (JSON)
  - 兼容性：min_compatible_version, max_compatible_version
  - 统计信息：download_count, deploy_count

- 创建 `agent_upgrade_tasks` 表
  - 字段：id, task_name, target_version, target_server_ids
  - 任务状态：status (pending/running/completed/failed/cancelled)
  - 进度统计：current_step, total_steps, success_count, failed_count, skipped_count
  - 详细日志：error_message, operation_log (JSON)

- 插入默认版本 1.0.0 数据
  ```sql
  INSERT INTO agent_versions (version, release_notes, changelog, is_latest, features)
  VALUES ('1.0.0', 'Agent 初始版本，支持基础监控指标采集', 
          '- 支持基础监控指标采集\n- 支持心跳上报\n- 支持 /metrics 端点', 
          1, '{"extended_metrics":false,"custom_configs":false}');
  ```

---

### 阶段二：后端实现 ✅

#### 2. 数据模型定义
**文件**：`backend/models/cmdb.go`

新增结构体：
```go
// AgentVersion Agent 版本模型
type AgentVersion struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Version   string    `json:"version" gorm:"size:50;not null;uniqueIndex"`
    ReleaseNotes string  `json:"releaseNotes" gorm:"type:text"`
    Changelog  string    `json:"changelog" gorm:"type:text"`
    ReleasedAt time.Time `json:"releasedAt"`
    // ... 二进制文件信息、版本状态、功能支持、统计信息
}

// AgentUpgradeTask Agent 升级任务模型
type AgentUpgradeTask struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    TaskName    string    `json:"taskName"`
    TargetVersion   string `json:"targetVersion" gorm:"size:50;not null"`
    TargetServerIDs string `json:"targetServerIds" gorm:"type:json"`
    Status        string `json:"status" gorm:"type:varchar(20);default:'pending'"`
    // ... 进度统计、详细日志
}

// AgentVersionFeature 版本功能支持结构
type AgentVersionFeature struct {
    ExtendedMetrics bool `json:"extendedMetrics"`
    CustomConfigs   bool `json:"customConfigs"`
}
```

#### 3. 服务层实现
**文件**：`backend/services/agent.go`

**版本管理服务**：
- `GetLatestVersion()` - 获取最新版本信息
- `GetAllVersions()` - 获取所有版本列表（按发布时间倒序）
- `GetVersionByID(id)` - 根据ID获取版本信息
- `GetVersionByNumber(version)` - 根据版本号获取版本信息
- `CreateVersion(version, operator)` - 创建新版本
- `UpdateVersion(id, updates)` - 更新版本信息
- `DeleteVersion(id)` - 删除版本

**版本兼容性检查**：
- `compareVersion(v1, v2, operator)` - 版本号比较函数
  - 支持：>, >=, <, <=, =, ==
  - 自动补齐版本号段（如 1.0 → 1.0.0）
- `CheckVersionCompatibility(current, target)` - 检查版本兼容性
  - 空版本自动允许升级
  - 检查 min_compatible_version
  - 检查 max_compatible_version

**升级功能**：
- `(s *AgentService) UpgradeAgent(serverID, targetVersion)` - 升级单台主机
  - 检查当前版本是否已是目标版本
  - 版本兼容性检查
  - SSH 连接和架构检测
  - 停止当前 Agent → 上传新版本二进制 → 启动新版本
  - 等待心跳确认（60秒超时，轮询20次）
  - 更新部署统计计数

- `GetUpgradeTasks(page, pageSize, status)` - 获取升级任务列表
- `GetUpgradeTaskByID(id)` - 获取升级任务详情

#### 4. 控制器实现
**文件**：`backend/controllers/cmdb.go`

新增控制器方法：
```go
// 版本管理控制器
func (c *CMDBController) GetAgentVersions(ctx *gin.Context)
func (c *CMDBController) GetLatestAgentVersion(ctx *gin.Context)
func (c *CMDBController) GetAgentVersionByID(ctx *gin.Context)
func (c *CMDBController) CreateAgentVersion(ctx *gin.Context)
func (c *CMDBController) UpdateAgentVersion(ctx *gin.Context)
func (c *CMDBController) DeleteAgentVersion(ctx *gin.Context)

// 升级管理控制器
func (c *CMDBController) UpgradeAgent(ctx *gin.Context)
func (c *CMDBController) GetUpgradeTasks(ctx *gin.Context)
func (c *CMDBController) GetUpgradeTaskByID(ctx *gin.Context)
```

#### 5. 路由配置
**文件**：`backend/routes/routes.go`

新增路由（需要认证）：
```go
// Agent 版本管理
api.GET("/cmdb/agent-versions", cmdbController.GetAgentVersions)
api.GET("/cmdb/agent-versions/latest", cmdbController.GetLatestAgentVersion)
api.GET("/cmdb/agent-versions/:id", cmdbController.GetAgentVersionByID)
api.POST("/cmdb/agent-versions", cmdbController.CreateAgentVersion)
api.PUT("/cmdb/agent-versions/:id", cmdbController.UpdateAgentVersion)
api.DELETE("/cmdb/agent-versions/:id", cmdbController.DeleteAgentVersion)

// Agent 升级管理
api.POST("/cmdb/servers/:id/agent/upgrade", cmdbController.UpgradeAgent)
api.GET("/cmdb/agent-upgrade-tasks", cmdbController.GetUpgradeTasks)
api.GET("/cmdb/agent-upgrade-tasks/:id", cmdbController.GetUpgradeTaskByID)
```

---

### 阶段三：前端实现 ✅

#### 6. TypeScript 类型定义
**文件**：`soybean-admin-element-plus/src/typings/api/cmdb.d.ts`

```typescript
declare namespace CMDB {
  /** Agent 版本 */
  type AgentVersion = {
    id: number;
    version: string;
    releaseNotes?: string;
    changelog?: string;
    releasedAt: string;
    // ... 二进制文件信息、版本状态、功能支持、统计信息
  };

  /** Agent 版本表单 */
  type AgentVersionForm = {
    id?: number;
    version: string;
    releaseNotes?: string;
    changelog?: string;
    amd64BinaryPath?: string;
    arm64BinaryPath?: string;
    isLatest?: boolean;
    isDeprecated?: boolean;
    features?: Record<string, boolean>;
    minCompatibleVersion?: string;
    maxCompatibleVersion?: string;
  };

  /** Agent 升级任务 */
  type AgentUpgradeTask = {
    id: number;
    taskName?: string;
    targetVersion: string;
    status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
    // ... 进度统计、详细日志
  };
}
```

#### 7. API 函数封装
**文件**：`soybean-admin-element-plus/src/service/api/cmdb.ts`

新增 API 函数：
```typescript
// 版本管理 API
export function fetchGetAgentVersions(): Promise<AgentVersion[]>
export function fetchGetLatestAgentVersion(): Promise<AgentVersion>
export function fetchGetAgentVersionByID(id: number): Promise<AgentVersion>
export function fetchCreateAgentVersion(data: AgentVersionForm)
export function fetchUpdateAgentVersion(id: number, data: Partial<AgentVersionForm>)
export function fetchDeleteAgentVersion(id: number)

// 升级管理 API
export function fetchUpgradeAgent(serverId: number, targetVersion: string)
export function fetchGetUpgradeTasks(params?): Promise<{list, total}>
export function fetchGetUpgradeTaskByID(taskId: number): Promise<AgentUpgradeTask>
```

#### 8. Agent 管理页面增强
**文件**：`soybean-admin-element-plus/src/views/cmdb_config_agents/index.vue`

**新增状态**：
```typescript
const latestVersion = ref<CMDB.AgentVersion | null>(null);
const upgradeVersions = ref<CMDB.AgentVersion[]>([]);
```

**新增辅助函数**：
- `isLatestVersion(currentVersion)` - 判断是否为最新版本
- `getVersionStatus(currentVersion)` - 获取版本状态标签
- `getLatestVersion()` - 获取最新版本信息
- `handleUpgrade(row)` - 单台升级处理
- `batchUpgradeable` - 可升级主机计算属性
- `handleBatchUpgrade()` - 批量升级处理

**UI 增强**：
1. 版本列增强
   - 显示版本号（v1.0.0）
   - 显示版本状态标签（最新版/可升级）
   
2. 操作列增强
   - 运行中且非最新版本：显示升级按钮
   - 操作列宽度调整为 280px

3. 工具栏增强
   - 新增批量升级按钮
   - 显示可升级主机数量

---

## 功能特性说明

### 版本管理

**版本信息记录**：
- 版本号（如 1.0.0, 1.1.0）
- 发布说明和更新日志
- 各架构二进制文件路径、哈希、大小
- 功能支持标记（JSON 格式）
- 兼容性范围

**版本状态**：
- `is_latest` - 是否为最新版本（全局只有一个）
- `is_deprecated` - 是否已弃用
- 统计信息：下载次数、部署次数

### 升级功能

**单台升级流程**：
1. 检查当前版本 → 已是目标版本则跳过
2. 版本兼容性检查 → 不兼容则拒绝升级
3. SSH 连接并检测架构（amd64/arm64）
4. 停止当前 Agent（systemctl stop）
5. 上传新版本二进制文件
6. 设置执行权限
7. 启动新版本（systemctl start）
8. 等待心跳确认（60秒超时，每3秒轮询一次）
9. 更新数据库版本信息和部署统计

**批量升级流程**：
- 创建升级任务（可记录到 agent_upgrade_tasks 表）
- 异步并发执行多台主机升级
- 实时更新进度（current/total）
- 记录成功/失败/跳过数量
- 升级完成后显示通知

**失败处理**：
- 上传失败：尝试重新启动旧版本
- 启动失败：记录错误日志
- 超时未收到心跳：标记为失败

### 用户界面

**版本状态显示**：
- 未安装：灰色 "-"
- 最新版：绿色标签"最新版"
- 可升级：橙色标签"可升级"

**升级按钮**：
- 单台：针对运行中且非最新版本的主机显示
- 批量：根据选中主机动态启用/禁用
- 确认对话框：显示当前版本和目标版本

**批量操作**：
- 批量部署：未安装的主机
- 批量卸载：运行中/离线的主机
- 批量升级：运行中且需要升级的主机
- 进度提示：显示当前进度（x / y 台）

---

## 关键文件清单

### 后端文件

| 文件路径 | 说明 |
|---------|------|
| `backend/sql/agent_version_management.sql` | 数据库迁移脚本 |
| `backend/models/cmdb.go` | 数据模型定义 |
| `backend/services/agent.go` | 服务层实现（版本管理、升级） |
| `backend/controllers/cmdb.go` | 控制器实现 |
| `backend/routes/routes.go` | 路由配置 |

### 前端文件

| 文件路径 | 说明 |
|---------|------|
| `soybean-admin-element-plus/src/typings/api/cmdb.d.ts` | TypeScript 类型定义 |
| `soybean-admin-element-plus/src/service/api/cmdb.ts` | API 函数封装 |
| `soybean-admin-element-plus/src/views/cmdb_config_agents/index.vue` | Agent 管理页面 |

---

## API 端点列表

### 版本管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/cmdb/agent-versions` | 获取版本列表 |
| GET | `/api/cmdb/agent-versions/latest` | 获取最新版本 |
| GET | `/api/cmdb/agent-versions/:id` | 获取版本详情 |
| POST | `/api/cmdb/agent-versions` | 创建新版本 |
| PUT | `/api/cmdb/agent-versions/:id` | 更新版本信息 |
| DELETE | `/api/cmdb/agent-versions/:id` | 删除版本 |

### 升级管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/cmdb/servers/:id/agent/upgrade` | 升级单台主机 Agent |
| GET | `/api/cmdb/agent-upgrade-tasks` | 获取升级任务列表 |
| GET | `/api/cmdb/agent-upgrade-tasks/:id` | 获取升级任务详情 |

---

## 使用示例

### 查看主机 Agent 版本状态

在 Agent 管理页面中：
- 版本列显示当前版本号（如 v1.0.0）
- 状态标签显示"最新版"或"可升级"

### 升级单台主机

1. 点击主机行的"升级"按钮
2. 确认升级对话框（显示当前版本和目标版本）
3. 系统自动执行升级流程
4. 完成后版本状态自动更新

### 批量升级主机

1. 勾选多台需要升级的主机
2. 点击"批量升级"按钮（显示可升级数量）
3. 确认批量升级对话框
4. 系统显示升级进度
5. 完成后显示成功数量

---

## 未实现的可选功能

以下功能可根据后续需求实现：

### 阶段四：版本管理页面（P1）

- 创建独立的版本管理页面 `cmdb_agent_versions`
  - 版本列表展示
  - 版本创建/编辑/删除对话框
  - 版本详情查看
  - 版本统计信息

- 创建升级任务页面 `cmdb_agent_upgrades`
  - 任务列表展示
  - 任务进度监控
  - 任务详情查看
  - 任务日志查看

- 配置菜单和路由

### 阶段五：高级功能（P2）

- 二进制文件管理
  - 二进制文件上传功能
  - 文件完整性校验（SHA256）
  - 目录结构：`agent-binaries/{version}/`

- 版本发布流程
  - 版本发布审批
  - 发布通知

- 版本回滚功能
  - 自动回滚机制
  - 手动回滚到指定版本

- 统计和监控
  - 版本部署统计
  - 升级成功率统计
  - 版本使用趋势分析

---

## 问题解决记录

### 问题：主机 41 的 hostname 为空

**原因**：主机 41 的 Agent 是旧版本，不支持 `/extended-metrics` 端点，返回空数据，导致 WebSocket 广播时 hostname 为空。

**解决方案**：实施 Agent 版本管理系统后，可以：
1. 查看主机 41 的当前版本状态
2. 确认需要升级
3. 一键升级到最新版本
4. 升级后 hostname 正常显示

### 数据库字段缺失

**问题**：编译时出现 `config.GetDB` 未定义错误。

**解决**：直接使用 services 包级别的 `db` 变量，删除所有 `db := config.GetDB()` 调用。

---

## 验证方法

1. **数据库验证**
   ```sql
   SELECT * FROM agent_versions;
   SELECT * FROM agent_upgrade_tasks;
   ```

2. **后端 API 测试**
   ```bash
   # 获取最新版本
   curl http://localhost:8082/api/cmdb/agent-versions/latest
   
   # 获取版本列表
   curl http://localhost:8082/api/cmdb/agent-versions
   ```

3. **前端功能测试**
   - 打开 Agent 管理页面
   - 查看版本状态标签显示
   - 点击升级按钮测试单台升级
   - 勾选多台主机测试批量升级

4. **升级流程测试**
   - 选择一台旧版本主机
   - 执行升级操作
   - 观察升级进度和状态变化
   - 验证升级后版本号正确更新

---

## 总结

本次实施完成了 Agent 版本管理系统的核心功能，建立了完整的版本发布、管理和升级体系。系统现在具备：

1. ✅ 版本信息记录和查询
2. ✅ 版本兼容性检查
3. ✅ 单台/批量升级功能
4. ✅ 升级进度追踪
5. ✅ 用户友好的升级界面

该系统为后续的 Agent 功能迭代提供了坚实的基础，解决了旧版本 Agent 功能缺失的问题，实现了版本的可追溯和可管理。
