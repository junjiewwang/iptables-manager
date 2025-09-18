package models

import (
	"time"
	"gorm.io/gorm"
)

// IPTablesRule 规则模型
type IPTablesRule struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ChainName     string    `json:"chain_name" gorm:"size:50;not null"`
	RuleNumber    *int      `json:"rule_number" gorm:"index"`
	Target        string    `json:"target" gorm:"size:20;not null"`
	Protocol      string    `json:"protocol" gorm:"size:10"`
	SourceIP      string    `json:"source_ip" gorm:"size:45"`
	DestinationIP string    `json:"destination_ip" gorm:"size:45"`
	SourcePort    string    `json:"source_port" gorm:"size:20"`
	DestPort      string    `json:"destination_port" gorm:"size:20"`
	InterfaceIn   string    `json:"interface_in" gorm:"size:20"`
	InterfaceOut  string    `json:"interface_out" gorm:"size:20"`
	RuleText      string    `json:"rule_text" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:50;not null;index"`
	Operation string    `json:"operation" gorm:"size:100;not null"`
	Details   string    `json:"details" gorm:"type:text"`
	Timestamp time.Time `json:"timestamp" gorm:"autoCreateTime"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`
}

// User 用户模型
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Password string `json:"password" gorm:"size:255;not null"`
	Role     string `json:"role" gorm:"size:20;default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Statistics 统计信息模型
type Statistics struct {
	TotalRules       int               `json:"total_rules"`
	RulesByChain     map[string]int    `json:"rules_by_chain"`
	RecentOperations int               `json:"recent_operations"`
	SystemStatus     string            `json:"system_status"`
}

// TableName 设置表名
func (IPTablesRule) TableName() string {
	return "iptables_rules"
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

func (User) TableName() string {
	return "users"
}