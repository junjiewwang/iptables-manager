/**
 * TableItemFactory 单元测试
 */

import { describe, it, expect } from 'vitest'
import { TableItemFactory, TableConfigBuilder } from '../../utils/tableFactory'

describe('TableItemFactory', () => {
  describe('createRulesTableConfig', () => {
    it('should create basic rules table configuration', () => {
      const config = TableItemFactory.createRulesTableConfig()
      
      expect(config.columns).toHaveLength(8)
      expect(config.stripe).toBe(true)
      expect(config.border).toBe(true)
      expect(config.size).toBe('small')
      
      // 检查必要的列
      const columnProps = config.columns.map(col => col.prop)
      expect(columnProps).toContain('line_number')
      expect(columnProps).toContain('target')
      expect(columnProps).toContain('protocol')
      expect(columnProps).toContain('source')
      expect(columnProps).toContain('destination')
    })
  })

  describe('createDetailRulesTableConfig', () => {
    it('should create detailed rules table with chain and table columns', () => {
      const config = TableItemFactory.createDetailRulesTableConfig()
      
      expect(config.columns).toHaveLength(10) // 基础8列 + 链列 + 表列
      
      const columnProps = config.columns.map(col => col.prop)
      expect(columnProps).toContain('chain_name')
      expect(columnProps).toContain('table')
      
      // 检查列的顺序：链和表列应该在目标列之前
      const targetIndex = columnProps.indexOf('target')
      const chainIndex = columnProps.indexOf('chain_name')
      const tableIndex = columnProps.indexOf('table')
      
      expect(chainIndex).toBeLessThan(targetIndex)
      expect(tableIndex).toBeLessThan(targetIndex)
    })
  })

  describe('createGroupedRulesTableConfig', () => {
    it('should create grouped rules table with table column only', () => {
      const config = TableItemFactory.createGroupedRulesTableConfig()
      
      expect(config.columns).toHaveLength(9) // 基础8列 + 表列
      
      const columnProps = config.columns.map(col => col.prop)
      expect(columnProps).toContain('table')
      expect(columnProps).not.toContain('chain_name')
    })
  })

  describe('createInterfaceStatsTableConfig', () => {
    it('should create interface statistics table configuration', () => {
      const config = TableItemFactory.createInterfaceStatsTableConfig()
      
      const columnProps = config.columns.map(col => col.prop)
      expect(columnProps).toContain('name')
      expect(columnProps).toContain('type')
      expect(columnProps).toContain('state')
      expect(columnProps).toContain('ip_addresses')
      expect(columnProps).toContain('inRules')
      expect(columnProps).toContain('outRules')
      expect(columnProps).toContain('forwardRules')
    })
  })

  describe('addActionColumn', () => {
    it('should add action column to existing configuration', () => {
      const baseConfig = TableItemFactory.createRulesTableConfig()
      const actions = [
        { label: '编辑', type: 'primary' as const, handler: 'edit' },
        { label: '删除', type: 'danger' as const, handler: 'delete' }
      ]
      
      const configWithActions = TableItemFactory.addActionColumn(baseConfig, actions)
      
      expect(configWithActions.columns).toHaveLength(9) // 原8列 + 操作列
      
      const actionColumn = configWithActions.columns[configWithActions.columns.length - 1]
      expect(actionColumn.prop).toBe('actions')
      expect(actionColumn.label).toBe('操作')
      expect(actionColumn.fixed).toBe('right')
      expect(actionColumn.width).toBe(160) // 2 * 60 + 40
    })
  })

  describe('createCustomTableConfig', () => {
    it('should create custom table configuration', () => {
      const columns = [
        { prop: 'name', label: '名称', width: 100 },
        { prop: 'value', label: '值', width: 200 }
      ]
      const options = { stripe: false, border: false }
      
      const config = TableItemFactory.createCustomTableConfig(columns, options)
      
      expect(config.columns).toEqual(columns)
      expect(config.stripe).toBe(false)
      expect(config.border).toBe(false)
      expect(config.size).toBe('small') // 默认值
    })
  })
})

describe('TableConfigBuilder', () => {
  it('should build table configuration step by step', () => {
    const config = new TableConfigBuilder()
      .addColumn({ prop: 'name', label: '名称', width: 100 })
      .addColumn({ prop: 'type', label: '类型', width: 80 })
      .setOptions({ stripe: false, size: 'large' })
      .build()
    
    expect(config.columns).toHaveLength(2)
    expect(config.columns[0].prop).toBe('name')
    expect(config.columns[1].prop).toBe('type')
    expect(config.stripe).toBe(false)
    expect(config.size).toBe('large')
  })

  it('should add multiple columns at once', () => {
    const columns = [
      { prop: 'col1', label: 'Column 1' },
      { prop: 'col2', label: 'Column 2' }
    ]
    
    const config = new TableConfigBuilder()
      .addColumns(columns)
      .build()
    
    expect(config.columns).toHaveLength(2)
    expect(config.columns).toEqual(columns)
  })

  it('should reset builder state', () => {
    const builder = new TableConfigBuilder()
      .addColumn({ prop: 'test', label: 'Test' })
      .setOptions({ stripe: false })
    
    const configBefore = builder.build()
    expect(configBefore.columns).toHaveLength(1)
    expect(configBefore.stripe).toBe(false)
    
    builder.reset()
    const configAfter = builder.build()
    expect(configAfter.columns).toHaveLength(0)
    expect(configAfter.stripe).toBe(true) // 默认值
  })
})