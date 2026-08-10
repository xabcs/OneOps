package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (c *CMDBController) GetServerGroups(ctx *gin.Context) {
	groups, err := c.svc.GetServerGroups()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// GetAssetTree 获取完整的资产树（分组+服务器），一次性返回所有数据
func (c *CMDBController) GetAssetTree(ctx *gin.Context) {
	data, err := c.svc.GetAssetTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(data))
}

// GetServerGroupByID 获取主机分组详情
func (c *CMDBController) GetServerGroupByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	group, err := c.svc.GetServerGroupByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("分组不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(group))
}

// CreateServerGroup 创建主机分组
func (c *CMDBController) CreateServerGroup(ctx *gin.Context) {
	var group modelcmdb.ServerGroup
	if err := ctx.ShouldBindJSON(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.CreateServerGroup(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组创建成功"))
}

// UpdateServerGroup 更新主机分组
func (c *CMDBController) UpdateServerGroup(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerGroup(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组更新成功"))
}

// DeleteServerGroup 删除主机分组
func (c *CMDBController) DeleteServerGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteServerGroup(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组删除成功"))
}

// AssignServerToGroup 将服务器分配到分组
func (c *CMDBController) AssignServerToGroup(ctx *gin.Context) {
	var req struct {
		ServerID uint `json:"serverId" binding:"required"`
		GroupID  uint `json:"groupId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.AssignServerToGroup(req.ServerID, req.GroupID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// AssignServerToGroups 将服务器分配到多个分组
func (c *CMDBController) AssignServerToGroups(ctx *gin.Context) {
	var req struct {
		ServerID uint   `json:"serverId" binding:"required"`
		GroupIDs []uint `json:"groupIds" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.AssignServerToGroups(req.ServerID, req.GroupIDs); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// GetServersByGroup 获取指定分组下的服务器列表
func (c *CMDBController) GetServersByGroup(ctx *gin.Context) {
	groupIDStr := ctx.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的分组ID"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	servers, err := c.svc.GetServersByGroup(uint(groupID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 手动分页
	total := int64(len(servers))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(servers) {
		servers = []modelcmdb.Server{}
	} else if end > len(servers) {
		servers = servers[start:]
	} else {
		servers = servers[start:end]
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  servers,
		"total": total,
	}))
}
