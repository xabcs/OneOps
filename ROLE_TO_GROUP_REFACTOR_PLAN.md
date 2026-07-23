# 授权中心"角色"改名为"用户组"重构计划

## 重构目标

将授权中心的"角色(Role)"概念重命名为"用户组(Group)"，使术语更加准确和易于理解。

## 重构范围

### 1. 数据库表重命名

| 原表名 | 新表名 | 说明 |
|--------|--------|------|
| `auth_roles` | `auth_groups` | 用户组表 |
| `auth_user_roles` | `auth_user_groups` | 用户组成员关系表 |
| `role_bindings` | `group_bindings` | 用户组权限绑定表 |

### 2. 字段重命名

**auth_groups 表：**
- `role_code` → `group_code`
- `role_name` → `group_name`

**auth_user_groups 表：**
- `role_id` → `group_id`

**group_bindings 表：**
- `role_id` → `group_id`

### 3. 后端代码重命名

**模型：**
```go
AuthRole      → AuthGroup
AuthUserRole  → AuthUserGroup
RoleBinding   → GroupBinding
```

**服务方法：**
```go
CreateAuthRole      → CreateAuthGroup
GetAuthRoles        → GetAuthGroups
UpdateAuthRole      → UpdateAuthGroup
DeleteAuthRole      → DeleteAuthGroup
GetAllAuthRoles     → GetAllAuthGroups

AssignRoleToUser    → AssignUserToGroup
DeleteUserRole      → DeleteUserGroup

CreateRoleBinding   → CreateGroupBinding
GetRoleBindings     → GetGroupBindings
DeleteRoleBinding   → DeleteGroupBinding
```

**控制器：**
```go
GetAuthRoles        → GetAuthGroups
CreateAuthRole      → CreateAuthGroup
UpdateAuthRole      → UpdateAuthGroup
DeleteAuthRole      → DeleteAuthGroup
GetAllAuthRoles     → GetAllAuthGroups

AssignRoleToUser    → AssignUserToGroup
DeleteUserRole      → DeleteUserGroup

GetRoleBindings     → GetGroupBindings
CreateRoleBinding   → CreateGroupBinding
DeleteRoleBinding   → DeleteGroupBinding
```

### 4. 前端UI重命名

**菜单名称：**
- 角色管理 → 用户组管理
- 角色绑定 → 用户组权限配置
- 用户授权 → 用户组分配

**页面标题：**
- 角色管理 → 用户组管理
- 授权中心角色列表 → 用户组列表
- 添加角色 → 添加用户组
- 编辑角色 → 编辑用户组
- 分配角色 → 分配用户组
- 角色绑定列表 → 用户组权限配置列表
- 用户角色列表 → 用户所属用户组列表

**按钮文字：**
- 添加角色 → 添加用户组
- 分配角色 → 分配用户组

**表格列标题：**
- 角色名称 → 用户组名称
- 角色代码 → 用户组代码
- 外部角色名称 → 外部权限名称

### 5. 国际化文件

**中文（zh-cn.ts）：**
```typescript
route: {
  auth_roles: '用户组管理',
  auth_rolebindings: '用户组权限配置',
  auth_userauthorization: '用户组分配'
}

page: {
  auth: {
    roles: {
      title: '用户组管理',
      addGroup: '添加用户组',
      editGroup: '编辑用户组',
      groupName: '用户组名称',
      groupCode: '用户组代码'
    }
  }
}
```

**英文（en-us.ts）：**
```typescript
route: {
  auth_roles: 'User Groups',
  auth_rolebindings: 'Group Permission Config',
  auth_userauthorization: 'User Group Assignment'
}
```

## 重构步骤

### 阶段一：数据库迁移（优先）

1. 创建数据库迁移脚本
2. 备份现有数据
3. 执行表重命名
4. 更新外键约束
5. 验证数据完整性

### 阶段二：后端重构

1. 重命名模型文件和结构体
2. 重命名服务方法
3. 重命名控制器方法
4. 更新路由配置
5. 测试API接口

### 阶段三：前端重构

1. 更新国际化文件
2. 修改组件文件名
3. 更新UI文字
4. 测试前端功能

### 阶段四：文档更新

1. 更新API文档
2. 更新用户手册
3. 更新开发者文档

## 重构影响

### 受影响的功能

- ✅ 用户组管理（原角色管理）
- ✅ 用户组权限配置（原角色绑定）
- ✅ 用户组分配（原用户授权）
- ❌ 不影响：用户管理、应用管理、操作日志

### 数据兼容性

- 所有现有数据自动迁移到新表
- 数据关系保持不变
- 用户无需重新配置

## 风险评估

### 低风险
- 表重命名（数据库支持）
- UI文字修改（纯显示层）

### 中等风险
- 后端代码重构（需要全面测试）
- API路径变更（前端需要同步修改）

### 风险缓解措施
- 完整的数据库备份
- 分阶段实施
- 完整的测试用例
- 回滚方案

## 验收标准

### 功能验收
- ✅ 所有API接口正常工作
- ✅ 前端页面显示正确
- ✅ 数据完整性验证通过
- ✅ 用户组管理功能正常
- ✅ 用户组权限配置功能正常
- ✅ 用户组分配功能正常

### 文档验收
- ✅ API文档更新完成
- ✅ 用户手册更新完成
- ✅ 数据库文档更新完成

## 时间估算

- 数据库迁移：1小时
- 后端重构：2-3小时
- 前端重构：2-3小时
- 测试验证：1-2小时
- 文档更新：1小时

**总计：7-10小时**

## 回滚方案

如果重构出现问题，可以执行回滚：

1. 恢复数据库表到原名
2. 回滚代码到重构前版本
3. 恢复前端代码

## 下一步行动

确认开始重构后，将按以下顺序执行：

1. 创建数据库迁移脚本
2. 修改后端模型和代码
3. 修改前端UI和国际化
4. 执行全面测试
5. 更新文档

---

**准备开始重构？**
