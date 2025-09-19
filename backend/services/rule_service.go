package services

import (
	"fmt"
	"iptables-management-backend/config"
	"iptables-management-backend/models"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RuleService struct{}

// NewRuleService 创建规则服务实例
func NewRuleService() *RuleService {
	return &RuleService{}
}

// GetAllRules 获取所有规则（从数据库）
func (s *RuleService) GetAllRules() ([]models.IPTablesRule, error) {
	log.Println("[DEBUG] RuleService.GetAllRules called")

	var rules []models.IPTablesRule

	// 检查数据库连接
	if config.DB == nil {
		log.Println("[ERROR] Database connection is nil")
		return rules, fmt.Errorf("database connection is nil")
	}

	log.Println("[DEBUG] Executing database query: SELECT * FROM iptables_rules ORDER BY created_at DESC")

	err := config.DB.Order("created_at DESC").Find(&rules).Error
	if err != nil {
		log.Printf("[ERROR] Database query failed: %v", err)
		return rules, err
	}

	log.Printf("[DEBUG] Database query successful, found %d rules", len(rules))

	// 检查表是否存在
	var tableCount int
	err = config.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='iptables_rules'").Scan(&tableCount).Error
	if err != nil {
		log.Printf("[ERROR] Failed to check table existence: %v", err)
	} else {
		tableExists := tableCount > 0
		log.Printf("[DEBUG] Table 'iptables_rules' exists: %v", tableExists)
	}

	// 获取表的行数
	var count int64
	err = config.DB.Model(&models.IPTablesRule{}).Count(&count).Error
	if err != nil {
		log.Printf("[ERROR] Failed to count rules: %v", err)
	} else {
		log.Printf("[DEBUG] Total rules in database: %d", count)
	}

	return rules, nil
}

// GetSystemRules 从系统实时获取iptables规则
func (s *RuleService) GetSystemRules() ([]models.IPTablesRule, error) {
	log.Println("[DEBUG] RuleService.GetSystemRules called - fetching real iptables rules")

	var allRules []models.IPTablesRule
	tables := []string{"raw", "mangle", "nat", "filter"}

	for _, tableName := range tables {
		log.Printf("[DEBUG] Fetching rules from table: %s", tableName)

		rules, err := s.getTableRules(tableName)
		if err != nil {
			log.Printf("[ERROR] Failed to get rules from table %s: %v", tableName, err)
			continue
		}

		allRules = append(allRules, rules...)
		log.Printf("[DEBUG] Retrieved %d rules from table %s", len(rules), tableName)
	}

	log.Printf("[DEBUG] Total system rules retrieved: %d", len(allRules))
	return allRules, nil
}

// getTableRules 获取指定表的规则
func (s *RuleService) getTableRules(tableName string) ([]models.IPTablesRule, error) {
	cmd := exec.Command("iptables", "-t", tableName, "-L", "-n", "--line-numbers", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute iptables command: %v", err)
	}

	return s.parseIPTablesOutput(string(output), tableName)
}

// parseIPTablesOutput 解析iptables命令输出
func (s *RuleService) parseIPTablesOutput(output, tableName string) ([]models.IPTablesRule, error) {
	var rules []models.IPTablesRule
	lines := strings.Split(output, "\n")

	var currentChain string
	var chainPolicy string

	// 正则表达式匹配链头部信息
	chainRegex := regexp.MustCompile(`^Chain\s+(\S+)\s+\(policy\s+(\S+).*\)`)
	// 正则表达式匹配规则行
	ruleRegex := regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否是链头部
		if chainMatch := chainRegex.FindStringSubmatch(line); chainMatch != nil {
			currentChain = chainMatch[1]
			chainPolicy = chainMatch[2]
			log.Printf("[DEBUG] Found chain: %s with policy: %s", currentChain, chainPolicy)
			continue
		}

		// 检查是否是规则行
		if ruleMatch := ruleRegex.FindStringSubmatch(line); ruleMatch != nil && currentChain != "" {
			lineNum, _ := strconv.Atoi(ruleMatch[1])
			packets, _ := strconv.ParseInt(ruleMatch[2], 10, 64)
			bytes, _ := strconv.ParseInt(ruleMatch[3], 10, 64)
			target := ruleMatch[4]
			protocol := ruleMatch[5]
			opt := ruleMatch[6]
			inInterface := ruleMatch[7]

			// 解析剩余部分（出接口、源地址、目标地址等）
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
				Policy:       chainPolicy,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// SyncSystemRules 同步系统规则到数据库
func (s *RuleService) SyncSystemRules() error {
	log.Println("[DEBUG] Starting system rules synchronization")

	// 获取系统规则
	systemRules, err := s.GetSystemRules()
	if err != nil {
		return fmt.Errorf("failed to get system rules: %v", err)
	}

	// 清空现有规则
	err = config.DB.Exec("DELETE FROM iptables_rules").Error
	if err != nil {
		return fmt.Errorf("failed to clear existing rules: %v", err)
	}

	// 批量插入新规则
	if len(systemRules) > 0 {
		err = config.DB.Create(&systemRules).Error
		if err != nil {
			return fmt.Errorf("failed to insert system rules: %v", err)
		}
	}

	log.Printf("[DEBUG] Successfully synchronized %d rules to database", len(systemRules))
	return nil
}

// GetRuleByID 根据ID获取规则
func (s *RuleService) GetRuleByID(id uint) (*models.IPTablesRule, error) {
	var rule models.IPTablesRule
	err := config.DB.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateRule 创建新规则
func (s *RuleService) CreateRule(rule *models.IPTablesRule) error {
	// 生成规则文本
	rule.RuleText = s.generateRuleText(rule)
	return config.DB.Create(rule).Error
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(id uint, rule *models.IPTablesRule) error {
	// 生成规则文本
	rule.RuleText = s.generateRuleText(rule)
	return config.DB.Model(&models.IPTablesRule{}).Where("id = ?", id).Updates(rule).Error
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(id uint) error {
	return config.DB.Delete(&models.IPTablesRule{}, id).Error
}

// GetRulesByChain 根据链名获取规则
func (s *RuleService) GetRulesByChain(chainName string) ([]models.IPTablesRule, error) {
	var rules []models.IPTablesRule
	err := config.DB.Where("chain_name = ?", chainName).Order("rule_number ASC").Find(&rules).Error
	return rules, err
}

// generateRuleText 生成iptables规则文本
func (s *RuleService) generateRuleText(rule *models.IPTablesRule) string {
	ruleText := fmt.Sprintf("iptables -A %s", rule.ChainName)

	if rule.Protocol != "" {
		ruleText += fmt.Sprintf(" -p %s", rule.Protocol)
	}

	if rule.SourceIP != "" {
		ruleText += fmt.Sprintf(" -s %s", rule.SourceIP)
	}

	if rule.DestinationIP != "" {
		ruleText += fmt.Sprintf(" -d %s", rule.DestinationIP)
	}

	if rule.InterfaceIn != "" {
		ruleText += fmt.Sprintf(" -i %s", rule.InterfaceIn)
	}

	if rule.InterfaceOut != "" {
		ruleText += fmt.Sprintf(" -o %s", rule.InterfaceOut)
	}

	if rule.SourcePort != "" {
		ruleText += fmt.Sprintf(" --sport %s", rule.SourcePort)
	}

	if rule.DestPort != "" {
		ruleText += fmt.Sprintf(" --dport %s", rule.DestPort)
	}

	ruleText += fmt.Sprintf(" -j %s", rule.Target)

	return ruleText
}

// GetStatistics 获取统计信息
func (s *RuleService) GetStatistics() (*models.Statistics, error) {
	var totalRules int64
	config.DB.Model(&models.IPTablesRule{}).Count(&totalRules)

	// 按链统计规则数量
	var chainStats []struct {
		ChainName string
		Count     int64
	}
	config.DB.Model(&models.IPTablesRule{}).
		Select("chain_name, count(*) as count").
		Group("chain_name").
		Scan(&chainStats)

	rulesByChain := make(map[string]int)
	for _, stat := range chainStats {
		rulesByChain[stat.ChainName] = int(stat.Count)
	}

	// 获取今日操作数量
	var recentOps int64
	config.DB.Model(&models.OperationLog{}).
		Where("DATE(timestamp) = date('now')").
		Count(&recentOps)

	return &models.Statistics{
		TotalRules:       int(totalRules),
		RulesByChain:     rulesByChain,
		RecentOperations: int(recentOps),
		SystemStatus:     "正常",
	}, nil
}
