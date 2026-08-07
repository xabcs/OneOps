-- ======================================
-- OneOps 按钮级权限系统 - 初始权限数据
-- 版本: 1.0.0
-- 创建日期: 2026-08-03
-- 描述: 初始化基础系统权限数据
-- ======================================

-- 注意：执行此脚本前，请确保已执行 permission_tables.sql 创建表结构

-- ======================================
-- 1. 清理现有权限数据（可选，用于重新初始化）
-- ======================================
-- TRUNCATE TABLE sys_permissions;
-- TRUNCATE TABLE sys_role_permissions;
-- TRUNCATE TABLE sys_user_permissions;

-- ======================================
-- 2. 插入系统管理模块权限
-- ======================================

-- 2.1 用户管理权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 用户管理菜单权限
('system.user.menu', '用户管理', '用户管理菜单权限', 'system', 'user', 'view', 'menu', NULL, 1),

-- 用户管理按钮权限
('system.user.view', '查看用户', '查看用户列表和详情权限', 'system', 'user', 'view', 'button', LAST_INSERT_ID(), 1),
('system.user.create', '新增用户', '新增用户权限', 'system', 'user', 'create', 'button', LAST_INSERT_ID(), 2),
('system.user.update', '编辑用户', '编辑用户信息权限', 'system', 'user', 'update', 'button', LAST_INSERT_ID(), 3),
('system.user.delete', '删除用户', '删除用户权限', 'system', 'user', 'delete', 'button', LAST_INSERT_ID(), 4),
('system.user.reset_password', '重置密码', '重置用户密码权限', 'system', 'user', 'reset_password', 'button', LAST_INSERT_ID(), 5),
('system.user.unlock', '解锁用户', '解锁用户权限', 'system', 'user', 'unlock', 'button', LAST_INSERT_ID(), 6),
('system.user.assign_role', '分配角色', '为用户分配角色权限', 'system', 'user', 'assign_role', 'button', LAST_INSERT_ID(), 7),
('system.user.export', '导出用户', '导出用户数据权限', 'system', 'user', 'export', 'button', LAST_INSERT_ID(), 8),
('system.user.import', '导入用户', '导入用户数据权限', 'system', 'user', 'import', 'button', LAST_INSERT_ID(), 9);

-- 2.2 角色管理权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 角色管理菜单权限
('system.role.menu', '角色管理', '角色管理菜单权限', 'system', 'role', 'view', 'menu', NULL, 2),

-- 角色管理按钮权限
('system.role.view', '查看角色', '查看角色列表和详情权限', 'system', 'role', 'view', 'button', LAST_INSERT_ID(), 1),
('system.role.create', '新增角色', '新增角色权限', 'system', 'role', 'create', 'button', LAST_INSERT_ID(), 2),
('system.role.update', '编辑角色', '编辑角色信息权限', 'system', 'role', 'update', 'button', LAST_INSERT_ID(), 3),
('system.role.delete', '删除角色', '删除角色权限', 'system', 'role', 'delete', 'button', LAST_INSERT_ID(), 4),
('system.role.assign_permission', '分配权限', '为角色分配权限', 'system', 'role', 'assign_permission', 'button', LAST_INSERT_ID(), 5),
('system.role.view_users', '查看角色用户', '查看角色下的用户列表', 'system', 'role', 'view_users', 'button', LAST_INSERT_ID(), 6);

-- 2.3 权限管理权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 权限管理菜单权限
('system.permission.menu', '权限管理', '权限管理菜单权限', 'system', 'permission', 'view', 'menu', NULL, 3),

-- 权限管理按钮权限
('system.permission.view', '查看权限', '查看权限列表和详情权限', 'system', 'permission', 'view', 'button', LAST_INSERT_ID(), 1),
('system.permission.create', '新增权限', '新增权限权限', 'system', 'permission', 'create', 'button', LAST_INSERT_ID(), 2),
('system.permission.update', '编辑权限', '编辑权限信息权限', 'system', 'permission', 'update', 'button', LAST_INSERT_ID(), 3),
('system.permission.delete', '删除权限', '删除权限权限', 'system', 'permission', 'delete', 'button', LAST_INSERT_ID(), 4),
('system.permission.export', '导出权限', '导出权限配置权限', 'system', 'permission', 'export', 'button', LAST_INSERT_ID(), 5),
('system.permission.import', '导入权限', '导入权限配置权限', 'system', 'permission', 'import', 'button', LAST_INSERT_ID(), 6);

-- 2.4 菜单管理权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 菜单管理菜单权限
('system.menu.menu', '菜单管理', '菜单管理菜单权限', 'system', 'menu', 'view', 'menu', NULL, 4),

-- 菜单管理按钮权限
('system.menu.view', '查看菜单', '查看菜单列表和详情权限', 'system', 'menu', 'view', 'button', LAST_INSERT_ID(), 1),
('system.menu.create', '新增菜单', '新增菜单权限', 'system', 'menu', 'create', 'button', LAST_INSERT_ID(), 2),
('system.menu.update', '编辑菜单', '编辑菜单信息权限', 'system', 'menu', 'update', 'button', LAST_INSERT_ID(), 3),
('system.menu.delete', '删除菜单', '删除菜单权限', 'system', 'menu', 'delete', 'button', LAST_INSERT_ID(), 4),
('system.menu.sort', '菜单排序', '调整菜单顺序权限', 'system', 'menu', 'sort', 'button', LAST_INSERT_ID(), 5);

