package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"iptables-management-backend/services"
)

type TableHandler struct {
	tableService *services.TableService
	logService   *services.LogService
}

// NewTableHandler 创建表处理器实例
func NewTableHandler(tableService *services.TableService, logService *services.LogService) *TableHandler {
	return &TableHandler{
		tableService: tableService,
		logService:   logService,
	}
}

// GetAllTables 获取所有表信息
func (h *TableHandler) GetAllTables(c *gin.Context) {
	log.Println("[DEBUG] GetAllTables API called")

	tables, err := h.tableService.GetAllTables()
	if err != nil {
		log.Printf("[ERROR] Failed to get tables: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取表信息失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d tables", len(tables))
	c.JSON(http.StatusOK, tables)
}

// GetTableInfo 获取指定表信息
func (h *TableHandler) GetTableInfo(c *gin.Context) {
	tableName := c.Param("table")
	log.Printf("[DEBUG] GetTableInfo API called for table: %s", tableName)

	tableInfo, err := h.tableService.GetTableInfo(tableName)
	if err != nil {
		log.Printf("[ERROR] Failed to get table info for %s: %v", tableName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取表信息失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved table info for %s with %d chains", tableName, len(tableInfo.Chains))
	c.JSON(http.StatusOK, tableInfo)
}

// GetChainVerbose 获取指定链的详细信息
func (h *TableHandler) GetChainVerbose(c *gin.Context) {
	tableName := c.Param("table")
	chainName := c.Param("chain")
	log.Printf("[DEBUG] GetChainVerbose API called for table: %s, chain: %s", tableName, chainName)

	chainInfo, err := h.tableService.GetChainVerbose(tableName, chainName)
	if err != nil {
		log.Printf("[ERROR] Failed to get chain verbose info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取链详细信息失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved verbose info for chain %s with %d rules", chainName, len(chainInfo.Rules))
	c.JSON(http.StatusOK, chainInfo)
}

// GetSpecialChains 获取特殊链信息（如DOCKER相关链）
func (h *TableHandler) GetSpecialChains(c *gin.Context) {
	log.Println("[DEBUG] GetSpecialChains API called")

	// 获取一些特殊的链信息
	specialChains := []map[string]interface{}{}

	// FORWARD链详细信息
	forwardInfo, err := h.tableService.GetChainVerbose("filter", "FORWARD")
	if err == nil {
		specialChains = append(specialChains, map[string]interface{}{
			"name":  "FORWARD (详细)",
			"table": "filter",
			"chain": "FORWARD",
			"info":  forwardInfo,
		})
	}

	// NAT表的POSTROUTING链
	postRoutingInfo, err := h.tableService.GetChainVerbose("nat", "POSTROUTING")
	if err == nil {
		specialChains = append(specialChains, map[string]interface{}{
			"name":  "POSTROUTING (详细)",
			"table": "nat",
			"chain": "POSTROUTING",
			"info":  postRoutingInfo,
		})
	}

	// 尝试获取DOCKER相关链（可能不存在）
	dockerIsolationInfo, err := h.tableService.GetChainVerbose("filter", "DOCKER-ISOLATION-STAGE-2")
	if err == nil {
		specialChains = append(specialChains, map[string]interface{}{
			"name":  "DOCKER-ISOLATION-STAGE-2 (详细)",
			"table": "filter",
			"chain": "DOCKER-ISOLATION-STAGE-2",
			"info":  dockerIsolationInfo,
		})
	}

	log.Printf("[DEBUG] Retrieved %d special chains", len(specialChains))
	c.JSON(http.StatusOK, specialChains)
}
