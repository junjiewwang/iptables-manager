package services

import (
	"encoding/json"
	"fmt"
	"iptables-management-backend/models"
	"log"
	"strings"
	"sync"
	"time"
)

type TopologyService struct {
	tableService   *TableService
	networkService *NetworkService
	cache          *TopologyCache
}

// TopologyCache 缓存结构
type TopologyCache struct {
	data      *TopologyData
	timestamp time.Time
	ttl       time.Duration
	mutex     sync.RWMutex
}

// NewTopologyCache 创建缓存实例
func NewTopologyCache(ttl time.Duration) *TopologyCache {
	return &TopologyCache{
		ttl: ttl,
	}
}

// NewTopologyService 创建拓扑服务实例
func NewTopologyService() *TopologyService {
	return &TopologyService{
		tableService:   NewTableService(),
		networkService: NewNetworkService(),
		cache:          NewTopologyCache(30 * time.Second), // 30秒缓存
	}
}

// TopologyData 拓扑图数据结构
type TopologyData struct {
	Nodes []TopologyNode `json:"nodes"`
	Links []TopologyLink `json:"links"`
	Flow  []FlowPath     `json:"flow"`
}

// TopologyNode 拓扑图节点
type TopologyNode struct {
	ID            string            `json:"id"`
	Label         string            `json:"label"`
	Type          string            `json:"type"` // interface, table, chain, rule
	InterfaceName string            `json:"interface_name,omitempty"`
	InterfaceType string            `json:"interface_type,omitempty"`
	TableName     string            `json:"table_name,omitempty"`
	ChainName     string            `json:"chain_name,omitempty"`
	Policy        string            `json:"policy,omitempty"`
	RuleCount     int               `json:"rule_count,omitempty"`
	RuleNumber    int               `json:"rule_number,omitempty"`
	Packets       string            `json:"packets,omitempty"`
	Bytes         string            `json:"bytes,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Position      Position          `json:"position"`
	Layer         int               `json:"layer"` // 用于分层布局
}

// TopologyLink 拓扑图连接
type TopologyLink struct {
	ID         string            `json:"id"`
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	Type       string            `json:"type"` // interface_rule, rule_interface, input, output, forward
	Label      string            `json:"label,omitempty"`
	RuleText   string            `json:"rule_text,omitempty"`
	RuleNumber int               `json:"rule_number,omitempty"`
	ChainType  string            `json:"chain_type,omitempty"` // INPUT, OUTPUT, FORWARD
	Action     string            `json:"action,omitempty"`     // ACCEPT, DROP, REJECT
	Protocol   string            `json:"protocol,omitempty"`
	Port       string            `json:"port,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// FlowPath 数据流路径
type FlowPath struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Path        []string `json:"path"` // 节点ID序列
	Color       string   `json:"color"`
}

