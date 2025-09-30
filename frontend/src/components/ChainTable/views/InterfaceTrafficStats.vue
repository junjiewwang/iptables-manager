<template>
  <div class="interface-traffic-stats">
    <h4>流量统计</h4>
    <div class="traffic-grid">
      <div class="traffic-item">
        <span class="traffic-label">接收字节:</span>
        <span class="traffic-value">{{ formatBytes(interfaceData.statistics.rx_bytes) }}</span>
      </div>
      <div class="traffic-item">
        <span class="traffic-label">发送字节:</span>
        <span class="traffic-value">{{ formatBytes(interfaceData.statistics.tx_bytes) }}</span>
      </div>
      <div class="traffic-item">
        <span class="traffic-label">接收包数:</span>
        <span class="traffic-value">{{ interfaceData.statistics.rx_packets }}</span>
      </div>
      <div class="traffic-item">
        <span class="traffic-label">发送包数:</span>
        <span class="traffic-value">{{ interfaceData.statistics.tx_packets }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFormatters } from '@/composables/core/useFormatters'

interface Props {
  interfaceData: any
}

const props = defineProps<Props>()

// 使用格式化工具
const { formatBytes } = useFormatters()
</script>

<style scoped>
.interface-traffic-stats {
  margin-bottom: 20px;
}

.interface-traffic-stats h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.traffic-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.traffic-item {
  display: flex;
  justify-content: space-between;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.traffic-label {
  font-weight: 500;
  color: #666;
}

.traffic-value {
  color: #333;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
}
</style>