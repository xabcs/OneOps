package k8s

import (
	"net/http"
	"strconv"

	modelk8s "oneops/backend3/model/k8s"
	repok8s "oneops/backend3/repository/k8s"
	k8ssvc "oneops/backend3/service/k8s"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// DiagnosticController 诊断中心 REST 控制器（Arthas Tunnel 架构）
type DiagnosticController struct {
	svc *k8ssvc.DiagnosticService
}

// NewDiagnosticController 创建诊断控制器
func NewDiagnosticController(svc *k8ssvc.DiagnosticService) *DiagnosticController {
	return &DiagnosticController{svc: svc}
}

// GetCommands godoc
// @Summary      获取 Arthas 命令目录
// @Description  返回全量命令目录（内置默认 + 管理员风险覆盖），含风险分级与流式标记
// @Tags         K8s-诊断
// @Produce      json
// @Success      200 {object} utils.Response{data=[]k8ssvc.CommandCatalogItem}
// @Router       /k8s/diagnostic/commands [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetCommands(c *gin.Context) {
	catalog := ctrl.svc.GetCommandCatalog()
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": catalog})
}

// ListApps godoc
// @Summary      获取已接入诊断的应用列表
// @Description  从 tunnel-server 同步 agent 注册表并按应用分组返回
// @Tags         K8s-诊断
// @Produce      json
// @Success      200 {object} utils.Response{data=[]k8ssvc.AppSummary}
// @Failure      200 {object} utils.Response "内部错误"
// @Router       /k8s/diagnostic/apps [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) ListApps(c *gin.Context) {
	apps, err := ctrl.svc.ListApps()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	if apps == nil {
		apps = []k8ssvc.AppSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": apps})
}

// ListAppAgents godoc
// @Summary      获取应用实例列表
// @Description  返回指定应用下的 agent 实例（Pod）列表及在线状态
// @Tags         K8s-诊断
// @Produce      json
// @Param        appName path string true "应用名"
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/apps/{appName}/agents [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) ListAppAgents(c *gin.Context) {
	appName := c.Param("appName")
	agents, err := ctrl.svc.ListAppAgents(appName)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": agents})
}

// ExecuteOneShot godoc
// @Summary      执行一次性诊断命令
// @Description  经 tunnel 对指定 agent 执行单条 Arthas 命令（适用于只读命令；流式命令请使用专家终端）
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body body object true "执行请求" example({"agentId":"pod-x-default","command":"thread -n 5","timeout":60})
// @Success      200 {object} utils.Response
// @Failure      200 {object} utils.Response "命令被禁用 / agent 离线 / 执行失败"
// @Router       /k8s/diagnostic/execute [post]
// @Security     BearerAuth
func (ctrl *DiagnosticController) ExecuteOneShot(c *gin.Context) {
	var request struct {
		ClusterID string `json:"clusterId"`
		AgentID   string `json:"agentId" binding:"required"`
		Command   string `json:"command" binding:"required"`
		Timeout   int    `json:"timeout"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}
	username := c.GetString("username")
	if username == "" {
		username = "system"
	}

	result, err := ctrl.svc.ExecuteOneShot(&k8ssvc.OneShotRequest{
		ClusterID: request.ClusterID,
		AppName:   "",
		AgentID:   request.AgentID,
		Command:   request.Command,
		Timeout:   request.Timeout,
		UserID:    userID,
		Username:  username,
	})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if result.Status == "error" {
		c.JSON(http.StatusOK, gin.H{
			"code": 500, "message": "诊断执行失败: " + result.Error, "data": result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

// GetExecutions godoc
// @Summary      查询诊断执行历史
// @Description  分页查询执行记录，支持按应用/agent/操作者/命令/风险级筛选
// @Tags         K8s-诊断
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        pageSize  query int    false "每页数量" default(20)
// @Param        appName   query string false "应用名"
// @Param        agentId   query string false "agent 标识"
// @Param        username  query string false "操作者"
// @Param        command   query string false "命令关键字"
// @Param        riskLevel query string false "风险级 L0-L5"
// @Success      200 {object} utils.Response{data=dto.PageResult}
// @Router       /k8s/diagnostic/executions [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetExecutions(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	query := repok8s.ExecutionQuery{
		ClusterID: c.Query("clusterId"),
		AppName:   c.Query("appName"),
		AgentID:   c.Query("agentId"),
		Username:  c.Query("username"),
		Command:   c.Query("command"),
		RiskLevel: c.Query("riskLevel"),
		Page:      params.GetPage(),
		PageSize:  params.GetPageSize(),
	}

	list, total, err := ctrl.svc.GetExecutions(query)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(list, total, params)))
}

// GetOverview godoc
// @Summary      诊断中心总览
// @Description  接入应用数/agent 在线率/执行统计/高危操作趋势数据源
// @Tags         K8s-诊断
// @Produce      json
// @Success      200 {object} utils.Response{data=k8ssvc.OverviewData}
// @Router       /k8s/diagnostic/overview [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetOverview(c *gin.Context) {
	data, err := ctrl.svc.GetOverview()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

// GetSessionDetail godoc
// @Summary      会话详情（含 I/O 回放数据）
// @Tags         K8s-诊断
// @Param        sessionId path int true "会话 ID"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/sessions/{sessionId} [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetSessionDetail(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}
	session, err := ctrl.svc.GetSessionByID(uint(sessionID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("会话不存在"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": session})
}

// GetCommandOverrides godoc
// @Summary      获取命令风险覆盖表
// @Description  管理员维护的命令分级覆盖与拉黑列表
// @Tags         K8s-诊断
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/command-overrides [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetCommandOverrides(c *gin.Context) {
	overrides, err := ctrl.svc.GetCommandOverrides()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	if overrides == nil {
		overrides = []modelk8s.DiagnosticCommandOverride{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": overrides})
}

// SaveCommandOverride godoc
// @Summary      新增/更新命令风险覆盖
// @Description  设置命令风险级（L0-L5）或拉黑（disabled）；保存后命令目录即时生效
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body body k8ssvc.CommandOverrideInput true "覆盖内容" example({"command":"ognl","riskLevel":"L5","description":"任意表达式执行"})
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/command-overrides [post]
// @Security     BearerAuth
func (ctrl *DiagnosticController) SaveCommandOverride(c *gin.Context) {
	var input k8ssvc.CommandOverrideInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	input.UpdatedBy = c.GetString("username")

	if err := ctrl.svc.SaveCommandOverride(&input); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": ctrl.svc.GetCommandCatalog()})
}

// DeleteCommandOverride godoc
// @Summary      删除命令风险覆盖
// @Description  删除后该命令恢复内置默认风险分级
// @Tags         K8s-诊断
// @Param        id path int true "覆盖记录 ID"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/command-overrides/{id} [delete]
// @Security     BearerAuth
func (ctrl *DiagnosticController) DeleteCommandOverride(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的记录ID"))
		return
	}
	if err := ctrl.svc.DeleteCommandOverrideById(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": ctrl.svc.GetCommandCatalog()})
}
