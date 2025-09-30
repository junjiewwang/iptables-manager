<template>
  <div class="chain-detail-filters">
    <el-input
      :value="ruleSearchText"
      @input="$emit('update:ruleSearchText', $event)"
      placeholder="搜索规则"
      clearable
      size="small"
      style="width: 200px"
    >
      <template #prefix>
        <el-icon><Search /></el-icon>
      </template>
    </el-input>
    
    <el-select
      :value="tableFilter"
      @input="$emit('update:tableFilter', $event)"
      placeholder="表筛选"
      clearable
      size="small"
      style="width: 120px"
    >
      <el-option
        v-for="table in ['raw', 'mangle', 'nat', 'filter']"
        :key="table"
        :label="table"
        :value="table"
      />
    </el-select>
    
    <el-select
      :value="targetFilter"
      @input="$emit('update:targetFilter', $event)"
      placeholder="目标筛选"
      clearable
      size="small"
      style="width: 150px"
    >
      <el-option
        v-for="target in ['ACCEPT', 'DROP', 'REJECT', 'RETURN', 'MASQUERADE', 'SNAT', 'DNAT']"
        :key="target"
        :label="target"
        :value="target"
      />
    </el-select>
  </div>
</template>

<script setup lang="ts">
import { Search } from '@element-plus/icons-vue'

interface Props {
  ruleSearchText: string
  tableFilter: string
  targetFilter: string
}

interface Emits {
  (e: 'update:ruleSearchText', value: string): void
  (e: 'update:tableFilter', value: string): void
  (e: 'update:targetFilter', value: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
</script>

<style scoped>
.chain-detail-filters {
  display: flex;
  gap: 10px;
}

@media (max-width: 768px) {
  .chain-detail-filters {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>