package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ═══════════════════════════════════════════
// 统一分页基类
// ═══════════════════════════════════════════

// BasePageQuery 所有分页查询的公共基类
type BasePageQuery struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" json:"pageSize" binding:"omitempty,min=1,max=100"`
}

// GetPage 获取页码（默认1）
func (q *BasePageQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize 获取每页条数（默认10，最大100）
func (q *BasePageQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 10
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// GetOffset 计算 SQL 偏移量
func (q *BasePageQuery) GetOffset() int {
	return (q.GetPage() - 1) * q.GetPageSize()
}

// ═══════════════════════════════════════════
// 统一分页结果
// ═══════════════════════════════════════════

// PageResult 统一分页响应结构
type PageResult struct {
	List      interface{} `json:"list"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"pageSize"`
	Pages     int64       `json:"pages"`     // 新规范
	PageCount int64       `json:"pageCount"` // 兼容前端旧字段
}

// NewPageResult 构造分页结果
func NewPageResult(list interface{}, total int64, q BasePageQuery) PageResult {
	ps := int64(q.GetPageSize())
	pages := (total + ps - 1) / ps
	if pages < 0 {
		pages = 0
	}
	return PageResult{
		List:      list,
		Total:     total,
		Page:      q.GetPage(),
		PageSize:  q.GetPageSize(),
		Pages:     pages,
		PageCount: pages,
	}
}

// ═══════════════════════════════════════════
// 统一参数校验错误格式化
// ═══════════════════════════════════════════

// fieldNames 字段名中文映射
var fieldNames = map[string]string{
	"Username":  "用户名",
	"Password":  "密码",
	"Email":     "邮箱",
	"RealName":  "姓名",
	"Phone":     "手机号",
	"Hostname":  "主机名",
	"IP":        "IP地址",
	"InnerIP":   "内网IP",
	"Name":      "名称",
	"Code":      "编码",
	"Path":      "路径",
	"Namespace": "命名空间",
	"Description": "描述",
	"Status":    "状态",
	"Permission": "权限",
	"Icon":      "图标",
	"Sort":      "排序",
	"MenuType":  "菜单类型",
	"ParentID":  "父级ID",
	"SSHPort":   "SSH端口",
	"Env":       "环境",
	"Provider":  "供应商",
	"OS":        "操作系统",
	"Arch":      "架构",
	"CPU":       "CPU",
	"Memory":    "内存",
	"Disk":      "磁盘",
	"Remarks":   "备注",
	"RoleIDs":   "角色",
	"MenuIDs":   "菜单",
	"GroupID":   "分组",
	"TagID":     "标签",
	"CredentialID": "凭证",
	"LabelSelector": "标签选择器",
}

// FormatValidationError 将 validator 错误转为中文提示
func FormatValidationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			if cn, ok := fieldNames[field]; ok {
				field = cn
			}
			switch e.Tag() {
			case "required":
				return fmt.Sprintf("%s不能为空", field)
			case "min":
				return fmt.Sprintf("%s不能小于%s", field, e.Param())
			case "max":
				return fmt.Sprintf("%s不能大于%s", field, e.Param())
			case "email":
				return fmt.Sprintf("%s格式不正确", field)
			case "ip":
				return fmt.Sprintf("%s不是有效的IP地址", field)
			case "oneof":
				return fmt.Sprintf("%s值不合法", field)
			case "alphanum":
				return fmt.Sprintf("%s只能包含字母和数字", field)
			case "alphanumunderscore":
				return fmt.Sprintf("%s只能包含字母、数字和下划线", field)
			case "hostname":
				return fmt.Sprintf("%s格式不正确", field)
			case "url":
				return fmt.Sprintf("%s不是有效的URL", field)
			case "uri":
				return fmt.Sprintf("%s不是有效的URI", field)
			default:
				return fmt.Sprintf("%s校验失败(%s)", field, e.Tag())
			}
		}
	}
	return "请求参数错误"
}
