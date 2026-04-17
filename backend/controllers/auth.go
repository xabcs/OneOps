package controllers

import (
	"log"
	"net/http"

	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct{}

// 创建审计服务实例
var auditService = services.NewAuditService()

// NewAuthController 创建认证控制器
func NewAuthController() *AuthController {
	return &AuthController{}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("用户名和密码不能为空"))
		return
	}

	authService := services.NewAuthService()
	token, user, err := authService.Login(req.Username, req.Password)
	if err != nil {
		// 记录登录失败日志
		if logErr := auditService.LogLogin(
			0, // 用户ID未知
			req.Username,
			"",
			c.ClientIP(),
			c.Request.UserAgent(),
			"",
			"failed",
			"用户名或密码错误",
		); logErr != nil {
			// 记录日志失败，打印错误但不影响登录流程
			log.Printf("记录登录失败日志出错: %v", logErr)
		}

		if err == services.ErrInvalidPassword {
			c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户名或密码错误"))
		} else {
			c.JSON(http.StatusOK, utils.ErrorInternal("登录失败"))
		}
		return
	}

	// 获取用户信息（包含权限和菜单）
	userInfo, err := authService.GetUserInfo(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户信息失败"))
		return
	}

	// 记录登录成功日志
	if logErr := auditService.LogLogin(
		user.ID,
		user.Username,
		user.Nickname,
		c.ClientIP(),
		c.Request.UserAgent(),
		"", // 位置信息（可以后续通过IP解析）
		"success",
		"",
	); logErr != nil {
		// 记录日志失败，打印错误但不影响登录流程
		log.Printf("记录登录成功日志出错: %v", logErr)
	}

	// 构建响应数据
	responseData := gin.H{
		"token": token,
		"user":  userInfo.ToMap(),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(responseData, "登录成功"))
}

// GetUserInfo 获取用户信息
func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}

	authService := services.NewAuthService()
	userInfo, err := authService.GetUserInfo(userID.(uint))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户信息失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(userInfo.ToMap(), "success"))
}

// Logout 登出
func (ctrl *AuthController) Logout(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}

	// 记录登出日志
	auditService.LogLogout(userID.(uint))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("登出成功"))
}
