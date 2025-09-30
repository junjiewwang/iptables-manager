<template>
  <div class="layout-panel">
    <div class="panel-header">
      <h4>布局控制</h4>
    </div>
    
    <div class="layout-buttons">
      <button 
        title="垂直布局" 
        :class="{ active: currentDirection === 'TB' }"
        @click="$emit('layout-change', 'TB')"
      >
        <el-icon><ArrowDown /></el-icon>
        <span>垂直</span>
      </button>

      <button 
        title="水平布局" 
        :class="{ active: currentDirection === 'LR' }"
        @click="$emit('layout-change', 'LR')"
      >
        <el-icon><ArrowRight /></el-icon>
        <span>水平</span>
      </button>

      <button 
        title="自动适配" 
        @click="$emit('fit-view')"
      >
        <el-icon><FullScreen /></el-icon>
        <span>适配</span>
      </button>

      <button 
        title="重置布局" 
        @click="$emit('reset-layout')"
      >
        <el-icon><RefreshRight /></el-icon>
        <span>重置</span>
      </button>
    </div>

    <div class="layout-options">
      <div class="option-item">
        <label>节点间距</label>
        <el-slider
          :model-value="nodeSpacing"
          :min="50"
          :max="200"
          :step="10"
          size="small"
          @update:model-value="$emit('spacing-change', 'node', $event)"
        />
      </div>

      <div class="option-item">
        <label>层级间距</label>
        <el-slider
          :model-value="rankSpacing"
          :min="80"
          :max="300"
          :step="20"
          size="small"
          @update:model-value="$emit('spacing-change', 'rank', $event)"
        />
      </div>

      <div class="option-item">
        <el-checkbox 
          :model-value="animateLayout"
          @update:model-value="$emit('animate-toggle', $event)"
        >
          动画过渡
        </el-checkbox>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowDown, ArrowRight, FullScreen, RefreshRight } from '@element-plus/icons-vue'

interface Props {
  currentDirection: string
  nodeSpacing: number
  rankSpacing: number
  animateLayout: boolean
}

interface Emits {
  (e: 'layout-change', direction: string): void
  (e: 'fit-view'): void
  (e: 'reset-layout'): void
  (e: 'spacing-change', type: 'node' | 'rank', value: number): void
  (e: 'animate-toggle', value: boolean): void
}

defineProps<Props>()
defineEmits<Emits>()
</script>

<style scoped>
.layout-panel {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  min-width: 200px;
}

.panel-header {
  margin-bottom: 12px;
}

.panel-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.layout-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 16px;
}

.layout-buttons button {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #ffffff;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 12px;
}

.layout-buttons button:hover {
  background: #f3f4f6;
  border-color: #3b82f6;
  color: #3b82f6;
}

.layout-buttons button.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
}

.layout-options {
  border-top: 1px solid #e5e7eb;
  padding-top: 12px;
}

.option-item {
  margin-bottom: 12px;
}

.option-item:last-child {
  margin-bottom: 0;
}

.option-item label {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 6px;
}

:deep(.el-slider) {
  margin: 0;
}

:deep(.el-slider__runway) {
  height: 4px;
}

:deep(.el-slider__button) {
  width: 14px;
  height: 14px;
}

:deep(.el-checkbox) {
  font-size: 12px;
}

:deep(.el-checkbox__label) {
  color: #6b7280;
}
</style>