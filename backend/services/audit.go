package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/models"
	"oneops/backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditService 审计服务
type AuditService struct{}

// NewAuditService 创建审计服务
func NewAuditService() *AuditService {
	return &AuditService{}
}

// LogLogin 记录登录日志
func (s *AuditService) LogLogin(userID uint, username, nickname, ip, userAgent, location, status, failReason string) error {
	log := models.LoginLog{
		UserID:     userID,
		Username:   username,
		Nickname:   nickname,
		IP:         ip,
		UserAgent:  userAgent,
		Location:   location,
		Status:     status,
		FailReason: failReason,
		LoginTime:  time.Now(),
	}

	return db.Create(&log).Error
}

// LogLogout 记录登出日志
func (s *AuditService) LogLogout(userID uint) error {
	// 查找最近的成功登录记录
	var loginLog models.LoginLog
	err := db.Where("user_id = ? AND status = ?", userID, "success").
		Order("login_time DESC").
		First(&loginLog).Error

	if err != nil {
		return err
	}

	// 计算会话时长
	duration := int(time.Since(loginLog.LoginTime).Seconds())

	// 更新登出时间和会话时长
	now := time.Now()
	return db.Model(&loginLog).
		Updates(map[string]interface{}{
			"logout_time": &now,
			"duration":    duration,
		}).Error
}

// LogOperation 记录操作日志
func (s *AuditService) LogOperation(userID uint, username, nickname, module, action, description, method, path string,
	params, response interface{}, statusCode int, ip, userAgent string, duration int, status, errorMsg string) error {

	// 序列化参数和响应
	paramsJSON, _ := json.Marshal(params)
	responseJSON, _ := json.Marshal(response)

	log := models.OperationLog{
		UserID:      userID,
		Username:    username,
		Nickname:    nickname,
		Module:      module,
		Action:      action,
		Description: description,
		Method:      method,
		Path:        path,
		Params:      string(paramsJSON),
		Response:    string(responseJSON),
		StatusCode:  statusCode,
		IP:          ip,
		UserAgent:   userAgent,
		Duration:    duration,
		Status:      status,
		ErrorMsg:    errorMsg,
		OperateTime: time.Now(),
	}

	return db.Create(&log).Error
}

// LogSystemEvent 记录系统事件日志
func (s *AuditService) LogSystemEvent(level, source, category, message, details, ip string) error {
	log := models.SystemEventLog{
		Level:     level,
		Source:    source,
		Category:  category,
		Message:   message,
		Details:   details,
		IP:        ip,
		EventTime: time.Now(),
	}

	return db.Create(&log).Error
}

// LoginLogResponse 登录日志响应结构
type LoginLogResponse struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	IP       string `json:"ip"`
	Location string `json:"location"`
	Browser  string `json:"browser"`
	OS       string `json:"os"`
	Status   string `json:"status"`
	Msg      string `json:"msg"`
}

// GetLoginLogs 获取登录日志列表（用于前端显示）
func (s *AuditService) GetLoginLogs(query map[string]interface{}, page, pageSize int) ([]LoginLogResponse, int64, error) {
	var logs []models.LoginLog
	var total int64

	tx := db.Model(&models.LoginLog{})

	// 构建查询条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := tx.Order("login_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	response := make([]LoginLogResponse, 0, len(logs))
	for _, log := range logs {
		userAgentInfo := utils.ParseUserAgent(log.UserAgent)
		logResponse := LoginLogResponse{
			Time:     log.LoginTime.Format("2006-01-02 15:04:05"),
			Username: log.Username,
			IP:       log.IP,
			Location: log.Location,
			Browser:  userAgentInfo.Browser,
			OS:       userAgentInfo.OS,
			Status:   log.Status,
			Msg:      log.FailReason,
		}

		// 应用浏览器和操作系统过滤（在解析后过滤）
		if browser, ok := query["browser"].(string); ok && browser != "" {
			if userAgentInfo.Browser != browser {
				continue
			}
		}
		if os, ok := query["os"].(string); ok && os != "" {
			if userAgentInfo.OS != os {
				continue
			}
		}

		response = append(response, logResponse)
	}

	// 更新总数（应用浏览器和OS过滤后）
	total = int64(len(response))

	// 应用分页（在内存中分页，因为浏览器和OS过滤是在解析后进行的）
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	return response[start:end], total, nil
}

