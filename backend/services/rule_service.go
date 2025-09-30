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

// CompareSystemAndDatabaseRules 比对系统规则和数据库规则
func (s *RuleService) CompareSystemAndDatabaseRules() (bool, error) {
	log.Println("[DEBUG] Comparing system rules with database rules")

	// 获取系统规则
	systemRules, err := s.GetSystemRules()
	if err != nil {
		return false, fmt.Errorf("failed to get system rules: %v", err)
	}

	// 获取数据库规则
	dbRules, err := s.GetAllRules()
	if err != nil {
		return false, fmt.Errorf("failed to get database rules: %v", err)
	}

	// 比较规则数量
	if len(systemRules) != len(dbRules) {
		log.Printf("[DEBUG] Rule count mismatch: system=%d, database=%d", len(systemRules), len(dbRules))
		return false, nil
	}

	// 创建系统规则的映射，用于快速查找
	systemRuleMap := make(map[string]bool)
	for _, rule := range systemRules {
		key := fmt.Sprintf("%s-%s-%s-%s-%s", rule.ChainName, rule.Target, rule.Protocol, rule.SourceIP, rule.DestinationIP)
		systemRuleMap[key] = true
	}

	// 检查数据库规则是否都存在于系统规则中
	for _, rule := range dbRules {
		key := fmt.Sprintf("%s-%s-%s-%s-%s", rule.ChainName, rule.Target, rule.Protocol, rule.SourceIP, rule.DestinationIP)
		if !systemRuleMap[key] {
			log.Printf("[DEBUG] Database rule not found in system: %s", key)
			return false, nil
		}
	}

	log.Println("[DEBUG] System rules and database rules are consistent")
	return true, nil
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

// CleanupResult 清理结果结构
type CleanupResult struct {
	DuplicateRules   int      `json:"duplicate_rules"`
	InvalidBridges   int      `json:"invalid_bridges"`
	InvalidChains    int      `json:"invalid_chains"`
	InvalidTargets   int      `json:"invalid_targets"`
	TotalCleaned     int      `json:"total_cleaned"`
	CleanedRuleIDs   []uint   `json:"cleaned_rule_ids"`
	CleanedRuleTexts []string `json:"cleaned_rule_texts"`
}

// CleanInvalidRules 清除无效规则
func (s *RuleService) CleanInvalidRules(dryRun bool, username string) (*CleanupResult, error) {
	log.Printf("[DEBUG] CleanInvalidRules called with dryRun=%v, username=%s", dryRun, username)

	result := &CleanupResult{
		CleanedRuleIDs:   make([]uint, 0),
		CleanedRuleTexts: make([]string, 0),
	}

	// 获取所有规则
	var allRules []models.IPTablesRule
	err := config.DB.Find(&allRules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all rules: %v", err)
	}

	var rulesToDelete []uint

	// 1. 检查重复规则
	duplicateIDs := s.findDuplicateRules(allRules)
	result.DuplicateRules = len(duplicateIDs)
	rulesToDelete = append(rulesToDelete, duplicateIDs...)

	// 2. 检查无效网桥规则
	invalidBridgeIDs := s.findInvalidBridgeRules(allRules)
	result.InvalidBridges = len(invalidBridgeIDs)
	rulesToDelete = append(rulesToDelete, invalidBridgeIDs...)

	// 3. 检查无效链规则
	invalidChainIDs := s.findInvalidChainRules(allRules)
	result.InvalidChains = len(invalidChainIDs)
	rulesToDelete = append(rulesToDelete, invalidChainIDs...)

	// 4. 检查无效目标规则
	invalidTargetIDs := s.findInvalidTargetRules(allRules)
	result.InvalidTargets = len(invalidTargetIDs)
	rulesToDelete = append(rulesToDelete, invalidTargetIDs...)

	// 去重
	uniqueIDs := s.removeDuplicateIDs(rulesToDelete)
	result.TotalCleaned = len(uniqueIDs)
	result.CleanedRuleIDs = uniqueIDs

	// 获取要删除的规则文本
	for _, rule := range allRules {
		for _, id := range uniqueIDs {
			if rule.ID == id {
				result.CleanedRuleTexts = append(result.CleanedRuleTexts, rule.RuleText)
				break
			}
		}
	}

	// 如果不是dry-run模式，执行实际删除
	if !dryRun && len(uniqueIDs) > 0 {
		err = s.deleteRulesByIDs(uniqueIDs, username)
		if err != nil {
			return nil, fmt.Errorf("failed to delete invalid rules: %v", err)
		}
	}

	log.Printf("[DEBUG] CleanInvalidRules completed: %+v", result)
	return result, nil
}

// findDuplicateRules 查找重复规则
func (s *RuleService) findDuplicateRules(rules []models.IPTablesRule) []uint {
	var duplicateIDs []uint
	ruleMap := make(map[string][]uint)

	for _, rule := range rules {
		// 生成规则的唯一标识（忽略注释和空格）
		key := s.normalizeRuleKey(rule)
		ruleMap[key] = append(ruleMap[key], rule.ID)
	}

	// 找出重复的规则ID（保留第一个，删除其余的）
	for _, ids := range ruleMap {
		if len(ids) > 1 {
			// 保留第一个，其余的标记为重复
			duplicateIDs = append(duplicateIDs, ids[1:]...)
		}
	}

	log.Printf("[DEBUG] Found %d duplicate rules", len(duplicateIDs))
	return duplicateIDs
}

// normalizeRuleKey 标准化规则键（用于重复检测）
func (s *RuleService) normalizeRuleKey(rule models.IPTablesRule) string {
	// 移除空格和注释，生成标准化的规则键
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		strings.TrimSpace(rule.ChainName),
		strings.TrimSpace(rule.Target),
		strings.TrimSpace(rule.Protocol),
		strings.TrimSpace(rule.Source),
		strings.TrimSpace(rule.Destination),
		strings.TrimSpace(rule.InInterface),
		strings.TrimSpace(rule.OutInterface),
		strings.TrimSpace(rule.SourcePort),
		strings.TrimSpace(rule.DestPort))

	// 移除多余的空格和特殊字符
	key = regexp.MustCompile(`\s+`).ReplaceAllString(key, " ")
	return strings.ToLower(strings.TrimSpace(key))
}

