-- ======================================
-- 添加路由字段到 sys_permissions 表
-- 版本: 1.0.0
-- 日期: 2026-08-06
-- 描述: 为权限表添加路由映射字段，实现数据库驱动的权限检查
-- ======================================

-- 添加路由相关字段
ALTER TABLE sys_permissions
ADD COLUMN route_method VARCHAR(10) COMMENT 'HTTP方法: GET/POST/PUT/DELETE' AFTER action,
ADD COLUMN route_path VARCHAR(255) COMMENT '路由路径' AFTER route_method;

-- 添加索引，优化路由查询性能
CREATE INDEX idx_permissions_route ON sys_permissions(route_method, route_path);

-- ======================================
-- 更新现有权限的路由信息
-- ======================================

-- ========== 审计中心模块 ==========
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/login-logs'
WHERE code = 'audit.login_log.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/login-logs/export'
WHERE code = 'audit.login_log.export';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/operation-logs'
WHERE code = 'audit.operation_log.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/operation-logs/export'
WHERE code = 'audit.operation_log.export';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/system-event-logs'
WHERE code = 'audit.system_event.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/stats'
WHERE code = 'audit.stats.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/audit/modules'
WHERE code = 'audit.system_event.list';

-- ========== 系统管理模块 ==========
-- 菜单管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/sys_menus'
WHERE code = 'system.menu.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/sys_menus/tree'
WHERE code = 'system.menu.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/sys_menus'
WHERE code = 'system.menu.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/sys_menus/:id'
WHERE code = 'system.menu.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/sys_menus/:id'
WHERE code = 'system.menu.delete';

-- 角色管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/sys_roles'
WHERE code = 'system.role.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/sys_roles'
WHERE code = 'system.role.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/sys_roles/:id'
WHERE code = 'system.role.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/sys_roles/:id'
WHERE code = 'system.role.delete';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/sys_roles/:roleId/sys_permissions'
WHERE code = 'system.role.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/sys_roles/:roleId/sys_permissions'
WHERE code = 'system.role.assign_permissions';

-- 用户管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/sys_users'
WHERE code = 'system.user.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/sys_users'
WHERE code = 'system.user.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/sys_users/:id'
WHERE code = 'system.user.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/sys_users/:id'
WHERE code = 'system.user.delete';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/sys_users/:id/password'
WHERE code = 'system.user.reset_password';

-- 权限管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/sys_permissions'
WHERE code = 'system.permission.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/sys_permissions/tree'
WHERE code = 'system.permission.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/sys_permissions'
WHERE code = 'system.permission.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/sys_permissions/:id'
WHERE code = 'system.permission.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/sys_permissions/:id'
WHERE code = 'system.permission.delete';

-- 属性管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/attributes'
WHERE code = 'system.attribute.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/attributes/:id'
WHERE code = 'system.attribute.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/attributes'
WHERE code = 'system.attribute.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/attributes/:id'
WHERE code = 'system.attribute.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/attributes/:id'
WHERE code = 'system.attribute.delete';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/server-attributes/:serverId'
WHERE code = 'system.attribute.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/server-attributes/:serverId'
WHERE code = 'system.attribute.update';

-- 路由管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/route/getUserRoutes'
WHERE code = 'system.route.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/route/isRouteExist'
WHERE code = 'system.route.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/route/invalidateCache'
WHERE code = 'system.route.invalidate';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/route/debugCache'
WHERE code = 'system.route.debug';

-- ========== 授权中心模块 ==========
-- 应用管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications'
WHERE code = 'auth.application.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth_applications'
WHERE code = 'auth.application.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/auth_applications/:id'
WHERE code = 'auth.application.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/auth_applications/:id'
WHERE code = 'auth.application.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth_applications/:id/sync-sys_roles'
WHERE code = 'auth.application.sync';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id/sys_roles'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth_applications/:id/sync-sys_users'
WHERE code = 'auth.application.sync';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id/sys_users'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth_applications/:id/sync-groups'
WHERE code = 'auth.application.sync';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id/groups'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth_applications/:id/sync-rules'
WHERE code = 'auth.application.sync';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id/rules'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/:id/operation-logs'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/types'
WHERE code = 'auth.application.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth_applications/types/:type/config'
WHERE code = 'auth.application.view';

