<template>
  <g>
    <!-- 边线 -->
    <path
      :id="id"
      :style="edgeStyle"
      class="vue-flow__edge-path"
      :d="edgePath"
      :marker-end="markerEnd"
    />
    
    <!-- 边标签 -->
    <text
      v-if="data?.label"
      class="edge-label"
      :x="labelX"
      :y="labelY"
      text-anchor="middle"
      dominant-baseline="middle"
    >
      <tspan
        class="edge-label-bg"
        :x="labelX"
        :y="labelY"
      >
        {{ data.label }}
      </tspan>
      <tspan
        class="edge-label-text"
        :x="labelX"
        :y="labelY"
      >
        {{ data.label }}
      </tspan>
    </text>
  </g>
</template>

<script setup lang="ts">
import { getBezierPath, Position } from '@vue-flow/core'

interface Props {
  id: string
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  sourcePosition: string
  targetPosition: string
  data?: {
    label?: string
    description?: string
    sourcePosition?: string
    targetPosition?: string
  }
  markerEnd?: string
  style?: Record<string, any>
}

const props = defineProps<Props>()

/**
 * 将字符串连接点转换为Position枚举值
 */
const stringToPosition = (positionStr: string) => {
  switch (positionStr) {
    case 'top': return Position.Top
    case 'bottom': return Position.Bottom
    case 'left': return Position.Left
    case 'right': return Position.Right
    default: return Position.Bottom // 默认值
  }
}

/**
 * 获取实际的连接点信息（字符串格式）
 * 优先使用props中的连接点信息，如果没有则使用data中的
 */
const actualSourcePositionStr = computed(() => {
  // 以 props 为准，data 作为回退，确保与 VueFlow 根据节点把手计算出的坐标一致
  return props.sourcePosition || props.data?.sourcePosition
})

const actualTargetPositionStr = computed(() => {
  return props.targetPosition || props.data?.targetPosition
})

/**
 * 获取实际的连接点信息（Position枚举格式）
 * 用于传递给getBezierPath函数
 */
const actualSourcePosition = computed(() => {
  return stringToPosition(actualSourcePositionStr.value)
})

const actualTargetPosition = computed(() => {
  return stringToPosition(actualTargetPositionStr.value)
})

/**
 * 判断是否为水平布局（左→右或右→左）
 */
const isHorizontalLayout = computed(() => {
  return (
    actualSourcePositionStr.value === 'right' && actualTargetPositionStr.value === 'left'
  )
})

const isReverseHorizontalLayout = computed(() => {
  return (
    actualSourcePositionStr.value === 'left' && actualTargetPositionStr.value === 'right'
  )
})

/**
 * 计算贝塞尔曲线路径
 * 仅支持水平布局，简化日志与曲率计算
 */
const edgePath = computed(() => {
  // 调试：打印边的连接点信息
  console.log(`🔗 [FlowEdge] 边 ${props.id} 连接点信息:`)
  console.log(`  Props连接点: source=${props.sourcePosition}, target=${props.targetPosition}`)
  console.log(`  Data连接点: source=${props.data?.sourcePosition}, target=${props.data?.targetPosition}`)
  const sourceMismatch = props.sourcePosition && props.data?.sourcePosition && props.sourcePosition !== props.data.sourcePosition
  const targetMismatch = props.targetPosition && props.data?.targetPosition && props.targetPosition !== props.data.targetPosition
  if (sourceMismatch || targetMismatch) {
    console.warn(`  ⚠️ 连接点不一致: props(source=${props.sourcePosition}, target=${props.targetPosition}) ≠ data(source=${props.data?.sourcePosition}, target=${props.data?.targetPosition})，已按 props 优先处理`)
  }
  console.log(`  实际使用连接点(字符串): source=${actualSourcePositionStr.value}, target=${actualTargetPositionStr.value} (以props为准)`)
  console.log(`  实际使用连接点(枚举): source=${actualSourcePosition.value}, target=${actualTargetPosition.value}`)
  console.log(`  源节点: (${props.sourceX}, ${props.sourceY}) - ${actualSourcePositionStr.value}`)
  console.log(`  目标节点: (${props.targetX}, ${props.targetY}) - ${actualTargetPositionStr.value}`)
  console.log(`  布局判断: 水平=${isHorizontalLayout.value}, 反向水平=${isReverseHorizontalLayout.value}`)
  
  if (isHorizontalLayout.value) {
    console.log(`  ✅ [FlowEdge] 水平布局 (right → left)，使用曲率: ${getCurvature()}`)
  } else if (isReverseHorizontalLayout.value) {
    console.log(`  ✅ [FlowEdge] 反向水平布局 (left → right)，使用曲率: ${getCurvature()}`)
  } else {
    console.log(`  ⚠️ [FlowEdge] 未匹配到水平布局，仍使用默认曲率: ${getCurvature()}`)
  }
  
  // 仅按水平布局计算贝塞尔曲线
  const [path] = getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: actualSourcePosition.value,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: actualTargetPosition.value,
    curvature: getCurvature(),
  })
  
  console.log(`  📍 [FlowEdge] getBezierPath 参数:`)
  console.log(`    sourceX: ${props.sourceX}, sourceY: ${props.sourceY}`)
  console.log(`    targetX: ${props.targetX}, targetY: ${props.targetY}`)
  console.log(`    sourcePosition: ${actualSourcePosition.value} (${actualSourcePositionStr.value}), targetPosition: ${actualTargetPosition.value} (${actualTargetPositionStr.value})`)
  console.log(`    curvature: ${getCurvature()}`)
  console.log(`  🎨 [FlowEdge] 生成路径: ${path.substring(0, 80)}...`)
  
  return path
})

/**
 * 固定曲率（只保留水平布局）
 */
const getCurvature = () => {
  return 0.15
}

/**
 * 计算边的样式
 */
const edgeStyle = computed(() => ({
  strokeWidth: 2,
  ...props.style
}))

/**
 * 计算标签X坐标位置：水平布局居中
 */
const labelX = computed(() => {
  const midX = (props.sourceX + props.targetX) / 2
  return midX
})

/**
 * 计算标签Y坐标位置：根据左右方向微调
 */
const labelY = computed(() => {
  const midY = (props.sourceY + props.targetY) / 2
  if (isHorizontalLayout.value) {
    return midY - 15
  }
  if (isReverseHorizontalLayout.value) {
    return midY + 20
  }
  return midY - 10
})
</script>

<style scoped>
.vue-flow__edge-path {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: stroke-width 0.2s, stroke-dasharray 0.2s;
}

.edge-label {
  font-size: 12px;
  font-weight: 500;
  pointer-events: none;
  transition: transform 0.3s ease;
}

.edge-label-bg {
  fill: white;
  stroke: white;
  stroke-width: 4;
  opacity: 0.9;
}

.edge-label-text {
  fill: #374151;
}

/* 悬停效果 */
.vue-flow__edge:hover .vue-flow__edge-path {
  stroke-width: 3;
}

/* 选中效果 */
.vue-flow__edge.selected .vue-flow__edge-path {
  stroke-width: 3;
  stroke-dasharray: none !important;
}
</style>