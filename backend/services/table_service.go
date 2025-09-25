package services

import (
	"fmt"
	"iptables-management-backend/utils"
	"log"
	"os/exec"
	"strconv"
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
	LineNumber   int    `json:"line_number"`   // 规则行号
	Packets      uint64 `json:"packets"`       // 匹配包数
	Bytes        string `json:"bytes"`         // 匹配字节数(含单位如K/M/G)
	Target       string `json:"target"`        // 目标动作
	Protocol     string `json:"protocol"`      // 协议类型
	Options      string `json:"options"`       // 特殊选项
	InInterface  string `json:"in_interface"`  // 入站接口
	OutInterface string `json:"out_interface"` // 出站接口
	Source       string `json:"source"`        // 源地址
	Destination  string `json:"destination"`   // 目标地址
	RuleText     string `json:"rule_text"`     // 完整规则文本
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
	cmd := exec.Command("iptables", "-t", tableName, "-L", "-n", "--line-numbers", "-v")
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
					start := strings.Index(line, "policy")
					if start != -1 {
						rest := line[start+6:]
						rest = strings.TrimSpace(rest)
						policyEnd := strings.Index(rest, " ")
						if policyEnd != -1 {
							policy = rest[:policyEnd]
						}
					}
				}

				// 提取包和字节数
				if strings.Contains(line, "packets") {
					fields := strings.Fields(line)
					for i, field := range fields {
						if field == "packets," && i > 0 {
							packets = fields[i-1]
						}
						if field == "bytes)" && i > 0 {
							bytes = fields[i-1]
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

		// 跳过表头行 (num pkts bytes target prot opt in out source destination)
		if strings.Contains(line, "num") && strings.Contains(line, "pkts") && strings.Contains(line, "bytes") {
			continue
		}

		// 解析规则行
		if currentChain != nil {
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
	if len(fields) < 10 { // 至少需要10个字段: num, pkts, bytes, target, prot, opt, in, out, source, destination
		return nil
	}

	// 解析行号
	lineNumber, err := strconv.Atoi(fields[0])
	if err != nil {
		log.Printf("[WARN] Failed to parse line number: %v", err)
		return nil
	}

	// 解析包数（支持K/M/G单位）
	packets := utils.ParsePacketCount(fields[1])

	rule := &RuleInfo{
		LineNumber:   lineNumber,
		Packets:      packets,
		Bytes:        fields[2],
		Target:       fields[3],
		Protocol:     fields[4],
		Options:      fields[5],
		InInterface:  fields[6],
		OutInterface: fields[7],
		Source:       fields[8],
		Destination:  fields[9],
		RuleText:     line,
	}

	// 如果有额外的选项参数（如ctstate等），合并到Options中
	if len(fields) > 10 {
		extraOptions := strings.Join(fields[10:], " ")
		if rule.Options != "--" {
			rule.Options = rule.Options + " " + extraOptions
		} else {
			rule.Options = extraOptions
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

	// 解析包数（支持K/M/G单位）
	packets := utils.ParsePacketCount(fields[0])

	rule := &RuleInfo{
		LineNumber:   0, // 详细模式下没有行号
		Packets:      packets,
		Bytes:        fields[1],
		Target:       fields[2],
		Protocol:     fields[3],
		Options:      fields[4],
		InInterface:  "",
		OutInterface: "",
		Source:       fields[5],
		RuleText:     line,
	}

	if len(fields) > 6 {
		rule.Destination = fields[6]
	}

	if len(fields) > 7 {
		rule.Options = strings.Join(fields[7:], " ")
	}

	return rule
}