// findInvalidBridgeRules 查找无效网桥规则
func (s *RuleService) findInvalidBridgeRules(rules []models.IPTablesRule) []uint {
	var invalidIDs []uint

	// 获取系统中存在的网桥
	validBridges := s.getValidBridges()

	for _, rule := range rules {
		// 检查输入接口
		if rule.InInterface != "" && strings.HasPrefix(rule.InInterface, "br-") {
			if !s.isBridgeValid(rule.InInterface, validBridges) {
				invalidIDs = append(invalidIDs, rule.ID)
				continue
			}
		}

		// 检查输出接口
		if rule.OutInterface != "" && strings.HasPrefix(rule.OutInterface, "br-") {
			if !s.isBridgeValid(rule.OutInterface, validBridges) {
				invalidIDs = append(invalidIDs, rule.ID)
			}
		}
	}

	log.Printf("[DEBUG] Found %d invalid bridge rules", len(invalidIDs))
	return invalidIDs
}

// getValidBridges 获取系统中有效的网桥
func (s *RuleService) getValidBridges() []string {
	cmd := exec.Command("brctl", "show")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[WARNING] Failed to get bridge list: %v", err)
		return []string{}
	}

	var bridges []string
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}

		fields := strings.Fields(line)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "\t") {
			bridges = append(bridges, fields[0])
		}
	}

	log.Printf("[DEBUG] Found valid bridges: %v", bridges)
	return bridges
}

// isBridgeValid 检查网桥是否有效
func (s *RuleService) isBridgeValid(bridge string, validBridges []string) bool {
	for _, validBridge := range validBridges {
		if bridge == validBridge {
			return true
		}
	}
	return false
}

// findInvalidChainRules 查找引用不存在链的规则
func (s *RuleService) findInvalidChainRules(rules []models.IPTablesRule) []uint {
	var invalidIDs []uint

	// 获取系统中存在的链
	validChains := s.getValidChains()

	for _, rule := range rules {
		if !s.isChainValid(rule.ChainName, validChains) {
			invalidIDs = append(invalidIDs, rule.ID)
		}
	}

	log.Printf("[DEBUG] Found %d invalid chain rules", len(invalidIDs))
	return invalidIDs
}

