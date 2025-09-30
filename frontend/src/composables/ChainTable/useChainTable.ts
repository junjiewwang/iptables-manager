import { computed, ref } from 'vue'
import type { ChainTableData, IPTablesRule, NetworkInterface, ViewMode } from '@/types/ChainTable'
import { useFormatters } from '@/composables/core/useFormatters'
import { useTagTypes } from '@/composables/core/useTagTypes'
import { useChainTableData } from './useChainTableData'
import { useChainTableDetail } from './useChainTableDetail'
import { useChainTableStats } from './useChainTableStats'

/**
 * 链表视图核心逻辑组合式函数
 * 集成各个子组合式函数，提供统一的接口
 * 
 * 该组合式函数是链表视图的主要入口点，集成了以下功能：
 * - 数据加载与处理（从 useChainTableData）
 * - 详情视图管理（从 useChainTableDetail）
 * - 统计数据处理（从 useChainTableStats）
 * - 格式化工具（从 useFormatters）
 * - 标签类型管理（从 useTagTypes）
 * 
 * @example
 * // 基本用法
 * const {
 *   loading,
 *   chainTableData,
 *   refreshData,
 *   selectChain,
 *   viewInterfaceRules
 * } = useChainTable()
 * 
 * // 加载数据
 * await refreshData()
 * 
 * // 选择链
 * selectChain('INPUT', filteredRules)
 */
export function useChainTable() {
    // 导入工具函数
    const { formatBytes, formatDate } = useFormatters()
    const { getTableTagType, getTargetTagType, getChainTagType } = useTagTypes()
    
    // 导入子组合式函数
    const { 
        loading, 
        chainTableData, 
        interfaces, 
        refreshData, 
        autoSyncSystemRules, 
        loadChainTableData, 
        loadInterfaces 
    } = useChainTableData()
    
    const { 
        showChainDialog, 
        detailTitle, 
        detailRules, 
        groupByChain, 
        selectedChain, 
        sortedDetailRules, 
        groupedRules, 
        selectChain: selectChainBase, 
        selectChainInTable, 
        closeChainDialog, 
        viewInterfaceRules: viewInterfaceRulesBase 
    } = useChainTableDetail()
    
    const { 
        getChainRuleCount, 
        getFilteredChainRuleCount, 
        getInterfaceRuleCount 
    } = useChainTableStats()
    
    // 基本状态
    const viewMode = ref<ViewMode>('chain')
    
    // 计算属性
    const chains = computed(() => chainTableData.value.chains || [])
    const tables = computed(() => chainTableData.value.tables || [])

    // 封装选择链函数，传入当前链数据
    const selectChain = (chainName: string, filteredRules: IPTablesRule[]) => {
        selectChainBase(chainName, chains.value, filteredRules)
    }
    
    // 封装查看接口规则函数
    const viewInterfaceRules = (interfaceName: string, filteredRules: IPTablesRule[]) => {
        viewInterfaceRulesBase(interfaceName, filteredRules)
    }

    // 处理视图模式变化
    const handleViewModeChange = () => {
        selectedChain.value = ''
    }

    return {
        // 状态
        loading,
        viewMode,
        selectedChain,
        showChainDialog,
        detailTitle,
        detailRules,
        groupByChain,
        chainTableData,
        interfaces,

        // 计算属性
        chains,
        tables,
        sortedDetailRules,
        groupedRules,

        // 方法
        getChainRuleCount,
        getFilteredChainRuleCount,
        getInterfaceRuleCount,
        selectChain,
        selectChainInTable,
        closeChainDialog,
        refreshData,
        autoSyncSystemRules,
        loadChainTableData,
        loadInterfaces,
        handleViewModeChange,
        formatBytes,
        formatDate,
        getTableTagType,
        getTargetTagType,
        getChainTagType,
        viewInterfaceRules
    }
}