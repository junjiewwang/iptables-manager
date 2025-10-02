import dagre from '@dagrejs/dagre'
import { Position, useVueFlow } from '@vue-flow/core'

interface LayoutParams {
  nodes: any[]
  edges: any[]
  direction: string
  nodeSpacing?: number
  rankSpacing?: number
}

/**
 * 布局计算，仅支持水平(LR)。保留原API但忽略传入的direction，统一为LR。
 */
export function useLayout() {
  const { findNode } = useVueFlow()

  let graph: any = new dagre.graphlib.Graph()

  function layout(options: LayoutParams) {
    const { nodes, edges, nodeSpacing = 50, rankSpacing = 80 } = options

    // 始终新建图，避免历史残留
    const dagreGraph = new dagre.graphlib.Graph()
    graph = dagreGraph

    dagreGraph.setDefaultEdgeLabel(() => ({}))

    // 固定为水平布局(LR)
    dagreGraph.setGraph({
      rankdir: 'LR',
      nodesep: rankSpacing, // 同一层节点间距
      ranksep: nodeSpacing, // 层级间距
    })

    // 节点尺寸
    for (const node of nodes) {
      const graphNode = findNode(node.id)
      const width = (graphNode?.dimensions?.width || 180) + 20
      const height = (graphNode?.dimensions?.height || 100) + 20
      dagreGraph.setNode(node.id, { width, height })
    }

    // 边
    for (const edge of edges) {
      dagreGraph.setEdge(edge.source, edge.target)
    }

    // 运行布局
    dagre.layout(dagreGraph)

    // 返回带位置与把手的节点（LR: 右→左）
    return nodes.map((node) => {
      const nodeWithPosition = dagreGraph.node(node.id)
      return {
        ...node,
        targetPosition: Position.Left,
        sourcePosition: Position.Right,
        position: {
          x: nodeWithPosition.x,
          y: nodeWithPosition.y,
        },
      }
    })
  }

  return { graph, layout }
}