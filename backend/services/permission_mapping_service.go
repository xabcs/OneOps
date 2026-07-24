package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/logger"
	"oneops/backend/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionMappingService 权限映射服务
// 负责管理授权中心用户组与外部应用权限的映射关系
type PermissionMappingService struct {
	db              *gorm.DB
	adapterFactory *AdapterFactory
}

// NewPermissionMappingService 创建权限映射服务实例
func NewPermissionMappingService() *PermissionMappingService {
	return &PermissionMappingService{
		db:              db,
		adapterFactory: NewAdapterFactory(),
	}
}

// AssignPermissionToAuthGroup 为授权中心用户组分配外部应用权限
// 核心方法：实现权限映射的核心逻辑
func (s *PermissionMappingService) AssignPermissionToAuthGroup(authGroupID, appID uint, mappingType, externalID, externalName string, permissionDetail map[string]interface{}, grantedBy string) (*models.PermissionMappingResult, error) {
	// 1. 验证授权中心用户组是否存在
	var authGroup models.AuthGroup
	if err := s.db.First(&authGroup, authGroupID).Error; err != nil {
		return &models.PermissionMappingResult{
			Success: false,
			Message: fmt.Sprintf("授权中心用户组不存在 (ID: %d)", authGroupID),
		}, err
	}

	// 2. 验证外部应用是否存在
	var app models.Application
	if err := s.db.First(&app, appID).Error; err != nil {
		return &models.PermissionMappingResult{
			Success: false,
			Message: fmt.Sprintf("外部应用不存在 (ID: %d)", appID),
		}, err
	}

	// 3. 检查是否已存在相同的映射
	var existingMapping models.AuthGroupPermissionMapping
	err := s.db.Where("auth_group_id = ? AND app_id = ? AND mapping_type = ? AND external_id = ?",
		authGroupID, appID, mappingType, externalID).First(&existingMapping).Error

	if err == nil {
		// 映射已存在，更新权限详情
		permissionDetailJSON, _ := json.Marshal(permissionDetail)
		existingMapping.PermissionDetail = string(permissionDetailJSON)
		existingMapping.ExternalName = externalName
		existingMapping.IsEnabled = true
		existingMapping.GrantedBy = grantedBy
		existingMapping.GrantedAt = time.Now()

		if err := s.db.Save(&existingMapping).Error; err != nil {
			return &models.PermissionMappingResult{
				Success: false,
				Message: "更新权限映射失败",
			}, err
		}

		// 调用外部应用API同步权限
		if syncErr := s.syncPermissionToExternalApp(&existingMapping, &app); syncErr != nil {
			logger.Warn("同步权限到外部应用失败",
				zap.Uint("mappingId", existingMapping.ID),
				zap.Error(syncErr))
			return &models.PermissionMappingResult{
				Success:  true,
				Message:  "权限映射已更新，但外部应用同步失败",
				MappingID: existingMapping.ID,
				ExternalID: externalID,
				Warnings:  []string{syncErr.Error()},
			}, nil
		}

		return &models.PermissionMappingResult{
			Success:   true,
			Message:   "权限映射已更新",
			MappingID: existingMapping.ID,
			ExternalID: externalID,
		}, nil
	}

	// 4. 创建新的权限映射
	permissionDetailJSON, _ := json.Marshal(permissionDetail)
	mapping := &models.AuthGroupPermissionMapping{
		AuthGroupID:      authGroupID,
		AppID:            appID,
		MappingType:      mappingType,
		ExternalID:       externalID,
		ExternalName:     externalName,
		PermissionDetail: string(permissionDetailJSON),
		IsEnabled:        true,
		GrantedBy:        grantedBy,
		GrantedAt:        time.Now(),
		SyncStatus:       "pending",
	}

	if err := s.db.Create(mapping).Error; err != nil {
		return &models.PermissionMappingResult{
			Success: false,
			Message: "创建权限映射失败",
		}, err
	}

	// 5. 调用外部应用API同步权限
	if syncErr := s.syncPermissionToExternalApp(mapping, &app); syncErr != nil {
		logger.Warn("同步权限到外部应用失败",
			zap.Uint("mappingId", mapping.ID),
			zap.Error(syncErr))
		// 仍然返回成功，因为映射已创建，但标记同步状态为失败
		s.db.Model(mapping).Updates(map[string]interface{}{
			"sync_status":          "failed",
			"sync_error_message":   syncErr.Error(),
		})
		return &models.PermissionMappingResult{
			Success:   true,
			Message:   "权限映射已创建，但外部应用同步失败",
			MappingID: mapping.ID,
			ExternalID: externalID,
			Warnings:  []string{syncErr.Error()},
		}, nil
	}

	// 更新同步状态为成功
	now := time.Now()
	s.db.Model(mapping).Updates(map[string]interface{}{
		"sync_status":   "success",
		"last_synced_at": &now,
	})

	logger.Info("成功为授权中心用户组分配外部应用权限",
		zap.Uint("authGroupId", authGroupID),
		zap.String("authGroupName", authGroup.Name),
		zap.Uint("appId", appID),
		zap.String("appName", app.Name),
		zap.String("mappingType", mappingType),
		zap.String("externalId", externalID),
		zap.String("grantedBy", grantedBy))

	return &models.PermissionMappingResult{
		Success:   true,
		Message:   "权限分配成功",
		MappingID: mapping.ID,
		ExternalID: externalID,
	}, nil
}

