package system

import (
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// syncPermissions 同步系统权限数据（129个权限，6个模块）
func (i *Initializer) syncPermissions() error {
	logger.Info("开始同步权限数据...")

	db := database.GetDB()

	permissions := []modelsystem.Permission{
		// ========== 系统管理模块（27个） ==========
		// 菜单管理
		{Code: "system.menu.list", Name: "菜单列表", Description: "查看菜单列表", Module: "system", Resource: "menu", Action: "list", Level: 3, SortOrder: 1, Status: 1},
		{Code: "system.menu.create", Name: "创建菜单", Description: "创建新菜单", Module: "system", Resource: "menu", Action: "create", Level: 3, SortOrder: 2, Status: 1},
		{Code: "system.menu.update", Name: "更新菜单", Description: "更新菜单信息", Module: "system", Resource: "menu", Action: "update", Level: 3, SortOrder: 3, Status: 1},
		{Code: "system.menu.delete", Name: "删除菜单", Description: "删除菜单", Module: "system", Resource: "menu", Action: "delete", Level: 3, SortOrder: 4, Status: 1},
		// 角色管理
		{Code: "system.role.list", Name: "角色列表", Description: "查看角色列表", Module: "system", Resource: "role", Action: "list", Level: 3, SortOrder: 5, Status: 1},
		{Code: "system.role.view", Name: "查看角色", Description: "查看角色详情", Module: "system", Resource: "role", Action: "view", Level: 3, SortOrder: 6, Status: 1},
		{Code: "system.role.create", Name: "创建角色", Description: "创建新角色", Module: "system", Resource: "role", Action: "create", Level: 3, SortOrder: 7, Status: 1},
		{Code: "system.role.update", Name: "更新角色", Description: "更新角色信息", Module: "system", Resource: "role", Action: "update", Level: 3, SortOrder: 8, Status: 1},
		{Code: "system.role.delete", Name: "删除角色", Description: "删除角色", Module: "system", Resource: "role", Action: "delete", Level: 3, SortOrder: 9, Status: 1},
		{Code: "system.role.assign_permissions", Name: "分配权限", Description: "为角色分配权限", Module: "system", Resource: "role", Action: "assign_permissions", Level: 3, SortOrder: 10, Status: 1},
		// 用户管理
		{Code: "system.user.list", Name: "用户列表", Description: "查看用户列表", Module: "system", Resource: "user", Action: "list", Level: 3, SortOrder: 11, Status: 1},
		{Code: "system.user.create", Name: "创建用户", Description: "创建新用户", Module: "system", Resource: "user", Action: "create", Level: 3, SortOrder: 12, Status: 1},
		{Code: "system.user.update", Name: "更新用户", Description: "更新用户信息", Module: "system", Resource: "user", Action: "update", Level: 3, SortOrder: 13, Status: 1},
		{Code: "system.user.delete", Name: "删除用户", Description: "删除用户", Module: "system", Resource: "user", Action: "delete", Level: 3, SortOrder: 14, Status: 1},
		{Code: "system.user.reset_password", Name: "重置密码", Description: "重置用户密码", Module: "system", Resource: "user", Action: "reset_password", Level: 3, SortOrder: 15, Status: 1},
		// 权限管理
		{Code: "system.permission.list", Name: "权限列表", Description: "查看权限列表", Module: "system", Resource: "permission", Action: "list", Level: 3, SortOrder: 16, Status: 1},
		{Code: "system.permission.create", Name: "创建权限", Description: "创建新权限", Module: "system", Resource: "permission", Action: "create", Level: 3, SortOrder: 17, Status: 1},
		{Code: "system.permission.update", Name: "更新权限", Description: "更新权限信息", Module: "system", Resource: "permission", Action: "update", Level: 3, SortOrder: 18, Status: 1},
		{Code: "system.permission.delete", Name: "删除权限", Description: "删除权限", Module: "system", Resource: "permission", Action: "delete", Level: 3, SortOrder: 19, Status: 1},
		// 属性管理
		{Code: "cmdb.attribute.list", Name: "属性列表", Description: "查看属性定义列表", Module: "cmdb", Resource: "attribute", Action: "list", Level: 3, SortOrder: 20, Status: 1},
		{Code: "cmdb.attribute.view", Name: "查看属性", Description: "查看属性详细信息", Module: "cmdb", Resource: "attribute", Action: "view", Level: 3, SortOrder: 21, Status: 1},
		{Code: "cmdb.attribute.create", Name: "创建属性", Description: "创建新的属性定义", Module: "cmdb", Resource: "attribute", Action: "create", Level: 3, SortOrder: 22, Status: 1},
		{Code: "cmdb.attribute.update", Name: "更新属性", Description: "更新属性定义", Module: "cmdb", Resource: "attribute", Action: "update", Level: 3, SortOrder: 23, Status: 1},
		{Code: "cmdb.attribute.delete", Name: "删除属性", Description: "删除属性定义", Module: "cmdb", Resource: "attribute", Action: "delete", Level: 3, SortOrder: 24, Status: 1},
		// 路由管理
		{Code: "system.route.list", Name: "查看路由列表", Description: "查看系统路由列表", Module: "system", Resource: "route", Action: "list", Level: 3, SortOrder: 25, Status: 1},
		{Code: "system.route.invalidate", Name: "刷新路由缓存", Description: "刷新系统路由缓存", Module: "system", Resource: "route", Action: "invalidate", Level: 3, SortOrder: 26, Status: 1},
		{Code: "system.route.debug", Name: "调试路由", Description: "调试系统路由信息", Module: "system", Resource: "route", Action: "debug", Level: 3, SortOrder: 27, Status: 1},

		// ========== 审计中心模块（6个） ==========
		{Code: "audit.login_log.list", Name: "登录日志列表", Description: "查看登录日志列表", Module: "audit", Resource: "login_log", Action: "list", Level: 3, SortOrder: 28, Status: 1},
		{Code: "audit.login_log.export", Name: "导出登录日志", Description: "导出登录日志", Module: "audit", Resource: "login_log", Action: "export", Level: 3, SortOrder: 29, Status: 1},
		{Code: "audit.operation_log.list", Name: "操作日志列表", Description: "查看操作日志列表", Module: "audit", Resource: "operation_log", Action: "list", Level: 3, SortOrder: 30, Status: 1},
		{Code: "audit.operation_log.export", Name: "导出操作日志", Description: "导出操作日志", Module: "audit", Resource: "operation_log", Action: "export", Level: 3, SortOrder: 31, Status: 1},
		{Code: "audit.system_event.list", Name: "系统事件列表", Description: "查看系统事件列表", Module: "audit", Resource: "system_event", Action: "list", Level: 3, SortOrder: 32, Status: 1},
		{Code: "audit.stats.view", Name: "查看审计统计", Description: "查看审计统计数据", Module: "audit", Resource: "stats", Action: "view", Level: 3, SortOrder: 33, Status: 1},

		// ========== 授权中心模块（28个） ==========
		// 应用管理
		{Code: "auth.application.list", Name: "查看应用列表", Description: "查看应用权限列表", Module: "auth", Resource: "application", Action: "list", Level: 3, SortOrder: 34, Status: 1},
		{Code: "auth.application.view", Name: "查看应用详情", Description: "查看应用权限详细信息", Module: "auth", Resource: "application", Action: "view", Level: 3, SortOrder: 35, Status: 1},
		{Code: "auth.application.create", Name: "创建应用", Description: "创建新的应用权限", Module: "auth", Resource: "application", Action: "create", Level: 3, SortOrder: 36, Status: 1},
		{Code: "auth.application.update", Name: "更新应用", Description: "更新应用权限信息", Module: "auth", Resource: "application", Action: "update", Level: 3, SortOrder: 37, Status: 1},
		{Code: "auth.application.delete", Name: "删除应用", Description: "删除应用权限", Module: "auth", Resource: "application", Action: "delete", Level: 3, SortOrder: 38, Status: 1},
		{Code: "auth.application.sync", Name: "同步应用数据", Description: "同步应用权限数据", Module: "auth", Resource: "application", Action: "sync", Level: 3, SortOrder: 39, Status: 1},
		// 授权用户管理
		{Code: "auth.user.list", Name: "查看授权用户列表", Description: "查看授权中心用户列表", Module: "auth", Resource: "user", Action: "list", Level: 3, SortOrder: 40, Status: 1},
		{Code: "auth.user.view", Name: "查看授权用户详情", Description: "查看授权用户详细信息", Module: "auth", Resource: "user", Action: "view", Level: 3, SortOrder: 41, Status: 1},
		{Code: "auth.user.create", Name: "创建授权用户", Description: "创建新的授权用户", Module: "auth", Resource: "user", Action: "create", Level: 3, SortOrder: 42, Status: 1},
		{Code: "auth.user.update", Name: "更新授权用户", Description: "更新授权用户信息", Module: "auth", Resource: "user", Action: "update", Level: 3, SortOrder: 43, Status: 1},
		{Code: "auth.user.delete", Name: "删除授权用户", Description: "删除授权用户", Module: "auth", Resource: "user", Action: "delete", Level: 3, SortOrder: 44, Status: 1},
		// 授权用户组管理
		{Code: "auth.group.list", Name: "查看用户组列表", Description: "查看授权用户组列表", Module: "auth", Resource: "group", Action: "list", Level: 3, SortOrder: 45, Status: 1},
		{Code: "auth.group.view", Name: "查看用户组详情", Description: "查看授权用户组详细信息", Module: "auth", Resource: "group", Action: "view", Level: 3, SortOrder: 46, Status: 1},
		{Code: "auth.group.create", Name: "创建用户组", Description: "创建新的授权用户组", Module: "auth", Resource: "group", Action: "create", Level: 3, SortOrder: 47, Status: 1},
		{Code: "auth.group.update", Name: "更新用户组", Description: "更新授权用户组信息", Module: "auth", Resource: "group", Action: "update", Level: 3, SortOrder: 48, Status: 1},
		{Code: "auth.group.delete", Name: "删除用户组", Description: "删除授权用户组", Module: "auth", Resource: "group", Action: "delete", Level: 3, SortOrder: 49, Status: 1},
		// 用户组绑定管理
		{Code: "auth.group_binding.list", Name: "查看用户组绑定", Description: "查看用户组权限绑定列表", Module: "auth", Resource: "group_binding", Action: "list", Level: 3, SortOrder: 50, Status: 1},
		{Code: "auth.group_binding.view", Name: "查看用户组绑定详情", Description: "查看用户组权限绑定详情", Module: "auth", Resource: "group_binding", Action: "view", Level: 3, SortOrder: 51, Status: 1},
		{Code: "auth.group_binding.create", Name: "创建用户组绑定", Description: "创建用户组权限绑定", Module: "auth", Resource: "group_binding", Action: "create", Level: 3, SortOrder: 52, Status: 1},
		{Code: "auth.group_binding.delete", Name: "删除用户组绑定", Description: "删除用户组权限绑定", Module: "auth", Resource: "group_binding", Action: "delete", Level: 3, SortOrder: 53, Status: 1},
		// 用户组成员管理
		{Code: "auth.user_group.list", Name: "查看用户组成员", Description: "查看用户组成员列表", Module: "auth", Resource: "user_group", Action: "list", Level: 3, SortOrder: 54, Status: 1},
		{Code: "auth.user_group.assign", Name: "分配用户到组", Description: "将用户分配到用户组", Module: "auth", Resource: "user_group", Action: "assign", Level: 3, SortOrder: 55, Status: 1},
		{Code: "auth.user_group.delete", Name: "移除用户组成员", Description: "移除用户组成员", Module: "auth", Resource: "user_group", Action: "delete", Level: 3, SortOrder: 56, Status: 1},
		// 用户身份映射
		{Code: "auth.identity_mapping.list", Name: "查看身份映射", Description: "查看用户身份映射列表", Module: "auth", Resource: "identity_mapping", Action: "list", Level: 3, SortOrder: 57, Status: 1},
		{Code: "auth.identity_mapping.delete", Name: "删除身份映射", Description: "删除用户身份映射", Module: "auth", Resource: "identity_mapping", Action: "delete", Level: 3, SortOrder: 58, Status: 1},
		// 用户权限查询
		{Code: "auth.permission.query", Name: "查询用户权限", Description: "查询用户有效权限", Module: "auth", Resource: "permission", Action: "query", Level: 3, SortOrder: 59, Status: 1},
		{Code: "auth.permission.view", Name: "查看权限矩阵", Description: "查看用户权限矩阵", Module: "auth", Resource: "permission", Action: "view", Level: 3, SortOrder: 60, Status: 1},
		// 执行记录
		{Code: "auth.binding_execution.list", Name: "查看执行记录", Description: "查看权限绑定执行记录", Module: "auth", Resource: "binding_execution", Action: "list", Level: 3, SortOrder: 61, Status: 1},

		// ========== 资产管理 CMDB 模块（33个） ==========
		// 服务器管理
		{Code: "cmdb.server.list", Name: "服务器列表", Description: "查看服务器列表", Module: "cmdb", Resource: "server", Action: "list", Level: 3, SortOrder: 62, Status: 1},
		{Code: "cmdb.server.view", Name: "查看服务器", Description: "查看服务器详情", Module: "cmdb", Resource: "server", Action: "view", Level: 3, SortOrder: 63, Status: 1},
		{Code: "cmdb.server.create", Name: "创建服务器", Description: "创建新服务器", Module: "cmdb", Resource: "server", Action: "create", Level: 3, SortOrder: 64, Status: 1},
		{Code: "cmdb.server.update", Name: "更新服务器", Description: "更新服务器信息", Module: "cmdb", Resource: "server", Action: "update", Level: 3, SortOrder: 65, Status: 1},
		{Code: "cmdb.server.delete", Name: "删除服务器", Description: "删除服务器", Module: "cmdb", Resource: "server", Action: "delete", Level: 3, SortOrder: 66, Status: 1},
		{Code: "cmdb.server.connect", Name: "连接服务器", Description: "连接到服务器", Module: "cmdb", Resource: "server", Action: "connect", Level: 3, SortOrder: 67, Status: 1},
		// Agent管理
		{Code: "cmdb.agents.list", Name: "Agent列表", Description: "查看Agent列表", Module: "cmdb", Resource: "agents", Action: "list", Level: 3, SortOrder: 68, Status: 1},
		{Code: "cmdb.agents.view", Name: "查看Agent", Description: "查看Agent详细信息", Module: "cmdb", Resource: "agents", Action: "view", Level: 3, SortOrder: 69, Status: 1},
		{Code: "cmdb.agents.deploy", Name: "部署Agent", Description: "部署Agent到服务器", Module: "cmdb", Resource: "agents", Action: "deploy", Level: 3, SortOrder: 70, Status: 1},
		{Code: "cmdb.agents.restart", Name: "重启Agent", Description: "重启Agent服务", Module: "cmdb", Resource: "agents", Action: "restart", Level: 3, SortOrder: 71, Status: 1},
		{Code: "cmdb.agents.uninstall", Name: "卸载Agent", Description: "卸载Agent", Module: "cmdb", Resource: "agents", Action: "uninstall", Level: 3, SortOrder: 72, Status: 1},
		{Code: "cmdb.agents.upgrade", Name: "升级Agent", Description: "升级Agent版本", Module: "cmdb", Resource: "agents", Action: "upgrade", Level: 3, SortOrder: 73, Status: 1},
		// 主机分组管理
		{Code: "cmdb.group.list", Name: "分组列表", Description: "查看分组列表", Module: "cmdb", Resource: "group", Action: "list", Level: 3, SortOrder: 74, Status: 1},
		{Code: "cmdb.group.view", Name: "查看分组", Description: "查看分组详情", Module: "cmdb", Resource: "group", Action: "view", Level: 3, SortOrder: 75, Status: 1},
		{Code: "cmdb.group.create", Name: "创建分组", Description: "创建新分组", Module: "cmdb", Resource: "group", Action: "create", Level: 3, SortOrder: 76, Status: 1},
		{Code: "cmdb.group.update", Name: "更新分组", Description: "更新分组信息", Module: "cmdb", Resource: "group", Action: "update", Level: 3, SortOrder: 77, Status: 1},
		{Code: "cmdb.group.delete", Name: "删除分组", Description: "删除分组", Module: "cmdb", Resource: "group", Action: "delete", Level: 3, SortOrder: 78, Status: 1},
		{Code: "cmdb.group.assign", Name: "分配服务器", Description: "分配服务器到分组", Module: "cmdb", Resource: "group", Action: "assign", Level: 3, SortOrder: 79, Status: 1},
		// 业务系统管理
		{Code: "cmdb.business.list", Name: "业务列表", Description: "查看业务列表", Module: "cmdb", Resource: "business", Action: "list", Level: 3, SortOrder: 80, Status: 1},
		{Code: "cmdb.business.create", Name: "创建业务", Description: "创建新业务", Module: "cmdb", Resource: "business", Action: "create", Level: 3, SortOrder: 81, Status: 1},
		{Code: "cmdb.business.update", Name: "更新业务", Description: "更新业务信息", Module: "cmdb", Resource: "business", Action: "update", Level: 3, SortOrder: 82, Status: 1},
		{Code: "cmdb.business.delete", Name: "删除业务", Description: "删除业务", Module: "cmdb", Resource: "business", Action: "delete", Level: 3, SortOrder: 83, Status: 1},
		// 机房管理
		{Code: "cmdb.rooms.list", Name: "机房列表", Description: "查看机房列表", Module: "cmdb", Resource: "rooms", Action: "list", Level: 3, SortOrder: 84, Status: 1},
		{Code: "cmdb.rooms.create", Name: "创建机房", Description: "创建新机房", Module: "cmdb", Resource: "rooms", Action: "create", Level: 3, SortOrder: 85, Status: 1},
		{Code: "cmdb.rooms.update", Name: "更新机房", Description: "更新机房信息", Module: "cmdb", Resource: "rooms", Action: "update", Level: 3, SortOrder: 86, Status: 1},
		{Code: "cmdb.rooms.delete", Name: "删除机房", Description: "删除机房", Module: "cmdb", Resource: "rooms", Action: "delete", Level: 3, SortOrder: 87, Status: 1},
		// 标签管理
		{Code: "cmdb.tags.list", Name: "标签列表", Description: "查看标签列表", Module: "cmdb", Resource: "tags", Action: "list", Level: 3, SortOrder: 88, Status: 1},
		{Code: "cmdb.tags.create", Name: "创建标签", Description: "创建新标签", Module: "cmdb", Resource: "tags", Action: "create", Level: 3, SortOrder: 89, Status: 1},
		{Code: "cmdb.tags.update", Name: "更新标签", Description: "更新标签信息", Module: "cmdb", Resource: "tags", Action: "update", Level: 3, SortOrder: 90, Status: 1},
		{Code: "cmdb.tags.delete", Name: "删除标签", Description: "删除标签", Module: "cmdb", Resource: "tags", Action: "delete", Level: 3, SortOrder: 91, Status: 1},
		// 会话管理
		{Code: "cmdb.session.list", Name: "会话列表", Description: "查看堡垒机会话列表", Module: "cmdb", Resource: "session", Action: "list", Level: 3, SortOrder: 92, Status: 1},
		{Code: "cmdb.session.view", Name: "查看会话", Description: "查看会话详细信息", Module: "cmdb", Resource: "session", Action: "view", Level: 3, SortOrder: 93, Status: 1},
		{Code: "cmdb.session.terminate", Name: "终止会话", Description: "终止堡垒机会话", Module: "cmdb", Resource: "session", Action: "terminate", Level: 3, SortOrder: 94, Status: 1},
		// 访问策略管理
		{Code: "cmdb.access_policy.list", Name: "访问策略列表", Description: "查看访问策略列表", Module: "cmdb", Resource: "access_policy", Action: "list", Level: 3, SortOrder: 95, Status: 1},
		{Code: "cmdb.access_policy.create", Name: "创建访问策略", Description: "创建新的访问策略", Module: "cmdb", Resource: "access_policy", Action: "create", Level: 3, SortOrder: 96, Status: 1},
		{Code: "cmdb.access_policy.update", Name: "更新访问策略", Description: "更新访问策略信息", Module: "cmdb", Resource: "access_policy", Action: "update", Level: 3, SortOrder: 97, Status: 1},
		{Code: "cmdb.access_policy.delete", Name: "删除访问策略", Description: "删除访问策略", Module: "cmdb", Resource: "access_policy", Action: "delete", Level: 3, SortOrder: 98, Status: 1},
		// 凭据管理
		{Code: "cmdb.credential.create", Name: "创建凭据", Description: "创建SSH/访问凭据", Module: "cmdb", Resource: "credential", Action: "create", Level: 3, SortOrder: 132, Status: 1},
		{Code: "cmdb.credential.update", Name: "更新凭据", Description: "更新凭据信息", Module: "cmdb", Resource: "credential", Action: "update", Level: 3, SortOrder: 133, Status: 1},
		{Code: "cmdb.credential.delete", Name: "删除凭据", Description: "删除凭据", Module: "cmdb", Resource: "credential", Action: "delete", Level: 3, SortOrder: 134, Status: 1},
		{Code: "cmdb.credential.test", Name: "测试凭据", Description: "测试凭据连通性", Module: "cmdb", Resource: "credential", Action: "test", Level: 3, SortOrder: 135, Status: 1},
		// Agent 记录与版本管理
		{Code: "cmdb.agents.delete", Name: "删除Agent记录", Description: "删除Agent记录（仅清记录，不卸载）", Module: "cmdb", Resource: "agents", Action: "delete", Level: 3, SortOrder: 136, Status: 1},
		{Code: "cmdb.agent_version.manage", Name: "管理Agent版本", Description: "新增/更新/删除Agent版本记录", Module: "cmdb", Resource: "agent_version", Action: "manage", Level: 3, SortOrder: 137, Status: 1},

		// ========== 监控中心模块（16个） ==========
		// 监控数据
		{Code: "monitor.data.view", Name: "查看监控数据", Description: "查看监控中心数据", Module: "monitor", Resource: "data", Action: "view", Level: 3, SortOrder: 99, Status: 1},
		{Code: "monitor.data.refresh", Name: "刷新监控数据", Description: "刷新监控中心数据", Module: "monitor", Resource: "data", Action: "refresh", Level: 3, SortOrder: 100, Status: 1},
		{Code: "monitor.data.export", Name: "导出监控数据", Description: "导出监控中心数据", Module: "monitor", Resource: "data", Action: "export", Level: 3, SortOrder: 101, Status: 1},
		// 告警管理
		{Code: "monitor.alert.list", Name: "告警列表", Description: "查看告警列表", Module: "monitor", Resource: "alert", Action: "list", Level: 3, SortOrder: 102, Status: 1},
		{Code: "monitor.alert.view", Name: "查看告警", Description: "查看告警详情", Module: "monitor", Resource: "alert", Action: "view", Level: 3, SortOrder: 103, Status: 1},
		{Code: "monitor.alert.ack", Name: "确认告警", Description: "确认告警事件", Module: "monitor", Resource: "alert", Action: "ack", Level: 3, SortOrder: 104, Status: 1},
		{Code: "monitor.alert.handle", Name: "处理告警", Description: "处理告警事件", Module: "monitor", Resource: "alert", Action: "handle", Level: 3, SortOrder: 105, Status: 1},
		// 监控任务
		{Code: "monitor.task.list", Name: "任务列表", Description: "查看任务列表", Module: "monitor", Resource: "task", Action: "list", Level: 3, SortOrder: 106, Status: 1},
		{Code: "monitor.task.create", Name: "创建任务", Description: "创建新任务", Module: "monitor", Resource: "task", Action: "create", Level: 3, SortOrder: 107, Status: 1},
		{Code: "monitor.task.update", Name: "更新任务", Description: "更新任务信息", Module: "monitor", Resource: "task", Action: "update", Level: 3, SortOrder: 108, Status: 1},
		{Code: "monitor.task.delete", Name: "删除任务", Description: "删除任务", Module: "monitor", Resource: "task", Action: "delete", Level: 3, SortOrder: 109, Status: 1},
		{Code: "monitor.task.execute", Name: "执行任务", Description: "执行监控任务", Module: "monitor", Resource: "task", Action: "execute", Level: 3, SortOrder: 110, Status: 1},
		// 巡检报告
		{Code: "monitor.report.list", Name: "报告列表", Description: "查看巡检报告列表", Module: "monitor", Resource: "report", Action: "list", Level: 3, SortOrder: 111, Status: 1},
		{Code: "monitor.report.view", Name: "查看报告", Description: "查看巡检报告详情", Module: "monitor", Resource: "report", Action: "view", Level: 3, SortOrder: 112, Status: 1},
		{Code: "monitor.report.create", Name: "创建报告", Description: "创建新的巡检报告", Module: "monitor", Resource: "report", Action: "create", Level: 3, SortOrder: 113, Status: 1},
		{Code: "monitor.report.delete", Name: "删除报告", Description: "删除巡检报告", Module: "monitor", Resource: "report", Action: "delete", Level: 3, SortOrder: 114, Status: 1},
		// 告警规则管理
		{Code: "monitor.alert_rule.create", Name: "创建告警规则", Description: "创建告警规则", Module: "monitor", Resource: "alert_rule", Action: "create", Level: 3, SortOrder: 138, Status: 1},
		{Code: "monitor.alert_rule.update", Name: "更新告警规则", Description: "更新/启停告警规则", Module: "monitor", Resource: "alert_rule", Action: "update", Level: 3, SortOrder: 139, Status: 1},
		{Code: "monitor.alert_rule.delete", Name: "删除告警规则", Description: "删除告警规则", Module: "monitor", Resource: "alert_rule", Action: "delete", Level: 3, SortOrder: 140, Status: 1},
		// 通知渠道管理
		{Code: "monitor.notification.create", Name: "创建通知渠道", Description: "创建通知渠道", Module: "monitor", Resource: "notification", Action: "create", Level: 3, SortOrder: 141, Status: 1},
		{Code: "monitor.notification.update", Name: "更新通知渠道", Description: "更新/启停通知渠道", Module: "monitor", Resource: "notification", Action: "update", Level: 3, SortOrder: 142, Status: 1},
		{Code: "monitor.notification.delete", Name: "删除通知渠道", Description: "删除通知渠道", Module: "monitor", Resource: "notification", Action: "delete", Level: 3, SortOrder: 143, Status: 1},
		{Code: "monitor.notification.test", Name: "测试通知渠道", Description: "发送测试通知", Module: "monitor", Resource: "notification", Action: "test", Level: 3, SortOrder: 144, Status: 1},

		// ========== K8s管理模块（15个） ==========
		// 集群管理（list=单接口权限：仅集群列表端点；view=接口权限集合：列表+详情+nodes+namespaces）
		{Code: "k8s.cluster.list", Name: "集群列表", Description: "集群列表接口权限（单接口）", Module: "k8s", Resource: "cluster", Action: "list", Level: 3, SortOrder: 115, Status: 1},
		{Code: "k8s.cluster.view", Name: "查看集群", Description: "查看集群列表与详情（接口集合）", Module: "k8s", Resource: "cluster", Action: "view", Level: 3, SortOrder: 116, Status: 1},
		{Code: "k8s.cluster.create", Name: "创建集群", Description: "创建新集群", Module: "k8s", Resource: "cluster", Action: "create", Level: 3, SortOrder: 117, Status: 1},
		{Code: "k8s.cluster.update", Name: "更新集群", Description: "更新集群信息", Module: "k8s", Resource: "cluster", Action: "update", Level: 3, SortOrder: 118, Status: 1},
		{Code: "k8s.cluster.delete", Name: "删除集群", Description: "删除集群", Module: "k8s", Resource: "cluster", Action: "delete", Level: 3, SortOrder: 119, Status: 1},
		{Code: "k8s.cluster.connect", Name: "连接集群", Description: "连接到集群", Module: "k8s", Resource: "cluster", Action: "connect", Level: 3, SortOrder: 120, Status: 1},
		// K8s权限管理
		{Code: "k8s.permission.list", Name: "权限列表", Description: "查看K8s权限列表", Module: "k8s", Resource: "permission", Action: "list", Level: 3, SortOrder: 121, Status: 1},
		{Code: "k8s.permission.assign", Name: "分配权限", Description: "分配K8s权限给用户", Module: "k8s", Resource: "permission", Action: "assign", Level: 3, SortOrder: 122, Status: 1},
		{Code: "k8s.permission.revoke", Name: "撤销权限", Description: "撤销用户的K8s权限", Module: "k8s", Resource: "permission", Action: "revoke", Level: 3, SortOrder: 123, Status: 1},
		// K8s资源管理
		{Code: "k8s.resource.view", Name: "查看资源", Description: "查看K8s资源详情", Module: "k8s", Resource: "resource", Action: "view", Level: 3, SortOrder: 124, Status: 1},
		{Code: "k8s.resource.create", Name: "创建资源", Description: "创建新的K8s资源", Module: "k8s", Resource: "resource", Action: "create", Level: 3, SortOrder: 125, Status: 1},
		{Code: "k8s.resource.update", Name: "更新资源", Description: "更新K8s资源信息", Module: "k8s", Resource: "resource", Action: "update", Level: 3, SortOrder: 126, Status: 1},
		{Code: "k8s.resource.delete", Name: "删除资源", Description: "删除K8s资源", Module: "k8s", Resource: "resource", Action: "delete", Level: 3, SortOrder: 127, Status: 1},

		// 终端
		{Code: "k8s.terminal.connect", Name: "终端连接", Description: "连接Pod终端与会话管理", Module: "k8s", Resource: "terminal", Action: "connect", Level: 3, SortOrder: 127, Status: 1},
		// K8s诊断
		{Code: "k8s.diagnostic.execute", Name: "执行诊断", Description: "执行K8s诊断命令", Module: "k8s", Resource: "diagnostic", Action: "execute", Level: 3, SortOrder: 128, Status: 1},
		{Code: "k8s.diagnostic.view", Name: "查看诊断结果", Description: "查看K8s诊断结果和历史", Module: "k8s", Resource: "diagnostic", Action: "view", Level: 3, SortOrder: 129, Status: 1},
		// K8s原生 RBAC 代管（A 模式）
		{Code: "k8s.rbac.view", Name: "查看原生RBAC", Description: "查看集群内原生 ClusterRole/Role/Binding", Module: "k8s", Resource: "rbac", Action: "view", Level: 3, SortOrder: 130, Status: 1},
		{Code: "k8s.rbac.manage", Name: "管理原生RBAC", Description: "编辑集群内原生 ClusterRole/Role 规则", Module: "k8s", Resource: "rbac", Action: "manage", Level: 3, SortOrder: 131, Status: 1},

		// ========== 工单中心模块（11个） ==========
		// 工单：view=工单中心菜单可见（发起/待办/我相关工单）；list=查看全部工单（scope=all 与任意详情）
		{Code: "ticket.ticket.view", Name: "工单中心入口", Description: "查看工单中心菜单，发起与处理自己相关的工单", Module: "ticket", Resource: "ticket", Action: "view", Level: 3, SortOrder: 150, Status: 1},
		{Code: "ticket.ticket.list", Name: "查看全部工单", Description: "查看全部工单列表与任意工单详情", Module: "ticket", Resource: "ticket", Action: "list", Level: 3, SortOrder: 151, Status: 1},
		{Code: "ticket.ticket.reassign", Name: "改派审批人", Description: "改派工单当前节点的审批人（处理审批人离职/请假导致的节点卡死）", Module: "ticket", Resource: "ticket", Action: "reassign", Level: 3, SortOrder: 160, Status: 1},
		// 工单类型管理
		{Code: "ticket.type.list", Name: "工单类型列表", Description: "查看工单类型列表", Module: "ticket", Resource: "type", Action: "list", Level: 3, SortOrder: 152, Status: 1},
		{Code: "ticket.type.create", Name: "创建工单类型", Description: "创建新的工单类型（场景）", Module: "ticket", Resource: "type", Action: "create", Level: 3, SortOrder: 153, Status: 1},
		{Code: "ticket.type.update", Name: "更新工单类型", Description: "更新工单类型信息", Module: "ticket", Resource: "type", Action: "update", Level: 3, SortOrder: 154, Status: 1},
		{Code: "ticket.type.delete", Name: "删除工单类型", Description: "删除工单类型", Module: "ticket", Resource: "type", Action: "delete", Level: 3, SortOrder: 155, Status: 1},
		// 流程定义管理
		{Code: "ticket.workflow.list", Name: "流程定义列表", Description: "查看审批流程定义列表", Module: "ticket", Resource: "workflow", Action: "list", Level: 3, SortOrder: 156, Status: 1},
		{Code: "ticket.workflow.create", Name: "创建流程", Description: "创建审批流程定义", Module: "ticket", Resource: "workflow", Action: "create", Level: 3, SortOrder: 157, Status: 1},
		{Code: "ticket.workflow.update", Name: "更新流程", Description: "更新审批流程定义", Module: "ticket", Resource: "workflow", Action: "update", Level: 3, SortOrder: 158, Status: 1},
		{Code: "ticket.workflow.delete", Name: "删除流程", Description: "删除审批流程定义", Module: "ticket", Resource: "workflow", Action: "delete", Level: 3, SortOrder: 159, Status: 1},
		{Code: "ticket.notify.list", Name: "通知设置查看", Description: "查看工单通知事件矩阵", Module: "ticket", Resource: "notify", Action: "list", Level: 3, SortOrder: 160, Status: 1},
		{Code: "ticket.notify.update", Name: "通知设置更新", Description: "配置工单通知事件矩阵", Module: "ticket", Resource: "notify", Action: "update", Level: 3, SortOrder: 161, Status: 1},
	}

	for _, perm := range permissions {
		var existing modelsystem.Permission
		err := db.Where("code = ?", perm.Code).First(&existing).Error

		if err == nil {
			db.Model(&existing).Updates(map[string]interface{}{
				"name":        perm.Name,
				"description": perm.Description,
				"module":      perm.Module,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"level":       perm.Level,
				"sort_order":  perm.SortOrder,
				"status":      perm.Status,
			})
		} else {
			if err := db.Create(&perm).Error; err != nil {
				logger.Error("创建权限失败", zap.String("code", perm.Code), zap.Error(err))
			}
		}
	}

	var count int64
	db.Model(&modelsystem.Permission{}).Count(&count)
	logger.Info("权限数据同步完成", zap.Int64("total_permissions", count))

	return nil
}

// permissionModuleNames 模块（Level 1）节点中文名
var permissionModuleNames = map[string]string{
	"system":  "系统管理",
	"cmdb":    "资产管理",
	"audit":   "审计中心",
	"auth":    "授权中心",
	"monitor": "监控中心",
	"k8s":     "K8s管理",
	"ticket":  "工单中心",
}

// permissionResourceNames 资源（Level 2）节点中文名，键为 module.resource
var permissionResourceNames = map[string]string{
	"system.menu": "菜单管理", "system.role": "角色管理", "system.user": "用户管理",
	"system.permission": "权限管理", "system.route": "路由管理",
	"audit.login_log": "登录日志", "audit.operation_log": "操作日志",
	"audit.system_event": "系统事件", "audit.stats": "审计统计",
	"auth.application": "应用管理", "auth.user": "授权用户", "auth.group": "用户组",
	"auth.group_binding": "用户组绑定", "auth.user_group": "用户组成员",
	"auth.identity_mapping": "身份映射", "auth.permission": "权限查询",
	"auth.binding_execution": "执行记录",
	"cmdb.attribute":         "属性管理", "cmdb.server": "服务器管理", "cmdb.agents": "Agent管理",
	"cmdb.group": "主机分组", "cmdb.business": "业务管理", "cmdb.rooms": "机房管理",
	"cmdb.tags": "标签管理", "cmdb.session": "会话管理", "cmdb.access_policy": "访问策略",
	"monitor.data": "监控数据", "monitor.alert": "告警管理",
	"monitor.task": "监控任务", "monitor.report": "巡检报告",
	"k8s.cluster": "集群管理", "k8s.permission": "K8s权限", "k8s.resource": "资源管理",
	"k8s.terminal": "终端连接", "k8s.diagnostic": "诊断中心", "k8s.rbac": "原生RBAC",
	"ticket.ticket": "工单管理", "ticket.type": "工单类型", "ticket.workflow": "流程定义",
}

// syncPermissionHierarchy 补齐权限树层级并回填 parent_id（幂等）：
// Level 3 按钮权限由 seed 平铺写入，此处按 module / module.resource 约定
// 补建 Level 1 模块、Level 2 资源分组节点，并把 Level 3 挂到对应资源节点下。
// 分组节点仅作树形展示，不参与鉴权（角色只绑定 Level 3 权限）。
func (i *Initializer) syncPermissionHierarchy() error {
	db := database.GetDB()

	// 1. 按模块去重，补建 Level 1 节点（code = module）
	var modules []string
	if err := db.Model(&modelsystem.Permission{}).Where("level = 3").
		Distinct().Pluck("module", &modules).Error; err != nil {
		return err
	}

	moduleIDs := make(map[string]uint, len(modules))
	for idx, module := range modules {
		name := permissionModuleNames[module]
		if name == "" {
			name = module
		}
		var node modelsystem.Permission
		if err := db.Where("code = ? AND level = 1", module).First(&node).Error; err != nil {
			node = modelsystem.Permission{
				Code: module, Name: name, Module: module, Resource: module,
				Level: 1, SortOrder: idx + 1, Status: 1,
			}
			if err := db.Create(&node).Error; err != nil {
				logger.Error("创建模块权限节点失败", zap.String("module", module), zap.Error(err))
				continue
			}
		}
		moduleIDs[module] = node.ID
	}

	// 2. 按 module.resource 去重，补建 Level 2 节点（code = module.resource）
	type resourceGroup struct {
		Module   string
		Resource string
		MinSort  int
	}
	var groups []resourceGroup
	if err := db.Model(&modelsystem.Permission{}).Where("level = 3").
		Select("module, resource, MIN(sort_order) AS min_sort").
		Group("module, resource").Order("min_sort ASC").Scan(&groups).Error; err != nil {
		return err
	}

	resourceIDs := make(map[string]uint, len(groups))
	for _, g := range groups {
		code := g.Module + "." + g.Resource
		name := permissionResourceNames[code]
		if name == "" {
			name = g.Resource
		}
		parentID := moduleIDs[g.Module]

		var node modelsystem.Permission
		err := db.Where("code = ? AND level = 2", code).First(&node).Error
		if err != nil {
			node = modelsystem.Permission{
				Code: code, Name: name, Module: g.Module, Resource: g.Resource,
				Level: 2, ParentID: &parentID, SortOrder: g.MinSort, Status: 1,
			}
			if err := db.Create(&node).Error; err != nil {
				logger.Error("创建资源权限节点失败", zap.String("code", code), zap.Error(err))
				continue
			}
		} else if node.ParentID == nil || *node.ParentID != parentID {
			// 已存在但父级缺失/错位，修正
			db.Model(&node).Update("parent_id", parentID)
		}
		resourceIDs[code] = node.ID
	}

	// 3. 回填 Level 3 的 parent_id
	var pending []modelsystem.Permission
	if err := db.Where("level = 3 AND parent_id IS NULL").Find(&pending).Error; err != nil {
		return err
	}
	backfilled := 0
	for _, p := range pending {
		parentID, ok := resourceIDs[p.Module+"."+p.Resource]
		if !ok {
			continue
		}
		if err := db.Model(&modelsystem.Permission{ID: p.ID}).Update("parent_id", parentID).Error; err == nil {
			backfilled++
		}
	}

	logger.Info("权限层级同步完成",
		zap.Int("modules", len(moduleIDs)),
		zap.Int("resources", len(resourceIDs)),
		zap.Int("backfilled", backfilled))
	return nil
}

// syncDefaultPermissions 为内置角色分配默认权限
func (i *Initializer) syncDefaultPermissions() error {
	logger.Info("开始为内置角色分配默认权限...")

	db := database.GetDB()

	rolePermissions := map[string][]string{
		"admin": {"*.*.*"},
		"ops": {
			"system.user.list", "system.user.create", "system.user.update", "system.user.delete",
			"system.role.list", "system.role.update",
			"system.menu.list",
			"cmdb.server.list", "cmdb.server.view", "cmdb.server.create", "cmdb.server.update", "cmdb.server.delete", "cmdb.server.connect",
			"cmdb.business.list", "cmdb.business.create", "cmdb.business.update", "cmdb.business.delete",
			"cmdb.rooms.list", "cmdb.rooms.create", "cmdb.rooms.update", "cmdb.rooms.delete",
			"cmdb.tags.list", "cmdb.tags.create", "cmdb.tags.update", "cmdb.tags.delete",
			"cmdb.group.list", "cmdb.group.view", "cmdb.group.create", "cmdb.group.update", "cmdb.group.delete", "cmdb.group.assign",
			"cmdb.agents.list", "cmdb.agents.deploy", "cmdb.agents.restart", "cmdb.agents.uninstall",
			"monitor.data.view", "monitor.data.export",
			"monitor.alert.list", "monitor.alert.ack", "monitor.alert.handle",
			"monitor.task.list", "monitor.task.view", "monitor.task.create", "monitor.task.update", "monitor.task.delete", "monitor.task.execute",
			"k8s.cluster.list", "k8s.cluster.view", "k8s.cluster.create", "k8s.cluster.update", "k8s.cluster.delete", "k8s.cluster.connect",
			"k8s.resource.view", "k8s.resource.create", "k8s.resource.update", "k8s.resource.delete",
			"k8s.terminal.connect",
			"k8s.diagnostic.view", "k8s.diagnostic.execute",
			"k8s.permission.list", "k8s.permission.assign", "k8s.permission.revoke",
			"k8s.rbac.view", "k8s.rbac.manage",
			"audit.login_log.list", "audit.login_log.export",
			"audit.operation_log.list", "audit.operation_log.export",
			"audit.system_event.list",
			"audit.stats.view",
			// 工单中心：入口 + 全部工单查看 + 类型/流程管理 + 改派审批人 + 通知设置
			"ticket.ticket.view", "ticket.ticket.list", "ticket.ticket.reassign",
			"ticket.type.list", "ticket.type.create", "ticket.type.update", "ticket.type.delete",
			"ticket.workflow.list", "ticket.workflow.create", "ticket.workflow.update", "ticket.workflow.delete",
			"ticket.notify.list", "ticket.notify.update",
		},
		"auditor": {
			"cmdb.server.list", "cmdb.server.view",
			"cmdb.business.list",
			"cmdb.rooms.list",
			"cmdb.tags.list",
			"cmdb.group.list", "cmdb.group.view",
			"cmdb.agents.list",
			"monitor.data.view",
			"monitor.alert.list",
			"monitor.task.list", "monitor.task.view",
			"k8s.cluster.list", "k8s.cluster.view",
			"k8s.resource.view",
			"k8s.permission.list",
			"audit.login_log.list",
			"audit.operation_log.list",
			"audit.system_event.list",
			"audit.stats.view",
			// 工单中心：审计视角可看全部工单
			"ticket.ticket.view", "ticket.ticket.list",
		},
		"viewer": {
			"cmdb.server.list",
			"monitor.data.view",
		},
		"k8s_view": {
			// K8s 只读：可看集群/资源/诊断与授权绑定列表；授权操作（assign/revoke）归管理员角色
			"k8s.cluster.list", "k8s.cluster.view",
			"k8s.resource.view",
			"k8s.diagnostic.view",
			"k8s.permission.list",
		},
		"user": {
			"monitor.data.view",
			// 工单中心入口（发起/待办/我相关的工单）
			"ticket.ticket.view",
		},
		"test": {
			"system.user.list", "system.user.create",
			"system.role.list",
			"cmdb.server.list", "cmdb.server.create",
			"monitor.data.view",
			"k8s.cluster.list",
		},
	}

	for roleCode, permissionCodes := range rolePermissions {
		var role modelsystem.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			logger.Warn("角色不存在，跳过权限分配", zap.String("code", roleCode), zap.Error(err))
			continue
		}

		var permissions []modelsystem.Permission
		if err := db.Where("code IN ?", permissionCodes).Find(&permissions).Error; err != nil {
			logger.Warn("查询权限失败", zap.String("role", roleCode), zap.Error(err))
			continue
		}

		assignedCount := 0
		for _, perm := range permissions {
			var rolePerm modelsystem.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rolePerm).Error
			if err != nil {
				rolePerm = modelsystem.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
				if err := db.Create(&rolePerm).Error; err != nil {
					logger.Error("分配权限失败", zap.String("role", roleCode), zap.String("permission", perm.Code), zap.Error(err))
				} else {
					assignedCount++
				}
			}
		}

		// 只增不删：保留管理员对内置角色的手工授权调整，
		// 避免每次重启把 seed 清单外的绑定清空（角色定义收紧请走角色管理界面，
		// 改动会经 RoleService 同步 Casbin）
		logger.Info("角色权限分配完成",
			zap.String("role", roleCode),
			zap.Int("总权限数", len(permissions)),
			zap.Int("新分配", assignedCount))
	}

	return nil
}

// syncAPIPermissions 同步层级权限代码到Casbin
func (i *Initializer) syncAPIPermissions() error {
	logger.Info("开始同步权限代码到Casbin...")

	permService, err := GetPermissionService()
	if err != nil {
		logger.Warn("获取PermissionService失败，跳过权限初始化", zap.Error(err))
		return nil
	}

	if err := permService.ClearAllPolicies(); err != nil {
		logger.Error("清除旧策略失败", zap.Error(err))
		return err
	}

	if err := permService.SyncAllRolesToCasbin(); err != nil {
		logger.Error("同步角色权限失败", zap.Error(err))
		return err
	}

	policies, err := permService.GetAllPolicies()
	if err != nil {
		logger.Error("获取Casbin策略失败", zap.Error(err))
		return err
	}
	logger.Info("Casbin权限初始化完成", zap.Int("总策略数", len(policies)))

	return nil
}
