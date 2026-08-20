package system

import (
	"errors"
	"fmt"
	"strings"

	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// UserGroupService 平台用户组服务（组成员为 sys_users，组用于集群等资源的批量授权）
type UserGroupService struct {
	db *gorm.DB
}

// NewUserGroupService 创建用户组服务
func NewUserGroupService(db *gorm.DB) *UserGroupService {
	return &UserGroupService{db: db}
}

// GetUserGroups 分页查询用户组（可按名称/编码模糊过滤），含成员数
func (s *UserGroupService) GetUserGroups(page, pageSize int, keyword string) ([]map[string]interface{}, int64, error) {
	query := s.db.Model(&modelsystem.UserGroup{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var groups []modelsystem.UserGroup
	if err := query.Order("id").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(groups))
	for i, g := range groups {
		var memberCount int64
		s.db.Model(&modelsystem.UserGroupMember{}).Where("group_id = ?", g.ID).Count(&memberCount)
		result[i] = map[string]interface{}{
			"id":          g.ID,
			"code":        g.Code,
			"name":        g.Name,
			"description": g.Description,
			"status":      g.Status,
			"memberCount": memberCount,
			"createdAt":   g.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, total, nil
}

// GetUserGroupOptions 用户组选项（供选择器）
func (s *UserGroupService) GetUserGroupOptions() ([]map[string]interface{}, error) {
	var groups []modelsystem.UserGroup
	if err := s.db.Where("status = 1").Order("id").Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(groups))
	for i, g := range groups {
		result[i] = map[string]interface{}{"id": g.ID, "code": g.Code, "name": g.Name}
	}
	return result, nil
}

// CreateUserGroup 创建用户组
func (s *UserGroupService) CreateUserGroup(req modelsystem.CreateUserGroupRequest) error {
	code := strings.TrimSpace(req.Code)
	var count int64
	s.db.Model(&modelsystem.UserGroup{}).Where("code = ?", code).Count(&count)
	if count > 0 {
		return fmt.Errorf("用户组编码已存在: %s", code)
	}
	return s.db.Create(&modelsystem.UserGroup{
		Code:        code,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}).Error
}

// UpdateUserGroup 更新用户组（code 不可改）
func (s *UserGroupService) UpdateUserGroup(id uint, req modelsystem.UpdateUserGroupRequest) error {
	updates := map[string]interface{}{"name": req.Name, "description": req.Description}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	return s.db.Model(&modelsystem.UserGroup{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteUserGroup 删除用户组（级联删除组成员与集群组绑定）
func (s *UserGroupService) DeleteUserGroup(id uint) error {
	var group modelsystem.UserGroup
	if err := s.db.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("用户组不存在")
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&modelsystem.UserGroupMember{}).Error; err != nil {
			return err
		}
		// 集群组绑定表（k8s_cluster_group_bindings）由 k8s 模块维护，此处按表名清理
		if err := tx.Exec("DELETE FROM k8s_cluster_group_bindings WHERE group_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&group).Error
	})
}

// GetGroupMembers 查询组成员列表
func (s *UserGroupService) GetGroupMembers(groupID uint) ([]map[string]interface{}, error) {
	var members []modelsystem.UserGroupMember
	if err := s.db.Where("group_id = ?", groupID).Preload("User").Find(&members).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		username, nickname := "", ""
		if m.User != nil {
			username = m.User.Username
			nickname = m.User.Nickname
		}
		result[i] = map[string]interface{}{
			"userId":    m.UserID,
			"username":  username,
			"nickname":  nickname,
			"createdAt": m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

// AddGroupMembers 批量添加组成员（跳过已存在与无效用户）
func (s *UserGroupService) AddGroupMembers(groupID uint, userIDs []uint) (int, error) {
	var group modelsystem.UserGroup
	if err := s.db.First(&group, groupID).Error; err != nil {
		return 0, fmt.Errorf("用户组不存在")
	}

	added := 0
	for _, uid := range userIDs {
		var user modelsystem.User
		if err := s.db.First(&user, uid).Error; err != nil {
			continue
		}
		var count int64
		s.db.Model(&modelsystem.UserGroupMember{}).
			Where("group_id = ? AND user_id = ?", groupID, uid).Count(&count)
		if count > 0 {
			continue
		}
		if err := s.db.Create(&modelsystem.UserGroupMember{GroupID: groupID, UserID: uid}).Error; err == nil {
			added++
		}
	}
	return added, nil
}

// RemoveGroupMember 移除组成员
func (s *UserGroupService) RemoveGroupMember(groupID uint, userID uint) error {
	result := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&modelsystem.UserGroupMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("成员不存在")
	}
	return nil
}
