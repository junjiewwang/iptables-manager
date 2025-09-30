import dagre from '@dagrejs/dagre'
import { Position, useVueFlow } from '@vue-flow/core'
import { ref } from 'vue'

/**
 * Composable to run the layout algorithm on the graph.
 * It uses the `dagre` library to calculate the layout of the nodes and edges.
 * @returns Layout utility functions and state
 */
export function useLayout() {
  const { findNode } = useVueFlow()

  const graph = ref(new dagre.graphlib.Graph())
  const previousDirection = ref('TB')

  /**
   * Applies layout algorithm to nodes and edges
   * @param nodes - Array of nodes to layout
   * @param edges - Array of edges connecting the nodes
   * @param direction - Layout direction ('TB', 'BT', 'LR', 'RL')
   * @returns Array of nodes with updated positions and connection points
   */
  function layout(nodes, edges, direction) {
    // Create a new graph instance to handle node/edge removals
    const dagreGraph = new dagre.graphlib.Graph()
    graph.value = dagreGraph

    // Set default edge label function
    dagreGraph.setDefaultEdgeLabel(() => ({}))

    // Determine if layout is horizontal
    const isHorizontal = direction === 'LR' || direction === 'RL'
    
    // Configure graph layout direction and spacing
    dagreGraph.setGraph({ 
      rankdir: direction,
      // Adjust spacing based on layout direction
      nodesep: isHorizontal ? 100 : 80,  // Distance between nodes in same rank
      ranksep: isHorizontal ? 200 : 120, // Distance between ranks
      align: isHorizontal ? 'UL' : null, // Alignment for horizontal layout
      // Additional parameters for better layout
      marginx: 20,
      marginy: 20,
      acyclicer: 'greedy',     // Algorithm for making the graph acyclic
      ranker: 'network-simplex' // Ranking algorithm
    })

    previousDirection.value = direction

    console.log(`Applying ${isHorizontal ? 'horizontal' : 'vertical'} layout (${direction})`)

    // Add nodes to the graph with dimensions
    for (const node of nodes) {
      const graphNode = findNode(node.id)
      
      // Use actual node dimensions if available, otherwise use defaults
      // Add padding to ensure nodes don't overlap
      const width = (graphNode?.dimensions?.width || 180) + 20
      const height = (graphNode?.dimensions?.height || 100) + 20
      
      dagreGraph.setNode(node.id, { width, height })
    }

    // Add edges to the graph
    for (const edge of edges) {
      dagreGraph.setEdge(edge.source, edge.target)
    }

    // Run the layout algorithm
    dagre.layout(dagreGraph)

    // Update node positions and connection points based on layout direction
    return nodes.map((node) => {
      const nodeWithPosition = dagreGraph.node(node.id)
      
      // Set connection points based on layout direction
      let sourcePosition, targetPosition
      
      if (isHorizontal) {
        sourcePosition = direction === 'LR' ? Position.Right : Position.Left
        targetPosition = direction === 'LR' ? Position.Left : Position.Right
      } else {
        sourcePosition = direction === 'TB' ? Position.Bottom : Position.Top
        targetPosition = direction === 'TB' ? Position.Top : Position.Bottom
      }

      // Return node with updated position and connection points
      return {
        ...node,
        sourcePosition,
        targetPosition,
        position: { 
          x: Math.round(nodeWithPosition.x - (nodeWithPosition.width / 2)), 
          y: Math.round(nodeWithPosition.y - (nodeWithPosition.height / 2))
        },
      }
    })
  }

  return { graph, layout, previousDirection }
}