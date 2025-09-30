/**
 * 表格项工厂模式
 * 用于创建不同类型的表格配置项
 */

import type { IPTablesRule } from '@/types/ChainTable/types'

/**
 * 表格列配置接口
 */
export interface TableColumnConfig {
  prop: string
  label: string
  width?: string | number
  minWidth?: string | number
  fixed?: boolean | 'left' | 'right'
  sortable?: boolean
  showOverflowTooltip?: boolean
  formatter?: (row: any, column: any, cellValue: any, index: number) => string
  renderHeader?: (h: any, { column, $index }: any) => any
}

/**
 * 表格配置接口
 */
export interface TableConfig {
  columns: TableColumnConfig[]
  showSelection?: boolean
  showIndex?: boolean
  stripe?: boolean
  border?: boolean
  size?: 'large' | 'default' | 'small'
}

/**
 * 表格项工厂类
 */
export class TableItemFactory {
  /**
   * 创建基础规则表格配置
   */
  static createRulesTableConfig(): TableConfig {
    return {
      columns: [
        {
          prop: 'line_number',
          label: '#',
          width: 60
        },
        {
          prop: 'target',
          label: '目标',
          width: 120
        },
        {
          prop: 'protocol',
          label: '协议',
          width: 80
        },
        {
          prop: 'source',
          label: '源地址',
          width: 140
        },
        {
          prop: 'destination',
          label: '目标地址',
          width: 140
        },
        {
          prop: 'in_interface',
          label: '入接口',
          width: 100
        },
        {
          prop: 'out_interface',
          label: '出接口',
          width: 100
        },
        {
          prop: 'options',
          label: '选项',
          minWidth: 200,
          showOverflowTooltip: true
        }
      ],
      stripe: true,
      border: true,
      size: 'small'
    }
  }

  /**
   * 创建详细规则表格配置（包含链和表列）
   */
  static createDetailRulesTableConfig(): TableConfig {
    const baseConfig = this.createRulesTableConfig()
    
    // 在目标列前插入链和表列
    const targetIndex = baseConfig.columns.findIndex(col => col.prop === 'target')
    baseConfig.columns.splice(targetIndex, 0, 
      {
        prop: 'chain_name',
        label: '链',
        width: 120
      },
      {
        prop: 'table',
        label: '表',
        width: 80
      }
    )

    return baseConfig
  }

  /**
   * 创建分组规则表格配置（只包含表列）
   */
  static createGroupedRulesTableConfig(): TableConfig {
    const baseConfig = this.createRulesTableConfig()
    
    // 在目标列前插入表列
    const targetIndex = baseConfig.columns.findIndex(col => col.prop === 'target')
    baseConfig.columns.splice(targetIndex, 0, {
      prop: 'table',
      label: '表',
      width: 80
    })

    return baseConfig
  }

  /**
   * 创建接口统计表格配置
   */
  static createInterfaceStatsTableConfig(): TableConfig {
    return {
      columns: [
        {
          prop: 'name',
          label: '接口名称',
          width: 120
        },
        {
          prop: 'type',
          label: '类型',
          width: 100
        },
        {
          prop: 'state',
          label: '状态',
          width: 80
        },
        {
          prop: 'ip_addresses',
          label: 'IP地址',
          minWidth: 150,
          formatter: (row) => row.ip_addresses?.join(', ') || '无'
        },
        {
          prop: 'inRules',
          label: '入站规则',
          width: 100
        },
        {
          prop: 'outRules',
          label: '出站规则',
          width: 100
        },
        {
          prop: 'forwardRules',
          label: '转发规则',
          width: 100
        }
      ],
      stripe: true,
      border: true,
      size: 'small'
    }
  }

  /**
   * 创建自定义表格配置
   */
  static createCustomTableConfig(
    columns: TableColumnConfig[],
    options: Partial<TableConfig> = {}
  ): TableConfig {
    return {
      columns,
      stripe: true,
      border: true,
      size: 'small',
      ...options
    }
  }

  /**
   * 添加操作列到表格配置
   */
  static addActionColumn(
    config: TableConfig,
    actions: Array<{
      label: string
      type?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
      handler: string
    }>
  ): TableConfig {
    const actionColumn: TableColumnConfig = {
      prop: 'actions',
      label: '操作',
      width: actions.length * 60 + 40,
      fixed: 'right'
    }

    return {
      ...config,
      columns: [...config.columns, actionColumn]
    }
  }

  /**
   * 创建表格列的渲染函数
   */
  static createColumnRenderer(
    type: 'tag' | 'badge' | 'link' | 'button',
    options: any = {}
  ) {
    switch (type) {
      case 'tag':
        return (value: string) => ({
          type: 'tag',
          props: {
            type: options.getType?.(value) || 'default',
            size: options.size || 'small'
          },
          children: value
        })
      
      case 'badge':
        return (value: number) => ({
          type: 'badge',
          props: {
            value,
            hidden: !value
          }
        })
      
      case 'link':
        return (value: string, row: any) => ({
          type: 'link',
          props: {
            type: 'primary',
            onClick: () => options.onClick?.(row)
          },
          children: value
        })
      
      case 'button':
        return (value: string, row: any) => ({
          type: 'button',
          props: {
            size: 'small',
            type: options.type || 'primary',
            onClick: () => options.onClick?.(row)
          },
          children: options.label || value
        })
      
      default:
        return (value: any) => value
    }
  }
}

/**
 * 表格项构建器类
 */
export class TableConfigBuilder {
  private config: TableConfig

  constructor() {
    this.config = {
      columns: [],
      stripe: true,
      border: true,
      size: 'small'
    }
  }

  /**
   * 添加列
   */
  addColumn(column: TableColumnConfig): this {
    this.config.columns.push(column)
    return this
  }

  /**
   * 添加多列
   */
  addColumns(columns: TableColumnConfig[]): this {
    this.config.columns.push(...columns)
    return this
  }

  /**
   * 设置表格选项
   */
  setOptions(options: Partial<TableConfig>): this {
    Object.assign(this.config, options)
    return this
  }

  /**
   * 构建配置
   */
  build(): TableConfig {
    return { ...this.config }
  }

  /**
   * 重置构建器
   */
  reset(): this {
    this.config = {
      columns: [],
      stripe: true,
      border: true,
      size: 'small'
    }
    return this
  }
}