// getValidChains 获取系统中有效的链
func (s *RuleService) getValidChains() map[string][]string {
	validChains := make(map[string][]string)
	tables := []string{"raw", "mangle", "nat", "filter"}

	for _, table := range tables {
		cmd := exec.Command("iptables", "-t", table, "-L", "-n")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("[WARNING] Failed to get chains for table %s: %v", table, err)
			continue
		}

		chains := s.parseChainNames(string(output))
		validChains[table] = chains
	}

	log.Printf("[DEBUG] Found valid chains: %v", validChains)
	return validChains
}

// parseChainNames 解析链名称
func (s *RuleService) parseChainNames(output string) []string {
	var chains []string
	chainRegex := regexp.MustCompile(`^Chain\s+(\S+)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if match := chainRegex.FindStringSubmatch(line); match != nil {
			chains = append(chains, match[1])
		}
	}

	return chains
}

// isChainValid 检查链是否有效
func (s *RuleService) isChainValid(chainName string, validChains map[string][]string) bool {
	for _, chains := range validChains {
		for _, validChain := range chains {
			if chainName == validChain {
				return true
			}
		}
	}
	return false
}

// findInvalidTargetRules 查找目标指向不存在模块的规则
func (s *RuleService) findInvalidTargetRules(rules []models.IPTablesRule) []uint {
	var invalidIDs []uint

	// 定义有效的目标
	validTargets := []string{
		"ACCEPT", "DROP", "REJECT", "LOG", "RETURN",
		"DNAT", "SNAT", "MASQUERADE", "REDIRECT",
		"MARK", "TOS", "TTL", "ULOG", "NFQUEUE",
		"CONNMARK", "DSCP", "ECN", "HL", "HMARK",
		"IDLETIMER", "LED", "NFLOG", "RATEEST",
		"SECMARK", "SET", "SYNPROXY", "TCPMSS",
		"TCPOPTSTRIP", "TEE", "TPROXY", "TRACE",
	}

	// 获取系统中的自定义链（也可以作为目标）
	validChains := s.getValidChains()

	for _, rule := range rules {
		if !s.isTargetValid(rule.Target, validTargets, validChains) {
			invalidIDs = append(invalidIDs, rule.ID)
		}
	}

	log.Printf("[DEBUG] Found %d invalid target rules", len(invalidIDs))
	return invalidIDs
}

// isTargetValid 检查目标是否有效
func (s *RuleService) isTargetValid(target string, validTargets []string, validChains map[string][]string) bool {
	// 检查是否是标准目标
	for _, validTarget := range validTargets {
		if target == validTarget {
			return true
		}
	}

	// 检查是否是自定义链
	for _, chains := range validChains {
		for _, chain := range chains {
			if target == chain {
				return true
			}
		}
	}

	return false
}

// removeDuplicateIDs 去除重复的ID
func (s *RuleService) removeDuplicateIDs(ids []uint) []uint {
	seen := make(map[uint]bool)
	var unique []uint

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}

// deleteRulesByIDs 根据ID批量删除规则
func (s *RuleService) deleteRulesByIDs(ids []uint, username string) error {
	// 开始事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取要删除的规则详情用于日志记录
	var rulesToDelete []models.IPTablesRule
	err := tx.Where("id IN ?", ids).Find(&rulesToDelete).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除规则
	err = tx.Where("id IN ?", ids).Delete(&models.IPTablesRule{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录操作日志
	for _, rule := range rulesToDelete {
		logEntry := &models.OperationLog{
			Username:  username,
			Operation: "清除无效规则",
			Details:   fmt.Sprintf("删除规则: %s (ID: %d)", rule.RuleText, rule.ID),
			IPAddress: "system",
		}

		err = tx.Create(logEntry).Error
		if err != nil {
			log.Printf("[WARNING] Failed to create log entry: %v", err)
		}
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Successfully deleted %d invalid rules", len(ids))
	return nil
}