// GetLoginLogsForExport 获取登录日志列表（用于导出）
func (s *AuditService) GetLoginLogsForExport(query map[string]interface{}, page, pageSize int) ([]models.LoginLog, int64, error) {
	var logs []models.LoginLog
	var total int64

	tx := db.Model(&models.LoginLog{})

	// 构建查询条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := tx.Order("login_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetOperationLogs 获取操作日志列表
func (s *AuditService) GetOperationLogs(query map[string]interface{}, page, pageSize int) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	tx := db.Model(&models.OperationLog{})

	// 构建查询条件
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
		// 尝试将字符串转换为数字
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := tx.Order("operate_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	// 格式化时间字段
	for i := range logs {
		logs[i].Time = logs[i].OperateTime.Format("2006-01-02 15:04:05")
	}

	return logs, total, nil
}

// GetSystemEventLogs 获取系统事件日志列表
func (s *AuditService) GetSystemEventLogs(query map[string]interface{}, page, pageSize int) ([]models.SystemEventLog, int64, error) {
	var logs []models.SystemEventLog
	var total int64

	tx := db.Model(&models.SystemEventLog{})

	// 构建查询条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := tx.Order("event_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// GetAuditStats 获取审计统计信息
func (s *AuditService) GetAuditStats() (gin.H, error) {
	var stats gin.H = make(gin.H)

	// 登录统计
	var loginStats []struct {
		Status string
		Count  int64
	}
	db.Model(&models.LoginLog{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&loginStats)

	stats["login"] = gin.H{
		"total":     0,
		"success":   0,
		"failed":    0,
		"today":     0,
		"thisWeek":  0,
		"thisMonth": 0,
	}

	for _, stat := range loginStats {
		stats["login"].(gin.H)["total"] = stats["login"].(gin.H)["total"].(int) + int(stat.Count)
		if stat.Status == "success" {
			stats["login"].(gin.H)["success"] = stat.Count
		} else {
			stats["login"].(gin.H)["failed"] = stat.Count
		}
	}

	// 操作统计
	var opStats []struct {
		Status string
		Count  int64
	}
	db.Model(&models.OperationLog{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&opStats)

	stats["operation"] = gin.H{
		"total":   0,
		"success": 0,
		"failed":  0,
	}

	for _, stat := range opStats {
		stats["operation"].(gin.H)["total"] = stats["operation"].(gin.H)["total"].(int) + int(stat.Count)
		if stat.Status == "success" {
			stats["operation"].(gin.H)["success"] = stat.Count
		} else {
			stats["operation"].(gin.H)["failed"] = stat.Count
		}
	}

	// 今日登录次数
	var todayLoginCount int64
	today := time.Now().Format("2006-01-02")
	db.Model(&models.LoginLog{}).
		Where("DATE(login_time) = ?", today).
		Count(&todayLoginCount)
	stats["login"].(gin.H)["today"] = todayLoginCount

	return stats, nil
}

// GetModules 获取可用的审计模块列表（从menu表获取一级菜单）
func (s *AuditService) GetModules() []string {
	var modules []string

	// 从menu表中查询所有一级菜单（parent_id = 0）
	err := db.Model(&models.Menu{}).
		Where("parent_id = 0").
		Where("status = 1"). // 只查询启用的菜单
		Order("sort ASC").
		Pluck("name", &modules).Error

	if err != nil {
		// 如果查询失败，返回空列表
		return []string{}
	}

	return modules
}
