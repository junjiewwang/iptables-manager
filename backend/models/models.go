package models

import (
	"time"
)

// IPTablesRule 规则模型
type IPTablesRule struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Table         string    `json:"table" gorm:"size:20;not null;index"`
	ChainName     string    `json:"chain_name" gorm:"size:50;not null"`
	LineNumber    int       `json:"line_number" gorm:"index"`
	RuleNumber    *int      `json:"rule_number" gorm:"index"`
	Target        string    `json:"target" gorm:"size:20;not null"`
	Protocol      string    `json:"protocol" gorm:"size:10"`
	Source        string    `json:"source" gorm:"size:45"`
	Destination   string    `json:"destination" gorm:"size:45"`
	SourceIP      string    `json:"source_ip" gorm:"size:45"`
	DestinationIP string    `json:"destination_ip" gorm:"size:45"`
	SourcePort    string    `json:"source_port" gorm:"size:20"`
	DestPort      string    `json:"destination_port" gorm:"size:20"`
	InInterface   string    `json:"in_interface" gorm:"size:20"`
	OutInterface  string    `json:"out_interface" gorm:"size:20"`
	InterfaceIn   string    `json:"interface_in" gorm:"size:20"`
	InterfaceOut  string    `json:"interface_out" gorm:"size:20"`
	Options       string    `json:"options" gorm:"size:100"`
	Extra         string    `json:"extra" gorm:"type:text"`
	Packets       int64     `json:"packets" gorm:"default:0"`
	Bytes         int64     `json:"bytes" gorm:"default:0"`
	Policy        string    `json:"policy" gorm:"size:20"`
	RuleText      string    `json:"rule_text" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NetworkInterface 网络接口模型
type NetworkInterface struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	State       string         `json:"state"`
	IPAddresses []string       `json:"ip_addresses"`
	MACAddress  string         `json:"mac_address"`
	MTU         int            `json:"mtu"`
	IsUp        bool           `json:"is_up"`
	IsDocker    bool           `json:"is_docker"`
	DockerType  string         `json:"docker_type,omitempty"`
	Statistics  InterfaceStats `json:"statistics"`
}

// InterfaceStats 接口统计信息
type InterfaceStats struct {
	RxBytes   int64 `json:"rx_bytes"`
	TxBytes   int64 `json:"tx_bytes"`
	RxPackets int64 `json:"rx_packets"`
	TxPackets int64 `json:"tx_packets"`
	RxErrors  int64 `json:"rx_errors"`
	TxErrors  int64 `json:"tx_errors"`
}

// DockerBridge Docker网桥信息
type DockerBridge struct {
	Name       string            `json:"name"`
	NetworkID  string            `json:"network_id"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	IPAMConfig DockerIPAMConfig  `json:"ipam_config"`
	Containers []DockerContainer `json:"containers"`
	Rules      []IPTablesRule    `json:"rules"`
	Interface  NetworkInterface  `json:"interface"`
}

// DockerIPAMConfig Docker IPAM配置
type DockerIPAMConfig struct {
	Driver  string            `json:"driver"`
	Config  []DockerSubnet    `json:"config"`
	Options map[string]string `json:"options"`
}

// DockerSubnet Docker子网配置
type DockerSubnet struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

// DockerContainer Docker容器信息
type DockerContainer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
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
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"size:255;not null"`
	Role      string    `json:"role" gorm:"size:20;default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Statistics 统计信息模型
type Statistics struct {
	TotalRules       int            `json:"total_rules"`
	RulesByChain     map[string]int `json:"rules_by_chain"`
	RecentOperations int            `json:"recent_operations"`
	SystemStatus     string         `json:"system_status"`
}

// NetworkConnection 网络连接信息
type NetworkConnection struct {
	Protocol       string `json:"protocol"`
	LocalAddress   string `json:"local_address"`
	ForeignAddress string `json:"foreign_address"`
	State          string `json:"state"`
}

// RouteEntry 路由表条目
type RouteEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Source      string `json:"source"`
	Metric      int    `json:"metric"`
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
