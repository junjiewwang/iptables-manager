/**
 * API服务工厂 Composable
 * 提供统一的API服务创建和管理
 * 
 * 该组合式函数实现了工厂模式，用于创建和管理不同类型的API服务：
 * - 链表API服务（chainTableService）
 * - 网络接口API服务（networkService）
 * - 系统API服务（systemService）
 * 
 * 每个服务都封装了相关的API调用，并使用 useErrorHandler 进行统一的错误处理
 * 使用动态导入实现代码分割，减小初始加载体积
 * 
 * @example
 * // 基本用法
 * const { chainTableService, networkService, systemService } = useApiServiceFactory()
 * 
 * // 使用链表服务
 * const tables = await chainTableService.getAllTables()
 * 
 * // 使用网络服务
 * const interfaces = await networkService.getInterfaces()
 * 
 * // 使用系统服务
 * await systemService.syncSystemRules()
 */

import { ref, shallowRef } from 'vue'
import { useErrorHandler } from '@/composables/core/useErrorHandler'

// 使用动态导入API模块
const apiServiceModule = () => import('@/api').then(module => module.apiService)
const networkAPIModule = () => import('@/api').then(module => module.networkAPI)
const tablesAPIModule = () => import('@/api').then(module => module.tablesAPI)

/**
 * API服务工厂 Hook
 * 
 * @returns 包含不同类型 API 服务的对象
 */
