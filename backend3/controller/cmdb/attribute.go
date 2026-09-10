package cmdb

import (
	"encoding/json"
	"net/http"
	"strconv"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	syssvc "oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
)

// AttributeController 属性定义控制器
type AttributeController struct {
	svc *AttributeService
}

// NewAttributeController 创建属性定义控制器
func NewAttributeController(svc *AttributeService) *AttributeController {
	return &AttributeController{
		svc: svc,
	}
}

// GetAttributeDefinitions godoc
// @Summary      获取属性定义列表
// @Description  获取属性定义列表，支持按分类筛选
// @Tags         CMDB-属性
// @Produce      json
// @Param        category  query  string  false  "属性分类"
// @Success      200  {object}  utils.Response  "属性定义列表"
// @Failure      200  {object}  utils.Response  "获取属性列表失败"
// @Router       /cmdb/attributes [get]
// @Security     BearerAuth
func (c *AttributeController) GetAttributeDefinitions(ctx *gin.Context) {
	category := ctx.Query("category")

	attributes, err := c.svc.GetAttributeDefinitions(category)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取属性列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(attributes))
}

// GetAttributeDefinitionByID godoc
// @Summary      获取属性定义详情
// @Description  根据属性 ID 获取属性定义详情
// @Tags         CMDB-属性
// @Produce      json
// @Param        id  path  int  true  "属性 ID"
// @Success      200  {object}  utils.Response{data=modelsystem.AttributeDefinition}
// @Failure      200  {object}  utils.Response  "无效的 ID / 属性不存在"
// @Router       /cmdb/attributes/{id} [get]
// @Security     BearerAuth
func (c *AttributeController) GetAttributeDefinitionByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	attribute, err := c.svc.GetAttributeDefinitionByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("属性不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(attribute))
}

// CreateAttributeDefinition godoc
// @Summary      创建属性定义
// @Description  创建一个新的属性定义（需要管理员权限）
// @Tags         CMDB-属性
// @Accept       json
// @Produce      json
// @Param        attr  body      modelsystem.AttributeDefinition  true  "属性定义信息"
// @Success      200   {object}  utils.Response  "属性创建成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 无权限 / 创建失败"
// @Router       /cmdb/attributes [post]
// @Security     BearerAuth
func (c *AttributeController) CreateAttributeDefinition(ctx *gin.Context) {
	var attr modelsystem.AttributeDefinition
	if err := ctx.ShouldBindJSON(&attr); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
		return
	}

	if err := c.svc.CreateAttributeDefinition(&attr); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性创建成功"))
}

// UpdateAttributeDefinitionRequest 属性定义更新请求。
// 指针字段：nil 表示未提交该字段、不参与更新——天然实现部分更新语义与字段白名单（防止 mass assignment），
// 同时可挂 binding 校验。字段名与模型 JSON 契约一致（驼峰）
type UpdateAttributeDefinitionRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=100"`
	Key          *string `json:"key" binding:"omitempty,min=1,max=50"`
	Category     *string `json:"category" binding:"omitempty,min=1,max=50"`
	Type         *string `json:"type" binding:"omitempty,min=1,max=30"`
	Options      *string `json:"options"`
	Required     *bool   `json:"required"`
	DefaultValue *string `json:"defaultValue"`
	SortOrder    *int    `json:"sortOrder"`
	Status       *int    `json:"status" binding:"omitempty,min=0,max=1"`
	Description  *string `json:"description"`
}

// UpdateAttributeDefinition godoc
// @Summary      更新属性定义
// @Description  根据属性 ID 更新属性定义信息（部分字段更新，需要管理员权限）
// @Tags         CMDB-属性
// @Accept       json
// @Produce      json
// @Param        id    path      int                                true  "属性 ID"
// @Param        attr  body      UpdateAttributeDefinitionRequest   true  "需要更新的字段（仅提交要改的字段）"
// @Success      200   {object}  utils.Response  "属性更新成功"
// @Failure      200   {object}  utils.Response  "无效的 ID / 请求参数错误 / 无权限 / 更新失败"
// @Router       /cmdb/attributes/{id} [put]
// @Security     BearerAuth
func (c *AttributeController) UpdateAttributeDefinition(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var req UpdateAttributeDefinitionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
		return
	}

	// 仅非 nil 字段进入更新集（部分更新），列名统一为蛇形；attr_key 为实际列名（原 key 为 MySQL 保留字，已迁移改名）
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Key != nil {
		updates["attr_key"] = *req.Key
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Options != nil {
		updates["options"] = *req.Options
	}
	if req.Required != nil {
		updates["required"] = *req.Required
	}
	if req.DefaultValue != nil {
		updates["default_value"] = *req.DefaultValue
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("没有可更新的字段"))
		return
	}

	if err := c.svc.UpdateAttributeDefinition(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性更新成功"))
}

