# 系统管理功能对接文档

## 概述

本文档记录了 soybean-admin-element-plus 前端系统管理功能与 OneOps 后端（Go + Gin）的对接情况。

## 对接状态

✅ **用户管理** - 已完成对接
✅ **角色管理** - 已完成对接
✅ **菜单管理** - 已完成对接

## 功能清单

### 1. 用户管理

**功能特性：**
- ✅ 用户列表查询
- ✅ 用户搜索（按用户名、昵称、邮箱）
- ✅ 用户状态显示（启用/禁用）
- ✅ 分页显示

**API 接口：**
- `GET /api/system/users` - 获取用户列表
- `POST /api/system/users` - 创建用户
- `PUT /api/system/users/:id` - 更新用户
- `DELETE /api/system/users/:id` - 删除用户

**数据字段：**
```typescript
{
  id: number;           // 用户ID
  username: string;     // 用户名
  nickname: string;     // 昵称
  avatar: string;       // 头像
  email: string;        // 邮箱
  roleIds: number[];    // 角色ID列表
  status: string;       // 状态：active/inactive
  homePath: string;     // 首页路径
  createdAt: string;    // 创建时间
  updatedAt: string;    // 更新时间
}
```

### 2. 角色管理

**功能特性：**
- ✅ 角色列表查询
- ✅ 角色搜索（按名称、编码）
- ✅ 角色状态显示（启用/禁用）
- ✅ 分页显示

**API 接口：**
- `GET /api/system/roles` - 获取角色列表
- `POST /api/system/roles` - 创建角色
- `PUT /api/system/roles/:id` - 更新角色
- `DELETE /api/system/roles/:id` - 删除角色

**数据字段：**
```typescript
{
  id: number;           // 角色ID
  name: string;         // 角色名称
  code: string;         // 角色编码
  description: string;  // 角色描述
  menuIds: number[];    // 菜单ID列表
  status: number;       // 状态：1-启用，0-禁用
  createdAt: string;    // 创建时间
  updatedAt: string;    // 更新时间
}
```

### 3. 菜单管理

**功能特性：**
- ✅ 菜单树形展示
- ✅ 菜单层级管理
- ✅ 菜单权限配置

**API 接口：**
- `GET /api/system/menus` - 获取菜单列表
- `POST /api/system/menus` - 创建菜单
- `PUT /api/system/menus/:id` - 更新菜单
- `DELETE /api/system/menus/:id` - 删除菜单

**数据字段：**
```typescript
{
  id: number;           // 菜单ID
  name: string;         // 菜单名称
  icon: string;         // 菜单图标
  path: string;         // 路由路径
  permission: string;   // 权限标识
  parentId: number;     // 父菜单ID
  sort: number;         // 排序
  status: number;       // 状态：1-启用，0-禁用
  children?: Menu[];    // 子菜单
  createdAt: string;    // 创建时间
  updatedAt: string;    // 更新时间
}
```

## 字段映射

### 用户字段映射

| 前端显示名称 | 后端字段名 | 类型 | 说明 |
|------------|----------|------|------|
| 用户名 | `username` | string | 用户登录名 |
| 昵称 | `nickname` | string | 用户显示名称 |
| 邮箱 | `email` | string | 用户邮箱 |
| 状态 | `status` | string | active/inactive |
| 角色 | `roleIds` | number[] | 角色ID数组 |

### 角色字段映射

| 前端显示名称 | 后端字段名 | 类型 | 说明 |
|------------|----------|------|------|
| 角色名称 | `name` | string | 角色显示名称 |
| 角色编码 | `code` | string | 角色唯一标识 |
| 角色描述 | `description` | string | 角色说明 |
| 状态 | `status` | number | 1-启用，0-禁用 |

### 菜单字段映射

| 前端显示名称 | 后端字段名 | 类型 | 说明 |
|------------|----------|------|------|
| 菜单名称 | `name` | string | 菜单显示名称 |
| 菜单图标 | `icon` | string | 图标名称 |
| 路由路径 | `path` | string | 前端路由路径 |
| 权限标识 | `permission` | string | 权限控制标识 |
| 父级菜单 | `parentId` | number | 父菜单ID，0表示根菜单 |
| 排序 | `sort` | number | 显示顺序 |

## 权限要求

所有系统管理接口都需要 JWT 认证：

```typescript
headers: {
  'Authorization': `Bearer ${token}`
}
```

默认管理员账号：
- 用户名：`admin`
- 密码：`123456`

## 访问路径

- **用户管理**：http://localhost:9527/manage/user
- **角色管理**：http://localhost:9527/manage/role
- **菜单管理**：http://localhost:9527/manage/menu

## 前端路由配置

系统管理页面已集成到前端路由中，登录后会根据用户权限动态显示。

## 样式说明

系统管理页面使用 soybean-admin-element-plus 的统一样式风格：

- **主题**：支持亮色/暗色主题切换
- **布局**：响应式布局，适配各种屏幕尺寸
- **组件**：使用 Element Plus 组件库
- **表格**：支持分页、排序、筛选
- **表单**：统一的表单样式和验证规则

## 开发说明

### API 调用示例

```typescript
import { fetchGetUserList } from '@/service/api';

// 获取用户列表
const { data, error } = await fetchGetUserList({
  current: 1,
  size: 10,
  username: 'admin',
  status: 'active'
});
```

### 数据处理

所有API响应都经过统一处理，返回格式：

```typescript
{
  code: 200,        // 状态码
  success: true,    // 是否成功
  data: {...},      // 数据
  message: ''       // 消息
}
```

## 测试账号

系统预置的测试数据：

**用户：**
- admin / 123456（超级管理员）
- test / 123456（测试用户）

**角色：**
- 超级管理员（拥有所有权限）
- 普通用户（基础权限）

**菜单：**
- 仪表盘概览
- 资产管理
- 自动化任务
- 监控中心
- 系统管理
- 操作审计

## 注意事项

1. **权限控制**：系统管理功能需要相应的用户权限才能访问
2. **数据验证**：所有输入数据都经过前端和后端双重验证
3. **错误处理**：API错误会统一显示在页面提示中
4. **Token刷新**：Token过期后会自动刷新（如配置了刷新机制）

## 后续优化建议

1. **批量操作**：添加批量删除、批量启用/禁用功能
2. **导出功能**：支持用户、角色数据导出为Excel
3. **操作日志**：记录系统管理操作的历史日志
4. **权限细粒度控制**：实现按钮级别的权限控制
5. **数据缓存**：优化角色和菜单数据的缓存策略

## 故障排查

### API 请求失败

1. 检查后端服务是否正常运行
2. 确认 Token 是否有效
3. 查看浏览器控制台的网络请求详情

### 页面显示异常

1. 清除浏览器缓存
2. 检查前端服务是否正常运行
3. 查看浏览器控制台的错误信息

### 权限错误

1. 确认当前用户是否有相应权限
2. 检查角色配置是否正确
3. 验证菜单权限设置

## 更新日志

**2025-05-09**
- ✅ 完成用户管理功能对接
- ✅ 完成角色管理功能对接
- ✅ 完成菜单管理功能对接
- ✅ 更新API接口路径
- ✅ 修改类型定义匹配后端结构
- ✅ 调整字段名称映射
