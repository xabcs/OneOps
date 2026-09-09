package system

import (
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// initMenus 初始化菜单数据（动态路由模式）
func (i *Initializer) initMenus() error {
	db := database.GetDB()

	menus := []modelsystem.Menu{
		// 一级菜单
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "系统管理", Icon: "carbon:cloud-service-management", Path: "/manage", Permission: "", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "用户管理", Icon: "ic:round-manage-accounts", Path: "/manage/user", Permission: "system.user.view", Sort: 1, Status: 1, ParentID: 2},
		{ID: 4, Name: "角色管理", Icon: "carbon:user-role", Path: "/manage/role", Permission: "system.role.view", Sort: 2, Status: 1, ParentID: 2},
		{ID: 5, Name: "菜单管理", Icon: "material-symbols:route", Path: "/manage/menu", Permission: "system.menu.view", Sort: 3, Status: 1, ParentID: 2},
		{ID: 13, Name: "关于", Icon: "fluent:book-information-24-regular", Path: "/about", Permission: "", Sort: 5, Status: 1, ParentID: 0},
		{ID: 14, Name: "用户中心", Icon: "mdi:user-circle-outline", Path: "/user-center", Permission: "", Sort: 6, Status: 1, ParentID: 0},
		{ID: 20, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", Sort: 3, Status: 1, ParentID: 0},
		{ID: 21, Name: "主机管理", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", Sort: 1, Status: 1, ParentID: 20},
		{ID: 22, Name: "业务管理", Icon: "mdi:sitemap", Path: "/cmdb/business", Permission: "cmdb:business:query", Sort: 2, Status: 1, ParentID: 20},
		{ID: 23, Name: "机房管理", Icon: "mdi:office-building-marker", Path: "/cmdb/rooms", Permission: "cmdb:room:query", Sort: 3, Status: 1, ParentID: 20},
		{ID: 24, Name: "标签管理", Icon: "mdi:tag-multiple", Path: "/cmdb/tags", Permission: "cmdb:tag:query", Sort: 4, Status: 1, ParentID: 20},
		{ID: 25, Name: "变更记录", Icon: "mdi:history", Path: "/cmdb/changes", Permission: "cmdb:change:query", Sort: 5, Status: 1, ParentID: 20},
		{ID: 26, Name: "SSH凭证", Icon: "mdi:key-variant", Path: "/cmdb/ssh-credentials", Permission: "cmdb:credential:query", Sort: 6, Status: 1, ParentID: 20},
	}

	for _, menu := range menus {
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
	}

	return nil
}

// syncMenus 同步菜单数据（自动检测并添加新菜单）
func (i *Initializer) syncMenus() error {
	logger.Info("开始同步菜单数据...")

	db := database.GetDB()

	// 定义动态路由菜单（用于 SoybeanAdmin 动态路由模式）
	// V4 - 匹配新的前端模块化目录结构
	menus := []modelsystem.Menu{
		// ========== 一级菜单 ==========
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", MenuType: "menu", Sort: 1, Status: 1, ParentID: 0},

		{ID: 2, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", MenuType: "directory", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "监控中心", Icon: "mdi:chart-line", Path: "/monitoring", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 0},
		{ID: 7, Name: "授权中心", Icon: "mdi:shield-account", Path: "/auth", Permission: "", MenuType: "directory", Sort: 4, Status: 1, ParentID: 0},
		{ID: 4, Name: "审计中心", Icon: "mdi:file-document", Path: "/audit", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 0},
		{ID: 5, Name: "K8s管理", Icon: "mdi:kubernetes", Path: "/k8s", Permission: "", MenuType: "directory", Sort: 6, Status: 1, ParentID: 0},
		{ID: 6, Name: "系统管理", Icon: "mdi:cog", Path: "/manage", Permission: "", MenuType: "directory", Sort: 8, Status: 1, ParentID: 0},

		// ========== CMDB 二级菜单 (ID: 20-39) ==========
		{ID: 20, Name: "主机管理", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 2},
		{ID: 21, Name: "业务管理", Icon: "mdi:sitemap", Path: "/cmdb/business", Permission: "cmdb:business:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 2},
		// 凭证管理目录
		{ID: 22, Name: "凭证管理", Icon: "mdi:key", Path: "/cmdb/credentials", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 2},
		{ID: 23, Name: "访问凭证", Icon: "mdi:key-variant", Path: "/cmdb/credentials/access", Permission: "cmdb:credentials:access", MenuType: "menu", Sort: 1, Status: 1, ParentID: 22},
		{ID: 24, Name: "SSH密钥", Icon: "mdi:ssh", Path: "/cmdb/credentials/ssh", Permission: "cmdb:credentials:ssh", MenuType: "menu", Sort: 2, Status: 1, ParentID: 22},
		// 访问策略
		{ID: 25, Name: "访问策略", Icon: "mdi:shield-lock", Path: "/cmdb/policies", Permission: "cmdb:policies:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 2},
		// 配置管理目录
		{ID: 26, Name: "配置管理", Icon: "mdi:cog", Path: "/cmdb/config", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 2},
		{ID: 27, Name: "业务配置", Icon: "mdi:sitemap", Path: "/cmdb/config/business", Permission: "cmdb:config:business", MenuType: "menu", Sort: 1, Status: 1, ParentID: 26},
		{ID: 28, Name: "机房管理", Icon: "mdi:server", Path: "/cmdb/config/rooms", Permission: "cmdb:rooms:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 26},
		{ID: 29, Name: "标签管理", Icon: "mdi:tag-multiple", Path: "/cmdb/config/tags", Permission: "cmdb:tags:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 26},
		{ID: 30, Name: "代理配置", Icon: "mdi:robot", Path: "/cmdb/config/agents", Permission: "cmdb:agents:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 26},
		{ID: 39, Name: "属性管理", Icon: "mdi:format-list-bulleted-type", Path: "/cmdb/config/attributes", Permission: "cmdb.attribute.view", MenuType: "menu", Sort: 5, Status: 1, ParentID: 26},
		// 资产总览
		{ID: 31, Name: "资产总览", Icon: "mdi:chart-pie", Path: "/cmdb/dashboard", Permission: "cmdb:dashboard:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 2},
		// 审计记录目录
		{ID: 32, Name: "审计记录", Icon: "mdi:history", Path: "/cmdb/audit", Permission: "", MenuType: "directory", Sort: 7, Status: 1, ParentID: 2},
		{ID: 33, Name: "变更记录", Icon: "mdi:file-document", Path: "/cmdb/audit/changes", Permission: "cmdb:audit:changes", MenuType: "menu", Sort: 1, Status: 1, ParentID: 32},
		// 命令审计目录
		{ID: 34, Name: "命令审计", Icon: "mdi:terminal", Path: "/cmdb/audit/command", Permission: "", MenuType: "directory", Sort: 2, Status: 1, ParentID: 32},
		{ID: 35, Name: "命令历史", Icon: "mdi:history", Path: "/cmdb/audit/command/history", Permission: "cmdb:audit:command:history", MenuType: "menu", Sort: 1, Status: 1, ParentID: 34},
		{ID: 36, Name: "在线会话", Icon: "mdi:laptop", Path: "/cmdb/audit/online", Permission: "cmdb:audit:online", MenuType: "menu", Sort: 3, Status: 1, ParentID: 32},
		{ID: 37, Name: "历史会话", Icon: "mdi:history", Path: "/cmdb/audit/sessions", Permission: "cmdb:audit:sessions", MenuType: "menu", Sort: 4, Status: 1, ParentID: 32},
		{ID: 38, Name: "命令记录", Icon: "mdi:code-tags", Path: "/cmdb/audit/commands", Permission: "cmdb:audit:commands", MenuType: "menu", Sort: 5, Status: 1, ParentID: 32},
		// web 终端：新标签页打开的工作台（open_type=new-tab），可见性跟随主机管理（resource=server）
		{ID: 60, Name: "Web终端", Icon: "mdi:console", Path: "/webterminal", Permission: "", MenuType: "menu", OpenType: "new-tab", Sort: 8, Status: 1, ParentID: 2},

		// ========== 监控中心二级菜单 (ID: 40-49) ==========
		{ID: 40, Name: "监控概览", Icon: "mdi:chart-line", Path: "/monitoring/overview", Permission: "monitoring:overview:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 3},
		{ID: 41, Name: "主机监控", Icon: "mdi:server-network", Path: "/monitoring/servers", Permission: "monitoring:servers:query", MenuType: "directory", Sort: 2, Status: 1, ParentID: 3},
		{ID: 42, Name: "告警管理", Icon: "mdi:alert-circle", Path: "/monitoring/alerts", Permission: "monitoring:alerts:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 3},
		{ID: 43, Name: "趋势分析", Icon: "mdi:chart-areaspline", Path: "/monitoring/trends", Permission: "monitoring:trends:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 3},
		{ID: 44, Name: "巡检报告", Icon: "mdi:file-document", Path: "/monitoring/reports", Permission: "monitoring:reports:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 3},
		{ID: 45, Name: "监控设置", Icon: "mdi:cog", Path: "/monitoring/settings", Permission: "monitoring:settings:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 3},

		// ========== 审计中心二级菜单 (ID: 50-59) ==========
		{ID: 50, Name: "登录审计", Icon: "mdi:login", Path: "/audit/login", Permission: "audit.login.view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 4},
		{ID: 51, Name: "操作审计", Icon: "mdi:account-edit", Path: "/audit/operation", Permission: "audit.operation.view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 4},
		{ID: 52, Name: "系统事件", Icon: "mdi:information", Path: "/audit/system", Permission: "audit.system.view", MenuType: "menu", Sort: 3, Status: 1, ParentID: 4},

		// ========== 授权中心二级菜单 (ID: 100-107) ==========
		{ID: 100, Name: "用户", Icon: "mdi:account", Path: "/auth/users", Permission: "auth:user:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 7},
		{ID: 101, Name: "用户组", Icon: "mdi:shield-account", Path: "/auth/roles", Permission: "auth:role:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 7},
		{ID: 102, Name: "应用", Icon: "mdi:application", Path: "/auth/applications", Permission: "auth:app:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 7},
		{ID: 103, Name: "权限映射", Icon: "mdi:link", Path: "/auth/rolebindings", Permission: "auth:binding:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 7},
		{ID: 104, Name: "用户授权", Icon: "mdi:account-key", Path: "/auth/userauthorization", Permission: "auth:authorization:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 7},
		{ID: 105, Name: "操作日志", Icon: "mdi:file-document", Path: "/auth/operationlogs", Permission: "auth:log:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 7},
		{ID: 106, Name: "用户身份映射", Icon: "mdi:account-switch", Path: "/auth/user-identities", Permission: "auth:identity:query", MenuType: "menu", Sort: 7, Status: 1, ParentID: 7},
		{ID: 107, Name: "用户有效权限", Icon: "mdi:shield-check", Path: "/auth/user-permissions", Permission: "auth:permission:query", MenuType: "menu", Sort: 8, Status: 1, ParentID: 7},

		// ========== K8s管理二级菜单 (ID: 80-99) ==========
		{ID: 80, Name: "集群管理", Icon: "mdi:server-network", Path: "/k8s/clusters", Permission: "k8s:cluster:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 5},
		{ID: 87, Name: "工作负载", Icon: "mdi:cube-outline", Path: "/k8s/workloads", Permission: "k8s:workload:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 5},
		{ID: 90, Name: "诊断中心", Icon: "mdi:stethoscope", Path: "/k8s/diagnostic", Permission: "k8s:diagnostic:execute", MenuType: "menu", Sort: 3, Status: 1, ParentID: 5},
		{ID: 88, Name: "网络", Icon: "mdi:network-outline", Path: "/k8s/network", Permission: "k8s:network:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 5},
		{ID: 89, Name: "配置管理", Icon: "mdi:cog", Path: "/k8s/config", Permission: "k8s:config:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 5},
		// 安全管理目录：集群访问授权的统一入口（与系统管理的平台角色明确分层——此处管理的是集群内 RBAC 权限）
		{ID: 93, Name: "安全管理", Icon: "mdi:shield-lock", Path: "/k8s/security", Permission: "", MenuType: "directory", Sort: 6, Status: 1, ParentID: 5},
		{ID: 94, Name: "授权管理", Icon: "mdi:account-key", Path: "/k8s/security/authorization", Permission: "k8s.permission.list", MenuType: "menu", Sort: 1, Status: 1, ParentID: 93},
		{ID: 92, Name: "角色管理", Icon: "mdi:shield-key", Path: "/k8s/security/roles", Permission: "k8s.rbac.view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 93},

		// ========== 系统管理二级菜单 (ID: 70-79) ==========
		{ID: 70, Name: "用户管理", Icon: "mdi:account-multiple", Path: "/manage/user", Permission: "system.user.view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 6},
		{ID: 71, Name: "角色管理", Icon: "mdi:shield-account", Path: "/manage/role", Permission: "system.role.view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 6},
		{ID: 72, Name: "菜单管理", Icon: "mdi:menu", Path: "/manage/menu", Permission: "system.menu.view", MenuType: "menu", Sort: 3, Status: 1, ParentID: 6},
		{ID: 73, Name: "权限管理", Icon: "mdi:shield-key", Path: "/manage/permission", Permission: "system.permission.list", MenuType: "menu", Sort: 4, Status: 1, ParentID: 6},
		{ID: 74, Name: "用户组", Icon: "mdi:account-group", Path: "/manage/user-group", Permission: "system.user.view", MenuType: "menu", Sort: 5, Status: 1, ParentID: 6},
		// 通知渠道：平台级渠道池（邮件/企微/钉钉机器人），供监控告警与工单通知共同引用
		{ID: 75, Name: "通知渠道", Icon: "mdi:bell-badge", Path: "/manage/notification-channels", Permission: "monitor.notification.update", MenuType: "menu", Sort: 6, Status: 1, ParentID: 6},

		// ========== 工单中心 (ID: 110-119) ==========
		{ID: 110, Name: "工单中心", Icon: "mdi:clipboard-text-clock", Path: "/ticket", Permission: "ticket.ticket.view", MenuType: "directory", Sort: 7, Status: 1, ParentID: 0},
		{ID: 111, Name: "我的工单", Icon: "mdi:ticket-confirmation", Path: "/ticket/center", Permission: "ticket.ticket.view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 110},
		{ID: 112, Name: "工单类型", Icon: "mdi:format-list-checks", Path: "/ticket/types", Permission: "ticket.type.list", MenuType: "menu", Sort: 2, Status: 1, ParentID: 110},
		{ID: 113, Name: "审批流程", Icon: "mdi:source-branch", Path: "/ticket/workflows", Permission: "ticket.workflow.list", MenuType: "menu", Sort: 3, Status: 1, ParentID: 110},
		// 通知设置：事件矩阵（审批事件 × 渠道绑定），渠道本身在系统管理-通知渠道维护
		{ID: 114, Name: "通知设置", Icon: "mdi:bell-cog", Path: "/ticket/notify-settings", Permission: "ticket.notify.list", MenuType: "menu", Sort: 4, Status: 1, ParentID: 110},
	}

	addedCount := 0
	updatedCount := 0

	for _, menu := range menus {
		var existingMenu modelsystem.Menu
		err := db.Where("id = ?", menu.ID).First(&existingMenu).Error

		if err == nil {
			// 菜单已存在，更新数据（保持数据同步）
			// 注意：不更新 sort 字段，保留用户在菜单管理中修改的排序
			db.Model(&existingMenu).Updates(map[string]interface{}{
				"name":       menu.Name,
				"icon":       menu.Icon,
				"path":       menu.Path,
				"permission": menu.Permission,
				"resource":   menu.Resource,
				"parent_id":  menu.ParentID,
				"status":     menu.Status,
				"menu_type":  menu.MenuType,
				"open_type":  menu.OpenType,
			})
			updatedCount++
			logger.Debug("更新菜单",
				zap.String("name", menu.Name),
				zap.Uint("id", menu.ID))
		} else {
			// 菜单不存在，添加新菜单
			if err := db.Create(&menu).Error; err != nil {
				logger.Error("添加菜单失败",
					zap.String("name", menu.Name),
					zap.Any("error", err))
				return err
			}
			addedCount++
			logger.Info("添加新菜单",
				zap.String("name", menu.Name),
				zap.Uint("id", menu.ID),
				zap.String("path", menu.Path))
		}
	}
	// 更新菜单的 resource 字段（根据路径映射到权限码的 resource）
	logger.Info("开始更新菜单resource字段...")

	resourceMappings := []struct {
		path     string
		resource string
	}{
		{"/manage/user", "user"},
		{"/manage/role", "role"},
		{"/manage/menu", "menu"},
		{"/manage/permission", "permission"},
		{"/cmdb/servers", "server"},
		// web 终端与主机管理共享 resource：有主机管理权限即可见终端入口
		{"/webterminal", "server"},
		{"/cmdb/business", "business"},
		{"/cmdb/config/rooms", "rooms"},
		{"/cmdb/config/tags", "tags"},
		{"/cmdb/config/agents", "agents"},
		{"/cmdb/config/attributes", "attribute"},
		{"/cmdb/config/business", "config_business"},
		{"/k8s/clusters", "cluster"},
		{"/k8s/workloads", "resource"},
		{"/k8s/network", "resource"},
		{"/k8s/config", "resource"},
		{"/k8s/diagnostic", "diagnostic"},
		{"/k8s/security/authorization", "permission"},
		{"/k8s/security/roles", "rbac"},
		{"/monitoring/overview", "overview"},
		{"/monitoring/servers", "monitoring_servers"},
		{"/monitoring/alerts", "alerts"},
		{"/monitoring/trends", "trends"},
		{"/monitoring/reports", "reports"},
		{"/monitoring/settings", "settings"},
		{"/audit/login", "login_audit"},
		{"/audit/operation", "operation_audit"},
		{"/audit/system", "system_audit"},
		// 工单中心
		{"/ticket/center", "ticket"},
		{"/ticket/types", "type"},
		{"/ticket/workflows", "workflow"},
		{"/ticket/notify-settings", "notify"},
		// 系统管理
		{"/manage/notification-channels", "notification"},
	}

	for _, mapping := range resourceMappings {
		result := db.Model(&modelsystem.Menu{}).Where("path = ?", mapping.path).Update("resource", mapping.resource)
		if result.Error != nil {
			logger.Warn("更新菜单resource字段失败",
				zap.String("path", mapping.path),
				zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			logger.Info("更新菜单resource字段",
				zap.String("path", mapping.path),
				zap.String("resource", mapping.resource))
		}
	}

	logger.Info("菜单resource字段更新完成")

	// 删除废弃的 K8s 会话审计菜单（如果存在）
	deletedK8sAuditResult := db.Where("path = ?", "/k8s/audit/sessions").Delete(&modelsystem.Menu{})
	if deletedK8sAuditResult.Error != nil {
		logger.Error("删除废弃K8s会话审计菜单失败", zap.Error(deletedK8sAuditResult.Error))
		return deletedK8sAuditResult.Error
	}
	if deletedK8sAuditResult.RowsAffected > 0 {
		logger.Info("已删除废弃K8s会话审计菜单", zap.Int64("count", deletedK8sAuditResult.RowsAffected))
	}

	deletedResult := db.Where("path = ? OR name IN ?", "/cmdb/groups", []string{"主机分组", "cmdb_groups"}).Delete(&modelsystem.Menu{})
	if deletedResult.Error != nil {
		logger.Error("删除废弃主机分组菜单失败", zap.Error(deletedResult.Error))
		return deletedResult.Error
	}
	if deletedResult.RowsAffected > 0 {
		logger.Info("已删除废弃主机分组菜单", zap.Int64("count", deletedResult.RowsAffected))
	}

	// 删除废弃的 terminal 菜单（已改名为 webterminal）
	deletedTerminalResult := db.Where("path IN ?", []string{"/terminal", "/terminal/workbench"}).Delete(&modelsystem.Menu{})
	if deletedTerminalResult.Error != nil {
		logger.Error("删除废弃终端菜单失败", zap.Error(deletedTerminalResult.Error))
		return deletedTerminalResult.Error
	}
	if deletedTerminalResult.RowsAffected > 0 {
		logger.Info("已删除废弃终端菜单", zap.Int64("count", deletedTerminalResult.RowsAffected))
	}

	// 删除废弃的集群角色菜单：三档（cluster-viewer/operator/admin）已下线，
	// 集群授权统一走集群详情-原生授权（native-bindings）
	deletedClusterRolesResult := db.Where("path = ?", "/k8s/cluster-roles").Delete(&modelsystem.Menu{})
	if deletedClusterRolesResult.Error != nil {
		logger.Error("删除废弃集群角色菜单失败", zap.Error(deletedClusterRolesResult.Error))
		return deletedClusterRolesResult.Error
	}
	if deletedClusterRolesResult.RowsAffected > 0 {
		logger.Info("已删除废弃集群角色菜单", zap.Int64("count", deletedClusterRolesResult.RowsAffected))
	}

	logger.Info("菜单同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(menus)))

	return nil
}

// SyncMenus 公开的菜单同步方法（用于外部调用）
func (i *Initializer) SyncMenus() error {
	return i.syncMenus()
}
