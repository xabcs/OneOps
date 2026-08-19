package modelsystem

import "time"

// PermissionRoute 权限-路由映射模型
// 运行时权限校验的唯一数据来源：RequirePermissionFromDB 按 (method, path) 匹配得到权限码集合
// 同一端点可配置多个权限码（OR 语义，用户持有任一即可访问）：
//   - 单接口权限码：只保护一个端点（如 k8s.cluster.list → GET /api/k8s/clusters）
//   - 集合权限码：保护一组端点（如 k8s.cluster.view → 列表+详情+nodes+namespaces）
// 一个权限码也可对应多条路由（如 k8s.resource.view 保护十几个资源查询端点）
type PermissionRoute struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	PermissionCode string    `json:"permissionCode" gorm:"type:varchar(100);not null;index:idx_permission_route"`
	Method         string    `json:"method" gorm:"type:varchar(10);not null;index:idx_permission_route"`
	Path           string    `json:"path" gorm:"type:varchar(255);not null;index:idx_permission_route"`
	// IsSeed 是否由启动种子管理：true 时随代码修正更新；页面新建/修改后置 false，此后种子不再覆盖（页面优先）
	IsSeed    bool      `json:"isSeed" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (PermissionRoute) TableName() string { return "sys_permission_routes" }

// --- DTO ---

type CreatePermissionRouteRequest struct {
	PermissionCode string `json:"permissionCode" binding:"required,max=100"`
	Method         string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
	Path           string `json:"path" binding:"required,max=255"`
}

type UpdatePermissionRouteRequest struct {
	PermissionCode string `json:"permissionCode" binding:"required,max=100"`
}
