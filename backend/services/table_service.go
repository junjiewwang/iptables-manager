package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type TableService struct{}

// NewTableService 创建表服务实例
func NewTableService() *TableService {
	return &TableService{}
}

// TableInfo 表信息结构
type TableInfo struct {
	TableName string      `json:"table_name"`
	Chains    []ChainInfo `json:"chains"`
}

// ChainInfo 链信息结构
type ChainInfo struct {
	ChainName string     `json:"chain_name"`
	Policy    string     `json:"policy"`
	Packets   string     `json:"packets"`
	Bytes     string     `json:"bytes"`
	Rules     []RuleInfo `json:"rules"`
}

// RuleInfo 规则信息结构
type RuleInfo struct {
	LineNumber  string `json:"line_number"`
	Packets     string `json:"packets"`
	Bytes       string `json:"bytes"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	Options     string `json:"options"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	RuleText    string `json:"rule_text"`
}

// GetAllTables 获取所有表的信息
func (s *TableService) GetAllTables() ([]TableInfo, error) {
	log.Println("[DEBUG] TableService.GetAllTables called")

	tables := []string{"raw", "mangle", "nat", "filter"}
	var allTables []TableInfo

	for _, tableName := range tables {
		tableInfo, err := s.GetTableInfo(tableName)
		if err != nil {
			log.Printf("[ERROR] Failed to get table info for %s: %v", tableName, err)
			continue
		}
		allTables = append(allTables, *tableInfo)
	}

	return allTables, nil
}

// GetTableInfo 获取指定表的信息
func (s *TableService) GetTableInfo(tableName string) (*TableInfo, error) {
	log.Printf("[DEBUG] Getting table info for: %s", tableName)

	// 执行 iptables -t <table> -L -n --line-numbers 命令
	cmd := exec.Command("iptables", "-t", tableName, "-L", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[ERROR] Failed to execute iptables command for table %s: %v", tableName, err)
		return nil, fmt.Errorf("failed to get table info: %v", err)
	}

	tableInfo := &TableInfo{
		TableName: tableName,
		Chains:    []ChainInfo{},
	}

	// 解析输出
	chains := s.parseTableOutput(string(output))
	tableInfo.Chains = chains

	return tableInfo, nil
}

// GetChainVerbose 获取指定链的详细信息
func (s *TableService) GetChainVerbose(tableName, chainName string) (*ChainInfo, error) {
	log.Printf("[DEBUG] Getting verbose info for chain %s in table %s", chainName, tableName)

	var cmd *exec.Cmd
	if tableName == "filter" {
		// 对于filter表，可以省略-t参数
		cmd = exec.Command("iptables", "-L", chainName, "-v")
	} else {
		cmd = exec.Command("iptables", "-t", tableName, "-L", chainName, "-v")
	}

	output, err := cmd.Output()
	if err != nil {
		log.Printf("[ERROR] Failed to execute iptables verbose command: %v", err)
		return nil, fmt.Errorf("failed to get chain verbose info: %v", err)
	}

	// 解析详细输出
	chainInfo := s.parseChainVerboseOutput(chainName, string(output))
	return chainInfo, nil
}

// parseTableOutput 解析iptables表输出
func (s *TableService) parseTableOutput(output string) []ChainInfo {
	var chains []ChainInfo
	var currentChain *ChainInfo

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否是链头部信息
		if strings.HasPrefix(line, "Chain ") {
			if currentChain != nil {
				chains = append(chains, *currentChain)
			}

			// 解析链信息: Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				chainName := parts[1]
				policy := "ACCEPT"
				packets := "0"
				bytes := "0"

				// 提取策略信息
				if strings.Contains(line, "policy") {
					for i, part := range parts {
						if part == "policy" && i+1 < len(parts) {
							policy = parts[i+1]
							break
						}
					}
				}

				// 提取包和字节数
				if strings.Contains(line, "packets") {
					for i, part := range parts {
						if part == "packets," && i-1 >= 0 {
							packets = parts[i-1]
						}
						if part == "bytes)" && i-1 >= 0 {
							bytes = parts[i-1]
						}
					}
				}

				currentChain = &ChainInfo{
					ChainName: chainName,
					Policy:    policy,
					Packets:   packets,
					Bytes:     bytes,
					Rules:     []RuleInfo{},
				}
			}
			continue
		}

		// 跳过表头
		if strings.Contains(line, "target") && strings.Contains(line, "prot") {
			continue
		}

		// 解析规则行
		if currentChain != nil && len(strings.Fields(line)) >= 3 {
			rule := s.parseRuleLine(line)
			if rule != nil {
				currentChain.Rules = append(currentChain.Rules, *rule)
			}
		}
	}

	// 添加最后一个链
	if currentChain != nil {
		chains = append(chains, *currentChain)
	}

	return chains
}

// parseRuleLine 解析规则行
func (s *TableService) parseRuleLine(line string) *RuleInfo {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return nil
	}

	rule := &RuleInfo{
		RuleText: line,
	}

	// 如果第一个字段是数字，说明有行号
	if len(fields) >= 6 {
		rule.LineNumber = fields[0]
		rule.Target = fields[1]
		rule.Protocol = fields[2]
		rule.Options = fields[3]
		rule.Source = fields[4]
		rule.Destination = fields[5]

		// 剩余部分作为选项
		if len(fields) > 6 {
			rule.Options = strings.Join(fields[6:], " ")
		}
	} else {
		rule.Target = fields[0]
		rule.Protocol = fields[1]
		rule.Source = fields[2]
		if len(fields) > 3 {
			rule.Destination = fields[3]
		}
	}

	return rule
}

// parseChainVerboseOutput 解析链的详细输出
func (s *TableService) parseChainVerboseOutput(chainName, output string) *ChainInfo {
	chainInfo := &ChainInfo{
		ChainName: chainName,
		Rules:     []RuleInfo{},
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析链头部信息
		if strings.HasPrefix(line, "Chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				policy := "ACCEPT"
				packets := "0"
				bytes := "0"

				if strings.Contains(line, "policy") {
					for i, part := range parts {
						if part == "policy" && i+1 < len(parts) {
							policy = parts[i+1]
							break
						}
					}
				}

				if strings.Contains(line, "packets") {
					for i, part := range parts {
						if part == "packets," && i-1 >= 0 {
							packets = parts[i-1]
						}
						if part == "bytes)" && i-1 >= 0 {
							bytes = parts[i-1]
						}
					}
				}

				chainInfo.Policy = policy
				chainInfo.Packets = packets
				chainInfo.Bytes = bytes
			}
			continue
		}

		// 跳过表头
		if strings.Contains(line, "pkts") && strings.Contains(line, "bytes") {
			continue
		}

		// 解析详细规则行
		if len(strings.Fields(line)) >= 6 {
			rule := s.parseVerboseRuleLine(line)
			if rule != nil {
				chainInfo.Rules = append(chainInfo.Rules, *rule)
			}
		}
	}

	return chainInfo
}

// parseVerboseRuleLine 解析详细规则行
func (s *TableService) parseVerboseRuleLine(line string) *RuleInfo {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	rule := &RuleInfo{
		Packets:  fields[0],
		Bytes:    fields[1],
		Target:   fields[2],
		Protocol: fields[3],
		Options:  fields[4],
		Source:   fields[5],
		RuleText: line,
	}

	if len(fields) > 6 {
		rule.Destination = fields[6]
	}

	if len(fields) > 7 {
		rule.Options = strings.Join(fields[7:], " ")
	}

	return rule
}
