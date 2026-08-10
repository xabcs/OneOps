-- ======================================
-- OneOps 权限初始化脚本
-- Level: 1=模块, 2=页面, 3=按钮, 4=API
-- ======================================

-- 清理现有数据
TRUNCATE TABLE sys_permissions;
TRUNCATE TABLE sys_role_permissions;

-- 插入系统管理权限
INSERT INTO sys_permissions (code, name, description, module, resource, action, level, status, sort_order, created_at, updated_at) VALUES
-- 用户管理
('system.user.view', '查看用户', '查看用户列表和详情', 'system', 'user', 'view', 3, 1, 1, NOW(), NOW()),
('system.user.create', '创建用户', '创建新用户', 'system', 'user', 'create', 3, 1, 2, NOW(), NOW()),
('system.user.update', '编辑用户', '编辑用户信息', 'system', 'user', 'update', 3, 1, 3, NOW(), NOW()),
('system.user.delete', '删除用户', '删除用户', 'system', 'user', 'delete', 3, 1, 4, NOW(), NOW()),
('system.user.reset_password', '重置密码', '重置用户密码', 'system', 'user', 'reset_password', 3, 1, 5, NOW(), NOW()),

-- 角色管理
('system.role.view', '查看角色', '查看角色列表和详情', 'system', 'role', 'view', 3, 1, 1, NOW(), NOW()),
('system.role.create', '创建角色', '创建新角色', 'system', 'role', 'create', 3, 1, 2, NOW(), NOW()),
('system.role.update', '编辑角色', '编辑角色信息', 'system', 'role', 'update', 3, 1, 3, NOW(), NOW()),
('system.role.delete', '删除角色', '删除角色', 'system', 'role', 'delete', 3, 1, 4, NOW(), NOW()),
('system.role.assign_permission', '分配权限', '为角色分配权限', 'system', 'role', 'assign_permission', 3, 1, 5, NOW(), NOW()),

-- 权限管理
('system.permission.view', '查看权限', '查看权限列表和详情', 'system', 'permission', 'view', 3, 1, 1, NOW(), NOW()),
('system.permission.create', '创建权限', '创建新权限', 'system', 'permission', 'create', 3, 1, 2, NOW(), NOW()),
('system.permission.update', '编辑权限', '编辑权限信息', 'system', 'permission', 'update', 3, 1, 3, NOW(), NOW()),
('system.permission.delete', '删除权限', '删除权限', 'system', 'permission', 'delete', 3, 1, 4, NOW(), NOW());

-- 为admin角色分配所有权限
INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT 1, id, NOW() FROM sys_permissions;
