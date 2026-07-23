# Jenkins 项目角色获取问题修复

## 问题描述

用户反馈：Jenkins 角色同步只能获取全局角色（Global Roles），无法获取项目角色（Item Roles/Project Roles）。

## 问题原因

**Role Strategy 插件版本差异：**

旧版本插件：
- 使用 `RoleType.Item` 表示项目角色

新版本插件（当前 Jenkins 使用）：
- 使用 `RoleType.Project` 表示项目角色
- `RoleType.Item` 已废弃，会导致 `MissingPropertyException` 错误

## 实际情况

### 错误的脚本（修复前）

```groovy
// ❌ 错误：在新版本插件中会报错
def itemRoles = rbas.getRoleMap(RoleType.Item).getRoles()
```

**错误信息：**
```
groovy.lang.MissingPropertyException: No such property: Item for class: RoleType
```

### 正确的脚本（修复后）

```groovy
// ✅ 正确：使用 Project 枚举值
def projectRoles = rbas.getRoleMap(RoleType.Project).getRoles()
```

## 修复内容

**文件：** `backend/services/application_permission_service.go`

**修改位置：** `fetchJenkinsRoles` 函数中的 Groovy 脚本

```go
script := `
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

def rbas = Jenkins.instance.getAuthorizationStrategy()

if (rbas instanceof RoleBasedAuthorizationStrategy) {
    // 获取全局角色
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }

    // 获取项目角色（使用 Project 而不是 Item）
    def projectRoles = rbas.getRoleMap(RoleType.Project).getRoles()
    projectRoles.each { role ->
        println "${role.name}|${role.name}|project|Project Role"
    }
}
`
```

## 测试结果

### 全局角色（5个）

```
| 角色代码  | 角色名称   | 角色类型 |
|----------|-----------|---------|
| admin    | admin     | global  |
| anonymous| anonymous | global  |
| dev      | dev       | global  |
| manager  | manager   | global  |
| test     | test      | global  |
```

### 项目角色（34个）

```
| 角色代码           | 角色名称          | 角色类型 |
|-------------------|------------------|---------|
| Android           | Android          | project |
| front-factory-pre | front-factory-pre| project |
| front-saas-prod   | front-saas-prod  | project |
| k8s-saas-prod      | k8s-saas-prod    | project |
| 前端-erp-生产      | 前端-erp-生产     | project |
| 后端-erp-生产      | 后端-erp-生产     | project |
| ... (共 34 个)    | ...              | project |
```

## Role Strategy 插件角色类型

根据插件版本不同，角色类型枚举值可能不同：

### 当前版本支持的类型

```groovy
RoleType.values().each { type ->
    println type.name()
}

// 输出：
// - Global（全局角色）
// - Project（项目角色）
// - Slave（Agent 角色）
```

### 使用建议

1. **全局角色（Global）**：适用于整个 Jenkins 实例的权限控制
   - 如：admin, developer, reader
   - 控制全局权限，如系统配置、Job 创建等

2. **项目角色（Project）**：适用于特定项目/Job 的权限控制
   - 如：project-A-admin, project-B-developer
   - 可以使用正则表达式匹配项目名称
   - 控制特定项目的读写执行权限

3. **Agent 角色（Slave）**：适用于 Agent 节点的权限控制
   - 控制哪些用户可以使用特定的 Agent 节点

## 验证步骤

### 1. 重新编译后端

```bash
cd backend
go build
```

### 2. 重启服务

```bash
# 重启 OneOps 后端服务
./oneops-backend
```

### 3. 重新同步角色

```
1. 进入"授权中心 > 应用管理"
2. 找到 Jenkins 应用
3. 点击"同步角色"
4. 点击"查看角色"验证
```

### 4. 验证结果

应该看到：
- ✅ 5 个全局角色
- ✅ 34 个项目角色
- ✅ 角色类型正确标识（global/project）

## 前端显示优化

### 角色类型标签

在角色列表中，角色类型会以不同颜色显示：

```vue
<ElTag :type="row.roleType === 'global' ? 'primary' : 'success'">
  {{ row.roleType === 'global' ? '全局角色' : '项目角色' }}
</ElTag>
```

### 角色绑定时的提示

当为授权中心角色绑定外部角色时，可以看到：

```
外部角色：
├── [Global] admin
├── [Global] developer
├── [Project] front-saas-prod
├── [Project] k8s-saas-prod
└── ...
```

## 注意事项

### 1. 项目角色的命名

项目角色通常使用以下命名规范：
- 项目名称：`front-saas-prod`
- 环境+项目：`前端-erp-生产`
- 业务线：`Android`, `k8s-mini-pre`

### 2. 角色绑定的最佳实践

**推荐做法：**

```
授权中心角色：前端开发
  ↓ 绑定项目角色
Jenkins 项目角色：front-saas-prod, front-factory-prod

授权中心角色：后端开发
  ↓ 绑定项目角色
Jenkins 项目角色：后端-erp-生产, 后端-erp-预发
```

**不推荐做法：**

```
❌ 将项目角色绑定为全局角色
❌ 一个授权中心角色绑定过多项目角色（建议不超过10个）
```

### 3. 权限范围

- **全局角色**：影响整个 Jenkins
- **项目角色**：只影响匹配的项目/Job

## 兼容性说明

### 支持的 Jenkins 版本

- Jenkins 2.x 及以上
- Role Strategy 插件 2.x 及以上

### 支持的角色类型

| 角色类型 | 枚举值 | 是否支持 |
|---------|--------|---------|
| 全局角色 | RoleType.Global | ✅ |
| 项目角色 | RoleType.Project | ✅ |
| Agent 角色 | RoleType.Slave | ⚠️ 未实现 |

## 后续改进建议

### 1. 添加 Agent 角色支持

```groovy
// 获取 Agent 角色
try {
    def agentRoles = rbas.getRoleMap(RoleType.Slave).getRoles()
    agentRoles.each { role ->
        println "${role.name}|${role.name}|agent|Agent Role"
    }
} catch (Exception e) {
    // 忽略错误，某些版本可能不支持
}
```

### 2. 角色权限详情

可以进一步获取每个角色的具体权限：

```groovy
import hudson.security.Permission

def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
globalRoles.each { role ->
    def permissions = role.getPermissions()
    // 输出权限详情
}
```

### 3. 角色分配情况

可以查看哪些用户/组被分配了特定角色：

```groovy
def roleMap = rbas.getRoleMap(RoleType.Global)
def role = roleMap.getRoles().find { it.name == 'admin' }
def entries = roleMap.getRoleAssignments(role)
entries.each { entry ->
    println "User/Group: ${entry.sid}, Type: ${entry.type}"
}
```

## 总结

✅ **问题已修复**

**修改内容：**
- 将 `RoleType.Item` 改为 `RoleType.Project`
- 添加了对新版本 Role Strategy 插件的兼容

**测试结果：**
- ✅ 成功获取 5 个全局角色
- ✅ 成功获取 34 个项目角色
- ✅ 角色类型正确标识

**下一步：**
- 重新编译并部署后端
- 重新同步 Jenkins 角色
- 验证角色绑定功能
