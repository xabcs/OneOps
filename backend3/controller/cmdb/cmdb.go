package cmdb

import (
	. "oneops/backend3/service/cmdb"
)

// CMDBController CMDB控制器
type CMDBController struct {
	svc      *CMDBService
	agentSvc *AgentService
}

// NewCMDBController 创建CMDB控制器
func NewCMDBController(svc *CMDBService, agentSvc *AgentService) *CMDBController {
	return &CMDBController{
		svc:      svc,
		agentSvc: agentSvc,
	}
}
