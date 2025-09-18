package handlers

import (
	"net/http"

	"iptables-management-backend/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	logService  *services.LogService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *services.AuthService, logService *services.LogService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logService:  logService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		// 记录登录失败日志
		h.logService.LogOperation(req.Username, "登录失败", err.Error(), c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 记录登录成功日志
	h.logService.LogOperation(response.Username, "登录成功", "用户登录系统", c.ClientIP())

	c.JSON(http.StatusOK, response)
}