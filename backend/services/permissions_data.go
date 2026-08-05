package services

import "oneops/backend/models"

// 权限数据定义文件
// 包含系统中所有模块的权限代码定义

// GetAllSystemPermissions 获取系统中所有模块的权限定义
func GetAllSystemPermissions() []models.Permission {
	return []models.Permission{
		// ========== 系统管理模块 ==========
		{Code: "system", Name: "系统管理", Module: "system", Resource: "", Action: "", Level: 1, Status: 1},

		// 用户管理
		{Code: "system.user", Name: "用户管理", Module: "system", Resource: "user", Action: "", Level: 2, Status: 1},
		{Code: "system.user.list", Name: "用户列表", Module: "system", Resource: "user", Action: "list", Level: 3, Status: 1},
		{Code: "system.user.view", Name: "查看用户", Module: "system", Resource: "user", Action: "view", Level: 3, Status: 1},
		{Code: "system.user.create", Name: "创建用户", Module: "system", Resource: "user", Action: "create", Level: 3, Status: 1},
		{Code: "system.user.update", Name: "更新用户", Module: "system", Resource: "user", Action: "update", Level: 3, Status: 1},
		{Code: "system.user.delete", Name: "删除用户", Module: "system", Resource: "user", Action: "delete", Level: 3, Status: 1},
		{Code: "system.user.reset_password", Name: "重置密码", Module: "system", Resource: "user", Action: "reset_password", Level: 3, Status: 1},

		// 角色管理
		{Code: "system.role", Name: "角色管理", Module: "system", Resource: "role", Action: "", Level: 2, Status: 1},
		{Code: "system.role.list", Name: "角色列表", Module: "system", Resource: "role", Action: "list", Level: 3, Status: 1},
		{Code: "system.role.view", Name: "查看角色", Module: "system", Resource: "role", Action: "view", Level: 3, Status: 1},
		{Code: "system.role.create", Name: "创建角色", Module: "system", Resource: "role", Action: "create", Level: 3, Status: 1},
		{Code: "system.role.update", Name: "更新角色", Module: "system", Resource: "role", Action: "update", Level: 3, Status: 1},
		{Code: "system.role.delete", Name: "删除角色", Module: "system", Resource: "role", Action: "delete", Level: 3, Status: 1},
		{Code: "system.role.assign_permissions", Name: "分配权限", Module: "system", Resource: "role", Action: "assign_permissions", Level: 3, Status: 1},

		// 菜单管理
		{Code: "system.menu", Name: "菜单管理", Module: "system", Resource: "menu", Action: "", Level: 2, Status: 1},
		{Code: "system.menu.list", Name: "菜单列表", Module: "system", Resource: "menu", Action: "list", Level: 3, Status: 1},
		{Code: "system.menu.view", Name: "查看菜单", Module: "system", Resource: "menu", Action: "view", Level: 3, Status: 1},
		{Code: "system.menu.create", Name: "创建菜单", Module: "system", Resource: "menu", Action: "create", Level: 3, Status: 1},
		{Code: "system.menu.update", Name: "更新菜单", Module: "system", Resource: "menu", Action: "update", Level: 3, Status: 1},
		{Code: "system.menu.delete", Name: "删除菜单", Module: "system", Resource: "menu", Action: "delete", Level: 3, Status: 1},

		// 权限管理
		{Code: "system.permission", Name: "权限管理", Module: "system", Resource: "permission", Action: "", Level: 2, Status: 1},
		{Code: "system.permission.list", Name: "权限列表", Module: "system", Resource: "permission", Action: "list", Level: 3, Status: 1},
		{Code: "system.permission.view", Name: "查看权限", Module: "system", Resource: "permission", Action: "view", Level: 3, Status: 1},
		{Code: "system.permission.create", Name: "创建权限", Module: "system", Resource: "permission", Action: "create", Level: 3, Status: 1},
		{Code: "system.permission.update", Name: "更新权限", Module: "system", Resource: "permission", Action: "update", Level: 3, Status: 1},
		{Code: "system.permission.delete", Name: "删除权限", Module: "system", Resource: "permission", Action: "delete", Level: 3, Status: 1},

		// ========== CMDB模块 ==========
		{Code: "cmdb", Name: "CMDB管理", Module: "cmdb", Resource: "", Action: "", Level: 1, Status: 1},

		// 服务器管理
		{Code: "cmdb.server", Name: "服务器管理", Module: "cmdb", Resource: "server", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.server.list", Name: "服务器列表", Module: "cmdb", Resource: "server", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.server.view", Name: "查看服务器", Module: "cmdb", Resource: "server", Action: "view", Level: 3, Status: 1},
		{Code: "cmdb.server.create", Name: "创建服务器", Module: "cmdb", Resource: "server", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.server.update", Name: "更新服务器", Module: "cmdb", Resource: "server", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.server.delete", Name: "删除服务器", Module: "cmdb", Resource: "server", Action: "delete", Level: 3, Status: 1},
		{Code: "cmdb.server.connect", Name: "连接服务器", Module: "cmdb", Resource: "server", Action: "connect", Level: 3, Status: 1},

		// 分组管理
		{Code: "cmdb.group", Name: "分组管理", Module: "cmdb", Resource: "group", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.group.list", Name: "分组列表", Module: "cmdb", Resource: "group", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.group.view", Name: "查看分组", Module: "cmdb", Resource: "group", Action: "view", Level: 3, Status: 1},
		{Code: "cmdb.group.create", Name: "创建分组", Module: "cmdb", Resource: "group", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.group.update", Name: "更新分组", Module: "cmdb", Resource: "group", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.group.delete", Name: "删除分组", Module: "cmdb", Resource: "group", Action: "delete", Level: 3, Status: 1},
		{Code: "cmdb.group.assign", Name: "分配服务器", Module: "cmdb", Resource: "group", Action: "assign", Level: 3, Status: 1},

		// 业务管理
		{Code: "cmdb.business", Name: "业务管理", Module: "cmdb", Resource: "business", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.business.list", Name: "业务列表", Module: "cmdb", Resource: "business", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.business.view", Name: "查看业务", Module: "cmdb", Resource: "business", Action: "view", Level: 3, Status: 1},
		{Code: "cmdb.business.create", Name: "创建业务", Module: "cmdb", Resource: "business", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.business.update", Name: "更新业务", Module: "cmdb", Resource: "business", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.business.delete", Name: "删除业务", Module: "cmdb", Resource: "business", Action: "delete", Level: 3, Status: 1},

		// 机房管理
		{Code: "cmdb.rooms", Name: "机房管理", Module: "cmdb", Resource: "rooms", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.rooms.list", Name: "机房列表", Module: "cmdb", Resource: "rooms", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.rooms.view", Name: "查看机房", Module: "cmdb", Resource: "rooms", Action: "view", Level: 3, Status: 1},
		{Code: "cmdb.rooms.create", Name: "创建机房", Module: "cmdb", Resource: "rooms", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.rooms.update", Name: "更新机房", Module: "cmdb", Resource: "rooms", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.rooms.delete", Name: "删除机房", Module: "cmdb", Resource: "rooms", Action: "delete", Level: 3, Status: 1},

		// 标签管理
		{Code: "cmdb.tags", Name: "标签管理", Module: "cmdb", Resource: "tags", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.tags.list", Name: "标签列表", Module: "cmdb", Resource: "tags", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.tags.view", Name: "查看标签", Module: "cmdb", Resource: "tags", Action: "view", Level: 3, Status: 1},
		{Code: "cmdb.tags.create", Name: "创建标签", Module: "cmdb", Resource: "tags", Action: "create", Level: 3, Status: 1},
		{Code: "cmdb.tags.update", Name: "更新标签", Module: "cmdb", Resource: "tags", Action: "update", Level: 3, Status: 1},
		{Code: "cmdb.tags.delete", Name: "删除标签", Module: "cmdb", Resource: "tags", Action: "delete", Level: 3, Status: 1},

		// Agent管理
		{Code: "cmdb.agents", Name: "Agent管理", Module: "cmdb", Resource: "agents", Action: "", Level: 2, Status: 1},
		{Code: "cmdb.agents.list", Name: "Agent列表", Module: "cmdb", Resource: "agents", Action: "list", Level: 3, Status: 1},
		{Code: "cmdb.agents.deploy", Name: "部署Agent", Module: "cmdb", Resource: "agents", Action: "deploy", Level: 3, Status: 1},
		{Code: "cmdb.agents.restart", Name: "重启Agent", Module: "cmdb", Resource: "agents", Action: "restart", Level: 3, Status: 1},
		{Code: "cmdb.agents.uninstall", Name: "卸载Agent", Module: "cmdb", Resource: "agents", Action: "uninstall", Level: 3, Status: 1},

		// ========== 监控模块 ==========
		{Code: "monitor", Name: "监控管理", Module: "monitor", Resource: "", Action: "", Level: 1, Status: 1},

		// 监控数据
		{Code: "monitor.data", Name: "监控数据", Module: "monitor", Resource: "data", Action: "", Level: 2, Status: 1},
		{Code: "monitor.data.view", Name: "查看监控数据", Module: "monitor", Resource: "data", Action: "view", Level: 3, Status: 1},
		{Code: "monitor.data.export", Name: "导出监控数据", Module: "monitor", Resource: "data", Action: "export", Level: 3, Status: 1},

		// 告警管理
		{Code: "monitor.alert", Name: "告警管理", Module: "monitor", Resource: "alert", Action: "", Level: 2, Status: 1},
		{Code: "monitor.alert.list", Name: "告警列表", Module: "monitor", Resource: "alert", Action: "list", Level: 3, Status: 1},
		{Code: "monitor.alert.view", Name: "查看告警", Module: "monitor", Resource: "alert", Action: "view", Level: 3, Status: 1},
		{Code: "monitor.alert.ack", Name: "确认告警", Module: "monitor", Resource: "alert", Action: "ack", Level: 3, Status: 1},
		{Code: "monitor.alert.handle", Name: "处理告警", Module: "monitor", Resource: "alert", Action: "handle", Level: 3, Status: 1},

		// 监控任务
		{Code: "monitor.task", Name: "监控任务", Module: "monitor", Resource: "task", Action: "", Level: 2, Status: 1},
		{Code: "monitor.task.list", Name: "任务列表", Module: "monitor", Resource: "task", Action: "list", Level: 3, Status: 1},
		{Code: "monitor.task.view", Name: "查看任务", Module: "monitor", Resource: "task", Action: "view", Level: 3, Status: 1},
		{Code: "monitor.task.create", Name: "创建任务", Module: "monitor", Resource: "task", Action: "create", Level: 3, Status: 1},
		{Code: "monitor.task.update", Name: "更新任务", Module: "monitor", Resource: "task", Action: "update", Level: 3, Status: 1},
		{Code: "monitor.task.delete", Name: "删除任务", Module: "monitor", Resource: "task", Action: "delete", Level: 3, Status: 1},
		{Code: "monitor.task.execute", Name: "执行任务", Module: "monitor", Resource: "task", Action: "execute", Level: 3, Status: 1},

		// ========== K8s模块 ==========
		{Code: "k8s", Name: "K8s管理", Module: "k8s", Resource: "", Action: "", Level: 1, Status: 1},

		// 集群管理
		{Code: "k8s.cluster", Name: "集群管理", Module: "k8s", Resource: "cluster", Action: "", Level: 2, Status: 1},
		{Code: "k8s.cluster.list", Name: "集群列表", Module: "k8s", Resource: "cluster", Action: "list", Level: 3, Status: 1},
		{Code: "k8s.cluster.view", Name: "查看集群", Module: "k8s", Resource: "cluster", Action: "view", Level: 3, Status: 1},
		{Code: "k8s.cluster.create", Name: "创建集群", Module: "k8s", Resource: "cluster", Action: "create", Level: 3, Status: 1},
		{Code: "k8s.cluster.update", Name: "更新集群", Module: "k8s", Resource: "cluster", Action: "update", Level: 3, Status: 1},
		{Code: "k8s.cluster.delete", Name: "删除集群", Module: "k8s", Resource: "cluster", Action: "delete", Level: 3, Status: 1},
		{Code: "k8s.cluster.connect", Name: "连接集群", Module: "k8s", Resource: "cluster", Action: "connect", Level: 3, Status: 1},

		// K8s资源管理
		{Code: "k8s.resource", Name: "资源管理", Module: "k8s", Resource: "resource", Action: "", Level: 2, Status: 1},
		{Code: "k8s.resource.view", Name: "查看资源", Module: "k8s", Resource: "resource", Action: "view", Level: 3, Status: 1},
		{Code: "k8s.resource.create", Name: "创建资源", Module: "k8s", Resource: "resource", Action: "create", Level: 3, Status: 1},
		{Code: "k8s.resource.update", Name: "更新资源", Module: "k8s", Resource: "resource", Action: "update", Level: 3, Status: 1},
		{Code: "k8s.resource.delete", Name: "删除资源", Module: "k8s", Resource: "resource", Action: "delete", Level: 3, Status: 1},

		// K8s权限管理
		{Code: "k8s.permission", Name: "K8s权限", Module: "k8s", Resource: "permission", Action: "", Level: 2, Status: 1},
		{Code: "k8s.permission.list", Name: "权限列表", Module: "k8s", Resource: "permission", Action: "list", Level: 3, Status: 1},
		{Code: "k8s.permission.assign", Name: "分配权限", Module: "k8s", Resource: "permission", Action: "assign", Level: 3, Status: 1},
		{Code: "k8s.permission.revoke", Name: "撤销权限", Module: "k8s", Resource: "permission", Action: "revoke", Level: 3, Status: 1},

		// ========== 审计模块 ==========
		{Code: "audit", Name: "审计管理", Module: "audit", Resource: "", Action: "", Level: 1, Status: 1},

		// 登录日志
		{Code: "audit.login_log", Name: "登录日志", Module: "audit", Resource: "login_log", Action: "", Level: 2, Status: 1},
		{Code: "audit.login_log.list", Name: "登录日志列表", Module: "audit", Resource: "login_log", Action: "list", Level: 3, Status: 1},
		{Code: "audit.login_log.view", Name: "查看登录日志", Module: "audit", Resource: "login_log", Action: "view", Level: 3, Status: 1},
		{Code: "audit.login_log.export", Name: "导出登录日志", Module: "audit", Resource: "login_log", Action: "export", Level: 3, Status: 1},

		// 操作日志
		{Code: "audit.operation_log", Name: "操作日志", Module: "audit", Resource: "operation_log", Action: "", Level: 2, Status: 1},
		{Code: "audit.operation_log.list", Name: "操作日志列表", Module: "audit", Resource: "operation_log", Action: "list", Level: 3, Status: 1},
		{Code: "audit.operation_log.view", Name: "查看操作日志", Module: "audit", Resource: "operation_log", Action: "view", Level: 3, Status: 1},
		{Code: "audit.operation_log.export", Name: "导出操作日志", Module: "audit", Resource: "operation_log", Action: "export", Level: 3, Status: 1},

		// 系统事件日志
		{Code: "audit.system_event", Name: "系统事件日志", Module: "audit", Resource: "system_event", Action: "", Level: 2, Status: 1},
		{Code: "audit.system_event.list", Name: "事件日志列表", Module: "audit", Resource: "system_event", Action: "list", Level: 3, Status: 1},
		{Code: "audit.system_event.view", Name: "查看事件日志", Module: "audit", Resource: "system_event", Action: "view", Level: 3, Status: 1},

		// 审计统计
		{Code: "audit.stats", Name: "审计统计", Module: "audit", Resource: "stats", Action: "", Level: 2, Status: 1},
		{Code: "audit.stats.view", Name: "查看统计", Module: "audit", Resource: "stats", Action: "view", Level: 3, Status: 1},
	}
}