-- 授权用户管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth-sys_users'
WHERE code = 'auth.user.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth-sys_users/list'
WHERE code = 'auth.user.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth-sys_users/:id/password'
WHERE code = 'auth.user.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth-sys_users'
WHERE code = 'auth.user.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/auth-sys_users/:id'
WHERE code = 'auth.user.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/auth-sys_users/:id'
WHERE code = 'auth.user.delete';

-- 授权用户组管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth-groups'
WHERE code = 'auth.group.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/auth-groups/list'
WHERE code = 'auth.group.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/auth-groups'
WHERE code = 'auth.group.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/system/auth-groups/:id'
WHERE code = 'auth.group.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/auth-groups/:id'
WHERE code = 'auth.group.delete';

-- 用户组绑定管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/groups/:id/bindings'
WHERE code = 'auth.group_binding.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/groups/:id/bindings'
WHERE code = 'auth.group_binding.create';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/groups/bindings/:id'
WHERE code = 'auth.group_binding.delete';

-- 用户组成员管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/sys_users/:id/groups'
WHERE code = 'auth.user_group.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/system/sys_users/assign-group'
WHERE code = 'auth.user_group.assign';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/sys_users/:id/groups/:groupId'
WHERE code = 'auth.user_group.delete';

-- 用户身份映射
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/user-identity-mappings'
WHERE code = 'auth.identity_mapping.list';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/system/user-identity-mappings/:id'
WHERE code = 'auth.identity_mapping.delete';

-- 用户权限查询
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/user-sys_permissions'
WHERE code = 'auth.permission.query';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/user-sys_permissions/matrix'
WHERE code = 'auth.permission.view';

-- 执行记录
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/system/group-bindings/:bindingId/executions'
WHERE code = 'auth.binding_execution.list';

-- ========== 资产管理模块 (CMDB) ==========
-- 服务器管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers'
WHERE code = 'cmdb.server.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/stats'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/config'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/connect'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/connect'
WHERE code = 'cmdb.server.connect';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/permission'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/sync-metrics'
WHERE code = 'cmdb.server.update';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers'
WHERE code = 'cmdb.server.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/cmdb_servers/:id'
WHERE code = 'cmdb.server.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/cmdb_servers/:id'
WHERE code = 'cmdb.server.delete';

-- Agent管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agents'
WHERE code = 'cmdb.agents.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/agent/status'
WHERE code = 'cmdb.agents.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/agent/deploy'
WHERE code = 'cmdb.agents.deploy';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/agents/batch-deploy'
WHERE code = 'cmdb.agents.deploy';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/agent/restart'
WHERE code = 'cmdb.agents.restart';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/agent/uninstall'
WHERE code = 'cmdb.agents.uninstall';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/agents/batch-uninstall'
WHERE code = 'cmdb.agents.uninstall';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/agents/:id'
WHERE code = 'cmdb.agents.uninstall';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agent-versions'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agent-versions/latest'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agent-versions/:id'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/agent-versions'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/agent-versions/:id'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/agent-versions/:id'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/agent/upgrade'
WHERE code = 'cmdb.agents.upgrade';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agent-upgrade-tasks'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/agent-upgrade-tasks/:id'
WHERE code = 'cmdb.agents.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/cmdb_servers/:id/test-connection'
WHERE code = 'cmdb.server.update';

-- Agent 监控增强
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/extended-metrics'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/metrics/history'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/hardware'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/processes'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/services'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/network'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_servers/:id/security'
WHERE code = 'cmdb.server.view';

-- 主机分组管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/groups'
WHERE code = 'cmdb.group.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/groups/:id'
WHERE code = 'cmdb.group.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/asset-tree'
WHERE code = 'cmdb.group.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/groups'
WHERE code = 'cmdb.group.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/groups/:id'
WHERE code = 'cmdb.group.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/groups/:id'
WHERE code = 'cmdb.group.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/groups/assign'
WHERE code = 'cmdb.group.assign';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/groups/assign-multi'
WHERE code = 'cmdb.group.assign';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/group-cmdb_servers/:groupId'
WHERE code = 'cmdb.group.view';

-- 业务系统管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/business-units'
WHERE code = 'cmdb.business.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/business-units'
WHERE code = 'cmdb.business.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/business-units/:id'
WHERE code = 'cmdb.business.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/business-units/:id'
WHERE code = 'cmdb.business.delete';

