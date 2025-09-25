package controllers

import (
	"iptables-management-backend/models"
	"iptables-management-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TunnelController struct {
	networkService *services.NetworkService
}

// NewTunnelController 创建隧道控制器实例
func NewTunnelController() *TunnelController {
	return &TunnelController{
		networkService: services.NewNetworkService(),
	}
}

// GetTunnelInterfaceRules 获取隧道接口相关的iptables规则
// @Summary 获取隧道接口相关的iptables规则
// @Description 获取指定隧道接口相关的所有iptables规则
// @Tags tunnel
// @Accept json
// @Produce json
// @Param interface_name path string true "隧道接口名称"
// @Success 200 {object} map[string]interface{} "成功返回规则列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/{interface_name}/rules [get]
func (tc *TunnelController) GetTunnelInterfaceRules(c *gin.Context) {
	interfaceName := c.Param("interface_name")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "接口名称不能为空",
		})
		return
	}

	rules, err := tc.networkService.GetTunnelInterfaceRules(interfaceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取隧道接口规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"interface_name": interfaceName,
		"rules_count":    len(rules),
		"rules":          rules,
	})
}

// GetTunnelInterfaceInfo 获取隧道接口详细信息
// @Summary 获取隧道接口详细信息
// @Description 获取指定隧道接口的详细配置和状态信息
// @Tags tunnel
// @Accept json
// @Produce json
// @Param interface_name path string true "隧道接口名称"
// @Success 200 {object} map[string]interface{} "成功返回接口信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "接口不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/{interface_name}/info [get]
func (tc *TunnelController) GetTunnelInterfaceInfo(c *gin.Context) {
	interfaceName := c.Param("interface_name")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "接口名称不能为空",
		})
		return
	}

	tunnelInfo, err := tc.networkService.GetTunnelInterfaceInfo(interfaceName)
	if err != nil {
		if err.Error() == "interface "+interfaceName+" not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "隧道接口不存在: " + interfaceName,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取隧道接口信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"tunnel_info": tunnelInfo,
	})
}

// AnalyzeTunnelDockerCommunication 分析隧道接口与Docker网桥的通信
// @Summary 分析隧道接口与Docker网桥的通信
// @Description 分析指定隧道接口与Docker网桥之间的通信路径和规则
// @Tags tunnel
// @Accept json
// @Produce json
// @Param tunnel_interface query string true "隧道接口名称"
// @Param docker_bridge query string true "Docker网桥名称"
// @Success 200 {object} map[string]interface{} "成功返回分析结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/analyze-communication [get]
func (tc *TunnelController) AnalyzeTunnelDockerCommunication(c *gin.Context) {
	tunnelInterface := c.Query("tunnel_interface")
	dockerBridge := c.Query("docker_bridge")

	if tunnelInterface == "" || dockerBridge == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "隧道接口名称和Docker网桥名称都不能为空",
		})
		return
	}

	analysis, err := tc.networkService.AnalyzeTunnelDockerCommunication(tunnelInterface, dockerBridge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "分析通信路径失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"analysis": analysis,
	})
}

// GetTunnelInterfaces 获取所有隧道接口列表
// @Summary 获取所有隧道接口列表
// @Description 获取系统中所有的隧道接口(tun/tap)列表
// @Tags tunnel
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回隧道接口列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/interfaces [get]
func (tc *TunnelController) GetTunnelInterfaces(c *gin.Context) {
	interfaces, err := tc.networkService.GetAllInterfaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取网络接口失败: " + err.Error(),
		})
		return
	}

	var tunnelInterfaces []models.NetworkInterface
	for _, iface := range interfaces {
		if iface.Type == "tunnel" {
			tunnelInterfaces = append(tunnelInterfaces, iface)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"tunnel_interfaces": tunnelInterfaces,
		"count":             len(tunnelInterfaces),
	})
}

