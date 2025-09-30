<template>
  <div class="main-content">
    <!-- 数据流图视图 -->
    <ChainFlowView
      v-if="viewMode === 'chain'"
      :flow-elements="flowElements"
      :topo-settings="topoSettings"
      @update:flow-elements="$emit('update:flowElements', $event)"
      @update:topo-settings="$emit('update:topoSettings', $event)"
      @node-click="events.onNodeClick"
      @edge-click="events.onEdgeClick"
      @node-drag-stop="events.onNodeDragStop"
      @select-chain-table="events.onSelectChainTable"
      @settings-changed="events.onTopoSettingsChanged"
      @strategy-applied="events.onTopoStrategyApplied"
      @settings-reset="events.onTopoSettingsReset"
    />
    
    <!-- 表视图 -->
    <TableView
      v-else-if="viewMode === 'table'"
      :tables="tables"
      @edit-rule="events.onEditRule"
      @delete-rule="events.onDeleteRule"
      @add-rule="events.onAddRule"
    />
    
    <!-- 接口视图 -->
    <InterfaceView
      v-else-if="viewMode === 'interface'"
      :interfaces="interfaces"
      :filtered-interface-data="filteredInterfaceData"
      :active-interfaces-count="activeInterfacesCount"
      :docker-interfaces-count="dockerInterfacesCount"
      :total-interface-rules="totalInterfaceRules"
      @view-interface-rules="events.onViewInterfaceRules"
    />
    
    <!-- 链详情对话框 -->
    <ChainDetailDialog
      :show-dialog="showChainDialog"
      :title="detailTitle"
      :selected-chain="selectedChain"
      :group-by-chain="groupByChain"
      :rule-search-text="ruleSearchText"
      :table-filter="tableFilter"
      :target-filter="targetFilter"
      :grouped-rules="groupedRules"
      :filtered-detail-rules="filteredDetailRules"
      @update:show-dialog="$emit('update:showChainDialog', $event)"
      @update:group-by-chain="$emit('update:groupByChain', $event)"
      @update:rule-search-text="$emit('update:ruleSearchText', $event)"
      @update:table-filter="$emit('update:tableFilter', $event)"
      @update:target-filter="$emit('update:targetFilter', $event)"
      @closed="events.onCloseChainDialog"
      @edit-rule="events.onEditRule"
      @delete-rule="events.onDeleteRule"
      @add-rule="events.onAddRule"
    />
  </div>
</template>

<script setup lang="ts">
import ChainFlowView from './ChainFlowView.vue'
import TableView from './TableView.vue'
import InterfaceView from './InterfaceView.vue'
import ChainDetailDialog from '../dialogs/ChainDetailDialog.vue'
import { useMainTableEvents } from '@/composables/useMainTableEvents'
import type { 
  ViewMode, 
  TopoSettings, 
  IPTablesRule, 
  ChainInfo, 
  TableInfo, 
  NetworkInterface 
} from '@/types/ChainTable/types'

// 定义组件属性
interface Props {
  viewMode: ViewMode
  flowElements: any[]
  topoSettings: TopoSettings
  chains: ChainInfo[]
  tables: TableInfo[]
  interfaces: NetworkInterface[]
  filteredInterfaceData: any[]
  activeInterfacesCount: number
  dockerInterfacesCount: number
  totalInterfaceRules: number
  showChainDialog: boolean
  detailTitle: string
  detailRules: IPTablesRule[]
  selectedChain: string
  groupByChain: boolean
  ruleSearchText: string
  tableFilter: string
  targetFilter: string
  groupedRules: Record<string, IPTablesRule[]>
  filteredDetailRules: IPTablesRule[]
}

// 定义事件
interface Emits {
  (e: 'update:flowElements', value: any[]): void
  (e: 'update:topoSettings', value: TopoSettings): void
  (e: 'update:showChainDialog', value: boolean): void
  (e: 'update:groupByChain', value: boolean): void
  (e: 'update:ruleSearchText', value: string): void
  (e: 'update:tableFilter', value: string): void
  (e: 'update:targetFilter', value: string): void
  (e: 'node-click', event: any): void
  (e: 'edge-click', event: any): void
  (e: 'node-drag-stop', event: any): void
  (e: 'select-chain-table', tableName: string, chainName: string): void
  (e: 'close-chain-dialog'): void
  (e: 'edit-rule', rule: IPTablesRule): void
  (e: 'delete-rule', rule: IPTablesRule): void
  (e: 'add-rule', chainName: string, tableName?: string): void
  (e: 'view-interface-rules', interfaceName: string): void
  (e: 'topo-settings-changed', settings: TopoSettings): void
  (e: 'topo-strategy-applied', strategyType: string): void
  (e: 'topo-settings-reset'): void
}

const props = withDefaults(defineProps<Props>(), {
  viewMode: 'chain',
  flowElements: () => [],
  chains: () => [],
  tables: () => [],
  interfaces: () => [],
  filteredInterfaceData: () => [],
  activeInterfacesCount: 0,
  dockerInterfacesCount: 0,
  totalInterfaceRules: 0,
  showChainDialog: false,
  detailTitle: '',
  detailRules: () => [],
  selectedChain: '',
  groupByChain: true,
  ruleSearchText: '',
  tableFilter: '',
  targetFilter: '',
  groupedRules: () => ({}),
  filteredDetailRules: () => []
})

const emit = defineEmits<Emits>()

// 使用事件处理 composable
const events = useMainTableEvents(emit)
</script>

<style scoped>
.main-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.1);
  overflow: hidden;
}
</style>