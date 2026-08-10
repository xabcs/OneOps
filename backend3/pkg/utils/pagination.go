package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

// ParsePagination 从 gin.Context 解析分页参数，统一使用 page/pageSize
func ParsePagination(c *gin.Context) PaginationParams {
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			pageSize = size
		}
	}

	if pageSize > 100 {
		pageSize = 100
	}

	return PaginationParams{Page: page, PageSize: pageSize}
}

// BuildOffset 计算数据库偏移量
func BuildOffset(p PaginationParams) int {
	return (p.Page - 1) * p.PageSize
}

// PaginatedData 统一分页响应结构
type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Pages    int         `json:"pages"`
}

// BuildPaginatedResponse 构建统一分页响应（放入 Response.Data）
func BuildPaginatedResponse(data interface{}, total int64, p PaginationParams) PaginatedData {
	pages := 0
	if p.PageSize > 0 {
		pages = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
	return PaginatedData{
		List:     data,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
		Pages:    pages,
	}
}
