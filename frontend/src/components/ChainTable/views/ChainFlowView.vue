<template>
  <div class="dataflow-view">
    <div class="vue-flow-wrapper">
      <VueFlow
        :nodes="nodes"
        :edges="edges"
        :default-viewport="{ zoom: 0.8 }"
        :min-zoom="0.5"
        :max-zoom="2"
        :snap-to-grid="true"
        :snap-grid="[20, 20]"
        :node-draggable="true"
        :auto-connect="false"
        :connection-mode="ConnectionMode.Strict"
        :fit-view-on-init="false"
        :elevate-edges-on-select="true"
        class="dataflow-diagram"
        @node-click="onNodeClick"
        @edge-click="onEdgeClick"
        @nodes-initialized="onNodesInitialized"
      >
        <!-- 背景 -->
        <Background 
          :pattern-color="'#e2e8f0'" 
          :gap="20" 
          variant="lines"
        />
        
        <!-- 控制面板 -->
        <Controls />
        
        <!-- 小地图 -->
        <MiniMap />

        <!-- 布局控制面板 -->
        <Panel position="top-right" class="layout-control-panel">
          <LayoutPanel
            :current-direction="layoutDirection"
            :node-spacing="layoutSettings.nodeSpacing"
            :rank-spacing="layoutSettings.rankSpacing"
            :animate-layout="layoutSettings.animateLayout"
            @layout-change="onLayoutChange"
            @fit-view="onFitView"
            @reset-layout="onResetLayout"
            @spacing-change="onSpacingChange"
            @animate-toggle="onAnimateToggle"
          />
        </Panel>

        <!-- 自定义节点模板 -->
        <template #node-network="nodeProps">
          <NetworkNode :nodeProps="nodeProps" />
        </template>

        <template #node-chain="nodeProps">
          <ChainNode
            :nodeProps="nodeProps"
            @select-chain-table="onSelectChainTable"
          />
        </template>

        <template #node-decision="nodeProps">
          <DecisionNode :nodeProps="nodeProps" />
        </template>

        <template #node-process="nodeProps">
          <ProcessNode :nodeProps="nodeProps" />
        </template>

        <!-- 自定义边模板 -->
        <template #edge-flow="flowEdgeProps">
          <FlowEdge
            :id="flowEdgeProps.id"
            :source-x="flowEdgeProps.sourceX"
            :source-y="flowEdgeProps.sourceY"
            :target-x="flowEdgeProps.targetX"
            :target-y="flowEdgeProps.targetY"
            :source-position="flowEdgeProps.sourcePosition"
            :target-position="flowEdgeProps.targetPosition"
            :data="flowEdgeProps.data"
            :marker-end="flowEdgeProps.markerEnd"
            :style="flowEdgeProps.style"
          />
        </template>
      </VueFlow>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick } from 'vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'
import { VueFlow, MarkerType, Panel, useVueFlow, Position, ConnectionMode } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import FlowEdge from './edges/FlowEdge.vue'
import LayoutPanel from '../LayoutPanel.vue'
import { useLayout } from '@/composables/ChainTable/useLayout'
import ChainNode from './nodes/ChainNode.vue'
import NetworkNode from './nodes/NetworkNode.vue'
import DecisionNode from './nodes/DecisionNode.vue'
import ProcessNode from './nodes/ProcessNode.vue'

interface Props {
  flowElements?: any[]
  topoSettings?: any
}

