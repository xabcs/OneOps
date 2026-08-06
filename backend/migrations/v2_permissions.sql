-- ======================================
-- OneOps 权限数据初始化 SQL
-- 版本: 2.0.0
-- 创建日期: 2026-08-06
-- 描述: 完整的系统权限定义（129个权限）
-- ======================================

-- 使用 UPSERT 语法，可安全重复执行
-- ON DUPLICATE KEY UPDATE 确保数据最新

-- ======================================
-- 系统管理模块（27个权限）
-- ======================================

-- 菜单管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.menu.list', '菜单列表', '查看菜单列表', 'system', 'menu', 'list', 3, 1, 1, NOW(), NOW()),
('system.menu.create', '创建菜单', '创建新菜单', 'system', 'menu', 'create', 3, 2, 1, NOW(), NOW()),
('system.menu.update', '更新菜单', '更新菜单信息', 'system', 'menu', 'update', 3, 3, 1, NOW(), NOW()),
('system.menu.delete', '删除菜单', '删除菜单', 'system', 'menu', 'delete', 3, 4, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 角色管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.role.list', '角色列表', '查看角色列表', 'system', 'role', 'list', 3, 5, 1, NOW(), NOW()),
('system.role.view', '查看角色', '查看角色详情', 'system', 'role', 'view', 3, 6, 1, NOW(), NOW()),
('system.role.create', '创建角色', '创建新角色', 'system', 'role', 'create', 3, 7, 1, NOW(), NOW()),
('system.role.update', '更新角色', '更新角色信息', 'system', 'role', 'update', 3, 8, 1, NOW(), NOW()),
('system.role.delete', '删除角色', '删除角色', 'system', 'role', 'delete', 3, 9, 1, NOW(), NOW()),
('system.role.assign_permissions', '分配权限', '为角色分配权限', 'system', 'role', 'assign_permissions', 3, 10, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 用户管理（5个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.user.list', '用户列表', '查看用户列表', 'system', 'user', 'list', 3, 11, 1, NOW(), NOW()),
('system.user.create', '创建用户', '创建新用户', 'system', 'user', 'create', 3, 12, 1, NOW(), NOW()),
('system.user.update', '更新用户', '更新用户信息', 'system', 'user', 'update', 3, 13, 1, NOW(), NOW()),
('system.user.delete', '删除用户', '删除用户', 'system', 'user', 'delete', 3, 14, 1, NOW(), NOW()),
('system.user.reset_password', '重置密码', '重置用户密码', 'system', 'user', 'reset_password', 3, 15, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 权限管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.permission.list', '权限列表', '查看权限列表', 'system', 'permission', 'list', 3, 16, 1, NOW(), NOW()),
('system.permission.create', '创建权限', '创建新权限', 'system', 'permission', 'create', 3, 17, 1, NOW(), NOW()),
('system.permission.update', '更新权限', '更新权限信息', 'system', 'permission', 'update', 3, 18, 1, NOW(), NOW()),
('system.permission.delete', '删除权限', '删除权限', 'system', 'permission', 'delete', 3, 19, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 属性管理（5个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.attribute.list', '属性列表', '查看属性定义列表', 'system', 'attribute', 'list', 3, 20, 1, NOW(), NOW()),
('system.attribute.view', '查看属性', '查看属性详细信息', 'system', 'attribute', 'view', 3, 21, 1, NOW(), NOW()),
('system.attribute.create', '创建属性', '创建新的属性定义', 'system', 'attribute', 'create', 3, 22, 1, NOW(), NOW()),
('system.attribute.update', '更新属性', '更新属性定义', 'system', 'attribute', 'update', 3, 23, 1, NOW(), NOW()),
('system.attribute.delete', '删除属性', '删除属性定义', 'system', 'attribute', 'delete', 3, 24, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 路由管理（3个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('system.route.list', '查看路由列表', '查看系统路由列表', 'system', 'route', 'list', 3, 25, 1, NOW(), NOW()),
('system.route.invalidate', '刷新路由缓存', '刷新系统路由缓存', 'system', 'route', 'invalidate', 3, 26, 1, NOW(), NOW()),
('system.route.debug', '调试路由', '调试系统路由信息', 'system', 'route', 'debug', 3, 27, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- 审计中心模块（6个权限）
-- ======================================

-- 登录日志（2个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('audit.login_log.list', '登录日志列表', '查看登录日志列表', 'audit', 'login_log', 'list', 3, 28, 1, NOW(), NOW()),
('audit.login_log.export', '导出登录日志', '导出登录日志', 'audit', 'login_log', 'export', 3, 29, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 操作日志（2个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('audit.operation_log.list', '操作日志列表', '查看操作日志列表', 'audit', 'operation_log', 'list', 3, 30, 1, NOW(), NOW()),
('audit.operation_log.export', '导出操作日志', '导出操作日志', 'audit', 'operation_log', 'export', 3, 31, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 系统事件（1个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('audit.system_event.list', '系统事件列表', '查看系统事件列表', 'audit', 'system_event', 'list', 3, 32, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 审计统计（1个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('audit.stats.view', '查看审计统计', '查看审计统计数据', 'audit', 'stats', 'view', 3, 33, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- 授权中心模块（28个权限）
-- ======================================

-- 应用管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.application.list', '查看应用列表', '查看应用权限列表', 'auth', 'application', 'list', 3, 34, 1, NOW(), NOW()),
('auth.application.view', '查看应用详情', '查看应用权限详细信息', 'auth', 'application', 'view', 3, 35, 1, NOW(), NOW()),
('auth.application.create', '创建应用', '创建新的应用权限', 'auth', 'application', 'create', 3, 36, 1, NOW(), NOW()),
('auth.application.update', '更新应用', '更新应用权限信息', 'auth', 'application', 'update', 3, 37, 1, NOW(), NOW()),
('auth.application.delete', '删除应用', '删除应用权限', 'auth', 'application', 'delete', 3, 38, 1, NOW(), NOW()),
('auth.application.sync', '同步应用数据', '同步应用权限数据', 'auth', 'application', 'sync', 3, 39, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 授权用户管理（5个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.user.list', '查看授权用户列表', '查看授权中心用户列表', 'auth', 'user', 'list', 3, 40, 1, NOW(), NOW()),
('auth.user.view', '查看授权用户详情', '查看授权用户详细信息', 'auth', 'user', 'view', 3, 41, 1, NOW(), NOW()),
('auth.user.create', '创建授权用户', '创建新的授权用户', 'auth', 'user', 'create', 3, 42, 1, NOW(), NOW()),
('auth.user.update', '更新授权用户', '更新授权用户信息', 'auth', 'user', 'update', 3, 43, 1, NOW(), NOW()),
('auth.user.delete', '删除授权用户', '删除授权用户', 'auth', 'user', 'delete', 3, 44, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 授权用户组管理（5个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.group.list', '查看用户组列表', '查看授权用户组列表', 'auth', 'group', 'list', 3, 45, 1, NOW(), NOW()),
('auth.group.view', '查看用户组详情', '查看授权用户组详细信息', 'auth', 'group', 'view', 3, 46, 1, NOW(), NOW()),
('auth.group.create', '创建用户组', '创建新的授权用户组', 'auth', 'group', 'create', 3, 47, 1, NOW(), NOW()),
('auth.group.update', '更新用户组', '更新授权用户组信息', 'auth', 'group', 'update', 3, 48, 1, NOW(), NOW()),
('auth.group.delete', '删除用户组', '删除授权用户组', 'auth', 'group', 'delete', 3, 49, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 用户组绑定管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.group_binding.list', '查看用户组绑定', '查看用户组权限绑定列表', 'auth', 'group_binding', 'list', 3, 50, 1, NOW(), NOW()),
('auth.group_binding.view', '查看用户组绑定详情', '查看用户组权限绑定详情', 'auth', 'group_binding', 'view', 3, 51, 1, NOW(), NOW()),
('auth.group_binding.create', '创建用户组绑定', '创建用户组权限绑定', 'auth', 'group_binding', 'create', 3, 52, 1, NOW(), NOW()),
('auth.group_binding.delete', '删除用户组绑定', '删除用户组权限绑定', 'auth', 'group_binding', 'delete', 3, 53, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 用户组成员管理（3个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.user_group.list', '查看用户组成员', '查看用户组成员列表', 'auth', 'user_group', 'list', 3, 54, 1, NOW(), NOW()),
('auth.user_group.assign', '分配用户到组', '将用户分配到用户组', 'auth', 'user_group', 'assign', 3, 55, 1, NOW(), NOW()),
('auth.user_group.delete', '移除用户组成员', '移除用户组成员', 'auth', 'user_group', 'delete', 3, 56, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 用户身份映射（2个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.identity_mapping.list', '查看身份映射', '查看用户身份映射列表', 'auth', 'identity_mapping', 'list', 3, 57, 1, NOW(), NOW()),
('auth.identity_mapping.delete', '删除身份映射', '删除用户身份映射', 'auth', 'identity_mapping', 'delete', 3, 58, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 用户权限查询（2个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.permission.query', '查询用户权限', '查询用户有效权限', 'auth', 'permission', 'query', 3, 59, 1, NOW(), NOW()),
('auth.permission.view', '查看权限矩阵', '查看用户权限矩阵', 'auth', 'permission', 'view', 3, 60, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 执行记录（1个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('auth.binding_execution.list', '查看执行记录', '查看权限绑定执行记录', 'auth', 'binding_execution', 'list', 3, 61, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- 资产管理模块 - CMDB（33个权限）
-- ======================================

-- 服务器管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.server.list', '服务器列表', '查看服务器列表', 'cmdb', 'server', 'list', 3, 62, 1, NOW(), NOW()),
('cmdb.server.view', '查看服务器', '查看服务器详情', 'cmdb', 'server', 'view', 3, 63, 1, NOW(), NOW()),
('cmdb.server.create', '创建服务器', '创建新服务器', 'cmdb', 'server', 'create', 3, 64, 1, NOW(), NOW()),
('cmdb.server.update', '更新服务器', '更新服务器信息', 'cmdb', 'server', 'update', 3, 65, 1, NOW(), NOW()),
('cmdb.server.delete', '删除服务器', '删除服务器', 'cmdb', 'server', 'delete', 3, 66, 1, NOW(), NOW()),
('cmdb.server.connect', '连接服务器', '连接到服务器', 'cmdb', 'server', 'connect', 3, 67, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- Agent管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.agents.list', 'Agent列表', '查看Agent列表', 'cmdb', 'agents', 'list', 3, 68, 1, NOW(), NOW()),
('cmdb.agents.view', '查看Agent', '查看Agent详细信息', 'cmdb', 'agents', 'view', 3, 69, 1, NOW(), NOW()),
('cmdb.agents.deploy', '部署Agent', '部署Agent到服务器', 'cmdb', 'agents', 'deploy', 3, 70, 1, NOW(), NOW()),
('cmdb.agents.restart', '重启Agent', '重启Agent服务', 'cmdb', 'agents', 'restart', 3, 71, 1, NOW(), NOW()),
('cmdb.agents.uninstall', '卸载Agent', '卸载Agent', 'cmdb', 'agents', 'uninstall', 3, 72, 1, NOW(), NOW()),
('cmdb.agents.upgrade', '升级Agent', '升级Agent版本', 'cmdb', 'agents', 'upgrade', 3, 73, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 主机分组管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.group.list', '分组列表', '查看分组列表', 'cmdb', 'group', 'list', 3, 74, 1, NOW(), NOW()),
('cmdb.group.view', '查看分组', '查看分组详情', 'cmdb', 'group', 'view', 3, 75, 1, NOW(), NOW()),
('cmdb.group.create', '创建分组', '创建新分组', 'cmdb', 'group', 'create', 3, 76, 1, NOW(), NOW()),
('cmdb.group.update', '更新分组', '更新分组信息', 'cmdb', 'group', 'update', 3, 77, 1, NOW(), NOW()),
('cmdb.group.delete', '删除分组', '删除分组', 'cmdb', 'group', 'delete', 3, 78, 1, NOW(), NOW()),
('cmdb.group.assign', '分配服务器', '分配服务器到分组', 'cmdb', 'group', 'assign', 3, 79, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 业务系统管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.business.list', '业务列表', '查看业务列表', 'cmdb', 'business', 'list', 3, 80, 1, NOW(), NOW()),
('cmdb.business.create', '创建业务', '创建新业务', 'cmdb', 'business', 'create', 3, 81, 1, NOW(), NOW()),
('cmdb.business.update', '更新业务', '更新业务信息', 'cmdb', 'business', 'update', 3, 82, 1, NOW(), NOW()),
('cmdb.business.delete', '删除业务', '删除业务', 'cmdb', 'business', 'delete', 3, 83, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 机房管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.rooms.list', '机房列表', '查看机房列表', 'cmdb', 'rooms', 'list', 3, 84, 1, NOW(), NOW()),
('cmdb.rooms.create', '创建机房', '创建新机房', 'cmdb', 'rooms', 'create', 3, 85, 1, NOW(), NOW()),
('cmdb.rooms.update', '更新机房', '更新机房信息', 'cmdb', 'rooms', 'update', 3, 86, 1, NOW(), NOW()),
('cmdb.rooms.delete', '删除机房', '删除机房', 'cmdb', 'rooms', 'delete', 3, 87, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 标签管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.tags.list', '标签列表', '查看标签列表', 'cmdb', 'tags', 'list', 3, 88, 1, NOW(), NOW()),
('cmdb.tags.create', '创建标签', '创建新标签', 'cmdb', 'tags', 'create', 3, 89, 1, NOW(), NOW()),
('cmdb.tags.update', '更新标签', '更新标签信息', 'cmdb', 'tags', 'update', 3, 90, 1, NOW(), NOW()),
('cmdb.tags.delete', '删除标签', '删除标签', 'cmdb', 'tags', 'delete', 3, 91, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 会话管理（3个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.session.list', '会话列表', '查看堡垒机会话列表', 'cmdb', 'session', 'list', 3, 92, 1, NOW(), NOW()),
('cmdb.session.view', '查看会话', '查看会话详细信息', 'cmdb', 'session', 'view', 3, 93, 1, NOW(), NOW()),
('cmdb.session.terminate', '终止会话', '终止堡垒机会话', 'cmdb', 'session', 'terminate', 3, 94, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 访问策略管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('cmdb.access_policy.list', '访问策略列表', '查看访问策略列表', 'cmdb', 'access_policy', 'list', 3, 95, 1, NOW(), NOW()),
('cmdb.access_policy.create', '创建访问策略', '创建新的访问策略', 'cmdb', 'access_policy', 'create', 3, 96, 1, NOW(), NOW()),
('cmdb.access_policy.update', '更新访问策略', '更新访问策略信息', 'cmdb', 'access_policy', 'update', 3, 97, 1, NOW(), NOW()),
('cmdb.access_policy.delete', '删除访问策略', '删除访问策略', 'cmdb', 'access_policy', 'delete', 3, 98, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- 监控中心模块（16个权限）
-- ======================================

-- 监控数据（3个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('monitor.data.view', '查看监控数据', '查看监控中心数据', 'monitor', 'data', 'view', 3, 99, 1, NOW(), NOW()),
('monitor.data.refresh', '刷新监控数据', '刷新监控中心数据', 'monitor', 'data', 'refresh', 3, 100, 1, NOW(), NOW()),
('monitor.data.export', '导出监控数据', '导出监控中心数据', 'monitor', 'data', 'export', 3, 101, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 告警管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('monitor.alert.list', '告警列表', '查看告警列表', 'monitor', 'alert', 'list', 3, 102, 1, NOW(), NOW()),
('monitor.alert.view', '查看告警', '查看告警详情', 'monitor', 'alert', 'view', 3, 103, 1, NOW(), NOW()),
('monitor.alert.ack', '确认告警', '确认告警事件', 'monitor', 'alert', 'ack', 3, 104, 1, NOW(), NOW()),
('monitor.alert.handle', '处理告警', '处理告警事件', 'monitor', 'alert', 'handle', 3, 105, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 监控任务（5个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('monitor.task.list', '任务列表', '查看任务列表', 'monitor', 'task', 'list', 3, 106, 1, NOW(), NOW()),
('monitor.task.create', '创建任务', '创建新任务', 'monitor', 'task', 'create', 3, 107, 1, NOW(), NOW()),
('monitor.task.update', '更新任务', '更新任务信息', 'monitor', 'task', 'update', 3, 108, 1, NOW(), NOW()),
('monitor.task.delete', '删除任务', '删除任务', 'monitor', 'task', 'delete', 3, 109, 1, NOW(), NOW()),
('monitor.task.execute', '执行任务', '执行监控任务', 'monitor', 'task', 'execute', 3, 110, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- 巡检报告（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('monitor.report.list', '报告列表', '查看巡检报告列表', 'monitor', 'report', 'list', 3, 111, 1, NOW(), NOW()),
('monitor.report.view', '查看报告', '查看巡检报告详情', 'monitor', 'report', 'view', 3, 112, 1, NOW(), NOW()),
('monitor.report.create', '创建报告', '创建新的巡检报告', 'monitor', 'report', 'create', 3, 113, 1, NOW(), NOW()),
('monitor.report.delete', '删除报告', '删除巡检报告', 'monitor', 'report', 'delete', 3, 114, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- K8s管理模块（15个权限）
-- ======================================

-- 集群管理（6个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('k8s.cluster.list', '集群列表', '查看集群列表', 'k8s', 'cluster', 'list', 3, 115, 1, NOW(), NOW()),
('k8s.cluster.view', '查看集群', '查看集群详情', 'k8s', 'cluster', 'view', 3, 116, 1, NOW(), NOW()),
('k8s.cluster.create', '创建集群', '创建新集群', 'k8s', 'cluster', 'create', 3, 117, 1, NOW(), NOW()),
('k8s.cluster.update', '更新集群', '更新集群信息', 'k8s', 'cluster', 'update', 3, 118, 1, NOW(), NOW()),
('k8s.cluster.delete', '删除集群', '删除集群', 'k8s', 'cluster', 'delete', 3, 119, 1, NOW(), NOW()),
('k8s.cluster.connect', '连接集群', '连接到集群', 'k8s', 'cluster', 'connect', 3, 120, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- K8s权限管理（3个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('k8s.permission.list', '权限列表', '查看K8s权限列表', 'k8s', 'permission', 'list', 3, 121, 1, NOW(), NOW()),
('k8s.permission.assign', '分配权限', '分配K8s权限给用户', 'k8s', 'permission', 'assign', 3, 122, 1, NOW(), NOW()),
('k8s.permission.revoke', '撤销权限', '撤销用户的K8s权限', 'k8s', 'permission', 'revoke', 3, 123, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- K8s资源管理（4个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('k8s.resource.view', '查看资源', '查看K8s资源详情', 'k8s', 'resource', 'view', 3, 124, 1, NOW(), NOW()),
('k8s.resource.create', '创建资源', '创建新的K8s资源', 'k8s', 'resource', 'create', 3, 125, 1, NOW(), NOW()),
('k8s.resource.update', '更新资源', '更新K8s资源信息', 'k8s', 'resource', 'update', 3, 126, 1, NOW(), NOW()),
('k8s.resource.delete', '删除资源', '删除K8s资源', 'k8s', 'resource', 'delete', 3, 127, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- K8s诊断（2个）
INSERT INTO permissions (code, name, description, module, resource, action, level, sort_order, status, created_at, updated_at) VALUES
('k8s.diagnostic.execute', '执行诊断', '执行K8s诊断命令', 'k8s', 'diagnostic', 'execute', 3, 128, 1, NOW(), NOW()),
('k8s.diagnostic.view', '查看诊断结果', '查看K8s诊断结果和历史', 'k8s', 'diagnostic', 'view', 3, 129, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description), sort_order=VALUES(sort_order), updated_at=NOW();

-- ======================================
-- 验证统计
-- ======================================
-- 执行完成后，可运行以下查询验证：
-- SELECT COUNT(*) as total_permissions FROM permissions;
-- SELECT module, COUNT(*) as count FROM permissions GROUP BY module ORDER BY module;

-- 预期结果：
-- system: 27
-- audit: 6
-- auth: 28
-- cmdb: 33
-- monitor: 16
-- k8s: 15
-- 总计: 129
