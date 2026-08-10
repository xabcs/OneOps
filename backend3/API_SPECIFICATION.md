# OneOps Backend3 API 标准规范文档

> 版本：v3.0 | 更新日期：2026-08-10

---

## 一、概述

### 1.1 基础信息

| 项目 | 值 |
|---|---|
| Base URL | `/api` |
| 协议 | HTTP/HTTPS |
| 认证方式 | JWT Bearer Token |
| 字符编码 | UTF-8 |
| 时间格式 | ISO 8601（`2026-08-10T16:30:00+08:00`） |
| 数据交换格式 | JSON |

### 1.2 认证方式

除标注为"公开"的接口外，所有接口均需在请求头中携带 JWT Token：

```
Authorization: Bearer <token>
```

未携带或 Token 过期返回 `401`，权限不足返回 `403`。

---

## 二、统一响应规范

### 2.1 响应结构

所有接口统一返回以下 JSON 结构：

```json
{
  "code": 200,
  "success": true,
  "data": {},
  "message": "success"
}
```

| 字段 | 类型 | 必返回 | 说明 |
|---|---|---|---|
| `code` | int | 是 | 业务状态码 |
| `success` | bool | 是 | 是否成功 |
| `data` | any | 否 | 业务数据，失败时可能不存在 |
| `message` | string | 是 | 提示信息 |

### 2.2 状态码定义

| code | 含义 | 场景 |
|---|---|---|
| 200 | 成功 | 请求处理成功 |
| 400 | 参数错误 | 请求参数校验失败 |
| 401 | 未授权 | Token 缺失、过期或无效 |
| 403 | 禁止访问 | 权限不足 |
| 404 | 资源不存在 | 查询的目标不存在 |
| 500 | 内部错误 | 服务端异常 |
| 40001 | 主机名重复 | 创建/更新服务器时主机名冲突 |
| 40002 | IP重复 | 创建/更新服务器时IP冲突 |
| 40003 | 服务器不存在 | 操作的目标服务器不存在 |
| 40004 | SSH凭证无效 | 凭证认证失败 |

### 2.3 成功响应示例

**单条数据**：
```json
{
  "code": 200,
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "超级管理员"
  },
  "message": "success"
}
```

**操作确认**：
```json
{
  "code": 200,
  "success": true,
  "data": null,
  "message": "创建成功"
}
```

### 2.4 错误响应示例

```json
{
  "code": 400,
  "success": false,
  "message": "主机名不能为空"
}
```

---

## 三、统一分页规范

### 3.1 请求参数

所有分页查询接口统一接受以下 Query 参数：

| 参数 | 类型 | 默认值 | 约束 | 说明 |
|---|---|---|---|---|
| `page` | int | 1 | ≥ 1 | 页码，从1开始 |
| `pageSize` | int | 10 | 1 ~ 100 | 每页条数 |

**请求示例**：
```
GET /api/cmdb/servers?page=2&pageSize=20&hostname=web01&env=prod
```

### 3.2 响应格式

分页接口的 `data` 字段统一为：

