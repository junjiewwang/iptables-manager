<template>
  <div class="chain-table-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>IPTables 五链四表可视化</h1>
      <p class="description">展示PREROUTING、INPUT、FORWARD、OUTPUT、POSTROUTING五链与raw、mangle、nat、filter四表的关系</p>
    </div>

    <!-- 控制面板 -->
    <div class="control-panel">
      <TableToolbar
          v-model:viewMode="viewMode"
          :loading="loading"
          :cleaningRules="cleaningRules"
          @refresh-data="refreshData"
          @clean-invalid-rules="cleanInvalidRules"

      />

      <FilterPanel
          v-model:activeFilterPanels="activeFilterPanels"
          v-model:selectedInterfaces="selectedInterfaces"
          v-model:selectedProtocols="selectedProtocols"
          v-model:selectedTargets="selectedTargets"
          v-model:ipRangeFilter="ipRangeFilter"
          v-model:portRangeFilter="portRangeFilter"
          :interfaces="interfaces"
          :activeFiltersCount="activeFiltersCount"
          @apply-filters="applyFilters"
          @clear-filters="clearAllFilters"
      />
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated/>
    </div>

    <!-- 主要内容区域 -->
    <MainTable
        v-else
        :viewMode="viewMode"
        :chains="chains"
        :tables="tables"
        :interfaces="interfaces"
        :filteredInterfaceData="filteredInterfaceData"
        :activeInterfacesCount="activeInterfacesCount"
        :dockerInterfacesCount="dockerInterfacesCount"
        :totalInterfaceRules="totalInterfaceRules"
        v-model:showChainDialog="showChainDialog"
        :detailTitle="detailTitle"
        :detailRules="detailRules"
        :selectedChain="selectedChain"
        v-model:groupByChain="groupByChain"
        v-model:ruleSearchText="ruleSearchText"
        v-model:tableFilter="tableFilter"
        v-model:targetFilter="targetFilter"
        :groupedRules="groupedRules"
        :filteredDetailRules="filteredDetailRules"
        @select-chain-table="selectChainInTable"
        @close-chain-dialog="closeChainDialog"
        @edit-rule="editRuleFromDetail"
        @delete-rule="deleteRuleFromDetail"
        @add-rule="openAddRuleDialog"
        @view-interface-rules="viewInterfaceRules"
    />


    <!-- 添加/编辑规则对话框 -->
    <EditFormDialog
        v-model:visible="showAddRuleDialog"
        :isEditRule="isEditRule"
        :ruleForm="ruleForm"
        :ruleFormRules="ruleFormRules"
        :interfaces="interfaces"
        ref="ruleFormRef"
        @submit="submitRuleForm"
        @chain-change="handleChainChange"
        @reset="resetRuleForm"
    />
  </div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, ref, watch} from 'vue'
import {ElMessage} from 'element-plus'
import TableToolbar from './TableToolbar.vue'
import FilterPanel from './FilterPanel.vue'
import MainTable from './views/MainTable.vue'
import EditFormDialog from './dialogs/EditFormDialog.vue'
import {useChainTable} from '@/composables/ChainTable/useChainTable'
import {useTableFilters} from '@/composables/ChainTable/useTableFilters'
import {useTableActions} from '@/composables/ChainTable/useTableActions'

// 使用组合式函数
const {
  loading,
  viewMode,
  selectedChain,
  showChainDialog,
  detailTitle,
  detailRules,
  groupByChain,
  chainTableData,
  interfaces,

  chains,
  tables,
  sortedDetailRules,
  groupedRules,
  refreshData,
  selectChain,
  closeChainDialog,
  getInterfaceRuleCount,
  viewInterfaceRules
} = useChainTable()

const {
  activeFilterPanels,
  selectedInterfaces,
  selectedProtocols,
  selectedTargets,
  ipRangeFilter,
  portRangeFilter,
  selectedInterfaceTypes,
  interfaceStatusFilter,
  tableFilter,
  targetFilter,
  ruleSearchText,
  activeFiltersCount,
  hasActiveFilters,
  applyFiltersToRules,
  applyFiltersToInterfaces,
  clearAllFilters
} = useTableFilters()

const {
  rulesLoading,
  ruleDialogVisible,
  isEditRule,
  rules,
  ruleFormRef,
  ruleForm,
  ruleFormRules,
  cleaningRules,
  showCleanupDialog,
  cleanupResult,
  loadRules,
  openAddRuleDialog,
  editRule,
  deleteRule,
  submitRuleForm,
  resetRuleForm,
  handleChainChange,
  cleanInvalidRules
} = useTableActions()


