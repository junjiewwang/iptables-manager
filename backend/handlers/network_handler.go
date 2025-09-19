package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"iptables-management-backend/services"
)

type NetworkHandler struct {
	networkService *services.NetworkService
	logService     *services.LogService
}

// NewNetworkHandler 创建网络处理器实例
func NewNetworkHandler(networkService *services.NetworkService, logService *services.LogService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
		logService:     logService,
	}
}

// GetInterfaces 获取所有网络接口
func (h *NetworkHandler) GetInterfaces(c *gin.Context) {
	log.Println("[DEBUG] GetInterfaces API called")

	interfaces, err := h.networkService.GetAllInterfaces()
	if err != nil {
		log.Printf("[ERROR] Failed to get network interfaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取网络接口失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d network interfaces", len(interfaces))
	c.JSON(http.StatusOK, interfaces)
}

// GetDockerBridges 获取Docker网桥信息
func (h *NetworkHandler) GetDockerBridges(c *gin.Context) {
	log.Println("[DEBUG] GetDockerBridges API called")

	bridges, err := h.networkService.GetDockerBridges()
	if err != nil {
		log.Printf("[ERROR] Failed to get Docker bridges: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取Docker网桥失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d Docker bridges", len(bridges))
	c.JSON(http.StatusOK, bridges)
}

// GetBridgeRules 获取指定网桥的规则
func (h *NetworkHandler) GetBridgeRules(c *gin.Context) {
	bridgeName := c.Param("name")
	log.Printf("[DEBUG] GetBridgeRules API called for bridge: %s", bridgeName)

	rules, err := h.networkService.GetBridgeRules(bridgeName)
	if err != nil {
		log.Printf("[ERROR] Failed to get rules for bridge %s: %v", bridgeName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取网桥规则失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d rules for bridge %s", len(rules), bridgeName)
	c.JSON(http.StatusOK, rules)
}

// GetNetworkConnections 获取网络连接信息
func (h *NetworkHandler) GetNetworkConnections(c *gin.Context) {
	log.Println("[DEBUG] GetNetworkConnections API called")

	connections, err := h.networkService.GetNetworkConnections()
	if err != nil {
		log.Printf("[ERROR] Failed to get network connections: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取网络连接失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d network connections", len(connections))
	c.JSON(http.StatusOK, connections)
}

// GetRouteTable 获取路由表信息
func (h *NetworkHandler) GetRouteTable(c *gin.Context) {
	log.Println("[DEBUG] GetRouteTable API called")

	routes, err := h.networkService.GetRouteTable()
	if err != nil {
		log.Printf("[ERROR] Failed to get route table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取路由表失败"})
		return
	}

	log.Printf("[DEBUG] Retrieved %d route entries", len(routes))
	c.JSON(http.StatusOK, routes)
}
