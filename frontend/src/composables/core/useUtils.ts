/**
 * 通用工具函数 Composable
 * 提供各种通用工具函数
 */

import type { IPTablesRule } from '@/types/ChainTable'

/**
 * 通用工具函数 Hook
 */
export function useUtils() {
  /**
   * 按行号排序规则
   * @param rules 规则数组
   * @returns 排序后的规则数组
   */
  const sortRulesByLineNumber = (rules: IPTablesRule[]): IPTablesRule[] => {
    return [...rules].sort((a, b) => {
      const lineA = parseInt(a.line_number || '0', 10)
      const lineB = parseInt(b.line_number || '0', 10)
      return lineA - lineB
    })
  }

  /**
   * 格式化规则数据
   * @param rule 规则对象
   * @param index 索引
   * @param chainName 链名
   * @param tableName 表名
   * @returns 格式化后的规则对象
   */
  const formatRuleData = (
    rule: Partial<IPTablesRule>, 
    index: number, 
    chainName?: string, 
    tableName?: string
  ): IPTablesRule => ({
    ...rule as IPTablesRule,
    line_number: rule.line_number || (index + 1).toString(),
    chain_name: rule.chain_name || chainName || '未指定链',
    table: rule.table || tableName || 'filter'
  })

  /**
   * 转换API数据格式为ChainTableData
   * @param tableDataArray API返回的表数据数组
   * @returns 转换后的ChainTableData对象
   */
  const convertApiDataToChainTableData = (tableDataArray: any[]): any => {
    const convertedData: any = {
      chains: [],
      tables: []
    }

    // 用于去重链名
    const chainMap = new Map()

    // 处理每个表的数据
    tableDataArray.forEach((tableItem: any) => {
      if (!tableItem?.table_name || !Array.isArray(tableItem.chains)) return

      // 添加表信息
      convertedData.tables.push({
        name: tableItem.table_name,
        total_rules: tableItem.chains.reduce((total: number, chain: any) =>
          total + (chain.rules?.length || 0), 0),
        chains: tableItem.chains.map((chain: any) => ({
          name: chain.chain_name,
          policy: chain.policy || 'ACCEPT',
          rules: chain.rules || []
        }))
      })

      // 处理链数据，合并相同链名的规则
      tableItem.chains.forEach((chain: any) => {
        if (!chain?.chain_name) return

        const chainKey = chain.chain_name
        if (!chainMap.has(chainKey)) {
          chainMap.set(chainKey, {
            name: chain.chain_name,
            policy: chain.policy || 'ACCEPT',
            rules: [],
            tables: []
          })
        }

        const existingChain = chainMap.get(chainKey)
        // 添加规则
        if (Array.isArray(chain.rules)) {
          existingChain.rules.push(...chain.rules)
        }
        // 添加表名
        if (!existingChain.tables.includes(tableItem.table_name)) {
          existingChain.tables.push(tableItem.table_name)
        }
      })
    })

    // 将Map转换为数组
    convertedData.chains = Array.from(chainMap.values())

    return convertedData
  }

  return {
    sortRulesByLineNumber,
    formatRuleData,
    convertApiDataToChainTableData
  }
}