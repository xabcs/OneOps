# Jumpserver 用户有效权限视图需求梳理

## 当前架构分析

### 数据模型

#### 1. **group_bindings** 表（用户组-应用-角色绑定）
```
用户组 -> 应用 -> 应用角色
```

#### 2. **application_authorization_rules** 表（应用授权规则）
Jumpserver 的授权规则配置，可能包含：
- 资产访问权限
- 节点权限
- 动作权限（查看、连接、文件传输等）

#### 3. **application_roles** 表（应用角色）
Jumpserver 的角色定义

### 与 Jenkins 的区别

| 特性 | Jenkins | Jumpserver |
|-----|---------|------------|
| **权限粒度** | 项目级别 | 资产/节点级别 |
| **权限类型** | Global/Project 角色 | 授权规则 |
| **权限内容** | 项目构建权限 | 资产访问权限、命令执行权限 |
| **矩阵维度** | 用户 × 角色 | 用户 × 授权规则 |

## 需求分析

### 1. **矩阵视图设计**

#### Jenkins 矩阵（已实现）
- **横轴**：用户列表
- **纵轴**：角色列表（Global + Project）
- **单元格**：是否有权限（✓/✗）

#### Jumpserver 矩阵（待实现）
- **横轴**：授权规则列表
- **纵轴**：用户列表
- **单元格**：是否有权限（✓/✗）

### 2. **授权规则内容**

Jumpserver 的授权规则应包含：
- **规则名称**：如"生产服务器访问权限"、"测试服务器访问权限"
- **资产数量**：该规则包含多少资产
- **节点数量**：该规则包含多少节点
- **权限类型**：查看、连接、上传、下载、命令执行等
- **有效期**：权限的有效时间范围

### 3. **数据查询需求**

#### 查询用户在 Jumpserver 的有效权限

```sql
SELECT
  用户信息,
  授权规则名称,
  资产范围,
  权限类型,
  有效期,
  来源用户组
FROM 用户
JOIN 用户组
JOIN 用户组绑定
JOIN 授权规则
WHERE 应用 = Jumpserver
```

## 实现方案

### 后端 API

#### 1. **获取 Jumpserver 授权规则列表**
```go
GET /api/v1/applications/{appId}/authorization-rules
```

返回：
```json
{
  "rules": [
    {
      "id": 1,
      "name": "生产服务器访问权限",
      "assets_count": 50,
      "nodes_count": 5,
      "permissions": ["view", "connect", "upload", "download"],
      "description": "访问生产环境服务器"
    }
  ]
}
```

#### 2. **获取 Jumpserver 用户-授权规则矩阵**
```go
GET /api/v1/system/user-permissions/matrix?appId=7
```

返回：
```json
{
  "view_type": "user-rule-matrix",
  "app_id": 7,
  "app_name": "jumpserver",
  "rules": [
    {
      "id": 1,
      "name": "生产服务器访问权限",
      "assets_count": 50,
      "nodes_count": 5
    }
  ],
  "users": [
    {
      "id": 7,
      "username": "mtest",
      "external_username": "mtest_jump"
    }
  ],
  "matrix": {
    "7": {
      "1": true
    }
  },
  "permissions_detail": {
    "7_1": {
      "status": "active",
      "permissions": ["view", "connect"],
      "group_name": "运维组",
      "assigned_at": "2026-07-20T10:00:00+08:00"
    }
  }
}
```

### 前端组件

#### JumpserverMatrix.vue

```vue
<template>
  <div class="jumpserver-matrix">
    <!-- 统计信息 -->
    <div>授权规则: X 条，用户: Y 个</div>

    <!-- 矩阵表格 -->
    <PermissionMatrix
      :rows="users"
      :columns="rules"
      :matrix="matrix"
      row-key="id"
      col-key="id"
      row-label="username"
      col-label="name"
    />
  </div>
</template>
```

### 数据结构映射

#### 授权规则 -> 矩阵列
- ID -> col-key
- 名称 -> 显示标签
- 资产数量/节点数量 -> 副标签

#### 用户 -> 矩阵行
- ID -> row-key
- 用户名 -> 显示标签
- 外部用户名 -> 副标签

#### 权限数据 -> 矩阵单元格
- 用户有该授权规则 -> true (✓)
- 用户无该授权规则 -> false (✗)

## 实现步骤

### Phase 1: 后端数据查询
1. ✅ 查询 `application_authorization_rules` 表，获取 Jumpserver 的授权规则
2. ✅ 查询 `group_bindings` 表，获取用户组与授权规则的绑定关系
3. ✅ 查询 `auth_user_groups` 表，获取用户与用户组的关系
4. ✅ 构建用户-授权规则矩阵数据结构

### Phase 2: 后端 API 实现
1. ⬜ 实现 `getUserRuleMatrix` 方法（在 `application_permission_service.go`）
2. ⬜ 查询授权规则列表
3. ⬜ 查询用户列表
4. ⬜ 构建矩阵数据

### Phase 3: 前端组件
1. ✅ 创建 `JumpserverMatrix.vue` 组件（已创建占位）
2. ⬜ 实现授权规则显示
3. ⬜ 实现资产数量/节点数量显示
4. ⬜ 实现权限详情显示

### Phase 4: 集成测试
1. ⬜ 测试 API 返回数据
2. ⬜ 测试前端矩阵渲染
3. ⬜ 测试权限详情显示

## 待确认问题

### 1. 授权规则数据结构
**问题**：`application_authorization_rules` 表的具体结构是什么？
**需要**：
- 表结构定义
- 示例数据
- 与 `group_bindings` 的关联关系

### 2. 授权规则如何绑定到用户组？
**问题**：用户组如何获得授权规则权限？
**可能方案**：
- `group_bindings` 中存储授权规则 ID
- 单独的 `group_binding_rules` 关联表
- `application_authorization_rules` 中存储用户组 ID

### 3. 权限详情包含哪些字段？
**问题**：权限详情应该显示什么信息？
**可能包含**：
- 授权的资产列表
- 授权的节点列表
- 具体的权限动作（查看、连接、上传等）
- 有效期
- 来源用户组

### 4. Jumpserver 外部用户如何管理？
**问题**：Jumpserver 中是否有外部用户概念？
**需要确认**：
- Jumpserver 是否需要在创建权限时创建外部用户
- 外部用户与授权中心用户的映射关系

## 下一步行动

1. **确认数据模型**：查看 `application_authorization_rules` 表结构和数据
2. **确认关联关系**：授权规则如何绑定到用户组
3. **实现后端查询**：编写 SQL 查询构建矩阵数据
4. **实现前端组件**：完成 JumpserverMatrix 组件
5. **集成测试**：验证完整流程

---

**关键差异总结**：
- Jenkins：用户 × 角色 矩阵
- Jumpserver：用户 × 授权规则 矩阵
- 权限粒度不同：Jenkins 是项目级别，Jumpserver 是资产/节点级别
