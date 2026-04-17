package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
)

// MenuController 菜单控制器
type MenuController struct{}

// NewMenuController 创建菜单控制器
func NewMenuController() *MenuController {
	return &MenuController{}
}

// GetMenus 获取所有菜单
func (ctrl *MenuController) GetMenus(c *gin.Context) {
	db := services.GetDB()

	var menus []models.Menu
	if err := db.Order("sort ASC").Find(&menus).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取菜单列表失败"))
		return
	}

	// 构建菜单树
	menuTree := ctrl.buildMenuTree(menus, 0)

	c.JSON(http.StatusOK, utils.SuccessWithData(menuTree))
}

// buildMenuTree 递归构建菜单树
func (ctrl *MenuController) buildMenuTree(menus []models.Menu, parentID uint) []models.Menu {
	var result []models.Menu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			menuItem := menu
			children := ctrl.buildMenuTree(menus, menu.ID)
			if len(children) > 0 {
				menuItem.Children = make([]*models.Menu, len(children))
				for i := range children {
					menuItem.Children[i] = &children[i]
				}
			}
			result = append(result, menuItem)
		}
	}
	return result
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name       string `json:"name" binding:"required"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
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

	db := services.GetDB()
	menu := models.Menu{
		Name:       req.Name,
		Icon:       req.Icon,
		Path:       req.Path,
		Permission: req.Permission,
		ParentID:   req.ParentID,
		Sort:       req.Sort,
		Status:     req.Status,
	}

	if err := db.Create(&menu).Error; err != nil {
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

	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	db := services.GetDB()
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
	updates["parent_id"] = req.ParentID
	updates["sort"] = req.Sort
	updates["status"] = req.Status

	if err := db.Model(&models.Menu{}).Where("id = ?", id).Updates(updates).Error; err != nil {
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

	db := services.GetDB()

	// 检查是否有子菜单
	var count int64
	db.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("该菜单下有子菜单，无法删除"))
		return
	}

	if err := db.Delete(&models.Menu{}, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除菜单失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
