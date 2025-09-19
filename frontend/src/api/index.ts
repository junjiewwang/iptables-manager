import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    console.log('[DEBUG] API Request:', config.method?.toUpperCase(), config.url)
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    console.error('[ERROR] API Request failed:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    console.log('[DEBUG] API Response:', response.config.method?.toUpperCase(), response.config.url, 
      'Status:', response.status, 'Data length:', Array.isArray(response.data) ? response.data.length : 'N/A')
    if (Array.isArray(response.data)) {
      console.log('[DEBUG] Response data sample:', response.data.slice(0, 2))
    }
    return response
  },
  (error) => {
    console.error('[ERROR] API Response failed:', error.config?.method?.toUpperCase(), error.config?.url, 
      'Status:', error.response?.status, 'Message:', error.message)
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export interface IPTablesRule {
  id?: number
  chain_name: string
  rule_number?: number
  target: string
  protocol?: string
  source_ip?: string
  destination_ip?: string
  source_port?: string
  destination_port?: string
  interface_in?: string
  interface_out?: string
  rule_text?: string
  created_at?: string
  updated_at?: string
}

export interface OperationLog {
  id: number
  username: string
  operation: string
  details: string
  timestamp: string
  ip_address: string
}

export interface Statistics {
  total_rules: number
  rules_by_chain: Record<string, number>
  recent_operations: number
  system_status: string
}

export interface RuleInfo {
  line_number?: string
  packets?: string
  bytes?: string
  target: string
  protocol: string
  options?: string
  source: string
  destination: string
  rule_text: string
}

export interface ChainInfo {
  chain_name: string
  policy: string
  packets: string
  bytes: string
  rules: RuleInfo[]
}

export interface TableInfo {
  table_name: string
  chains: ChainInfo[]
}

export interface NetworkInterface {
  name: string
  type: string
  state: string
  mac_address: string
  mtu: number
  is_up: boolean
  is_docker: boolean
  ip_addresses: string[]
  statistics: {
    rx_bytes: number
    tx_bytes: number
    rx_packets: number
    tx_packets: number
  }
}

export interface ChainVerbose {
  chain_name: string
  table_name: string
  policy: string
  packets: string
  bytes: string
  rules: RuleInfo[]
  references: number
}

export interface DockerBridge {
  name: string
  id: string
  driver: string
  network: string
  subnet: string
  gateway: string
  containers: number
  created: string
}

export interface BridgeRule {
  id: string
  chain: string
  rule: string
  target: string
  protocol: string
  source: string
  destination: string
}

export interface SpecialChain {
  name: string
  table: string
  type: 'builtin' | 'user' | 'custom'
  policy: string
  packets: string
  bytes: string
}

export interface NetworkConnection {
  protocol: string
  local_address: string
  local_port: number
  remote_address: string
  remote_port: number
  state: string
  pid: number
  process: string
}

export interface Route {
  destination: string
  gateway: string
  genmask: string
  flags: string
  interface: string
  metric: number
}

// 拓扑图相关接口定义
export interface TopologyNode {
  id: string
  label: string
  type: 'interface' | 'table' | 'chain' | 'rule'
  interface_name?: string
  interface_type?: string
  table_name?: string
  chain_name?: string
  policy?: string
  rule_count?: number
  rule_number?: number
  packets?: string
  bytes?: string
  properties?: Record<string, string>
  position: {
    x: number
    y: number
  }
  layer: number
}

export interface TopologyLink {
  id: string
  source: string
  target: string
  type: 'interface_rule' | 'rule_interface' | 'input' | 'output' | 'forward'
  label?: string
  rule_text?: string
  rule_number?: number
  chain_type?: string
  action?: string
  protocol?: string
  port?: string
  properties?: Record<string, string>
}

export interface FlowPath {
  id: string
  name: string
  description: string
  path: string[]
  color: string
}

export interface TopologyData {
  nodes: TopologyNode[]
  links: TopologyLink[]
  flow: FlowPath[]
}

export interface TopologyStats {
  total_nodes: number
  total_links: number
  total_flows: number
  node_types: Record<string, number>
  chain_types: Record<string, number>
  interface_types: Record<string, number>
  generated_at: number
}

export interface TopologyOptions {
  protocol_filter?: string
  chain_filter?: string
  interface_filter?: string
  rule_type_filter?: string
  page?: number
  page_size?: number
  include_stats?: boolean
  include_metadata?: boolean
}

export interface APIResponse<T> {
  success: boolean
  data: T
  meta?: {
    timestamp: number
    version?: string
    refreshed?: boolean
  }
  stats?: TopologyStats
  error?: {
    message: string
    code: string
    details?: string
  }
}

export interface TopologyHealth {
  status: 'healthy' | 'unhealthy'
  response_time: number
  timestamp: number
  error?: string
}

// 网络接口API
export const networkAPI = {
  // 获取网络接口
  getInterfaces: () => api.get<NetworkInterface[]>('/network/interfaces'),
  
  // 获取Docker网桥
  getDockerBridges: () => api.get<DockerBridge[]>('/docker/bridges'),
  
  // 获取网桥规则（新增）
  getBridgeRules: (bridgeName: string) => api.get<BridgeRule[]>(`/bridges/${bridgeName}/rules`),
  
  // 获取网络连接
  getNetworkConnections: () => api.get<NetworkConnection[]>('/network/connections'),
  
  // 获取路由表
  getRouteTable: () => api.get<Route[]>('/network/routes')
}

// 表管理API
export const tablesAPI = {
  // 获取所有表
  getAllTables: () => api.get<TableInfo[]>('/tables'),
  
  // 获取特定表信息（新增）
  getTableInfo: (table: string) => api.get<TableInfo>(`/tables/${table}`),
  
  // 获取特定链的详细信息
  getChainVerbose: (table: string, chain: string) => 
    api.get<ChainVerbose>(`/tables/${table}/chains/${chain}`),
  
  // 获取特殊链（新增）
  getSpecialChains: () => api.get<SpecialChain[]>('/special-chains')
}

// 规则管理API
export const rulesAPI = {
  // 获取所有规则
  getRules: () => api.get<IPTablesRule[]>('/rules'),
  
  // 获取规则统计
  getRuleStats: () => api.get<Statistics>('/statistics'),
  
  // 获取系统规则
  getSystemRules: () => api.get<IPTablesRule[]>('/rules/system'),
  
  // 同步系统规则（新增）
  syncSystemRules: () => api.post('/rules/sync'),
  
  // 添加规则
  addRule: (rule: Partial<IPTablesRule>) => api.post('/rules', rule),
  
  // 删除规则
  deleteRule: (id: number) => api.delete(`/rules/${id}`),
  
  // 更新规则
  updateRule: (id: number, rule: Partial<IPTablesRule>) => api.put(`/rules/${id}`, rule)
}

// 拓扑图API（新增）
export const topologyAPI = {
  // 获取拓扑数据（支持查询参数）
  getTopology: (options?: TopologyOptions) => {
    const params = new URLSearchParams()
    
    if (options?.protocol_filter) params.append('protocol', options.protocol_filter)
    if (options?.chain_filter) params.append('chain', options.chain_filter)
    if (options?.interface_filter) params.append('interface', options.interface_filter)
    if (options?.rule_type_filter) params.append('rule_type', options.rule_type_filter)
    if (options?.page) params.append('page', options.page.toString())
    if (options?.page_size) params.append('page_size', options.page_size.toString())
    if (options?.include_stats) params.append('include_stats', 'true')
    if (options?.include_metadata) params.append('include_metadata', 'true')
    
    return api.get<APIResponse<TopologyData>>(`/topology${params.toString() ? '?' + params.toString() : ''}`)
  },
  
  // 获取拓扑统计信息
  getTopologyStats: () => api.get<APIResponse<TopologyStats>>('/topology/stats'),
  
  // 强制刷新拓扑缓存
  refreshTopology: () => api.post<APIResponse<TopologyData>>('/topology/refresh'),
  
  // 导出拓扑数据
  exportTopology: (format: 'json' | 'csv' = 'json') => 
    api.get(`/topology/export?format=${format}`, {
      responseType: format === 'json' ? 'json' : 'text'
    }),
  
  // 获取拓扑服务健康状态
  getTopologyHealth: () => api.get<APIResponse<TopologyHealth>>('/topology/health')
}

// 日志API
export const logsAPI = {
  // 获取操作日志
  getLogs: (page = 1, pageSize = 20) => 
    api.get<OperationLog[]>(`/logs?page=${page}&page_size=${pageSize}`),
  
  // 获取日志统计
  getLogStats: () => api.get<Record<string, any>>('/logs/stats')
}

// 认证API
export const authAPI = {
  // 用户登录
  login: (username: string, password: string) => 
    api.post('/auth/login', { username, password }),
  
  // 用户登出
  logout: () => api.post('/auth/logout'),
  
  // 验证token
  validateToken: () => api.get('/auth/validate')
}

// 错误处理工具函数
export const handleAPIError = (error: any): string => {
  if (error.response?.data?.error) {
    const { message, code, details } = error.response.data.error
    return `${message}${details ? ` (${details})` : ''}`
  }
  return error.message || '未知错误'
}

// 网络错误重试机制
export const retryAPIRequest = async <T>(
  apiCall: () => Promise<T>,
  maxRetries = 3,
  delay = 1000
): Promise<T> => {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await apiCall()
    } catch (error) {
      if (i === maxRetries - 1) throw error
      console.warn(`[WARN] API request failed, retrying in ${delay}ms...`)
      await new Promise(resolve => setTimeout(resolve, delay))
      delay *= 2 // 指数退避
    }
  }
  throw new Error('Max retries exceeded')
}

// 统一的API服务
export const apiService = {
  // 拓扑图相关API
  ...topologyAPI,
  
  // 规则管理相关API
  ...rulesAPI,
  
  // 表管理相关API
  ...tablesAPI,
  
  // 网络接口相关API
  ...networkAPI,
  
  // 日志相关API
  ...logsAPI,
  
  // 认证相关API
  ...authAPI,
  
  // 备份功能（在Dashboard.vue中使用）
  backup: () => api.post('/system/backup'),
  
  // 获取系统统计信息（在Dashboard.vue中使用）
  getStatistics: () => api.get<Statistics>('/statistics'),
  
  // 获取系统状态（在Dashboard.vue中使用）
  getSystemStatus: () => api.get<{ status: string }>('/system/status')
}

export default api