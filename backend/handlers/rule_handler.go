package handlers

import (
	"net/http"
	"strconv"

	"iptables-management-backend/models"
	"iptables-management-backend/services"
	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	ruleService *services.RuleService
	logService  *services.LogService
}

// NewRuleHandler 创建规则处理器实例
func NewRuleHandler(ruleService *services.RuleService, logService *services.LogService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
		logService:  logService,
	}
}

// GetRules 获取所有规则
func (h *RuleHandler) GetRules(c *gin.Context) {
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取规则失败"})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// GetRule 获取单个规则
func (h *RuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	rule, err := h.ruleService.GetRuleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateRule 创建新规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule models.IPTablesRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleService.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建规则失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	h.logService.LogOperation(
		username.(string),
		"创建规则",
		"创建新的iptables规则: "+rule.RuleText,
		c.ClientIP(),
	)

	c.JSON(http.StatusCreated, rule)
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	var rule models.IPTablesRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleService.UpdateRule(uint(id), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新规则失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	h.logService.LogOperation(
		username.(string),
		"更新规则",
		"更新iptables规则ID: "+idStr,
		c.ClientIP(),
	)

	c.JSON(http.StatusOK, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	if err := h.ruleService.DeleteRule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除规则失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	h.logService.LogOperation(
		username.(string),
		"删除规则",
		"删除iptables规则ID: "+idStr,
		c.ClientIP(),
	)

	c.JSON(http.StatusOK, gin.H{"message": "规则删除成功"})
}

// GetStatistics 获取统计信息
func (h *RuleHandler) GetStatistics(c *gin.Context) {
	stats, err := h.ruleService.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}