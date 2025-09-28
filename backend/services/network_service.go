package services

import (
	"encoding/json"
	"fmt"
	"iptables-management-backend/models"
	"iptables-management-backend/utils"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
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

// GetTunnelInterfaceRules 获取与隧道接口相关的iptables规则
func (s *NetworkService) GetTunnelInterfaceRules(interfaceName string) ([]models.IPTablesRule, error) {
	log.Printf("[DEBUG] NetworkService.GetTunnelInterfaceRules called for interface: %s", interfaceName)

	var rules []models.IPTablesRule
	tables := []string{"raw", "mangle", "nat", "filter"}

	for _, tableName := range tables {
		cmd := exec.Command("iptables", "-t", tableName, "-L", "-n", "-v", "--line-numbers")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("[WARN] Failed to get rules from table %s: %v", tableName, err)
			continue
		}

		tableRules := s.parseTunnelRules(string(output), tableName, interfaceName)
		rules = append(rules, tableRules...)
	}

	log.Printf("[DEBUG] Found %d rules related to interface %s", len(rules), interfaceName)
	return rules, nil
}

// parseTunnelRules 解析与隧道接口相关的规则
func (s *NetworkService) parseTunnelRules(output, tableName, interfaceName string) []models.IPTablesRule {
	var rules []models.IPTablesRule
	lines := strings.Split(output, "\n")

	var currentChain string
	chainRegex := regexp.MustCompile(`^Chain\s+(\S+)\s+\(policy\s+(\S+).*\)`)
	ruleRegex := regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查链头部
		if chainMatch := chainRegex.FindStringSubmatch(line); chainMatch != nil {
			currentChain = chainMatch[1]
			continue
		}

		// 检查规则行是否包含指定接口
		if ruleMatch := ruleRegex.FindStringSubmatch(line); ruleMatch != nil && currentChain != "" {
			if strings.Contains(line, interfaceName) {
				lineNum, _ := strconv.Atoi(ruleMatch[1])
				packets, _ := strconv.ParseInt(ruleMatch[2], 10, 64)
				bytes, _ := strconv.ParseInt(ruleMatch[3], 10, 64)
				target := ruleMatch[4]
				protocol := ruleMatch[5]
				opt := ruleMatch[6]
				inInterface := ruleMatch[7]

				remaining := strings.TrimSpace(ruleMatch[8])
				parts := strings.Fields(remaining)

				var outInterface, source, destination, extra string
				if len(parts) > 0 {
					outInterface = parts[0]
				}
				if len(parts) > 1 {
					source = parts[1]
				}
				if len(parts) > 2 {
					destination = parts[2]
				}
				if len(parts) > 3 {
					extra = strings.Join(parts[3:], " ")
				}

				rule := models.IPTablesRule{
					Table:        tableName,
					ChainName:    currentChain,
					LineNumber:   lineNum,
					Target:       target,
					Protocol:     protocol,
					Source:       source,
					Destination:  destination,
					InInterface:  inInterface,
					OutInterface: outInterface,
					Options:      opt,
					Extra:        extra,
					Packets:      packets,
					Bytes:        bytes,
				}

				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// AnalyzeTunnelDockerCommunication 分析隧道接口与Docker网桥间的通信路径
func (s *NetworkService) AnalyzeTunnelDockerCommunication(tunnelInterface, dockerBridge string) (*models.TunnelDockerAnalysis, error) {
	log.Printf("[DEBUG] Analyzing communication between tunnel %s and docker bridge %s", tunnelInterface, dockerBridge)

	analysis := &models.TunnelDockerAnalysis{
		TunnelInterface: tunnelInterface,
		DockerBridge:    dockerBridge,
	}

	// 首先检查接口是否存在
	tunnelExists, err := s.checkInterfaceExists(tunnelInterface)
	if err != nil || !tunnelExists {
		return nil, fmt.Errorf("tunnel interface %s not found or not accessible", tunnelInterface)
	}

	bridgeExists, err := s.checkInterfaceExists(dockerBridge)
	if err != nil || !bridgeExists {
		return nil, fmt.Errorf("docker bridge %s not found or not accessible", dockerBridge)
	}

	// 获取接口IP信息
	var tunnelIPs []string

	// 对于tunnel接口，优先获取destination IP
	if strings.HasPrefix(tunnelInterface, "tun") || strings.HasPrefix(tunnelInterface, "tap") {
		log.Printf("[DEBUG] Detected tunnel interface %s, getting destination IP", tunnelInterface)

		// 首先尝试获取destination IP
		if destIP := s.getTunnelDestinationIP(tunnelInterface); destIP != "" {
			tunnelIPs = append(tunnelIPs, destIP)
			log.Printf("[DEBUG] Using tunnel destination IP: %s", destIP)
		}

		// 如果没有获取到destination IP，再尝试获取本地IP作为备选
		if len(tunnelIPs) == 0 {
			log.Printf("[DEBUG] No destination IP found, trying to get local IPs for %s", tunnelInterface)
			localIPs, err := s.getInterfaceIPs(tunnelInterface)
			if err != nil {
				log.Printf("[WARN] Failed to get tunnel interface local IPs: %v", err)
			} else {
				tunnelIPs = localIPs
				log.Printf("[DEBUG] Using tunnel local IPs: %v", tunnelIPs)
			}
		}
	} else {
		// 对于非tunnel接口，使用常规方法获取IP
		var err error
		tunnelIPs, err = s.getInterfaceIPs(tunnelInterface)
		if err != nil {
			log.Printf("[WARN] Failed to get interface IPs for %s: %v", tunnelInterface, err)
		}
	}

	bridgeIPs, err := s.getInterfaceIPs(dockerBridge)
	if err != nil {
		log.Printf("[WARN] Failed to get bridge interface IPs: %v", err)
	}

	log.Printf("[DEBUG] Final IP addresses - Tunnel (%s): %v, Bridge (%s): %v",
		tunnelInterface, tunnelIPs, dockerBridge, bridgeIPs)

	// 获取FORWARD链规则（精确匹配）
	forwardRules, err := s.getForwardRulesExact(tunnelInterface, dockerBridge)
	if err != nil {
		return nil, fmt.Errorf("failed to get forward rules: %v", err)
	}
	analysis.ForwardRules = forwardRules

	// 获取NAT规则（精确匹配）
	natRules, err := s.getNATRulesExact(tunnelInterface, dockerBridge)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT rules: %v", err)
	}
	analysis.NATRules = natRules

	// 获取Docker隔离规则（DOCKER-ISOLATION-STAGE-2链）
	isolationRules, err := s.getDockerIsolationRules(tunnelInterface, dockerBridge)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker isolation rules: %v", err)
	}
	analysis.IsolationRules = isolationRules

	// 实际连通性测试
	connectivityResult := s.testConnectivity(tunnelInterface, dockerBridge, tunnelIPs, bridgeIPs)

	// 生成通信路径（基于实际测试结果）
	analysis.CommunicationPath = s.generateCommunicationPathWithIsolation(tunnelInterface, dockerBridge, connectivityResult, isolationRules)

	// 计算统计信息（精确计算）
	analysis.Statistics = s.calculateTunnelDockerStatsExact(tunnelInterface, dockerBridge, forwardRules, natRules)

	// 生成建议（基于实际测试结果）
	analysis.Recommendations = s.generateRecommendationsWithTest(analysis, connectivityResult)

	return analysis, nil
}

// getForwardRules 获取FORWARD链中相关的规则
func (s *NetworkService) getForwardRulesExact(tunnelInterface, dockerBridge string) ([]models.IPTablesRule, error) {
	cmd := exec.Command("iptables", "-t", "filter", "-L", "FORWARD", "-n", "-v", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var rules []models.IPTablesRule
	lines := strings.Split(string(output), "\n")
	ruleRegex := regexp.MustCompile(`^\s*(\d+)\s+(\d+[KMG]?)\s+(\d+[KMG]?)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if ruleMatch := ruleRegex.FindStringSubmatch(line); ruleMatch != nil {
			lineNum, _ := strconv.Atoi(ruleMatch[1])
			packets := utils.ParsePacketCount(ruleMatch[2])
			bytes := utils.ParsePacketCount(ruleMatch[3])
			target := ruleMatch[4]
			protocol := ruleMatch[5]
			opt := ruleMatch[6]
			inInterface := ruleMatch[7]

			remaining := strings.TrimSpace(ruleMatch[8])
			parts := strings.Fields(remaining)

			var outInterface, source, destination, extra string
			if len(parts) > 0 {
				outInterface = parts[0]
			}
			if len(parts) > 1 {
				source = parts[1]
			}
			if len(parts) > 2 {
				destination = parts[2]
			}
			if len(parts) > 3 {
				extra = strings.Join(parts[3:], " ")
			}

			// 精确匹配：只包含涉及指定接口组合的规则
			isRelevant := false
			if (inInterface == tunnelInterface && outInterface == dockerBridge) ||
				(inInterface == dockerBridge && outInterface == tunnelInterface) ||
				(inInterface == tunnelInterface && outInterface == "*") ||
				(inInterface == "*" && outInterface == dockerBridge) {
				isRelevant = true
			}

			// 也检查规则文本中是否明确提到这两个接口
			if !isRelevant && (strings.Contains(line, tunnelInterface) && strings.Contains(line, dockerBridge)) {
				isRelevant = true
			}

			if isRelevant {
				rule := models.IPTablesRule{
					Table:        "filter",
					ChainName:    "FORWARD",
					LineNumber:   lineNum,
					Target:       target,
					Protocol:     protocol,
					Source:       source,
					Destination:  destination,
					InInterface:  inInterface,
					OutInterface: outInterface,
					Options:      opt,
					Extra:        extra,
					Packets:      int64(packets),
					Bytes:        int64(bytes),
				}

				rules = append(rules, rule)
			}
		}
	}

	return rules, nil
}

// generateCommunicationPath 生成通信路径
// 保留原有方法作为备用
func (s *NetworkService) generateCommunicationPath(tunnelInterface, dockerBridge string) []models.CommunicationStep {
	// 使用新的方法，但不进行实际测试
	connectivity := ConnectivityResult{}
	return s.generateCommunicationPathWithTest(tunnelInterface, dockerBridge, connectivity)
}

// generateCommunicationPathWithIsolation 生成包含隔离规则检查的通信路径
// analyzeIsolationRulesEffectiveness 分析隔离规则的有效性
func (s *NetworkService) analyzeIsolationRulesEffectiveness(isolationRules []models.IPTablesRule, tunnelInterface, dockerBridge string) IsolationAnalysisResult {
	result := IsolationAnalysisResult{}

	// 按行号排序规则，确保按执行顺序分析
	sortedRules := make([]models.IPTablesRule, len(isolationRules))
	copy(sortedRules, isolationRules)

	// 简单排序（按行号）
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[i].LineNumber > sortedRules[j].LineNumber {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	// 分析规则有效性
	hasEarlyReturn := false
	for _, rule := range sortedRules {
		if rule.Target == "RETURN" {
			// 检查是否是针对隧道接口的RETURN规则
			if (rule.InInterface == tunnelInterface || rule.InInterface == "any") &&
				(rule.OutInterface == dockerBridge || strings.HasPrefix(rule.OutInterface, "br-")) {
				hasEarlyReturn = true
				result.HasReturnRules = true
				log.Printf("[DEBUG] Found RETURN rule that bypasses isolation: line %d", rule.LineNumber)
				break
			}
		}
	}

	// 统计DROP规则
	for _, rule := range sortedRules {
		if rule.Target == "DROP" {
			// 检查是否影响当前通信路径
			affectsPath := false
			if (rule.InInterface == "any" || rule.InInterface == tunnelInterface) &&
				(rule.OutInterface == dockerBridge || strings.HasPrefix(rule.OutInterface, "br-")) {
				affectsPath = true
			}

			if affectsPath {
				if hasEarlyReturn {
					result.IneffectiveDrops++
				} else {
					result.EffectiveDrops++
				}
			}
		}
	}

	// 生成状态描述
	if result.EffectiveDrops > 0 {
		result.Status = fmt.Sprintf("检测到%d条有效DROP规则可能阻断通信", result.EffectiveDrops)
		result.Action = "隔离阻断"
	} else if result.IneffectiveDrops > 0 && result.HasReturnRules {
		result.Status = fmt.Sprintf("检测到%d条DROP规则，但已被RETURN规则覆盖", result.IneffectiveDrops)
		result.Action = "允许通过"
	} else if len(isolationRules) > 0 {
		result.Status = "存在隔离规则但不影响当前通信路径"
		result.Action = "允许通过"
	} else {
		result.Status = "无相关隔离规则"
		result.Action = "允许通过"
	}

	log.Printf("[DEBUG] Isolation analysis: %s (effective drops: %d, ineffective drops: %d, has returns: %v)",
		result.Status, result.EffectiveDrops, result.IneffectiveDrops, result.HasReturnRules)

	return result
}

func (s *NetworkService) generateCommunicationPathWithIsolation(tunnelInterface, dockerBridge string, connectivity ConnectivityResult, isolationRules []models.IPTablesRule) []models.CommunicationStep {
	steps := []models.CommunicationStep{
		{
			Step:        1,
			Description: fmt.Sprintf("数据包从%s接口进入", tunnelInterface),
			Table:       "raw",
			Chain:       "PREROUTING",
			Action:      "连接跟踪初始化",
			Interface:   tunnelInterface,
		},
		{
			Step:        2,
			Description: "包标记和修改处理",
			Table:       "mangle",
			Chain:       "PREROUTING",
			Action:      "包处理",
		},
		{
			Step:        3,
			Description: "DNAT规则检查",
			Table:       "nat",
			Chain:       "PREROUTING",
			Action:      "地址转换",
			Interface:   tunnelInterface,
		},
		{
			Step:        4,
			Description: "转发前包处理",
			Table:       "mangle",
			Chain:       "FORWARD",
			Action:      "包修改",
		},
		{
			Step: 5,
			Description: fmt.Sprintf("转发规则检查 (%s -> %s) - %s", tunnelInterface, dockerBridge, func() string {
				if connectivity.TunnelToBridge {
					return "允许转发"
				}
				return "转发被阻止"
			}()),
			Table:     "filter",
			Chain:     "FORWARD",
			Action:    "过滤决策",
			Interface: fmt.Sprintf("%s->%s", tunnelInterface, dockerBridge),
		},
	}

	// 智能分析Docker隔离规则的有效性
	isolationAnalysis := s.analyzeIsolationRulesEffectiveness(isolationRules, tunnelInterface, dockerBridge)

	steps = append(steps, models.CommunicationStep{
		Step:        6,
		Description: fmt.Sprintf("Docker隔离规则检查 (DOCKER-ISOLATION-STAGE-2) - %s", isolationAnalysis.Status),
		Table:       "filter",
		Chain:       "DOCKER-ISOLATION-STAGE-2",
		Action:      isolationAnalysis.Action,
		Interface:   fmt.Sprintf("%s->%s", tunnelInterface, dockerBridge),
	})

	// 继续原有的步骤
	steps = append(steps, []models.CommunicationStep{
		{
			Step:        7,
			Description: "转发后包处理",
			Table:       "mangle",
			Chain:       "POSTROUTING",
			Action:      "包修改",
		},
		{
			Step:        8,
			Description: "SNAT/MASQUERADE规则处理",
			Table:       "nat",
			Chain:       "POSTROUTING",
			Action:      "源地址转换",
			Interface:   dockerBridge,
		},
		{
			Step:        9,
			Description: fmt.Sprintf("数据包通过%s发送到目标 - 跳过连通性检查", dockerBridge),
			Table:       "output",
			Chain:       "OUTPUT",
			Action:      "包发送",
			Interface:   dockerBridge,
		},
	}...)

	// 如果有错误信息，添加具体的问题分析
	if connectivity.Error != "" {
		steps = append(steps, models.CommunicationStep{
			Step:        10,
			Description: fmt.Sprintf("规则分析问题: %s", connectivity.Error),
			Table:       "analysis",
			Chain:       "RULE_ANALYSIS",
			Action:      "问题定位",
		})
	}

	return steps
}

// calculateTunnelDockerStats 计算统计信息
func (s *NetworkService) calculateTunnelDockerStatsExact(tunnelInterface, dockerBridge string, forwardRules, natRules []models.IPTablesRule) models.TunnelDockerStats {
	var stats models.TunnelDockerStats

	// 精确计算特定接口间的统计信息
	for _, rule := range forwardRules {
		// 隧道到Docker的流量
		if rule.InInterface == tunnelInterface && (rule.OutInterface == dockerBridge || strings.HasPrefix(rule.OutInterface, "br-")) {
			stats.TunnelToDockerPackets += rule.Packets
			stats.TunnelToDockerBytes += rule.Bytes
			if rule.Target == "ACCEPT" {
				stats.ForwardedPackets += rule.Packets
			} else if rule.Target == "DROP" || rule.Target == "REJECT" {
				stats.DroppedPackets += rule.Packets
			}
		}

		// Docker到隧道的流量
		if (rule.InInterface == dockerBridge || strings.HasPrefix(rule.InInterface, "br-")) && rule.OutInterface == tunnelInterface {
			stats.DockerToTunnelPackets += rule.Packets
			stats.DockerToTunnelBytes += rule.Bytes
			if rule.Target == "ACCEPT" {
				stats.ForwardedPackets += rule.Packets
			} else if rule.Target == "DROP" || rule.Target == "REJECT" {
				stats.DroppedPackets += rule.Packets
			}
		}
	}

	// 如果没有精确匹配的规则，尝试从接口统计中获取数据
	if stats.TunnelToDockerPackets == 0 && stats.DockerToTunnelPackets == 0 {
		log.Printf("[DEBUG] No exact rule matches found, attempting to get interface statistics")
		tunnelStats, err := s.getInterfaceStatistics(tunnelInterface)
		if err == nil {
			// 使用接口统计作为参考（这是估算值）
			stats.TunnelToDockerPackets = tunnelStats.TxPackets / 10 // 估算10%流量到Docker
			stats.TunnelToDockerBytes = tunnelStats.TxBytes / 10
		}

		bridgeStats, err := s.getInterfaceStatistics(dockerBridge)
		if err == nil {
			stats.DockerToTunnelPackets = bridgeStats.TxPackets / 10
			stats.DockerToTunnelBytes = bridgeStats.TxBytes / 10
		}
	}

	return stats
}

// generateRecommendations 生成优化建议
// 检查接口是否存在
func (s *NetworkService) checkInterfaceExists(interfaceName string) (bool, error) {
	cmd := exec.Command("ip", "link", "show", interfaceName)
	err := cmd.Run()
	return err == nil, nil
}

// 获取接口IP地址
func (s *NetworkService) getInterfaceIPs(interfaceName string) ([]string, error) {
	cmd := exec.Command("ip", "addr", "show", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ips []string
	ipRegex := regexp.MustCompile(`inet\s+([0-9.]+/[0-9]+)`)
	matches := ipRegex.FindAllStringSubmatch(string(output), -1)
	for _, match := range matches {
		if len(match) > 1 {
			ips = append(ips, strings.Split(match[1], "/")[0])
		}
	}

	return ips, nil
}

// 连通性测试结果
type ConnectivityResult struct {
	TunnelToBridge bool
	BridgeToTunnel bool
	Error          string
}

// IsolationAnalysisResult 隔离规则分析结果
type IsolationAnalysisResult struct {
	Status           string // 状态描述
	Action           string // 动作描述
	EffectiveDrops   int    // 有效的DROP规则数量
	IneffectiveDrops int    // 无效的DROP规则数量（被RETURN规则覆盖）
	HasReturnRules   bool   // 是否存在RETURN规则
}

// hping3测试配置

// hping3测试结果

// 规则匹配结果
type ForwardRuleMatch struct {
	Found        bool     `json:"found"`
	MatchedRules []string `json:"matched_rules"`
	Details      string   `json:"details"`
}

// 解析后的规则详情
type ParsedForwardRule struct {
	LineNumber      int               `json:"line_number"`
	Target          string            `json:"target"`
	Protocol        string            `json:"protocol"`
	InInterface     string            `json:"in_interface"`
	OutInterface    string            `json:"out_interface"`
	Source          string            `json:"source"`
	Destination     string            `json:"destination"`
	ConntrackState  string            `json:"conntrack_state"`
	ExtraConditions map[string]string `json:"extra_conditions"`
	RawRule         string            `json:"raw_rule"`
}

// 测试连通性 - 使用hping3进行高级网络测试
func (s *NetworkService) testConnectivity(tunnelInterface, dockerBridge string, tunnelIPs, bridgeIPs []string) ConnectivityResult {
	// 获取调用者信息和时间戳
	pc, _, _, _ := runtime.Caller(1)
	callerFunc := runtime.FuncForPC(pc).Name()
	startTime := time.Now()

	// 记录方法调用和输入参数
	inputParams := map[string]interface{}{
		"tunnelInterface": tunnelInterface,
		"dockerBridge":    dockerBridge,
		"tunnelIPs":       tunnelIPs,
		"bridgeIPs":       bridgeIPs,
		"caller":          callerFunc,
		"timestamp":       startTime.Format(time.RFC3339Nano),
		"os":              runtime.GOOS,
		"arch":            runtime.GOARCH,
	}

	inputJSON, _ := json.Marshal(inputParams)
	log.Printf("[DEBUG] testConnectivity: Method called with parameters: %s", string(inputJSON))

	result := ConnectivityResult{}

	// 检查输入参数有效性
	if tunnelInterface == "" || dockerBridge == "" {
		result.Error = "Invalid input parameters: interface names cannot be empty"
		log.Printf("[DEBUG] testConnectivity: Input validation failed - empty interface names")
		return result
	}

	// 检查当前用户权限
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "unknown"
	}
	log.Printf("[DEBUG] testConnectivity: Running as user: %s", currentUser)

	// 检查是否有足够权限执行网络命令
	if os.Geteuid() != 0 {
		log.Printf("[DEBUG] testConnectivity: Warning - Not running as root (UID: %d), some network tests may fail", os.Geteuid())
	}

	// 跳过路由检查和ping检查步骤，直接进行iptables规则检查
	log.Printf("[DEBUG] testConnectivity: Skipping route and ping checks, proceeding directly to iptables rule analysis")

	// iptables规则检查 - 使用改进的规则匹配机制
	log.Printf("[DEBUG] testConnectivity: Step 1 - Checking iptables FORWARD rules using improved matching")

	// 获取完整的FORWARD规则列表
	forwardRules, forwardErr := s.getForwardRulesExact(tunnelInterface, dockerBridge)
	if forwardErr != nil {
		log.Printf("[DEBUG] testConnectivity: Failed to get FORWARD rules: %v", forwardErr)
		result.Error = fmt.Sprintf("Failed to retrieve FORWARD rules: %v", forwardErr)
	} else {
		log.Printf("[DEBUG] testConnectivity: Retrieved %d FORWARD rules for analysis", len(forwardRules))

		// 检查隧道到网桥的转发规则
		log.Printf("[DEBUG] testConnectivity: Analyzing FORWARD rules: %s -> %s", tunnelInterface, dockerBridge)
		tunnelToBridgeMatch := s.checkForwardRuleMatch(forwardRules, tunnelInterface, dockerBridge, "tunnel_to_bridge")
		result.TunnelToBridge = tunnelToBridgeMatch.Found

		forwardResult1 := map[string]interface{}{
			"direction":     "tunnel_to_bridge",
			"method":        "rule_analysis",
			"rules_checked": len(forwardRules),
			"found":         tunnelToBridgeMatch.Found,
			"matched_rules": tunnelToBridgeMatch.MatchedRules,
			"details":       tunnelToBridgeMatch.Details,
		}

		if tunnelToBridgeMatch.Found {
			log.Printf("[DEBUG] testConnectivity: FORWARD rule found (%s -> %s): %d matching rules",
				tunnelInterface, dockerBridge, len(tunnelToBridgeMatch.MatchedRules))
			for i, rule := range tunnelToBridgeMatch.MatchedRules {
				log.Printf("[DEBUG] testConnectivity: Matched rule %d: %s", i+1, rule)
			}
		} else {
			log.Printf("[DEBUG] testConnectivity: No FORWARD rule found (%s -> %s)", tunnelInterface, dockerBridge)
		}

		// 检查网桥到隧道的转发规则
		log.Printf("[DEBUG] testConnectivity: Analyzing FORWARD rules: %s -> %s", dockerBridge, tunnelInterface)
		bridgeToTunnelMatch := s.checkForwardRuleMatch(forwardRules, dockerBridge, tunnelInterface, "bridge_to_tunnel")
		result.BridgeToTunnel = bridgeToTunnelMatch.Found

		forwardResult2 := map[string]interface{}{
			"direction":     "bridge_to_tunnel",
			"method":        "rule_analysis",
			"rules_checked": len(forwardRules),
			"found":         bridgeToTunnelMatch.Found,
			"matched_rules": bridgeToTunnelMatch.MatchedRules,
			"details":       bridgeToTunnelMatch.Details,
		}

		if bridgeToTunnelMatch.Found {
			log.Printf("[DEBUG] testConnectivity: FORWARD rule found (%s -> %s): %d matching rules",
				dockerBridge, tunnelInterface, len(bridgeToTunnelMatch.MatchedRules))
			for i, rule := range bridgeToTunnelMatch.MatchedRules {
				log.Printf("[DEBUG] testConnectivity: Matched rule %d: %s", i+1, rule)
			}
		} else {
			log.Printf("[DEBUG] testConnectivity: No FORWARD rule found (%s -> %s)", dockerBridge, tunnelInterface)
		}

		// 记录iptables规则检查结果
		iptablesResults := []interface{}{forwardResult1, forwardResult2}
		iptablesJSON, _ := json.Marshal(iptablesResults)
		log.Printf("[DEBUG] testConnectivity: iptables rules check results: %s", string(iptablesJSON))
	}

	// 最终结果汇总
	totalDuration := time.Since(startTime)
	finalResult := map[string]interface{}{
		"tunnelInterface": tunnelInterface,
		"dockerBridge":    dockerBridge,
		"tunnelToBridge":  result.TunnelToBridge,
		"bridgeToTunnel":  result.BridgeToTunnel,
		"error":           result.Error,
		"totalDuration":   totalDuration.String(),
		"timestamp":       time.Now().Format(time.RFC3339Nano),
		"note":            "Route and ping checks have been disabled",
	}

	finalJSON, _ := json.Marshal(finalResult)
	log.Printf("[DEBUG] testConnectivity: Final connectivity test result: %s", string(finalJSON))

	// 性能统计
	if totalDuration > 10*time.Second {
		log.Printf("[DEBUG] testConnectivity: Warning - Connectivity test took longer than expected: %v", totalDuration)
	}

	log.Printf("[DEBUG] testConnectivity: Method completed in %v", totalDuration)

	return result
}

// 获取接口统计信息
type InterfaceStats struct {
	RxPackets int64
	TxPackets int64
	RxBytes   int64
	TxBytes   int64
}

func (s *NetworkService) getInterfaceStatistics(interfaceName string) (*InterfaceStats, error) {
	cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_packets", interfaceName))
	rxPacketsOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_packets", interfaceName))
	txPacketsOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", interfaceName))
	rxBytesOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", interfaceName))
	txBytesOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	rxPackets, _ := strconv.ParseInt(strings.TrimSpace(string(rxPacketsOutput)), 10, 64)
	txPackets, _ := strconv.ParseInt(strings.TrimSpace(string(txPacketsOutput)), 10, 64)
	rxBytes, _ := strconv.ParseInt(strings.TrimSpace(string(rxBytesOutput)), 10, 64)
	txBytes, _ := strconv.ParseInt(strings.TrimSpace(string(txBytesOutput)), 10, 64)

	return &InterfaceStats{
		RxPackets: rxPackets,
		TxPackets: txPackets,
		RxBytes:   rxBytes,
		TxBytes:   txBytes,
	}, nil
}

// 精确获取NAT规则
func (s *NetworkService) getNATRulesExact(tunnelInterface, dockerBridge string) ([]models.IPTablesRule, error) {
	var allRules []models.IPTablesRule
	chains := []string{"PREROUTING", "POSTROUTING", "OUTPUT"}

	for _, chain := range chains {
		cmd := exec.Command("iptables", "-t", "nat", "-L", chain, "-n", "-v", "--line-numbers")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("[WARN] Failed to get NAT rules from chain %s: %v", chain, err)
			continue
		}

		chainRules := s.parseNATChainRulesExact(string(output), chain, tunnelInterface, dockerBridge)
		allRules = append(allRules, chainRules...)
	}

	return allRules, nil
}

// getDockerIsolationRules 获取DOCKER-ISOLATION-STAGE-2链规则
func (s *NetworkService) getDockerIsolationRules(tunnelInterface, dockerBridge string) ([]models.IPTablesRule, error) {
	cmd := exec.Command("iptables", "-t", "filter", "-L", "DOCKER-ISOLATION-STAGE-2", "-n", "-v", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[WARN] Failed to get DOCKER-ISOLATION-STAGE-2 rules: %v", err)
		// 如果链不存在，返回空规则列表而不是错误
		return []models.IPTablesRule{}, nil
	}

	rules := s.parseDockerIsolationRules(string(output), tunnelInterface, dockerBridge)
	log.Printf("[DEBUG] Found %d DOCKER-ISOLATION-STAGE-2 rules related to %s and %s", len(rules), tunnelInterface, dockerBridge)

	return rules, nil
}

// parseDockerIsolationRules 解析Docker隔离规则
func (s *NetworkService) parseDockerIsolationRules(output, tunnelInterface, dockerBridge string) []models.IPTablesRule {
	var rules []models.IPTablesRule
	lines := strings.Split(output, "\n")
	ruleRegex := regexp.MustCompile(`^\s*(\d+)\s+(\d+[KMG]?)\s+(\d+[KMG]?)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if ruleMatch := ruleRegex.FindStringSubmatch(line); ruleMatch != nil {
			lineNum, _ := strconv.Atoi(ruleMatch[1])
			packets := utils.ParsePacketCount(ruleMatch[2])
			bytes := utils.ParsePacketCount(ruleMatch[3])
			target := ruleMatch[4]
			protocol := ruleMatch[5]
			opt := ruleMatch[6]
			inInterface := ruleMatch[7]

			remaining := strings.TrimSpace(ruleMatch[8])
			parts := strings.Fields(remaining)

			var outInterface, source, destination, extra string
			if len(parts) > 0 {
				outInterface = parts[0]
			}
			if len(parts) > 1 {
				source = parts[1]
			}
			if len(parts) > 2 {
				destination = parts[2]
			}
			if len(parts) > 3 {
				extra = strings.Join(parts[3:], " ")
			}

			// 扩展相关性检查，包括所有可能影响通信的规则
			isRelevant := false

			// 检查所有类型的规则（DROP、RETURN、ACCEPT等）
			if target == "DROP" || target == "RETURN" || target == "ACCEPT" {
				// 检查是否涉及指定的Docker网桥
				if strings.Contains(line, dockerBridge) || outInterface == dockerBridge || inInterface == dockerBridge {
					isRelevant = true
					log.Printf("[DEBUG] Found %s rule affecting bridge %s: %s", target, dockerBridge, line)
				}
			}

			if isRelevant {
				rule := models.IPTablesRule{
					Table:        "filter",
					ChainName:    "DOCKER-ISOLATION-STAGE-2",
					LineNumber:   lineNum,
					Target:       target,
					Protocol:     protocol,
					Source:       source,
					Destination:  destination,
					InInterface:  inInterface,
					OutInterface: outInterface,
					Options:      opt,
					Extra:        extra,
					Packets:      int64(packets),
					Bytes:        int64(bytes),
				}

				rules = append(rules, rule)
			}
		}
	}

	log.Printf("[DEBUG] Found %d relevant isolation rules for %s <-> %s", len(rules), tunnelInterface, dockerBridge)
	return rules
}

func (s *NetworkService) parseNATChainRulesExact(output, chain, tunnelInterface, dockerBridge string) []models.IPTablesRule {
	var rules []models.IPTablesRule
	lines := strings.Split(output, "\n")
	ruleRegex := regexp.MustCompile(`^\s*(\d+)\s+(\d+[KMG]?)\s+(\d+[KMG]?)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if ruleMatch := ruleRegex.FindStringSubmatch(line); ruleMatch != nil {
			lineNum, _ := strconv.Atoi(ruleMatch[1])
			packets := utils.ParsePacketCount(ruleMatch[2])
			bytes := utils.ParsePacketCount(ruleMatch[3])
			target := ruleMatch[4]
			protocol := ruleMatch[5]
			opt := ruleMatch[6]
			inInterface := ruleMatch[7]

			remaining := strings.TrimSpace(ruleMatch[8])
			parts := strings.Fields(remaining)

			var outInterface, source, destination, extra string
			if len(parts) > 0 {
				outInterface = parts[0]
			}
			if len(parts) > 1 {
				source = parts[1]
			}
			if len(parts) > 2 {
				destination = parts[2]
			}
			if len(parts) > 3 {
				extra = strings.Join(parts[3:], " ")
			}

			// 精确匹配：只包含涉及指定接口的NAT规则
			isRelevant := false
			if strings.Contains(line, tunnelInterface) || strings.Contains(line, dockerBridge) {
				isRelevant = true
			}

			// 检查MASQUERADE规则
			if target == "MASQUERADE" && (outInterface == tunnelInterface || inInterface == tunnelInterface) {
				isRelevant = true
			}

			if isRelevant {
				rule := models.IPTablesRule{
					Table:        "nat",
					ChainName:    chain,
					LineNumber:   lineNum,
					Target:       target,
					Protocol:     protocol,
					Source:       source,
					Destination:  destination,
					InInterface:  inInterface,
					OutInterface: outInterface,
					Options:      opt,
					Extra:        extra,
					Packets:      int64(packets),
					Bytes:        int64(bytes),
				}

				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// 基于实际测试结果生成通信路径
func (s *NetworkService) generateCommunicationPathWithTest(tunnelInterface, dockerBridge string, connectivity ConnectivityResult) []models.CommunicationStep {
	steps := []models.CommunicationStep{
		{
			Step:        1,
			Description: fmt.Sprintf("数据包从%s接口进入", tunnelInterface),
			Table:       "raw",
			Chain:       "PREROUTING",
			Action:      "连接跟踪初始化",
			Interface:   tunnelInterface,
		},
		{
			Step:        2,
			Description: "包标记和修改处理",
			Table:       "mangle",
			Chain:       "PREROUTING",
			Action:      "包处理",
		},
		{
			Step:        3,
			Description: "DNAT规则检查",
			Table:       "nat",
			Chain:       "PREROUTING",
			Action:      "地址转换",
			Interface:   tunnelInterface,
		},
		{
			Step:        4,
			Description: "路由决策 - 跳过路由检查",
			Table:       "routing",
			Chain:       "ROUTING_DECISION",
			Action:      "路由查找",
		},
		{
			Step:        5,
			Description: "转发前包处理",
			Table:       "mangle",
			Chain:       "FORWARD",
			Action:      "包修改",
		},
		{
			Step: 6,
			Description: fmt.Sprintf("转发规则检查 (%s -> %s) - %s", tunnelInterface, dockerBridge, func() string {
				if connectivity.TunnelToBridge {
					return "允许转发"
				}
				return "转发被阻止"
			}()),
			Table:     "filter",
			Chain:     "FORWARD",
			Action:    "过滤决策",
			Interface: fmt.Sprintf("%s->%s", tunnelInterface, dockerBridge),
		},
		{
			Step:        7,
			Description: "转发后包处理",
			Table:       "mangle",
			Chain:       "POSTROUTING",
			Action:      "包修改",
		},
		{
			Step:        8,
			Description: "SNAT/MASQUERADE规则处理",
			Table:       "nat",
			Chain:       "POSTROUTING",
			Action:      "源地址转换",
			Interface:   dockerBridge,
		},
		{
			Step:        9,
			Description: fmt.Sprintf("数据包通过%s发送到目标 - 跳过连通性检查", dockerBridge),
			Table:       "output",
			Chain:       "OUTPUT",
			Action:      "包发送",
			Interface:   dockerBridge,
		},
	}

	// 如果有错误信息，添加错误信息
	if connectivity.Error != "" {
		steps = append(steps, models.CommunicationStep{
			Step:        10,
			Description: fmt.Sprintf("规则分析失败: %s", connectivity.Error),
			Table:       "error",
			Chain:       "RULE_ANALYSIS",
			Action:      "分析失败",
		})
	}

	return steps
}

// 基于实际测试结果生成建议
func (s *NetworkService) generateRecommendationsWithTest(analysis *models.TunnelDockerAnalysis, connectivity ConnectivityResult) []string {
	var recommendations []string

	// 基于iptables规则分析结果的建议
	if !connectivity.TunnelToBridge {
		recommendations = append(recommendations,
			fmt.Sprintf("缺少%s到%s的转发规则，建议执行: iptables -I FORWARD 1 -i %s -o %s -j ACCEPT",
				analysis.TunnelInterface, analysis.DockerBridge, analysis.TunnelInterface, analysis.DockerBridge))
	}

	if !connectivity.BridgeToTunnel {
		recommendations = append(recommendations,
			fmt.Sprintf("缺少%s到%s的返回路径规则，建议执行: iptables -I FORWARD 2 -i %s -o %s -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT",
				analysis.DockerBridge, analysis.TunnelInterface, analysis.DockerBridge, analysis.TunnelInterface))
	}

	if connectivity.TunnelToBridge && connectivity.BridgeToTunnel {
		recommendations = append(recommendations, "iptables规则检查通过，转发规则配置正常")
	} else {
		// 检查是否有MASQUERADE规则
		hasMasqueradeRule := false
		for _, rule := range analysis.NATRules {
			if rule.Target == "MASQUERADE" && strings.Contains(rule.OutInterface, analysis.TunnelInterface) {
				hasMasqueradeRule = true
				break
			}
		}

		if !hasMasqueradeRule {
			recommendations = append(recommendations,
				fmt.Sprintf("缺少MASQUERADE规则，建议执行: iptables -t nat -A POSTROUTING -o %s -j MASQUERADE",
					analysis.TunnelInterface))
		}
	}

	// 检查丢包率
	if analysis.Statistics.DroppedPackets > 0 {
		totalPackets := analysis.Statistics.ForwardedPackets + analysis.Statistics.DroppedPackets
		if totalPackets > 0 {
			dropRate := float64(analysis.Statistics.DroppedPackets) / float64(totalPackets) * 100
			if dropRate > 5.0 {
				recommendations = append(recommendations,
					fmt.Sprintf("检测到较高的丢包率(%.2f%%)，建议检查防火墙规则配置", dropRate))
			}
		}
	}

	// 性能优化建议
	if len(analysis.ForwardRules) > 20 {
		recommendations = append(recommendations, "规则数量较多，建议优化规则顺序，将常用规则前置")
	}

	// 如果没有找到相关规则，提供基本配置建议
	if len(analysis.ForwardRules) == 0 {
		recommendations = append(recommendations, "未找到相关的FORWARD规则，请检查iptables配置")
	}

	if len(analysis.NATRules) == 0 {
		recommendations = append(recommendations, "未找到相关的NAT规则，可能需要配置MASQUERADE规则")
	}

	// 智能分析Docker隔离规则
	if len(analysis.IsolationRules) > 0 {
		isolationAnalysis := s.analyzeIsolationRulesEffectiveness(analysis.IsolationRules, analysis.TunnelInterface, analysis.DockerBridge)

		// 只有当存在有效的DROP规则时才提供相关建议
		if isolationAnalysis.EffectiveDrops > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("⚠️ 检测到%d条有效的Docker隔离DROP规则正在阻断通信", isolationAnalysis.EffectiveDrops))
			recommendations = append(recommendations,
				fmt.Sprintf("建议添加Docker隔离绕过规则: iptables -I DOCKER-ISOLATION-STAGE-2 1 -i %s -o %s -j RETURN",
					analysis.TunnelInterface, analysis.DockerBridge))
		} else if isolationAnalysis.IneffectiveDrops > 0 && isolationAnalysis.HasReturnRules {
			recommendations = append(recommendations,
				fmt.Sprintf("✅ 检测到%d条DROP规则，但已被RETURN规则有效覆盖，隔离规则不影响当前通信", isolationAnalysis.IneffectiveDrops))
		} else {
			recommendations = append(recommendations, "Docker隔离规则配置正常，不影响当前通信路径")
		}
	} else {
		log.Printf("[DEBUG] No Docker isolation rules found for analysis")
	}

	return recommendations
}

// parsePacketCount 解析包数，支持K/M/G单位
func (s *NetworkService) parsePacketCount(countStr string) int64 {
	if countStr == "" || countStr == "--" {
		return 0
	}

	// 移除可能的逗号
	countStr = strings.ReplaceAll(countStr, ",", "")

	// 检查是否有单位后缀
	if len(countStr) == 0 {
		return 0
	}

	lastChar := countStr[len(countStr)-1]
	var multiplier int64 = 1
	var numStr string

	switch lastChar {
	case 'K', 'k':
		multiplier = 1000
		numStr = countStr[:len(countStr)-1]
	case 'M', 'm':
		multiplier = 1000000
		numStr = countStr[:len(countStr)-1]
	case 'G', 'g':
		multiplier = 1000000000
		numStr = countStr[:len(countStr)-1]
	default:
		// 没有单位，直接解析
		numStr = countStr
	}

	// 解析数值部分
	if numStr == "" {
		return 0
	}

	// 支持小数点（如 2.3K）
	if strings.Contains(numStr, ".") {
		floatVal, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			log.Printf("[WARN] Failed to parse float packet count '%s': %v", numStr, err)
			return 0
		}
		return int64(floatVal * float64(multiplier))
	}

	// 整数解析
	intVal, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		log.Printf("[WARN] Failed to parse packet count '%s': %v", numStr, err)
		return 0
	}

	return intVal * multiplier
}

// checkForwardRuleMatch 检查FORWARD规则匹配
func (s *NetworkService) checkForwardRuleMatch(rules []models.IPTablesRule, inInterface, outInterface, direction string) ForwardRuleMatch {
	match := ForwardRuleMatch{
		Found:        false,
		MatchedRules: []string{},
		Details:      "",
	}

	log.Printf("[DEBUG] checkForwardRuleMatch: Checking %d rules for %s (%s -> %s)",
		len(rules), direction, inInterface, outInterface)

	var matchDetails []string

	for i, rule := range rules {
		log.Printf("[TRACE] checkForwardRuleMatch: Rule %d - In:%s Out:%s Target:%s Extra:%s",
			i+1, rule.InInterface, rule.OutInterface, rule.Target, rule.Extra)

		// 解析规则详情
		parsedRule := s.parseForwardRule(rule)

		// 检查是否匹配条件
		if s.ruleMatchesCriteria(parsedRule, inInterface, outInterface) {
			match.Found = true
			ruleDesc := fmt.Sprintf("Line %d: %s %s -i %s -o %s",
				rule.LineNumber, rule.Target, rule.Protocol, rule.InInterface, rule.OutInterface)

			// 添加额外条件信息
			if parsedRule.ConntrackState != "" {
				ruleDesc += fmt.Sprintf(" --ctstate %s", parsedRule.ConntrackState)
			}

			if len(parsedRule.ExtraConditions) > 0 {
				for key, value := range parsedRule.ExtraConditions {
					ruleDesc += fmt.Sprintf(" %s %s", key, value)
				}
			}

			match.MatchedRules = append(match.MatchedRules, ruleDesc)
			matchDetails = append(matchDetails, fmt.Sprintf("Rule %d matches criteria", rule.LineNumber))

			log.Printf("[DEBUG] checkForwardRuleMatch: Rule %d matches - %s", rule.LineNumber, ruleDesc)
		}
	}

	if match.Found {
		match.Details = fmt.Sprintf("Found %d matching rules: %s",
			len(match.MatchedRules), strings.Join(matchDetails, "; "))
		log.Printf("[DEBUG] checkForwardRuleMatch: Match successful for %s - %s", direction, match.Details)
	} else {
		match.Details = fmt.Sprintf("No matching rules found for %s -> %s", inInterface, outInterface)
		log.Printf("[DEBUG] checkForwardRuleMatch: No matches found for %s", direction)
	}

	return match
}

// parseForwardRule 解析FORWARD规则详情
func (s *NetworkService) parseForwardRule(rule models.IPTablesRule) ParsedForwardRule {
	parsed := ParsedForwardRule{
		LineNumber:      rule.LineNumber,
		Target:          rule.Target,
		Protocol:        rule.Protocol,
		InInterface:     rule.InInterface,
		OutInterface:    rule.OutInterface,
		Source:          rule.Source,
		Destination:     rule.Destination,
		ExtraConditions: make(map[string]string),
		RawRule:         fmt.Sprintf("%s %s %s %s %s %s %s", rule.Target, rule.Protocol, rule.InInterface, rule.OutInterface, rule.Source, rule.Destination, rule.Extra),
	}

	// 解析额外条件
	if rule.Extra != "" {
		extraParts := strings.Fields(rule.Extra)
		for i := 0; i < len(extraParts); i++ {
			switch extraParts[i] {
			case "ctstate":
				if i+1 < len(extraParts) {
					parsed.ConntrackState = extraParts[i+1]
					i++ // 跳过下一个参数
				}
			case "--ctstate":
				if i+1 < len(extraParts) {
					parsed.ConntrackState = extraParts[i+1]
					i++ // 跳过下一个参数
				}
			case "-m":
				if i+1 < len(extraParts) {
					parsed.ExtraConditions["module"] = extraParts[i+1]
					i++ // 跳过下一个参数
				}
			case "--dport":
				if i+1 < len(extraParts) {
					parsed.ExtraConditions["dport"] = extraParts[i+1]
					i++ // 跳过下一个参数
				}
			case "--sport":
				if i+1 < len(extraParts) {
					parsed.ExtraConditions["sport"] = extraParts[i+1]
					i++ // 跳过下一个参数
				}
			}
		}
	}

	log.Printf("[TRACE] parseForwardRule: Parsed rule %d - ConntrackState:%s, ExtraConditions:%v",
		parsed.LineNumber, parsed.ConntrackState, parsed.ExtraConditions)

	return parsed
}

// ruleMatchesCriteria 检查规则是否匹配指定条件
func (s *NetworkService) ruleMatchesCriteria(rule ParsedForwardRule, inInterface, outInterface string) bool {
	// 基本接口匹配检查
	interfaceMatch := false

	// 精确匹配
	if rule.InInterface == inInterface && rule.OutInterface == outInterface {
		interfaceMatch = true
		log.Printf("[TRACE] ruleMatchesCriteria: Exact interface match - In:%s Out:%s", inInterface, outInterface)
	}

	// 通配符匹配
	if rule.InInterface == inInterface && (rule.OutInterface == "*" || rule.OutInterface == "any") {
		interfaceMatch = true
		log.Printf("[TRACE] ruleMatchesCriteria: Wildcard out interface match - In:%s Out:*", inInterface)
	}

	if (rule.InInterface == "*" || rule.InInterface == "any") && rule.OutInterface == outInterface {
		interfaceMatch = true
		log.Printf("[TRACE] ruleMatchesCriteria: Wildcard in interface match - In:* Out:%s", outInterface)
	}

	// 检查目标动作是否为ACCEPT（允许转发）
	targetMatch := (rule.Target == "ACCEPT")

	// 协议匹配（all表示所有协议）
	protocolMatch := (rule.Protocol == "all" || rule.Protocol == "")

	// 综合判断
	matches := interfaceMatch && targetMatch && protocolMatch

	log.Printf("[TRACE] ruleMatchesCriteria: Rule %d evaluation - Interface:%v Target:%v Protocol:%v => Final:%v",
		rule.LineNumber, interfaceMatch, targetMatch, protocolMatch, matches)

	// 如果基本条件匹配，还需要检查是否有阻止性的额外条件
	if matches {
		// 检查conntrack状态 - RELATED,ESTABLISHED是常见的允许状态
		if rule.ConntrackState != "" {
			validStates := []string{"RELATED,ESTABLISHED", "ESTABLISHED,RELATED", "NEW,RELATED,ESTABLISHED"}
			stateValid := false
			for _, validState := range validStates {
				if rule.ConntrackState == validState {
					stateValid = true
					break
				}
			}
			if !stateValid {
				log.Printf("[TRACE] ruleMatchesCriteria: Rule %d rejected due to restrictive conntrack state: %s",
					rule.LineNumber, rule.ConntrackState)
				return false
			}
		}

		log.Printf("[DEBUG] ruleMatchesCriteria: Rule %d fully matches criteria", rule.LineNumber)
	}

	return matches
}

// checkHpingAvailability 检查hping3是否可用
func (s *NetworkService) checkHpingAvailability() bool {
	log.Printf("[DEBUG] checkHpingAvailability: Checking if hping3 is installed")

	// 检查hping3命令是否存在
	cmd := exec.Command("which", "hping3")
	err := cmd.Run()
	if err != nil {
		log.Printf("[DEBUG] checkHpingAvailability: hping3 not found in PATH")

		// 尝试常见的安装路径
		commonPaths := []string{
			"/usr/bin/hping3",
			"/usr/sbin/hping3",
			"/usr/local/bin/hping3",
			"/usr/local/sbin/hping3",
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				log.Printf("[DEBUG] checkHpingAvailability: Found hping3 at %s", path)
				return true
			}
		}

		log.Printf("[DEBUG] checkHpingAvailability: hping3 not found. Install with: apt-get install hping3 (Ubuntu/Debian) or yum install hping3 (CentOS/RHEL)")
		return false
	}

	// 检查版本信息
	versionCmd := exec.Command("hping3", "--version")
	output, err := versionCmd.CombinedOutput()
	if err != nil {
		log.Printf("[DEBUG] checkHpingAvailability: hping3 found but version check failed: %v", err)
		return true // 即使版本检查失败，也认为可用
	}

	log.Printf("[DEBUG] checkHpingAvailability: hping3 available - %s", strings.TrimSpace(string(output)))
	return true
}

// getTunnelDestinationIP 获取tunnel接口的destination IP
func (s *NetworkService) getTunnelDestinationIP(tunnelInterface string) string {
	log.Printf("[DEBUG] getTunnelDestinationIP: Getting destination IP for %s", tunnelInterface)

	// 方法1: 尝试使用 ip addr show 命令
	cmd := exec.Command("ip", "addr", "show", tunnelInterface)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[DEBUG] getTunnelDestinationIP: Failed to get interface info with ip command: %v", err)
	} else {
		log.Printf("[DEBUG] getTunnelDestinationIP: ip addr output: %s", string(output))

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "inet ") && strings.Contains(line, "peer") {
				// 解析 "inet 192.168.252.1 peer 192.168.252.2/32" 格式
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "peer" && i+1 < len(parts) {
						peerAddr := strings.Split(parts[i+1], "/")[0]
						log.Printf("[DEBUG] getTunnelDestinationIP: Found peer destination IP: %s", peerAddr)
						return peerAddr
					}
				}
			}
		}
	}

	// 方法2: 尝试使用 ifconfig 命令 (适用于显示destination的情况)
	log.Printf("[DEBUG] getTunnelDestinationIP: Trying ifconfig command for %s", tunnelInterface)
	ifconfigCmd := exec.Command("ifconfig", tunnelInterface)
	ifconfigOutput, ifconfigErr := ifconfigCmd.Output()
	if ifconfigErr != nil {
		log.Printf("[DEBUG] getTunnelDestinationIP: Failed to get interface info with ifconfig: %v", ifconfigErr)
	} else {
		log.Printf("[DEBUG] getTunnelDestinationIP: ifconfig output: %s", string(ifconfigOutput))

		lines := strings.Split(string(ifconfigOutput), "\n")
		for _, line := range lines {
			// 解析 "inet 192.168.252.1  netmask 255.255.255.255  destination 192.168.252.2" 格式
			if strings.Contains(line, "inet ") && strings.Contains(line, "destination") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "destination" && i+1 < len(parts) {
						destAddr := parts[i+1]
						log.Printf("[DEBUG] getTunnelDestinationIP: Found destination IP: %s", destAddr)
						return destAddr
					}
				}
			}
		}
	}

	// 方法3: 尝试从 /proc/net/route 获取点对点接口的目标地址
	log.Printf("[DEBUG] getTunnelDestinationIP: Trying to get destination from route table")
	routeCmd := exec.Command("ip", "route", "show", "dev", tunnelInterface)
	routeOutput, routeErr := routeCmd.Output()
	if routeErr != nil {
		log.Printf("[DEBUG] getTunnelDestinationIP: Failed to get route info: %v", routeErr)
	} else {
		log.Printf("[DEBUG] getTunnelDestinationIP: route output: %s", string(routeOutput))

		lines := strings.Split(string(routeOutput), "\n")
		for _, line := range lines {
			// 查找点对点路由，格式类似: "192.168.252.2 dev tun0 proto kernel scope link src 192.168.252.1"
			if strings.Contains(line, tunnelInterface) && !strings.Contains(line, "0.0.0.0") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					// 第一个字段通常是目标地址
					destAddr := parts[0]
					// 验证是否为有效IP地址
					if net.ParseIP(destAddr) != nil {
						log.Printf("[DEBUG] getTunnelDestinationIP: Found route destination IP: %s", destAddr)
						return destAddr
					}
				}
			}
		}
	}

	log.Printf("[DEBUG] getTunnelDestinationIP: No destination IP found for %s using any method", tunnelInterface)
	return ""
}

// runHpingTest 执行hping3测试

// GetTunnelInterfaceInfo 获取隧道接口详细信息
func (s *NetworkService) GetTunnelInterfaceInfo(interfaceName string) (*models.TunnelInterfaceInfo, error) {
	log.Printf("[DEBUG] Getting tunnel interface info for: %s", interfaceName)

	// 获取基础网络接口信息
	interfaces, err := s.GetAllInterfaces()
	if err != nil {
		return nil, err
	}

	var baseInterface models.NetworkInterface
	found := false
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			baseInterface = iface
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	tunnelInfo := &models.TunnelInterfaceInfo{
		NetworkInterface: baseInterface,
	}

	// 确定隧道类型
	if strings.HasPrefix(interfaceName, "tun") {
		tunnelInfo.TunnelType = "tun"
	} else if strings.HasPrefix(interfaceName, "tap") {
		tunnelInfo.TunnelType = "tap"
	} else {
		tunnelInfo.TunnelType = "unknown"
	}

	// 获取隧道配置信息
	tunnelInfo.LocalAddress, tunnelInfo.PeerAddress = s.getTunnelAddresses(interfaceName)

	// 获取相关规则
	tunnelInfo.RelatedRules, _ = s.GetTunnelInterfaceRules(interfaceName)

	// 获取连接的网桥
	tunnelInfo.ConnectedBridges = s.getConnectedBridges(interfaceName)

	return tunnelInfo, nil
}

// getTunnelAddresses 获取隧道地址信息
func (s *NetworkService) getTunnelAddresses(interfaceName string) (string, string) {
	cmd := exec.Command("ip", "addr", "show", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") && strings.Contains(line, "peer") {
			// 解析 "inet 192.168.252.1 peer 192.168.252.2/32" 格式
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "inet" && i+1 < len(parts) {
					localAddr := parts[i+1]
					if i+3 < len(parts) && parts[i+2] == "peer" {
						peerAddr := strings.Split(parts[i+3], "/")[0]
						return localAddr, peerAddr
					}
				}
			}
		}
	}

	return "", ""
}

// getConnectedBridges 获取连接的网桥列表
func (s *NetworkService) getConnectedBridges(interfaceName string) []string {
	var bridges []string

	// 通过iptables规则分析连接的网桥
	rules, err := s.GetTunnelInterfaceRules(interfaceName)
	if err != nil {
		return bridges
	}

	bridgeSet := make(map[string]bool)
	for _, rule := range rules {
		if rule.InInterface != interfaceName && strings.HasPrefix(rule.InInterface, "docker") {
			bridgeSet[rule.InInterface] = true
		}
		if rule.OutInterface != interfaceName && strings.HasPrefix(rule.OutInterface, "docker") {
			bridgeSet[rule.OutInterface] = true
		}
		if rule.InInterface != interfaceName && strings.HasPrefix(rule.InInterface, "br-") {
			bridgeSet[rule.InInterface] = true
		}
		if rule.OutInterface != interfaceName && strings.HasPrefix(rule.OutInterface, "br-") {
			bridgeSet[rule.OutInterface] = true
		}
	}

	for bridge := range bridgeSet {
		bridges = append(bridges, bridge)
	}

	return bridges
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

	// 设置主要IP地址（取第一个IP地址）
	if len(iface.IPAddresses) > 0 {
		bridge.IPAddress = iface.IPAddresses[0]
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

// FixConnectivity 修复隧道接口与Docker网桥之间的连通性问题
func (s *NetworkService) FixConnectivity(tunnelInterface, dockerBridge string) (*models.ConnectivityFixResult, error) {
	log.Printf("[DEBUG] NetworkService.FixConnectivity called with tunnel: %s, bridge: %s", tunnelInterface, dockerBridge)

	result := &models.ConnectivityFixResult{
		TunnelInterface: tunnelInterface,
		DockerBridge:    dockerBridge,
		FixedIssues:     []string{},
		AppliedRules:    []string{},
		Success:         false,
	}

	// 1. 检查接口是否存在
	tunnelExists, err := s.checkInterfaceExists(tunnelInterface)
	if err != nil {
		return result, fmt.Errorf("failed to check tunnel interface: %v", err)
	}
	if !tunnelExists {
		return result, fmt.Errorf("tunnel interface %s does not exist", tunnelInterface)
	}

	bridgeExists, err := s.checkInterfaceExists(dockerBridge)
	if err != nil {
		return result, fmt.Errorf("failed to check docker bridge: %v", err)
	}
	if !bridgeExists {
		return result, fmt.Errorf("docker bridge %s does not exist", dockerBridge)
	}

	// 2. 应用修复规则（按手动脚本的顺序执行）
	fixedCount := 0

	log.Printf("[DEBUG] Starting connectivity fix with script-compatible order")

	// 3.1 首先添加FORWARD规则（对应脚本中的步骤1和2）
	if err := s.ensureForwardRulesOptimized(tunnelInterface, dockerBridge, result); err != nil {
		log.Printf("[WARN] Failed to ensure forward rules: %v", err)
	} else {
		fixedCount++
		log.Printf("[DEBUG] ✓ FORWARD rules applied")
	}

	// 3.2 然后处理Docker隔离规则（对应脚本中的步骤3）
	if err := s.fixDockerIsolationRulesOptimized(tunnelInterface, dockerBridge, result); err != nil {
		log.Printf("[WARN] Failed to fix Docker isolation rules: %v", err)
	} else {
		fixedCount++
		log.Printf("[DEBUG] ✓ Docker isolation rules applied")
	}

	// 3.3 确保接口状态正常
	if err := s.ensureInterfaceState(tunnelInterface, dockerBridge, result); err != nil {
		log.Printf("[WARN] Failed to ensure interface state: %v", err)
	} else {
		fixedCount++
	}

	// 3.4 清理可能阻塞的规则（最后执行，避免干扰前面的规则）
	if err := s.cleanupBlockingRulesOptimized(tunnelInterface, dockerBridge, result); err != nil {
		log.Printf("[WARN] Failed to cleanup blocking rules: %v", err)
	} else {
		fixedCount++
	}

	// 4. 验证修复结果
	if fixedCount > 0 {
		// 等待规则生效
		time.Sleep(2 * time.Second)
		result.Success = true
		result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("成功应用 %d 项修复", fixedCount))

		// 记录详细的修复信息
		log.Printf("[DEBUG] === Connectivity Fix Summary ===")
		log.Printf("[DEBUG] Tunnel Interface: %s", tunnelInterface)
		log.Printf("[DEBUG] Docker Bridge: %s", dockerBridge)
		log.Printf("[DEBUG] Applied Rules Count: %d", len(result.AppliedRules))
		for i, rule := range result.AppliedRules {
			log.Printf("[DEBUG] Rule %d: %s", i+1, rule)
		}
		log.Printf("[DEBUG] Fixed Issues Count: %d", len(result.FixedIssues))
		for i, issue := range result.FixedIssues {
			log.Printf("[DEBUG] Issue %d: %s", i+1, issue)
		}
		log.Printf("[DEBUG] === End Summary ===")
	} else {
		log.Printf("[DEBUG] No rules were applied during connectivity fix")
	}

	log.Printf("[DEBUG] Connectivity fix completed. Success: %v, Fixed issues: %d", result.Success, len(result.FixedIssues))
	return result, nil
}

// fixDockerIsolationRules 修复Docker隔离规则问题
func (s *NetworkService) fixDockerIsolationRules(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	log.Printf("[DEBUG] Checking Docker isolation rules for %s -> %s", tunnelInterface, dockerBridge)

	// 获取DOCKER-ISOLATION-STAGE-2链的规则
	isolationRules, err := s.getDockerIsolationRules(tunnelInterface, dockerBridge)
	if err != nil {
		log.Printf("[WARN] Failed to get Docker isolation rules: %v", err)
		return nil // 不返回错误，因为链可能不存在
	}

	if len(isolationRules) == 0 {
		log.Printf("[DEBUG] No Docker isolation rules found")
		return nil
	}

	// 检查是否有阻断规则
	hasBlockingRules := false
	for _, rule := range isolationRules {
		if rule.Target == "DROP" {
			hasBlockingRules = true
			log.Printf("[DEBUG] Found blocking isolation rule: %s -> %s (line %d)",
				rule.InInterface, rule.OutInterface, rule.LineNumber)
		}
	}

	if !hasBlockingRules {
		log.Printf("[DEBUG] No blocking isolation rules found")
		return nil
	}

	// 尝试在DOCKER-ISOLATION-STAGE-2链的开头添加允许规则
	// 这样可以在DROP规则之前允许特定的隧道接口通信
	allowCmd := exec.Command("iptables", "-I", "DOCKER-ISOLATION-STAGE-2", "1",
		"-i", tunnelInterface, "-o", dockerBridge, "-j", "ACCEPT")

	if err := allowCmd.Run(); err != nil {
		log.Printf("[WARN] Failed to add isolation bypass rule: %v", err)
		// 尝试另一种方法：添加返回规则
		returnCmd := exec.Command("iptables", "-I", "DOCKER-ISOLATION-STAGE-2", "1",
			"-i", dockerBridge, "-o", tunnelInterface, "-j", "ACCEPT")

		if err := returnCmd.Run(); err != nil {
			log.Printf("[WARN] Failed to add isolation return rule: %v", err)
			return err
		} else {
			result.FixedIssues = append(result.FixedIssues,
				fmt.Sprintf("添加Docker隔离返回规则: %s -> %s", dockerBridge, tunnelInterface))
			result.AppliedRules = append(result.AppliedRules,
				fmt.Sprintf("iptables -I DOCKER-ISOLATION-STAGE-2 1 -i %s -o %s -j ACCEPT", dockerBridge, tunnelInterface))
		}
	} else {
		result.FixedIssues = append(result.FixedIssues,
			fmt.Sprintf("添加Docker隔离绕过规则: %s -> %s", tunnelInterface, dockerBridge))
		result.AppliedRules = append(result.AppliedRules,
			fmt.Sprintf("iptables -I DOCKER-ISOLATION-STAGE-2 1 -i %s -o %s -j ACCEPT", tunnelInterface, dockerBridge))

		// 同时添加返回路径规则
		returnCmd := exec.Command("iptables", "-I", "DOCKER-ISOLATION-STAGE-2", "2",
			"-i", dockerBridge, "-o", tunnelInterface, "-j", "ACCEPT")

		if err := returnCmd.Run(); err == nil {
			result.FixedIssues = append(result.FixedIssues,
				fmt.Sprintf("添加Docker隔离返回规则: %s -> %s", dockerBridge, tunnelInterface))
			result.AppliedRules = append(result.AppliedRules,
				fmt.Sprintf("iptables -I DOCKER-ISOLATION-STAGE-2 2 -i %s -o %s -j ACCEPT", dockerBridge, tunnelInterface))
		}
	}

	log.Printf("[DEBUG] Docker isolation rules fix completed")
	return nil
}

// checkIsolationRuleExists 检查隔离规则是否已存在
func (s *NetworkService) checkIsolationRuleExists(inInterface, outInterface, target string) bool {
	// 使用 iptables -C 命令精确检查Docker隔离规则是否存在
	cmd := exec.Command("iptables", "-C", "DOCKER-ISOLATION-STAGE-2", "-i", inInterface, "-o", outInterface, "-j", target)
	err := cmd.Run()

	if err == nil {
		log.Printf("[DEBUG] Found existing isolation rule: %s -> %s (%s)", inInterface, outInterface, target)
		return true
	}

	log.Printf("[DEBUG] Isolation rule %s -> %s (%s) not found", inInterface, outInterface, target)
	return false
}

// checkForwardRuleExists 检查FORWARD规则是否已存在
func (s *NetworkService) checkForwardRuleExists(inInterface, outInterface string) bool {
	// 使用 iptables -C 命令精确检查规则是否存在
	// 这是检查iptables规则的标准方法
	cmd := exec.Command("iptables", "-C", "FORWARD", "-i", inInterface, "-o", outInterface, "-j", "ACCEPT")
	err := cmd.Run()

	if err == nil {
		log.Printf("[DEBUG] Found existing FORWARD rule: %s -> %s", inInterface, outInterface)
		return true
	}

	log.Printf("[DEBUG] FORWARD rule %s -> %s not found", inInterface, outInterface)
	return false
}

// checkConntrackRuleExists 检查是否存在conntrack状态跟踪规则
func (s *NetworkService) checkConntrackRuleExists(inInterface, outInterface string) bool {
	// 使用 iptables -C 命令精确检查conntrack规则是否存在
	cmd := exec.Command("iptables", "-C", "FORWARD", "-i", inInterface, "-o", outInterface,
		"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	err := cmd.Run()

	if err == nil {
		log.Printf("[DEBUG] Found existing conntrack rule: %s -> %s", inInterface, outInterface)
		return true
	}

	log.Printf("[DEBUG] Conntrack rule %s -> %s not found", inInterface, outInterface)
	return false
}

// checkNATRuleExists 检查NAT规则是否已存在
func (s *NetworkService) checkNATRuleExists(rulePattern string) bool {
	cmd := exec.Command("iptables", "-t", "nat", "-L", "POSTROUTING", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), rulePattern)
}

// fixDockerIsolationRulesOptimized 优化的Docker隔离规则修复
func (s *NetworkService) fixDockerIsolationRulesOptimized(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	log.Printf("[DEBUG] Fixing Docker isolation rules for %s <-> %s", tunnelInterface, dockerBridge)

	// 检查DOCKER-ISOLATION-STAGE-2链是否存在
	checkCmd := exec.Command("iptables", "-L", "DOCKER-ISOLATION-STAGE-2", "-n")
	if err := checkCmd.Run(); err != nil {
		log.Printf("[INFO] DOCKER-ISOLATION-STAGE-2 chain does not exist, skipping isolation rules")
		return nil
	}

	// 3. 绕过 Docker 隔离规则 (与手动脚本保持一致)
	// 添加隧道接口到网桥的RETURN规则
	log.Printf("[DEBUG] Checking if isolation RETURN rule exists: %s -> %s", tunnelInterface, dockerBridge)
	if !s.checkIsolationRuleExists(tunnelInterface, dockerBridge, "RETURN") {
		log.Printf("[DEBUG] Adding isolation RETURN rule: %s -> %s", tunnelInterface, dockerBridge)
		returnCmd1 := exec.Command("iptables", "-I", "DOCKER-ISOLATION-STAGE-2", "1",
			"-i", tunnelInterface, "-o", dockerBridge, "-j", "RETURN")

		if err := returnCmd1.Run(); err != nil {
			log.Printf("[ERROR] Failed to add isolation RETURN rule: %v", err)
			return fmt.Errorf("failed to add isolation RETURN rule: %v", err)
		} else {
			result.FixedIssues = append(result.FixedIssues,
				fmt.Sprintf("绕过Docker隔离: %s -> %s", tunnelInterface, dockerBridge))
			result.AppliedRules = append(result.AppliedRules,
				fmt.Sprintf("iptables -I DOCKER-ISOLATION-STAGE-2 1 -i %s -o %s -j RETURN", tunnelInterface, dockerBridge))
			log.Printf("[SUCCESS] ✓ Added isolation bypass rule: %s -> %s", tunnelInterface, dockerBridge)
		}
	} else {
		log.Printf("[INFO] Isolation RETURN rule %s -> %s already exists, skipping", tunnelInterface, dockerBridge)
	}

	log.Printf("[DEBUG] Docker isolation rules fix completed for %s <-> %s", tunnelInterface, dockerBridge)
	return nil
}

// ensureForwardRulesOptimized 优化的FORWARD规则确保方法
func (s *NetworkService) ensureForwardRulesOptimized(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	log.Printf("[DEBUG] Ensuring FORWARD rules for %s <-> %s", tunnelInterface, dockerBridge)

	// 1. 允许 tun0 → 容器网桥的转发 (插入到FORWARD链的第1位)
	log.Printf("[DEBUG] Checking if FORWARD rule exists: %s -> %s", tunnelInterface, dockerBridge)
	if !s.checkForwardRuleExists(tunnelInterface, dockerBridge) {
		log.Printf("[DEBUG] Adding FORWARD rule: %s -> %s", tunnelInterface, dockerBridge)
		cmd1 := exec.Command("iptables", "-I", "FORWARD", "1", "-i", tunnelInterface, "-o", dockerBridge, "-j", "ACCEPT")
		if err := cmd1.Run(); err != nil {
			log.Printf("[ERROR] Failed to add forward rule %s -> %s: %v", tunnelInterface, dockerBridge, err)
			return fmt.Errorf("failed to add forward rule: %v", err)
		} else {
			rule1 := fmt.Sprintf("iptables -I FORWARD 1 -i %s -o %s -j ACCEPT", tunnelInterface, dockerBridge)
			result.AppliedRules = append(result.AppliedRules, rule1)
			result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("添加FORWARD规则: %s -> %s", tunnelInterface, dockerBridge))
			log.Printf("[SUCCESS] ✓ Added forward rule: %s -> %s", tunnelInterface, dockerBridge)
		}
	} else {
		log.Printf("[INFO] FORWARD rule %s -> %s already exists, skipping", tunnelInterface, dockerBridge)
	}

	// 2. 允许回包 (插入到FORWARD链的第2位，使用conntrack模块)
	log.Printf("[DEBUG] Checking if conntrack rule exists: %s -> %s", dockerBridge, tunnelInterface)
	if !s.checkConntrackRuleExists(dockerBridge, tunnelInterface) {
		log.Printf("[DEBUG] Adding conntrack rule: %s -> %s", dockerBridge, tunnelInterface)
		cmd2 := exec.Command("iptables", "-I", "FORWARD", "2", "-i", dockerBridge, "-o", tunnelInterface,
			"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
		if err := cmd2.Run(); err != nil {
			log.Printf("[ERROR] Failed to add conntrack rule %s -> %s: %v", dockerBridge, tunnelInterface, err)
			return fmt.Errorf("failed to add conntrack rule: %v", err)
		} else {
			rule2 := fmt.Sprintf("iptables -I FORWARD 2 -i %s -o %s -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT", dockerBridge, tunnelInterface)
			result.AppliedRules = append(result.AppliedRules, rule2)
			result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("添加回包规则: %s -> %s", dockerBridge, tunnelInterface))
			log.Printf("[SUCCESS] ✓ Added conntrack rule: %s -> %s", dockerBridge, tunnelInterface)
		}
	} else {
		log.Printf("[INFO] Conntrack rule %s -> %s already exists, skipping", dockerBridge, tunnelInterface)
	}

	log.Printf("[DEBUG] FORWARD rules optimization completed for %s <-> %s", tunnelInterface, dockerBridge)
	return nil
}

// ensureNATRulesOptimized 优化的NAT规则确保方法
func (s *NetworkService) ensureNATRulesOptimized(tunnelInterface, dockerBridge string, tunnelIPs, bridgeIPs []string, result *models.ConnectivityFixResult) error {
	if len(bridgeIPs) == 0 {
		return fmt.Errorf("no bridge IPs available")
	}

	bridgeIP := bridgeIPs[0]
	bridgeNetwork := s.getNetworkFromIP(bridgeIP)

	// 检查并添加MASQUERADE规则
	masqueradePattern := fmt.Sprintf("MASQUERADE.*%s.*%s", bridgeNetwork, tunnelInterface)
	if !s.checkNATRuleExists(masqueradePattern) {
		rule1 := fmt.Sprintf("-A POSTROUTING -s %s -o %s -j MASQUERADE", bridgeNetwork, tunnelInterface)
		if err := s.applyIPTablesRule("nat", rule1); err == nil {
			result.AppliedRules = append(result.AppliedRules, rule1)
			result.FixedIssues = append(result.FixedIssues, "添加MASQUERADE规则")
		}
	} else {
		log.Printf("[DEBUG] MASQUERADE rule already exists")
	}

	// 检查并添加SNAT规则（如果需要）
	if len(tunnelIPs) > 0 {
		tunnelIP := tunnelIPs[0]
		snatPattern := fmt.Sprintf("SNAT.*%s.*%s.*%s", bridgeNetwork, tunnelInterface, tunnelIP)
		if !s.checkNATRuleExists(snatPattern) {
			rule2 := fmt.Sprintf("-A POSTROUTING -s %s -o %s -j SNAT --to-source %s", bridgeNetwork, tunnelInterface, tunnelIP)
			if err := s.applyIPTablesRule("nat", rule2); err == nil {
				result.AppliedRules = append(result.AppliedRules, rule2)
				result.FixedIssues = append(result.FixedIssues, "添加SNAT规则")
			}
		} else {
			log.Printf("[DEBUG] SNAT rule already exists")
		}
	}

	return nil
}

// cleanupBlockingRulesOptimized 优化的阻塞规则清理方法
func (s *NetworkService) cleanupBlockingRulesOptimized(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	// 查找并删除可能阻塞的DROP或REJECT规则
	cmd := exec.Command("iptables", "-L", "FORWARD", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	deletedCount := 0

	for _, line := range lines {
		if strings.Contains(line, tunnelInterface) && strings.Contains(line, dockerBridge) {
			if strings.Contains(line, "DROP") || strings.Contains(line, "REJECT") {
				// 提取行号并删除规则
				fields := strings.Fields(line)
				if len(fields) > 0 {
					lineNum := fields[0]
					deleteCmd := exec.Command("iptables", "-D", "FORWARD", lineNum)
					if err := deleteCmd.Run(); err == nil {
						deletedCount++
						result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("删除阻塞规则: %s", line))
						result.AppliedRules = append(result.AppliedRules, fmt.Sprintf("iptables -D FORWARD %s", lineNum))
					}
				}
			}
		}
	}

	if deletedCount > 0 {
		log.Printf("[DEBUG] Cleaned up %d blocking rules", deletedCount)
	}

	return nil
}

// ensureForwardRules 确保FORWARD链有正确的规则
func (s *NetworkService) ensureForwardRules(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	// 允许从隧道接口到Docker网桥的流量
	rule1 := fmt.Sprintf("-A FORWARD -i %s -o %s -j ACCEPT", tunnelInterface, dockerBridge)
	if err := s.applyIPTablesRule("filter", rule1); err == nil {
		result.AppliedRules = append(result.AppliedRules, rule1)
		result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("添加FORWARD规则: %s -> %s", tunnelInterface, dockerBridge))
	}

	// 允许从Docker网桥到隧道接口的流量
	rule2 := fmt.Sprintf("-A FORWARD -i %s -o %s -j ACCEPT", dockerBridge, tunnelInterface)
	if err := s.applyIPTablesRule("filter", rule2); err == nil {
		result.AppliedRules = append(result.AppliedRules, rule2)
		result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("添加FORWARD规则: %s -> %s", dockerBridge, tunnelInterface))
	}

	// 允许已建立连接的流量
	rule3 := fmt.Sprintf("-A FORWARD -i %s -o %s -m state --state RELATED,ESTABLISHED -j ACCEPT", dockerBridge, tunnelInterface)
	if err := s.applyIPTablesRule("filter", rule3); err == nil {
		result.AppliedRules = append(result.AppliedRules, rule3)
		result.FixedIssues = append(result.FixedIssues, "添加状态跟踪规则")
	}

	return nil
}

// ensureNATRules 确保NAT规则正确配置
func (s *NetworkService) ensureNATRules(tunnelInterface, dockerBridge string, tunnelIPs, bridgeIPs []string, result *models.ConnectivityFixResult) error {
	if len(bridgeIPs) == 0 {
		return fmt.Errorf("no bridge IPs available")
	}

	bridgeIP := bridgeIPs[0]
	bridgeNetwork := s.getNetworkFromIP(bridgeIP)

	// 添加MASQUERADE规则，允许从隧道接口访问Docker网络
	rule1 := fmt.Sprintf("-A POSTROUTING -s %s -o %s -j MASQUERADE", bridgeNetwork, tunnelInterface)
	if err := s.applyIPTablesRule("nat", rule1); err == nil {
		result.AppliedRules = append(result.AppliedRules, rule1)
		result.FixedIssues = append(result.FixedIssues, "添加MASQUERADE规则")
	}

	// 添加SNAT规则（如果需要）
	if len(tunnelIPs) > 0 {
		tunnelIP := tunnelIPs[0]
		rule2 := fmt.Sprintf("-A POSTROUTING -s %s -o %s -j SNAT --to-source %s", bridgeNetwork, tunnelInterface, tunnelIP)
		if err := s.applyIPTablesRule("nat", rule2); err == nil {
			result.AppliedRules = append(result.AppliedRules, rule2)
			result.FixedIssues = append(result.FixedIssues, "添加SNAT规则")
		}
	}

	return nil
}

// ensureInterfaceState 确保接口状态正常
func (s *NetworkService) ensureInterfaceState(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	// 确保接口启用
	cmd1 := exec.Command("ip", "link", "set", tunnelInterface, "up")
	if err := cmd1.Run(); err == nil {
		result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("启用接口: %s", tunnelInterface))
	}

	cmd2 := exec.Command("ip", "link", "set", dockerBridge, "up")
	if err := cmd2.Run(); err == nil {
		result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("启用接口: %s", dockerBridge))
	}

	// 启用IP转发
	cmd3 := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd3.Run(); err == nil {
		result.FixedIssues = append(result.FixedIssues, "启用IP转发")
	}

	return nil
}

// cleanupBlockingRules 清理可能阻塞的规则
func (s *NetworkService) cleanupBlockingRules(tunnelInterface, dockerBridge string, result *models.ConnectivityFixResult) error {
	// 查找并删除可能阻塞的DROP或REJECT规则
	cmd := exec.Command("iptables", "-L", "FORWARD", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, tunnelInterface) && strings.Contains(line, dockerBridge) {
			if strings.Contains(line, "DROP") || strings.Contains(line, "REJECT") {
				// 提取行号并删除规则
				fields := strings.Fields(line)
				if len(fields) > 0 {
					lineNum := fields[0]
					deleteCmd := exec.Command("iptables", "-D", "FORWARD", lineNum)
					if err := deleteCmd.Run(); err == nil {
						result.FixedIssues = append(result.FixedIssues, fmt.Sprintf("删除阻塞规则: %s", line))
					}
				}
			}
		}
	}

	return nil
}

// applyIPTablesRule 应用iptables规则
func (s *NetworkService) applyIPTablesRule(table, rule string) error {
	args := []string{"-t", table}
	args = append(args, strings.Fields(rule)...)

	cmd := exec.Command("iptables", args...)
	return cmd.Run()
}

// getNetworkFromIP 从IP地址获取网络地址
func (s *NetworkService) getNetworkFromIP(ip string) string {
	// 简单的网络地址计算，假设是/24网络
	parts := strings.Split(ip, ".")
	if len(parts) >= 3 {
		return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
	}
	return ip + "/32"
}
