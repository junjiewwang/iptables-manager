package services

import (
	"fmt"
	"iptables-management-backend/models"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type NetworkService struct{}

// NewNetworkService 创建网络服务实例
func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

// GetAllInterfaces 获取所有网络接口
func (s *NetworkService) GetAllInterfaces() ([]models.NetworkInterface, error) {
	log.Println("[DEBUG] NetworkService.GetAllInterfaces called")

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var result []models.NetworkInterface
	for _, iface := range interfaces {
		netInterface, err := s.parseInterface(iface)
		if err != nil {
			log.Printf("[WARN] Failed to parse interface %s: %v", iface.Name, err)
			continue
		}
		result = append(result, netInterface)
	}

	log.Printf("[DEBUG] Retrieved %d network interfaces", len(result))
	return result, nil
}

// parseInterface 解析网络接口信息
func (s *NetworkService) parseInterface(iface net.Interface) (models.NetworkInterface, error) {
	netInterface := models.NetworkInterface{
		Name:       iface.Name,
		MACAddress: iface.HardwareAddr.String(),
		MTU:        iface.MTU,
		IsUp:       iface.Flags&net.FlagUp != 0,
		IsDocker:   s.isDockerInterface(iface.Name),
	}

	// 设置接口类型和Docker类型
	netInterface.Type = s.getInterfaceType(iface.Name)
	if netInterface.IsDocker {
		netInterface.DockerType = s.getDockerInterfaceType(iface.Name)
	}

	// 设置状态
	if netInterface.IsUp {
		netInterface.State = "UP"
	} else {
		netInterface.State = "DOWN"
	}

	// 获取IP地址
	addrs, err := iface.Addrs()
	if err == nil {
		for _, addr := range addrs {
			netInterface.IPAddresses = append(netInterface.IPAddresses, addr.String())
		}
	}

	// 获取统计信息
	stats, err := s.getInterfaceStats(iface.Name)
	if err == nil {
		netInterface.Statistics = stats
	} else {
		// 如果获取统计信息失败，使用默认值
		netInterface.Statistics = models.InterfaceStats{
			RxBytes:   0,
			TxBytes:   0,
			RxPackets: 0,
			TxPackets: 0,
			RxErrors:  0,
			TxErrors:  0,
		}
	}

	return netInterface, nil
}

// isDockerInterface 判断是否为Docker接口
func (s *NetworkService) isDockerInterface(name string) bool {
	dockerPatterns := []string{
		"^docker\\d+$",
		"^br-[a-f0-9]{12}$",
		"^veth[a-f0-9]+$",
	}

	for _, pattern := range dockerPatterns {
		matched, _ := regexp.MatchString(pattern, name)
		if matched {
			return true
		}
	}

	return false
}

// getInterfaceType 获取接口类型
func (s *NetworkService) getInterfaceType(name string) string {
	if strings.HasPrefix(name, "lo") {
		return "loopback"
	}
	if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") {
		return "ethernet"
	}
	if strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "wlan") {
		return "wireless"
	}
	if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") {
		return "bridge"
	}
	if strings.HasPrefix(name, "veth") {
		return "veth"
	}
	if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap") {
		return "tunnel"
	}
	return "unknown"
}

// getDockerInterfaceType 获取Docker接口类型
func (s *NetworkService) getDockerInterfaceType(name string) string {
	if strings.HasPrefix(name, "docker") {
		return "default_bridge"
	}
	if strings.HasPrefix(name, "br-") {
		return "custom_bridge"
	}
	if strings.HasPrefix(name, "veth") {
		return "container_veth"
	}
	return "unknown"
}

// getInterfaceStats 获取接口统计信息
func (s *NetworkService) getInterfaceStats(name string) (models.InterfaceStats, error) {
	var stats models.InterfaceStats

	// 从 /proc/net/dev 读取统计信息
	cmd := exec.Command("cat", "/proc/net/dev")
	output, err := cmd.Output()
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, name+":") {
			fields := strings.Fields(line)
			if len(fields) >= 17 {
				stats.RxBytes, _ = strconv.ParseInt(fields[1], 10, 64)
				stats.RxPackets, _ = strconv.ParseInt(fields[2], 10, 64)
				stats.RxErrors, _ = strconv.ParseInt(fields[3], 10, 64)
				stats.TxBytes, _ = strconv.ParseInt(fields[9], 10, 64)
				stats.TxPackets, _ = strconv.ParseInt(fields[10], 10, 64)
				stats.TxErrors, _ = strconv.ParseInt(fields[11], 10, 64)
			}
			break
		}
	}

	return stats, nil
}