```json
{
  "code": 200,
  "success": true,
  "data": {
    "list": [
      { "id": 1, "hostname": "web01" },
      { "id": 2, "hostname": "web02" }
    ],
    "total": 100,
    "page": 2,
    "pageSize": 20,
    "pages": 5,
    "pageCount": 5
  },
  "message": "success"
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `list` | array | 当前页数据列表 |
| `total` | int64 | 总记录数 |
| `page` | int | 当前页码 |
| `pageSize` | int | 每页条数 |
| `pages` | int64 | 总页数（推荐使用） |
| `pageCount` | int64 | 总页数（兼容字段，与 pages 值相同） |

---

## 四、统一参数校验规范

### 4.1 校验错误响应

参数校验失败时返回中文友好提示：

```json
{
  "code": 400,
  "success": false,
  "message": "邮箱格式不正确"
}
```

### 4.2 常见校验提示

| 校验规则 | 提示格式 | 示例 |
|---|---|---|
| required | {字段}不能为空 | 用户名不能为空 |
| min | {字段}不能小于{值} | 密码不能小于6 |
| max | {字段}不能大于{值} | 主机名不能大于100 |
| email | {字段}格式不正确 | 邮箱格式不正确 |
| ip | {字段}不是有效的IP地址 | IP地址不是有效的IP地址 |
| oneof | {字段}值不合法 | 状态值不合法 |
| alphanum | {字段}只能包含字母和数字 | 编码只能包含字母和数字 |

### 4.3 字段名中文映射

| 英文字段名 | 中文名 |
|---|---|
| Username | 用户名 |
| Password | 密码 |
| Email | 邮箱 |
| RealName | 姓名 |
| Phone | 手机号 |
| Hostname | 主机名 |
| IP | IP地址 |
| Name | 名称 |
| Code | 编码 |
| Path | 路径 |
| Namespace | 命名空间 |
| Description | 描述 |
| Status | 状态 |
| Permission | 权限 |

---

## 五、API 接口清单

### 5.1 认证模块 — `/api`

| 方法 | 路径 | 说明 | 认证 | 分页 |
|---|---|---|---|---|
| POST | `/auth/login` | 用户登录 | 公开 | — |
| POST | `/auth/logout` | 退出登录 | 需要 | — |
| GET | `/auth/captcha` | 获取验证码 | 公开 | — |
| GET | `/auth/routes` | 获取用户路由 | 需要 | — |
| GET | `/auth/info` | 获取当前用户信息 | 需要 | — |

---

### 5.2 系统管理 — `/api/system`

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/menus` | 菜单列表 | system.menu.list | — |
| GET | `/menus/tree` | 菜单树 | system.menu.list | — |
| POST | `/menus` | 创建菜单 | system.menu.create | — |
| PUT | `/menus/:id` | 更新菜单 | system.menu.update | — |
| DELETE | `/menus/:id` | 删除菜单 | system.menu.delete | — |
| GET | `/roles` | 角色列表 | system.role.list | ✅ |
| POST | `/roles` | 创建角色 | system.role.create | — |
| PUT | `/roles/:id` | 更新角色 | system.role.update | — |
| DELETE | `/roles/:id` | 删除角色 | system.role.delete | — |
| GET | `/users` | 用户列表 | system.user.list | ✅ |
| POST | `/users` | 创建用户 | system.user.create | — |
| PUT | `/users/:id` | 更新用户 | system.user.update | — |
| DELETE | `/users/:id` | 删除用户 | system.user.delete | — |
| PUT | `/users/:id/password` | 重置密码 | system.user.reset_password | — |
| GET | `/attributes` | 属性定义列表 | system.attribute.list | — |
| POST | `/attributes` | 创建属性定义 | system.attribute.create | — |
| PUT | `/attributes/:id` | 更新属性定义 | system.attribute.update | — |
| DELETE | `/attributes/:id` | 删除属性定义 | system.attribute.delete | — |
| GET | `/attributes/:id` | 属性定义详情 | system.attribute.view | — |
| GET | `/server-attributes/:serverId` | 主机属性 | system.attribute.view | — |
| POST | `/server-attributes/:serverId` | 保存主机属性 | system.attribute.update | — |
| GET | `/permissions` | 权限列表 | system.permission.list | — |
| GET | `/permissions/tree` | 权限树 | system.permission.list | — |
| POST | `/permissions` | 创建权限 | system.permission.create | — |
| PUT | `/permissions/:id` | 更新权限 | system.permission.update | — |
| DELETE | `/permissions/:id` | 删除权限 | system.permission.delete | — |
| POST | `/permissions/check` | 校验权限 | system.permission.list | — |
| GET | `/roles/:roleId/permissions` | 角色权限列表 | system.role.view | — |
| POST | `/roles/:roleId/permissions` | 分配角色权限 | system.role.assign_permissions | — |
| GET | `/user/permissions` | 当前用户权限 | — | — |

---

### 5.3 资产管理 — `/api/cmdb`

