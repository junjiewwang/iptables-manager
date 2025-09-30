import { ref, computed } from 'vue'
import type { FilterConditions, IPTablesRule } from '@/types/ChainTable/types'

/**
 * 表格筛选逻辑组合式函数
 * 处理所有与筛选相关的状态和操作
 */
export function useTableFilters() {
  // 筛选面板状态
  const activeFilterPanels = ref(['filters'])
  
  // 筛选条件
  const selectedInterfaces = ref<string[]>([])
  const selectedProtocols = ref<string[]>([])
  const selectedTargets = ref<string[]>([])
  const ipRangeFilter = ref('')
  const portRangeFilter = ref('')
  
  // 接口视图筛选数据
  const selectedInterfaceTypes = ref<string[]>([])
  const interfaceStatusFilter = ref('')
  
  // 规则管理筛选
  const tableFilter = ref('')
  const targetFilter = ref('')
  const ruleSearchText = ref('')
  
  // 计算属性：活跃筛选条件数量
  const activeFiltersCount = computed(() => {
    let count = 0
    if (selectedInterfaces.value.length > 0) count++
    if (selectedProtocols.value.length > 0) count++
    if (selectedTargets.value.length > 0) count++
    if (ipRangeFilter.value) count++
    if (portRangeFilter.value) count++
    return count
  })
  
  // 计算属性：是否有活跃筛选条件
  const hasActiveFilters = computed(() => {
    return activeFiltersCount.value > 0
  })
  
  // 获取当前筛选条件
  const getFilterConditions = (): FilterConditions => {
    return {
      selectedInterfaces: selectedInterfaces.value,
      selectedProtocols: selectedProtocols.value,
      selectedTargets: selectedTargets.value,
      ipRangeFilter: ipRangeFilter.value,
      portRangeFilter: portRangeFilter.value
    }
  }
  
  // 应用筛选条件到规则列表
  const applyFiltersToRules = (rules: IPTablesRule[]): IPTablesRule[] => {
    let filtered = [...rules]
    
    // 按接口筛选
    if (selectedInterfaces.value.length > 0) {
      filtered = filtered.filter((rule: any) => 
        selectedInterfaces.value.includes(rule.in_interface) ||
        selectedInterfaces.value.includes(rule.out_interface)
      )
    }
    
    // 按协议筛选
    if (selectedProtocols.value.length > 0) {
      filtered = filtered.filter((rule: any) => 
        selectedProtocols.value.includes(rule.protocol?.toLowerCase())
      )
    }
    
    // 按目标动作筛选
    if (selectedTargets.value.length > 0) {
      filtered = filtered.filter((rule: any) => 
        selectedTargets.value.includes(rule.target)
      )
    }
    
    // 按IP范围筛选
    if (ipRangeFilter.value) {
      const ipPattern = ipRangeFilter.value.toLowerCase()
      filtered = filtered.filter((rule: any) => 
        rule.source?.toLowerCase().includes(ipPattern) ||
        rule.destination?.toLowerCase().includes(ipPattern) ||
        rule.source_ip?.toLowerCase().includes(ipPattern) ||
        rule.destination_ip?.toLowerCase().includes(ipPattern)
      )
    }
    
    // 按端口范围筛选
    if (portRangeFilter.value) {
      const portPattern = portRangeFilter.value.toLowerCase()
      filtered = filtered.filter((rule: any) => 
        rule.options?.toLowerCase().includes(portPattern) ||
        rule.source_port?.toLowerCase().includes(portPattern) ||
        rule.destination_port?.toLowerCase().includes(portPattern)
      )
    }
    
    // 按搜索文本筛选
    if (ruleSearchText.value) {
      const searchText = ruleSearchText.value.toLowerCase()
      filtered = filtered.filter((rule: any) => 
        rule.chain_name?.toLowerCase().includes(searchText) ||
        rule.target?.toLowerCase().includes(searchText) ||
        rule.source?.toLowerCase().includes(searchText) ||
        rule.destination?.toLowerCase().includes(searchText) ||
        rule.protocol?.toLowerCase().includes(searchText) ||
        rule.in_interface?.toLowerCase().includes(searchText) ||
        rule.out_interface?.toLowerCase().includes(searchText) ||
        rule.options?.toLowerCase().includes(searchText)
      )
    }
    
    // 按表筛选
    if (tableFilter.value) {
      filtered = filtered.filter((rule: any) => rule.table === tableFilter.value)
    }
    
    // 按目标筛选
    if (targetFilter.value) {
      filtered = filtered.filter((rule: any) => rule.target === targetFilter.value)
    }
    
    return filtered
  }
  
  // 筛选接口数据
  const applyFiltersToInterfaces = (interfaces: any[]) => {
    let filtered = [...interfaces]
    
    // 按接口类型筛选
    if (selectedInterfaceTypes.value.length > 0) {
      filtered = filtered.filter((iface: any) => 
        selectedInterfaceTypes.value.includes(iface.type)
      )
    }
    
    // 按接口状态筛选
    if (interfaceStatusFilter.value) {
      switch (interfaceStatusFilter.value) {
        case 'up':
          filtered = filtered.filter((iface: any) => iface.is_up)
          break
        case 'down':
          filtered = filtered.filter((iface: any) => !iface.is_up)
          break
        case 'docker':
          filtered = filtered.filter((iface: any) => iface.is_docker)
          break
      }
    }
    
    // 应用全局筛选条件：如果设置了接口筛选，只显示被选中的接口
    if (selectedInterfaces.value.length > 0) {
      filtered = filtered.filter((iface: any) => 
        selectedInterfaces.value.includes(iface.name)
      )
    }
    
    return filtered
  }
  
  // 切换接口筛选
  const toggleInterface = (interfaceName: string) => {
    const index = selectedInterfaces.value.indexOf(interfaceName)
    if (index > -1) {
      selectedInterfaces.value.splice(index, 1)
    } else {
      selectedInterfaces.value.push(interfaceName)
    }
  }
  
  // 切换协议筛选
  const toggleProtocol = (protocol: string) => {
    const index = selectedProtocols.value.indexOf(protocol)
    if (index > -1) {
      selectedProtocols.value.splice(index, 1)
    } else {
      selectedProtocols.value.push(protocol)
    }
  }
  
  // 切换目标筛选
  const toggleTarget = (target: string) => {
    const index = selectedTargets.value.indexOf(target)
    if (index > -1) {
      selectedTargets.value.splice(index, 1)
    } else {
      selectedTargets.value.push(target)
    }
  }
  
  // 切换接口类型筛选
  const toggleInterfaceType = (type: string) => {
    const index = selectedInterfaceTypes.value.indexOf(type)
    if (index > -1) {
      selectedInterfaceTypes.value.splice(index, 1)
    } else {
      selectedInterfaceTypes.value.push(type)
    }
  }
  
  // 切换接口状态筛选
  const toggleInterfaceStatus = (status: string) => {
    if (interfaceStatusFilter.value === status) {
      interfaceStatusFilter.value = ''
    } else {
      interfaceStatusFilter.value = status
    }
  }
  
  // 清空所有筛选条件
  const clearAllFilters = () => {
    selectedInterfaces.value = []
    selectedProtocols.value = []
    selectedTargets.value = []
    ipRangeFilter.value = ''
    portRangeFilter.value = ''
    ruleSearchText.value = ''
    tableFilter.value = ''
    targetFilter.value = ''
  }
  
  // 清空接口筛选条件
  const clearInterfaceFilters = () => {
    selectedInterfaceTypes.value = []
    interfaceStatusFilter.value = ''
  }
  
  return {
    // 状态
    activeFilterPanels,
    selectedInterfaces,
    selectedProtocols,
    selectedTargets,
    ipRangeFilter,
    portRangeFilter,
    selectedInterfaceTypes,
    interfaceStatusFilter,
    tableFilter,
    targetFilter,
    ruleSearchText,
    
    // 计算属性
    activeFiltersCount,
    hasActiveFilters,
    
    // 方法
    getFilterConditions,
    applyFiltersToRules,
    applyFiltersToInterfaces,
    toggleInterface,
    toggleProtocol,
    toggleTarget,
    toggleInterfaceType,
    toggleInterfaceStatus,
    clearAllFilters,
    clearInterfaceFilters
  }
}