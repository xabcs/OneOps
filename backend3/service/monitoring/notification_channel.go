package monitoring

import (
	"encoding/json"
	"fmt"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID          uint            `json:"id"`
	ChannelType string          `json:"channelType"`
	ChannelName string          `json:"channelName"`
	Config      json.RawMessage `json:"config"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

// GetNotificationChannels 获取通知渠道列表
func (s *MonitoringService) GetNotificationChannels() ([]NotificationChannel, error) {
	var channels []struct {
		ID          uint            `gorm:"column:id"`
		ChannelType string          `gorm:"column:channel_type"`
		ChannelName string          `gorm:"column:channel_name"`
		Config      json.RawMessage `gorm:"column:config"`
		Enabled     bool            `gorm:"column:enabled"`
		CreatedAt   time.Time       `gorm:"column:created_at"`
		UpdatedAt   time.Time       `gorm:"column:updated_at"`
	}

	err := getDB().Table("mon_notification_channels").Find(&channels).Error
	if err != nil {
		return nil, fmt.Errorf("查询通知渠道失败: %w", err)
	}

	result := make([]NotificationChannel, len(channels))
	for i, c := range channels {
		result[i] = NotificationChannel{
			ID:          c.ID,
			ChannelType: c.ChannelType,
			ChannelName: c.ChannelName,
			Config:      c.Config,
			Enabled:     c.Enabled,
			CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, nil
}

// CreateNotificationChannel 创建通知渠道
func (s *MonitoringService) CreateNotificationChannel(channel *NotificationChannel) (uint, error) {
	now := time.Now()
	channel.CreatedAt = now.Format("2006-01-02 15:04:05")
	channel.UpdatedAt = channel.CreatedAt

	result := getDB().Exec("INSERT INTO mon_notification_channels (channel_type, channel_name, config, enabled) VALUES (?, ?, ?, ?)",
		channel.ChannelType, channel.ChannelName, channel.Config, channel.Enabled)
	if result.Error != nil {
		return 0, result.Error
	}

	var id uint
	getDB().Raw("SELECT LAST_INSERT_ID() as id").Scan(&id)
	return id, nil
}

// UpdateNotificationChannel 更新通知渠道
func (s *MonitoringService) UpdateNotificationChannel(channel *NotificationChannel) error {
	query := `UPDATE mon_notification_channels
	          SET channel_type = ?, channel_name = ?, config = ?, enabled = ?
	          WHERE id = ?`

	return getDB().Exec(query, channel.ChannelType, channel.ChannelName, channel.Config, channel.Enabled, channel.ID).Error
}

// DeleteNotificationChannel 删除通知渠道
func (s *MonitoringService) DeleteNotificationChannel(id uint) error {
	return getDB().Exec("DELETE FROM mon_notification_channels WHERE id = ?", id).Error
}

// TestNotificationChannel 测试通知渠道
func (s *MonitoringService) TestNotificationChannel(id uint) error {
	var channel struct {
		ChannelType string          `gorm:"column:channel_type"`
		Config      json.RawMessage `gorm:"column:config"`
	}

	err := getDB().Table("mon_notification_channels").Select("channel_type, config").
		Where("id = ?", id).First(&channel).Error
	if err != nil {
		return fmt.Errorf("获取通知渠道失败: %w", err)
	}

	switch channel.ChannelType {
	case "email":
		logger.Info("发送邮件测试通知", zap.Uint("channelId", id))
	case "wechat":
		logger.Info("发送企业微信测试通知", zap.Uint("channelId", id))
	case "dingtalk":
		logger.Info("发送钉钉测试通知", zap.Uint("channelId", id))
	default:
		return fmt.Errorf("不支持的通知渠道类型: %s", channel.ChannelType)
	}

	return nil
}
