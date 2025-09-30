<template>
  <div class="interface-rules-stats">
    <div class="stats-header">
      <h4>规则统计</h4>
      <el-button 
        size="small" 
        type="primary" 
        @click="onViewInterfaceRules(interfaceData.name)"
        :disabled="!hasRules"
      >
        查看规则
      </el-button>
    </div>
    
    <div class="rules-grid">
      <div class="rule-stat-item input">
        <div class="rule-stat-icon input">
          <el-icon><ArrowLeft /></el-icon>
        </div>
        <div class="rule-stat-info">
          <div class="rule-stat-number">{{ interfaceData.inRules }}</div>
          <div class="rule-stat-label">入站规则</div>
        </div>
      </div>
      
      <div class="rule-stat-item output">
        <div class="rule-stat-icon output">
          <el-icon><ArrowRight /></el-icon>
        </div>
        <div class="rule-stat-info">
          <div class="rule-stat-number">{{ interfaceData.outRules }}</div>
          <div class="rule-stat-label">出站规则</div>
        </div>
      </div>
      
      <div class="rule-stat-item forward">
        <div class="rule-stat-icon forward">
          <el-icon><Right /></el-icon>
        </div>
        <div class="rule-stat-info">
          <div class="rule-stat-number">{{ interfaceData.forwardRules }}</div>
          <div class="rule-stat-label">转发规则</div>
        </div>
      </div>
      
      <div class="rule-stat-item total">
        <div class="rule-stat-icon total">
          <el-icon><List /></el-icon>
        </div>
        <div class="rule-stat-info">
          <div class="rule-stat-number">{{ totalRules }}</div>
          <div class="rule-stat-label">总规则数</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, ArrowRight, Right, List } from '@element-plus/icons-vue'

interface Props {
  interfaceData: any
}

interface Emits {
  (e: 'view-interface-rules', interfaceName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const totalRules = computed(() => {
  return props.interfaceData.inRules + props.interfaceData.outRules + props.interfaceData.forwardRules
})

const hasRules = computed(() => {
  return totalRules.value > 0
})

const onViewInterfaceRules = (interfaceName: string) => {
  emit('view-interface-rules', interfaceName)
}
</script>

<style scoped>
.interface-rules-stats {
  margin-bottom: 20px;
}

.stats-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stats-header h4 {
  margin: 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.rule-stat-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.rule-stat-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 16px;
}

.rule-stat-icon.input {
  background: #e3f2fd;
  color: #1976d2;
}

.rule-stat-icon.output {
  background: #fff3e0;
  color: #f57c00;
}

.rule-stat-icon.forward {
  background: #e8f5e8;
  color: #388e3c;
}

.rule-stat-icon.total {
  background: #f3e5f5;
  color: #7b1fa2;
}

.rule-stat-item.total {
  grid-column: span 2;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border: 2px solid #dee2e6;
}

.rule-stat-info {
  flex: 1;
}

.rule-stat-number {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  line-height: 1;
}

.rule-stat-label {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
</style>