package controller

import (
	"log"
	"net/http"
	"time"

	"oneops/backend/container"
	"oneops/backend/logger"
	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController 认证控制器
type AuthController struct {
	container *container.ServiceContainer
}

// NewAuthController 创建认证控制器
func NewAuthController(cnt *container.ServiceContainer) *AuthController {
	return &AuthController{
		container: cnt,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (ctrl *AuthController) Login(c *gin.Context) {
	startTime := time.Now()
	logger.Debug("[登录调试] 登录请求开始",
		zap.String("username", c.PostForm("username")),
		zap.String("clientIP", c.ClientIP()))

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Debug("[登录调试] 参数验证失败", zap.Error(err))
		c.JSON(http.StatusOK, utils.ErrorBadRequest("用户名和密码不能为空"))
		return
	}

	logger.Debug("[登录调试] 参数验证成功", zap.String("username", req.Username))

	authService := ctrl.container.AuthService()
	auditService := ctrl.container.AuditService()

	logger.Debug("[登录调试] 开始调用authService.Login")
	loginStart := time.Now()

	token, user, err := authService.Login(req.Username, req.Password)

	logger.Debug("[登录调试] authService.Login完成",
		zap.Duration("耗时", time.Since(loginStart)),
		zap.Error(err))

	if err != nil {
		logger.Debug("[登录调试] 登录失败", zap.Error(err))

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

	logger.Debug("[登录调试] 登录成功，开始获取用户信息",
		zap.Uint("userID", user.ID))

	// 获取用户信息（包含权限和菜单）
	userInfoStart := time.Now()
	userInfo, err := authService.GetUserInfo(user.ID)

	logger.Debug("[登录调试] authService.GetUserInfo完成",
		zap.Duration("耗时", time.Since(userInfoStart)),
		zap.Error(err))

	if err != nil {
		logger.Error("[登录调试] 获取用户信息失败", zap.Error(err))
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户信息失败"))
		return
	}

	logger.Debug("[登录调试] 开始记录登录日志")

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

	logger.Debug("[登录调试] 记录登录日志完成，准备返回响应")

	// 构建响应数据
	responseData := gin.H{
		"token": token,
		"user":  userInfo.ToMap(),
	}

	logger.Debug("[登录调试] 登录请求完成",
		zap.Duration("总耗时", time.Since(startTime)))

	c.JSON(http.StatusOK, utils.SuccessResponse(responseData, "登录成功"))
}

// GetUserInfo 获取用户信息
func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	authService := ctrl.container.AuthService()
	uid, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}
	userInfo, err := authService.GetUserInfo(uid)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户信息失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(userInfo.ToMap(), "success"))
}

// Logout 登出
func (ctrl *AuthController) Logout(c *gin.Context) {

	// 记录登出日志
	auditService := ctrl.container.AuditService()
	uid, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}
	auditService.LogLogout(uid)

	c.JSON(http.StatusOK, utils.SuccessWithMessage("登出成功"))
}
