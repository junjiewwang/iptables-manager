// 定义IPTables规则类型
export interface IPTablesRule {
  id?: number
  chain_name: string
  table: string
  target: string
  protocol: string
  source?: string
  destination?: string
  source_ip?: string
  destination_ip?: string
  source_port?: string
  destination_port?: string
  in_interface?: string
  out_interface?: string
  options?: string
  rule_text?: string
  line_number?: string
}

// 定义链信息类型
export interface ChainInfo {
  name: string
  policy?: string
  tables?: string[]
  rules?: IPTablesRule[]
}

// 定义表信息类型
export interface TableInfo {
  name: string
  total_rules?: number
  chains?: {
    name: string
    policy?: string
    rules?: IPTablesRule[]
  }[]
}

// 定义链表数据类型
export interface ChainTableData {
  chains: ChainInfo[]
  tables: TableInfo[]
  interfaceRules?: Record<string, IPTablesRule[]>
}

// 定义网络接口类型
export interface NetworkInterface {
  name: string
  type: string
  state: string
  ip_addresses: string[]
  mac_address: string
  mtu: number
  is_up: boolean
  is_docker: boolean
  statistics: {
    rx_bytes: number
    tx_bytes: number
    rx_packets: number
    tx_packets: number
  }
}


// 定义筛选条件类型
export interface FilterConditions {
  selectedInterfaces: string[]
  selectedProtocols: string[]
  selectedTargets: string[]
  ipRangeFilter: string
  portRangeFilter: string
}

// 定义规则表单类型
export interface RuleForm {
  id?: number
  chain_name: string
  table: string
  target: string
  protocol: string
  source_ip: string
  destination_ip: string
  source_port: string
  destination_port: string
  in_interface: string
  out_interface: string
  options: string
}

// 定义清理结果类型
export interface CleanupResult {
  total_cleaned: number
  duplicate_rules: number
  invalid_bridges: number
  invalid_chains: number
  invalid_targets: number
}

// 定义视图模式类型
export type ViewMode = 'chain' | 'table' | 'interface'