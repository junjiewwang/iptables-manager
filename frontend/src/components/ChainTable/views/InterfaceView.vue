<template>
  <div class="interface-view">
    <InterfaceStatsPanel 
      :interfaces-count="interfaces.length"
      :active-interfaces-count="activeInterfacesCount"
      :docker-interfaces-count="dockerInterfacesCount"
      :total-interface-rules="totalInterfaceRules"
    />
    
    <div class="interfaces-grid">
      <InterfaceCard
        v-for="iface in filteredInterfaceData" 
        :key="iface.name"
        :interface-data="iface"
        @view-interface-rules="onViewInterfaceRules"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import InterfaceStatsPanel from './InterfaceStatsPanel.vue'
import InterfaceCard from './InterfaceCard.vue'
import type { NetworkInterface } from '../types'

interface Props {
  interfaces: NetworkInterface[]
  filteredInterfaceData: any[]
  activeInterfacesCount: number
  dockerInterfacesCount: number
  totalInterfaceRules: number
}

interface Emits {
  (e: 'view-interface-rules', interfaceName: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const onViewInterfaceRules = (interfaceName: string) => {
  emit('view-interface-rules', interfaceName)
}
</script>

<style scoped>
.interface-view {
  padding: 20px;
}

.interfaces-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(450px, 1fr));
  gap: 20px;
}

@media (max-width: 768px) {
  .interfaces-grid {
    grid-template-columns: 1fr;
  }
}
</style>