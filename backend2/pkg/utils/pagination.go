package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
}

// ParsePagination 解析分页参数
func ParsePagination(c *gin.Context) PaginationParams {
	page := 1
	pageSize := 10

	// 解析页码
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if currentStr := c.Query("current"); currentStr != "" {
		if p, err := strconv.Atoi(currentStr); err == nil && p > 0 {
			page = p
		}
	}

	// 解析每页大小
	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			pageSize = size
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			pageSize = size
		}
	}

	// 限制每页最大数量
	if pageSize > 100 {
		pageSize = 100
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// BuildOffset 计算偏移量
func BuildOffset(p PaginationParams) int {
	return (p.Page - 1) * p.PageSize
}

// BuildPaginatedResponse 构建分页响应
func BuildPaginatedResponse(data interface{}, total int64, p PaginationParams) map[string]interface{} {
	return map[string]interface{}{
		"records": data,
		"current": p.Page,
		"size":    p.PageSize,
		"total":   total,
		"pages":   (total + int64(p.PageSize) - 1) / int64(p.PageSize),
	}
}

// BuildSearchResponse 构建搜索响应（兼容不同命名）
func BuildSearchResponse(data interface{}, total int64, p PaginationParams) map[string]interface{} {
	return map[string]interface{}{
		"list":      data,
		"total":     total,
		"page":      p.Page,
		"pageSize":  p.PageSize,
		"pageCount": (total + int64(p.PageSize) - 1) / int64(p.PageSize),
	}
}
