package system

import (
	"errors"
	"fmt"
	"strings"

	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// GetPermissionCodesByRoute 运行时权限校验入口：按 (method, path) 匹配得到权限码集合
// 同一端点可配置多个权限码（OR 语义）：单接口码与集合码并存，用户持有任一即可访问
// 数据来源 sys_permission_routes（seed 初始化 + 权限管理页维护）。
// 全量内存缓存 + 本实例写操作整体失效：单实例内即改即生效；多实例部署需改为共享缓存或加 TTL
func (s *PermissionService) GetPermissionCodesByRoute(method, path string) ([]string, error) {
	key := method + " " + path

	s.routeCacheMu.RLock()
	codes, ok := s.routeCache[key]
	s.routeCacheMu.RUnlock()
	if ok {
		return codes, nil
	}

	var routes []modelsystem.PermissionRoute
	if err := s.db.Where("method = ? AND path = ?", method, path).Find(&routes).Error; err != nil {
		return nil, err
	}
	codes = make([]string, 0, len(routes))
	for _, r := range routes {
		codes = append(codes, r.PermissionCode)
	}

	s.routeCacheMu.Lock()
	s.routeCache[key] = codes
	s.routeCacheMu.Unlock()
	return codes, nil
}

// invalidateRouteCache 映射增删改后整体失效缓存（懒加载，重建成本 = 每路由 1 条查询）
func (s *PermissionService) invalidateRouteCache() {
	s.routeCacheMu.Lock()
	s.routeCache = make(map[string][]string)
	s.routeCacheMu.Unlock()
}

// GetPermissionRoutes 查询权限路由映射，可按权限码过滤
func (s *PermissionService) GetPermissionRoutes(permissionCode string) ([]modelsystem.PermissionRoute, error) {
	var routes []modelsystem.PermissionRoute
	query := s.db.Order("permission_code, path")
	if permissionCode != "" {
		query = query.Where("permission_code = ?", permissionCode)
	}
	if err := query.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

// CreatePermissionRoute 新增权限路由映射（页面维护入口，IsSeed=false，种子不再覆盖）
// 同一端点允许不同权限码并存；仅拒绝完全相同的 (method, path, code) 重复行
func (s *PermissionService) CreatePermissionRoute(req modelsystem.CreatePermissionRouteRequest) error {
	code := strings.TrimSpace(req.PermissionCode)
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	path := strings.TrimSpace(req.Path)

	if !strings.HasPrefix(path, "/api/") {
		return fmt.Errorf("路径必须以 /api/ 开头")
	}

	// 权限码必须存在于目录且启用
	var count int64
	if err := s.db.Model(&modelsystem.Permission{}).
		Where("code = ? AND status = 1", code).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("权限码不存在或已停用: %s", code)
	}

	// 完全相同的三元组才视为重复
	var existing int64
	if err := s.db.Model(&modelsystem.PermissionRoute{}).
		Where("method = ? AND path = ? AND permission_code = ?", method, path, code).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return fmt.Errorf("该映射已存在: %s %s → %s", method, path, code)
	}

	if err := s.db.Create(&modelsystem.PermissionRoute{
		PermissionCode: code,
		Method:         method,
		Path:           path,
	}).Error; err != nil {
		return err
	}
	s.invalidateRouteCache()
	return nil
}

// UpdatePermissionRoute 修改映射归属的权限码（端点 method/path 不可改，删了重建）
func (s *PermissionService) UpdatePermissionRoute(id uint, req modelsystem.UpdatePermissionRouteRequest) error {
	code := strings.TrimSpace(req.PermissionCode)

	var count int64
	if err := s.db.Model(&modelsystem.Permission{}).
		Where("code = ? AND status = 1", code).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("权限码不存在或已停用: %s", code)
	}

	var route modelsystem.PermissionRoute
	if err := s.db.First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("映射不存在")
		}
		return err
	}

	// 改码后若与既有行完全重复则拒绝
	var dup int64
	if err := s.db.Model(&modelsystem.PermissionRoute{}).
		Where("method = ? AND path = ? AND permission_code = ? AND id != ?", route.Method, route.Path, code, id).
		Count(&dup).Error; err != nil {
		return err
	}
	if dup > 0 {
		return fmt.Errorf("该端点已存在此权限码映射: %s", code)
	}

	// 页面修改过的映射脱离种子管理，后续启动不再被 seed 覆盖
	return s.db.Model(&route).Updates(map[string]interface{}{
		"permission_code": code,
		"is_seed":         false,
	}).Error
}

// DeletePermissionRoute 删除权限路由映射
// 该端点若还有其他权限码行则继续受保护；一行不剩则变为拒绝访问（fail-closed）
func (s *PermissionService) DeletePermissionRoute(id uint) error {
	if err := s.db.Delete(&modelsystem.PermissionRoute{}, id).Error; err != nil {
		return err
	}
	s.invalidateRouteCache()
	return nil
}