-- 机房管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/rooms'
WHERE code = 'cmdb.rooms.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/cmdb_cabinets'
WHERE code = 'cmdb.rooms.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/rooms'
WHERE code = 'cmdb.rooms.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/rooms/:id'
WHERE code = 'cmdb.rooms.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/rooms/:id'
WHERE code = 'cmdb.rooms.delete';

-- 标签管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/tags'
WHERE code = 'cmdb.tags.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/tags'
WHERE code = 'cmdb.tags.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/tags/:id'
WHERE code = 'cmdb.tags.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/tags/:id'
WHERE code = 'cmdb.tags.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/tags/assign'
WHERE code = 'cmdb.tags.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/server-tags/:serverId/:tagId'
WHERE code = 'cmdb.tags.update';

-- SSH凭证管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/ssh-credentials'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/ssh-credentials/:id'
WHERE code = 'cmdb.server.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/ssh-credentials'
WHERE code = 'cmdb.server.update';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/ssh-credentials/:id'
WHERE code = 'cmdb.server.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/ssh-credentials/:id'
WHERE code = 'cmdb.server.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/ssh-credentials/:id/test'
WHERE code = 'cmdb.server.view';

-- 资产变更记录
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/asset-changes'
WHERE code = 'cmdb.server.view';

-- 会话管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions'
WHERE code = 'cmdb.session.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/list'
WHERE code = 'cmdb.session.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/active'
WHERE code = 'cmdb.session.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/active-memory'
WHERE code = 'cmdb.session.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/stats'
WHERE code = 'cmdb.session.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/:id'
WHERE code = 'cmdb.session.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/sessions/:id/terminate'
WHERE code = 'cmdb.session.terminate';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/:id/commands'
WHERE code = 'cmdb.session.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/sessions/:id/file-transfers'
WHERE code = 'cmdb.session.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/sessions/:id/resize'
WHERE code = 'cmdb.session.view';

-- 命令审计
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/commands'
WHERE code = 'cmdb.session.view';

-- 文件传输审计
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/file-transfers'
WHERE code = 'cmdb.session.view';

-- 访问策略管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/access-policies'
WHERE code = 'cmdb.access_policy.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/cmdb/access-policies/:id'
WHERE code = 'cmdb.access_policy.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/cmdb/access-policies'
WHERE code = 'cmdb.access_policy.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/cmdb/access-policies/:id'
WHERE code = 'cmdb.access_policy.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/cmdb/access-policies/:id'
WHERE code = 'cmdb.access_policy.delete';

-- ========== 监控中心模块 ==========
-- 监控数据
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/grafana/url'
WHERE code = 'monitor.data.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/stats'
WHERE code = 'monitor.data.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/refresh'
WHERE code = 'monitor.data.refresh';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/alert/handle'
WHERE code = 'monitor.alert.handle';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/overview'
WHERE code = 'monitor.data.view';

-- 告警管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/alerts'
WHERE code = 'monitor.alert.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/alerts/:id/acknowledge'
WHERE code = 'monitor.alert.ack';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/alerts/stats'
WHERE code = 'monitor.alert.list';

-- 告警规则管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/alerts/rules'
WHERE code = 'monitor.task.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/alerts/rules'
WHERE code = 'monitor.task.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/monitoring/alerts/rules/:id'
WHERE code = 'monitor.task.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/monitoring/alerts/rules/:id'
WHERE code = 'monitor.task.delete';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/monitoring/alerts/rules/:id/status'
WHERE code = 'monitor.task.update';

-- 通知渠道管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/notifications/channels'
WHERE code = 'monitor.task.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/notifications/channels'
WHERE code = 'monitor.task.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/monitoring/notifications/channels/:id'
WHERE code = 'monitor.task.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/monitoring/notifications/channels/:id'
WHERE code = 'monitor.task.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/notifications/channels/:id/test'
WHERE code = 'monitor.task.execute';

-- 巡检报告
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/reports'
WHERE code = 'monitor.report.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/monitoring/reports'
WHERE code = 'monitor.report.create';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/reports/:id'
WHERE code = 'monitor.report.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/monitoring/reports/:id/export'
WHERE code = 'monitor.data.export';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/monitoring/reports/:id'
WHERE code = 'monitor.report.delete';

