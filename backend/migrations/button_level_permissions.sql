-- ======================================
-- 按钮级权限系统设计
-- 支持最细粒度的权限控制
-- ======================================

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(100) UNIQUE NOT NULL COMMENT '权限编码：module.resource.action',
  name VARCHAR(50) NOT NULL COMMENT '权限名称',
  description VARCHAR(200) COMMENT '权限描述',
  module VARCHAR(50) NOT NULL COMMENT '模块：system, business, auth',
  resource VARCHAR(50) NOT NULL COMMENT '资源：user, role, permission',
  action VARCHAR(50) NOT NULL COMMENT '操作：view, create, update, delete, export, etc',
  level ENUM('module', 'page', 'button', 'api') DEFAULT 'button' COMMENT '权限层级',
  parent_id BIGINT DEFAULT NULL COMMENT '父权限ID',
  sort_order INT DEFAULT 0 COMMENT '排序',
  status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_module (module),
  INDEX idx_resource (resource),
  INDEX idx_level (level),
  INDEX idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  role_id BIGINT NOT NULL COMMENT '角色ID',
  permission_id BIGINT NOT NULL COMMENT '权限ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_role_permission (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- 用户权限关联表（直接分配给用户的权限）
CREATE TABLE IF NOT EXISTS user_permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL COMMENT '用户ID',
  permission_id BIGINT NOT NULL COMMENT '权限ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_permission (user_id, permission_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户权限关联表';

