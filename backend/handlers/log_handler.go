package handlers

import (
	"net/http"

	"iptables-management-backend/models"
	"iptables-management-backend/services"
	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *services.LogService
}

// NewLogHandler 创建日志处理器实例
func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// GetLogs 获取所有操作日志
func (h *LogHandler) GetLogs(c *gin.Context) {
	username := c.Query("username")
	
	var logs []models.OperationLog
	var err error
	
	if username != "" {
		logs, err = h.logService.GetLogsByUser(username)
	} else {
		logs, err = h.logService.GetAllLogs()
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败"})
		return
	}

	c.JSON(http.StatusOK, logs)
}