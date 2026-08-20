package k8s

import (
	"encoding/json"
	"fmt"
	modelk8s "oneops/backend3/model/k8s"
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// ClusterRepository K8s集群数据访问层
type ClusterRepository struct {
	db *gorm.DB
}

// NewClusterRepository 创建K8s集群仓库
func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

// ClusterQuery 集群分页查询条件
type ClusterQuery struct {
	Filter *modelk8s.K8sClusterFilter
	// AuthorizedIDs 数据权限：用户可见的集群ID集合（nil 表示不限制，如超管）
	// 非空时 count 与 list 同条件过滤，保证 total 与 list 一致
	AuthorizedIDs *[]uint
	Page          int
	PageSize      int
}

// FindActiveByID 根据ID和启用状态（status=1）查询集群
func (r *ClusterRepository) FindActiveByID(clusterID uint) (*modelk8s.K8sCluster, error) {
	var cluster modelk8s.K8sCluster
	if err := r.db.Where("id = ? AND status = 1", clusterID).First(&cluster).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

// FindByID 根据ID查询集群
func (r *ClusterRepository) FindByID(clusterID uint) (*modelk8s.K8sCluster, error) {
	var cluster modelk8s.K8sCluster
	if err := r.db.First(&cluster, clusterID).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

// FindWithPagination 分页查询集群列表
func (r *ClusterRepository) FindWithPagination(q ClusterQuery) ([]modelk8s.K8sCluster, int64, error) {
	var clusters []modelk8s.K8sCluster
	var total int64

	query := r.db.Model(&modelk8s.K8sCluster{})

	if q.AuthorizedIDs != nil {
		query = query.Where("id IN ?", *q.AuthorizedIDs)
	}

	if q.Filter != nil {
		if q.Filter.Name != nil && *q.Filter.Name != "" {
			query = query.Where("name LIKE ?", "%"+*q.Filter.Name+"%")
		}
		if q.Filter.Status != nil {
			query = query.Where("status = ?", *q.Filter.Status)
		}
		if q.Filter.ClusterType != nil && *q.Filter.ClusterType != "" {
			query = query.Where("cluster_type = ?", *q.Filter.ClusterType)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	return clusters, total, nil
}

// FindAuthorizedClusterIDs 查询用户可见的全部集群ID（数据权限范围）
// 三档下线后：直绑/组绑定为历史存量，原生 RBAC 绑定为现行授权，任一存在即可见
func (r *ClusterRepository) FindAuthorizedClusterIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		SELECT cluster_id FROM k8s_cluster_role_bindings WHERE user_id = ?
		UNION
		SELECT gb.cluster_id FROM k8s_cluster_group_bindings gb
			JOIN sys_user_group_members m ON m.group_id = gb.group_id
			WHERE m.user_id = ?
		UNION
		SELECT cluster_id FROM k8s_native_role_bindings WHERE subject_type = 'user' AND user_id = ?
		UNION
		SELECT nb.cluster_id FROM k8s_native_role_bindings nb
			JOIN sys_user_group_members m ON m.group_id = nb.group_id
			WHERE nb.subject_type = 'group' AND m.user_id = ?`,
		userID, userID, userID, userID).Scan(&ids).Error
	return ids, err
}

// IsSuperAdmin 判定用户是否平台超级管理员（角色 ID=1），
// 超管是唯一允许以平台身份执行集群操作的白名单
func (r *ClusterRepository) IsSuperAdmin(userID uint) bool {
	user, err := r.FindUserByID(userID)
	if err != nil {
		return false
	}
	var roleIDs []uint
	if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
		return false
	}
	for _, roleID := range roleIDs {
		if roleID == 1 {
			return true
		}
	}
	return false
}

// Create 创建集群
func (r *ClusterRepository) Create(cluster *modelk8s.K8sCluster) error {
	return r.db.Create(cluster).Error
}

// UpdateByID 根据集群记录更新指定字段
func (r *ClusterRepository) UpdateByID(cluster *modelk8s.K8sCluster, updates map[string]interface{}) error {
	return r.db.Model(cluster).Updates(updates).Error
}

// DeleteClusterCascade 级联删除集群及其关联的角色绑定和会话记录（事务）
func (r *ClusterRepository) DeleteClusterCascade(clusterID uint, cluster *modelk8s.K8sCluster) error {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("cluster_id = ?", clusterID).Delete(&modelk8s.ClusterRoleBinding{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("cluster_id = ?", clusterID).Delete(&modelk8s.K8sSession{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(cluster).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// FindUserByID 根据ID查询用户
func (r *ClusterRepository) FindUserByID(userID uint) (*modelsystem.User, error) {
	var user modelsystem.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserGroupByID 根据ID查询平台用户组
func (r *ClusterRepository) FindUserGroupByID(groupID uint) (*modelsystem.UserGroup, error) {
	var group modelsystem.UserGroup
	if err := r.db.First(&group, groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// FindSubjectOptions 授权主体候选（k8s 原生授权表单专用，仅返回所需字段）
// 与系统管理页的 users/options 解耦：授权上下文的权限归属 k8s.permission.*，不借用 system.user.list
// 用户附带 groupIds：授权管理页用户视角需叠加"经组继承"的有效授权
func (r *ClusterRepository) FindSubjectOptions() (users []map[string]interface{}, groups []map[string]interface{}, err error) {
	var userRows []struct {
		ID       uint
		Username string
		Nickname string
	}
	if err = r.db.Model(&modelsystem.User{}).
		Where("status = ?", "active").
		Select("id, username, nickname").
		Order("id ASC").
		Find(&userRows).Error; err != nil {
		return nil, nil, err
	}

	if err = r.db.Model(&modelsystem.UserGroup{}).
		Select("id, name, code").
		Order("id ASC").
		Find(&groups).Error; err != nil {
		return nil, nil, err
	}

	// 组成员关系 → 用户 groupIds
	var members []modelsystem.UserGroupMember
	if err = r.db.Find(&members).Error; err != nil {
		return nil, nil, err
	}
	userGroups := make(map[uint][]uint)
	for _, m := range members {
		userGroups[m.UserID] = append(userGroups[m.UserID], m.GroupID)
	}

	users = make([]map[string]interface{}, 0, len(userRows))
	for _, u := range userRows {
		users = append(users, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
			"groupIds": userGroups[u.ID],
		})
	}
	return users, groups, nil
}

// FindClusterOptions 集群候选（授权管理页专用，仅返回 id/名称/状态）
// 权限归属 k8s.permission.list，与授权查看对齐，不借用 k8s.cluster.list
func (r *ClusterRepository) FindClusterOptions() ([]map[string]interface{}, error) {
	var clusters []map[string]interface{}
	if err := r.db.Model(&modelk8s.K8sCluster{}).
		Select("id, name, status").
		Order("id ASC").
		Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

// ========== 原生 RBAC 绑定（B 模式） ==========

// FindNativeBindingsByCluster 查询集群的原生角色绑定（预加载用户与用户组）
func (r *ClusterRepository) FindNativeBindingsByCluster(clusterID uint) ([]modelk8s.K8sNativeRoleBinding, error) {
	var bindings []modelk8s.K8sNativeRoleBinding
	if err := r.db.Where("cluster_id = ?", clusterID).
		Preload("User").
		Preload("Group").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// FindAllNativeBindings 全局原生绑定列表（授权管理页）：clusterID=0 表示全部集群
func (r *ClusterRepository) FindAllNativeBindings(clusterID uint) ([]modelk8s.K8sNativeRoleBinding, error) {
	var bindings []modelk8s.K8sNativeRoleBinding
	query := r.db.
		Preload("User").
		Preload("Group").
		Preload("Cluster")
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if err := query.Order("cluster_id ASC, id ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// FindNativeBindingByID 根据ID查询原生角色绑定
func (r *ClusterRepository) FindNativeBindingByID(bindingID uint) (*modelk8s.K8sNativeRoleBinding, error) {
	var binding modelk8s.K8sNativeRoleBinding
	if err := r.db.First(&binding, bindingID).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// CreateNativeBinding 创建原生角色绑定记录
func (r *ClusterRepository) CreateNativeBinding(binding *modelk8s.K8sNativeRoleBinding) error {
	return r.db.Create(binding).Error
}

// UpdateNativeBindingK8sName 回写集群内 Binding 资源名
func (r *ClusterRepository) UpdateNativeBindingK8sName(binding *modelk8s.K8sNativeRoleBinding, k8sName string) error {
	return r.db.Model(binding).Update("k8s_binding_name", k8sName).Error
}

// DeleteNativeBinding 删除原生角色绑定记录
func (r *ClusterRepository) DeleteNativeBinding(bindingID uint) error {
	return r.db.Delete(&modelk8s.K8sNativeRoleBinding{}, bindingID).Error
}

// FindNativeImpersonation 查询用户在某集群的原生身份（B 模式执行层依据）
// 返回：模拟用户名（oneops-{username}）、需附加的模拟组（oneops-group-{code}...）、是否存在任何原生绑定
func (r *ClusterRepository) FindNativeImpersonation(userID uint, clusterID uint) (string, []string, bool, error) {
	var groups []string
	if err := r.db.Raw(`
		SELECT DISTINCT nb.impersonation_name FROM k8s_native_role_bindings nb
		JOIN sys_user_group_members m ON m.group_id = nb.group_id
		WHERE nb.subject_type = 'group' AND nb.cluster_id = ? AND m.user_id = ?`,
		clusterID, userID).Scan(&groups).Error; err != nil {
		return "", nil, false, err
	}

	var cnt int64
	if err := r.db.Raw(`
		SELECT COUNT(*) FROM k8s_native_role_bindings
		WHERE subject_type = 'user' AND user_id = ? AND cluster_id = ?`,
		userID, clusterID).Scan(&cnt).Error; err != nil {
		return "", nil, false, err
	}

	if cnt == 0 && len(groups) == 0 {
		return "", nil, false, nil
	}

	var username string
	if err := r.db.Raw(`SELECT username FROM sys_users WHERE id = ?`, userID).Scan(&username).Error; err != nil {
		return "", nil, false, err
	}
	if username == "" {
		return "", nil, false, fmt.Errorf("用户不存在: %d", userID)
	}

	return "oneops-" + username, groups, true, nil
}
