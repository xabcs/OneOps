package system

import (
	"sort"
	"strings"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// publicRoutes 无需权限映射的路由：
// 登录引导/路由守卫接口（仅认证）、WebSocket 与 Agent 上报（自行验证凭证）
var publicRoutes = map[string]bool{
	"POST:/api/login":                  true,
	"POST:/api/logout":                 true,
	"GET:/api/user/info":               true,
	"GET:/api/route/getConstantRoutes": true,
	"GET:/api/route/getUserRoutes":     true,
	"GET:/api/route/isRouteExist":      true,
	"POST:/api/cmdb/agent/heartbeat":   true,
	"GET:/api/cmdb/sessions/:id/ws":    true,
	"GET:/api/k8s/terminal/ws":         true,
	"GET:/api/monitoring/ws":           true,
	// 工单站内消息：仅认证（本人收件箱数据，无角色维度）
	"GET:/api/ticket/messages":              true,
	"GET:/api/ticket/messages/unread-count": true,
	"PUT:/api/ticket/messages/read":         true,
	"GET:/api/ticket/my-notify-setting":     true, // 用户通知偏好（本人自助）
	"PUT:/api/ticket/my-notify-setting":     true,
}

// AuditRoutePermissions 启动对账（只读，不写库）：
//  1. 受保护路由缺少权限映射 → Error 日志点名（这些路由会被中间件拒绝访问）
//  2. 映射指向未注册的路由 → Warn 日志（代码已删除路由，映射成为死数据）
//
// 权限映射本身由 seed（syncPermissionRoutes）与权限管理页共同维护，本函数仅提供可观测性
func AuditRoutePermissions(engine *gin.Engine) error {
	db := database.GetDB()

	var mappings []modelsystem.PermissionRoute
	if err := db.Find(&mappings).Error; err != nil {
		return err
	}
	mapped := make(map[string][]string, len(mappings)) // "METHOD:path" → permissionCodes（OR 并联）
	for _, m := range mappings {
		mapped[m.Method+":"+m.Path] = append(mapped[m.Method+":"+m.Path], m.PermissionCode)
	}

	unprotected := []string{}
	deadMappings := make(map[string]string)

	registered := make(map[string]bool)
	for _, r := range engine.Routes() {
		if !strings.HasPrefix(r.Path, "/api/") {
			continue // Swagger、静态资源等非业务路由
		}
		key := r.Method + ":" + r.Path
		registered[key] = true

		if publicRoutes[key] {
			continue
		}
		if _, ok := mapped[key]; !ok {
			unprotected = append(unprotected, key)
		}
	}

	// 反向对账：映射存在但路由未注册
	for key, codes := range mapped {
		if !registered[key] {
			deadMappings[key] = strings.Join(codes, " | ")
		}
	}

	sort.Strings(unprotected)
	for _, key := range unprotected {
		logger.Error("[AuditRoutePermissions] 受保护路由缺少权限映射，访问将被拒绝（请在权限管理页或 seed 中补映射）",
			zap.String("route", key))
	}
	for key, code := range deadMappings {
		logger.Warn("[AuditRoutePermissions] 权限映射指向未注册的路由（死数据，可清理）",
			zap.String("route", key),
			zap.String("permission_code", code))
	}

	logger.Info("[AuditRoutePermissions] 路由权限对账完成",
		zap.Int("registered_api_routes", len(registered)),
		zap.Int("mapped_routes", len(mapped)),
		zap.Int("unprotected", len(unprotected)),
		zap.Int("dead_mappings", len(deadMappings)))

	return nil
}
