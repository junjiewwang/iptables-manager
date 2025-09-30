<template>
  <el-card>
    <!-- 主要控制区 -->
    <div class="main-controls">
      <div class="view-tabs">
        <el-tabs :value="viewMode"
                 @input="$emit('update:viewMode', $event)" @tab-change="handleViewModeChange" type="card">
          <el-tab-pane label="链视图" name="chain">
            <template #label>
              <el-icon><Share /></el-icon>
              链视图
            </template>
          </el-tab-pane>
          <el-tab-pane label="表视图" name="table">
            <template #label>
              <el-icon><Grid /></el-icon>
              表视图
            </template>
          </el-tab-pane>
          <el-tab-pane label="接口视图" name="interface">
            <template #label>
              <el-icon><Connection /></el-icon>
              接口视图
            </template>
          </el-tab-pane>
        </el-tabs>
      </div>
      
      <div class="action-buttons">
        <el-button @click="refreshData" :loading="loading" type="primary">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
        <el-button @click="cleanInvalidRules" :loading="cleaningRules" type="warning">
          <el-icon><Delete /></el-icon>
          清除无效规则
        </el-button>
        <el-button v-if="viewMode === 'chain'" @click="showTopoSettings" type="success">
          <el-icon><Setting /></el-icon>
          拓扑设置
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { Share, Grid, Connection, Refresh, Delete, Setting } from '@element-plus/icons-vue'
import type { ViewMode } from './types'

// 定义组件属性
const props = defineProps({
  viewMode: {
    type: String as () => ViewMode,
    default: 'chain'
  },
  loading: {
    type: Boolean,
    default: false
  },
  cleaningRules: {
    type: Boolean,
    default: false
  }
})

// 定义事件
const emit = defineEmits([
  'update:viewMode',
  'refresh-data',
  'clean-invalid-rules',
  'show-topo-settings'
])

// 处理视图模式变化
const handleViewModeChange = (tab: string) => {
  emit('update:viewMode', tab)
}

// 刷新数据
const refreshData = () => {
  emit('refresh-data')
}

// 清除无效规则
const cleanInvalidRules = () => {
  emit('clean-invalid-rules')
}

// 显示拓扑设置
const showTopoSettings = () => {
  emit('show-topo-settings')
}
</script>

<style scoped>
.main-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.view-tabs {
  flex: 1;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

@media (max-width: 768px) {
  .main-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
}
</style>