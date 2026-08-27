package controllerticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	serviceticket "oneops/backend3/service/ticket"
)

// WorkflowController 工单流程/类型控制器
type WorkflowController struct {
	svc *serviceticket.WorkflowService
}

func NewWorkflowController(svc *serviceticket.WorkflowService) *WorkflowController {
	return &WorkflowController{svc: svc}
}

// parseID 解析路径参数 ID
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return 0, false
	}
	return uint(id), true
}

// parsePage 解析分页参数
func parsePage(c *gin.Context) (dto.BasePageQuery, bool) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return params, false
	}
	return params, true
}

// ────────────────────────── 流程定义 ──────────────────────────

// GetWorkflows godoc
// @Summary      获取流程定义列表
// @Tags         工单-流程定义
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows [get]
// @Security     BearerAuth
func (ctrl *WorkflowController) GetWorkflows(c *gin.Context) {
	params, ok := parsePage(c)
	if !ok {
		return
	}
	keyword := c.Query("keyword")
	status := -1
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = v
		}
	}
	var typeID uint
	if s := c.Query("typeId"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 32); err == nil {
			typeID = uint(v)
		}
	}
	list, total, err := ctrl.svc.GetWorkflows(keyword, status, typeID, params.GetOffset(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取流程列表失败"))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(list, total, params)))
}

// GetWorkflowsByType godoc
// @Summary      获取某场景下启用的流程（发起工单用，仅需登录）
// @Tags         工单-流程定义
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows/by-type/{typeId} [get]
// @Security     BearerAuth
func (ctrl *WorkflowController) GetWorkflowsByType(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("typeId"), 10, 32)
	if err != nil || typeID == 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的场景ID"))
		return
	}
	list, err := ctrl.svc.GetEnabledWorkflowsByType(uint(typeID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取流程失败"))
		return
	}
	// 精简字段
	items := make([]gin.H, 0, len(list))
	for _, wf := range list {
		items = append(items, gin.H{
			"id": wf.ID, "name": wf.Name, "code": wf.Code,
			"description": wf.Description, "nodeCount": len(wf.Nodes),
		})
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(items))
}

// GetWorkflow godoc
// @Summary      获取流程详情（含节点）
// @Tags         工单-流程定义
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows/{id} [get]
// @Security     BearerAuth
func (ctrl *WorkflowController) GetWorkflow(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	wf, err := ctrl.svc.GetWorkflow(id)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("流程不存在"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(wf))
}

// CreateWorkflow godoc
// @Summary      创建流程定义
// @Tags         工单-流程定义
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows [post]
// @Security     BearerAuth
func (ctrl *WorkflowController) CreateWorkflow(c *gin.Context) {
	var req serviceticket.WorkflowSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	if err := ctrl.svc.CreateWorkflow(req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建成功"))
}

// UpdateWorkflow godoc
// @Summary      更新流程定义
// @Tags         工单-流程定义
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows/{id} [put]
// @Security     BearerAuth
func (ctrl *WorkflowController) UpdateWorkflow(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.WorkflowSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	if err := ctrl.svc.UpdateWorkflow(id, req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteWorkflow godoc
// @Summary      删除流程定义
// @Tags         工单-流程定义
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/workflows/{id} [delete]
// @Security     BearerAuth
func (ctrl *WorkflowController) DeleteWorkflow(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := ctrl.svc.DeleteWorkflow(id); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// ────────────────────────── 工单类型 ──────────────────────────

// GetTypes godoc
// @Summary      获取工单类型列表
// @Tags         工单-类型管理
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/types [get]
// @Security     BearerAuth
func (ctrl *WorkflowController) GetTypes(c *gin.Context) {
	params, ok := parsePage(c)
	if !ok {
		return
	}
	keyword := c.Query("keyword")
	status := -1
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = v
		}
	}
	list, total, err := ctrl.svc.GetTypes(keyword, status, params.GetOffset(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取工单类型列表失败"))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(list, total, params)))
}

// GetTypeOptions godoc
// @Summary      获取启用的工单类型选项（发起工单用，仅需登录）
// @Tags         工单-类型管理
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/types/options [get]
// @Security     BearerAuth
func (ctrl *WorkflowController) GetTypeOptions(c *gin.Context) {
	list, err := ctrl.svc.GetEnabledTypes()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取工单类型失败"))
		return
	}
	items := make([]gin.H, 0, len(list))
	for _, t := range list {
		items = append(items, gin.H{
			"id": t.ID, "name": t.Name, "code": t.Code, "icon": t.Icon,
			"description": t.Description, "formSchema": t.FormSchema,
		})
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(items))
}

// CreateType godoc
// @Summary      创建工单类型
// @Tags         工单-类型管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/types [post]
// @Security     BearerAuth
func (ctrl *WorkflowController) CreateType(c *gin.Context) {
	var req serviceticket.TypeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	if err := ctrl.svc.CreateType(req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建成功"))
}

// UpdateType godoc
// @Summary      更新工单类型
// @Tags         工单-类型管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/types/{id} [put]
// @Security     BearerAuth
func (ctrl *WorkflowController) UpdateType(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TypeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	if err := ctrl.svc.UpdateType(id, req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteType godoc
// @Summary      删除工单类型
// @Tags         工单-类型管理
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/types/{id} [delete]
// @Security     BearerAuth
func (ctrl *WorkflowController) DeleteType(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := ctrl.svc.DeleteType(id); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
