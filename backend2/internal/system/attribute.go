package system

import (
	"net/http"
	"oneops/backend2/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AttributeController 属性定义控制器
type AttributeController struct {
	attributeService *AttributeService
}

// NewAttributeController 创建属性定义控制器
func NewAttributeController() *AttributeController {
	return &AttributeController{
		attributeService: NewAttributeService(),
	}
}

// GetAttributeDefinitions 获取属性定义列表
func (c *AttributeController) GetAttributeDefinitions(ctx *gin.Context) {
	category := ctx.Query("category")

	attributes, err := c.attributeService.GetAttributeDefinitions(category)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取属性列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(attributes))
}

// GetAttributeDefinitionByID 获取属性定义详情
func (c *AttributeController) GetAttributeDefinitionByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	attribute, err := c.attributeService.GetAttributeDefinitionByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("属性不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(attribute))
}

// CreateAttributeDefinition 创建属性定义
func (c *AttributeController) CreateAttributeDefinition(ctx *gin.Context) {
	var attr AttributeDefinition
	if err := ctx.ShouldBindJSON(&attr); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
		return
	}

	if err := c.attributeService.CreateAttributeDefinition(&attr); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性创建成功"))
}

// UpdateAttributeDefinition 更新属性定义
func (c *AttributeController) UpdateAttributeDefinition(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 检查管理员权限
	if !c.isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, utils.ErrorInternal("无权限"))
		return
	}

	if err := c.attributeService.UpdateAttributeDefinition(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性更新成功"))
}

// DeleteAttributeDefinition 删除属性定义
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

	if err := c.attributeService.DeleteAttributeDefinition(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("属性删除成功"))
}

// GetServerAttributes 获取主机的属性列表
func (c *AttributeController) GetServerAttributes(ctx *gin.Context) {
	serverIDStr := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	attributes, err := c.attributeService.GetServerAttributes(uint(serverID))
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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.attributeService.ValidateAttributeValue(req.AttributeID, req.Value); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("验证通过"))
}

// SaveServerAttributes 保存主机的属性列表
func (c *AttributeController) SaveServerAttributes(ctx *gin.Context) {
	serverIDStr := ctx.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	var attributes []ServerAttribute
	if err := ctx.ShouldBindJSON(&attributes); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.attributeService.SaveServerAttributes(uint(serverID), attributes); err != nil {
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