// 添加规则对话框
const showAddRuleDialog = ref(false)

// 筛选后的表格规则
const filteredTableRules = computed(() => {
  // 收集所有表中的所有规则
  let allRules: any[] = []

  // 收集所有表中的所有规则
  tables.value.forEach((table: any) => {
    if (table.chains && Array.isArray(table.chains)) {
      table.chains.forEach((chain: any) => {
        if (chain.rules && Array.isArray(chain.rules)) {
          chain.rules.forEach((rule: any) => {
            allRules.push({
              ...rule,
              table: table.name,
              chain_name: chain.name,
              line_number: rule.line_number || allRules.length + 1
            })
          })
        }
      })
    }
  })

  // 应用筛选条件
  return applyFiltersToRules(allRules)
})

// 筛选后的详细规则
const filteredDetailRules = computed(() => {
  return applyFiltersToRules(sortedDetailRules.value)
})

// 筛选后的接口数据
const filteredInterfaceData = computed(() => {
  let filtered = applyFiltersToInterfaces(interfaces.value)

  // 添加规则计数
  return filtered.map((iface: any) => ({
    ...iface,
    inRules: getInterfaceRuleCount(iface.name, 'in', filteredTableRules.value),
    outRules: getInterfaceRuleCount(iface.name, 'out', filteredTableRules.value),
    forwardRules: getInterfaceRuleCount(iface.name, 'forward', filteredTableRules.value)
  }))
})

// 活跃接口数量
const activeInterfacesCount = computed(() => {
  return filteredInterfaceData.value.filter((iface: any) => iface.is_up).length
})

// Docker接口数量
const dockerInterfacesCount = computed(() => {
  return filteredInterfaceData.value.filter((iface: any) => iface.is_docker).length
})

// 总接口规则数量
const totalInterfaceRules = computed(() => {
  return filteredInterfaceData.value.reduce((total: number, iface: any) => {
    return total + iface.inRules + iface.outRules + iface.forwardRules
  }, 0)
})

// 应用筛选
const applyFilters = () => {
  ElMessage.success(`已应用 ${activeFiltersCount.value} 个筛选条件`)
}

// 从详细规则页面编辑规则
const editRuleFromDetail = (rule: any) => {
  // 将详细规则数据转换为规则表单数据
  isEditRule.value = true
  showAddRuleDialog.value = true
  Object.assign(ruleForm.value, {
    id: rule.id,
    chain_name: rule.chain_name || '',
    table: rule.table || '',
    target: rule.target || '',
    protocol: rule.protocol || '',
    source_ip: rule.source || '',
    destination_ip: rule.destination || '',
    destination_port: rule.destination_port || ''
  })
}

// 从详细规则页面删除规则
const deleteRuleFromDetail = async (rule: any) => {
  const success = await deleteRule(rule)
  if (success) {
    await refreshData()
    // 重新加载详细规则数据
    if (selectedChain.value) {
      selectChain(selectedChain.value, filteredTableRules.value)
    }
  }
}

// 选择链表（用于流程图组件）
const selectChainInTable = (tableName: string, chainName: string) => {
  // 构建链标识符
  const chainKey = `${tableName}.${chainName}`
  selectChain(chainKey, filteredTableRules.value)
}


// 初始化
onMounted(async () => {
  // 加载数据
  await refreshData()

  // 确保数据加载完成后再初始化流程图
  nextTick(() => {

  })
})

// 监听筛选条件变化，更新拓扑图
watch([selectedInterfaces, selectedProtocols, selectedTargets, ipRangeFilter, portRangeFilter], () => {
  if (chainTableData.value && chainTableData.value.chains) {
    nextTick(() => {

    })
  }
}, {deep: true})
</script>

<style scoped>
.chain-table-view {
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
  color: white;
}

.page-header h1 {
  font-size: 2.5rem;
  margin-bottom: 10px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.description {
  font-size: 1.1rem;
  opacity: 0.9;
}

.control-panel {
  margin-bottom: 20px;
}

.loading-container {
  padding: 40px;
  background: white;
  border-radius: 12px;
  margin-top: 20px;
}

.settings-options {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .chain-table-view {
    padding: 10px;
  }

  .page-header h1 {
    font-size: 2rem;
  }

  .settings-options {
    flex-direction: column;
    gap: 10px;
  }
}
</style>