// DeleteAttributeDefinition godoc
// @Summary      删除属性定义
// @Description  根据属性 ID 删除属性定义（需要管理员权限）
// @Tags         CMDB-属性
// @Produce      json
// @Param        id  path  int  true  "属性 ID"
// @Success      200  {object}  utils.Response  "属性删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 无权限 / 删除失败"
// @Router       /cmdb/attributes/{id} [delete]
// @Security     BearerAuth
func (c *AttributeController) DeleteAttributeDefinition(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
		return
	}

	if err := c.svc.DeleteAttributeDefinition(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性删除成功"))
}

// GetServerAttributes godoc
// @Summary      获取主机属性列表
// @Description  根据主机 ID 获取主机的属性列表
// @Tags         CMDB-属性
// @Produce      json
// @Param        serverId  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response  "主机属性列表"
// @Failure      200  {object}  utils.Response  "无效的主机 ID / 获取主机属性失败"
// @Router       /cmdb/server-attributes/{serverId} [get]
// @Security     BearerAuth
func (c *AttributeController) GetServerAttributes(ctx *gin.Context) {
	serverIDStr := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	attributes, err := c.svc.GetServerAttributes(uint(serverID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取主机属性失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(attributes))
}

// ValidateServerAttribute godoc
// @Summary      验证主机属性值
// @Description  验证给定属性 ID 的值是否符合属性定义的规则
// @Tags         CMDB-属性
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "验证请求"
// @Success      200   {object}  utils.Response  "验证通过"
// @Failure      200   {object}  utils.Response  "验证失败"
// @Router       /cmdb/attributes/validate [post]
// @Security     BearerAuth
func (c *AttributeController) ValidateServerAttribute(ctx *gin.Context) {
	var req struct {
		AttributeID uint   `json:"attributeId" binding:"required"`
		Value       string `json:"value" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.ValidateAttributeValue(req.AttributeID, req.Value); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("验证通过"))
}

// SaveServerAttributes godoc
// @Summary      保存主机属性
// @Description  保存指定主机的属性列表（覆盖更新）
// @Tags         CMDB-属性
// @Accept       json
// @Produce      json
// @Param        serverId    path      int                            true  "主机 ID"
// @Param        attributes  body      []modelcmdb.ServerAttribute    true  "属性列表"
// @Success      200         {object}  utils.Response  "属性保存成功"
// @Failure      200         {object}  utils.Response  "无效的主机 ID / 请求参数错误 / 保存失败"
// @Router       /cmdb/server-attributes/{serverId} [post]
// @Security     BearerAuth
func (c *AttributeController) SaveServerAttributes(ctx *gin.Context) {
	serverIDStr := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	var attributes []modelcmdb.ServerAttribute
	if err := ctx.ShouldBindJSON(&attributes); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.SaveServerAttributes(uint(serverID), attributes); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性保存成功"))
}

// isAdmin 检查当前用户是否拥有管理员权限（通过统一的权限服务判断）
func (c *AttributeController) isAdmin(ctx *gin.Context) bool {
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return false
	}
	permSvc, err := syssvc.GetPermissionService()
	if err != nil {
		return false
	}
	allowed, err := permSvc.HasPermission(userID, "cmdb.attribute.update")
	if err != nil {
		return false
	}
	return allowed
}

// GetServersByAttributes godoc
// @Summary      按属性筛选主机
// @Description  根据属性键值对筛选主机列表
// @Tags         CMDB-属性
// @Produce      json
// @Param        filters   query  string  false  "属性筛选条件（JSON 格式，如 {\"env\":\"prod\"}）"
// @Param        page      query  int     false  "页码"  default(1)
// @Param        pageSize  query  int     false  "每页数量"  default(20)
// @Success      200  {object}  utils.Response  "主机列表"
// @Router       /cmdb/servers-by-attributes [get]
// @Security     BearerAuth
func (c *AttributeController) GetServersByAttributes(ctx *gin.Context) {
	filtersStr := ctx.Query("filters")
	if filtersStr == "" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("filters 参数不能为空"))
		return
	}

	filters := make(map[string]string)
	if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("filters 参数格式错误"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	servers, total, err := c.svc.GetServersByAttributes(filters, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("按属性筛选主机失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(map[string]interface{}{
		"list":     servers,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}
