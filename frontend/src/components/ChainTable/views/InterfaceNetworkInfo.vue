<template>
  <div class="interface-network-info">
    <div class="network-section">
      <h4>IP地址</h4>
      <div class="address-list">
        <el-tag 
          v-for="(ip, index) in interfaceData.ip_addresses" 
          :key="index" 
          class="address-tag" 
          size="small"
        >
          {{ ip }}
        </el-tag>
        <span class="no-address" v-if="!interfaceData.ip_addresses?.length">
          无IP地址
        </span>
      </div>
    </div>
    
    <div class="network-section" v-if="interfaceData.mac_address">
      <h4>MAC地址</h4>
      <div class="mac-address">{{ interfaceData.mac_address }}</div>
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
.interface-network-info {
  margin-bottom: 20px;
}

.network-section {
  margin-bottom: 16px;
}

.network-section h4 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 14px;
  font-weight: 600;
}

.address-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.address-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.no-address {
  color: #999;
  font-style: italic;
}

.mac-address {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #f5f5f5;
  padding: 4px 8px;
  border-radius: 4px;
  color: #666;
}
</style>