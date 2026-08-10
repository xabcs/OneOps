package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"

	"gorm.io/gorm"
)

// AgentRepository Agent数据访问层
type AgentRepository struct {
	db *gorm.DB
}

// NewAgentRepository 创建Agent仓库
func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// ========== 服务器相关查询（Agent专用） ==========

// FindServerForAgent 加载主机及其系统运维凭证
func (r *AgentRepository) FindServerForAgent(serverID uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	if err := r.db.Preload("SystemCredential").First(&server, serverID).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// FindServerByID 根据ID获取服务器
func (r *AgentRepository) FindServerByID(serverID uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	if err := r.db.First(&server, serverID).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// FindServerByIPOrHostname 根据IP或主机名查找服务器（用于心跳上报）
func (r *AgentRepository) FindServerByIPOrHostname(ip, hostname string) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	err := r.db.Where("ip = ? OR hostname = ?", ip, hostname).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// UpdateServerFields 更新服务器指定字段
func (r *AgentRepository) UpdateServerFields(serverID uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.Server{}).Where("id = ?", serverID).Updates(updates).Error
}

// PluckAgentServerIDs 获取所有可能运行Agent的主机ID
func (r *AgentRepository) PluckAgentServerIDs() ([]uint, error) {
	var serverIDs []uint
	err := r.db.Model(&modelcmdb.Server{}).
		Where("agent_status IN ('running', 'offline', 'failed')").
		Pluck("id", &serverIDs).Error
	return serverIDs, err
}

// MarkHeartbeatTimeouts 将超过阈值未上报心跳的主机状态置为 offline
func (r *AgentRepository) MarkHeartbeatTimeouts(threshold interface{}) (int64, error) {
	result := r.db.Model(&modelcmdb.Server{}).
		Where("agent_status = 'running' AND last_heartbeat_at < ?", threshold).
		Update("agent_status", "offline")
	return result.RowsAffected, result.Error
}

// ========== AgentVersion CRUD ==========

// FindLatestVersion 获取最新版本信息
func (r *AgentRepository) FindLatestVersion() (*modelcmdb.AgentVersion, error) {
	var version modelcmdb.AgentVersion
	err := r.db.Where("is_latest = ?", true).First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// FindAllVersions 获取所有版本列表
func (r *AgentRepository) FindAllVersions() ([]modelcmdb.AgentVersion, error) {
	var versions []modelcmdb.AgentVersion
	err := r.db.Order("released_at DESC").Find(&versions).Error
	return versions, err
}

// FindVersionByID 根据ID获取版本信息
func (r *AgentRepository) FindVersionByID(versionID uint) (*modelcmdb.AgentVersion, error) {
	var version modelcmdb.AgentVersion
	err := r.db.First(&version, versionID).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// FindVersionByNumber 根据版本号获取版本信息
func (r *AgentRepository) FindVersionByNumber(versionNumber string) (*modelcmdb.AgentVersion, error) {
	var version modelcmdb.AgentVersion
	err := r.db.Where("version = ?", versionNumber).First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// FindVersionByNumberOrNil 根据版本号获取版本信息（不包装错误消息）
func (r *AgentRepository) FindVersionByNumberOrNil(versionNumber string) (*modelcmdb.AgentVersion, error) {
	var version modelcmdb.AgentVersion
	err := r.db.Where("version = ?", versionNumber).First(&version).Error
	return &version, err
}

// FindExistingVersion 检查版本号是否已存在
func (r *AgentRepository) FindExistingVersion(versionNumber string) (*modelcmdb.AgentVersion, error) {
	var version modelcmdb.AgentVersion
	err := r.db.Where("version = ?", versionNumber).First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// ClearLatestFlags 清除所有最新版本标记
func (r *AgentRepository) ClearLatestFlags() error {
	return r.db.Model(&modelcmdb.AgentVersion{}).Where("is_latest = ?", true).Update("is_latest", false).Error
}

// ClearLatestFlagsExcludeID 清除其他版本的最新标记（排除指定ID）
func (r *AgentRepository) ClearLatestFlagsExcludeID(excludeID uint) error {
	return r.db.Model(&modelcmdb.AgentVersion{}).Where("is_latest = ? AND id != ?", true, excludeID).Update("is_latest", false).Error
}

// CreateVersion 创建新版本
func (r *AgentRepository) CreateVersion(version *modelcmdb.AgentVersion) error {
	return r.db.Create(version).Error
}

// UpdateVersion 更新版本信息
func (r *AgentRepository) UpdateVersion(versionID uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.AgentVersion{}).Where("id = ?", versionID).Updates(updates).Error
}

// DeleteVersion 删除版本
func (r *AgentRepository) DeleteVersion(version *modelcmdb.AgentVersion) error {
	return r.db.Delete(version).Error
}

// CountServersByVersion 统计使用指定版本的主机数量
func (r *AgentRepository) CountServersByVersion(versionNumber string) (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.Server{}).Where("agent_version = ?", versionNumber).Count(&count).Error
	return count, err
}

// IncrementDeployCount 增加版本部署次数
func (r *AgentRepository) IncrementDeployCount(versionNumber string) error {
	return r.db.Model(&modelcmdb.AgentVersion{}).Where("version = ?", versionNumber).
		UpdateColumn("deploy_count", gorm.Expr("deploy_count + ?", 1)).Error
}

// ========== AgentUpgradeTask CRUD ==========

// CreateUpgradeTask 创建升级任务
func (r *AgentRepository) CreateUpgradeTask(task *modelcmdb.AgentUpgradeTask) error {
	return r.db.Create(task).Error
}

// UpdateUpgradeTask 更新升级任务
func (r *AgentRepository) UpdateUpgradeTask(taskID uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.AgentUpgradeTask{}).Where("id = ?", taskID).Updates(updates).Error
}

// FindUpgradeTasks 获取升级任务列表
func (r *AgentRepository) FindUpgradeTasks(page, pageSize int, status string) ([]modelcmdb.AgentUpgradeTask, int64, error) {
	var tasks []modelcmdb.AgentUpgradeTask
	var total int64

	query := r.db.Model(&modelcmdb.AgentUpgradeTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// FindUpgradeTaskByID 获取升级任务详情
func (r *AgentRepository) FindUpgradeTaskByID(taskID uint) (*modelcmdb.AgentUpgradeTask, error) {
	var task modelcmdb.AgentUpgradeTask
	err := r.db.First(&task, taskID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
