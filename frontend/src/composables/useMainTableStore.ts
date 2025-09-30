/**
 * MainTable 状态管理 Composable
 * 使用类似 Pinia 的状态管理模式
 */

import { ref, computed, reactive } from 'vue'
import type { 
  ViewMode, 
  IPTablesRule, 
  ChainInfo, 
  TableInfo, 
  NetworkInterface
} from '../components/ChainTable/types'

/**
 * MainTable 状态接口
 */
export interface MainTableState {
  // 视图相关
  viewMode: ViewMode
  showChainDialog: boolean
  detailTitle: string
  selectedChain: string
  
  // 数据相关
  chains: ChainInfo[]
  tables: TableInfo[]
  interfaces: NetworkInterface[]
  detailRules: IPTablesRule[]
  
  // 筛选相关
  groupByChain: boolean
  ruleSearchText: string
  tableFilter: string
  targetFilter: string
  
  // 拓扑设置已移除
  
  // 统计数据
  activeInterfacesCount: number
  dockerInterfacesCount: number
  totalInterfaceRules: number
}

/**
 * MainTable 状态管理 Hook
 */
export function useMainTableStore() {
  // 响应式状态
  const state = reactive<MainTableState>({
    viewMode: 'chain',
    showChainDialog: false,
    detailTitle: '',
    selectedChain: '',
    chains: [],
    tables: [],
    interfaces: [],
    detailRules: [],
    groupByChain: true,
    ruleSearchText: '',
    tableFilter: '',
    targetFilter: '',
    // 拓扑设置已移除
    activeInterfacesCount: 0,
    dockerInterfacesCount: 0,
    totalInterfaceRules: 0
  })

  // 计算属性
  const filteredDetailRules = computed(() => {
    let rules = state.detailRules

    // 搜索文本筛选
    if (state.ruleSearchText) {
      const searchText = state.ruleSearchText.toLowerCase()
      rules = rules.filter(rule => 
        rule.target?.toLowerCase().includes(searchText) ||
        rule.protocol?.toLowerCase().includes(searchText) ||
        rule.source?.toLowerCase().includes(searchText) ||
        rule.destination?.toLowerCase().includes(searchText) ||
        rule.options?.toLowerCase().includes(searchText)
      )
    }

    // 表筛选
    if (state.tableFilter) {
      rules = rules.filter(rule => rule.table === state.tableFilter)
    }

    // 目标筛选
    if (state.targetFilter) {
      rules = rules.filter(rule => rule.target === state.targetFilter)
    }

    return rules
  })

  const groupedRules = computed(() => {
    const grouped: Record<string, IPTablesRule[]> = {}
    
    filteredDetailRules.value.forEach(rule => {
      const chainName = rule.chain_name
      if (!grouped[chainName]) {
        grouped[chainName] = []
      }
      grouped[chainName].push(rule)
    })

    return grouped
  })

  const filteredInterfaceData = computed(() => {
    return state.interfaces.map(iface => ({
      ...iface,
      inRules: getInterfaceRuleCount(iface.name, 'input'),
      outRules: getInterfaceRuleCount(iface.name, 'output'),
      forwardRules: getInterfaceRuleCount(iface.name, 'forward')
    }))
  })

  // Actions
  const actions = {
    /**
     * 设置视图模式
     */
    setViewMode(mode: ViewMode) {
      state.viewMode = mode
    },

    /**
     * 显示链详情对话框
     */
    showChainDetail(chainName: string, title: string, rules: IPTablesRule[]) {
      state.selectedChain = chainName
      state.detailTitle = title
      state.detailRules = rules
      state.showChainDialog = true
    },

    /**
     * 隐藏链详情对话框
     */
    hideChainDetail() {
      state.showChainDialog = false
      state.selectedChain = ''
      state.detailTitle = ''
      state.detailRules = []
      // 重置筛选条件
      state.ruleSearchText = ''
      state.tableFilter = ''
      state.targetFilter = ''
    },

    /**
     * 设置分组模式
     */
    setGroupByChain(groupBy: boolean) {
      state.groupByChain = groupBy
    },

    /**
     * 设置搜索文本
     */
    setRuleSearchText(text: string) {
      state.ruleSearchText = text
    },

    /**
     * 设置表筛选
     */
    setTableFilter(table: string) {
      state.tableFilter = table
    },

    /**
     * 设置目标筛选
     */
    setTargetFilter(target: string) {
      state.targetFilter = target
    },

    /**
     * 更新链数据
     */
    updateChains(chains: ChainInfo[]) {
      state.chains = chains
    },

    /**
     * 更新表数据
     */
    updateTables(tables: TableInfo[]) {
      state.tables = tables
    },

    /**
     * 更新接口数据
     */
    updateInterfaces(interfaces: NetworkInterface[]) {
      state.interfaces = interfaces
      // 更新统计数据
      state.activeInterfacesCount = interfaces.filter(iface => iface.is_up).length
      state.dockerInterfacesCount = interfaces.filter(iface => iface.is_docker).length
      state.totalInterfaceRules = calculateTotalInterfaceRules(interfaces)
    },

    // 拓扑设置更新方法已移除

    /**
     * 重置所有筛选条件
     */
    resetFilters() {
      state.ruleSearchText = ''
      state.tableFilter = ''
      state.targetFilter = ''
    },

    /**
     * 重置状态
     */
    reset() {
      state.viewMode = 'chain'
      state.showChainDialog = false
      state.detailTitle = ''
      state.selectedChain = ''
      state.chains = []
      state.tables = []
      state.interfaces = []
      state.detailRules = []
      state.groupByChain = true
      this.resetFilters()
    }
  }

  // 辅助函数
  function getInterfaceRuleCount(interfaceName: string, direction: 'input' | 'output' | 'forward'): number {
    // 这里应该根据实际的规则数据计算
    // 暂时返回模拟数据
    return Math.floor(Math.random() * 10)
  }

  function calculateTotalInterfaceRules(interfaces: NetworkInterface[]): number {
    // 计算所有接口相关的规则总数
    return interfaces.reduce((total, iface) => {
      return total + getInterfaceRuleCount(iface.name, 'input') + 
             getInterfaceRuleCount(iface.name, 'output') + 
             getInterfaceRuleCount(iface.name, 'forward')
    }, 0)
  }

  // Getters
  const getters = {
    /**
     * 获取当前视图模式
     */
    getCurrentViewMode: () => state.viewMode,

    /**
     * 获取是否显示链详情对话框
     */
    isChainDialogVisible: () => state.showChainDialog,

    /**
     * 获取筛选后的规则数量
     */
    getFilteredRulesCount: () => filteredDetailRules.value.length,

    /**
     * 获取分组数量
     */
    getGroupedRulesCount: () => Object.keys(groupedRules.value).length,

    /**
     * 检查是否有活跃的筛选条件
     */
    hasActiveFilters: () => {
      return !!(state.ruleSearchText || state.tableFilter || state.targetFilter)
    },

    /**
     * 获取接口统计摘要
     */
    getInterfaceStatsSummary: () => ({
      total: state.interfaces.length,
      active: state.activeInterfacesCount,
      docker: state.dockerInterfacesCount,
      totalRules: state.totalInterfaceRules
    })
  }

  return {
    // 状态
    state: readonly(state),
    
    // 计算属性
    filteredDetailRules,
    groupedRules,
    filteredInterfaceData,
    
    // 操作方法
    ...actions,
    
    // 获取方法
    ...getters
  }
}

/**
 * 创建只读状态的辅助函数
 */
function readonly<T extends object>(obj: T): Readonly<T> {
  return new Proxy(obj, {
    set() {
      console.warn('Cannot modify readonly state directly. Use actions instead.')
      return false
    },
    deleteProperty() {
      console.warn('Cannot delete properties from readonly state.')
      return false
    }
  })
}