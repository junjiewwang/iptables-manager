<template>
  <el-card class="interface-card">
    <template #header>
      <div class="interface-title">
        <el-icon class="interface-icon"><Connection /></el-icon>
        <h3>{{ interfaceData.name }}</h3>
        <div class="interface-badges">
          <el-tag size="small" :type="interfaceData.is_up ? 'success' : 'danger'">
            {{ interfaceData.is_up ? 'UP' : 'DOWN' }}
          </el-tag>
          <el-tag size="small" type="warning" v-if="interfaceData.is_docker">
            Docker
          </el-tag>
          <el-tag size="small" type="info">
            {{ interfaceData.type }}
          </el-tag>
        </div>
      </div>
    </template>
    
    <div class="interface-card-content" :class="{ 'docker-interface': interfaceData.is_docker }">
      <InterfaceBasicInfo :interface-data="interfaceData" />
      <InterfaceNetworkInfo :interface-data="interfaceData" />
      <InterfaceRulesStats 
        :interface-data="interfaceData"
        @view-interface-rules="onViewInterfaceRules"
      />
      <InterfaceTrafficStats :interface-data="interfaceData" />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { Connection } from '@element-plus/icons-vue'
import InterfaceBasicInfo from './InterfaceBasicInfo.vue'
import InterfaceNetworkInfo from './InterfaceNetworkInfo.vue'
import InterfaceRulesStats from './InterfaceRulesStats.vue'
import InterfaceTrafficStats from './InterfaceTrafficStats.vue'

interface Props {
  interfaceData: any
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
.interface-card-content {
  height: 100%;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.interface-card-content.docker-interface {
  border-left: 4px solid #ff9800;
}

.interface-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.interface-icon {
  font-size: 20px;
  color: #1976d2;
}

.interface-badges {
  display: flex;
  gap: 8px;
}
</style>