#### 5.3.1 服务器管理

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/servers` | 服务器列表 | cmdb.server.list | ✅ |
| POST | `/servers` | 创建服务器 | cmdb.server.create | — |
| PUT | `/servers/:id` | 更新服务器 | cmdb.server.update | — |
| DELETE | `/servers/:id` | 删除服务器 | cmdb.server.delete | — |
| GET | `/servers/stats` | 服务器统计 | cmdb.server.list | — |
| POST | `/servers/config` | 服务器配置 | cmdb.server.list | — |
| GET | `/servers/:id` | 服务器详情 | cmdb.server.view | — |
| GET | `/servers/:id/connect` | 连接信息 | cmdb.server.connect | — |
| POST | `/servers/:id/connect` | 建立连接 | cmdb.server.connect | — |
| GET | `/servers/:id/permission` | 连接权限检查 | cmdb.server.connect | — |
| POST | `/servers/:id/sync-metrics` | 同步指标 | cmdb.server.update | — |

#### 5.3.2 Agent 管理

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| POST | `/agent/heartbeat` | Agent心跳 | **公开** | — |
| GET | `/agents` | Agent列表 | cmdb.agents.list | ✅ |
| POST | `/agents/batch-deploy` | 批量部署 | cmdb.agents.deploy | — |
| POST | `/agents/batch-uninstall` | 批量卸载 | cmdb.agents.uninstall | — |
| DELETE | `/agents/:id` | 删除记录 | cmdb.agents.uninstall | — |
| POST | `/servers/:id/agent/deploy` | 部署Agent | cmdb.agents.deploy | — |
| POST | `/servers/:id/agent/restart` | 重启Agent | cmdb.agents.restart | — |
| POST | `/servers/:id/agent/uninstall` | 卸载Agent | cmdb.agents.uninstall | — |
| GET | `/servers/:id/agent/status` | Agent状态 | cmdb.agents.view | — |
| POST | `/servers/:id/test-connection` | 测试连接 | cmdb.server.connect | — |

#### 5.3.3 Agent 版本管理

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/agent-versions` | 版本列表 | cmdb.agents.list | ✅ |
| GET | `/agent-versions/latest` | 最新版本 | cmdb.agents.list | — |
| GET | `/agent-versions/:id` | 版本详情 | cmdb.agents.view | — |
| POST | `/agent-versions` | 创建版本 | cmdb.agents.deploy | — |
| PUT | `/agent-versions/:id` | 更新版本 | cmdb.agents.upgrade | — |
| DELETE | `/agent-versions/:id` | 删除版本 | cmdb.agents.uninstall | — |
| POST | `/servers/:id/agent/upgrade` | 升级Agent | cmdb.agents.upgrade | — |
| GET | `/agent-upgrade-tasks` | 升级任务列表 | cmdb.agents.list | ✅ |
| GET | `/agent-upgrade-tasks/:id` | 升级任务详情 | cmdb.agents.view | — |

#### 5.3.4 主机分组

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/groups` | 分组列表 | cmdb.group.list | — |
| GET | `/groups/:id` | 分组详情 | cmdb.group.view | — |
| GET | `/asset-tree` | 资产树 | cmdb.group.list | — |
| POST | `/groups` | 创建分组 | cmdb.group.create | — |
| PUT | `/groups/:id` | 更新分组 | cmdb.group.update | — |
| DELETE | `/groups/:id` | 删除分组 | cmdb.group.delete | — |
| POST | `/groups/assign` | 分配到分组 | cmdb.group.assign | — |
| POST | `/groups/assign-multi` | 多分组分配 | cmdb.group.assign | — |
| GET | `/group-servers/:groupId` | 分组下主机 | cmdb.group.view | — |

#### 5.3.5 业务系统

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/business-units` | 业务列表 | cmdb.business.list | — |
| POST | `/business-units` | 创建业务 | cmdb.business.create | — |
| PUT | `/business-units/:id` | 更新业务 | cmdb.business.update | — |
| DELETE | `/business-units/:id` | 删除业务 | cmdb.business.delete | — |

#### 5.3.6 机房机柜

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/rooms` | 机房列表 | cmdb.rooms.list | — |
| POST | `/rooms` | 创建机房 | cmdb.rooms.create | — |
| PUT | `/rooms/:id` | 更新机房 | cmdb.rooms.update | — |
| DELETE | `/rooms/:id` | 删除机房 | cmdb.rooms.delete | — |
| GET | `/cabinets` | 机柜列表 | cmdb.rooms.list | — |

#### 5.3.7 标签管理

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/tags` | 标签列表 | cmdb.tags.list | — |
| POST | `/tags` | 创建标签 | cmdb.tags.create | — |
| PUT | `/tags/:id` | 更新标签 | cmdb.tags.update | — |
| DELETE | `/tags/:id` | 删除标签 | cmdb.tags.delete | — |
| POST | `/tags/assign` | 分配标签 | cmdb.tags.update | — |
| DELETE | `/server-tags/:serverId/:tagId` | 移除标签 | cmdb.tags.delete | — |