// GetDockerBridges 获取Docker网桥信息（使用系统原生命令）
func (s *NetworkService) GetDockerBridges() ([]models.DockerBridge, error) {
	log.Println("[DEBUG] NetworkService.GetDockerBridges called")

	var bridges []models.DockerBridge

	// 获取所有网络接口
	interfaces, err := s.GetAllInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	// 筛选Docker相关的网桥接口
	for _, iface := range interfaces {
		if s.isDockerBridge(iface.Name) {
			bridge, err := s.createBridgeFromInterface(iface)
			if err != nil {
				log.Printf("[WARN] Failed to create bridge from interface %s: %v", iface.Name, err)
				continue
			}
			bridges = append(bridges, bridge)
		}
	}

	log.Printf("[DEBUG] Retrieved %d Docker bridges", len(bridges))
	return bridges, nil
}

// isDockerBridge 判断是否为Docker网桥
func (s *NetworkService) isDockerBridge(name string) bool {
	// Docker默认网桥
	if name == "docker0" {
		return true
	}
	// Docker自定义网桥 (br-xxxxxxxxxxxx格式)
	if matched, _ := regexp.MatchString("^br-[a-f0-9]{12}$", name); matched {
		return true
	}
	return false
}

// createBridgeFromInterface 从网络接口创建Docker网桥信息
func (s *NetworkService) createBridgeFromInterface(iface models.NetworkInterface) (models.DockerBridge, error) {
	bridge := models.DockerBridge{
		Name:      iface.Name,
		Driver:    "bridge",
		Scope:     "local",
		Interface: iface,
	}

	// 根据接口名称设置网络ID和类型
	if iface.Name == "docker0" {
		bridge.NetworkID = "default"
	} else if strings.HasPrefix(iface.Name, "br-") {
		// 从br-xxxxxxxxxxxx中提取网络ID
		bridge.NetworkID = strings.TrimPrefix(iface.Name, "br-")
	}

	// 获取网桥的网络配置信息
	ipamConfig, err := s.getBridgeIPAMConfig(iface.Name)
	if err != nil {
		log.Printf("[WARN] Failed to get IPAM config for %s: %v", iface.Name, err)
		// 使用默认的IPAM配置
		bridge.IPAMConfig = models.DockerIPAMConfig{
			Driver:  "default",
			Config:  []models.DockerSubnet{},
			Options: make(map[string]string),
		}
	} else {
		bridge.IPAMConfig = ipamConfig
	}

	// 获取连接到网桥的容器信息
	containers, err := s.getBridgeContainers(iface.Name)
	if err != nil {
		log.Printf("[WARN] Failed to get containers for %s: %v", iface.Name, err)
		// 使用空的容器列表
		bridge.Containers = []models.DockerContainer{}
	} else {
		bridge.Containers = containers
	}

	return bridge, nil
}

// getBridgeIPAMConfig 获取网桥的IPAM配置（使用系统原生命令）
func (s *NetworkService) getBridgeIPAMConfig(bridgeName string) (models.DockerIPAMConfig, error) {
	config := models.DockerIPAMConfig{
		Driver: "default",
	}

	// 使用ip命令获取网桥的IP配置
	cmd := exec.Command("ip", "addr", "show", bridgeName)
	output, err := cmd.Output()
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			// 解析inet行，格式如: inet 172.17.0.1/16 brd 172.17.255.255 scope global docker0
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				cidr := fields[1]
				if ip, ipNet, err := net.ParseCIDR(cidr); err == nil {
					subnet := models.DockerSubnet{
						Subnet:  ipNet.String(),
						Gateway: ip.String(),
					}
					config.Config = append(config.Config, subnet)
				}
			}
		}
	}

	return config, nil
}

// getBridgeContainers 获取连接到网桥的容器信息（使用系统原生命令）
func (s *NetworkService) getBridgeContainers(bridgeName string) ([]models.DockerContainer, error) {
	var containers []models.DockerContainer

	// 使用ip命令获取连接到网桥的veth接口
	cmd := exec.Command("ip", "link", "show", "master", bridgeName)
	output, err := cmd.Output()
	if err != nil {
		// 如果命令失败，返回空列表而不是错误
		return containers, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "veth") {
			// 解析veth接口信息
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				vethName := strings.TrimSuffix(fields[1], ":")

				// 尝试获取veth接口的对端信息
				container, err := s.getContainerFromVeth(vethName)
				if err == nil {
					containers = append(containers, container)
				}
			}
		}
	}

	return containers, nil
}

