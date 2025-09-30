/**
 * MainTable 事件处理 Composable
 * 统一管理MainTable组件的事件处理逻辑
 */

import type { IPTablesRule } from '@/types/ChainTable/types'

export interface MainTableEvents {
  onSelectChainTable: (tableName: string, chainName: string) => void
  onCloseChainDialog: () => void
  onEditRule: (rule: IPTablesRule) => void
  onDeleteRule: (rule: IPTablesRule) => void
  onAddRule: (chainName: string, tableName?: string) => void
  onViewInterfaceRules: (interfaceName: string) => void
}

/**
 * MainTable 事件处理 Hook
 */
export function useMainTableEvents(emit: any): MainTableEvents {
  // 移除了拓扑图相关的事件处理函数

  /**
   * 选择链和表事件
   */
  const onSelectChainTable = (tableName: string, chainName: string) => {
    emit('select-ChainTable', tableName, chainName)
  }

  /**
   * 关闭链详情对话框事件
   */
  const onCloseChainDialog = () => {
    emit('close-chain-dialog')
  }

  /**
   * 编辑规则事件
   */
  const onEditRule = (rule: IPTablesRule) => {
    emit('edit-rule', rule)
  }

  /**
   * 删除规则事件
   */
  const onDeleteRule = (rule: IPTablesRule) => {
    emit('delete-rule', rule)
  }

  /**
   * 添加规则事件
   */
  const onAddRule = (chainName: string, tableName?: string) => {
    emit('add-rule', chainName, tableName)
  }

  /**
   * 查看接口规则事件
   */
  const onViewInterfaceRules = (interfaceName: string) => {
    emit('view-interface-rules', interfaceName)
  }


  // 移除了拓扑图设置重置函数

  return {
    onSelectChainTable,
    onCloseChainDialog,
    onEditRule,
    onDeleteRule,
    onAddRule,
    onViewInterfaceRules
  }
}