export function useApiServiceFactory() {
  const { wrapAsync } = useErrorHandler()
  
  // 使用shallowRef缓存API模块实例
  const apiServiceRef = shallowRef(null)
  const networkAPIRef = shallowRef(null)
  const tablesAPIRef = shallowRef(null)
  
  // 加载API模块的函数
  const loadApiService = async () => {
    if (!apiServiceRef.value) {
      apiServiceRef.value = await apiServiceModule()
    }
    return apiServiceRef.value
  }
  
  const loadNetworkAPI = async () => {
    if (!networkAPIRef.value) {
      networkAPIRef.value = await networkAPIModule()
    }
    return networkAPIRef.value
  }
  
  const loadTablesAPI = async () => {
    if (!tablesAPIRef.value) {
      tablesAPIRef.value = await tablesAPIModule()
    }
    return tablesAPIRef.value
  }

  /**
   * 链表API服务
   * 处理与链表、规则相关的API调用
   * @type {Object}
   */
  const chainTableService = {
    /**
     * 获取所有表数据
     * 
     * @async
     * @returns {Promise<any>} 表数据响应
     * @description 从后端获取所有表数据，包含链和规则信息
     */
    getAllTables: async () => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.getAllTables(),
        '获取所有表数据',
        { rethrow: true }
      )
    },

    /**
     * 获取表详情
     * 
     * @async
     * @param {string} tableName 表名称
     * @returns {Promise<any>} 表详情响应
     * @description 从后端获取指定表的详细信息
     */
    getTableDetail: async (tableName: string) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.getTableDetail(tableName),
        `获取表 ${tableName} 详情`,
        { rethrow: true }
      )
    },

    /**
     * 获取链详情
     * 
     * @async
     * @param {string} tableName 表名称
     * @param {string} chainName 链名称
     * @returns {Promise<any>} 链详情响应
     * @description 从后端获取指定表中指定链的详细信息
     */
    getChainDetail: async (tableName: string, chainName: string) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.getChainDetail(tableName, chainName),
        `获取链 ${chainName} 详情`,
        { rethrow: true }
      )
    },

    /**
     * 获取链详细信息
     * 
     * @async
     * @param {string} tableName 表名称
     * @param {string} chainName 链名称
     * @returns {Promise<any>} 链详细信息响应
     * @description 从后端获取指定表中指定链的详细信息
     */
    getChainVerbose: async (tableName: string, chainName: string) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.getChainVerbose(tableName, chainName),
        `获取链 ${chainName} 详情`,
        { rethrow: true }
      )
    },

    /**
     * 获取特殊链
     * 
     * @async
     * @returns {Promise<any>} 特殊链响应
     * @description 从后端获取特殊链信息
     */
    getSpecialChains: async () => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.getSpecialChains(),
        '获取特殊链',
        { rethrow: true }
      )
    },

    /**
     * 添加规则
     * 
     * @async
     * @param {any} rule 规则对象
     * @returns {Promise<any>} 添加规则的响应
     * @description 向后端发送请求添加新的规则
     */
    addRule: async (rule: any) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.addRule(rule),
        '添加规则',
        { rethrow: true }
      )
    },

    /**
     * 编辑规则
     * 
     * @async
     * @param {string} ruleId 规则ID
     * @param {any} rule 更新后的规则对象
     * @returns {Promise<any>} 编辑规则的响应
     * @description 向后端发送请求更新现有规则
     */
    editRule: async (ruleId: string, rule: any) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.editRule(ruleId, rule),
        '编辑规则',
        { rethrow: true }
      )
    },

    /**
     * 删除规则
     * 
     * @async
     * @param {string} ruleId 规则ID
     * @returns {Promise<any>} 删除规则的响应
     * @description 向后端发送请求删除指定的规则
     */
    deleteRule: async (ruleId: string) => {
      const tablesAPI = await loadTablesAPI()
      return await wrapAsync(
        async () => await tablesAPI.deleteRule(ruleId),
        '删除规则',
        { rethrow: true }
      )
    }
  }

  /**
   * 网络API服务
   * 处理与网络接口相关的API调用
   * @type {Object}
   */
  const networkService = {
    /**
     * 获取所有网络接口
     * 
     * @async
     * @returns {Promise<any>} 网络接口数据响应
     * @description 从后端获取所有网络接口信息
     */
    getInterfaces: async () => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getInterfaces(),
        '获取网络接口',
        { rethrow: true }
      )
    },

    /**
     * 获取Docker网桥
     * 
     * @async
     * @returns {Promise<any>} Docker网桥数据响应
     * @description 从后端获取Docker网桥信息
     */
    getDockerBridges: async () => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getDockerBridges(),
        '获取Docker网桥',
        { rethrow: true }
      )
    },

    /**
     * 获取网桥规则
     * 
     * @async
     * @param {string} bridgeName 网桥名称
     * @returns {Promise<any>} 网桥规则响应
     * @description 从后端获取指定网桥的规则
     */
    getBridgeRules: async (bridgeName: string) => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getBridgeRules(bridgeName),
        `获取网桥 ${bridgeName} 规则`,
        { rethrow: true }
      )
    },

    /**
     * 获取网络连接
     * 
     * @async
     * @returns {Promise<any>} 网络连接数据响应
     * @description 从后端获取网络连接信息
     */
    getNetworkConnections: async () => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getNetworkConnections(),
        '获取网络连接',
        { rethrow: true }
      )
    },

    /**
     * 获取路由表
     * 
     * @async
     * @returns {Promise<any>} 路由表数据响应
     * @description 从后端获取路由表信息
     */
    getRouteTable: async () => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getRouteTable(),
        '获取路由表',
        { rethrow: true }
      )
    },

    /**
     * 获取接口详情
     * 
     * @async
     * @param {string} interfaceName 网络接口名称
     * @returns {Promise<any>} 接口详情响应
     * @description 从后端获取指定网络接口的详细信息
     */
    getInterfaceDetail: async (interfaceName: string) => {
      const networkAPI = await loadNetworkAPI()
      return await wrapAsync(
        async () => await networkAPI.getInterfaceDetail(interfaceName),
        `获取接口 ${interfaceName} 详情`,
        { rethrow: true }
      )
    }
  }

  /**
   * 系统API服务
   * 处理与系统规则相关的API调用
   * @type {Object}
   */
  const systemService = {
    /**
     * 比对系统和数据库规则
     * 
     * @async
     * @returns {Promise<any>} 比对结果响应，包含 consistent 字段表示是否一致
     * @description 检查当前系统中的规则与数据库中存储的规则是否一致
     */
    compareSystemAndDatabaseRules: async () => {
      const apiService = await loadApiService()
      return await wrapAsync(
        async () => await apiService.compareSystemAndDatabaseRules(),
        '比对系统和数据库规则',
        { rethrow: true }
      )
    },

    /**
     * 同步系统规则
     * 
     * @async
     * @returns {Promise<any>} 同步结果响应
     * @description 将当前系统中的规则同步到数据库中，或将数据库中的规则同步到系统中
     */
    syncSystemRules: async () => {
      const apiService = await loadApiService()
      return await wrapAsync(
        async () => await apiService.syncSystemRules(),
        '同步系统规则',
        { rethrow: true }
      )
    },

    /**
     * 获取系统统计信息
     * 
     * @async
     * @returns {Promise<any>} 系统统计信息响应
     * @description 从后端获取系统统计信息
     */
    getStatistics: async () => {
      const apiService = await loadApiService()
      return await wrapAsync(
        async () => await apiService.getStatistics(),
        '获取系统统计信息',
        { rethrow: true }
      )
    },

    /**
     * 获取系统状态
     * 
     * @async
     * @returns {Promise<any>} 系统状态响应
     * @description 从后端获取系统状态信息
     */
    getSystemStatus: async () => {
      const apiService = await loadApiService()
      return await wrapAsync(
        async () => await apiService.getSystemStatus(),
        '获取系统状态',
        { rethrow: true }
      )
    },

    /**
     * 系统备份
     * 
     * @async
     * @returns {Promise<any>} 备份结果响应
     * @description 触发系统备份操作
     */
    backup: async () => {
      const apiService = await loadApiService()
      return await wrapAsync(
        async () => await apiService.backup(),
        '系统备份',
        { rethrow: true }
      )
    }
  }

  return {
    chainTableService,
    networkService,
    systemService
  }
}