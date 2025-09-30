/**
 * 标签类型管理 Composable
 * 统一管理表格、目标、链等标签的类型映射
 */

export interface TagTypeConfig {
  [key: string]: string
}

/**
 * 表格标签类型配置
 */
const TABLE_TAG_TYPES: TagTypeConfig = {
  'raw': 'info',
  'mangle': 'warning', 
  'nat': 'success',
  'filter': 'danger'
}

/**
 * 目标标签类型配置
 */
const TARGET_TAG_TYPES: TagTypeConfig = {
  'ACCEPT': 'success',
  'DROP': 'danger',
  'REJECT': 'warning',
  'RETURN': 'info',
  'MASQUERADE': 'primary',
  'SNAT': 'primary',
  'DNAT': 'primary'
}

/**
 * 链标签类型配置
 */
const CHAIN_TAG_TYPES: TagTypeConfig = {
  'PREROUTING': 'primary',
  'INPUT': 'success',
  'FORWARD': 'warning',
  'OUTPUT': 'info',
  'POSTROUTING': 'danger'
}

/**
 * 标签类型管理 Hook
 */
export function useTagTypes() {
  /**
   * 获取表的标签类型
   */
  const getTableTagType = (tableName: string): string => {
    return TABLE_TAG_TYPES[tableName] || 'default'
  }

  /**
   * 获取目标的标签类型
   */
  const getTargetTagType = (target: string): string => {
    return TARGET_TAG_TYPES[target] || 'default'
  }

  /**
   * 获取链的标签类型
   */
  const getChainTagType = (chainName: string): string => {
    return CHAIN_TAG_TYPES[chainName] || 'default'
  }

  /**
   * 获取分组中的表名列表
   */
  const getTablesInGroup = (rules: any[]): string[] => {
    const tables = new Set(rules.map(rule => rule.table).filter(Boolean))
    return Array.from(tables)
  }

  return {
    getTableTagType,
    getTargetTagType,
    getChainTagType,
    getTablesInGroup,
    // 导出配置供其他地方使用
    TABLE_TAG_TYPES,
    TARGET_TAG_TYPES,
    CHAIN_TAG_TYPES
  }
}