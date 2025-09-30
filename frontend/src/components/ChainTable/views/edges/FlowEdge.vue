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
import { computed } from 'vue'
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
  }
  markerEnd?: string
  style?: Record<string, any>
}

const props = defineProps<Props>()

/**
 * 判断是否为水平布局
 * 通过检查源节点和目标节点的连接点位置来确定
 */
const isHorizontalLayout = computed(() => {
  return (
    (props.sourcePosition === Position.Right && props.targetPosition === Position.Left) ||
    (props.sourcePosition === Position.Left && props.targetPosition === Position.Right)
  )
})

/**
 * 判断是否为垂直布局
 * 通过检查源节点和目标节点的连接点位置来确定
 */
const isVerticalLayout = computed(() => {
  return (
    (props.sourcePosition === Position.Bottom && props.targetPosition === Position.Top) ||
    (props.sourcePosition === Position.Top && props.targetPosition === Position.Bottom)
  )
})

/**
 * 计算控制点偏移量，使曲线更平滑
 * 根据布局方向和节点间距离动态调整
 */
const getControlPointOffset = () => {
  // 计算源点和目标点之间的距离
  const dx = props.targetX - props.sourceX
  const dy = props.targetY - props.sourceY
  const distance = Math.sqrt(dx * dx + dy * dy)
  
  // 根据布局方向和距离调整控制点偏移量
  if (isHorizontalLayout.value) {
    // 水平布局时，使用较小的偏移量，避免曲线过于弯曲
    return Math.min(distance * 0.3, 100)
  } else if (isVerticalLayout.value) {
    // 垂直布局时，使用较大的偏移量，使曲线更明显
    return Math.min(distance * 0.5, 150)
  } else {
    // 对角线布局时，使用中等偏移量
    return Math.min(distance * 0.4, 120)
  }
}

/**
 * 计算贝塞尔曲线路径
 * 根据布局方向动态调整曲线参数
 */
const edgePath = computed(() => {
  // 获取控制点偏移量
  const offset = getControlPointOffset()
  
  // 根据布局方向调整贝塞尔曲线参数
  const [path] = getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition as Position,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition as Position,
    // 水平布局时使用较小的曲率，垂直布局时使用较大的曲率
    curvature: isHorizontalLayout.value ? 0.15 : 0.3,
    // 根据布局方向调整偏移量
    offset: offset,
  })
  
  return path
})

/**
 * 计算边的样式
 */
const edgeStyle = computed(() => ({
  strokeWidth: 2,
  ...props.style
}))

/**
 * 计算标签X坐标位置
 * 根据布局方向调整
 */
const labelX = computed(() => {
  const midX = (props.sourceX + props.targetX) / 2
  
  // 水平布局时不需要额外偏移X坐标
  if (isHorizontalLayout.value) {
    return midX
  }
  
  // 对角线布局时，根据源点和目标点的相对位置调整X坐标
  const dx = props.targetX - props.sourceX
  return midX + (dx > 0 ? 10 : -10)
})

/**
 * 计算标签Y坐标位置
 * 根据布局方向调整
 */
const labelY = computed(() => {
  const midY = (props.sourceY + props.targetY) / 2
  
  // 根据布局方向调整垂直偏移
  if (isHorizontalLayout.value) {
    // 水平布局时，标签位于边的上方
    return midY - 15
  } else if (isVerticalLayout.value) {
    // 垂直布局时，标签位于边的右侧
    return midY - 5
  } else {
    // 对角线布局时，标签位于边的中心偏上
    return midY - 10
  }
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