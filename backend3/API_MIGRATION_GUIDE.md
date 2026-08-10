# OneOps Backend3 API 改造文档 — 前端对接参考

## 一、响应体统一结构

### 1.1 通用响应格式（无变化）

所有接口统一返回以下结构：

```json
{
  "code": 200,
  "success": true,
  "data": {},
  "message": "success"
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `code` | int | 业务状态码（200=成功, 400=参数错误, 401=未授权, 403=禁止, 500=内部错误） |
| `success` | bool | 是否成功 |
| `data` | any | 业务数据（成功时存在） |
| `message` | string | 提示信息 |

### 1.2 成功响应示例

**普通数据**：
```json
{
  "code": 200,
  "success": true,
  "data": { "id": 1, "username": "admin" },
  "message": "success"
}
```

**仅提示**：
```json
{
  "code": 200,
  "success": true,
  "data": null,
  "message": "创建成功"
}
```

### 1.3 错误响应示例

```json
{
  "code": 400,
  "success": false,
  "message": "用户名不能为空"
}
```

---

## 二、分页响应格式变化（重要）

### 2.1 改造前（3种格式并存）

**格式A** — 部分接口：
```json
{
  "code": 200,
  "success": true,
  "data": {
    "records": [...],
    "current": 1,
    "size": 10,
    "total": 100,
    "pages": 10
  },
  "message": "success"
}
```

**格式B** — 部分接口：
```json
{
  "code": 200,
  "success": true,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "pageCount": 10
  },
  "message": "success"
}
```

**格式C** — 部分接口（缺少总页数）：
```json
{
  "code": 200,
  "success": true,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 10
  },
  "message": "success"
}
```

### 2.2 改造后（统一格式）

**所有分页接口统一为**：

```json
{
  "code": 200,
  "success": true,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "pages": 10,
    "pageCount": 10
  },
  "message": "success"
}
```

### 2.3 字段变更对照表

| 改造前字段 | 改造后字段 | 说明 | 兼容性 |
|---|---|---|---|
| `records` | `list` | 数据列表字段名统一 | **破坏性变更**：原使用 `records` 的前端需改为 `list` |
| `current` | `page` | 当前页码字段名统一 | **破坏性变更**：原使用 `current` 的前端需改为 `page` |
| `size` | `pageSize` | 每页大小字段名统一 | **破坏性变更**：原使用 `size` 的前端需改为 `pageSize` |
| `pageCount` | `pageCount` + `pages` | 同时输出两个字段 | **向后兼容**：原使用 `pageCount` 的代码无需修改，新增 `pages` 为推荐字段 |
| — | `pages` | 总页数（新增） | **新增字段**：格式C 的接口现在也有总页数了 |

### 2.4 前端改造建议

```javascript
// ❌ 改造前（可能存在）
const list = response.data.records      // 格式A
const list = response.data.list         // 格式B/C
const page = response.data.current      // 格式A
const page = response.data.page         // 格式B/C
const pageSize = response.data.size     // 格式A
const pageSize = response.data.pageSize // 格式B/C