#### 5.3.8 SSH凭证

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/ssh-credentials` | 凭证列表 | cmdb.server.list | — |
| GET | `/ssh-credentials/:id` | 凭证详情 | cmdb.server.view | — |
| POST | `/ssh-credentials` | 创建凭证 | cmdb.server.create | — |
| PUT | `/ssh-credentials/:id` | 更新凭证 | cmdb.server.update | — |
| DELETE | `/ssh-credentials/:id` | 删除凭证 | cmdb.server.delete | — |
| POST | `/ssh-credentials/:id/test` | 测试凭证 | cmdb.server.connect | — |

#### 5.3.9 堡垒机会话

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/sessions` | 会话列表 | cmdb.session.list | ✅ |
| GET | `/sessions/list` | 会话列表(轻量) | cmdb.session.list | ✅ |
| GET | `/sessions/active` | 活跃会话 | cmdb.session.list | — |
| GET | `/sessions/active-memory` | 内存活跃会话 | cmdb.session.list | — |
| GET | `/sessions/stats` | 会话统计 | cmdb.session.list | — |
| GET | `/sessions/:id` | 会话详情 | cmdb.session.view | — |
| POST | `/sessions/:id/terminate` | 终止会话 | cmdb.session.terminate | — |
| GET | `/sessions/:id/commands` | 会话命令 | cmdb.session.view | ✅ |
| GET | `/sessions/:id/file-transfers` | 文件传输 | cmdb.session.view | ✅ |
| POST | `/sessions/:id/resize` | 调整终端 | cmdb.session.view | — |

#### 5.3.10 命令与文件审计

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/commands` | 命令记录 | cmdb.session.list | ✅ |
| GET | `/file-transfers` | 文件传输记录 | cmdb.session.list | ✅ |

#### 5.3.11 访问策略

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/access-policies` | 策略列表 | cmdb.access_policy.list | ✅ |
| GET | `/access-policies/:id` | 策略详情 | cmdb.access_policy.list | — |
| POST | `/access-policies` | 创建策略 | cmdb.access_policy.create | — |
| PUT | `/access-policies/:id` | 更新策略 | cmdb.access_policy.update | — |
| DELETE | `/access-policies/:id` | 删除策略 | cmdb.access_policy.delete | — |

#### 5.3.12 资产变更

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/asset-changes` | 变更记录 | cmdb.server.list | ✅ |

---

### 5.4 K8s管理 — `/api/k8s`

#### 5.4.1 集群管理

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/clusters` | 集群列表 | k8s.cluster.list | ✅ |
| POST | `/clusters` | 创建集群 | k8s.cluster.create | — |
| PUT | `/clusters/:id` | 更新集群 | k8s.cluster.update | — |
| DELETE | `/clusters/:id` | 删除集群 | k8s.cluster.delete | — |
| GET | `/clusters/:id` | 集群详情 | k8s.cluster.view | — |
| POST | `/clusters/:id/test` | 测试连接 | k8s.cluster.connect | — |

