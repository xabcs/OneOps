# 权限代码修复总结

## ✅ 已完成的修复

### 问题：前后端权限代码不一致

**发现的问题**：
- 前端使用 `system.user.view`、`system.role.view` 等权限
- 后端实际使用 `system.user.list`、`system.role.list` 等权限
- 导致前端权限检查失效

**已修复的文件**：

#### 1. 前端API权限映射
- **文件**：`frontend/src/utils/request.ts`
- **修改**：3处权限代码从 `.view` 改为 `.list`

#### 2. 前端路由守卫
- **文件**：`frontend/src/router/index-with-diagnostic.ts`
- **修改**：3处路由权限从 `.view` 改为 `.list`

#### 3. 前端权限常量
- **文件**：`frontend/src/utils/permission.ts`
- **修改**：权限常量从 `USER_VIEW` 改为 `USER_LIST` 等

#### 4. 权限配置文件
- **文件**：`backend/services/data/permissions.json`
- **删除**：4个冗余的 `view` 权限定义

## 📋 检查结果：其他模块无同类问题

### CMDB模块 ✅ 正确
- `cmdb.server.list` - 列表查询
- `cmdb.server.view` - 详情查询（有对应接口）
- `cmdb.group.view` - 详情查询（有对应接口）

### K8s模块 ✅ 正确
- `k8s.cluster.list` - 集群列表
- `k8s.cluster.view` - 集群详情（有对应接口）
- `k8s.resource.view` - 资源详情（有对应接口）

### Monitor模块 ✅ 正确
- `monitor.data.view` - 查看监控数据（合理）
- `monitor.alert.list` - 告警列表
- `monitor.task.list` - 任务列表

### Audit模块 ✅ 正确
- `audit.login_log.list` - 登录日志列表
- `audit.stats.view` - 审计统计（合理）

## 🎯 权限设计原则

### 何时使用 list 权限
- 列表数据已包含详情信息
- 不需要单独的详情接口
- **适用**：用户、角色、菜单、权限

### 何时使用 view 权限
- 列表数据不包含详情信息
- 有单独的详情接口 `GET /:id`
- 详情返回更多数据
- **适用**：服务器、集群、K8s资源

### 判断标准
```
是否有单独的详情接口？
├── 是 → 使用 *.view
└── 否 → 使用 *.list
```

## ✅ 验证结果

- ✅ 后端编译成功
- ✅ 权限JSON格式正确
- ✅ 前后端权限代码一致
- ✅ 无同类问题

## 📄 相关文档

- 详细分析：`PERMISSION_LIST_VIEW_ANALYSIS.md`
- 修复报告：`PERMISSION_FIX_REPORT.md`