// Position 节点位置
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// GetTopologyData 获取完整的拓扑图数据（带缓存）
func (s *TopologyService) GetTopologyData() (*TopologyData, error) {
	log.Println("[DEBUG] TopologyService.GetTopologyData called")

	// 检查缓存
	if cached := s.cache.get(); cached != nil {
		log.Println("[DEBUG] Returning cached topology data")
		return cached, nil
	}

	// 获取网络接口数据
	interfaces, err := s.networkService.GetAllInterfaces()
	if err != nil {
		log.Printf("[ERROR] Failed to get interfaces for topology: %v", err)
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	// 获取所有表数据
	tables, err := s.tableService.GetAllTables()
	if err != nil {
		log.Printf("[ERROR] Failed to get tables for topology: %v", err)
		return nil, fmt.Errorf("failed to get tables: %v", err)
	}

	topology := &TopologyData{
		Nodes: []TopologyNode{},
		Links: []TopologyLink{},
		Flow:  []FlowPath{},
	}

	// 生成以网络接口为中心的拓扑
	s.generateInterfaceNodes(interfaces, topology)
	s.generateRuleNodes(tables, topology)
	s.generateInterfaceRuleLinks(interfaces, tables, topology)
	s.generateNetworkFlowPaths(topology)

	// 数据验证和清理
	s.validateAndCleanTopology(topology)

	// 缓存结果
	s.cache.set(topology)

	log.Printf("[DEBUG] Generated topology with %d nodes, %d links, %d flows",
		len(topology.Nodes), len(topology.Links), len(topology.Flow))

	return topology, nil
}

// GetTopologyDataWithOptions 获取带参数的拓扑数据（支持过滤和分页）
func (s *TopologyService) GetTopologyDataWithOptions(options TopologyOptions) (*TopologyData, error) {
	log.Printf("[DEBUG] GetTopologyDataWithOptions called with options: %+v", options)

	baseTopology, err := s.GetTopologyData()
	if err != nil {
		return nil, err
	}

	// 应用过滤条件
	filteredTopology := s.applyFilters(baseTopology, options)

	// 应用分页
	if options.Pagination != nil {
		filteredTopology = s.applyPagination(filteredTopology, *options.Pagination)
	}

	return filteredTopology, nil
}

// TopologyOptions 查询参数
type TopologyOptions struct {
	ProtocolFilter  string             `json:"protocol_filter,omitempty"`
	ChainFilter     string             `json:"chain_filter,omitempty"`
	InterfaceFilter string             `json:"interface_filter,omitempty"`
	RuleTypeFilter  string             `json:"rule_type_filter,omitempty"`
	Pagination      *PaginationOptions `json:"pagination,omitempty"`
	IncludeStats    bool               `json:"include_stats,omitempty"`
	IncludeMetadata bool               `json:"include_metadata,omitempty"`
}

// PaginationOptions 分页参数
type PaginationOptions struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// validateAndCleanTopology 验证和清理拓扑数据
func (s *TopologyService) validateAndCleanTopology(topology *TopologyData) {
	// 移除孤立节点
	validNodeIDs := make(map[string]bool)
	for _, link := range topology.Links {
		validNodeIDs[link.Source] = true
		validNodeIDs[link.Target] = true
	}

	var validNodes []TopologyNode
	for _, node := range topology.Nodes {
		if validNodeIDs[node.ID] || node.Type == "interface" {
			validNodes = append(validNodes, node)
		}
	}
	topology.Nodes = validNodes

	// 验证数据完整性
	for i := range topology.Nodes {
		if topology.Nodes[i].Properties == nil {
			topology.Nodes[i].Properties = make(map[string]string)
		}
	}

	for i := range topology.Links {
		if topology.Links[i].Properties == nil {
			topology.Links[i].Properties = make(map[string]string)
		}
	}
}

// applyFilters 应用过滤条件
func (s *TopologyService) applyFilters(topology *TopologyData, options TopologyOptions) *TopologyData {
	if options.ProtocolFilter == "" && options.ChainFilter == "" && options.InterfaceFilter == "" {
		return topology
	}

	filtered := &TopologyData{
		Nodes: []TopologyNode{},
		Links: []TopologyLink{},
		Flow:  topology.Flow, // 保持流不变
	}

	// 过滤节点
	filteredNodeIDs := make(map[string]bool)
	for _, node := range topology.Nodes {
		if s.nodeMatchesFilter(node, options) {
			filtered.Nodes = append(filtered.Nodes, node)
			filteredNodeIDs[node.ID] = true
		}
	}

	// 过滤连接
	for _, link := range topology.Links {
		if filteredNodeIDs[link.Source] && filteredNodeIDs[link.Target] {
			filtered.Links = append(filtered.Links, link)
		}
	}

	return filtered
}

// nodeMatchesFilter 检查节点是否匹配过滤条件
func (s *TopologyService) nodeMatchesFilter(node TopologyNode, options TopologyOptions) bool {
	if options.ProtocolFilter != "" {
		if node.Type == "rule" && !strings.Contains(strings.ToLower(node.Properties["protocol"]), strings.ToLower(options.ProtocolFilter)) {
			return false
		}
	}

	if options.ChainFilter != "" {
		if node.Type == "rule" && !strings.EqualFold(node.ChainName, options.ChainFilter) {
			return false
		}
	}

	if options.InterfaceFilter != "" {
		if node.Type == "interface" && !strings.Contains(strings.ToLower(node.InterfaceName), strings.ToLower(options.InterfaceFilter)) {
			return false
		}
	}

	return true
}

// applyPagination 应用分页
func (s *TopologyService) applyPagination(topology *TopologyData, pagination PaginationOptions) *TopologyData {
	if pagination.PageSize <= 0 {
		pagination.PageSize = 50
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize

	if start >= len(topology.Nodes) {
		return &TopologyData{
			Nodes: []TopologyNode{},
			Links: []TopologyLink{},
			Flow:  topology.Flow,
		}
	}

	if end > len(topology.Nodes) {
		end = len(topology.Nodes)
	}

	paginated := &TopologyData{
		Nodes: topology.Nodes[start:end],
		Links: []TopologyLink{},
		Flow:  topology.Flow,
	}

	// 获取分页节点相关的连接
	paginatedNodeIDs := make(map[string]bool)
	for _, node := range paginated.Nodes {
		paginatedNodeIDs[node.ID] = true
	}

	for _, link := range topology.Links {
		if paginatedNodeIDs[link.Source] && paginatedNodeIDs[link.Target] {
			paginated.Links = append(paginated.Links, link)
		}
	}

	return paginated
}

// GetTopologyStats 获取拓扑统计信息
func (s *TopologyService) GetTopologyStats() (*TopologyStats, error) {
	topology, err := s.GetTopologyData()
	if err != nil {
		return nil, err
	}

	stats := &TopologyStats{
		TotalNodes:     len(topology.Nodes),
		TotalLinks:     len(topology.Links),
		TotalFlows:     len(topology.Flow),
		NodeTypes:      make(map[string]int),
		ChainTypes:     make(map[string]int),
		InterfaceTypes: make(map[string]int),
	}

	for _, node := range topology.Nodes {
		stats.NodeTypes[node.Type]++
		if node.Type == "interface" {
			stats.InterfaceTypes[node.InterfaceType]++
		} else if node.Type == "rule" {
			stats.ChainTypes[node.ChainName]++
		}
	}

	return stats, nil
}

// TopologyStats 拓扑统计信息
type TopologyStats struct {
	TotalNodes     int            `json:"total_nodes"`
	TotalLinks     int            `json:"total_links"`
	TotalFlows     int            `json:"total_flows"`
	NodeTypes      map[string]int `json:"node_types"`
	ChainTypes     map[string]int `json:"chain_types"`
	InterfaceTypes map[string]int `json:"interface_types"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

// cache相关方法
func (c *TopologyCache) get() *TopologyData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.data == nil || time.Since(c.timestamp) > c.ttl {
		return nil
	}

	// 深拷贝数据以避免并发问题
	dataBytes, _ := json.Marshal(c.data)
	var data TopologyData
	json.Unmarshal(dataBytes, &data)
	return &data
}

func (c *TopologyCache) set(data *TopologyData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 深拷贝数据
	dataBytes, _ := json.Marshal(data)
	var copiedData TopologyData
	json.Unmarshal(dataBytes, &copiedData)

	c.data = &copiedData
	c.timestamp = time.Now()
}

func (c *TopologyCache) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = nil
	c.timestamp = time.Time{}
}

// generateInterfaceNodes 生成网络接口节点
func (s *TopologyService) generateInterfaceNodes(interfaces []models.NetworkInterface, topology *TopologyData) {
	// 分层布局：外部网络在顶部，内部网络在底部
	externalY := 50
	internalY := 400
	dockerY := 600

	externalX := 100
	internalX := 100
	dockerX := 100

	for _, iface := range interfaces {
		var layer int
		var posY int
		var posX *int

		// 根据接口类型确定层级和位置
		if s.isExternalInterface(iface) {
			layer = 1 // 外部网络层
			posY = externalY
			posX = &externalX
		} else if iface.IsDocker {
			layer = 3 // Docker网络层
			posY = dockerY
			posX = &dockerX
		} else {
			layer = 2 // 内部网络层
			posY = internalY
			posX = &internalX
		}

		interfaceNode := TopologyNode{
			ID:            fmt.Sprintf("interface_%s", iface.Name),
			Label:         iface.Name,
			Type:          "interface",
			InterfaceName: iface.Name,
			InterfaceType: iface.Type,
			Layer:         layer,
			Position: Position{
				X: *posX,
				Y: posY,
			},
			Properties: map[string]string{
				"type":        iface.Type,
				"state":       iface.State,
				"mac_address": iface.MACAddress,
				"mtu":         fmt.Sprintf("%d", iface.MTU),
				"is_up":       fmt.Sprintf("%t", iface.IsUp),
				"is_docker":   fmt.Sprintf("%t", iface.IsDocker),
				"rx_bytes":    fmt.Sprintf("%d", iface.Statistics.RxBytes),
				"tx_bytes":    fmt.Sprintf("%d", iface.Statistics.TxBytes),
				"rx_packets":  fmt.Sprintf("%d", iface.Statistics.RxPackets),
				"tx_packets":  fmt.Sprintf("%d", iface.Statistics.TxPackets),
			},
		}

		// 添加IP地址信息
		if len(iface.IPAddresses) > 0 {
			interfaceNode.Properties["ip_addresses"] = strings.Join(iface.IPAddresses, ", ")
		}

		topology.Nodes = append(topology.Nodes, interfaceNode)
		*posX += 200 // 水平间距
	}
}

// generateRuleNodes 生成重要的iptables规则节点
func (s *TopologyService) generateRuleNodes(tables []TableInfo, topology *TopologyData) {
	ruleY := 250 // 规则层位置
	ruleX := 100

	for _, table := range tables {
		for _, chain := range table.Chains {
			// 只为INPUT、OUTPUT、FORWARD链生成规则节点
			if !s.isMainChain(chain.ChainName) {
				continue
			}

			for i, rule := range chain.Rules {
				if s.isImportantRule(rule) {
					ruleNode := TopologyNode{
						ID:         fmt.Sprintf("rule_%s_%s_%d", table.TableName, chain.ChainName, i),
						Label:      s.getRuleLabel(rule),
						Type:       "rule",
						TableName:  table.TableName,
						ChainName:  chain.ChainName,
						RuleNumber: i + 1,
						Packets:    fmt.Sprintf("%d", rule.Packets),
						Bytes:      rule.Bytes,
						Layer:      2, // 规则在中间层
						Position: Position{
							X: ruleX,
							Y: ruleY,
						},
						Properties: map[string]string{
							"target":        rule.Target,
							"protocol":      rule.Protocol,
							"source":        rule.Source,
							"destination":   rule.Destination,
							"line_number":   fmt.Sprintf("%d", rule.LineNumber),
							"interface_in":  s.extractInterface(rule.RuleText, "-i"),
							"interface_out": s.extractInterface(rule.RuleText, "-o"),
							"source_port":   s.extractPort(rule.RuleText, "--sport"),
							"dest_port":     s.extractPort(rule.RuleText, "--dport"),
						},
					}
					topology.Nodes = append(topology.Nodes, ruleNode)
					ruleX += 150 // 水平间距
				}
			}
		}
	}
}

// generateInterfaceRuleLinks 生成网络接口和规则之间的连接
func (s *TopologyService) generateInterfaceRuleLinks(interfaces []models.NetworkInterface, tables []TableInfo, topology *TopologyData) {
	// 为每个规则创建与相关网络接口的连接
	for _, table := range tables {
		for _, chain := range table.Chains {
			if !s.isMainChain(chain.ChainName) {
				continue
			}

			for i, rule := range chain.Rules {
				if !s.isImportantRule(rule) {
					continue
				}

				ruleID := fmt.Sprintf("rule_%s_%s_%d", table.TableName, chain.ChainName, i)

				// 提取规则中的接口信息
				inInterface := s.extractInterface(rule.RuleText, "-i")
				outInterface := s.extractInterface(rule.RuleText, "-o")

				// 创建输入接口连接
				if inInterface != "" && s.interfaceExists(inInterface, interfaces) {
					link := TopologyLink{
						ID:         fmt.Sprintf("link_%s_to_%s", fmt.Sprintf("interface_%s", inInterface), ruleID),
						Source:     fmt.Sprintf("interface_%s", inInterface),
						Target:     ruleID,
						Type:       "input",
						ChainType:  chain.ChainName,
						Action:     rule.Target,
						Protocol:   rule.Protocol,
						RuleNumber: i + 1,
						RuleText:   rule.RuleText,
						Label:      fmt.Sprintf("%s → %s", inInterface, rule.Target),
						Properties: map[string]string{
							"direction":     "input",
							"rule_number":   fmt.Sprintf("%d", i+1),
							"table":         table.TableName,
							"chain":         chain.ChainName,
							"in_interface":  inInterface,
							"out_interface": outInterface,
						},
					}
					topology.Links = append(topology.Links, link)
				}

				// 创建输出接口连接
				if outInterface != "" && s.interfaceExists(outInterface, interfaces) {
					link := TopologyLink{
						ID:         fmt.Sprintf("link_%s_to_%s", ruleID, fmt.Sprintf("interface_%s", outInterface)),
						Source:     ruleID,
						Target:     fmt.Sprintf("interface_%s", outInterface),
						Type:       "output",
						ChainType:  chain.ChainName,
						Action:     rule.Target,
						Protocol:   rule.Protocol,
						RuleNumber: i + 1,
						RuleText:   rule.RuleText,
						Label:      fmt.Sprintf("%s → %s", rule.Target, outInterface),
						Properties: map[string]string{
							"direction":     "output",
							"rule_number":   fmt.Sprintf("%d", i+1),
							"table":         table.TableName,
							"chain":         chain.ChainName,
							"in_interface":  inInterface,
							"out_interface": outInterface,
						},
					}
					topology.Links = append(topology.Links, link)
				}

				// 对于FORWARD链，创建接口到接口的直接连接
				if chain.ChainName == "FORWARD" && inInterface != "" && outInterface != "" {
					link := TopologyLink{
						ID:         fmt.Sprintf("link_%s_direct_%s", fmt.Sprintf("interface_%s", inInterface), fmt.Sprintf("interface_%s", outInterface)),
						Source:     fmt.Sprintf("interface_%s", inInterface),
						Target:     fmt.Sprintf("interface_%s", outInterface),
						Type:       "forward",
						ChainType:  chain.ChainName,
						Action:     rule.Target,
						Protocol:   rule.Protocol,
						RuleNumber: i + 1,
						RuleText:   rule.RuleText,
						Label:      fmt.Sprintf("%s → %s (%s)", inInterface, outInterface, rule.Target),
						Properties: map[string]string{
							"direction":     "forward",
							"rule_number":   fmt.Sprintf("%d", i+1),
							"table":         table.TableName,
							"chain":         chain.ChainName,
							"in_interface":  inInterface,
							"out_interface": outInterface,
						},
					}
					topology.Links = append(topology.Links, link)
				}

				// 如果没有指定接口，根据链类型创建通用连接
				if inInterface == "" && outInterface == "" {
					s.createGenericInterfaceLinks(ruleID, rule, chain, table, interfaces, topology)
				}
			}
		}
	}
}

// createGenericInterfaceLinks 为没有指定接口的规则创建通用连接
func (s *TopologyService) createGenericInterfaceLinks(ruleID string, rule RuleInfo, chain ChainInfo, table TableInfo, interfaces []models.NetworkInterface, topology *TopologyData) {
	switch chain.ChainName {
	case "INPUT":
		// INPUT链：从外部接口到规则
		for _, iface := range interfaces {
			if s.isExternalInterface(iface) {
				link := TopologyLink{
					ID:        fmt.Sprintf("link_%s_to_%s", fmt.Sprintf("interface_%s", iface.Name), ruleID),
					Source:    fmt.Sprintf("interface_%s", iface.Name),
					Target:    ruleID,
					Type:      "input",
					ChainType: chain.ChainName,
					Action:    rule.Target,
					Protocol:  rule.Protocol,
					RuleText:  rule.RuleText,
					Label:     fmt.Sprintf("INPUT: %s", rule.Target),
				}
				topology.Links = append(topology.Links, link)
			}
		}
	case "OUTPUT":
		// OUTPUT链：从规则到外部接口
		for _, iface := range interfaces {
			if s.isExternalInterface(iface) {
				link := TopologyLink{
					ID:        fmt.Sprintf("link_%s_to_%s", ruleID, fmt.Sprintf("interface_%s", iface.Name)),
					Source:    ruleID,
					Target:    fmt.Sprintf("interface_%s", iface.Name),
					Type:      "output",
					ChainType: chain.ChainName,
					Action:    rule.Target,
					Protocol:  rule.Protocol,
					RuleText:  rule.RuleText,
					Label:     fmt.Sprintf("OUTPUT: %s", rule.Target),
				}
				topology.Links = append(topology.Links, link)
			}
		}
	case "FORWARD":
		// FORWARD链：在接口之间转发
		externalInterfaces := s.getExternalInterfaces(interfaces)
		internalInterfaces := s.getInternalInterfaces(interfaces)

		for _, extIface := range externalInterfaces {
			for _, intIface := range internalInterfaces {
				// 外部到内部
				link1 := TopologyLink{
					ID:        fmt.Sprintf("link_%s_via_%s_to_%s", fmt.Sprintf("interface_%s", extIface.Name), ruleID, fmt.Sprintf("interface_%s", intIface.Name)),
					Source:    fmt.Sprintf("interface_%s", extIface.Name),
					Target:    fmt.Sprintf("interface_%s", intIface.Name),
					Type:      "forward",
					ChainType: chain.ChainName,
					Action:    rule.Target,
					Protocol:  rule.Protocol,
					RuleText:  rule.RuleText,
					Label:     fmt.Sprintf("FORWARD: %s", rule.Target),
				}
				topology.Links = append(topology.Links, link1)
			}
		}
	}
}

// generateNetworkFlowPaths 生成基于网络接口的数据流路径
func (s *TopologyService) generateNetworkFlowPaths(topology *TopologyData) {
	// 收集不同类型的节点
	interfaceNodes := make(map[string]TopologyNode)
	ruleNodes := make(map[string]TopologyNode)

	for _, node := range topology.Nodes {
		switch node.Type {
		case "interface":
			interfaceNodes[node.ID] = node
		case "rule":
			ruleNodes[node.ID] = node
		}
	}

	flows := []FlowPath{}

	// 生成入站数据流路径
	externalInterfaces := s.getExternalInterfaceNodes(interfaceNodes)
	inputRules := s.getRuleNodesByChain(ruleNodes, "INPUT")
	if len(externalInterfaces) > 0 && len(inputRules) > 0 {
		path := []string{}
		for _, iface := range externalInterfaces {
			path = append(path, iface.ID)
		}
		for _, rule := range inputRules {
			path = append(path, rule.ID)
		}

		flows = append(flows, FlowPath{
			ID:          "flow_input",
			Name:        "入站数据流",
			Description: "外部网络接口通过INPUT规则进入系统的数据流",
			Path:        path,
			Color:       "#4CAF50",
		})
	}

	// 生成出站数据流路径
	outputRules := s.getRuleNodesByChain(ruleNodes, "OUTPUT")
	if len(outputRules) > 0 && len(externalInterfaces) > 0 {
		path := []string{}
		for _, rule := range outputRules {
			path = append(path, rule.ID)
		}
		for _, iface := range externalInterfaces {
			path = append(path, iface.ID)
		}

		flows = append(flows, FlowPath{
			ID:          "flow_output",
			Name:        "出站数据流",
			Description: "系统通过OUTPUT规则向外部网络接口发送的数据流",
			Path:        path,
			Color:       "#2196F3",
		})
	}

	// 生成转发数据流路径
	forwardRules := s.getRuleNodesByChain(ruleNodes, "FORWARD")
	internalInterfaces := s.getInternalInterfaceNodes(interfaceNodes)
	if len(forwardRules) > 0 && len(externalInterfaces) > 0 && len(internalInterfaces) > 0 {
		path := []string{}
		for _, iface := range externalInterfaces {
			path = append(path, iface.ID)
		}
		for _, rule := range forwardRules {
			path = append(path, rule.ID)
		}
		for _, iface := range internalInterfaces {
			path = append(path, iface.ID)
		}

		flows = append(flows, FlowPath{
			ID:          "flow_forward",
			Name:        "转发数据流",
			Description: "外部网络接口通过FORWARD规则转发到内部网络接口的数据流",
			Path:        path,
			Color:       "#FF9800",
		})
	}

	// 生成Docker网络流路径
	dockerInterfaces := s.getDockerInterfaceNodes(interfaceNodes)
	if len(dockerInterfaces) > 0 {
		dockerRules := s.getDockerRelatedRules(ruleNodes)
		if len(dockerRules) > 0 {
			path := []string{}
			for _, iface := range dockerInterfaces {
				path = append(path, iface.ID)
			}
			for _, rule := range dockerRules {
				path = append(path, rule.ID)
			}

			flows = append(flows, FlowPath{
				ID:          "flow_docker",
				Name:        "Docker网络流",
				Description: "Docker容器网络的数据流路径",
				Path:        path,
				Color:       "#9C27B0",
			})
		}
	}

	topology.Flow = flows
}

// isExternalInterface 判断是否为外部网络接口
func (s *TopologyService) isExternalInterface(iface models.NetworkInterface) bool {
	externalTypes := []string{"ethernet", "wifi", "wlan", "ppp", "tunnel"}
	for _, t := range externalTypes {
		if strings.Contains(strings.ToLower(iface.Type), t) {
			return true
		}
	}
	// 根据接口名称判断
	externalNames := []string{"eth", "wlan", "wifi", "ppp", "wan", "tun", "tap"}
	for _, name := range externalNames {
		if strings.HasPrefix(strings.ToLower(iface.Name), name) {
			return true
		}
	}
	return false
}

// isMainChain 判断是否为主要链
func (s *TopologyService) isMainChain(chainName string) bool {
	mainChains := []string{"INPUT", "OUTPUT", "FORWARD", "PREROUTING", "POSTROUTING"}
	for _, chain := range mainChains {
		if chainName == chain {
			return true
		}
	}
	return false
}

// isImportantRule 判断是否为重要规则
func (s *TopologyService) isImportantRule(rule RuleInfo) bool {
	// 显示有特殊目标的规则
	importantTargets := []string{"ACCEPT", "DROP", "REJECT", "DNAT", "SNAT", "MASQUERADE", "REDIRECT", "LOG"}
	for _, target := range importantTargets {
		if strings.Contains(strings.ToUpper(rule.Target), target) {
			return true
		}
	}

	// 显示有接口指定的规则（包括tun和docker接口）
	if strings.Contains(rule.RuleText, "-i ") || strings.Contains(rule.RuleText, "-o ") {
		// 特别检查tun和docker接口
		if strings.Contains(rule.RuleText, "tun") ||
			strings.Contains(rule.RuleText, "docker") ||
			strings.Contains(rule.RuleText, "br-") {
			return true
		}
		return true
	}

	// 显示有端口指定的规则
	if strings.Contains(rule.RuleText, "--dport") || strings.Contains(rule.RuleText, "--sport") {
		return true
	}

	// 显示有特定协议的规则
	if rule.Protocol != "all" && rule.Protocol != "" {
		return true
	}

	// 注意：RuleInfo结构体中没有ChainName字段，需要从上下文获取

	return false
}

// extractInterface 从规则文本中提取接口名称
func (s *TopologyService) extractInterface(ruleText, flag string) string {
	parts := strings.Fields(ruleText)
	for i, part := range parts {
		if part == flag && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractPort 从规则文本中提取端口号
func (s *TopologyService) extractPort(ruleText, flag string) string {
	parts := strings.Fields(ruleText)
	for i, part := range parts {
		if part == flag && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// interfaceExists 检查接口是否存在
func (s *TopologyService) interfaceExists(interfaceName string, interfaces []models.NetworkInterface) bool {
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			return true
		}
	}
	return false
}

// getExternalInterfaces 获取外部接口列表
func (s *TopologyService) getExternalInterfaces(interfaces []models.NetworkInterface) []models.NetworkInterface {
	var external []models.NetworkInterface
	for _, iface := range interfaces {
		if s.isExternalInterface(iface) {
			external = append(external, iface)
		}
	}
	return external
}

// getInternalInterfaces 获取内部接口列表
func (s *TopologyService) getInternalInterfaces(interfaces []models.NetworkInterface) []models.NetworkInterface {
	var internal []models.NetworkInterface
	for _, iface := range interfaces {
		if !s.isExternalInterface(iface) && !iface.IsDocker {
			internal = append(internal, iface)
		}
	}
	return internal
}

// getExternalInterfaceNodes 获取外部接口节点
func (s *TopologyService) getExternalInterfaceNodes(interfaceNodes map[string]TopologyNode) []TopologyNode {
	var nodes []TopologyNode
	for _, node := range interfaceNodes {
		if node.Layer == 1 { // 外部网络层
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getInternalInterfaceNodes 获取内部接口节点
func (s *TopologyService) getInternalInterfaceNodes(interfaceNodes map[string]TopologyNode) []TopologyNode {
	var nodes []TopologyNode
	for _, node := range interfaceNodes {
		if node.Layer == 2 { // 内部网络层
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getDockerInterfaceNodes 获取Docker接口节点
func (s *TopologyService) getDockerInterfaceNodes(interfaceNodes map[string]TopologyNode) []TopologyNode {
	var nodes []TopologyNode
	for _, node := range interfaceNodes {
		if node.Layer == 3 { // Docker网络层
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getRuleNodesByChain 根据链名称获取规则节点
func (s *TopologyService) getRuleNodesByChain(ruleNodes map[string]TopologyNode, chainName string) []TopologyNode {
	var nodes []TopologyNode
	for _, node := range ruleNodes {
		if node.ChainName == chainName {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getDockerRelatedRules 获取Docker相关的规则节点
func (s *TopologyService) getDockerRelatedRules(ruleNodes map[string]TopologyNode) []TopologyNode {
	var nodes []TopologyNode
	for _, node := range ruleNodes {
		interfaceIn := strings.ToLower(node.Properties["interface_in"])
		interfaceOut := strings.ToLower(node.Properties["interface_out"])
		source := strings.ToLower(node.Properties["source"])
		destination := strings.ToLower(node.Properties["destination"])

		// 检查Docker相关接口
		if strings.Contains(interfaceIn, "docker") ||
			strings.Contains(interfaceOut, "docker") ||
			strings.Contains(interfaceIn, "br-") ||
			strings.Contains(interfaceOut, "br-") {
			nodes = append(nodes, node)
		}

		// 检查Docker常用网段
		if strings.Contains(source, "172.17") ||
			strings.Contains(destination, "172.17") ||
			strings.Contains(source, "172.18") ||
			strings.Contains(destination, "172.18") {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getRuleLabel 获取规则标签
func (s *TopologyService) getRuleLabel(rule RuleInfo) string {
	if rule.Target != "" {
		return fmt.Sprintf("%s", rule.Target)
	}
	if rule.Protocol != "" && rule.Protocol != "all" {
		return fmt.Sprintf("%s", rule.Protocol)
	}
	return fmt.Sprintf("Rule %d", rule.LineNumber)
}