interface Emits {
  (e: 'update:flowElements', value: any[]): void
  (e: 'node-click', event: any): void
  (e: 'edge-click', event: any): void
  (e: 'select-chain-table', tableName: string, chainName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 使用VueFlow和布局功能
const { fitView } = useVueFlow()
const { layout } = useLayout()

// 布局设置
const layoutDirection = ref('TB')
const layoutSettings = reactive({
  nodeSpacing: 80,
  rankSpacing: 120,
  animateLayout: true
})

// 初始节点位置（用于重置）
const initialNodePositions = new Map()

// 定义网络数据包处理流程的节点
const nodes = ref([
  // 外部网络
  {
    id: 'external-network',
    type: 'network',
    position: { x: 50, y: 50 },
    data: {
      label: '外部网络',
      icon: 'globe',
      nodeType: 'network',
      color: '#ffffff',
      borderColor: '#d1d5db',
      description: '来自外部的网络流量'
    },
    style: {
      width: 160,
      height: 90,
      borderRadius: '16px',
      border: '2px solid #d1d5db',
      backgroundColor: '#ffffff',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // PREROUTING 链
  {
    id: 'prerouting',
    type: 'chain',
    position: { x: 300, y: 50 },
    data: {
      label: 'PREROUTING',
      chainName: 'PREROUTING',
      tables: ['raw', 'mangle', 'nat'],
      ruleCount: 3,
      color: '#fecaca',
      borderColor: '#ef4444',
      description: '路由前处理'
    },
    style: {
      width: 220,
      height: 140,
      borderRadius: '8px',
      border: '2px solid #ef4444',
      backgroundColor: '#fecaca',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // 路由决策
  {
    id: 'routing-decision',
    type: 'decision',
    position: { x: 550, y: 50 },
    data: {
      label: '路由决策',
      icon: 'route',
      color: '#fed7aa',
      borderColor: '#f97316',
      description: '确定数据包的目的地'
    },
    style: {
      width: 180,
      height: 120,
      borderRadius: '12px',
      border: '2px solid #f97316',
      backgroundColor: '#fed7aa',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // INPUT 链
  {
    id: 'input',
    type: 'chain',
    position: { x: 400, y: 200 },
    data: {
      label: 'INPUT',
      chainName: 'INPUT',
      tables: ['mangle', 'filter', 'nat'],
      ruleCount: 0,
      color: '#dcfce7',
      borderColor: '#22c55e',
      description: '发往本机的数据包'
    },
    style: {
      width: 220,
      height: 140,
      borderRadius: '8px',
      border: '2px solid #22c55e',
      backgroundColor: '#dcfce7',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // 本地处理
  {
    id: 'local-process',
    type: 'process',
    position: { x: 400, y: 350 },
    data: {
      label: '本地处理',
      icon: 'cog',
      color: '#e9d5ff',
      borderColor: '#a855f7',
      description: '本地应用程序处理'
    },
    style: {
      width: 180,
      height: 100,
      borderRadius: '8px',
      border: '2px solid #a855f7',
      backgroundColor: '#e9d5ff',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // OUTPUT 链
  {
    id: 'output',
    type: 'chain',
    position: { x: 400, y: 500 },
    data: {
      label: 'OUTPUT',
      chainName: 'OUTPUT',
      tables: ['raw', 'mangle', 'nat', 'filter'],
      ruleCount: 2,
      color: '#dcfce7',
      borderColor: '#22c55e',
      description: '本机发出的数据包'
    },
    style: {
      width: 220,
      height: 140,
      borderRadius: '8px',
      border: '2px solid #22c55e',
      backgroundColor: '#dcfce7',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // FORWARD 链
  {
    id: 'forward',
    type: 'chain',
    position: { x: 700, y: 200 },
    data: {
      label: 'FORWARD',
      chainName: 'FORWARD',
      tables: ['mangle', 'filter'],
      ruleCount: 41,
      color: '#dbeafe',
      borderColor: '#3b82f6',
      description: '转发的数据包'
    },
    style: {
      width: 220,
      height: 140,
      borderRadius: '8px',
      border: '2px solid #3b82f6',
      backgroundColor: '#dbeafe',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // POSTROUTING 链
  {
    id: 'postrouting',
    type: 'chain',
    position: { x: 550, y: 650 },
    data: {
      label: 'POSTROUTING',
      chainName: 'POSTROUTING',
      tables: ['mangle', 'nat'],
      ruleCount: 14,
      color: '#fef3c7',
      borderColor: '#f59e0b',
      description: '路由后处理'
    },
    style: {
      width: 220,
      height: 140,
      borderRadius: '8px',
      border: '2px solid #f59e0b',
      backgroundColor: '#fef3c7',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  },
  
  // 内部网络
  {
    id: 'internal-network',
    type: 'network',
    position: { x: 550, y: 800 },
    data: {
      label: '内部网络',
      icon: 'network',
      nodeType: 'network',
      color: '#ffffff',
      borderColor: '#d1d5db',
      description: '内部网络目的地'
    },
    style: {
      width: 160,
      height: 90,
      borderRadius: '16px',
      border: '2px solid #d1d5db',
      backgroundColor: '#ffffff',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)'
    }
  }
])

// 定义连接边
const edges = ref([
  // 外部网络 → PREROUTING
  {
    id: 'e-external-prerouting',
    source: 'external-network',
    target: 'prerouting',
    type: 'flow',
    data: { 
      label: '入站数据包',
      description: '从外部网络进入的数据包'
    },
    style: { 
      stroke: '#3b82f6', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed,
    animated: true
  },
  
  // PREROUTING → 路由决策
  {
    id: 'e-prerouting-routing',
    source: 'prerouting',
    target: 'routing-decision',
    type: 'flow',
    data: { 
      label: '预路由处理',
      description: '经过PREROUTING链处理后的数据包'
    },
    style: { 
      stroke: '#3b82f6', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // 路由决策 → INPUT (发往本机)
  {
    id: 'e-routing-input',
    source: 'routing-decision',
    target: 'input',
    type: 'flow',
    data: { 
      label: '发往本机',
      description: '目的地为本机的数据包'
    },
    style: { 
      stroke: '#22c55e', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // 路由决策 → FORWARD (转发)
  {
    id: 'e-routing-forward',
    source: 'routing-decision',
    target: 'forward',
    type: 'flow',
    data: { 
      label: '转发处理',
      description: '需要转发的数据包'
    },
    style: { 
      stroke: '#f59e0b', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // INPUT → 本地处理
  {
    id: 'e-input-local',
    source: 'input',
    target: 'local-process',
    type: 'flow',
    data: { 
      label: '输入过滤',
      description: '经过INPUT链过滤后的数据包'
    },
    style: { 
      stroke: '#22c55e', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // 本地处理 → OUTPUT
  {
    id: 'e-local-output',
    source: 'local-process',
    target: 'output',
    type: 'flow',
    data: { 
      label: '本地响应',
      description: '本地应用程序生成的响应数据包'
    },
    style: { 
      stroke: '#a855f7', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // OUTPUT → POSTROUTING
  {
    id: 'e-output-postrouting',
    source: 'output',
    target: 'postrouting',
    type: 'flow',
    data: { 
      label: '出站数据包',
      description: '经过OUTPUT链处理后的数据包'
    },
    style: { 
      stroke: '#22c55e', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // FORWARD → POSTROUTING
  {
    id: 'e-forward-postrouting',
    source: 'forward',
    target: 'postrouting',
    type: 'flow',
    data: { 
      label: '转发数据包',
      description: '经过FORWARD链处理后的数据包'
    },
    style: { 
      stroke: '#f59e0b', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed
  },
  
  // POSTROUTING → 内部网络
  {
    id: 'e-postrouting-internal',
    source: 'postrouting',
    target: 'internal-network',
    type: 'flow',
    data: { 
      label: '路由后处理',
      description: '经过POSTROUTING链处理后的数据包'
    },
    style: { 
      stroke: '#3b82f6', 
      strokeWidth: 2,
      strokeDasharray: '5,5' 
    },
    markerEnd: MarkerType.ArrowClosed,
    animated: true
  }
])

// 布局相关方法
const applyLayout = async (direction: string = layoutDirection.value) => {
  try {
    console.log(`应用布局: ${direction}，节点间距: ${layoutSettings.nodeSpacing}，层级间距: ${layoutSettings.rankSpacing}`)
    
    // 更新布局方向
    layoutDirection.value = direction
    
    // 获取布局后的节点
    const layoutedNodes = layout(nodes.value, edges.value, direction)
    
    if (layoutSettings.animateLayout) {
      // 使用更平滑的动画过渡
      const transitionDuration = 800 // 增加总过渡时间（毫秒）
      const staggerDelay = 10 // 减少节点之间的延迟（毫秒）
      
      // 为每个节点应用动画
      layoutedNodes.forEach((layoutedNode: any, index: number) => {
        setTimeout(() => {
          const nodeIndex = nodes.value.findIndex((n: any) => n.id === layoutedNode.id)
          if (nodeIndex !== -1) {
            // 创建节点的副本并更新位置和连接点
            const updatedNode = { 
              ...nodes.value[nodeIndex],
              position: layoutedNode.position,
              sourcePosition: layoutedNode.sourcePosition,
              targetPosition: layoutedNode.targetPosition
            }
            
            // 更新节点
            nodes.value[nodeIndex] = updatedNode
          }
        }, index * staggerDelay)
      })

      // 等待所有动画完成后再调整视图
      setTimeout(() => fitView({ padding: 0.2, duration: 800 }),
        layoutedNodes.length * staggerDelay + transitionDuration)
    } else {
      // 不使用动画，直接更新节点位置和连接点
      nodes.value = layoutedNodes.map((layoutedNode: any) => {
        // 确保保留原始节点的所有属性，只更新位置和连接点
        const originalNode = nodes.value.find((n: any) => n.id === layoutedNode.id)
        if (originalNode) {
          return {
            ...originalNode,
            position: layoutedNode.position,
            sourcePosition: layoutedNode.sourcePosition,
            targetPosition: layoutedNode.targetPosition
          }
        }
        return layoutedNode
      })
      
      // 立即调整视图
      nextTick(() => fitView({ padding: 0.2 }))
    }
  } catch (error) {
    console.error('布局应用失败:', error)
  }
}

const onNodesInitialized = () => {
  // 保存初始位置
  nodes.value.forEach((node: any) => {
    initialNodePositions.set(node.id, { ...node.position })
  })
  
  // 应用初始布局
  applyLayout('TB')
}

const onLayoutChange = (direction: string) => {
  console.log(`布局方向改变: ${direction}`)
  
  // 确保布局方向有效
  if (!['TB', 'BT', 'LR', 'RL'].includes(direction)) {
    console.warn(`无效的布局方向: ${direction}，使用默认值: TB`)
    direction = 'TB'
  }
  
  // 应用新布局
  applyLayout(direction)
}

const onFitView = () => {
  fitView({ padding: 0.1, duration: 800 })
}

const onResetLayout = () => {
  console.log(`重置到初始位置`)
  // 重置到初始位置
  nodes.value.forEach((node: any) => {
    const initialPos = initialNodePositions.get(node.id)
    if (initialPos) {
      node.position = { ...initialPos }
    }
  })
  
  nextTick(() => {
    applyLayout(layoutDirection.value)
  })
}

const onSpacingChange = (type: 'node' | 'rank', value: number) => {
  if (type === 'node') {
    layoutSettings.nodeSpacing = value
  } else {
    layoutSettings.rankSpacing = value
  }
  
  // 重新应用布局
  applyLayout(layoutDirection.value)
}

const onAnimateToggle = (value: boolean) => {
  layoutSettings.animateLayout = value
}

// 事件处理
const onNodeClick = (event: any) => {
  emit('node-click', event)
}

const onEdgeClick = (event: any) => {
  emit('edge-click', event)
}

const onSelectChainTable = (tableName: string, chainName: string) => {
  emit('select-chain-table', tableName, chainName)
}
</script>

<style scoped>
.dataflow-view {
  height: 800px;
  position: relative;
  background: #f8fafc;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.vue-flow-wrapper {
  height: 100%;
  width: 100%;
}

.dataflow-diagram {
  background: #f8fafc;
}

.layout-control-panel {
  z-index: 1000;
}

/* 节点动画过渡 */
:deep(.vue-flow__node) {
  transition: transform 0.3s ease-in-out, box-shadow 0.3s ease-in-out;
}

:deep(.vue-flow__node:hover) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
  z-index: 10;
}

/* 边动画效果 */
:deep(.vue-flow__edge-path) {
  transition: stroke-width 0.2s ease, stroke-dasharray 0.2s ease;
}

:deep(.vue-flow__edge:hover .vue-flow__edge-path) {
  stroke-width: 3;
}

/* 选中节点效果 */
:deep(.vue-flow__node.selected) {
  box-shadow: 0 0 0 2px #3b82f6;
}

/* 选中边效果 */
:deep(.vue-flow__edge.selected .vue-flow__edge-path) {
  stroke-width: 3;
  stroke-dasharray: none !important;
}

/* 改进节点样式 */
:deep(.vue-flow__node[data-type="chain"]) {
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

:deep(.vue-flow__node[data-type="decision"]) {
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

:deep(.vue-flow__node[data-type="network"]) {
  border-radius: 16px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

:deep(.vue-flow__node[data-type="process"]) {
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* 改进控制面板样式 */
:deep(.vue-flow__controls) {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

:deep(.vue-flow__controls-button) {
  border-radius: 4px;
  transition: background-color 0.2s ease;
}

/* 改进小地图样式 */
:deep(.vue-flow__minimap) {
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dataflow-view {
    height: 600px;
  }
  
  .layout-control-panel {
    position: fixed !important;
    top: 10px !important;
    right: 10px !important;
  }
}

@media (max-width: 480px) {
  .layout-control-panel {
    transform: scale(0.9);
    transform-origin: top right;
  }
}
</style>