// GetDockerBridges 获取所有Docker网桥列表
// @Summary 获取所有Docker网桥列表
// @Description 获取系统中所有的Docker网桥接口列表，包含IP地址信息
// @Tags tunnel
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回Docker网桥列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/docker-bridges [get]
func (tc *TunnelController) GetDockerBridges(c *gin.Context) {
	dockerBridges, err := tc.networkService.GetDockerBridges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取Docker网桥失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"docker_bridges": dockerBridges,
		"count":          len(dockerBridges),
	})
}

// GenerateTunnelDockerRules 生成隧道与Docker通信规则
// @Summary 生成隧道与Docker通信规则
// @Description 根据指定参数生成隧道接口与Docker网桥通信的iptables规则
// @Tags tunnel
// @Accept json
// @Produce json
// @Param request body TunnelDockerRuleRequest true "规则生成请求"
// @Success 200 {object} map[string]interface{} "成功返回生成的规则"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/generate-rules [post]
func (tc *TunnelController) GenerateTunnelDockerRules(c *gin.Context) {
	var request TunnelDockerRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数格式错误: " + err.Error(),
		})
		return
	}

	if request.TunnelInterface == "" || request.DockerBridge == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "隧道接口和Docker网桥名称不能为空",
		})
		return
	}

	rules := tc.generateCommunicationRules(request)

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"generated_rules":  rules,
		"rules_count":      len(rules),
		"tunnel_interface": request.TunnelInterface,
		"docker_bridge":    request.DockerBridge,
	})
}

// TunnelDockerRuleRequest 隧道Docker规则生成请求
type TunnelDockerRuleRequest struct {
	TunnelInterface string `json:"tunnel_interface" binding:"required"`
	DockerBridge    string `json:"docker_bridge" binding:"required"`
	Direction       string `json:"direction"` // bidirectional, inbound, outbound
	Protocol        string `json:"protocol"`  // tcp, udp, icmp, all
	SourcePort      string `json:"source_port"`
	DestPort        string `json:"dest_port"`
	Action          string `json:"action"` // ACCEPT, DROP, REJECT
	EnableNAT       bool   `json:"enable_nat"`
	EnableLogging   bool   `json:"enable_logging"`
}

// generateCommunicationRules 生成通信规则
func (tc *TunnelController) generateCommunicationRules(request TunnelDockerRuleRequest) []string {
	var rules []string

	// 设置默认值
	if request.Direction == "" {
		request.Direction = "bidirectional"
	}
	if request.Protocol == "" {
		request.Protocol = "all"
	}
	if request.Action == "" {
		request.Action = "ACCEPT"
	}

	// 生成基础转发规则
	if request.Direction == "bidirectional" || request.Direction == "inbound" {
		rule := "iptables -A FORWARD -i " + request.TunnelInterface + " -o " + request.DockerBridge
		if request.Protocol != "all" {
			rule += " -p " + request.Protocol
		}
		if request.DestPort != "" {
			rule += " --dport " + request.DestPort
		}
		if request.EnableLogging {
			rule += " -j LOG --log-prefix \"TUNNEL-TO-DOCKER: \""
			rules = append(rules, rule)
			rule = "iptables -A FORWARD -i " + request.TunnelInterface + " -o " + request.DockerBridge
			if request.Protocol != "all" {
				rule += " -p " + request.Protocol
			}
			if request.DestPort != "" {
				rule += " --dport " + request.DestPort
			}
		}
		rule += " -j " + request.Action
		rules = append(rules, rule)
	}

	if request.Direction == "bidirectional" || request.Direction == "outbound" {
		rule := "iptables -A FORWARD -i " + request.DockerBridge + " -o " + request.TunnelInterface + " -m conntrack --ctstate RELATED,ESTABLISHED"
		if request.EnableLogging {
			rule += " -j LOG --log-prefix \"DOCKER-TO-TUNNEL: \""
			rules = append(rules, rule)
			rule = "iptables -A FORWARD -i " + request.DockerBridge + " -o " + request.TunnelInterface + " -m conntrack --ctstate RELATED,ESTABLISHED"
		}
		rule += " -j " + request.Action
		rules = append(rules, rule)
	}

	// 生成NAT规则
	if request.EnableNAT {
		natRule := "iptables -t nat -A POSTROUTING -o " + request.TunnelInterface + " -j MASQUERADE"
		rules = append(rules, natRule)

		// Docker网络的MASQUERADE规则
		dockerNatRule := "iptables -t nat -A POSTROUTING -s 172.17.0.0/16 -o " + request.TunnelInterface + " -j MASQUERADE"
		rules = append(rules, dockerNatRule)
	}

	// 生成INPUT规则（允许隧道接口的基本通信）
	inputRule := "iptables -A INPUT -i " + request.TunnelInterface + " -j " + request.Action
	rules = append(rules, inputRule)

	// 生成OUTPUT规则
	outputRule := "iptables -A OUTPUT -o " + request.TunnelInterface + " -j " + request.Action
	rules = append(rules, outputRule)

	return rules
}

