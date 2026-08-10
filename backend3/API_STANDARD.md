# OneOps API 统一规范

> 本文档定义所有后端接口的统一响应格式，供前端团队对接参考。

---

## 一、统一响应结构

所有接口（无论成功/失败）HTTP 状态码统一返回 **200**，业务结果通过 body 中的 `code` 字段判断。

### 1.1 成功响应

```json
{
  "code": 200,
  "success": true,
  "message": "操作成功",
  "data": {}
}
```

### 1.2 错误响应

```json
{
  "code": 400,
  "success": false,
  "message": "请求参数错误",
  "data": null
}
```

### 1.3 错误码定义

| code | 含义 | 说明 |
|------|------|------|
| 200 | 成功 | |
| 400 | 请求参数错误 | 参数校验失败、格式错误 |
| 401 | 未授权 | Token 无效或过期 |
| 403 | 禁止访问 | 权限不足 |
| 404 | 资源不存在 | |
| 500 | 服务器内部错误 | |

> HTTP 状态码统一为 200，前端通过 `response.data.code === 200` 判断成功。

---

## 二、统一分页格式

### 2.1 请求参数

所有分页接口统一使用以下查询参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码，从 1 开始 |
| `pageSize` | int | 10 | 每页条数，最大 100 |

**变更说明**（前端需改）：
- 原 `current` 参数 → 改为 `page`
- 原 `size` 参数 → 改为 `pageSize`

### 2.2 分页响应结构

分页接口的 `data` 统一为以下结构：

```json
{
  "code": 200,
  "success": true,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "pages": 10
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `list` | array | 当前页数据列表 |
| `total` | int | 总记录数 |
| `page` | int | 当前页码 |
| `pageSize` | int | 每页条数 |
| `pages` | int | 总页数 |

**变更说明**（前端需改）：
- 原 `records` 字段 → 改为 `list`
- 原 `current` 字段 → 改为 `page`
- 原 `size` 字段 → 改为 `pageSize`
- 分页字段位置统一在 `data` 内部（原来 k8s/cluster、k8s/permission 在根级别）

---

## 三、受影响的接口清单

### 3.1 分页参数变更（current/size → page/pageSize）

| 接口 | 方法 | 路径 | 原参数 | 新参数 |
|------|------|------|--------|--------|
| 用户列表 | GET | `/api/manage/user` | `current`, `size` | `page`, `pageSize` |
| 角色列表 | GET | `/api/manage/role` | `current`, `size` | `page`, `pageSize` |
| 应用列表 | GET | `/api/auth/applications` | `current`, `size` | `page`, `pageSize` |
| AuthUser 列表 | GET | `/api/auth/auth-users` | `current`, `size` | `page`, `pageSize` |
| 用户组列表 | GET | `/api/auth/group-bindings` | `current`, `size` | `page`, `pageSize` |
| 操作日志(授权) | GET | `/api/auth/operation-logs` | `current`, `size` | `page`, `pageSize` |
| 权限矩阵 | GET | `/api/auth/permission-matrix` | `current`, `size` | `page`, `pageSize` |
| 身份映射 | GET | `/api/auth/identity-mappings` | `current`, `size` | `page`, `pageSize` |

### 3.2 分页响应字段变更（records → list）

| 接口 | 方法 | 路径 | 原响应字段 | 新响应字段 |
|------|------|------|-----------|-----------|
| 用户列表 | GET | `/api/manage/user` | `records/current/size` | `list/page/pageSize/pages` |
| 角色列表 | GET | `/api/manage/role` | `records/current/size` | `list/page/pageSize/pages` |
| 应用列表 | GET | `/api/auth/applications` | `records/current/size` | `list/page/pageSize/pages` |
| AuthUser | GET | `/api/auth/auth-users` | `records/current/size` | `list/page/pageSize/pages` |
| 用户组 | GET | `/api/auth/group-bindings` | `records/current/size` | `list/page/pageSize/pages` |
| 操作日志(授权) | GET | `/api/auth/operation-logs` | `records/current/size` | `list/page/pageSize/pages` |
| 权限矩阵 | GET | `/api/auth/permission-matrix` | `records/current/size` | `list/page/pageSize/pages` |
| 身份映射 | GET | `/api/auth/identity-mappings` | `records/current/size` | `list/page/pageSize/pages` |

### 3.3 分页字段位置变更（根级 → data 内部）

| 接口 | 方法 | 路径 | 原格式 | 新格式 |
|------|------|------|--------|--------|
| K8s 集群列表 | GET | `/api/k8s/clusters` | `total/page/pageSize` 在 JSON 根级 | 移入 `data` 内部 |
| 集群权限列表 | GET | `/api/k8s/clusters/:id/permissions` | `total/page/pageSize` 在 JSON 根级 | 移入 `data` 内部 |

### 3.4 分页字段补全（不变，仅增加）

以下接口原来只有 `list/total`，现在补充 `page/pageSize/pages`（**向后兼容，前端可选使用**）：

| 接口 | 方法 | 路径 |
|------|------|------|
| 服务器列表 | GET | `/api/cmdb/servers` |
| Deployment 列表 | GET | `/api/k8s/:clusterId/deployments` |
| Pod 列表 | GET | `/api/k8s/:clusterId/pods` |
| Service 列表 | GET | `/api/k8s/:clusterId/services` |
| Ingress 列表 | GET | `/api/k8s/:clusterId/ingresses` |
| ConfigMap 列表 | GET | `/api/k8s/:clusterId/configmaps` |
| Secret 列表 | GET | `/api/k8s/:clusterId/secrets` |
| StatefulSet 列表 | GET | `/api/k8s/:clusterId/statefulsets` |
| Job 列表 | GET | `/api/k8s/:clusterId/jobs` |
| CronJob 列表 | GET | `/api/k8s/:clusterId/cronjobs` |

### 3.5 手写 gin.H 改为标准封装（无格式变化）

以下接口原来手写 `gin.H` 返回，现统一改用标准封装函数，**响应 JSON 格式不变**：

| 接口 | 路径 |
|------|------|
| WebSocket 终端相关 | `/api/k8s/terminal/*` |
| 诊断中心 | `/api/k8s/:clusterId/diagnostic/*` |
| 菜单树 | `/api/manage/menu/tree` |

---

## 四、前端适配 Checklist

```
[ ] 1. 分页请求参数：current → page, size → pageSize
[ ] 2. 分页响应取值：records → list, current → page, size → pageSize
[ ] 3. K8s 集群列表：分页字段从 root 移到 data 内部
[ ] 4. HTTP 拦截器：统一按 response.data.code === 200 判断成功
[ ] 5. 错误提示：统一取 response.data.message
```