-- API权限映射表（自动将API映射到按钮权限）
CREATE TABLE IF NOT EXISTS api_permission_mappings (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  method VARCHAR(10) NOT NULL COMMENT 'HTTP方法：GET, POST, PUT, DELETE',
  path VARCHAR(200) NOT NULL COMMENT 'API路径',
  permission_code VARCHAR(100) NOT NULL COMMENT '对应的权限编码',
  description VARCHAR(200) COMMENT '描述',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_api (method, path),
  INDEX idx_permission (permission_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API权限映射表';

-- ======================================
-- 用户管理完整按钮级权限示例
-- ======================================

-- 1. 用户管理模块和页面权限
INSERT INTO permissions (code, name, description, module, resource, action, level, parent_id, sort_order) VALUES
-- 模块级权限
('system', '系统管理', '系统管理模块', 'system', '', '', 'module', NULL, 1),

-- 页面级权限
('system.user', '用户管理', '用户管理页面', 'system', 'user', 'view', 'page', LAST_INSERT_ID(), 1),
('system.role', '角色管理', '角色管理页面', 'system', 'role', 'view', 'page', LAST_INSERT_ID(), 2),
('system.permission', '权限管理', '权限管理页面', 'system', 'permission', 'view', 'page', LAST_INSERT_ID(), 3);

-- 2. 用户管理按钮级权限
SET @user_page_id = LAST_INSERT_ID();
INSERT INTO permissions (code, name, description, module, resource, action, level, parent_id, sort_order) VALUES
-- 基础CRUD按钮
('system.user.view', '查看用户', '查看用户列表和详情', 'system', 'user', 'view', 'button', @user_page_id, 1),
('system.user.create', '新增用户', '创建新用户', 'system', 'user', 'create', 'button', @user_page_id, 2),
('system.user.update', '编辑用户', '编辑用户信息', 'system', 'user', 'update', 'button', @user_page_id, 3),
('system.user.delete', '删除用户', '删除用户', 'system', 'user', 'delete', 'button', @user_page_id, 4),

-- 用户管理特殊按钮
('system.user.batch_delete', '批量删除用户', '批量删除多个用户', 'system', 'user', 'batch_delete', 'button', @user_page_id, 5),
('system.user.reset_password', '重置密码', '重置用户密码', 'system', 'user', 'reset_password', 'button', @user_page_id, 6),
('system.user.unlock', '解锁用户', '解锁被锁定的用户', 'system', 'user', 'unlock', 'button', @user_page_id, 7),
('system.user.lock', '锁定用户', '锁定用户账号', 'system', 'user', 'lock', 'button', @user_page_id, 8),

-- 角色分配按钮
('system.user.assign_role', '分配角色', '为用户分配角色', 'system', 'user', 'assign_role', 'button', @user_page_id, 9),
('system.user.remove_role', '移除角色', '移除用户角色', 'system', 'user', 'remove_role', 'button', @user_page_id, 10),
('system.user.view_roles', '查看用户角色', '查看用户的角色列表', 'system', 'user', 'view_roles', 'button', @user_page_id, 11),

-- 数据导入导出按钮
('system.user.export', '导出用户', '导出用户数据', 'system', 'user', 'export', 'button', @user_page_id, 12),
('system.user.import', '导入用户', '导入用户数据', 'system', 'user', 'import', 'button', @user_page_id, 13),
('system.user.download_template', '下载导入模板', '下载用户导入模板', 'system', 'user', 'download_template', 'button', @user_page_id, 14),

-- 用户状态管理
('system.user.enable', '启用用户', '启用用户账号', 'system', 'user', 'enable', 'button', @user_page_id, 15),
('system.user.disable', '禁用用户', '禁用用户账号', 'system', 'user', 'disable', 'button', @user_page_id, 16),

-- 用户信息管理
('system.user.update_profile', '修改个人信息', '用户修改个人资料', 'system', 'user', 'update_profile', 'button', @user_page_id, 17),
('system.user.update_avatar', '修改头像', '修改用户头像', 'system', 'user', 'update_avatar', 'button', @user_page_id, 18),

-- 高级功能
('system.user.copy', '复制用户', '复制用户信息创建新用户', 'system', 'user', 'copy', 'button', @user_page_id, 19),
('system.user.view_history', '查看操作历史', '查看用户操作历史', 'system', 'user', 'view_history', 'button', @user_page_id, 20),
('system.user.approve', '审批用户', '审批用户注册或变更', 'system', 'user', 'approve', 'button', @user_page_id, 21),
('system.user.reject', '拒绝用户', '拒绝用户注册或变更', 'system', 'user', 'reject', 'button', @user_page_id, 22);

-- 3. API权限映射（自动关联按钮权限到API）
INSERT INTO api_permission_mappings (method, path, permission_code, description) VALUES
-- 用户查看相关API
('GET', '/api/system/users', 'system.user.view', '获取用户列表'),
('GET', '/api/system/users/:id', 'system.user.view', '获取用户详情'),
('GET', '/api/system/users/:id/roles', 'system.user.view_roles', '获取用户角色'),
('GET', '/api/system/users/:id/history', 'system.user.view_history', '获取用户操作历史'),

-- 用户创建相关API
('POST', '/api/system/users', 'system.user.create', '创建用户'),
('POST', '/api/system/users/:id/approve', 'system.user.approve', '审批用户'),
('POST', '/api/system/users/:id/reject', 'system.user.reject', '拒绝用户'),

-- 用户更新相关API
('PUT', '/api/system/users/:id', 'system.user.update', '更新用户信息'),
('PUT', '/api/system/users/:id/profile', 'system.user.update_profile', '更新个人资料'),
('PUT', '/api/system/users/:id/avatar', 'system.user.update_avatar', '更新用户头像'),
('PUT', '/api/system/users/:id/password', 'system.user.reset_password', '重置用户密码'),
('PUT', '/api/system/users/:id/enable', 'system.user.enable', '启用用户'),
('PUT', '/api/system/users/:id/disable', 'system.user.disable', '禁用用户'),
('PUT', '/api/system/users/:id/lock', 'system.user.lock', '锁定用户'),
('PUT', '/api/system/users/:id/unlock', 'system.user.unlock', '解锁用户'),

-- 用户删除相关API
('DELETE', '/api/system/users/:id', 'system.user.delete', '删除用户'),
('DELETE', '/api/system/users/batch', 'system.user.batch_delete', '批量删除用户'),
('DELETE', '/api/system/users/:id/roles/:role_id', 'system.user.remove_role', '移除用户角色'),

-- 用户角色管理API
('POST', '/api/system/users/:id/roles', 'system.user.assign_role', '为用户分配角色'),

-- 用户导入导出API
('GET', '/api/system/users/export', 'system.user.export', '导出用户数据'),
('POST', '/api/system/users/import', 'system.user.import', '导入用户数据'),
('GET', '/api/system/users/import/template', 'system.user.download_template', '下载导入模板'),

-- 用户复制API
('POST', '/api/system/users/:id/copy', 'system.user.copy', '复制用户');

-- ======================================
-- 角色管理完整按钮级权限示例
-- ======================================
SET @role_page_id = (SELECT id FROM permissions WHERE code = 'system.role');

INSERT INTO permissions (code, name, description, module, resource, action, level, parent_id, sort_order) VALUES
-- 基础CRUD按钮
('system.role.view', '查看角色', '查看角色列表和详情', 'system', 'role', 'view', 'button', @role_page_id, 1),
('system.role.create', '新增角色', '创建新角色', 'system', 'role', 'create', 'button', @role_page_id, 2),
('system.role.update', '编辑角色', '编辑角色信息', 'system', 'role', 'update', 'button', @role_page_id, 3),
('system.role.delete', '删除角色', '删除角色', 'system', 'role', 'delete', 'button', @role_page_id, 4),

-- 权限分配按钮
('system.role.assign_permission', '分配权限', '为角色分配权限', 'system', 'role', 'assign_permission', 'button', @role_page_id, 5),
('system.role.remove_permission', '移除权限', '移除角色权限', 'system', 'role', 'remove_permission', 'button', @role_page_id, 6),
('system.role.copy_permissions', '复制权限', '复制其他角色的权限', 'system', 'role', 'copy_permissions', 'button', @role_page_id, 7),

-- 角色用户管理
('system.role.view_users', '查看角色用户', '查看角色下的用户列表', 'system', 'role', 'view_users', 'button', @role_page_id, 8),
('system.role.add_users', '添加用户到角色', '批量添加用户到角色', 'system', 'role', 'add_users', 'button', @role_page_id, 9),
('system.role.remove_users', '从角色移除用户', '批量移除角色用户', 'system', 'role', 'remove_users', 'button', @role_page_id, 10),

-- 数据操作
('system.role.export', '导出角色', '导出角色数据', 'system', 'role', 'export', 'button', @role_page_id, 11),
('system.role.import', '导入角色', '导入角色数据', 'system', 'role', 'import', 'button', @role_page_id, 12),

-- 角色状态管理
('system.role.enable', '启用角色', '启用角色', 'system', 'role', 'enable', 'button', @role_page_id, 13),
('system.role.disable', '禁用角色', '禁用角色', 'system', 'role', 'disable', 'button', @role_page_id, 14);

-- ======================================
-- 权限管理完整按钮级权限示例
-- ======================================
SET @permission_page_id = (SELECT id FROM permissions WHERE code = 'system.permission');

INSERT INTO permissions (code, name, description, module, resource, action, level, parent_id, sort_order) VALUES
-- 基础CRUD按钮
('system.permission.view', '查看权限', '查看权限列表和详情', 'system', 'permission', 'view', 'button', @permission_page_id, 1),
('system.permission.create', '新增权限', '创建新权限', 'system', 'permission', 'create', 'button', @permission_page_id, 2),
('system.permission.update', '编辑权限', '编辑权限信息', 'system', 'permission', 'update', 'button', @permission_page_id, 3),
('system.permission.delete', '删除权限', '删除权限', 'system', 'permission', 'delete', 'button', @permission_page_id, 4),

-- 权限组织管理
('system.permission.view_tree', '查看权限树', '查看权限树形结构', 'system', 'permission', 'view_tree', 'button', @permission_page_id, 5),
('system.permission.reorder', '重新排序', '调整权限排序', 'system', 'permission', 'reorder', 'button', @permission_page_id, 6),

-- 权限分配管理
('system.permission.assign_role', '分配角色', '为权限分配角色', 'system', 'permission', 'assign_role', 'button', @permission_page_id, 7),
('system.permission.view_roles', '查看权限角色', '查看哪些角色拥有此权限', 'system', 'permission', 'view_roles', 'button', @permission_page_id, 8);

-- ======================================
-- 示例角色配置
-- ======================================

-- 超级管理员（所有权限）
INSERT INTO roles (name, code, description, status) VALUES
('超级管理员', 'super_admin', '拥有所有权限', 1);

SET @super_admin_role_id = LAST_INSERT_ID();

-- 为超级管理员分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT @super_admin_role_id, id FROM permissions WHERE module = 'system';

-- 用户查看员（只能查看）
INSERT INTO roles (name, code, description, status) VALUES
('用户查看员', 'user_viewer', '只能查看用户，不能操作', 1);

SET @user_viewer_role_id = LAST_INSERT_ID();

-- 为用户查看员分配查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT @user_viewer_role_id, id FROM permissions
WHERE code IN (
  'system', 'system.user', 'system.user.view',
  'system.user.view_roles', 'system.user.view_history'
);

-- 用户管理员（完整用户管理权限）
INSERT INTO roles (name, code, description, status) VALUES
('用户管理员', 'user_manager', '可以管理用户，但不能管理系统其他部分', 1);

SET @user_manager_role_id = LAST_INSERT_ID();

-- 为用户管理员分配所有用户相关权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT @user_manager_role_id, id FROM permissions
WHERE resource = 'user' AND level = 'button';

-- 数据导入导出专员（特定权限）
INSERT INTO roles (name, code, description, status) VALUES
('数据导入导出专员', 'data_import_export', '负责用户数据的导入导出', 1);

SET @data_role_id = LAST_INSERT_ID();

-- 为数据导入导出专员分配特定权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT @data_role_id, id FROM permissions
WHERE code IN (
  'system', 'system.user', 'system.user.view',
  'system.user.export', 'system.user.import',
  'system.user.download_template'
);

-- 用户审批员（审批权限）
INSERT INTO roles (name, code, description, status) VALUES
('用户审批员', 'user_approver', '负责审批用户注册和变更', 1);

SET @approver_role_id = LAST_INSERT_ID();

-- 为审批员分配审批权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT @approver_role_id, id FROM permissions
WHERE code IN (
  'system', 'system.user', 'system.user.view',
  'system.user.approve', 'system.user.reject',
  'system.user.enable', 'system.user.disable'
);

-- ======================================
-- 索引优化
-- ======================================

-- 为权限查询添加复合索引
CREATE INDEX idx_permission_lookup ON permissions(module, resource, action);
CREATE INDEX idx_button_permissions ON permissions(level, parent_id, status);

-- 为角色权限查询优化
CREATE INDEX idx_role_permissions_lookup ON role_permissions(role_id, permission_id);

-- 为用户权限查询优化
CREATE INDEX idx_user_permissions_lookup ON user_permissions(user_id, permission_id);

-- 为API权限映射查询优化
CREATE INDEX idx_api_permission_lookup ON api_permission_mappings(method, path);