-- ======================================
-- 3. 插入审计模块权限
-- ======================================

-- 3.1 登录日志权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 登录日志菜单权限
('audit.login_log.menu', '登录日志', '登录日志查看权限', 'audit', 'login_log', 'view', 'menu', NULL, 10),

-- 登录日志按钮权限
('audit.login_log.view', '查看登录日志', '查看登录日志列表权限', 'audit', 'login_log', 'view', 'button', LAST_INSERT_ID(), 1),
('audit.login_log.export', '导出登录日志', '导出登录日志权限', 'audit', 'login_log', 'export', 'button', LAST_INSERT_ID(), 2);

-- 3.2 操作日志权限
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 操作日志菜单权限
('audit.operation_log.menu', '操作日志', '操作日志查看权限', 'audit', 'operation_log', 'view', 'menu', NULL, 11),

-- 操作日志按钮权限
('audit.operation_log.view', '查看操作日志', '查看操作日志列表权限', 'audit', 'operation_log', 'view', 'button', LAST_INSERT_ID(), 1),
('audit.operation_log.export', '导出操作日志', '导出操作日志权限', 'audit', 'operation_log', 'export', 'button', LAST_INSERT_ID(), 2);

-- ======================================
-- 4. 插入系统配置权限
-- ======================================

INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
-- 系统设置菜单权限
('system.settings.menu', '系统设置', '系统设置菜单权限', 'system', 'settings', 'view', 'menu', NULL, 20),

-- 系统设置按钮权限
('system.settings.view', '查看设置', '查看系统设置权限', 'system', 'settings', 'view', 'button', LAST_INSERT_ID(), 1),
('system.settings.update', '修改设置', '修改系统设置权限', 'system', 'settings', 'update', 'button', LAST_INSERT_ID(), 2),
('system.settings.restart', '重启服务', '重启系统服务权限', 'system', 'settings', 'restart', 'button', LAST_INSERT_ID(), 3);

-- ======================================
-- 5. 插入超级管理员权限（通配符权限）
-- ======================================
INSERT INTO `sys_permissions` (`code`, `name`, `description`, `module`, `resource`, `action`, `level`, `parent_id`, `sort_order`) VALUES
('*.*.*', '超级管理员', '拥有所有权限的超级管理员', '*', '*', '*', 'menu', NULL, 0),
('system.*.*', '系统管理所有权限', '系统模块的所有权限', 'system', '*', '*', 'menu', NULL, 1);

-- ======================================
-- 6. 为默认管理员角色分配权限
-- ======================================
-- 假设默认管理员角色ID为1，如果没有请先创建角色
-- 为超级管理员角色分配所有权限
INSERT INTO `sys_role_permissions` (`role_id`, `permission_id`)
SELECT 1, `id` FROM `sys_permissions` WHERE `code` IN ('*.*.*', 'system.*.*');

-- ======================================
-- 7. 创建权限模板（常用权限组合）
-- ======================================
-- 系统管理员权限模板（包含用户、角色、权限管理）
-- 这个可以在代码中实现，或者创建单独的权限模板表

-- ======================================
-- 8. 验证数据插入
-- ======================================
-- 验证权限数量
SELECT
    module,
    level,
    COUNT(*) as count
FROM sys_permissions
GROUP BY module, level
ORDER BY module, level;

-- 验证权限树结构
SELECT
    p1.code,
    p1.name,
    p1.level,
    p2.code as parent_code,
    p2.name as parent_name
FROM sys_permissions p1
LEFT JOIN sys_permissions p2 ON p1.parent_id = p2.id
ORDER BY p1.sort_order;

-- ======================================
-- 9. 更新权限树的父级关系
-- ======================================
-- 由于自动插入的权限需要正确的parent_id，这里需要手动更新
-- 注意：这个部分需要根据实际插入的ID进行调整

-- 更新用户管理按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'system.user.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'system.user.%' AND p.code != 'system.user.menu';

-- 更新角色管理按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'system.role.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'system.role.%' AND p.code != 'system.role.menu';

-- 更新权限管理按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'system.permission.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'system.permission.%' AND p.code != 'system.permission.menu';

-- 更新菜单管理按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'system.menu.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'system.menu.%' AND p.code != 'system.menu.menu';

-- 更新登录日志按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'audit.login_log.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'audit.login_log.%' AND p.code != 'audit.login_log.menu';

-- 更新操作日志按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'audit.operation_log.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'audit.operation_log.%' AND p.code != 'audit.operation_log.menu';

-- 更新系统设置按钮权限的parent_id
UPDATE sys_permissions p
JOIN sys_permissions parent ON parent.code = 'system.settings.menu'
SET p.parent_id = parent.id
WHERE p.code LIKE 'system.settings.%' AND p.code != 'system.settings.menu';

-- ======================================
-- 10. 完成提示
-- ======================================
-- 权限数据初始化完成！
-- 总权限数量：SELECT COUNT(*) FROM sys_permissions;
-- 菜单权限数量：SELECT COUNT(*) FROM sys_permissions WHERE level = 'menu';
-- 按钮权限数量：SELECT COUNT(*) FROM sys_permissions WHERE level = 'button';

-- 下一步：
-- 1. 执行数据库迁移脚本
-- 2. 验证权限数据是否正确插入
-- 3. 为现有角色分配相应权限
-- 4. 测试权限功能