<template>
  <el-collapse :value="activeFilterPanels"
               @input="$emit('update:activeFilterPanels', $event)" class="filter-panel">
    <el-collapse-item title="筛选条件" name="filters">
      <template #title>
        <div class="filter-title">
          <el-icon>
            <Filter/>
          </el-icon>
          <span>筛选条件</span>
          <el-badge :value="activeFiltersCount" :hidden="activeFiltersCount === 0" type="primary"/>
        </div>
      </template>

      <div class="filter-content">
        <!-- 快捷筛选标签 -->
        <div class="quick-filters">
          <div class="filter-group">
            <label class="filter-label">网络接口:</label>
            <div class="filter-tags">
              <el-tag
                  v-for="iface in interfaces"
                  :key="iface.name"
                  :type="selectedInterfaces.includes(iface.name) ? 'primary' : 'info'"
                  :effect="selectedInterfaces.includes(iface.name) ? 'dark' : 'plain'"
                  @click="toggleInterface(iface.name)"
                  class="filter-tag"
              >
                {{ iface.name }}
              </el-tag>
            </div>
          </div>

          <div class="filter-group">
            <label class="filter-label">协议类型:</label>
            <div class="filter-tags">
              <el-tag
                  v-for="protocol in availableProtocols"
                  :key="protocol"
                  :type="selectedProtocols.includes(protocol) ? 'success' : 'info'"
                  :effect="selectedProtocols.includes(protocol) ? 'dark' : 'plain'"
                  @click="toggleProtocol(protocol)"
                  class="filter-tag"
              >
                {{ protocol.toUpperCase() }}
              </el-tag>
            </div>
          </div>

          <div class="filter-group">
            <label class="filter-label">目标动作:</label>
            <div class="filter-tags">
              <el-tag
                  v-for="target in availableTargets"
                  :key="target"
                  :type="getTargetTagType(target)"
                  :effect="selectedTargets.includes(target) ? 'dark' : 'plain'"
                  @click="toggleTarget(target)"
                  class="filter-tag"
              >
                {{ target }}
              </el-tag>
            </div>
          </div>
        </div>

        <!-- 高级筛选 -->
        <div class="advanced-filters">
          <el-row :gutter="16">
            <el-col :span="8">
              <el-input
                  :value="ipRangeFilter"
                  @input="$emit('update:ipRangeFilter', $event)"
                  placeholder="IP地址范围 (如: 192.168.1.0/24)"
                  clearable
                  size="small"
              >
                <template #prefix>
                  <el-icon>
                    <Location/>
                  </el-icon>
                </template>
              </el-input>
            </el-col>
            <el-col :span="8">
              <el-input
                  :value="portRangeFilter"
                  @input="$emit('update:portRangeFilter', $event)"
                  placeholder="端口范围 (如: 80,443,8000-9000)"
                  clearable
                  size="small"
              >
                <template #prefix>
                  <el-icon>
                    <Connection/>
                  </el-icon>
                </template>
              </el-input>
            </el-col>
            <el-col :span="8">
              <div class="filter-actions">
                <el-button size="small" @click="clearAllFilters">
                  <el-icon>
                    <Delete/>
                  </el-icon>
                  清空筛选
                </el-button>
                <el-button size="small" type="primary" @click="applyFilters">
                  <el-icon>
                    <Search/>
                  </el-icon>
                  应用筛选
                </el-button>
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </el-collapse-item>
  </el-collapse>
</template>

<script setup lang="ts">
import {computed as _computed} from 'vue'
import {Connection, Delete, Filter, Location, Search} from '@element-plus/icons-vue'
import type {NetworkInterface} from './types'

// 重命名computed以避免冲突
const computed = _computed

// 定义组件属性
const props = defineProps({
  activeFilterPanels: {
    type: Array as () => string[],
    default: () => ['filters']
  },
  selectedInterfaces: {
    type: Array as () => string[],
    default: () => []
  },
  selectedProtocols: {
    type: Array as () => string[],
    default: () => []
  },
  selectedTargets: {
    type: Array as () => string[],
    default: () => []
  },
  ipRangeFilter: {
    type: String,
    default: ''
  },
  portRangeFilter: {
    type: String,
    default: ''
  },
  interfaces: {
    type: Array as () => NetworkInterface[],
    default: () => []
  },
  activeFiltersCount: {
    type: Number,
    default: 0
  }
})

// 定义事件
const emit = defineEmits([
  'update:activeFilterPanels',
  'update:selectedInterfaces',
  'update:selectedProtocols',
  'update:selectedTargets',
  'update:ipRangeFilter',
  'update:portRangeFilter',
  'apply-filters',
  'clear-filters'
])

// 可用选项
const availableProtocols = ['tcp', 'udp', 'icmp', 'all']
const availableTargets = ['ACCEPT', 'DROP', 'REJECT', 'RETURN', 'MASQUERADE', 'SNAT', 'DNAT']

// 获取目标的标签类型
const getTargetTagType = (target: string): string => {
  const types: Record<string, string> = {
    'ACCEPT': 'success',
    'DROP': 'danger',
    'REJECT': 'warning',
    'RETURN': 'info',
    'MASQUERADE': 'primary',
    'SNAT': 'primary',
    'DNAT': 'primary'
  }
  return types[target] || 'default'
}

// 切换接口筛选
const toggleInterface = (interfaceName: string) => {
  const newInterfaces = [...props.selectedInterfaces]
  const index = newInterfaces.indexOf(interfaceName)
  if (index > -1) {
    newInterfaces.splice(index, 1)
  } else {
    newInterfaces.push(interfaceName)
  }
  emit('update:selectedInterfaces', newInterfaces)
}

// 切换协议筛选
const toggleProtocol = (protocol: string) => {
  const newProtocols = [...props.selectedProtocols]
  const index = newProtocols.indexOf(protocol)
  if (index > -1) {
    newProtocols.splice(index, 1)
  } else {
    newProtocols.push(protocol)
  }
  emit('update:selectedProtocols', newProtocols)
}

// 切换目标筛选
const toggleTarget = (target: string) => {
  const newTargets = [...props.selectedTargets]
  const index = newTargets.indexOf(target)
  if (index > -1) {
    newTargets.splice(index, 1)
  } else {
    newTargets.push(target)
  }
  emit('update:selectedTargets', newTargets)
}

// 清空所有筛选条件
const clearAllFilters = () => {
  emit('update:selectedInterfaces', [])
  emit('update:selectedProtocols', [])
  emit('update:selectedTargets', [])
  emit('update:ipRangeFilter', '')
  emit('update:portRangeFilter', '')
  emit('clear-filters')
}

// 应用筛选
const applyFilters = () => {
  emit('apply-filters')
}
</script>

<style scoped>
.filter-panel {
  margin-top: 16px;
}

.filter-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.filter-content {
  padding: 16px 0;
}

.quick-filters {
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
}

.filter-label {
  min-width: 80px;
  font-weight: 500;
  color: #606266;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-tag {
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-tag:hover {
  transform: scale(1.05);
}

.advanced-filters {
  border-top: 1px solid #e4e7ed;
  padding-top: 16px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .filter-group {
    flex-direction: column;
    align-items: flex-start;
  }

  .filter-label {
    min-width: auto;
  }
}
</style>