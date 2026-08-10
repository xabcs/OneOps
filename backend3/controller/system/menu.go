package system

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"
	syssvc "oneops/backend3/service/system"
)

// MenuController 菜单控制器
type MenuController struct {
	svc *syssvc.MenuService
}

// NewMenuController 创建菜单控制器
func NewMenuController(svc *syssvc.MenuService) *MenuController {
	return &MenuController{svc: svc}
}

// GetMenus 获取所有菜单
func (ctrl *MenuController) GetMenus(c *gin.Context) {
	menuTree, err := ctrl.svc.GetAllAsTree()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取菜单列表失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(menuTree))
}

// GetMenuTree 获取菜单树（专用于前端菜单管理）
func (ctrl *MenuController) GetMenuTree(c *gin.Context) {
	menuTree, err := ctrl.svc.GetAllAsTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取菜单树失败",
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    menuTree,
	})
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name       string `json:"name" binding:"required"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
	MenuType   string `json:"menuType"`
	ParentID   uint   `json:"parentId"`
	Sort       int    `json:"sort"`
	Status     int    `json:"status"`
}

// CreateMenu 创建菜单
func (ctrl *MenuController) CreateMenu(c *gin.Context) {
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	menu := modelsystem.Menu{
		Name:       req.Name,
		Icon:       req.Icon,
		Path:       req.Path,
		Permission: req.Permission,
		MenuType:   req.MenuType,
		ParentID:   req.ParentID,
		Sort:       req.Sort,
		Status:     req.Status,
	}

	if err := ctrl.svc.Create(&menu); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建菜单失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(menu))
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
	MenuType   string `json:"menuType"`
	ParentID   uint   `json:"parentId"`
	Sort       int    `json:"sort"`
	Status     int    `json:"status"`
}

// UpdateMenu 更新菜单
func (ctrl *MenuController) UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的菜单ID"))
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("读取请求失败"))
		return
	}

	var rawBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("JSON解析失败"))
		return
	}

	var req UpdateMenuRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数解析失败"))
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Permission != "" {
		updates["permission"] = req.Permission
	}
	if req.MenuType != "" {
		updates["menu_type"] = req.MenuType
	}
	if _, ok := rawBody["parentId"]; ok {
		updates["parent_id"] = req.ParentID
	}
	if _, ok := rawBody["sort"]; ok {
		updates["sort"] = req.Sort
	}
	if _, ok := rawBody["status"]; ok {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("没有提供要更新的字段"))
		return
	}

	if err := ctrl.svc.Update(id, updates); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新菜单失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteMenu 删除菜单
func (ctrl *MenuController) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的菜单ID"))
		return
	}

	if err := ctrl.svc.Delete(id); err != nil {
		if errors.Is(err, syssvc.ErrMenuHasChildren) {
			c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
			return
		}
		c.JSON(http.StatusOK, utils.ErrorInternal("删除菜单失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
