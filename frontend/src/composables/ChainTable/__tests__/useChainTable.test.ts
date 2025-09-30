import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useChainTable } from '../useChainTable'
import { getMockChainTableData, getMockInterfaces } from './mockData'

// Mock API services
vi.mock('@/api', () => ({
    apiService: {
        compareSystemAndDatabaseRules: vi.fn(),
        syncSystemRules: vi.fn()
    },
    tablesAPI: {
        getAllTables: vi.fn()
    },
    networkAPI: {
        getInterfaces: vi.fn()
    }
}))

// Mock Element Plus
vi.mock('element-plus', () => ({
    ElMessage: {
        success: vi.fn(),
        error: vi.fn(),
        warning: vi.fn()
    }
}))

describe('useChainTable', () => {
    let chainTable: ReturnType<typeof useChainTable>

    beforeEach(() => {
        vi.clearAllMocks()
        chainTable = useChainTable()
    })

    describe('初始化状态', () => {
        it('应该正确初始化所有状态', () => {
            expect(chainTable.loading.value).toBe(false)
            expect(chainTable.viewMode.value).toBe('chain')
            expect(chainTable.selectedChain.value).toBe('')
            expect(chainTable.showChainDialog.value).toBe(false)
            expect(chainTable.detailTitle.value).toBe('')
            expect(chainTable.detailRules.value).toEqual([])
            expect(chainTable.groupByChain.value).toBe(true)
            expect(chainTable.chainTableData.value).toEqual({
                chains: [],
                tables: []
            })
            expect(chainTable.interfaces.value).toEqual([])
        })
    })

    describe('工具函数', () => {
        it('应该正确获取链规则数量', () => {
            const mockData = getMockChainTableData()
            chainTable.chainTableData.value = mockData

            const inputRuleCount = chainTable.getChainRuleCount('INPUT')
            expect(inputRuleCount).toBe(3)

            const nonExistentRuleCount = chainTable.getChainRuleCount('NON_EXISTENT')
            expect(nonExistentRuleCount).toBe(0)
        })

        it('应该正确获取筛选后的链规则数量', () => {
            const mockData = getMockChainTableData()
            const filteredRules = mockData.chains.find(c => c.name === 'INPUT')?.rules || []

            const count = chainTable.getFilteredChainRuleCount('INPUT', filteredRules)
            expect(count).toBe(3)
        })

        it('应该正确获取接口规则数量', () => {
            const mockData = getMockChainTableData()
            const allRules = mockData.chains.flatMap(c => c.rules || [])

            const inCount = chainTable.getInterfaceRuleCount('eth0', 'in', allRules)
            expect(inCount).toBe(2) // PREROUTING 和 FORWARD 链中各有一个规则的 in_interface 是 eth0

            const outCount = chainTable.getInterfaceRuleCount('eth0', 'out', allRules)
            expect(outCount).toBe(1) // POSTROUTING 链中有一个规则的 out_interface 是 eth0

            const forwardCount = chainTable.getInterfaceRuleCount('eth0', 'forward', allRules)
            expect(forwardCount).toBe(1) // FORWARD 链中有一个规则涉及 eth0
        })
    })

    describe('格式化工具', () => {
        it('应该正确格式化字节数', () => {
            expect(chainTable.formatBytes(0)).toBe('0 B')
            expect(chainTable.formatBytes(1024)).toBe('1 KB')
            expect(chainTable.formatBytes(1048576)).toBe('1 MB')
            expect(chainTable.formatBytes(1073741824)).toBe('1 GB')
        })

        it('应该正确格式化日期', () => {
            const testDate = '2023-01-01T12:00:00Z'
            const formatted = chainTable.formatDate(testDate)
            expect(formatted).toContain('2023')
            
            expect(chainTable.formatDate('')).toBe('-')
        })
    })

    describe('标签类型获取', () => {
        it('应该正确获取表标签类型', () => {
            expect(chainTable.getTableTagType('raw')).toBe('info')
            expect(chainTable.getTableTagType('mangle')).toBe('warning')
            expect(chainTable.getTableTagType('nat')).toBe('success')
            expect(chainTable.getTableTagType('filter')).toBe('danger')
            expect(chainTable.getTableTagType('unknown')).toBe('default')
        })

        it('应该正确获取目标标签类型', () => {
            expect(chainTable.getTargetTagType('ACCEPT')).toBe('success')
            expect(chainTable.getTargetTagType('DROP')).toBe('danger')
            expect(chainTable.getTargetTagType('REJECT')).toBe('warning')
            expect(chainTable.getTargetTagType('RETURN')).toBe('info')
            expect(chainTable.getTargetTagType('MASQUERADE')).toBe('primary')
            expect(chainTable.getTargetTagType('UNKNOWN')).toBe('default')
        })

        it('应该正确获取链标签类型', () => {
            expect(chainTable.getChainTagType('PREROUTING')).toBe('primary')
            expect(chainTable.getChainTagType('INPUT')).toBe('success')
            expect(chainTable.getChainTagType('FORWARD')).toBe('warning')
            expect(chainTable.getChainTagType('OUTPUT')).toBe('info')
            expect(chainTable.getChainTagType('POSTROUTING')).toBe('danger')
            expect(chainTable.getChainTagType('UNKNOWN')).toBe('default')
        })
    })

    describe('链选择功能', () => {
        it('应该正确选择链', () => {
            const mockData = getMockChainTableData()
            chainTable.chainTableData.value = mockData
            
            const inputRules = mockData.chains.find(c => c.name === 'INPUT')?.rules || []
            chainTable.selectChain('INPUT', inputRules)

            expect(chainTable.selectedChain.value).toBe('INPUT')
            expect(chainTable.showChainDialog.value).toBe(true)
            expect(chainTable.detailTitle.value).toBe('INPUT 链详细规则')
            expect(chainTable.detailRules.value).toHaveLength(3)
        })

        it('应该正确选择表中的链', () => {
            const mockData = getMockChainTableData()
            chainTable.chainTableData.value = mockData

            chainTable.selectChainInTable('filter', 'INPUT')

            expect(chainTable.showChainDialog.value).toBe(true)
            expect(chainTable.detailTitle.value).toBe('FILTER.INPUT 详细规则')
            expect(chainTable.detailRules.value).toHaveLength(3)
        })

        it('应该正确关闭链详情对话框', () => {
            chainTable.showChainDialog.value = true
            chainTable.detailRules.value = [{ id: 1, rule_text: 'test' }]
            chainTable.groupByChain.value = false

            chainTable.closeChainDialog()

            expect(chainTable.showChainDialog.value).toBe(false)
            expect(chainTable.detailRules.value).toEqual([])
            expect(chainTable.groupByChain.value).toBe(true)
        })
    })

    describe('视图模式处理', () => {
        it('应该正确处理视图模式变化', () => {
            chainTable.selectedChain.value = 'INPUT'
            
            chainTable.handleViewModeChange()
            
            expect(chainTable.selectedChain.value).toBe('')
        })
    })

    describe('接口规则查看', () => {
        it('应该正确查看接口规则', () => {
            const mockData = getMockChainTableData()
            const allRules = mockData.chains.flatMap(c => c.rules || [])
            
            chainTable.viewInterfaceRules('eth0', allRules)

            expect(chainTable.selectedChain.value).toBe('接口 eth0')
            expect(chainTable.detailTitle.value).toBe('接口 eth0 相关规则')
            expect(chainTable.showChainDialog.value).toBe(true)
            expect(chainTable.detailRules.value.length).toBeGreaterThan(0)
        })
    })

    describe('计算属性', () => {
        it('应该正确计算链和表', () => {
            const mockData = getMockChainTableData()
            chainTable.chainTableData.value = mockData

            expect(chainTable.chains.value).toHaveLength(5)
            expect(chainTable.tables.value).toHaveLength(4)
        })

        it('应该正确排序详细规则', () => {
            chainTable.detailRules.value = [
                { id: 2, line_number: '2', rule_text: 'rule 2' },
                { id: 1, line_number: '1', rule_text: 'rule 1' },
                { id: 3, line_number: '3', rule_text: 'rule 3' }
            ]

            const sorted = chainTable.sortedDetailRules.value
            expect(sorted[0].line_number).toBe('1')
            expect(sorted[1].line_number).toBe('2')
            expect(sorted[2].line_number).toBe('3')
        })

        it('应该正确按链分组规则', () => {
            chainTable.detailRules.value = [
                { id: 1, chain_name: 'INPUT', rule_text: 'input rule 1' },
                { id: 2, chain_name: 'OUTPUT', rule_text: 'output rule 1' },
                { id: 3, chain_name: 'INPUT', rule_text: 'input rule 2' }
            ]

            const grouped = chainTable.groupedRules.value
            expect(grouped['INPUT']).toHaveLength(2)
            expect(grouped['OUTPUT']).toHaveLength(1)
        })
    })
})