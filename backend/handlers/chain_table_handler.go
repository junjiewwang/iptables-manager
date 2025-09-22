package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"iptables-management-backend/models"
	"iptables-management-backend/services"
)

type ChainTableHandler struct {
	tableService   *services.TableService
	networkService *services.NetworkService
	ruleService    *services.RuleService
	logService     *services.LogService
}

// NewChainTableHandler 创建五链四表处理器实例
func NewChainTableHandler(tableService *services.TableService, networkService *services.NetworkService, ruleService *services.RuleService, logService *services.LogService) *ChainTableHandler {
	return &ChainTableHandler{
		tableService:   tableService,
		networkService: networkService,
		ruleService:    ruleService,
		logService:     logService,
	}
}

// ChainData 链数据结构
type ChainData struct {
	Name   string                `json:"name"`
	Tables []ChainTableData      `json:"tables"`
	Rules  []models.IPTablesRule `json:"rules"`
}

// ChainTableData 链中的表数据
type ChainTableData struct {
	Name  string                `json:"name"`
	Rules []models.IPTablesRule `json:"rules"`
}

// TableData 表数据结构
type TableData struct {
	Name       string      `json:"name"`
	TotalRules int         `json:"total_rules"`
	Chains     []ChainInfo `json:"chains"`
}

// ChainInfo 链信息
type ChainInfo struct {
	Name   string                `json:"name"`
	Policy string                `json:"policy,omitempty"`
	Rules  []models.IPTablesRule `json:"rules"`
}

// ChainTableResponse 五链四表响应数据
type ChainTableResponse struct {
	Chains         []ChainData                      `json:"chains"`
	Tables         []TableData                      `json:"tables"`
	InterfaceRules map[string][]models.IPTablesRule `json:"interface_rules"`
}

// GetChainTableData 获取五链四表聚合数据
func (h *ChainTableHandler) GetChainTableData(c *gin.Context) {
	interfaceName := c.Query("interface")
	log.Printf("[DEBUG] GetChainTableData API called with interface: %s", interfaceName)

	// 获取所有规则
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		log.Printf("[ERROR] Failed to get rules: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取规则失败"})
		return
	}

	// 如果指定了接口，过滤规则
	if interfaceName != "" {
		rules = h.filterRulesByInterface(rules, interfaceName)
	}

	// 构建响应数据
	response := h.buildChainTableResponse(rules)

	log.Printf("[DEBUG] Retrieved chain-table data with %d chains, %d tables", len(response.Chains), len(response.Tables))
	c.JSON(http.StatusOK, response)
}

