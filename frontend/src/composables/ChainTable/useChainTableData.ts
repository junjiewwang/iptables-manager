/**
 * 链表数据加载 Composable
 * 处理链表数据的加载和转换
 * 
 * 该组合式函数负责以下功能：
 * - 从 API 加载链表数据
 * - 从 API 加载网络接口数据
 * - 自动比对并同步系统规则
 * - 管理数据加载状态
 * 
 * 使用 useApiServiceFactory 进行 API 调用，确保错误处理一致性
 * 
 * @example
 * // 基本用法
 * const { 
 *   loading, 
 *   chainTableData, 
 *   interfaces, 
 *   refreshData 
 * } = useChainTableData()
 * 
 * // 加载数据
 * await refreshData()
 * 
 * // 访问链表数据
 * console.log(chainTableData.value.chains)
 * console.log(chainTableData.value.tables)
 */

import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { ChainTableData, NetworkInterface } from '@/types/ChainTable'
import { useUtils } from '@/composables/core/useUtils'
import { useApiServiceFactory } from '@/composables/core/useApiServiceFactory'

/**
 * 链表数据加载 Hook
 */
export function useChainTableData() {
  // 导入工具函数
  const { convertApiDataToChainTableData } = useUtils()
  const { chainTableService, networkService, systemService } = useApiServiceFactory()

  // 数据状态
  const loading = ref(false)
  const chainTableData = ref<ChainTableData>({
    chains: [],
    tables: []
  })
  const interfaces = ref<NetworkInterface[]>([])

  /**
   * 刷新数据
   */
  const refreshData = async () => {
    loading.value = true
    try {
      // 自动比对并同步系统规则
      const synced = await autoSyncSystemRules()

      // 加载数据
      await Promise.all([
        loadChainTableData(),
        loadInterfaces()
      ])

      // 根据是否进行了同步显示不同的消息
      if (synced) {
        ElMessage.success('检测到数据不一致，已自动同步并刷新数据')
      } else {
        ElMessage.success('数据刷新成功')
      }

      return true
    } catch (error: any) {
      console.error('刷新数据失败:', error)
      ElMessage.error('数据刷新失败: ' + (error.message || '未知错误'))
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 自动比对并同步系统规则
   */
  const autoSyncSystemRules = async () => {
    // 先比对系统规则和数据库规则
    const compareResult = await systemService.compareSystemAndDatabaseRules()

    if (!compareResult.data.consistent) {
      // 如果不一致，自动同步
      await systemService.syncSystemRules()
      return true // 返回true表示进行了同步
    } else {
      return false // 返回false表示无需同步
    }
  }

  /**
   * 加载链表数据
   */
  const loadChainTableData = async () => {
    const response = await chainTableService.getAllTables()

    if (!response?.data) {
      throw new Error('API返回数据为空')
    }

    // 检查数据结构并转换
    if (Array.isArray(response.data)) {
      // API返回的是数组格式
      chainTableData.value = convertApiDataToChainTableData(response.data)
    } else if (response.data.chains && Array.isArray(response.data.chains)) {
      // 已经是目标格式
      chainTableData.value = response.data
    } else {
      throw new Error('API返回数据格式异常')
    }
    
    return chainTableData.value
  }

  /**
   * 加载网络接口
   */
  const loadInterfaces = async () => {
    const response = await networkService.getInterfaces()
    interfaces.value = response.data || []
    return interfaces.value
  }

  return {
    loading,
    chainTableData,
    interfaces,
    refreshData,
    autoSyncSystemRules,
    loadChainTableData,
    loadInterfaces
  }
}