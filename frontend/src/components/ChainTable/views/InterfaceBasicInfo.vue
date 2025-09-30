<template>
  <div class="interface-basic-info">
    <div class="info-row">
      <span class="info-label">类型:</span>
      <span class="info-value">{{ interfaceData.type }}</span>
    </div>
    <div class="info-row">
      <span class="info-label">状态:</span>
      <span class="info-value">{{ interfaceData.state }}</span>
    </div>
    <div class="info-row">
      <span class="info-label">MTU:</span>
      <span class="info-value">{{ interfaceData.mtu }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  interfaceData: any
}

const props = defineProps<Props>()
</script>

<style scoped>
.interface-basic-info {
  margin-bottom: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 6px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-weight: 500;
  color: #666;
}

.info-value {
  color: #333;
}
</style>