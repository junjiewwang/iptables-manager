package handlers

import (
	"github.com/gin-gonic/gin"
	"iptables-management-backend/services"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TopologyHandler struct {
	topologyService *services.TopologyService
}

// NewTopologyHandler 创建拓扑处理器实例
func NewTopologyHandler(topologyService *services.TopologyService) *TopologyHandler {
	return &TopologyHandler{
		topologyService: topologyService,
	}
}

// GetTopology 获取拓扑图数据（支持查询参数）
func (h *TopologyHandler) GetTopology(c *gin.Context) {
	log.Println("[DEBUG] GetTopology API called")

	// 解析查询参数
	options := services.TopologyOptions{
		ProtocolFilter:  c.Query("protocol"),
		ChainFilter:     c.Query("chain"),
		InterfaceFilter: c.Query("interface"),
		RuleTypeFilter:  c.Query("rule_type"),
		IncludeStats:    c.Query("include_stats") == "true",
		IncludeMetadata: c.Query("include_metadata") == "true",
	}

	// 解析分页参数
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			pageSize := 50 // 默认分页大小
			if ps := c.Query("page_size"); ps != "" {
				if psNum, err := strconv.Atoi(ps); err == nil && psNum > 0 {
					pageSize = psNum
				}
			}
			options.Pagination = &services.PaginationOptions{
				Page:     pageNum,
				PageSize: pageSize,
			}
		}
	}

	var topology interface{}
	var err error

	// 根据是否有查询参数选择不同的服务方法
	if options.ProtocolFilter != "" || options.ChainFilter != "" ||
		options.InterfaceFilter != "" || options.Pagination != nil {
		topology, err = h.topologyService.GetTopologyDataWithOptions(options)
	} else {
		topology, err = h.topologyService.GetTopologyData()
	}

	if err != nil {
		log.Printf("[ERROR] Failed to get topology data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get topology data",
			"details": err.Error(),
			"code":    "TOPOLOGY_FETCH_ERROR",
		})
		return
	}

	response := gin.H{
		"success": true,
		"data":    topology,
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"version":   "1.0",
		},
	}

	// 添加统计信息
	if options.IncludeStats {
		if stats, err := h.topologyService.GetTopologyStats(); err == nil {
			response["stats"] = stats
		}
	}

	log.Printf("[DEBUG] Returning topology data with meta information")
	c.JSON(http.StatusOK, response)
}

// GetTopologyStats 获取拓扑统计信息
func (h *TopologyHandler) GetTopologyStats(c *gin.Context) {
	log.Println("[DEBUG] GetTopologyStats API called")

	stats, err := h.topologyService.GetTopologyStats()
	if err != nil {
		log.Printf("[ERROR] Failed to get topology stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get topology statistics",
			"details": err.Error(),
			"code":    "STATS_FETCH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
		},
	})
}

// RefreshTopology 强制刷新拓扑缓存
func (h *TopologyHandler) RefreshTopology(c *gin.Context) {
	log.Println("[DEBUG] RefreshTopology API called")

	// 这里可以添加缓存失效逻辑
	// h.topologyService.InvalidateCache()

	topology, err := h.topologyService.GetTopologyData()
	if err != nil {
		log.Printf("[ERROR] Failed to refresh topology data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to refresh topology data",
			"details": err.Error(),
			"code":    "REFRESH_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    topology,
		"meta": gin.H{
			"timestamp": time.Now().Unix(),
			"refreshed": true,
		},
	})
}

// ExportTopology 导出拓扑数据
func (h *TopologyHandler) ExportTopology(c *gin.Context) {
	log.Println("[DEBUG] ExportTopology API called")

	format := c.DefaultQuery("format", "json")

	topology, err := h.topologyService.GetTopologyData()
	if err != nil {
		log.Printf("[ERROR] Failed to export topology data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to export topology data",
			"details": err.Error(),
			"code":    "EXPORT_ERROR",
		})
		return
	}

	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=topology.json")
		c.JSON(http.StatusOK, topology)
	case "csv":
		// 这里可以实现CSV导出逻辑
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=topology.csv")
		// 实现CSV转换逻辑
		c.String(http.StatusOK, "CSV export not implemented yet")
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported format. Use 'json' or 'csv'",
			"code":  "INVALID_FORMAT",
		})
	}
}

// GetTopologyHealth 获取拓扑服务健康状态
func (h *TopologyHandler) GetTopologyHealth(c *gin.Context) {
	log.Println("[DEBUG] GetTopologyHealth API called")

	// 简单的健康检查
	start := time.Now()
	_, err := h.topologyService.GetTopologyData()
	responseTime := time.Since(start)

	health := gin.H{
		"status":        "healthy",
		"response_time": responseTime.Milliseconds(),
		"timestamp":     time.Now().Unix(),
	}

	if err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"data":    health,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}
