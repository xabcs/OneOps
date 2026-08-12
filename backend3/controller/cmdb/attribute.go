package cmdb

import (
	"net/http"
	"strconv"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

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
// @Router       /system/attributes [get]
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
// @Router       /system/attributes/{id} [get]
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
// @Router       /system/attributes [post]
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

// UpdateAttributeDefinition godoc
// @Summary      更新属性定义
// @Description  根据属性 ID 更新属性定义信息（部分字段更新，需要管理员权限）
// @Tags         CMDB-属性
// @Accept       json
// @Produce      json
// @Param        id    path      int                            true  "属性 ID"
// @Param        attr  body      modelsystem.AttributeDefinition  true  "需要更新的字段"
// @Success      200   {object}  utils.Response  "属性更新成功"
// @Failure      200   {object}  utils.Response  "无效的 ID / 请求参数错误 / 无权限 / 更新失败"
// @Router       /system/attributes/{id} [put]
// @Security     BearerAuth
func (c *AttributeController) UpdateAttributeDefinition(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
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
// @Router       /system/attributes/{id} [delete]
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
// @Router       /system/server-attributes/{serverId} [get]
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

// ValidateServerAttribute 验证主机属性值
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
// @Router       /system/server-attributes/{serverId} [post]
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

// isAdmin 检查是否是管理员
func (c *AttributeController) isAdmin(ctx *gin.Context) bool {
	// 从上下文获取用户信息
	user := ctx.GetString("username")
	// TODO: 实现真正的权限检查
	// 现在简单处理：admin用户是管理员
	return user == "admin"
}
