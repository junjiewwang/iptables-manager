import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { useLayout } from '../useLayout'

// Mock dagre
vi.mock('@dagrejs/dagre', () => ({
  default: {
    graphlib: {
      Graph: vi.fn(() => ({
        setDefaultEdgeLabel: vi.fn(),
        setGraph: vi.fn(),
        setNode: vi.fn(),
        setEdge: vi.fn(),
        node: vi.fn(() => ({ x: 100, y: 100, width: 180, height: 80 }))
      }))
    },
    layout: vi.fn()
  }
}))

// Mock useVueFlow
vi.mock('@vue-flow/core', () => ({
  useVueFlow: () => ({
    findNode: vi.fn(() => ({
      dimensions: { width: 180, height: 80 }
    }))
  }),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right'
  }
}))

describe('useLayout', () => {
  it('应该正确初始化布局功能', () => {
    const { layout, relayout, previousDirection } = useLayout()
    
    expect(layout).toBeDefined()
    expect(relayout).toBeDefined()
    expect(previousDirection.value).toBe('TB')
  })

  it('应该正确应用布局算法', () => {
    const { layout } = useLayout()
    
    const nodes = [
      { id: '1', position: { x: 0, y: 0 } },
      { id: '2', position: { x: 0, y: 0 } }
    ]
    
    const edges = [
      { source: '1', target: '2' }
    ]
    
    const result = layout(nodes, edges, 'TB')
    
    expect(result).toHaveLength(2)
    expect(result[0]).toHaveProperty('position')
    expect(result[0]).toHaveProperty('targetPosition')
    expect(result[0]).toHaveProperty('sourcePosition')
  })

  it('应该根据方向设置正确的连接点位置', () => {
    const { layout } = useLayout()
    
    const nodes = [{ id: '1', position: { x: 0, y: 0 } }]
    const edges = []
    
    // 测试垂直布局
    const verticalResult = layout(nodes, edges, 'TB')
    expect(verticalResult[0].targetPosition).toBe('top')
    expect(verticalResult[0].sourcePosition).toBe('bottom')
    
    // 测试水平布局
    const horizontalResult = layout(nodes, edges, 'LR')
    expect(horizontalResult[0].targetPosition).toBe('left')
    expect(horizontalResult[0].sourcePosition).toBe('right')
  })

  it('应该正确处理重新布局', () => {
    const { layout, relayout, previousDirection } = useLayout()
    
    const nodes = [{ id: '1', position: { x: 0, y: 0 } }]
    const edges = []
    
    // 首次布局
    layout(nodes, edges, 'LR')
    expect(previousDirection.value).toBe('LR')
    
    // 重新布局应该使用之前的方向
    const result = relayout(nodes, edges)
    expect(result).toBeDefined()
  })
})