-- ========== K8s管理模块 ==========
-- 集群管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters'
WHERE code = 'k8s.cluster.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id'
WHERE code = 'k8s.cluster.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters'
WHERE code = 'k8s.cluster.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id'
WHERE code = 'k8s.cluster.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id'
WHERE code = 'k8s.cluster.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/test'
WHERE code = 'k8s.cluster.connect';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/nodes'
WHERE code = 'k8s.cluster.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/namespaces'
WHERE code = 'k8s.cluster.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/sys_users'
WHERE code = 'k8s.permission.list';

-- K8s权限管理
UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/sys_permissions'
WHERE code = 'k8s.permission.assign';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/sys_permissions/:userId'
WHERE code = 'k8s.permission.revoke';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/sys_users/clusters'
WHERE code = 'k8s.permission.list';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/sys_users/:userId/role'
WHERE code = 'k8s.permission.list';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/sys_permissions/batch-assign'
WHERE code = 'k8s.permission.assign';

-- K8s资源管理 - Deployments
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/deployments'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/deployments/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/deployments/:namespace/:name/pods'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/deployments'
WHERE code = 'k8s.resource.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/deployments'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/deployments'
WHERE code = 'k8s.resource.delete';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/deployments/scale'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/deployments/restart'
WHERE code = 'k8s.resource.update';

-- K8s资源管理 - StatefulSets
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/statefulsets'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/statefulsets/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/statefulsets/:namespace/:name/pods'
WHERE code = 'k8s.resource.view';

-- K8s资源管理 - DaemonSets
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/daemonsets'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/daemonsets/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/daemonsets/:namespace/:name/pods'
WHERE code = 'k8s.resource.view';

-- K8s资源管理 - Jobs
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/jobs'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/jobs/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/jobs/:namespace/:name/pods'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/jobs'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - CronJobs
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/cronjobs'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/cronjobs/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/cronjobs/:namespace/:name/pods'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/cronjobs'
WHERE code = 'k8s.resource.delete';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/cronjobs/suspend'
WHERE code = 'k8s.resource.update';

-- K8s资源管理 - Services
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/services'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/services/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/services'
WHERE code = 'k8s.resource.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/services'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/services'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - Ingresses
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/ingresses'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/ingresses/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/ingresses'
WHERE code = 'k8s.resource.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/ingresses'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/ingresses'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - Pods
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/pods'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/pods/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/pods'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/pods/:namespace/:name/logs'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/pods'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - ConfigMaps
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/configmaps'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/configmaps/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/configmaps'
WHERE code = 'k8s.resource.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/configmaps'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/configmaps'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - Secrets
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/secrets'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/secrets/:namespace/:name'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/clusters/:id/secrets'
WHERE code = 'k8s.resource.create';

UPDATE sys_permissions SET route_method = 'PUT', route_path = '/api/k8s/clusters/:id/secrets'
WHERE code = 'k8s.resource.update';

UPDATE sys_permissions SET route_method = 'DELETE', route_path = '/api/k8s/clusters/:id/secrets'
WHERE code = 'k8s.resource.delete';

-- K8s资源管理 - Events
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/clusters/:id/events'
WHERE code = 'k8s.resource.view';

-- K8s终端管理
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/terminal/active'
WHERE code = 'k8s.cluster.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/terminal/sessions/:sessionId/terminate'
WHERE code = 'k8s.cluster.view';

-- K8s诊断功能
UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/diagnostic/commands'
WHERE code = 'k8s.diagnostic.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/diagnostic/pods/:clusterId/:namespace'
WHERE code = 'k8s.resource.view';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/diagnostic/namespaces/:clusterId'
WHERE code = 'k8s.cluster.view';

UPDATE sys_permissions SET route_method = 'POST', route_path = '/api/k8s/diagnostic/execute'
WHERE code = 'k8s.diagnostic.execute';

UPDATE sys_permissions SET route_method = 'GET', route_path = '/api/k8s/diagnostic/history'
WHERE code = 'k8s.diagnostic.view';

-- ======================================
-- 验证统计
-- ======================================
-- 执行完成后运行：
-- SELECT module, resource, action, route_method, route_path FROM sys_permissions WHERE route_method IS NOT NULL ORDER BY module, resource, action;
-- SELECT COUNT(*) as total_with_routes FROM sys_permissions WHERE route_method IS NOT NULL;
