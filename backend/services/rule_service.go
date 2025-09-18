package services

import (
	"fmt"
	"iptables-management-backend/config"
	"iptables-management-backend/models"
)

type RuleService struct{}

// NewRuleService 创建规则服务实例
func NewRuleService() *RuleService {
	return &RuleService{}
}

// GetAllRules 获取所有规则
func (s *RuleService) GetAllRules() ([]models.IPTablesRule, error) {
	var rules []models.IPTablesRule
	err := config.DB.Order("created_at DESC").Find(&rules).Error
	return rules, err
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
		Where("DATE(timestamp) = CURDATE()").
		Count(&recentOps)
	
	return &models.Statistics{
		TotalRules:       int(totalRules),
		RulesByChain:     rulesByChain,
		RecentOperations: int(recentOps),
		SystemStatus:     "正常",
	}, nil
}