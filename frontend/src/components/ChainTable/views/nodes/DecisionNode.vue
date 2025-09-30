<template>
  <div 
    class="decision-node"
    :style="nodeStyle"
    @click="nodeProps.events.nodeClick"
  >
    <div class="node-icon">
      <i :class="`fa fa-${nodeProps.data.icon || 'code-branch'}`"></i>
    </div>
    <div class="node-title">{{ nodeProps.data.label }}</div>
    <div class="node-description" v-if="nodeProps.data.description">
      {{ nodeProps.data.description }}
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
      icon: string
      description?: string
      color: string
      borderColor: string
    }
  }
}

const props = defineProps<Props>()

const nodeStyle = computed(() => ({
  backgroundColor: props.nodeProps.data.color,
  borderColor: props.nodeProps.data.borderColor,
  borderWidth: '2px',
  borderStyle: 'solid'
}))
</script>

<style scoped>
.decision-node {
  min-width: 160px;
  min-height: 100px;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.decision-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.node-icon {
  font-size: 24px;
  margin-bottom: 8px;
  color: #4b5563;
}

.node-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  text-align: center;
  margin-bottom: 4px;
}

.node-description {
  font-size: 12px;
  color: #6b7280;
  text-align: center;
}
</style>