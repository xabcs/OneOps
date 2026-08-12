package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 主机分组管理 ==========

// GetServerGroups godoc
// @Summary      获取主机分组列表
// @Description  获取主机分组列表（树形结构）
// @Tags         CMDB-分组
// @Produce      json
// @Success      200  {object}  utils.Response  "分组树形列表"
// @Failure      200  {object}  utils.Response  "获取分组列表失败"
// @Router       /cmdb/groups [get]
// @Security     BearerAuth
func (c *CMDBController) GetServerGroups(ctx *gin.Context) {
	groups, err := c.svc.GetServerGroups()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// GetAssetTree godoc
// @Summary      获取资产树
// @Description  获取完整的资产树（分组 + 服务器），一次性返回所有数据
// @Tags         CMDB-分组
// @Produce      json
// @Success      200  {object}  utils.Response  "资产树"
// @Failure      200  {object}  utils.Response  "获取资产树失败"
// @Router       /cmdb/asset-tree [get]
// @Security     BearerAuth
func (c *CMDBController) GetAssetTree(ctx *gin.Context) {
	data, err := c.svc.GetAssetTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(data))
}

// GetServerGroupByID godoc
// @Summary      获取主机分组详情
// @Description  根据分组 ID 获取主机分组详情
// @Tags         CMDB-分组
// @Produce      json
// @Param        id  path  int  true  "分组 ID"
// @Success      200  {object}  utils.Response{data=modelcmdb.ServerGroup}
// @Failure      200  {object}  utils.Response  "无效的 ID / 分组不存在"
// @Router       /cmdb/groups/{id} [get]
// @Security     BearerAuth
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

// CreateServerGroup godoc
// @Summary      创建主机分组
// @Description  创建一个新的主机分组
// @Tags         CMDB-分组
// @Accept       json
// @Produce      json
// @Param        group  body      modelcmdb.ServerGroup  true  "分组信息"
// @Success      200    {object}  utils.Response  "分组创建成功"
// @Failure      200    {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/groups [post]
// @Security     BearerAuth
func (c *CMDBController) CreateServerGroup(ctx *gin.Context) {
	var group modelcmdb.ServerGroup
	if err := ctx.ShouldBindJSON(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.CreateServerGroup(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组创建成功"))
}

// UpdateServerGroup godoc
// @Summary      更新主机分组
// @Description  根据分组 ID 更新主机分组信息（部分字段更新）
// @Tags         CMDB-分组
// @Accept       json
// @Produce      json
// @Param        id     path      int                  true  "分组 ID"
// @Param        group  body      modelcmdb.ServerGroup  true  "需要更新的字段"
// @Success      200    {object}  utils.Response  "分组更新成功"
// @Failure      200    {object}  utils.Response  "无效的 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/groups/{id} [put]
// @Security     BearerAuth
func (c *CMDBController) UpdateServerGroup(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerGroup(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组更新成功"))
}

// DeleteServerGroup godoc
// @Summary      删除主机分组
// @Description  根据分组 ID 删除主机分组
// @Tags         CMDB-分组
// @Produce      json
// @Param        id  path  int  true  "分组 ID"
// @Success      200  {object}  utils.Response  "分组删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/groups/{id} [delete]
// @Security     BearerAuth
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

// AssignServerToGroup godoc
// @Summary      将服务器分配到分组
// @Description  将指定服务器分配到单个分组
// @Tags         CMDB-分组
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "分配请求"  examples({\"serverId\":1,\"groupId\":2})
// @Success      200   {object}  utils.Response  "服务器分配成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 分配失败"
// @Router       /cmdb/groups/assign [post]
// @Security     BearerAuth
func (c *CMDBController) AssignServerToGroup(ctx *gin.Context) {
	var req struct {
		ServerID uint `json:"serverId" binding:"required"`
		GroupID  uint `json:"groupId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.AssignServerToGroup(req.ServerID, req.GroupID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// AssignServerToGroups godoc
// @Summary      将服务器分配到多个分组
// @Description  将指定服务器分配到多个分组
// @Tags         CMDB-分组
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "分配请求"  examples({\"serverId\":1,\"groupIds\":[1,2,3]})
// @Success      200   {object}  utils.Response  "服务器分配成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 分配失败"
// @Router       /cmdb/groups/assign-multi [post]
// @Security     BearerAuth
func (c *CMDBController) AssignServerToGroups(ctx *gin.Context) {
	var req struct {
		ServerID uint   `json:"serverId" binding:"required"`
		GroupIDs []uint `json:"groupIds" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.AssignServerToGroups(req.ServerID, req.GroupIDs); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// GetServersByGroup godoc
// @Summary      获取分组下的服务器列表
// @Description  分页获取指定分组下的服务器列表
// @Tags         CMDB-分组
// @Produce      json
// @Param        groupId   path  int  true  "分组 ID"
// @Param        page      query int  false "页码"     default(1)
// @Param        pageSize  query int  false "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "无效的分组 ID / 请求参数错误 / 获取失败"
// @Router       /cmdb/group-servers/{groupId} [get]
// @Security     BearerAuth
func (c *CMDBController) GetServersByGroup(ctx *gin.Context) {
	groupIDStr := ctx.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的分组ID"))
		return
	}

	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	servers, err := c.svc.GetServersByGroup(uint(groupID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 手动分页
	total := int64(len(servers))
	start := (params.GetPage() - 1) * params.GetPageSize()
	end := start + params.GetPageSize()

	if start >= len(servers) {
		servers = []modelcmdb.Server{}
	} else if end > len(servers) {
		servers = servers[start:]
	} else {
		servers = servers[start:end]
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(servers, total, params)))
}