// getContainerFromVeth 从veth接口获取容器信息
func (s *NetworkService) getContainerFromVeth(vethName string) (models.DockerContainer, error) {
	container := models.DockerContainer{}

	// 获取veth接口的IP地址
	cmd := exec.Command("ip", "addr", "show", vethName)
	output, err := cmd.Output()
	if err != nil {
		return container, err
	}

	// 解析IP地址（虽然veth接口通常没有IP，但我们可以尝试）
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "link/ether ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				container.MACAddress = fields[1]
			}
		}
	}

	// 设置基本信息
	container.ID = vethName
	container.Name = "container-" + vethName[4:] // 移除veth前缀

	return container, nil
}

// GetBridgeRules 获取指定网桥的iptables规则
func (s *NetworkService) GetBridgeRules(bridgeName string) ([]models.IPTablesRule, error) {
	log.Printf("[DEBUG] Getting rules for bridge: %s", bridgeName)

	ruleService := NewRuleService()
	allRules, err := ruleService.GetSystemRules()
	if err != nil {
		return nil, err
	}

	var bridgeRules []models.IPTablesRule
	for _, rule := range allRules {
		// 检查规则是否与指定网桥相关
		if s.isRuleRelatedToBridge(rule, bridgeName) {
			bridgeRules = append(bridgeRules, rule)
		}
	}

	log.Printf("[DEBUG] Found %d rules for bridge %s", len(bridgeRules), bridgeName)
	return bridgeRules, nil
}

// GetNetworkConnections 获取网络连接信息（使用netstat命令）
func (s *NetworkService) GetNetworkConnections() ([]models.NetworkConnection, error) {
	log.Println("[DEBUG] Getting network connections using netstat")

	cmd := exec.Command("netstat", "-tuln")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute netstat: %v", err)
	}

	var connections []models.NetworkConnection
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "tcp") || strings.HasPrefix(line, "udp") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				connection := models.NetworkConnection{
					Protocol:     fields[0],
					LocalAddress: fields[3],
				}
				if len(fields) >= 5 {
					connection.ForeignAddress = fields[4]
				}
				if len(fields) >= 6 {
					connection.State = fields[5]
				}
				connections = append(connections, connection)
			}
		}
	}

	log.Printf("[DEBUG] Found %d network connections", len(connections))
	return connections, nil
}

// GetRouteTable 获取路由表信息（使用ip route命令）
func (s *NetworkService) GetRouteTable() ([]models.RouteEntry, error) {
	log.Println("[DEBUG] Getting route table using ip route")

	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ip route: %v", err)
	}

	var routes []models.RouteEntry
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		route := s.parseRouteEntry(line)
		if route.Destination != "" {
			routes = append(routes, route)
		}
	}

	log.Printf("[DEBUG] Found %d route entries", len(routes))
	return routes, nil
}

// parseRouteEntry 解析路由条目
func (s *NetworkService) parseRouteEntry(line string) models.RouteEntry {
	fields := strings.Fields(line)
	route := models.RouteEntry{}

	if len(fields) == 0 {
		return route
	}

	// 解析目标网络
	if fields[0] == "default" {
		route.Destination = "0.0.0.0/0"
	} else {
		route.Destination = fields[0]
	}

	// 解析其他字段
	for i := 1; i < len(fields)-1; i++ {
		switch fields[i] {
		case "via":
			if i+1 < len(fields) {
				route.Gateway = fields[i+1]
			}
		case "dev":
			if i+1 < len(fields) {
				route.Interface = fields[i+1]
			}
		case "src":
			if i+1 < len(fields) {
				route.Source = fields[i+1]
			}
		case "metric":
			if i+1 < len(fields) {
				if metric, err := strconv.Atoi(fields[i+1]); err == nil {
					route.Metric = metric
				}
			}
		}
	}

	return route
}

// isRuleRelatedToBridge 判断规则是否与网桥相关
func (s *NetworkService) isRuleRelatedToBridge(rule models.IPTablesRule, bridgeName string) bool {
	// 检查输入/输出接口
	if rule.InInterface == bridgeName || rule.OutInterface == bridgeName {
		return true
	}

	// 检查规则文本中是否包含网桥名称
	if strings.Contains(rule.Extra, bridgeName) {
		return true
	}

	// 检查Docker相关的规则
	if bridgeName == "docker0" || strings.HasPrefix(bridgeName, "br-") {
		if strings.Contains(rule.ChainName, "DOCKER") ||
			strings.Contains(rule.Target, "DOCKER") ||
			strings.Contains(rule.Extra, "DOCKER") {
			return true
		}
	}

	return false
}