#### 5.4.2 K8s权限

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/clusters/:clusterId/permissions` | 权限列表 | k8s.permission.list | — |
| POST | `/clusters/:clusterId/permissions` | 分配权限 | k8s.permission.assign | — |
| DELETE | `/clusters/:clusterId/permissions/:id` | 撤销权限 | k8s.permission.revoke | — |

#### 5.4.3 K8s资源操作

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/clusters/:clusterId/namespaces/:namespace/deployments` | Deployment列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/deployments/:name` | Deployment详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/deployments` | 创建Deployment | k8s.resource.create | — |
| PUT | `/clusters/:clusterId/namespaces/:namespace/deployments/:name` | 更新Deployment | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/deployments/:name` | 删除Deployment | k8s.resource.delete | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/deployments/:name/restart` | 重启Deployment | k8s.resource.update | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/deployments/:name/scale` | 扩缩容 | k8s.resource.update | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/statefulsets` | StatefulSet列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/statefulsets/:name` | StatefulSet详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/statefulsets/:name/restart` | 重启StatefulSet | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/statefulsets/:name` | 删除StatefulSet | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/daemonsets` | DaemonSet列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/daemonsets/:name` | DaemonSet详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/daemonsets/:name/restart` | 重启DaemonSet | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/daemonsets/:name` | 删除DaemonSet | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/pods` | Pod列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/pods/:name` | Pod详情 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/pods/:name/logs` | Pod日志 | k8s.resource.view | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/pods/:name` | 删除Pod | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/services` | Service列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/services/:name` | Service详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/services` | 创建Service | k8s.resource.create | — |
| PUT | `/clusters/:clusterId/namespaces/:namespace/services/:name` | 更新Service | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/services/:name` | 删除Service | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/ingresses` | Ingress列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/ingresses/:name` | Ingress详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/ingresses` | 创建Ingress | k8s.resource.create | — |
| PUT | `/clusters/:clusterId/namespaces/:namespace/ingresses/:name` | 更新Ingress | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/ingresses/:name` | 删除Ingress | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/configmaps` | ConfigMap列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/configmaps/:name` | ConfigMap详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/configmaps` | 创建ConfigMap | k8s.resource.create | — |
| PUT | `/clusters/:clusterId/namespaces/:namespace/configmaps/:name` | 更新ConfigMap | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/configmaps/:name` | 删除ConfigMap | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/secrets` | Secret列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/secrets/:name` | Secret详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/secrets` | 创建Secret | k8s.resource.create | — |
| PUT | `/clusters/:clusterId/namespaces/:namespace/secrets/:name` | 更新Secret | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/secrets/:name` | 删除Secret | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/jobs` | Job列表 | k8s.resource.view | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/jobs/:name` | 删除Job | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/cronjobs` | CronJob列表 | k8s.resource.view | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/cronjobs/:name` | CronJob详情 | k8s.resource.view | — |
| POST | `/clusters/:clusterId/namespaces/:namespace/cronjobs/:name/suspend` | 挂起CronJob | k8s.resource.update | — |
| DELETE | `/clusters/:clusterId/namespaces/:namespace/cronjobs/:name` | 删除CronJob | k8s.resource.delete | — |
| GET | `/clusters/:clusterId/namespaces/:namespace/events` | Event列表 | k8s.resource.view | — |

#### 5.4.4 K8s诊断

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/clusters/:clusterId/diagnostic/commands` | 诊断命令 | k8s.diagnostic.execute | — |
| GET | `/clusters/:clusterId/diagnostic/namespaces` | 诊断命名空间 | k8s.diagnostic.execute | — |
| GET | `/clusters/:clusterId/diagnostic/java-pods` | Java Pod列表 | k8s.diagnostic.execute | — |
| POST | `/clusters/:clusterId/diagnostic/execute` | 执行诊断 | k8s.diagnostic.execute | — |
| GET | `/clusters/:clusterId/diagnostic/history` | 诊断历史 | k8s.diagnostic.view | ✅ |

#### 5.4.5 K8s终端

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/terminal` | WebSocket终端 | k8s.cluster.connect | — |

---

### 5.5 审计中心 — `/api/audit`

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/login-logs` | 登录日志 | audit.login_log.list | ✅ |
| GET | `/login-logs/export` | 导出登录日志 | audit.login_log.export | — |
| GET | `/operation-logs` | 操作日志 | audit.operation_log.list | ✅ |
| GET | `/operation-logs/export` | 导出操作日志 | audit.operation_log.export | — |
| GET | `/system-events` | 系统事件 | audit.system_event.list | ✅ |
| GET | `/stats` | 审计统计 | audit.stats.view | — |

---

