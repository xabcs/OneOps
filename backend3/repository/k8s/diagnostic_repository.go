package k8s

import (
	modelk8s "oneops/backend3/model/k8s"

	"gorm.io/gorm"
)

// DiagnosticRepository 诊断数据访问层
type DiagnosticRepository struct {
	db *gorm.DB
}

// NewDiagnosticRepository 创建诊断仓库
func NewDiagnosticRepository(db *gorm.DB) *DiagnosticRepository {
	return &DiagnosticRepository{db: db}
}

// DiagnosticHistoryQuery 诊断历史分页查询条件
type DiagnosticHistoryQuery struct {
	ClusterID string
	Namespace string
	PodName   string
	Page      int
	PageSize  int
}

// Create 创建诊断历史记录
func (r *DiagnosticRepository) Create(history *modelk8s.DiagnosticHistory) error {
	return r.db.Create(history).Error
}

// FindHistoryWithPagination 分页查询诊断历史
func (r *DiagnosticRepository) FindHistoryWithPagination(q DiagnosticHistoryQuery) ([]modelk8s.DiagnosticHistory, int64, error) {
	var histories []modelk8s.DiagnosticHistory
	var total int64

	query := r.db.Model(&modelk8s.DiagnosticHistory{})

	if q.ClusterID != "" {
		query = query.Where("cluster_id = ?", q.ClusterID)
	}
	if q.Namespace != "" {
		query = query.Where("namespace = ?", q.Namespace)
	}
	if q.PodName != "" {
		query = query.Where("pod_name = ?", q.PodName)
	}

	query.Count(&total)

	offset := (q.Page - 1) * q.PageSize
	if err := query.Offset(offset).Limit(q.PageSize).Order("timestamp DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}
