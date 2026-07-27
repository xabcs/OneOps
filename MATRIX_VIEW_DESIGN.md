# 矩阵视图数据结构设计

## 概述

不同应用类型使用不同的矩阵视图来展示权限关系，提供更直观的权限可视化。

## 应用类型与矩阵视图映射

| 应用类型 | 矩阵类型 | 横轴 | 纵轴 | 单元格内容 |
|---------|---------|------|------|-----------|
| jenkins | 角色-用户 | 用户 | 角色 | 是否有权限 |
| jumpserver | 用户-授权规则 | 授权规则 | 用户 | 是否有权限 |
| gitlab | 角色-用户 | 用户 | 角色/组 | 是否有权限 |

## 数据结构定义

### 1. Jenkins 矩阵视图（角色-用户矩阵）

```json
{
  "view_type": "role-user-matrix",
  "app_id": 1,
  "app_name": "jenkins",
  "roles": [
    {
      "id": 250,
      "role_code": "admin",
      "role_name": "admin",
      "role_type": "global"
    },
    {
      "id": 252,
      "role_code": "dev",
      "role_name": "dev",
      "role_type": "project"
    }
  ],
  "users": [
    {
      "id": 7,
      "username": "mtest",
      "nickname": "mtest",
      "external_username": "mtest"
    },
    {
      "id": 8,
      "username": "rtest",
      "nickname": "rtest",
      "external_username": "rtest"
    }
  ],
  "matrix": {
    "250": {
      "7": true,
      "8": false
    },
    "252": {
      "7": false,
      "8": true
    }
  },
  "permissions_detail": {
    "250_7": {
      "status": "active",
      "group_name": "工厂实习组",
      "assigned_at": "2026-07-22T17:30:36.511+08:00"
    }
  }
}
```

### 2. Jumpserver 矩阵视图（用户-授权规则矩阵）

```json
{
  "view_type": "user-rule-matrix",
  "app_id": 2,
  "app_name": "jumpserver",
  "users": [
    {
      "id": 7,
      "username": "mtest",
      "nickname": "mtest",
      "external_username": "mtest_jp"
    }
  ],
  "rules": [
    {
      "id": 1,
      "rule_name": "生产服务器访问权限",
      "assets_count": 10,
      "nodes_count": 2
    },
    {
      "id": 2,
      "rule_name": "测试服务器访问权限",
      "assets_count": 5,
      "nodes_count": 1
    }
  ],
  "matrix": {
    "7": {
      "1": true,
      "2": true
    }
  },
  "permissions_detail": {
    "7_1": {
      "status": "active",
      "group_name": "运维组",
      "assigned_at": "2026-07-20T10:00:00+08:00"
    }
  }
}
```

### 3. 通用列表视图

```json
{
  "view_type": "list",
  "app_id": 3,
  "records": [...],
  "total": 100
}
```

## API 设计

### GET /api/v1/system/user-permissions/matrix

**请求参数：**
- `appId`: 应用ID（必填）
- `viewType`: 视图类型（matrix/list，默认matrix）

**响应示例：**

```json
{
  "code": 200,
  "success": true,
  "data": {
    "view_type": "role-user-matrix",
    "app_id": 1,
    "app_name": "jenkins",
    "roles": [...],
    "users": [...],
    "matrix": {...},
    "permissions_detail": {...}
  }
}
```

## 前端组件设计

### 1. 通用矩阵组件 (PermissionMatrix.vue)

**Props:**
- `rows`: 行数据数组
- `columns`: 列数据数组
- `matrix`: 矩阵数据对象
- `rowKey`: 行数据的唯一键字段
- `colKey`: 列数据的唯一键字段
- `rowLabel`: 行显示标签字段
- `colLabel`: 列显示标签字段

**功能:**
- 渲染矩阵表格
- 支持点击单元格查看详情
- 支持批量操作（整行/整列）

### 2. Jenkins 矩阵组件 (JenkinsMatrix.vue)

**特点:**
- 继承通用矩阵组件
- 横轴显示用户
- 纵轴显示角色
- 显示角色类型标签（global/project）
- 支持按角色类型筛选

### 3. Jumpserver 矩阵组件 (JumpserverMatrix.vue)

**特点:**
- 继承通用矩阵组件
- 横轴显示授权规则
- 纵轴显示用户
- 显示资产数量和节点数量
- 支持查看规则详情

## 交互设计

### 1. 视图切换

```
[列表视图] [矩阵视图]
```

- 默认显示矩阵视图（如果应用支持）
- 列表视图用于查看详细信息和导出

### 2. 单元格交互

**有权限（✓）：**
- 显示绿色标签
- 鼠标悬停显示详情（来源用户组、分配时间）
- 点击可撤销权限（需确认）

**无权限（✗）：**
- 显示灰色标签
- 点击可授权（打开授权对话框）

### 3. 批量操作

**整列操作：**
- 点击列头（用户）
- 显示操作菜单：批量授权、批量撤销

**整行操作：**
- 点击行头（角色/规则）
- 显示操作菜单：批量授权、批量撤销

## 实现优先级

1. **Phase 1（MVP）：**
   - 实现 Jenkins 角色-用户矩阵
   - 支持视图切换
   - 支持单元格点击查看详情

2. **Phase 2：**
   - 实现 Jumpserver 用户-授权规则矩阵
   - 支持批量操作

3. **Phase 3：**
   - 支持导出矩阵数据
   - 支持权限变更历史
   - 支持权限对比视图

## 技术考虑

### 性能优化

- **数据分页**：用户/角色数量大时，支持分页或虚拟滚动
- **懒加载**：权限详情按需加载
- **缓存**：矩阵数据缓存，减少重复查询

### 响应式设计

- **小屏幕**：矩阵视图自动转为卡片列表
- **大屏幕**：支持固定列头和行头

### 可访问性

- 支持键盘导航
- 支持屏幕阅读器
- 高对比度模式

## 扩展性

### 支持新应用类型

1. 在数据库配置矩阵类型
2. 实现对应的数据查询逻辑
3. 创建专用矩阵组件（可选，也可使用通用组件）

### 自定义矩阵配置

```json
{
  "matrix_config": {
    "row_entity": "role",
    "col_entity": "user",
    "row_label_field": "role_name",
    "col_label_field": "username",
    "cell_value": "has_permission"
  }
}
```
