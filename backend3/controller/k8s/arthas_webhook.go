package k8s

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/utils"
	k8ssvc "oneops/backend3/service/k8s"
)

// ArthasWebhookController Arthas Webhook 自动注入管理
type ArthasWebhookController struct {
	svc *k8ssvc.ArthasWebhookService
}

// NewArthasWebhookController 构造
func NewArthasWebhookController(svc *k8ssvc.ArthasWebhookService) *ArthasWebhookController {
	return &ArthasWebhookController{svc: svc}
}

// parseClusterID 解析 clusterId（query 或 body）
func parseClusterID(c *gin.Context) (uint, bool) {
	// query 优先
	if v := c.Query("clusterId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil || id == 0 {
			c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的 clusterId"))
			return 0, false
		}
		return uint(id), true
	}
	return 0, true // 允许缺省（全局配置场景）
}

// GetStatus godoc
// @Summary      查询集群 Webhook 注入状态
// @Description  返回是否启用（MWC 存在）、tunnel WS 地址、init 镜像与触发规则
// @Tags         K8s-诊断
// @Produce      json
// @Param        clusterId query int true "集群 ID"
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/webhook [get]
// @Security     BearerAuth
func (ctrl *ArthasWebhookController) GetStatus(c *gin.Context) {
	clusterID, ok := parseClusterID(c)
	if !ok {
		return
	}
	status, err := ctrl.svc.GetWebhookStatus(clusterID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": status})
}

// webhookOpRequest 启用/禁用请求
type webhookOpRequest struct {
	ClusterID uint `json:"clusterId" binding:"required"`
}

// Enable godoc
// @Summary      启用集群 Webhook 自动注入
// @Description  在目标集群创建/更新 MutatingWebhookConfiguration（Pod label oneops-arthas-injection=enabled 触发）
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body body webhookOpRequest true "集群 ID" example({"clusterId":1})
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/webhook/enable [post]
// @Security     BearerAuth
func (ctrl *ArthasWebhookController) Enable(c *gin.Context) {
	var req webhookOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("clusterId 不能为空"))
		return
	}
	if err := ctrl.svc.EnableWebhook(req.ClusterID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已启用：为 Pod template 打 label oneops-arthas-injection=enabled 后滚动重启即接入"})
}

// Disable godoc
// @Summary      禁用集群 Webhook 自动注入
// @Description  删除 MutatingWebhookConfiguration（已注入的 Pod 不受影响，新 Pod 不再注入）
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body body webhookOpRequest true "集群 ID" example({"clusterId":1})
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/webhook/disable [post]
// @Security     BearerAuth
func (ctrl *ArthasWebhookController) Disable(c *gin.Context) {
	var req webhookOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("clusterId 不能为空"))
		return
	}
	if err := ctrl.svc.DisableWebhook(req.ClusterID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已禁用（存量已注入 Pod 不受影响）"})
}

// webhookConfigRequest 注入配置保存请求
type webhookConfigRequest struct {
	ClusterID  uint   `json:"clusterId"`  // 0 = 全局默认
	TunnelWS   string `json:"tunnelWS"`   // agent 反连地址（集群内），如 ws://arthas-tunnel.test:7777/ws
	InitImage  string `json:"initImage"`  // initContainer 镜像
	WebhookURL string `json:"webhookURL"` // apiserver 回调地址（全局）
	TunnelURL  string `json:"tunnelURL"`  // 平台 HTTP 探测地址，如 http://1.2.3.4:8080
	TunnelSess string `json:"tunnelSess"` // 平台 WS 会话地址，如 ws://1.2.3.4:7777
}

// SaveConfig godoc
// @Summary      保存 Webhook 注入配置
// @Description  回调地址/tunnel WS 地址/init 镜像（clusterId=0 为全局默认，非 0 为集群覆盖；回调地址仅全局）
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body body webhookConfigRequest true "配置" example({"clusterId":0,"webhookURL":"https://192.168.1.10:9443/webhook/arthas-inject","tunnelWS":"ws://arthas-tunnel.oneops:7777/ws","initImage":"registry.cn-hangzhou.aliyuncs.com/oneops/arthas-agent:3.7.2"})
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/webhook/config [post]
// @Security     BearerAuth
func (ctrl *ArthasWebhookController) SaveConfig(c *gin.Context) {
	var req webhookConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求格式错误"))
		return
	}
	if req.TunnelWS == "" && req.InitImage == "" && req.WebhookURL == "" && req.TunnelURL == "" && req.TunnelSess == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("至少配置一项"))
		return
	}
	if err := ctrl.svc.SaveWebhookConfig(req.ClusterID, req.TunnelWS, req.InitImage, req.WebhookURL, req.TunnelURL, req.TunnelSess); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "配置已保存（对后续新建 Pod 生效）"})
}