// syncPermissionToExternalApp 调用外部应用API同步权限
func (s *PermissionMappingService) syncPermissionToExternalApp(mapping *models.AuthGroupPermissionMapping, app *models.Application) error {
	// 获取对应类型的适配器
	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		return fmt.Errorf("获取应用适配器失败: %w", err)
	}

	// 检查适配器是否支持权限操作
	type PermissionOperable interface {
		AssignPermissionToGroup(authGroupID uint, mapping *models.AuthGroupPermissionMapping) error
	}

	permAdapter, ok := adapter.(PermissionOperable)
	if !ok {
		return fmt.Errorf("应用类型 %s 不支持权限操作", app.Type)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 解析端点配置
	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
		}
	}

	// 调用适配器的权限分配方法
	if err := permAdapter.AssignPermissionToGroup(mapping.AuthGroupID, mapping); err != nil {
		return fmt.Errorf("分配权限到外部应用失败: %w", err)
	}

	logger.Info("成功同步权限到外部应用",
		zap.Uint("mappingId", mapping.ID),
		zap.Uint("appId", app.ID),
		zap.String("appName", app.Name))

	return nil
}

// GetAuthGroupPermissions 获取授权中心用户组的所有权限映射
func (s *PermissionMappingService) GetAuthGroupPermissions(authGroupID uint) ([]models.AuthGroupEffectivePermission, error) {
	var mappings []models.AuthGroupEffectivePermission

	query := `
		SELECT
			m.id,
			m.auth_group_id,
			g.name AS auth_group_name,
			m.app_id,
			a.name AS app_name,
			a.type AS app_type,
			m.mapping_type,
			m.external_id,
			m.external_name,
			m.permission_detail,
			m.is_enabled,
			m.expire_time
		FROM auth_group_permission_mappings m
		INNER JOIN auth_groups g ON m.auth_group_id = g.id
		INNER JOIN applications a ON m.app_id = a.id
		WHERE m.auth_group_id = ? AND m.is_enabled = true
		ORDER BY a.name, m.mapping_type
	`

	if err := s.db.Raw(query, authGroupID).Scan(&mappings).Error; err != nil {
		return nil, err
	}

	return mappings, nil
}

// GetUserEffectivePermissions 获取用户的有效权限（通过用户组继承）
func (s *PermissionMappingService) GetUserEffectivePermissions(userID uint) ([]models.AuthGroupEffectivePermission, error) {
	var permissions []models.AuthGroupEffectivePermission

	query := `
		SELECT DISTINCT
			m.id,
			m.auth_group_id,
			g.name AS auth_group_name,
			m.app_id,
			a.name AS app_name,
			a.type AS app_type,
			m.mapping_type,
			m.external_id,
			m.external_name,
			m.permission_detail,
			m.is_enabled,
			m.expire_time
		FROM auth_group_permission_mappings m
		INNER JOIN auth_groups g ON m.auth_group_id = g.id
		INNER JOIN applications a ON m.app_id = a.id
		INNER JOIN auth_user_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = ?
			AND m.is_enabled = true
			AND (m.expire_time IS NULL OR m.expire_time > NOW())
		ORDER BY a.name, m.mapping_type
	`

	if err := s.db.Raw(query, userID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// RevokePermissionFromAuthGroup 撤销授权中心用户组的外部应用权限
func (s *PermissionMappingService) RevokePermissionFromAuthGroup(mappingID uint, revokedBy string) error {
	// 1. 获取映射记录
	var mapping models.AuthGroupPermissionMapping
	if err := s.db.First(&mapping, mappingID).Error; err != nil {
		return fmt.Errorf("权限映射不存在 (ID: %d)", mappingID)
	}

	// 2. 获取应用信息
	var app models.Application
	if err := s.db.First(&app, mapping.AppID).Error; err != nil {
		return fmt.Errorf("应用不存在 (ID: %d)", mapping.AppID)
	}

	// 3. 调用外部应用API撤销权限
	if syncErr := s.revokePermissionFromExternalApp(&mapping, &app); syncErr != nil {
		logger.Warn("撤销外部应用权限失败",
			zap.Uint("mappingId", mappingID),
			zap.Error(syncErr))
		// 继续执行，删除映射记录
	}

	// 4. 删除映射记录
	if err := s.db.Delete(&mapping).Error; err != nil {
		return fmt.Errorf("删除权限映射失败: %w", err)
	}

	logger.Info("成功撤销授权中心用户组的外部应用权限",
		zap.Uint("mappingId", mappingID),
		zap.Uint("authGroupId", mapping.AuthGroupID),
		zap.Uint("appId", app.ID),
		zap.String("appName", app.Name),
		zap.String("revokedBy", revokedBy))

	return nil
}

// revokePermissionFromExternalApp 调用外部应用API撤销权限
func (s *PermissionMappingService) revokePermissionFromExternalApp(mapping *models.AuthGroupPermissionMapping, app *models.Application) error {
	// 获取对应类型的适配器
	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		return fmt.Errorf("获取应用适配器失败: %w", err)
	}

	// 检查适配器是否支持权限操作
	type PermissionOperable interface {
		RevokePermissionFromGroup(authGroupID uint, mapping *models.AuthGroupPermissionMapping) error
	}

	permAdapter, ok := adapter.(PermissionOperable)
	if !ok {
		return fmt.Errorf("应用类型 %s 不支持权限撤销操作", app.Type)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 解析端点配置
	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
		}
	}

	// 调用适配器的权限撤销方法
	if err := permAdapter.RevokePermissionFromGroup(mapping.AuthGroupID, mapping); err != nil {
		return fmt.Errorf("从外部应用撤销权限失败: %w", err)
	}

	logger.Info("成功从外部应用撤销权限",
		zap.Uint("mappingId", mapping.ID),
		zap.Uint("appId", app.ID),
		zap.String("appName", app.Name))

	return nil
}

