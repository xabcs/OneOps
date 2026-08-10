package k8s

import (
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
	Filter   *modelk8s.K8sClusterFilter
	Page     int
	PageSize int
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

// FindRoleBindingsByCluster 查询集群的角色绑定（预加载用户和角色）
func (r *ClusterRepository) FindRoleBindingsByCluster(clusterID uint) ([]modelk8s.ClusterRoleBinding, error) {
	var bindings []modelk8s.ClusterRoleBinding
	if err := r.db.Where("cluster_id = ?", clusterID).
		Preload("Role").
		Preload("User").
		Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// FindRoleBinding 查询用户的集群角色绑定
func (r *ClusterRepository) FindRoleBinding(userID uint, clusterID uint) (*modelk8s.ClusterRoleBinding, error) {
	var binding modelk8s.ClusterRoleBinding
	if err := r.db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// CreateRoleBinding 创建集群角色绑定
func (r *ClusterRepository) CreateRoleBinding(binding *modelk8s.ClusterRoleBinding) error {
	return r.db.Create(binding).Error
}

// UpdateRoleBindingRole 更新角色绑定的角色ID
func (r *ClusterRepository) UpdateRoleBindingRole(binding *modelk8s.ClusterRoleBinding, roleID uint) error {
	return r.db.Model(binding).Update("role_id", roleID).Error
}

// DeleteRoleBinding 删除用户的集群角色绑定，返回受影响行数
func (r *ClusterRepository) DeleteRoleBinding(userID uint, clusterID uint) (int64, error) {
	result := r.db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).Delete(&modelk8s.ClusterRoleBinding{})
	return result.RowsAffected, result.Error
}

// CountRoleBindings 统计用户的集群角色绑定数量
func (r *ClusterRepository) CountRoleBindings(userID uint, clusterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelk8s.ClusterRoleBinding{}).
		Where("user_id = ? AND cluster_id = ?", userID, clusterID).
		Count(&count).Error
	return count, err
}

// FindUserByID 根据ID查询用户
func (r *ClusterRepository) FindUserByID(userID uint) (*modelsystem.User, error) {
	var user modelsystem.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindRoleByID 根据ID查询角色
func (r *ClusterRepository) FindRoleByID(roleID uint) (*modelsystem.Role, error) {
	var role modelsystem.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
