# 权限管理功能测试指南

## 修复概述

本次修复彻底解决了权限管理系统的问题，确保：
- **给角色分配什么权限，用户就有什么权限**
- **前后端逻辑严谨，无兜底方案，无硬编码**
- **权限友好名称从数据库动态获取，不是写死的map**

## 修复内容

### 后端修复（services/auth.go）

1. **PermissionInfo 结构体定义**
   - 从嵌套位置移到包级别（line 151-154）
   - 包含 code 和 name 字段

2. **UserInfo 结构体更新**
   - 添加了 PermissionInfo 字段（line 165）
   - 字段缩进对齐

3. **GetUserInfo 方法**
   - 查询权限详细信息（lines 123-136）
   - 非管理员用户会查询数据库获取权限友好名称
   - 管理员用户（*.*.* 权限）不查询（显示权限码）

4. **ToMap 方法**
   - 返回 permissionInfo 字段（line 189）

### 前端修复（frontend/src/store/modules/auth/index.ts）

1. **删除硬编码的权限名称映射**
   - 删除了 loadPermissionNameMap 方法
   - 删除了 fetchGetPermissionList 导入

2. **优化 getPermissionName 方法**
   - 从登录返回的 permissionInfo 中查找权限名称
   - 如果找不到，回退到显示权限码

3. **userInfo 结构**
   - 包含 permissionInfo 字段（初始化为空数组）

## 测试步骤

### 1. 准备测试环境

```bash
# 启动后端
cd backend
go run main.go run

# 启动前端（另一个终端）
cd frontend
npm run dev
```

### 2. 测试管理员用户（admin）

1. **登录 admin 账号**
   - 用户名: admin
   - 密码: admin123

2. **验证权限数据**
   - 打开浏览器开发者工具（F12）
   - 查看 Application -> Local Storage -> Token
   - 查看 Network 标签的登录响应

3. **预期结果**
   ```json
   {
     "user": {
       "username": "admin",
       ...
     },
     "permissions": ["*.*.*"],
     "permissionInfo": [],
     "menuTree": [...]
   }
   ```
   - admin 用户应该有 *.*.* 权限（通配符权限）
   - permissionInfo 为空数组（管理员不需要查询权限详情）

### 3. 测试普通用户（test）

1. **登录 test 账号**
   - 用户名: test
   - 密码: test123

2. **查看登录响应**
   ```json
   {
     "user": {
       "username": "test",
       ...
     },
     "permissions": [
       "system.user.view",
       "system.role.view",
       ...
     ],
     "permissionInfo": [
       {"code": "system.user.view", "name": "查看用户"},
       {"code": "system.role.view", "name": "查看角色"},
       ...
     ],
     "menuTree": [...]
   }
   ```
   - permissions 包含角色分配的权限码
   - permissionInfo 包含权限友好名称

3. **验证权限友好名称**
   - 在用户管理页面尝试新增用户
   - 如果没有权限，应该看到："您需要【创建用户】权限才能新增用户"
   - 而不是："您需要 system.user.create 权限才能新增用户"

### 4. 测试权限分配流程

1. **使用 admin 账号登录**

2. **进入角色管理**
   - 选择 test 用户的角色
   - 分配新权限，例如："system.user.create"

3. **用 test 账号重新登录**
   - 验证新权限在 permissions 数组中
   - 验证权限友好名称在 permissionInfo 数组中
   - 验证可以执行新增用户操作

### 5. 测试权限控制

1. **测试用户新增用户**
   - 有 system.user.create 权限：可以新增
   - 没有 system.user.create 权限：提示需要【创建用户】权限

2. **测试用户编辑用户**
   - 有 system.user.update 权限：可以编辑
   - 没有 system.user.update 权限：提示需要【编辑用户】权限

3. **测试用户删除用户**
   - 有 system.user.delete 权限：可以删除
   - 没有 system.user.delete 权限：提示需要【删除用户】权限

## 关键验证点

### ✅ 数据驱动验证
- [ ] 权限友好名称来自数据库 permissions 表的 name 字段
- [ ] 前端没有任何硬编码的权限名称映射
- [ ] 权限码格式统一为点号分隔（system.user.view）

### ✅ 权限继承验证
- [ ] 角色分配权限后，用户立即获得相应权限
- [ ] 撤销角色权限后，用户立即失去相应权限
- [ ] admin 用户拥有所有权限（*.*.* 通配符）

### ✅ UI 提示验证
- [ ] 权限提示显示友好名称："【创建用户】权限"
- [ ] 权限提示不再显示权限码："system.user.create"
- [ ] 提示包含申请联系方式

### ✅ 前后端一致性
- [ ] 登录响应包含 permissions 和 permissionInfo
- [ ] 前端 authStore 正确存储这两个字段
- [ ] getPermissionName 方法从 permissionInfo 查找名称

## 故障排查

### 问题1: permissionInfo 为空数组

**原因**: 用户是管理员（有 *.*.* 权限）

**解决**: 这是正常行为，管理员不需要查询权限详情

### 问题2: 权限提示仍然显示权限码

**原因**:
1. 数据库 permissions 表的 name 字段为空
2. permissionInfo 中没有对应的权限码

**解决**:
```sql
-- 检查权限表数据
SELECT code, name FROM permissions WHERE code = 'system.user.view';

-- 如果 name 为空，更新它
UPDATE permissions SET name = '查看用户' WHERE code = 'system.user.view';
```

### 问题3: 后端编译失败

**检查**:
```bash
cd backend
go build -o /dev/null ./
```

**如果失败**: 查看 services/auth.go 的 PermissionInfo 结构体是否在包级别

### 问题4: 前端编译失败

**检查**:
```bash
cd frontend
npm run lint
```

**如果失败**: 查看 auth/index.ts 的 getPermissionName 方法语法

## 数据库权限数据示例

```sql
-- permissions 表应该有类似数据
INSERT INTO permissions (code, name, description) VALUES
('system.user.view', '查看用户', '可以查看用户列表'),
('system.user.create', '创建用户', '可以新增用户'),
('system.user.update', '编辑用户', '可以修改用户信息'),
('system.user.delete', '删除用户', '可以删除用户'),
('system.role.view', '查看角色', '可以查看角色列表'),
('system.role.create', '创建角色', '可以新增角色'),
('system.role.update', '编辑角色', '可以修改角色信息'),
('system.role.delete', '删除角色', '可以删除角色');
```

## 总结

修复后的权限管理系统具有以下特点：

1. **数据驱动**: 所有权限信息来自数据库，无硬编码
2. **严谨逻辑**: 权限分配 → 用户权限 → UI 控制，链路清晰
3. **友好提示**: 显示权限友好名称，提升用户体验
4. **易于维护**: 新增权限只需在数据库添加记录，前端自动支持