// ✅ 改造后（统一）
const list = response.data.list
const page = response.data.page
const pageSize = response.data.pageSize
const total = response.data.total
const pages = response.data.pages       // 推荐（与 pageCount 相同）
```

### 2.5 受影响的接口列表

以下接口的分页响应格式会发生变化：

#### 系统管理模块

| 接口 | 方法 | 路径 | 变化说明 |
|---|---|---|---|
| 用户列表 | GET | `/api/v1/users` | 分页格式统一 |
| 角色列表 | GET | `/api/v1/roles` | 分页格式统一 |

#### CMDB 模块

| 接口 | 方法 | 路径 | 变化说明 |
|---|---|---|---|
| 服务器列表 | GET | `/api/v1/cmdb/servers` | 分页格式统一 |
| Agent列表 | GET | `/api/v1/cmdb/agents` | 分页格式统一 |
| 会话列表 | GET | `/api/v1/cmdb/sessions` | 分页格式统一 |
| 命令记录 | GET | `/api/v1/cmdb/commands` | 分页格式统一 |
| 文件传输 | GET | `/api/v1/cmdb/file-transfers` | 分页格式统一 |
| 访问策略 | GET | `/api/v1/cmdb/access-policies` | 分页格式统一 |

#### K8s 模块

| 接口 | 方法 | 路径 | 变化说明 |
|---|---|---|---|
| 集群列表 | GET | `/api/v1/k8s/clusters` | 分页格式统一 |
| 诊断历史 | GET | `/api/v1/k8s/diagnostic/history` | 分页格式统一 |

#### 授权中心模块

| 接口 | 方法 | 路径 | 变化说明 |
|---|---|---|---|
| 应用列表 | GET | `/api/v1/auth/applications` | 分页格式统一 |
| 授权用户列表 | GET | `/api/v1/auth/users` | 分页格式统一 |
| 用户组列表 | GET | `/api/v1/auth/groups` | 分页格式统一 |
| 操作日志 | GET | `/api/v1/auth/operation-logs` | 分页格式统一 |
| 身份映射 | GET | `/api/v1/auth/identity-mappings` | 分页格式统一 |
| 用户组绑定 | GET | `/api/v1/auth/group-bindings` | 分页格式统一 |
| 权限矩阵 | GET | `/api/v1/auth/permission-matrix` | 分页格式统一 |

#### 审计模块

| 接口 | 方法 | 路径 | 变化说明 |
|---|---|---|---|
| 登录日志 | GET | `/api/v1/audit/login-logs` | 分页格式统一 |
| 操作日志 | GET | `/api/v1/audit/operation-logs` | 分页格式统一 |

---

## 三、分页请求参数统一

### 3.1 请求参数格式

所有分页查询接口统一接受以下参数（Query String）：

| 参数 | 类型 | 默认值 | 最大值 | 说明 |
|---|---|---|---|---|
| `page` | int | 1 | — | 页码，从1开始 |
| `pageSize` | int | 10 | 100 | 每页条数 |

**示例**：
```
GET /api/v1/cmdb/servers?page=2&pageSize=20&hostname=web
```

### 3.2 废弃的参数名

| 废弃参数 | 替换为 | 说明 |
|---|---|---|
| `current` | `page` | 页码参数统一为 `page` |
| `size` | `pageSize` | 每页大小统一为 `pageSize` |

> 注意：后端 `page` 和 `current` 均可识别（兼容期），但前端应统一使用 `page`。

---

## 四、参数校验错误响应优化

### 4.1 改造前

```json
{
  "code": 400,
  "success": false,
  "message": "Key: 'ServerQueryParams.IP' Error:Field validation for 'IP' failed on the 'ip' tag"
}
```

暴露了内部结构体字段名和校验标签，对前端不友好。

### 4.2 改造后

```json
{
  "code": 400,
  "success": false,
  "message": "IP地址格式不正确"
}
```

返回中文友好提示，常见校验提示：

| 校验规则 | 提示示例 |
|---|---|
| required | 用户名不能为空 |
| min | 密码不能小于6 |
| max | 主机名不能大于100 |
| email | 邮箱格式不正确 |
| ip | IP地址格式不正确 |
| oneof | 状态值不合法 |

---

## 五、非分页接口（无变化）

以下接口的响应格式不变：

- 单条数据查询：`GET /api/v1/cmdb/servers/:id`
- 创建：`POST /api/v1/cmdb/servers` → `SuccessWithData(created)`
- 更新：`PUT /api/v1/cmdb/servers/:id` → `SuccessWithMessage("更新成功")`
- 删除：`DELETE /api/v1/cmdb/servers/:id` → `SuccessWithMessage("删除成功")`
- 登录：`POST /api/v1/auth/login` → `SuccessWithData(token)`
- K8s资源操作（非数据库分页，从K8s API直接获取）→ 格式不变

---

## 六、前端改造检查清单

### 必须修改（破坏性变更）

- [ ] 全局搜索 `response.data.records` → 替换为 `response.data.list`
- [ ] 全局搜索 `response.data.current` → 替换为 `response.data.page`
- [ ] 全局搜索 `response.data.size` → 替换为 `response.data.pageSize`
- [ ] 请求参数中 `current=xxx` → 替换为 `page=xxx`
- [ ] 请求参数中 `size=xxx` → 替换为 `pageSize=xxx`

### 建议优化（非必须）

- [ ] 总页数优先使用 `pages` 字段（`pageCount` 兼容保留）
- [ ] 分页组件统一使用 `list/page/pageSize/total/pages` 五个字段
- [ ] 错误提示直接展示 `message` 字段（现在是中文友好提示）

### 不需要修改

- [x] 非分页接口的响应格式不变
- [x] `code`/`success`/`data`/`message` 顶层结构不变
- [x] K8s 资源列表（Pod/Deployment/Service 等从 K8s API 获取的列表）格式不变