// GetTunnelStatistics 获取隧道接口统计信息
// @Summary 获取隧道接口统计信息
// @Description 获取指定隧道接口的流量统计和性能指标
// @Tags tunnel
// @Accept json
// @Produce json
// @Param interface_name path string true "隧道接口名称"
// @Param hours query int false "统计时间范围(小时)" default(24)
// @Success 200 {object} map[string]interface{} "成功返回统计信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/{interface_name}/statistics [get]
func (tc *TunnelController) GetTunnelStatistics(c *gin.Context) {
	interfaceName := c.Param("interface_name")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "接口名称不能为空",
		})
		return
	}

	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		hours = 24
	}

	// 获取接口基础统计
	interfaces, err := tc.networkService.GetAllInterfaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取接口信息失败: " + err.Error(),
		})
		return
	}

	var targetInterface *models.NetworkInterface
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			targetInterface = &iface
			break
		}
	}

	if targetInterface == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "接口不存在: " + interfaceName,
		})
		return
	}

	// 获取相关规则统计
	rules, _ := tc.networkService.GetTunnelInterfaceRules(interfaceName)

	var totalPackets, totalBytes int64
	for _, rule := range rules {
		totalPackets += rule.Packets
		totalBytes += rule.Bytes
	}

	statistics := map[string]interface{}{
		"interface_name":     interfaceName,
		"time_range_hours":   hours,
		"interface_stats":    targetInterface.Statistics,
		"rules_count":        len(rules),
		"total_rule_packets": totalPackets,
		"total_rule_bytes":   totalBytes,
		"is_up":              targetInterface.IsUp,
		"mtu":                targetInterface.MTU,
		"ip_addresses":       targetInterface.IPAddresses,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"statistics": statistics,
	})
}

// FixConnectivity 修复隧道接口与Docker网桥的连通性问题
// @Summary 修复隧道接口与Docker网桥的连通性问题
// @Description 自动检测并修复隧道接口与Docker网桥之间的连通性问题
// @Tags tunnel
// @Accept json
// @Produce json
// @Param request body FixConnectivityRequest true "修复连通性请求"
// @Success 200 {object} map[string]interface{} "成功返回修复结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tunnel/fix-connectivity [post]
func (tc *TunnelController) FixConnectivity(c *gin.Context) {
	var request FixConnectivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数格式错误: " + err.Error(),
		})
		return
	}

	if request.TunnelInterface == "" || request.DockerBridge == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "隧道接口和Docker网桥名称不能为空",
		})
		return
	}

	result, err := tc.networkService.FixConnectivity(request.TunnelInterface, request.DockerBridge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "修复连通性失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// FixConnectivityRequest 修复连通性请求
type FixConnectivityRequest struct {
	TunnelInterface string `json:"tunnel_interface" binding:"required"`
	DockerBridge    string `json:"docker_bridge" binding:"required"`
}
