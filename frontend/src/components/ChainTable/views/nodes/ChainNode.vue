<template>
  <div 
    :class="[
      'chain-node', 
      nodeProps.data.chainType,
    ]"
    @click="nodeProps.events.nodeClick"
  >
    <!-- 动态连接点：根据节点的sourcePosition/targetPosition在相应边渲染Handle -->
    <Handle
      type="source"
      :position="nodeProps.sourcePosition ?? Position.Right"
      :id="`${nodeProps.id}-source`"
    />
    <Handle
      type="target"
      :position="nodeProps.targetPosition ?? Position.Left"
      :id="`${nodeProps.id}-target`"
    />

    <div class="chain-title">{{ nodeProps.data.label }}</div>
    <div class="chain-tables">
      <span 
        v-for="table in nodeProps.data.tables" 
        :key="table"
        :class="['table-tag', table]"
        @click.stop="selectChainTable(table, nodeProps.data.chainName)"
      >
        {{ table }}
      </span>
    </div>
    <div class="chain-stats">
      {{ nodeProps.data.ruleCount }} 条规则
    </div>
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

interface Props {
  nodeProps: any
}

interface Emits {
  (e: 'select-chain-table', tableName: string, chainName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const selectChainTable = (tableName: string, chainName: string) => {
  emit('select-chain-table', tableName, chainName)
}
</script>

<style scoped>
.chain-node {
  min-width: 120px;
  padding: 12px;
  border-radius: 8px;
  border: 2px solid #e1e5e9;
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
}

/* 美观处理：隐藏默认把手外观，但仍保留连接能力 */
:deep(.vue-flow__handle) {
  width: 8px;
  height: 8px;
  opacity: 0;
}

.chain-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.chain-node.dark-mode {
  background: linear-gradient(135deg, #2d3748 0%, #4a5568 100%);
  border-color: #4a5568;
  color: white;
}

.chain-node.gradient {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-color: #667eea;
}

.chain-node.flat {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: none;
}

.chain-node.glass {
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.chain-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 8px;
  text-align: center;
}

.chain-tables {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 8px;
  justify-content: center;
}

.table-tag {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.table-tag:hover {
  transform: scale(1.1);
}

.table-tag.raw {
  background: #e3f2fd;
  color: #1976d2;
}

.table-tag.mangle {
  background: #fff3e0;
  color: #f57c00;
}

.table-tag.nat {
  background: #e8f5e8;
  color: #388e3c;
}

.table-tag.filter {
  background: #ffebee;
  color: #d32f2f;
}

.chain-stats {
  font-size: 11px;
  color: #666;
  text-align: center;
}

.chain-node.dark-mode .chain-stats {
  color: #cbd5e0;
}
</style>