import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
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

// API接口
export const apiService = {
  // 认证相关
  login: (data: { username: string; password: string }) =>
    api.post('/login', data),

  // 规则管理
  getRules: () => api.get<IPTablesRule[]>('/rules'),
  addRule: (rule: IPTablesRule) => api.post('/rules', rule),
  updateRule: (id: number, rule: IPTablesRule) => api.put(`/rules/${id}`, rule),
  deleteRule: (id: number) => api.delete(`/rules/${id}`),

  // 统计信息
  getStatistics: () => api.get<Statistics>('/statistics'),

  // 操作日志
  getLogs: () => api.get<OperationLog[]>('/logs'),

  // 测试规则
  testRule: (data: { source_ip: string; destination_ip: string; port: number; protocol: string }) =>
    api.post('/test-rule', data),

  // 链管理
  listChainRules: (data: {
    chain_name?: string
    verbose?: boolean
    numeric?: boolean
    line_numbers?: boolean
  }) => api.post('/chains/list', data),
  
  createChain: (data: { chain_name: string }) => api.post('/chains/create', data),
  deleteChain: (chainName: string) => api.delete(`/chains/${chainName}`),
  flushChain: (data: { chain_name: string }) => api.post('/chains/flush', data),
  setChainPolicy: (data: { chain_name: string, policy: string }) => api.post('/chains/policy', data),

  // 备份和恢复
  backup: () => api.post('/backup'),
  restore: (data: any) => api.post('/restore', data)
}

export default api