### 5.6 监控中心 — `/api/monitoring`

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/overview` | 监控概览 | monitor.data.view | — |
| GET | `/servers` | 主机监控列表 | monitor.data.view | — |
| GET | `/servers/:id` | 主机监控详情 | monitor.data.view | — |
| GET | `/servers/:id/inspect` | 主机巡检 | monitor.data.view | — |
| POST | `/data` | Agent数据上报 | **公开** | — |
| GET | `/alerts` | 告警列表 | monitor.alert.list | ✅ |
| GET | `/alerts/rules` | 告警规则列表 | monitor.alert.list | — |
| POST | `/alerts/rules` | 创建告警规则 | monitor.alert.list | — |
| PUT | `/alerts/rules/:id` | 更新告警规则 | monitor.alert.list | — |
| DELETE | `/alerts/rules/:id` | 删除告警规则 | monitor.alert.list | — |
| POST | `/alerts/:id/ack` | 确认告警 | monitor.alert.ack | — |
| POST | `/reports/inspection` | 创建巡检报告 | monitor.report.create | — |
| GET | `/reports` | 报告列表 | monitor.report.list | ✅ |
| GET | `/reports/:id` | 报告详情 | monitor.report.view | — |
| DELETE | `/reports/:id` | 删除报告 | monitor.report.delete | — |
| GET | `/notification-channels` | 通知渠道列表 | monitor.data.view | — |
| POST | `/notification-channels` | 创建通知渠道 | monitor.data.view | — |
| PUT | `/notification-channels/:id` | 更新通知渠道 | monitor.data.view | — |
| DELETE | `/notification-channels/:id` | 删除通知渠道 | monitor.data.view | — |
| GET | `/ws` | WebSocket实时推送 | monitor.data.view | — |

---

### 5.7 授权中心 — `/api/auth`

| 方法 | 路径 | 说明 | 权限码 | 分页 |
|---|---|---|---|---|
| GET | `/applications` | 应用列表 | auth.application.list | ✅ |
| POST | `/applications` | 创建应用 | auth.application.create | — |
| PUT | `/applications/:id` | 更新应用 | auth.application.update | — |
| DELETE | `/applications/:id` | 删除应用 | auth.application.delete | — |
| POST | `/applications/:id/sync` | 同步应用数据 | auth.application.sync | — |
| GET | `/applications/:id/roles` | 应用角色列表 | auth.application.view | — |
| GET | `/users` | 授权用户列表 | auth.user.list | ✅ |
| POST | `/users` | 创建授权用户 | auth.user.create | — |
| PUT | `/users/:id` | 更新授权用户 | auth.user.update | — |
| DELETE | `/users/:id` | 删除授权用户 | auth.user.delete | — |
| GET | `/groups` | 用户组列表 | auth.group.list | ✅ |
| POST | `/groups` | 创建用户组 | auth.group.create | — |
| PUT | `/groups/:id` | 更新用户组 | auth.group.update | — |
| DELETE | `/groups/:id` | 删除用户组 | auth.group.delete | — |
| GET | `/group-bindings` | 用户组绑定列表 | auth.group_binding.list | ✅ |
| POST | `/group-bindings` | 创建绑定 | auth.group_binding.create | — |
| DELETE | `/group-bindings/:id` | 删除绑定 | auth.group_binding.delete | — |
| POST | `/group-members` | 添加组成员 | auth.user_group.assign | — |
| DELETE | `/group-members` | 移除组成员 | auth.user_group.delete | — |
| GET | `/operation-logs` | 操作日志 | auth.application.list | ✅ |
| GET | `/identity-mappings` | 身份映射列表 | auth.identity_mapping.list | ✅ |
| DELETE | `/identity-mappings/:id` | 删除身份映射 | auth.identity_mapping.delete | — |
| GET | `/permission-matrix` | 权限矩阵 | auth.permission.view | ✅ |
| GET | `/user-permissions/:userId` | 用户有效权限 | auth.permission.query | — |
| GET | `/users/:id/initial-password` | 用户初始密码 | auth.user.view | — |

---

## 六、请求体规范

### 6.1 Content-Type

| 场景 | Content-Type |
|---|---|
| 创建/更新资源 | `application/json` |
| 文件上传 | `multipart/form-data` |
| 查询参数 | Query String（GET 请求） |

### 6.2 通用字段约定

| 规则 | 说明 | 示例 |
|---|---|---|
| ID使用 uint | 所有资源ID为正整数 | `"id": 1` |
| 时间使用 ISO 8601 | 创建/更新时间 | `"created_at": "2026-08-10T16:30:00+08:00"` |
| 状态使用枚举字符串 | 不使用数字编码 | `"status": "active"` |
| 可选字段省略 | 不传等同于默认值 | 更新请求中只传需要修改的字段 |
| 列表字段使用数组 | 不使用逗号分隔字符串 | `"roleIds": [1, 2, 3]` |

### 6.3 创建 vs 更新请求差异

**创建请求**：必填字段使用值类型，`binding:"required"`

```json
{
  "hostname": "web01",
  "ip": "192.168.1.1",
  "sshPort": 22
}
```

**更新请求**：可选字段使用指针类型，仅传需要修改的字段

```json
{
  "hostname": "web01-new",
  "status": "inactive"
}
```

---

## 七、WebSocket 接口

### 7.1 K8s终端

```
ws://<host>/api/k8s/terminal?token=<jwt>&clusterId=1&namespace=default&pod=name&container=main&shell=bash
```

### 7.2 监控实时推送

```
ws://<host>/api/monitoring/ws?token=<jwt>
```

---

## 八、错误处理最佳实践

### 8.1 前端判断逻辑

```javascript
// 统一判断方式
if (response.data.success) {
    // 成功
    const data = response.data.data;
} else {
    // 失败
    const errorMsg = response.data.message;
    const errorCode = response.data.code;
}
```

### 8.2 常见错误码处理

| code | 处理方式 |
|---|---|
| 401 | 跳转登录页，清除本地 Token |
| 403 | 提示"权限不足"，显示联系管理员信息 |
| 40001 | 提示"主机名已存在"，让用户修改主机名 |
| 40002 | 提示"IP地址已存在"，让用户修改IP |
| 500 | 提示"系统异常"，建议刷新或联系管理员 |
