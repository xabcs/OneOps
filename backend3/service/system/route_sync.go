package system

import (
	"fmt"
	"strings"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// routePermOverride 存储无法通过约定自动推导的路由 → 权限码映射
// key 格式: "METHOD:/api/path"
var routePermOverride = map[string]string{
	// ========== 路由模块（特殊路径前缀） ==========
	"GET:/api/route/getUserRoutes":    "system.route.list",
	"POST:/api/route/invalidateCache": "system.route.invalidate",
	"GET:/api/route/debugCache":       "system.route.debug",
	"GET:/api/route/isRouteExist":     "system.route.list",

	// ========== 子资源操作 ==========
	// 角色权限分配
	"GET:/api/system/roles/:roleId/permissions":  "system.role.view",
	"POST:/api/system/roles/:roleId/permissions": "system.role.assign_permissions",
	// 用户密码重置
	"PUT:/api/system/users/:id/password": "system.user.reset_password",
	// 权限检查
	"POST:/api/system/permissions/check":  "system.permission.list",
	"GET:/api/system/permissions/tree":    "system.permission.list",
	"GET:/api/system/permissions/options": "system.permission.list",
	"GET:/api/system/user/permissions":    "system.permission.list",
	// 菜单树
	"GET:/api/system/menus/tree":    "system.menu.list",
	"GET:/api/system/users/options": "system.user.list",

	// CMDB Agent 操作
	"POST:/api/cmdb/servers/:id/agent/deploy":    "cmdb.agents.deploy",
	"POST:/api/cmdb/servers/:id/agent/restart":   "cmdb.agents.restart",
	"POST:/api/cmdb/servers/:id/agent/uninstall": "cmdb.agents.uninstall",
	"POST:/api/cmdb/servers/:id/agent/upgrade":   "cmdb.agents.upgrade",
	"GET:/api/cmdb/servers/:id/agent/status":     "cmdb.agents.list",
	// CMDB 服务器连接
	"POST:/api/cmdb/servers/:id/connect":   "cmdb.server.connect",
	"GET:/api/cmdb/servers/:id/connect":    "cmdb.server.connect",
	"GET:/api/cmdb/servers/:id/permission": "cmdb.server.connect",
	// CMDB 服务器子接口
	"GET:/api/cmdb/servers/stats":             "cmdb.server.list",
	"GET:/api/cmdb/servers/options":           "cmdb.server.list",
	"POST:/api/cmdb/servers/config":           "cmdb.server.view",
	"POST:/api/cmdb/servers/:id/sync-metrics": "cmdb.server.list",
	"GET:/api/cmdb/servers/:id":               "cmdb.server.view",
	// CMDB Agent 管理
	"GET:/api/cmdb/agents":                  "cmdb.agents.list",
	"POST:/api/cmdb/agents/batch-deploy":    "cmdb.agents.deploy",
	"POST:/api/cmdb/agents/batch-uninstall": "cmdb.agents.uninstall",
	"DELETE:/api/cmdb/agents/:id":           "cmdb.agents.list",
	// CMDB Agent 版本
	"GET:/api/cmdb/agent-versions":        "cmdb.agents.list",
	"GET:/api/cmdb/agent-versions/latest": "cmdb.agents.list",
	"GET:/api/cmdb/agent-versions/:id":    "cmdb.agents.list",
	"POST:/api/cmdb/agent-versions":       "cmdb.agents.list",
	"PUT:/api/cmdb/agent-versions/:id":    "cmdb.agents.list",
	"DELETE:/api/cmdb/agent-versions/:id": "cmdb.agents.list",
	// CMDB Agent 升级任务
	"GET:/api/cmdb/agent-upgrade-tasks":     "cmdb.agents.upgrade",
	"GET:/api/cmdb/agent-upgrade-tasks/:id": "cmdb.agents.upgrade",
	// CMDB 分组
	"POST:/api/cmdb/groups/assign":         "cmdb.group.assign",
	"POST:/api/cmdb/groups/assign-multi":   "cmdb.group.assign",
	"GET:/api/cmdb/group-servers/:groupId": "cmdb.group.list",
	"GET:/api/cmdb/asset-tree":             "cmdb.group.list",
	// CMDB 机房
	"GET:/api/cmdb/cabinets": "cmdb.rooms.list",
	// CMDB 标签
	"POST:/api/cmdb/tags/assign":                    "cmdb.tags.list",
	"DELETE:/api/cmdb/server-tags/:serverId/:tagId": "cmdb.tags.list",
	// CMDB SSH 凭证
	"GET:/api/cmdb/ssh-credentials":           "cmdb.server.list",
	"GET:/api/cmdb/ssh-credentials/:id":       "cmdb.server.list",
	"POST:/api/cmdb/ssh-credentials":          "cmdb.server.list",
	"PUT:/api/cmdb/ssh-credentials/:id":       "cmdb.server.list",
	"DELETE:/api/cmdb/ssh-credentials/:id":    "cmdb.server.list",
	"POST:/api/cmdb/ssh-credentials/:id/test": "cmdb.server.list",
	// CMDB 资产变更
	"GET:/api/cmdb/asset-changes": "cmdb.server.list",
	// CMDB 会话
	"GET:/api/cmdb/sessions":                    "cmdb.session.list",
	"GET:/api/cmdb/sessions/list":               "cmdb.session.list",
	"GET:/api/cmdb/sessions/active":             "cmdb.session.list",
	"GET:/api/cmdb/sessions/active-memory":      "cmdb.session.list",
	"GET:/api/cmdb/sessions/stats":              "cmdb.session.list",
	"GET:/api/cmdb/sessions/:id":                "cmdb.session.view",
	"POST:/api/cmdb/sessions/:id/terminate":     "cmdb.session.terminate",
	"GET:/api/cmdb/sessions/:id/commands":       "cmdb.session.view",
	"GET:/api/cmdb/sessions/:id/file-transfers": "cmdb.session.view",
	"POST:/api/cmdb/sessions/:id/resize":        "cmdb.session.view",
	// CMDB 命令审计
	"GET:/api/cmdb/commands":       "cmdb.session.list",
	"GET:/api/cmdb/file-transfers": "cmdb.session.list",
	// CMDB 访问策略
	"GET:/api/cmdb/access-policies":        "cmdb.access_policy.list",
	"GET:/api/cmdb/access-policies/:id":    "cmdb.access_policy.list",
	"POST:/api/cmdb/access-policies":       "cmdb.access_policy.create",
	"PUT:/api/cmdb/access-policies/:id":    "cmdb.access_policy.update",
	"DELETE:/api/cmdb/access-policies/:id": "cmdb.access_policy.delete",
	// CMDB 主机属性
	"GET:/api/system/server-attributes/:serverId":  "system.attribute.view",
	"POST:/api/system/server-attributes/:serverId": "system.attribute.update",

	// K8s 集群操作
	"POST:/api/k8s/clusters/:id/test":              "k8s.cluster.connect",
	"GET:/api/k8s/clusters/:id/nodes":              "k8s.cluster.view",
	"GET:/api/k8s/clusters/:id/namespaces":         "k8s.cluster.view",
	"GET:/api/k8s/clusters/:id/users":              "k8s.permission.list",
	"GET:/api/k8s/clusters/:id/users/:userId/role": "k8s.permission.list",
	"POST:/api/k8s/permissions/batch-assign":       "k8s.permission.assign",
	// K8s 资源操作（部署、服务等统一映射到 resource 权限）
	"GET:/api/k8s/clusters/:id/deployments":          "k8s.resource.view",
	"POST:/api/k8s/clusters/:id/deployments":         "k8s.resource.create",
	"PUT:/api/k8s/clusters/:id/deployments":          "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/deployments":       "k8s.resource.delete",
	"POST:/api/k8s/clusters/:id/deployments/scale":   "k8s.resource.update",
	"POST:/api/k8s/clusters/:id/deployments/restart": "k8s.resource.update",
	"GET:/api/k8s/clusters/:id/statefulsets":         "k8s.resource.view",
	"GET:/api/k8s/clusters/:id/daemonsets":           "k8s.resource.view",
	"GET:/api/k8s/clusters/:id/jobs":                 "k8s.resource.view",
	"DELETE:/api/k8s/clusters/:id/jobs":              "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/cronjobs":             "k8s.resource.view",
	"DELETE:/api/k8s/clusters/:id/cronjobs":          "k8s.resource.delete",
	"PUT:/api/k8s/clusters/:id/cronjobs/suspend":     "k8s.resource.update",
	"GET:/api/k8s/clusters/:id/services":             "k8s.resource.view",
	"POST:/api/k8s/clusters/:id/services":            "k8s.resource.create",
	"PUT:/api/k8s/clusters/:id/services":             "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/services":          "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/ingresses":            "k8s.resource.view",
	"POST:/api/k8s/clusters/:id/ingresses":           "k8s.resource.create",
	"PUT:/api/k8s/clusters/:id/ingresses":            "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/ingresses":         "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/pods":                 "k8s.resource.view",
	"PUT:/api/k8s/clusters/:id/pods":                 "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/pods":              "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/configmaps":           "k8s.resource.view",
	"POST:/api/k8s/clusters/:id/configmaps":          "k8s.resource.create",
	"PUT:/api/k8s/clusters/:id/configmaps":           "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/configmaps":        "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/secrets":              "k8s.resource.view",
	"POST:/api/k8s/clusters/:id/secrets":             "k8s.resource.create",
	"PUT:/api/k8s/clusters/:id/secrets":              "k8s.resource.update",
	"DELETE:/api/k8s/clusters/:id/secrets":           "k8s.resource.delete",
	"GET:/api/k8s/clusters/:id/events":               "k8s.resource.view",
	// K8s 终端
	"GET:/api/k8s/terminal/active":                         "k8s.cluster.view",
	"POST:/api/k8s/terminal/sessions/:sessionId/terminate": "k8s.cluster.view",
	// K8s 诊断
	"GET:/api/k8s/diagnostic/commands":                   "k8s.diagnostic.view",
	"GET:/api/k8s/diagnostic/namespaces/:clusterId":      "k8s.diagnostic.view",
	"GET:/api/k8s/diagnostic/pods/:clusterId/:namespace": "k8s.diagnostic.view",
	// K8s 权限
	"GET:/api/k8s/users/clusters":                      "k8s.permission.list",
	"POST:/api/k8s/clusters/:id/permissions":           "k8s.permission.assign",
	"DELETE:/api/k8s/clusters/:id/permissions/:userId": "k8s.permission.revoke",

	// 监控模块
	"GET:/api/monitoring/alerts/stats":                     "monitor.alert.list",
	"GET:/api/monitoring/alerts/rules":                     "monitor.alert.list",
	"POST:/api/monitoring/alerts/rules":                    "monitor.alert.list",
	"PUT:/api/monitoring/alerts/rules/:id":                 "monitor.alert.list",
	"DELETE:/api/monitoring/alerts/rules/:id":              "monitor.alert.list",
	"PUT:/api/monitoring/alerts/rules/:id/status":          "monitor.alert.list",
	"GET:/api/monitoring/notifications/channels":           "monitor.alert.list",
	"POST:/api/monitoring/notifications/channels":          "monitor.alert.list",
	"PUT:/api/monitoring/notifications/channels/:id":       "monitor.alert.list",
	"DELETE:/api/monitoring/notifications/channels/:id":    "monitor.alert.list",
	"POST:/api/monitoring/notifications/channels/:id/test": "monitor.alert.list",
	"GET:/api/monitoring/reports/:id":                      "monitor.report.list",
	"GET:/api/monitoring/reports/:id/export":               "monitor.report.list",
	"DELETE:/api/monitoring/reports/:id":                   "monitor.report.list",

	// 审计模块
	"GET:/api/audit/modules": "audit.stats.view",

	// 授权中心
	"GET:/api/system/applications/options":                 "auth.application.list",
	"GET:/api/system/applications/types":                   "auth.application.list",
	"GET:/api/system/applications/types/:type/config":      "auth.application.list",
	"POST:/api/system/applications/:id/sync-roles":         "auth.application.sync",
	"GET:/api/system/applications/:id/roles":               "auth.application.list",
	"POST:/api/system/applications/:id/sync-users":         "auth.application.sync",
	"GET:/api/system/applications/:id/users":               "auth.application.list",
	"POST:/api/system/applications/:id/sync-groups":        "auth.application.sync",
	"GET:/api/system/applications/:id/groups":              "auth.application.list",
	"POST:/api/system/applications/:id/sync-rules":         "auth.application.sync",
	"GET:/api/system/applications/:id/rules":               "auth.application.list",
	"GET:/api/system/applications/:id/operation-logs":      "auth.application.list",
	"GET:/api/system/auth-users/list":                      "auth.user.list",
	"GET:/api/system/auth-users/:id/password":              "auth.user.list",
	"GET:/api/system/auth-users":                           "auth.user.list",
	"POST:/api/system/auth-users":                          "auth.user.create",
	"PUT:/api/system/auth-users/:id":                       "auth.user.update",
	"DELETE:/api/system/auth-users/:id":                    "auth.user.delete",
	"GET:/api/system/auth-groups":                          "auth.group.list",
	"GET:/api/system/auth-groups/list":                     "auth.group.list",
	"GET:/api/system/auth-groups/options":                  "auth.group.list",
	"POST:/api/system/auth-groups":                         "auth.group.create",
	"PUT:/api/system/auth-groups/:id":                      "auth.group.update",
	"DELETE:/api/system/auth-groups/:id":                   "auth.group.delete",
	"GET:/api/system/groups/:id/bindings":                  "auth.group.list",
	"POST:/api/system/groups/:id/bindings":                 "auth.group.create",
	"DELETE:/api/system/groups/bindings/:id":               "auth.group.delete",
	"GET:/api/system/users/:id/groups":                     "auth.user.list",
	"POST:/api/system/users/assign-group":                  "auth.user.update",
	"DELETE:/api/system/users/:id/groups/:groupId":         "auth.user.update",
	"GET:/api/system/user-identity-mappings":               "auth.user.list",
	"DELETE:/api/system/user-identity-mappings/:id":        "auth.user.delete",
	"GET:/api/system/user-permissions":                     "auth.application.list",
	"GET:/api/system/user-permissions/matrix":              "auth.application.list",
	"GET:/api/system/group-bindings/:bindingId/executions": "auth.group.list",
}

// moduleMap 路径模块 → 权限模块映射
var moduleMap = map[string]string{
	"system":     "system",
	"cmdb":       "cmdb",
	"k8s":        "k8s",
	"monitoring": "monitor",
	"audit":      "audit",
	"route":      "system",
}

// singularize 简单的英文复数 → 单数转换
func singularize(word string) string {
	// 特殊: ies → y (categories → category)
	if strings.HasSuffix(word, "ies") && len(word) > 3 {
		return strings.TrimSuffix(word, "ies") + "y"
	}
	// 特殊: 以 ses/xes/ches/shes/zes 结尾的去掉 es (boxes → box, buses → bus)
	if len(word) > 3 {
		for _, suffix := range []string{"ses", "xes", "ches", "shes", "zes"} {
			if strings.HasSuffix(word, suffix) {
				return strings.TrimSuffix(word, "es")
			}
		}
	}
	// 普通复数: 去掉末尾 s (roles → role, users → user)
	if strings.HasSuffix(word, "s") && len(word) > 1 {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// derivePermissionCode 根据约定从 method + path 推导权限码
func derivePermissionCode(method, path string) (string, bool) {
	// 去掉 /api/ 前缀
	trimmed := strings.TrimPrefix(path, "/api/")
	segments := strings.Split(trimmed, "/")
	if len(segments) < 2 {
		return "", false
	}

	module, ok := moduleMap[segments[0]]
	if !ok {
		return "", false
	}

	resource := singularize(segments[1])

	// 判断是否有 :id 参数
	hasID := false
	if len(segments) > 2 {
		hasID = strings.HasPrefix(segments[2], ":")
	}

	switch method {
	case "GET":
		if hasID {
			return fmt.Sprintf("%s.%s.view", module, resource), true
		}
		return fmt.Sprintf("%s.%s.list", module, resource), true
	case "POST":
		return fmt.Sprintf("%s.%s.create", module, resource), true
	case "PUT":
		return fmt.Sprintf("%s.%s.update", module, resource), true
	case "DELETE":
		return fmt.Sprintf("%s.%s.delete", module, resource), true
	}
	return "", false
}

// SyncRoutePermissions 从 Gin 引擎提取已注册路由，自动匹配权限码并更新 sys_permissions 表
// 在 routes.SetupRoutes(r) 之后调用
func SyncRoutePermissions(engine *gin.Engine) error {
	svc, err := GetPermissionService()
	if err != nil {
		return fmt.Errorf("获取 PermissionService 失败: %w", err)
	}

	// 获取所有已注册路由
	routes := engine.Routes()

	// 收集 route → permissionCode 映射
	// 一个权限码可以对应多个路由，但 sys_permissions 只存一个主路由
	// 所以：优先保留 CRUD 主操作路由，跳过子路由
	type routeEntry struct {
		Method string
		Path   string
	}
	codeToRoute := make(map[string]routeEntry) // permissionCode → best route
	routeToCode := make(map[string]string)     // "METHOD:path" → permissionCode

	matched, unmatched := 0, 0

	for _, r := range routes {
		// 跳过 Swagger、静态文件等非业务路由
		if !strings.HasPrefix(r.Path, "/api/") {
			continue
		}
		// 跳过无需认证的路由（login, logout, heartbeat, ws 等）
		if shouldSkipRoute(r.Method, r.Path) {
			continue
		}

		routeKey := r.Method + ":" + r.Path

		// 1. 优先检查 override 表
		permCode, found := routePermOverride[routeKey]

		// 2. 尝试约定推导
		if !found {
			permCode, found = derivePermissionCode(r.Method, r.Path)
		}

		if !found {
			unmatched++
			logger.Debug("[SyncRoutePermissions] 路由未匹配权限",
				zap.String("method", r.Method),
				zap.String("path", r.Path))
			continue
		}

		routeToCode[routeKey] = permCode
		matched++

		// 为每个权限码选择最佳路由（优先 CRUD 主路由模式）
		// 主路由模式优先级: GET /resource > POST /resource > PUT /resource/:id > DELETE /resource/:id
		if existing, exists := codeToRoute[permCode]; exists {
			if isBetterRoute(r.Method, r.Path, existing.Method, existing.Path) {
				codeToRoute[permCode] = routeEntry{Method: r.Method, Path: r.Path}
			}
		} else {
			codeToRoute[permCode] = routeEntry{Method: r.Method, Path: r.Path}
		}
	}

	// 按权限码更新 route_method / route_path
	// 只更新 route 为空的权限，不覆盖管理页面手动配置的值
	updated := 0
	for code, route := range codeToRoute {
		result := svc.db.Model(&modelsystem.Permission{}).
			Where("code = ? AND (route_method = '' OR route_method IS NULL)", code).
			Updates(map[string]interface{}{
				"route_method": route.Method,
				"route_path":   route.Path,
			})
		if result.RowsAffected > 0 {
			updated++
		}
	}

	logger.Info("[SyncRoutePermissions] 路由权限同步完成",
		zap.Int("total_routes", len(routes)),
		zap.Int("matched", matched),
		zap.Int("unmatched", unmatched),
		zap.Int("permissions_updated", updated))

	return nil
}

// shouldSkipRoute 判断路由是否无需权限校验（无需认证的公共接口）
func shouldSkipRoute(method, path string) bool {
	skipPaths := map[string]bool{
		"POST:/api/login":                  true,
		"GET:/api/route/getConstantRoutes": true,
		"GET:/api/user/info":               true,
		"POST:/api/logout":                 true,
		"POST:/api/cmdb/agent/heartbeat":   true,
		"GET:/api/k8s/terminal/ws":         true,
		"GET:/api/monitoring/ws":           true,
	}
	return skipPaths[method+":"+path]
}

// isBetterRoute 判断新路由是否比已有路由更适合作为权限的主路由
// 优先级: 集合操作(list/create) > 单条操作(view/update/delete)
func isBetterRoute(newMethod, newPath, oldMethod, oldPath string) bool {
	// GET /resource (list) 是最高优先级
	if newMethod == "GET" && !strings.Contains(newPath, ":id") {
		return true
	}
	// POST /resource (create) 次之
	if newMethod == "POST" && !strings.Contains(newPath, ":id") {
		return oldMethod != "GET" || strings.Contains(oldPath, ":id")
	}
	return false
}
