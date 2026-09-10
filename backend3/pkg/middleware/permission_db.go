package middleware

import (
	"strings"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequirePermissionFromDB 权限检查中间件（数据库驱动，无缓存）
//
// 解析顺序：sys_permission_routes 按 (method, path) 匹配权限码集合 → 用户持有任一码即通过（OR 语义）
// 同一端点可同时配置单接口码与集合码（如 k8s.cluster.list 与 k8s.cluster.view 保护同一列表端点）
// 映射缺失时拒绝访问（fail-closed）：受保护组内的路由必须有映射，启动对账
// （AuditRoutePermissions）会列出全部缺失项，新增路由需在权限管理页或 seed 中补映射
func RequirePermissionFromDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, utils.ErrorUnauthorized("未授权访问"))
			c.Abort()
			return
		}

		// 获取当前路由信息
		method := c.Request.Method // GET/POST/PUT/DELETE
		path := c.FullPath()       // /api/audit/login-logs

		// 获取权限服务
		permService, err := system.GetPermissionService()
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限服务初始化失败"))
			c.Abort()
			return
		}

		// 通过 service 方法查询路由所需权限码集合（OR 语义）
		codes, err := permService.GetPermissionCodesByRoute(method, path)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限映射查询失败"))
			c.Abort()
			return
		}
		if len(codes) == 0 {
			// fail-closed：无映射即拒绝，避免新路由在配置前静默放行
			logger.Error("路由缺少权限映射，已拒绝访问",
				zap.String("route", method+" "+path),
				zap.Any("user_id", userID))
			c.JSON(403, utils.ErrorForbidden("接口未配置权限映射: "+path))
			c.Abort()
			return
		}

		// 检查用户是否持有任一权限码（通过 Casbin）；
		// 优先复用认证中间件预取的角色（gin 上下文），未取到时服务内回退查库
		var roles []*modelsystem.Role
		if v, exists := c.Get("user_roles"); exists {
			roles, _ = v.([]*modelsystem.Role)
		}
		hasPermission, err := permService.HasAnyPermissionWithRoles(userID.(uint), roles, codes)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败"))
			c.Abort()
			return
		}

		if !hasPermission {
			// 记录权限不足的详细信息
			username, _ := c.Get("username")
			logger.Warn("权限不足",
				zap.Any("user_id", userID),
				zap.Any("username", username),
				zap.String("route", method+" "+path),
				zap.String("required_permissions", strings.Join(codes, " | ")))

			c.JSON(403, utils.ErrorForbidden("权限不足: 需要 "+strings.Join(codes, " 或 ")+" 权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}
