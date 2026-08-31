package system

import (
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// seedPermissionRoutes 受保护路由 → 权限码的初始映射（仅初始化，页面配置优先）
// 语义约定：
//   - 常规 CRUD 端点映射到 <module>.<resource>.<list|view|create|update|delete>
//   - 动词端点（deploy/terminate/ack 等）映射到业务动作权限码
//   - 一码多路由是常态（如 k8s.resource.view 保护所有资源查询端点）
//   - 一路由多码并存（OR 语义）：单接口码与集合码同时保护一个端点（如 GET /api/k8s/clusters
//     同时受 k8s.cluster.list 单接口码与 k8s.cluster.view 集合码保护）
//
// 同步规则：按 (method, path, code) 三元组判定，缺失则插入；页面维护过（IsSeed=false）的行不动
func (i *Initializer) syncPermissionRoutes() error {
	logger.Info("开始同步权限路由映射...")

	db := database.GetDB()

	type routeDef struct {
		Method string
		Path   string
		Code   string
	}

	routes := []routeDef{
		// ========== 路由管理 ==========
		{"POST", "/api/route/invalidateCache", "system.route.invalidate"},
		{"GET", "/api/route/debugCache", "system.route.debug"},

		// ========== 工单中心：流程定义管理 ==========
		// 说明：工单发起/审批/详情/评论及 /ticket/types/options 仅需登录（审批权由流程节点配置决定），
		// 不在本表登记；此处仅登记管理端 CRUD 的权限映射。
		{"GET", "/api/ticket/workflows", "ticket.workflow.list"},
		{"GET", "/api/ticket/workflows/:id", "ticket.workflow.list"},
		{"POST", "/api/ticket/workflows", "ticket.workflow.create"},
		{"PUT", "/api/ticket/workflows/:id", "ticket.workflow.update"},
		{"DELETE", "/api/ticket/workflows/:id", "ticket.workflow.delete"},
		// 工单类型管理
		{"GET", "/api/ticket/types", "ticket.type.list"},
		{"POST", "/api/ticket/types", "ticket.type.create"},
		{"PUT", "/api/ticket/types/:id", "ticket.type.update"},
		{"DELETE", "/api/ticket/types/:id", "ticket.type.delete"},
		// 工单治理：改派审批人（管理员能力，解决审批人离职卡死）
		{"POST", "/api/ticket/tickets/:id/reassign", "ticket.ticket.reassign"},
		// 工单通知设置（事件矩阵）
		{"GET", "/api/ticket/notify-policies", "ticket.notify.list"},
		{"PUT", "/api/ticket/notify-policies/:event", "ticket.notify.update"},
		{"GET", "/api/ticket/notify-logs", "ticket.notify.list"},

		// ========== 系统管理：菜单/角色/用户/权限 ==========
		{"GET", "/api/system/menus", "system.menu.list"},
		{"GET", "/api/system/menus/tree", "system.menu.list"},
		{"POST", "/api/system/menus", "system.menu.create"},
		{"PUT", "/api/system/menus/:id", "system.menu.update"},
		{"DELETE", "/api/system/menus/:id", "system.menu.delete"},
		{"GET", "/api/system/roles", "system.role.list"},
		{"GET", "/api/system/roles/options", "system.role.list"},
		{"GET", "/api/system/roles/menu-paths", "system.role.list"},
		{"GET", "/api/system/roles/:id/users", "system.role.list"},
		{"POST", "/api/system/roles", "system.role.create"},
		{"PUT", "/api/system/roles/:id", "system.role.update"},
		{"DELETE", "/api/system/roles/:id", "system.role.delete"},
		{"GET", "/api/system/roles/:id/permissions", "system.role.view"},
		{"POST", "/api/system/roles/:id/permissions", "system.role.assign_permissions"},
		{"GET", "/api/system/users", "system.user.list"},
		{"GET", "/api/system/users/options", "system.user.list"},
		{"POST", "/api/system/users", "system.user.create"},
		{"PUT", "/api/system/users/:id", "system.user.update"},
		{"DELETE", "/api/system/users/:id", "system.user.delete"},
		{"PUT", "/api/system/users/:id/password", "system.user.reset_password"},
		{"GET", "/api/system/permissions", "system.permission.list"},
		{"GET", "/api/system/permissions/tree", "system.permission.list"},
		{"GET", "/api/system/permissions/options", "system.permission.list"},
		{"POST", "/api/system/permissions", "system.permission.create"},
		{"PUT", "/api/system/permissions/:id", "system.permission.update"},
		{"DELETE", "/api/system/permissions/:id", "system.permission.delete"},
		{"POST", "/api/system/permissions/check", "system.permission.list"},
		{"GET", "/api/system/user/permissions", "system.permission.list"},

		// ========== 权限路由映射管理（本表自身的维护入口） ==========
		{"GET", "/api/system/permissions/routes", "system.permission.list"},
		{"POST", "/api/system/permissions/routes", "system.permission.create"},
		{"PUT", "/api/system/permissions/routes/:id", "system.permission.update"},
		{"DELETE", "/api/system/permissions/routes/:id", "system.permission.delete"},

		// ========== 授权中心：应用管理 ==========
		{"GET", "/api/system/applications", "auth.application.list"},
		{"GET", "/api/system/applications/options", "auth.application.list"},
		{"POST", "/api/system/applications", "auth.application.create"},
		{"PUT", "/api/system/applications/:id", "auth.application.update"},
		{"DELETE", "/api/system/applications/:id", "auth.application.delete"},
		{"POST", "/api/system/applications/:id/sync-roles", "auth.application.sync"},
		{"GET", "/api/system/applications/:id/roles", "auth.application.list"},
		{"POST", "/api/system/applications/:id/sync-users", "auth.application.sync"},
		{"GET", "/api/system/applications/:id/users", "auth.application.list"},
		{"POST", "/api/system/applications/:id/sync-groups", "auth.application.sync"},
		{"GET", "/api/system/applications/:id/groups", "auth.application.list"},
		{"POST", "/api/system/applications/:id/sync-rules", "auth.application.sync"},
		{"GET", "/api/system/applications/:id/rules", "auth.application.list"},
		{"GET", "/api/system/applications/:id/operation-logs", "auth.application.list"},
		{"GET", "/api/system/applications/types", "auth.application.list"},
		{"GET", "/api/system/applications/types/:type/config", "auth.application.list"},

		// ========== 授权中心：用户/用户组/绑定 ==========
		{"GET", "/api/system/auth-users", "auth.user.list"},
		{"GET", "/api/system/auth-users/list", "auth.user.list"},
		{"GET", "/api/system/auth-users/:id/password", "auth.user.list"},
		{"POST", "/api/system/auth-users", "auth.user.create"},
		{"PUT", "/api/system/auth-users/:id", "auth.user.update"},
		{"DELETE", "/api/system/auth-users/:id", "auth.user.delete"},
		{"GET", "/api/system/auth-groups", "auth.group.list"},
		{"GET", "/api/system/auth-groups/list", "auth.group.list"},
		{"GET", "/api/system/auth-groups/options", "auth.group.list"},
		{"POST", "/api/system/auth-groups", "auth.group.create"},
		{"PUT", "/api/system/auth-groups/:id", "auth.group.update"},
		{"DELETE", "/api/system/auth-groups/:id", "auth.group.delete"},
		{"GET", "/api/system/groups/:id/bindings", "auth.group.list"},
		{"POST", "/api/system/groups/:id/bindings", "auth.group.create"},
		{"DELETE", "/api/system/groups/bindings/:id", "auth.group.delete"},
		{"GET", "/api/system/users/:id/groups", "auth.user.list"},
		{"POST", "/api/system/users/assign-group", "auth.user.update"},
		{"DELETE", "/api/system/users/:id/groups/:groupId", "auth.user.update"},
		{"GET", "/api/system/user-identity-mappings", "auth.user.list"},
		{"DELETE", "/api/system/user-identity-mappings/:id", "auth.user.delete"},
		{"GET", "/api/system/user-permissions", "auth.application.list"},
		{"GET", "/api/system/user-permissions/matrix", "auth.application.list"},
		{"GET", "/api/system/group-bindings/:bindingId/executions", "auth.group.list"},

		// ========== CMDB：服务器管理 ==========
		{"GET", "/api/cmdb/servers", "cmdb.server.list"},
		{"POST", "/api/cmdb/servers", "cmdb.server.create"},
		{"PUT", "/api/cmdb/servers/:id", "cmdb.server.update"},
		{"DELETE", "/api/cmdb/servers/:id", "cmdb.server.delete"},
		{"GET", "/api/cmdb/servers/stats", "cmdb.server.list"},
		{"GET", "/api/cmdb/servers/options", "cmdb.server.list"},
		{"POST", "/api/cmdb/servers/config", "cmdb.server.view"},
		{"GET", "/api/cmdb/servers/:id", "cmdb.server.view"},
		{"GET", "/api/cmdb/servers/:id/connect", "cmdb.server.connect"},
		{"POST", "/api/cmdb/servers/:id/connect", "cmdb.server.connect"},
		{"GET", "/api/cmdb/servers/:id/permission", "cmdb.server.connect"},
		{"POST", "/api/cmdb/servers/:id/sync-metrics", "cmdb.server.list"},
		{"POST", "/api/cmdb/servers/:id/test-connection", "cmdb.server.create"},

		// ========== CMDB：Agent 管理 ==========
		{"POST", "/api/cmdb/servers/:id/agent/deploy", "cmdb.agents.deploy"},
		{"POST", "/api/cmdb/servers/:id/agent/restart", "cmdb.agents.restart"},
		{"POST", "/api/cmdb/servers/:id/agent/uninstall", "cmdb.agents.uninstall"},
		{"POST", "/api/cmdb/servers/:id/agent/upgrade", "cmdb.agents.upgrade"},
		{"GET", "/api/cmdb/servers/:id/agent/status", "cmdb.agents.list"},
		{"GET", "/api/cmdb/agents", "cmdb.agents.list"},
		{"POST", "/api/cmdb/agents/batch-deploy", "cmdb.agents.deploy"},
		{"POST", "/api/cmdb/agents/batch-uninstall", "cmdb.agents.uninstall"},
		{"DELETE", "/api/cmdb/agents/:id", "cmdb.agents.delete"},
		{"GET", "/api/cmdb/agent-versions", "cmdb.agents.list"},
		{"GET", "/api/cmdb/agent-versions/latest", "cmdb.agents.list"},
		{"GET", "/api/cmdb/agent-versions/:id", "cmdb.agents.list"},
		{"POST", "/api/cmdb/agent-versions", "cmdb.agent_version.manage"},
		{"PUT", "/api/cmdb/agent-versions/:id", "cmdb.agent_version.manage"},
		{"DELETE", "/api/cmdb/agent-versions/:id", "cmdb.agent_version.manage"},
		{"GET", "/api/cmdb/agent-upgrade-tasks", "cmdb.agents.upgrade"},
		{"GET", "/api/cmdb/agent-upgrade-tasks/:id", "cmdb.agents.upgrade"},

		// ========== CMDB：分组/业务/机房/标签 ==========
		{"GET", "/api/cmdb/groups", "cmdb.group.list"},
		{"GET", "/api/cmdb/groups/:id", "cmdb.group.view"},
		{"GET", "/api/cmdb/asset-tree", "cmdb.group.list"},
		{"POST", "/api/cmdb/groups", "cmdb.group.create"},
		{"PUT", "/api/cmdb/groups/:id", "cmdb.group.update"},
		{"DELETE", "/api/cmdb/groups/:id", "cmdb.group.delete"},
		{"POST", "/api/cmdb/groups/assign", "cmdb.group.assign"},
		{"POST", "/api/cmdb/groups/assign-multi", "cmdb.group.assign"},
		{"GET", "/api/cmdb/group-servers/:groupId", "cmdb.group.list"},
		{"GET", "/api/cmdb/business-units", "cmdb.business.list"},
		{"POST", "/api/cmdb/business-units", "cmdb.business.create"},
		{"PUT", "/api/cmdb/business-units/:id", "cmdb.business.update"},
		{"DELETE", "/api/cmdb/business-units/:id", "cmdb.business.delete"},
		{"GET", "/api/cmdb/rooms", "cmdb.rooms.list"},
		{"POST", "/api/cmdb/rooms", "cmdb.rooms.create"},
		{"PUT", "/api/cmdb/rooms/:id", "cmdb.rooms.update"},
		{"DELETE", "/api/cmdb/rooms/:id", "cmdb.rooms.delete"},
		{"GET", "/api/cmdb/cabinets", "cmdb.rooms.list"},
		{"GET", "/api/cmdb/tags", "cmdb.tags.list"},
		{"POST", "/api/cmdb/tags", "cmdb.tags.create"},
		{"PUT", "/api/cmdb/tags/:id", "cmdb.tags.update"},
		{"DELETE", "/api/cmdb/tags/:id", "cmdb.tags.delete"},
		{"POST", "/api/cmdb/tags/assign", "cmdb.tags.list"},
		{"DELETE", "/api/cmdb/server-tags/:serverId/:tagId", "cmdb.tags.list"},

		// ========== CMDB：SSH凭证/资产变更 ==========
		{"GET", "/api/cmdb/ssh-credentials", "cmdb.server.list"},
		{"GET", "/api/cmdb/ssh-credentials/:id", "cmdb.server.list"},
		{"POST", "/api/cmdb/ssh-credentials", "cmdb.credential.create"},
		{"PUT", "/api/cmdb/ssh-credentials/:id", "cmdb.credential.update"},
		{"DELETE", "/api/cmdb/ssh-credentials/:id", "cmdb.credential.delete"},
		{"POST", "/api/cmdb/ssh-credentials/:id/test", "cmdb.credential.test"},
		{"GET", "/api/cmdb/asset-changes", "cmdb.server.list"},

		// ========== CMDB：会话/审计/访问策略 ==========
		{"GET", "/api/cmdb/sessions", "cmdb.session.list"},
		{"GET", "/api/cmdb/sessions/list", "cmdb.session.list"},
		{"GET", "/api/cmdb/sessions/active", "cmdb.session.list"},
		{"GET", "/api/cmdb/sessions/active-memory", "cmdb.session.list"},
		{"GET", "/api/cmdb/sessions/stats", "cmdb.session.list"},
		{"GET", "/api/cmdb/sessions/:id", "cmdb.session.view"},
		{"POST", "/api/cmdb/sessions/:id/terminate", "cmdb.session.terminate"},
		{"GET", "/api/cmdb/sessions/:id/commands", "cmdb.session.view"},
		{"GET", "/api/cmdb/sessions/:id/file-transfers", "cmdb.session.view"},
		{"POST", "/api/cmdb/sessions/:id/resize", "cmdb.session.view"},
		{"GET", "/api/cmdb/commands", "cmdb.session.list"},
		{"GET", "/api/cmdb/file-transfers", "cmdb.session.list"},
		{"GET", "/api/cmdb/access-policies", "cmdb.access_policy.list"},
		{"GET", "/api/cmdb/access-policies/:id", "cmdb.access_policy.list"},
		{"POST", "/api/cmdb/access-policies", "cmdb.access_policy.create"},
		{"PUT", "/api/cmdb/access-policies/:id", "cmdb.access_policy.update"},
		{"DELETE", "/api/cmdb/access-policies/:id", "cmdb.access_policy.delete"},

		// ========== CMDB：属性管理 ==========
		{"GET", "/api/cmdb/attributes", "cmdb.attribute.list"},
		{"POST", "/api/cmdb/attributes", "cmdb.attribute.create"},
		{"PUT", "/api/cmdb/attributes/:id", "cmdb.attribute.update"},
		{"DELETE", "/api/cmdb/attributes/:id", "cmdb.attribute.delete"},
		{"GET", "/api/cmdb/attributes/:id", "cmdb.attribute.view"},
		{"POST", "/api/cmdb/attributes/validate", "cmdb.attribute.view"},
		{"GET", "/api/cmdb/server-attributes/:serverId", "cmdb.attribute.view"},
		{"POST", "/api/cmdb/server-attributes/:serverId", "cmdb.attribute.update"},
		{"GET", "/api/cmdb/servers-by-attributes", "cmdb.attribute.view"},

		// ========== K8s：集群与权限 ==========
		// 集群列表端点双码保护：单接口码 list（只授列表）与集合码 view（列表+详情+nodes+namespaces）
		{"GET", "/api/k8s/clusters", "k8s.cluster.list"},
		{"GET", "/api/k8s/clusters", "k8s.cluster.view"},

		// 用户组管理（系统模块，复用用户管理权限码）
		{"GET", "/api/system/user-groups", "system.user.list"},
		{"GET", "/api/system/user-groups/options", "system.user.list"},
		{"GET", "/api/system/user-groups/:id/members", "system.user.list"},
		{"POST", "/api/system/user-groups", "system.user.create"},
		{"PUT", "/api/system/user-groups/:id", "system.user.update"},
		{"DELETE", "/api/system/user-groups/:id", "system.user.delete"},
		{"POST", "/api/system/user-groups/:id/members", "system.user.update"},
		{"DELETE", "/api/system/user-groups/:id/members/:userId", "system.user.update"},

		// 原生 RBAC 绑定与对象代管（集群授权唯一入口：K8s管理-安全管理-授权管理）
		{"GET", "/api/k8s/native-bindings", "k8s.permission.list"},
		{"GET", "/api/k8s/clusters/options", "k8s.permission.list"},
		{"GET", "/api/k8s/clusters/:id/native-bindings", "k8s.permission.list"},
		{"GET", "/api/k8s/clusters/:id/permission/subject-options", "k8s.permission.list"},
		{"GET", "/api/k8s/clusters/:id/permission/role-options", "k8s.permission.list"},
		{"POST", "/api/k8s/clusters/:id/native-bindings", "k8s.permission.assign"},
		{"DELETE", "/api/k8s/clusters/:id/native-bindings/:bindingId", "k8s.permission.revoke"},
		{"GET", "/api/k8s/clusters/:id/rbac/clusterroles", "k8s.rbac.view"},
		{"GET", "/api/k8s/clusters/:id/rbac/clusterroles/:name", "k8s.rbac.view"},
		{"PUT", "/api/k8s/clusters/:id/rbac/clusterroles", "k8s.rbac.manage"},
		{"GET", "/api/k8s/clusters/:id/rbac/roles", "k8s.rbac.view"},
		{"GET", "/api/k8s/clusters/:id/rbac/roles/:namespace/:name", "k8s.rbac.view"},
		{"PUT", "/api/k8s/clusters/:id/rbac/roles/:namespace", "k8s.rbac.manage"},
		{"GET", "/api/k8s/clusters/:id/rbac/clusterrolebindings", "k8s.rbac.view"},
		{"GET", "/api/k8s/clusters/:id/rbac/clusterrolebindings/:name", "k8s.rbac.view"},
		{"GET", "/api/k8s/clusters/:id/rbac/rolebindings", "k8s.rbac.view"},
		{"GET", "/api/k8s/clusters/:id/rbac/rolebindings/:namespace/:name", "k8s.rbac.view"},
		{"POST", "/api/k8s/clusters", "k8s.cluster.create"},
		{"GET", "/api/k8s/clusters/:id", "k8s.cluster.view"},
		{"PUT", "/api/k8s/clusters/:id", "k8s.cluster.update"},
		{"DELETE", "/api/k8s/clusters/:id", "k8s.cluster.delete"},
		{"POST", "/api/k8s/clusters/:id/test", "k8s.cluster.connect"},
		{"GET", "/api/k8s/clusters/:id/nodes", "k8s.cluster.view"},
		{"GET", "/api/k8s/clusters/:id/namespaces", "k8s.cluster.view"},
		{"GET", "/api/k8s/users/clusters", "k8s.permission.list"},

		// ========== K8s：资源管理（一码多路由） ==========
		{"GET", "/api/k8s/clusters/:id/deployments", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/deployments/:namespace/:name", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/deployments/:namespace/:name/pods", "k8s.resource.view"},
		{"POST", "/api/k8s/clusters/:id/deployments", "k8s.resource.create"},
		{"PUT", "/api/k8s/clusters/:id/deployments", "k8s.resource.update"},
		{"DELETE", "/api/k8s/clusters/:id/deployments", "k8s.resource.delete"},
		{"POST", "/api/k8s/clusters/:id/deployments/scale", "k8s.resource.update"},
		{"POST", "/api/k8s/clusters/:id/deployments/restart", "k8s.resource.update"},
		{"GET", "/api/k8s/clusters/:id/statefulsets", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/statefulsets/:namespace/:name", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/statefulsets/:namespace/:name/pods", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/daemonsets", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/daemonsets/:namespace/:name", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/daemonsets/:namespace/:name/pods", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/jobs", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/jobs/:namespace/:name", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/jobs/:namespace/:name/pods", "k8s.resource.view"},
		{"DELETE", "/api/k8s/clusters/:id/jobs", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/cronjobs", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/cronjobs/:namespace/:name", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/cronjobs/:namespace/:name/pods", "k8s.resource.view"},
		{"DELETE", "/api/k8s/clusters/:id/cronjobs", "k8s.resource.delete"},
		{"PUT", "/api/k8s/clusters/:id/cronjobs/suspend", "k8s.resource.update"},
		{"GET", "/api/k8s/clusters/:id/services", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/services/:namespace/:name", "k8s.resource.view"},
		{"POST", "/api/k8s/clusters/:id/services", "k8s.resource.create"},
		{"PUT", "/api/k8s/clusters/:id/services", "k8s.resource.update"},
		{"DELETE", "/api/k8s/clusters/:id/services", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/ingresses", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/ingresses/:namespace/:name", "k8s.resource.view"},
		{"POST", "/api/k8s/clusters/:id/ingresses", "k8s.resource.create"},
		{"PUT", "/api/k8s/clusters/:id/ingresses", "k8s.resource.update"},
		{"DELETE", "/api/k8s/clusters/:id/ingresses", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/pods", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/pods/:namespace/:name", "k8s.resource.view"},
		{"PUT", "/api/k8s/clusters/:id/pods", "k8s.resource.update"},
		{"GET", "/api/k8s/clusters/:id/pods/:namespace/:name/logs", "k8s.resource.view"},
		{"DELETE", "/api/k8s/clusters/:id/pods", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/configmaps", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/configmaps/:namespace/:name", "k8s.resource.view"},
		{"POST", "/api/k8s/clusters/:id/configmaps", "k8s.resource.create"},
		{"PUT", "/api/k8s/clusters/:id/configmaps", "k8s.resource.update"},
		{"DELETE", "/api/k8s/clusters/:id/configmaps", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/secrets", "k8s.resource.view"},
		{"GET", "/api/k8s/clusters/:id/secrets/:namespace/:name", "k8s.resource.view"},
		{"POST", "/api/k8s/clusters/:id/secrets", "k8s.resource.create"},
		{"PUT", "/api/k8s/clusters/:id/secrets", "k8s.resource.update"},
		{"DELETE", "/api/k8s/clusters/:id/secrets", "k8s.resource.delete"},
		{"GET", "/api/k8s/clusters/:id/events", "k8s.resource.view"},

		// ========== K8s：终端/诊断 ==========
		{"GET", "/api/k8s/terminal/active", "k8s.terminal.connect"},
		{"POST", "/api/k8s/terminal/sessions/:sessionId/terminate", "k8s.terminal.connect"},
		{"GET", "/api/k8s/diagnostic/commands", "k8s.diagnostic.view"},
		{"GET", "/api/k8s/diagnostic/pods/:clusterId/:namespace", "k8s.diagnostic.view"},
		{"GET", "/api/k8s/diagnostic/namespaces/:clusterId", "k8s.diagnostic.view"},
		{"POST", "/api/k8s/diagnostic/execute", "k8s.diagnostic.execute"},
		{"GET", "/api/k8s/diagnostic/history", "k8s.diagnostic.view"},

		// ========== 监控中心 ==========
		{"GET", "/api/monitoring/overview", "monitor.data.view"},
		{"GET", "/api/monitoring/alerts", "monitor.alert.list"},
		{"POST", "/api/monitoring/alerts/:id/acknowledge", "monitor.alert.ack"},
		{"GET", "/api/monitoring/alerts/stats", "monitor.alert.list"},
		{"GET", "/api/monitoring/alerts/rules", "monitor.alert.list"},
		{"POST", "/api/monitoring/alerts/rules", "monitor.alert_rule.create"},
		{"PUT", "/api/monitoring/alerts/rules/:id", "monitor.alert_rule.update"},
		{"DELETE", "/api/monitoring/alerts/rules/:id", "monitor.alert_rule.delete"},
		{"PUT", "/api/monitoring/alerts/rules/:id/status", "monitor.alert_rule.update"},
		{"GET", "/api/monitoring/notifications/channels", "monitor.alert.list"},
		{"POST", "/api/monitoring/notifications/channels", "monitor.notification.create"},
		{"PUT", "/api/monitoring/notifications/channels/:id", "monitor.notification.update"},
		{"DELETE", "/api/monitoring/notifications/channels/:id", "monitor.notification.delete"},
		{"POST", "/api/monitoring/notifications/channels/:id/test", "monitor.notification.test"},
		{"GET", "/api/monitoring/reports", "monitor.report.list"},
		{"POST", "/api/monitoring/reports", "monitor.report.create"},
		{"GET", "/api/monitoring/reports/:id", "monitor.report.list"},
		{"GET", "/api/monitoring/reports/:id/export", "monitor.report.list"},
		{"DELETE", "/api/monitoring/reports/:id", "monitor.report.delete"},

		// ========== 审计中心 ==========
		{"GET", "/api/audit/login-logs", "audit.login_log.list"},
		{"GET", "/api/audit/login-logs/export", "audit.login_log.export"},
		{"GET", "/api/audit/operation-logs", "audit.operation_log.list"},
		{"GET", "/api/audit/operation-logs/export", "audit.operation_log.export"},
		{"GET", "/api/audit/system-event-logs", "audit.system_event.list"},
		{"GET", "/api/audit/stats", "audit.stats.view"},
		{"GET", "/api/audit/modules", "audit.stats.view"},
	}

	// 校验映射的权限码都存在于权限目录（防拼写错导致校验恒 false）
	var permCount int64
	db.Model(&modelsystem.Permission{}).Count(&permCount)
	if permCount > 0 {
		codes := make(map[string]bool, len(routes))
		for _, r := range routes {
			codes[r.Code] = true
		}
		var perms []modelsystem.Permission
		codeList := make([]string, 0, len(codes))
		for code := range codes {
			codeList = append(codeList, code)
		}
		if err := db.Where("code IN ?", codeList).Find(&perms).Error; err == nil {
			existing := make(map[string]bool, len(perms))
			for _, p := range perms {
				existing[p.Code] = true
			}
			for code := range codes {
				if !existing[code] {
					logger.Warn("权限路由映射引用了不存在的权限码", zap.String("code", code))
				}
			}
		}
	}

	// 老库迁移：多码并存（OR 语义）需要去掉旧版的 (method, path) 唯一索引
	// GORM AutoMigrate 只加不删，这里显式清理
	if db.Migrator().HasIndex(&modelsystem.PermissionRoute{}, "uk_permission_route") {
		if err := db.Migrator().DropIndex(&modelsystem.PermissionRoute{}, "uk_permission_route"); err != nil {
			logger.Warn("清理旧唯一索引失败", zap.Error(err))
		} else {
			logger.Info("已移除旧唯一索引 uk_permission_route（端点可配置多个权限码）")
		}
	}

	inserted, updated := 0, 0
	for _, r := range routes {
		// 三元组 (method, path, code) 判定：同一端点允许不同权限码并存
		var existing modelsystem.PermissionRoute
		err := db.Where("method = ? AND path = ? AND permission_code = ?", r.Method, r.Path, r.Code).First(&existing).Error
		if err != nil {
			if err := db.Create(&modelsystem.PermissionRoute{
				PermissionCode: r.Code,
				Method:         r.Method,
				Path:           r.Path,
				IsSeed:         true,
			}).Error; err != nil {
				logger.Error("创建权限路由映射失败", zap.String("route", r.Method+" "+r.Path), zap.Error(err))
				continue
			}
			inserted++
		} else if !existing.IsSeed {
			// 页面维护过的记录（IsSeed=false）不由种子覆盖，仅标记跳过
			updated++
		}
	}

	logger.Info("权限路由映射同步完成",
		zap.Int("total", len(routes)),
		zap.Int("inserted", inserted),
		zap.Int("page_managed", updated))

	// 清理孤儿 seed 映射：代码中已移除/改码的旧 (method,path,code) 若仍以 is_seed=1 留库，
	// 在 OR 语义下会继续放行（写接口从读码改为专码后，残留旧行等于后门），必须同步删除。
	// 只清 is_seed=1 的行，页面手工维护（IsSeed=false）的映射不受影响
	var seedRoutes []modelsystem.PermissionRoute
	if err := db.Where("is_seed = ?", true).Find(&seedRoutes).Error; err != nil {
		logger.Warn("查询 seed 权限映射失败，跳过孤儿清理", zap.Error(err))
		return nil
	}
	current := make(map[string]bool, len(routes))
	for _, r := range routes {
		current[r.Method+"|"+r.Path+"|"+r.Code] = true
	}
	removed := 0
	for _, sr := range seedRoutes {
		key := sr.Method + "|" + sr.Path + "|" + sr.PermissionCode
		if !current[key] {
			if err := db.Delete(&sr).Error; err != nil {
				logger.Warn("清理孤儿 seed 权限映射失败", zap.String("route", key), zap.Error(err))
			} else {
				removed++
			}
		}
	}
	if removed > 0 {
		logger.Info("已清理孤儿 seed 权限映射", zap.Int("removed", removed))
	}

	return nil
}