// SyncAllPendingPermissions 同步所有待处理的权限映射
// 定时任务方法，用于权限同步失败重试
func (s *PermissionMappingService) SyncAllPendingPermissions() error {
	var pendingMappings []models.AuthGroupPermissionMapping

	// 查找待同步或同步失败的映射
	if err := s.db.Where("sync_status IN (?, ?) AND is_enabled = true", "pending", "failed").
		Find(&pendingMappings).Error; err != nil {
		return fmt.Errorf("查询待同步权限映射失败: %w", err)
	}

	logger.Info("开始同步待处理的权限映射",
		zap.Int("count", len(pendingMappings)))

	successCount := 0
	failedCount := 0

	for _, mapping := range pendingMappings {
		// 获取应用信息
		var app models.Application
		if err := s.db.First(&app, mapping.AppID).Error; err != nil {
			logger.Error("获取应用信息失败",
				zap.Uint("mappingId", mapping.ID),
				zap.Error(err))
			failedCount++
			continue
		}

		// 尝试同步
		if err := s.syncPermissionToExternalApp(&mapping, &app); err != nil {
			logger.Error("同步权限到外部应用失败",
				zap.Uint("mappingId", mapping.ID),
				zap.Error(err))
			s.db.Model(&mapping).Updates(map[string]interface{}{
				"sync_status":          "failed",
				"sync_error_message":   err.Error(),
			})
			failedCount++
		} else {
			now := time.Now()
			s.db.Model(&mapping).Updates(map[string]interface{}{
				"sync_status":    "success",
				"last_synced_at": &now,
				"sync_error_message": nil,
			})
			successCount++
		}
	}

	logger.Info("权限同步完成",
		zap.Int("total", len(pendingMappings)),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount))

	return nil
}

// CreatePermissionTemplate 创建权限模板
func (s *PermissionMappingService) CreatePermissionTemplate(appID uint, permissionCode, permissionName, permissionType, description string, template map[string]interface{}) (*models.ApplicationPermissionTemplate, error) {
	// 验证应用是否存在
	var app models.Application
	if err := s.db.First(&app, appID).Error; err != nil {
		return nil, fmt.Errorf("应用不存在 (ID: %d)", appID)
	}

	templateJSON, _ := json.Marshal(template)

	permTemplate := &models.ApplicationPermissionTemplate{
		AppID:          appID,
		PermissionCode:  permissionCode,
		PermissionName:  permissionName,
		PermissionType:  permissionType,
		Description:    description,
		Template:       string(templateJSON),
		IsSystemPreset: false,
	}

	if err := s.db.Create(permTemplate).Error; err != nil {
		return nil, fmt.Errorf("创建权限模板失败: %w", err)
	}

	logger.Info("成功创建权限模板",
		zap.Uint("appId", appID),
		zap.String("appName", app.Name),
		zap.String("permissionCode", permissionCode))

	return permTemplate, nil
}

// GetPermissionTemplates 获取应用的权限模板列表
func (s *PermissionMappingService) GetPermissionTemplates(appID uint) ([]models.ApplicationPermissionTemplate, error) {
	var templates []models.ApplicationPermissionTemplate

	if err := s.db.Where("app_id = ?", appID).
		Order("is_system_preset DESC, sort_order ASC, permission_name ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}

	return templates, nil
}
