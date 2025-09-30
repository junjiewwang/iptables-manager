<template>
  <div 
    class="chain-process-node"
    :style="nodeStyle"
    @click="handleClick"
  >
    <div class="node-header">
      <div class="node-title">{{ nodeProps.data.label }}</div>
      <div class="rule-count">{{ nodeProps.data.ruleCount }} 条规则</div>
    </div>
    
    <div class="tables-section">
      <div class="tables-label">关联表:</div>
      <div class="tables-list">
        <span 
          v-for="table in nodeProps.data.tables" 
          :key="table"
          class="table-tag"
          :class="`table-${table}`"
          @click.stop="handleTableClick(table)"
        >
          {{ table }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  nodeProps: {
    id: string
    data: {
      label: string
      chainName: string
      tables: string[]
      ruleCount: number
      color: string
      borderColor: string
    }
  }
}

interface Emits {
  (e: 'select-chain-table', tableName: string, chainName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const nodeStyle = computed(() => ({
  backgroundColor: props.nodeProps.data.color,
  borderColor: props.nodeProps.data.borderColor,
  borderWidth: '2px',
  borderStyle: 'solid'
}))

const handleClick = () => {
  console.log('Chain node clicked:', props.nodeProps.id)
}

const handleTableClick = (tableName: string) => {
  emit('select-chain-table', tableName, props.nodeProps.data.chainName)
}
</script>

<style scoped>
.chain-process-node {
  min-width: 160px;
  min-height: 100px;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.chain-process-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.node-header {
  margin-bottom: 12px;
}

.node-title {
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 4px;
}

.rule-count {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.tables-section {
  margin-top: 8px;
}

.tables-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 6px;
  font-weight: 500;
}

.tables-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.table-tag {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.table-tag:hover {
  transform: scale(1.05);
}

.table-raw {
  background-color: #fee2e2;
  color: #dc2626;
}

.table-mangle {
  background-color: #fef3c7;
  color: #d97706;
}

.table-nat {
  background-color: #dbeafe;
  color: #2563eb;
}

.table-filter {
  background-color: #dcfce7;
  color: #16a34a;
}
</style>