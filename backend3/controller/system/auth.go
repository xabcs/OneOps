package system

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"
)

// AuditLogger 审计日志接口（解耦对 audit 包的直接依赖）
type AuditLogger interface {
	LogLogin(userID uint, username, nickname, ip, userAgent, location, status, failReason string) error
	LogLogout(userID uint) error
}

// AuthController 认证控制器
type AuthController struct {
	authSvc  *system.AuthService
	auditSvc AuditLogger
}

// NewAuthController 创建认证控制器
func NewAuthController(authSvc *system.AuthService, auditSvc AuditLogger) *AuthController {
	return &AuthController{
		authSvc:  authSvc,
		auditSvc: auditSvc,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名密码登录，返回 JWT token 和用户信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "登录请求"
// @Success      200   {object}  utils.Response{data=object{token=string,user=object}}
// @Failure      200   {object}  utils.Response  "用户名或密码错误 / 登录失败"
// @Router       /login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	startTime := time.Now()
	logger.Debug("[登录调试] 登录请求开始",
		zap.String("username", c.PostForm("username")),
		zap.String("client_ip", c.ClientIP()))

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Debug("[登录调试] 参数验证失败", zap.Error(err))
		c.JSON(http.StatusOK, utils.ErrorBadRequest("用户名和密码不能为空"))
		return
	}

	logger.Debug("[登录调试] 参数验证成功", zap.String("username", req.Username))

	loginStart := time.Now()

	token, user, err := ctrl.authSvc.Login(req.Username, req.Password)

	logger.Debug("[登录调试] authService.Login完成",
		zap.Duration("耗时", time.Since(loginStart)),
		zap.Error(err))

	if err != nil {
		logger.Debug("[登录调试] 登录失败", zap.Error(err))

		// 区分失败原因：禁用账号单独提示；其余（含用户不存在）按防枚举口径统一为用户名或密码错误
		failReason := "登录失败"
		resp := utils.ErrorInternal("登录失败")
		switch {
		case errors.Is(err, system.ErrUserDisabled):
			failReason = "账号已被禁用，请联系管理员"
			resp = utils.ErrorUnauthorized(failReason)
		case errors.Is(err, system.ErrInvalidPassword), errors.Is(err, gorm.ErrRecordNotFound):
			failReason = "用户名或密码错误"
			resp = utils.ErrorUnauthorized(failReason)
		}

		// 记录登录失败日志
		if logErr := ctrl.auditSvc.LogLogin(
			0, // 用户ID未知
			req.Username,
			"",
			c.ClientIP(),
			c.Request.UserAgent(),
			"",
			"failed",
			failReason,
		); logErr != nil {
			// 记录日志失败，打印错误但不影响登录流程
			log.Printf("记录登录失败日志出错: %v", logErr)
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	logger.Debug("[登录调试] 登录成功，开始获取用户信息",
		zap.Uint("user_id", user.ID))

	// 获取用户信息（包含权限和菜单）
	userInfoStart := time.Now()
	userInfo, err := ctrl.authSvc.GetUserInfo(user.ID)

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
	if logErr := ctrl.auditSvc.LogLogin(
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

// GetUserInfo godoc
// @Summary      获取当前用户信息
// @Description  根据登录态获取当前登录用户的详细信息（含角色、权限、菜单）
// @Tags         认证
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "获取用户信息失败"
// @Router       /user/info [get]
// @Security     BearerAuth
func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	uid, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}
	userInfo, err := ctrl.authSvc.GetUserInfo(uid)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户信息失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(userInfo.ToMap(), "success"))
}

// Logout godoc
// @Summary      用户登出
// @Description  登出当前用户并记录登出日志
// @Tags         认证
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /logout [post]
// @Security     BearerAuth
func (ctrl *AuthController) Logout(c *gin.Context) {
	uid, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// 记录登出日志
	_ = ctrl.auditSvc.LogLogout(uid)

	c.JSON(http.StatusOK, utils.SuccessWithMessage("登出成功"))
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// UpdatePassword godoc
// @Summary      修改当前用户密码
// @Description  校验原密码后更新当前登录用户密码（注意：该 handler 尚未注册路由）
// @Tags         系统管理-认证
// @Accept       json
// @Produce      json
// @Param        request  body  ChangePasswordRequest  true  "原密码与新密码"
// @Success      200  {object}  utils.Response  "修改成功"
// @Failure      200  {object}  utils.Response  "请求参数错误 / 原密码错误 / 更新密码失败"
// @Router       /auth/password [put]
// @Security     BearerAuth
func (ctrl *AuthController) UpdatePassword(c *gin.Context) {
	uid, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	// 获取当前用户（含密码哈希）
	userInfo, err := ctrl.authSvc.GetUserInfo(uid)
	if err != nil || userInfo.User == nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("用户不存在"))
		return
	}

	// 验证原密码
	if !utils.CheckPassword(req.OldPassword, userInfo.User.Password) {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("原密码错误"))
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("密码加密失败"))
		return
	}

	if err := database.GetDB().
		Model(&modelsystem.User{}).
		Where("id = ?", uid).
		Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新密码失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("密码修改成功"))
}