// filterRulesByInterface 根据接口过滤规则
func (h *ChainTableHandler) filterRulesByInterface(rules []models.IPTablesRule, interfaceName string) []models.IPTablesRule {
	var filteredRules []models.IPTablesRule

	for _, rule := range rules {
		if rule.InInterface == interfaceName ||
			rule.OutInterface == interfaceName ||
			rule.InterfaceIn == interfaceName ||
			rule.InterfaceOut == interfaceName {
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules
}

// buildChainTableResponse 构建五链四表响应数据
func (h *ChainTableHandler) buildChainTableResponse(rules []models.IPTablesRule) ChainTableResponse {
	// 定义五链顺序
	chainOrder := []string{"PREROUTING", "INPUT", "FORWARD", "OUTPUT", "POSTROUTING"}
	// 定义四表
	tableNames := []string{"raw", "mangle", "nat", "filter"}

	// 按链分组规则
	chainRulesMap := make(map[string][]models.IPTablesRule)
	// 按表分组规则
	tableRulesMap := make(map[string][]models.IPTablesRule)
	// 按接口分组规则
	interfaceRulesMap := make(map[string][]models.IPTablesRule)

	for _, rule := range rules {
		// 按链分组
		chainRulesMap[rule.ChainName] = append(chainRulesMap[rule.ChainName], rule)

		// 按表分组
		tableRulesMap[rule.Table] = append(tableRulesMap[rule.Table], rule)

		// 按接口分组
		if rule.InInterface != "" {
			interfaceRulesMap[rule.InInterface] = append(interfaceRulesMap[rule.InInterface], rule)
		}
		if rule.OutInterface != "" {
			interfaceRulesMap[rule.OutInterface] = append(interfaceRulesMap[rule.OutInterface], rule)
		}
		if rule.InterfaceIn != "" {
			interfaceRulesMap[rule.InterfaceIn] = append(interfaceRulesMap[rule.InterfaceIn], rule)
		}
		if rule.InterfaceOut != "" {
			interfaceRulesMap[rule.InterfaceOut] = append(interfaceRulesMap[rule.InterfaceOut], rule)
		}
	}

	// 构建链数据
	var chains []ChainData
	for _, chainName := range chainOrder {
		chainRules := chainRulesMap[chainName]

		// 构建该链中的表数据
		var tables []ChainTableData
		for _, tableName := range tableNames {
			var tableRulesInChain []models.IPTablesRule
			for _, rule := range chainRules {
				if rule.Table == tableName {
					tableRulesInChain = append(tableRulesInChain, rule)
				}
			}

			tables = append(tables, ChainTableData{
				Name:  tableName,
				Rules: tableRulesInChain,
			})
		}

		chains = append(chains, ChainData{
			Name:   chainName,
			Tables: tables,
			Rules:  chainRules,
		})
	}

	// 构建表数据
	var tables []TableData
	for _, tableName := range tableNames {
		tableRules := tableRulesMap[tableName]

		// 按链分组该表的规则
		chainInfoMap := make(map[string][]models.IPTablesRule)
		for _, rule := range tableRules {
			chainInfoMap[rule.ChainName] = append(chainInfoMap[rule.ChainName], rule)
		}

		var chainInfos []ChainInfo
		for _, chainName := range chainOrder {
			if chainRules, exists := chainInfoMap[chainName]; exists {
				// 获取链的策略（如果有的话）
				policy := h.getChainPolicy(tableName, chainName, chainRules)

				chainInfos = append(chainInfos, ChainInfo{
					Name:   chainName,
					Policy: policy,
					Rules:  chainRules,
				})
			}
		}

		tables = append(tables, TableData{
			Name:       tableName,
			TotalRules: len(tableRules),
			Chains:     chainInfos,
		})
	}

	return ChainTableResponse{
		Chains:         chains,
		Tables:         tables,
		InterfaceRules: interfaceRulesMap,
	}
}

// getChainPolicy 获取链的策略
func (h *ChainTableHandler) getChainPolicy(tableName, chainName string, rules []models.IPTablesRule) string {
	// 尝试从规则中提取策略信息
	for _, rule := range rules {
		if rule.Policy != "" {
			return rule.Policy
		}
	}

	// 如果没有找到策略，尝试从iptables命令获取
	// 这里可以调用tableService来获取更详细的信息
	if chainInfo, err := h.tableService.GetChainVerbose(tableName, chainName); err == nil {
		// ChainInfo结构体中的Policy字段包含链的策略
		if chainInfo.Policy != "" {
			return chainInfo.Policy
		}
	}

	return ""
}

// GetInterfaceRuleStats 获取指定接口的规则统计
func (h *ChainTableHandler) GetInterfaceRuleStats(c *gin.Context) {
	interfaceName := c.Param("name")
	log.Printf("[DEBUG] GetInterfaceRuleStats API called for interface: %s", interfaceName)

	// 获取所有规则
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		log.Printf("[ERROR] Failed to get rules: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取规则失败"})
		return
	}

	// 过滤该接口相关的规则
	interfaceRules := h.filterRulesByInterface(rules, interfaceName)

	// 统计不同类型的规则
	stats := map[string]interface{}{
		"interface_name": interfaceName,
		"total_rules":    len(interfaceRules),
		"in_rules":       0,
		"out_rules":      0,
		"forward_rules":  0,
		"by_table":       make(map[string]int),
		"by_chain":       make(map[string]int),
		"rules":          interfaceRules,
	}

	tableStats := make(map[string]int)
	chainStats := make(map[string]int)

	for _, rule := range interfaceRules {
		// 统计表
		tableStats[rule.Table]++
		// 统计链
		chainStats[rule.ChainName]++

		// 统计方向
		if rule.InInterface == interfaceName || rule.InterfaceIn == interfaceName {
			stats["in_rules"] = stats["in_rules"].(int) + 1
		}
		if rule.OutInterface == interfaceName || rule.InterfaceOut == interfaceName {
			stats["out_rules"] = stats["out_rules"].(int) + 1
		}
		if rule.ChainName == "FORWARD" {
			stats["forward_rules"] = stats["forward_rules"].(int) + 1
		}
	}

	stats["by_table"] = tableStats
	stats["by_chain"] = chainStats

	log.Printf("[DEBUG] Retrieved rule stats for interface %s: %d total rules", interfaceName, len(interfaceRules))
	c.JSON(http.StatusOK, stats)
}
