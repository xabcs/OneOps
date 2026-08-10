package audit

import (
	"fmt"

	modelaudit "oneops/backend3/model/audit"
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// AuditRepository 审计数据访问层
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建审计仓库
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// ==================== 登录日志 ====================

// CreateLoginLog 创建登录日志
func (r *AuditRepository) CreateLoginLog(log *modelaudit.LoginLog) error {
	return r.db.Create(log).Error
}

// FindLastSuccessLoginLog 查找用户最近一次成功的登录记录
func (r *AuditRepository) FindLastSuccessLoginLog(userID uint) (*modelaudit.LoginLog, error) {
	var log modelaudit.LoginLog
	err := r.db.Where("user_id = ? AND status = ?", userID, "success").
		Order("login_time DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// UpdateLoginLogLogout 根据ID更新登录日志的登出时间和会话时长
func (r *AuditRepository) UpdateLoginLogLogout(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelaudit.LoginLog{}).Where("id = ?", id).Updates(updates).Error
}

// FindLoginLogs 分页查询登录日志
func (r *AuditRepository) FindLoginLogs(query map[string]interface{}, page, pageSize int) ([]modelaudit.LoginLog, int64, error) {
	var logs []modelaudit.LoginLog
	var total int64

	tx := r.db.Model(&modelaudit.LoginLog{})

	if username, ok := query["username"].(string); ok && username != "" {
		tx = tx.Where("username LIKE ?", "%"+username+"%")
	}
	if status, ok := query["status"].(string); ok && status != "" {
		tx = tx.Where("status = ?", status)
	}
	if location, ok := query["location"].(string); ok && location != "" {
		tx = tx.Where("location LIKE ?", "%"+location+"%")
	}
	if startTime, ok := query["startTime"].(string); ok && startTime != "" {
		tx = tx.Where("login_time >= ?", startTime)
	}
	if endTime, ok := query["endTime"].(string); ok && endTime != "" {
		tx = tx.Where("login_time <= ?", endTime)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("login_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ==================== 操作日志 ====================

// CreateOperationLog 创建操作日志
func (r *AuditRepository) CreateOperationLog(log *modelaudit.OperationLog) error {
	return r.db.Create(log).Error
}

// FindOperationLogs 分页查询操作日志
func (r *AuditRepository) FindOperationLogs(query map[string]interface{}, page, pageSize int) ([]modelaudit.OperationLog, int64, error) {
	var logs []modelaudit.OperationLog
	var total int64

	tx := r.db.Model(&modelaudit.OperationLog{})

	if username, ok := query["username"].(string); ok && username != "" {
		tx = tx.Where("username LIKE ?", "%"+username+"%")
	}
	if module, ok := query["module"].(string); ok && module != "" {
		tx = tx.Where("module = ?", module)
	}
	if status, ok := query["status"].(string); ok && status != "" {
		tx = tx.Where("status = ?", status)
	}
	if action, ok := query["action"].(string); ok && action != "" {
		tx = tx.Where("action LIKE ?", "%"+action+"%")
	}
	if method, ok := query["method"].(string); ok && method != "" {
		tx = tx.Where("method = ?", method)
	}
	if statusCode, ok := query["statusCode"].(int); ok && statusCode > 0 {
		tx = tx.Where("status_code = ?", statusCode)
	}
	if statusCodeStr, ok := query["statusCode"].(string); ok && statusCodeStr != "" {
		var code int
		if _, err := fmt.Sscanf(statusCodeStr, "%d", &code); err == nil {
			tx = tx.Where("status_code = ?", code)
		}
	}
	if path, ok := query["path"].(string); ok && path != "" {
		tx = tx.Where("path LIKE ?", "%"+path+"%")
	}
	if durationRange, ok := query["durationRange"].(string); ok && durationRange != "" {
		switch durationRange {
		case "fast":
			tx = tx.Where("duration < 100")
		case "normal":
			tx = tx.Where("duration >= 100 AND duration < 500")
		case "slow":
			tx = tx.Where("duration >= 500 AND duration < 1000")
		case "very-slow":
			tx = tx.Where("duration >= 1000")
		}
	}
	if startTime, ok := query["startTime"].(string); ok && startTime != "" {
		tx = tx.Where("operate_time >= ?", startTime)
	}
	if endTime, ok := query["endTime"].(string); ok && endTime != "" {
		tx = tx.Where("operate_time <= ?", endTime)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("operate_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ==================== 系统事件日志 ====================

// CreateSystemEventLog 创建系统事件日志
func (r *AuditRepository) CreateSystemEventLog(log *modelaudit.SystemEventLog) error {
	return r.db.Create(log).Error
}

// FindSystemEventLogs 分页查询系统事件日志
func (r *AuditRepository) FindSystemEventLogs(query map[string]interface{}, page, pageSize int) ([]modelaudit.SystemEventLog, int64, error) {
	var logs []modelaudit.SystemEventLog
	var total int64

	tx := r.db.Model(&modelaudit.SystemEventLog{})

	if level, ok := query["level"].(string); ok && level != "" {
		tx = tx.Where("level = ?", level)
	}
	if source, ok := query["source"].(string); ok && source != "" {
		tx = tx.Where("source LIKE ?", "%"+source+"%")
	}
	if category, ok := query["category"].(string); ok && category != "" {
		tx = tx.Where("category = ?", category)
	}
	if startTime, ok := query["startTime"].(string); ok && startTime != "" {
		tx = tx.Where("event_time >= ?", startTime)
	}
	if endTime, ok := query["endTime"].(string); ok && endTime != "" {
		tx = tx.Where("event_time <= ?", endTime)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("event_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ==================== 统计 ====================

// CountLoginLogsByStatus 按状态分组统计登录日志数量
func (r *AuditRepository) CountLoginLogsByStatus() (map[string]int64, error) {
	var stats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&modelaudit.LoginLog{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(stats))
	for _, s := range stats {
		result[s.Status] = s.Count
	}
	return result, nil
}

// CountOperationLogsByStatus 按状态分组统计操作日志数量
func (r *AuditRepository) CountOperationLogsByStatus() (map[string]int64, error) {
	var stats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&modelaudit.OperationLog{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(stats))
	for _, s := range stats {
		result[s.Status] = s.Count
	}
	return result, nil
}

// CountSystemEventLogsByLevel 按级别分组统计系统事件日志数量
func (r *AuditRepository) CountSystemEventLogsByLevel() (map[string]int64, error) {
	var stats []struct {
		Level string
		Count int64
	}
	if err := r.db.Model(&modelaudit.SystemEventLog{}).
		Select("level, count(*) as count").
		Group("level").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(stats))
	for _, s := range stats {
		result[s.Level] = s.Count
	}
	return result, nil
}

// CountLoginLogsOnDate 统计指定日期的登录次数
func (r *AuditRepository) CountLoginLogsOnDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&modelaudit.LoginLog{}).
		Where("DATE(created_at) = ?", date).
		Count(&count).Error
	return count, err
}

// CountLoginLogsSinceDate 统计指定日期起的登录次数
func (r *AuditRepository) CountLoginLogsSinceDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&modelaudit.LoginLog{}).
		Where("DATE(created_at) >= ?", date).
		Count(&count).Error
	return count, err
}

// ==================== 模块 ====================

// FindModules 查询可用的审计模块列表（一级菜单）
func (r *AuditRepository) FindModules() ([]string, error) {
	var modules []string
	err := r.db.Model(&modelsystem.Menu{}).
		Where("parent_id = 0").
		Where("status = 1").
		Order("sort ASC").
		Pluck("name", &modules).Error
	return modules, err
}
