package audit

import (
	"encoding/json"
	"fmt"
	"time"

	modelaudit "oneops/backend3/model/audit"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repoaudit "oneops/backend3/repository/audit"

	"go.uber.org/zap"
)

// AuditService 审计服务
type AuditService struct {
	repo *repoaudit.AuditRepository
}

// NewAuditService 创建审计服务
func NewAuditService(repo *repoaudit.AuditRepository) *AuditService {
	return &AuditService{
		repo: repo,
	}
}

// LogLogin 记录登录日志
func (s *AuditService) LogLogin(userID uint, username, nickname, ip, userAgent, location, status, failReason string) error {
	log := modelaudit.LoginLog{
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

	if err := s.repo.CreateLoginLog(&log); err != nil {
		logger.Error("记录登录日志失败", zap.Error(err))
		return err
	}
	return nil
}

// LogLogout 记录登出日志
func (s *AuditService) LogLogout(userID uint) error {
	loginLog, err := s.repo.FindLastSuccessLoginLog(userID)
	if err != nil {
		logger.Error("查找登录记录失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}

	duration := int(time.Since(loginLog.LoginTime).Seconds())
	now := time.Now()

	if err := s.repo.UpdateLoginLogLogout(loginLog.ID, map[string]interface{}{
		"logout_time": &now,
		"duration":    duration,
	}); err != nil {
		logger.Error("更新登出记录失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}
	return nil
}

// LogOperation 记录操作日志
func (s *AuditService) LogOperation(userID uint, username, nickname, module, action, description, method, path string,
	params, response interface{}, statusCode int, ip, userAgent string, duration int, status, errorMsg string) error {

	paramsJSON, _ := json.Marshal(params)
	responseJSON, _ := json.Marshal(response)

	log := modelaudit.OperationLog{
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

	if err := s.repo.CreateOperationLog(&log); err != nil {
		logger.Error("记录操作日志失败", zap.Error(err))
		return err
	}
	return nil
}

// LogSystemEvent 记录系统事件日志
func (s *AuditService) LogSystemEvent(level, source, category, message, details, ip string) error {
	log := modelaudit.SystemEventLog{
		Level:     level,
		Source:    source,
		Category:  category,
		Message:   message,
		Details:   details,
		IP:        ip,
		EventTime: time.Now(),
	}

	if err := s.repo.CreateSystemEventLog(&log); err != nil {
		logger.Error("记录系统事件日志失败", zap.Error(err))
		return err
	}
	return nil
}

// LoginLogResponse 登录日志响应结构
type LoginLogResponse struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"userId"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
	Location   string `json:"location"`
	Status     string `json:"status"`
	FailReason string `json:"failReason"`
	LoginTime  string `json:"loginTime"`
	LogoutTime string `json:"logoutTime,omitempty"`
	Duration   int    `json:"duration"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
}

// GetLoginLogs 获取登录日志列表
func (s *AuditService) GetLoginLogs(query map[string]interface{}, page, pageSize int) ([]LoginLogResponse, int64, error) {
	logs, _, err := s.repo.FindLoginLogs(query, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	response := make([]LoginLogResponse, 0, len(logs))
	for _, log := range logs {
		userAgentInfo := utils.ParseUserAgent(log.UserAgent)

		duration := 0
		if log.LogoutTime != nil {
			duration = int(log.LogoutTime.Sub(log.LoginTime).Seconds())
		} else if log.Duration > 0 {
			duration = log.Duration
		}

		var logoutTimeStr string
		if log.LogoutTime != nil {
			logoutTimeStr = log.LogoutTime.Format("2006-01-02 15:04:05")
		}

		logResponse := LoginLogResponse{
			ID:         log.ID,
			UserID:     log.UserID,
			Username:   log.Username,
			Nickname:   log.Nickname,
			IP:         log.IP,
			UserAgent:  log.UserAgent,
			Location:   log.Location,
			Status:     log.Status,
			FailReason: log.FailReason,
			LoginTime:  log.LoginTime.Format("2006-01-02 15:04:05"),
			LogoutTime: logoutTimeStr,
			Duration:   duration,
			Browser:    userAgentInfo.Browser,
			OS:         userAgentInfo.OS,
		}

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

	total := int64(len(response))

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
func (s *AuditService) GetLoginLogsForExport(query map[string]interface{}, page, pageSize int) ([]modelaudit.LoginLog, int64, error) {
	return s.repo.FindLoginLogs(query, page, pageSize)
}

// GetOperationLogs 获取操作日志列表
func (s *AuditService) GetOperationLogs(query map[string]interface{}, page, pageSize int) ([]modelaudit.OperationLog, int64, error) {
	logs, total, err := s.repo.FindOperationLogs(query, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	for i := range logs {
		logs[i].Time = logs[i].OperateTime.Format("2006-01-02 15:04:05")
	}

	return logs, total, nil
}

// GetSystemEventLogs 获取系统事件日志列表
func (s *AuditService) GetSystemEventLogs(query map[string]interface{}, page, pageSize int) ([]modelaudit.SystemEventLog, int64, error) {
	logs, total, err := s.repo.FindSystemEventLogs(query, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	for i := range logs {
		logs[i].Time = logs[i].EventTime.Format("2006-01-02 15:04:05")
	}

	return logs, total, nil
}

// formatDuration 格式化会话时长
func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%d分钟", minutes)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, remainingMinutes)
	}
	return fmt.Sprintf("%d小时", hours)
}

// GetAuditStats 获取审计统计信息
func (s *AuditService) GetAuditStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 登录统计
	loginCounts, _ := s.repo.CountLoginLogsByStatus()
	var loginTotal int64
	var loginSuccess int64
	var loginFailed int64
	for status, count := range loginCounts {
		loginTotal += count
		if status == "success" {
			loginSuccess = count
		} else {
			loginFailed += count
		}
	}

	// 今日登录次数
	today := time.Now().Format("2006-01-02")
	todayLoginCount, _ := s.repo.CountLoginLogsOnDate(today)

	// 本周登录次数
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	if time.Now().Weekday() == 0 {
		weekStart = time.Now().AddDate(0, 0, -6)
	}
	weekLoginCount, _ := s.repo.CountLoginLogsSinceDate(weekStart.Format("2006-01-02"))

	// 本月登录次数
	monthLoginCount, _ := s.repo.CountLoginLogsSinceDate(time.Now().Format("2006-01-01"))

	stats["login"] = map[string]interface{}{
		"total":     loginTotal,
		"success":   loginSuccess,
		"failed":    loginFailed,
		"today":     todayLoginCount,
		"thisWeek":  weekLoginCount,
		"thisMonth": monthLoginCount,
	}

	// 操作统计
	opCounts, _ := s.repo.CountOperationLogsByStatus()
	var opTotal int64
	var opSuccess int64
	var opFailed int64
	for status, count := range opCounts {
		opTotal += count
		if status == "success" {
			opSuccess = count
		} else {
			opFailed += count
		}
	}
	stats["operation"] = map[string]interface{}{
		"total":   opTotal,
		"success": opSuccess,
		"failed":  opFailed,
	}

	// 系统事件统计
	eventCounts, _ := s.repo.CountSystemEventLogsByLevel()
	var systemTotal int64
	for _, count := range eventCounts {
		systemTotal += count
	}
	stats["system"] = map[string]interface{}{
		"total":    systemTotal,
		"info":     eventCounts["info"],
		"warning":  eventCounts["warning"],
		"error":    eventCounts["error"],
		"critical": eventCounts["critical"],
	}

	return stats, nil
}

// GetModules 获取可用的审计模块列表（从menu表获取一级菜单）
func (s *AuditService) GetModules() []string {
	modules, err := s.repo.FindModules()
	if err != nil {
		logger.Error("获取审计模块列表失败", zap.Error(err))
		return []string{}
	}
	return modules
}
