package utils

import (
	"fmt"

	"gorm.io/gorm"
)

// QueryBuilder 查询构建器
// 提供链式API来构建数据库查询
type QueryBuilder struct {
	db        *gorm.DB
	query     *gorm.DB
	whereArgs []interface{}
	orderBy   []string
	page      int
	pageSize  int
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		query: db,
	}
}

// Where 添加WHERE条件
// 支持多种格式：
//   - Where("name = ?", "John")
//   - Where("name LIKE ?", "%John%")
//   - Where("age > ?", 18)
//   - Where("status IN ?", []string{"active", "pending"})
func (qb *QueryBuilder) Where(query string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Where(query, args...)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

// WhereOr 添加OR条件
func (qb *QueryBuilder) WhereOr(query string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Or(query, args...)
	return qb
}

// WhereEqual 添加等于条件（便捷方法）
func (qb *QueryBuilder) WhereEqual(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s = ?", field), value)
}

// WhereNotEqual 添加不等于条件
func (qb *QueryBuilder) WhereNotEqual(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s != ?", field), value)
}

// WhereLike 添加LIKE条件（自动添加%通配符）
//   - WhereLike("name", "John") => name LIKE '%John%'
func (qb *QueryBuilder) WhereLike(field string, value string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
}

// WhereLikeLeft 左模糊匹配
//   - WhereLikeLeft("name", "John") => name LIKE 'John%'
func (qb *QueryBuilder) WhereLikeLeft(field string, value string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s LIKE ?", field), value+"%")
}

// WhereLikeRight 右模糊匹配
//   - WhereLikeRight("name", "John") => name LIKE '%John'
func (qb *QueryBuilder) WhereLikeRight(field string, value string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value)
}

// WhereIn 添加IN条件
func (qb *QueryBuilder) WhereIn(field string, values interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IN ?", field), values)
}

// WhereNotIn 添加NOT IN条件
func (qb *QueryBuilder) WhereNotIn(field string, values interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s NOT IN ?", field), values)
}

// WhereGreaterThan 添加大于条件
func (qb *QueryBuilder) WhereGreaterThan(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s > ?", field), value)
}

// WhereGreaterThanOrEqual 添加大于等于条件
func (qb *QueryBuilder) WhereGreaterThanOrEqual(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s >= ?", field), value)
}

// WhereLessThan 添加小于条件
func (qb *QueryBuilder) WhereLessThan(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s < ?", field), value)
}

// WhereLessThanOrEqual 添加小于等于条件
func (qb *QueryBuilder) WhereLessThanOrEqual(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s <= ?", field), value)
}

// WhereBetween 添加BETWEEN条件
func (qb *QueryBuilder) WhereBetween(field string, min, max interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), min, max)
}

// WhereNull 添加IS NULL条件
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NULL", field))
}

// WhereNotNull 添加IS NOT NULL条件
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NOT NULL", field))
}

// OrderBy 添加排序
//   - OrderBy("created_at DESC")
//   - OrderBy("name ASC")
func (qb *QueryBuilder) OrderBy(order string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

// OrderByDesc 按字段降序排序（便捷方法）
func (qb *QueryBuilder) OrderByDesc(field string) *QueryBuilder {
	return qb.OrderBy(fmt.Sprintf("%s DESC", field))
}

// OrderByAsc 按字段升序排序（便捷方法）
func (qb *QueryBuilder) OrderByAsc(field string) *QueryBuilder {
	return qb.OrderBy(fmt.Sprintf("%s ASC", field))
}

// Paginate 设置分页参数
func (qb *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
	qb.page = page
	qb.pageSize = pageSize
	return qb
}

// BuildQuery 构建最终的GORM查询（内部方法）
func (qb *QueryBuilder) buildQuery() *gorm.DB {
	query := qb.query

	// 应用排序
	if len(qb.orderBy) > 0 {
		for _, order := range qb.orderBy {
			query = query.Order(order)
		}
	}

	// 应用分页
	if qb.page > 0 && qb.pageSize > 0 {
		offset := (qb.page - 1) * qb.pageSize
		query = query.Offset(offset).Limit(qb.pageSize)
	}

	return query
}

// Count 获取记录总数（不受分页影响）
func (qb *QueryBuilder) Count() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

// First 获取第一条记录
func (qb *QueryBuilder) First(dest interface{}) error {
	return qb.buildQuery().First(dest).Error
}

// Last 获取最后一条记录
func (qb *QueryBuilder) Last(dest interface{}) error {
	return qb.buildQuery().Last(dest).Error
}

// Find 查找多条记录
func (qb *QueryBuilder) Find(dest interface{}) error {
	return qb.buildQuery().Find(dest).Error
}

// FindWithCount 查找记录并返回总数（用于分页）
func (qb *QueryBuilder) FindWithCount(dest interface{}) (int64, error) {
	// 先获取总数
	total, err := qb.Count()
	if err != nil {
		return 0, err
	}

	// 再查找记录
	if err := qb.Find(dest); err != nil {
		return 0, err
	}

	return total, nil
}

// Pluck 查询单个字段的值到数组
func (qb *QueryBuilder) Pluck(field string, dest interface{}) error {
	return qb.buildQuery().Pluck(field, dest).Error
}

// Select 指定查询字段
func (qb *QueryBuilder) Select(query string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Select(query, args...)
	return qb
}

// GroupBy 添加分组
func (qb *QueryBuilder) GroupBy(field string) *QueryBuilder {
	qb.query = qb.query.Group(field)
	return qb
}

// Having 添加HAVING条件
func (qb *QueryBuilder) Having(query string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Having(query, args...)
	return qb
}

// Joins 添加关联查询
func (qb *QueryBuilder) Joins(query string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Joins(query, args...)
	return qb
}

// Preload 预加载关联数据
func (qb *QueryBuilder) Preload(field string) *QueryBuilder {
	qb.query = qb.query.Preload(field)
	return qb
}

// Reset 重置查询构建器（保留db连接）
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.query = qb.db
	qb.whereArgs = nil
	qb.orderBy = nil
	qb.page = 0
	qb.pageSize = 0
	return qb
}

// GetQuery 获取原始GORM查询对象（高级用法）
func (qb *QueryBuilder) GetQuery() *gorm.DB {
	return qb.buildQuery()
}
