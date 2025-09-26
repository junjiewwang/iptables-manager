<template>
  <div class="topology-container">
    <div class="topology-header">
      <h2>IPTables链路拓扑图</h2>
      <div class="topology-controls">
        <!-- 视图模式切换 -->
        <el-radio-group v-model="viewMode" size="small" @change="onViewModeChange">
          <el-radio-button label="flow">数据流视图</el-radio-button>
          <el-radio-button label="chain">链路架构</el-radio-button>
        </el-radio-group>
        
        <el-divider direction="vertical" />
        
        <!-- 数据流选择 -->
        <el-select 
          v-model="selectedFlow" 
          placeholder="选择数据流" 
          size="small" 
          style="width: 150px"
          clearable
          @change="onFlowChange"
        >
          <el-option label="转发流量" value="forward" />
          <el-option label="本地入站" value="input" />
          <el-option label="本地出站" value="output" />
        </el-select>
        
        <el-divider direction="vertical" />
        
        <!-- 过滤控件 -->
        <el-select 
          v-model="protocolFilter" 
          placeholder="协议过滤" 
          size="small" 
          style="width: 120px"
          clearable
          @change="applyFilters"
        >
          <el-option label="全部" value="" />
          <el-option label="TCP" value="tcp" />
          <el-option label="UDP" value="udp" />
          <el-option label="ICMP" value="icmp" />
        </el-select>
        
        <el-input 
          v-model="portFilter" 
          placeholder="端口过滤" 
          size="small" 
          style="width: 100px"
          clearable
          @input="applyFilters"
        />
        
        <el-divider direction="vertical" />
        
        <!-- 控制按钮 -->
        <el-button @click="resetView" size="small" :icon="Refresh">重置视图</el-button>
        <el-button @click="fitView" size="small" :icon="FullScreen">适应画布</el-button>
        <el-button @click="autoOptimizeLayout" size="small" type="primary" :icon="Star">智能优化</el-button>
        <el-button @click="optimizeArrowPositions" size="small" type="success" :icon="Position">优化箭头</el-button>
            <el-button @click="standardizeConnectionPaths" size="small" type="info" :icon="Position">标准化路径</el-button>
            <el-button @click="fixKeyConnections" size="small" type="warning" :icon="Tools">修复关键连接</el-button>
        <el-button @click="detectDenseAreas" size="small" type="warning" :icon="Search">检测密集区域</el-button>
        <el-button 
          @click="manualAdjustMode ? disableManualAdjust() : enableManualAdjust()" 
          size="small" 
          :type="manualAdjustMode ? 'danger' : 'info'"
          :icon="manualAdjustMode ? Close : Edit"
        >
          {{ manualAdjustMode ? '退出调整' : '手动调整' }}
        </el-button>
        <el-button @click="saveLayoutConfiguration" size="small" type="info" :icon="DocumentCopy">保存布局</el-button>
        <el-button @click="exportTopology" size="small" :icon="Download">导出</el-button>
      </div>
    </div>

    <div class="topology-content">
      <div class="topology-sidebar">
        <!-- 图例卡片 -->
        <el-card class="legend-card">
          <template #header>
            <span>图例</span>
          </template>
          <div class="legend-items">
            <!-- 节点类型图例 -->
            <div class="legend-section">
              <h4>节点类型</h4>
              <div class="legend-item">
                <div class="legend-icon interface-icon"></div>
                <span>网络接口</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon chain-icon"></div>
                <span>IPTables链</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon decision-icon"></div>
                <span>路由决策</span>
              </div>
              <div class="legend-item">
                <div class="legend-icon process-icon"></div>
                <span>本地进程</span>
              </div>
            </div>
            
            <!-- 连接类型图例 -->
            <div class="legend-section">
              <h4>数据流类型</h4>
              <div class="legend-item">
                <div class="legend-line forward-flow"></div>
                <span>转发流量</span>
              </div>
              <div class="legend-item">
                <div class="legend-line input-flow"></div>
                <span>本地入站</span>
              </div>
              <div class="legend-item">
                <div class="legend-line output-flow"></div>
                <span>本地出站</span>
              </div>
              <div class="legend-item">
                <div class="legend-line return-flow"></div>
                <span>返回路径</span>
              </div>
            </div>
            
            <!-- 表处理顺序 -->
            <div class="legend-section">
              <h4>表处理顺序</h4>
              <div class="legend-item">
                <el-tag size="small" type="danger">raw</el-tag>
                <span>连接跟踪</span>
              </div>
              <div class="legend-item">
                <el-tag size="small" type="warning">mangle</el-tag>
                <span>数据包修改</span>
              </div>
              <div class="legend-item">
                <el-tag size="small" type="info">nat</el-tag>
                <span>地址转换</span>
              </div>
              <div class="legend-item">
                <el-tag size="small" type="success">filter</el-tag>
                <span>数据包过滤</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 节点信息卡片 -->
        <el-card class="node-info-card" v-if="selectedNodeInfo">
          <template #header>
            <span>节点信息</span>
          </template>
          <div class="node-info-content">
            <el-descriptions :column="1" size="small">
              <el-descriptions-item label="节点ID">{{ selectedNodeInfo.id }}</el-descriptions-item>
              <el-descriptions-item label="节点类型">{{ selectedNodeInfo.type }}</el-descriptions-item>
              <el-descriptions-item label="标签">{{ selectedNodeInfo.data?.label }}</el-descriptions-item>
              <el-descriptions-item v-if="selectedNodeInfo.data?.chainType" label="链类型">
                {{ selectedNodeInfo.data.chainType }}
              </el-descriptions-item>
              <el-descriptions-item v-if="selectedNodeInfo.data?.tables" label="包含表">
                <el-tag 
                  v-for="table in selectedNodeInfo.data.tables" 
                  :key="table" 
                  size="small" 
                  :type="getTableTagType(table)"
                  style="margin-right: 4px;"
                >
                  {{ table }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item v-if="selectedNodeInfo.data?.ruleCount" label="规则数量">
                {{ selectedNodeInfo.data.ruleCount }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </el-card>

        <!-- 流量统计卡片 -->
        <el-card class="stats-card">
          <template #header>
            <span>流量统计</span>
          </template>
          <div class="stats-content">
            <div class="stat-item">
              <span class="stat-label">总节点数:</span>
              <span class="stat-value">{{ flowElements.length }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">链节点:</span>
              <span class="stat-value">{{ chainNodes.length }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">接口节点:</span>
              <span class="stat-value">{{ interfaceNodes.length }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">当前流量:</span>
              <span class="stat-value">{{ selectedFlow || '无' }}</span>
            </div>
          </div>
        </el-card>
      </div>

      <div class="topology-main">
        <div class="vue-flow-container" v-loading="loading" element-loading-text="加载拓扑图数据...">
          <VueFlow
            v-model="flowElements"
            class="iptables-flow"
            :default-viewport="{ zoom: 0.8 }"
            :min-zoom="0.2"
            :max-zoom="3"
            :snap-to-grid="true"
            :snap-grid="[15, 15]"
            :fit-view-on-init="true"
            :nodes-draggable="true"
            :edges-updatable="false"
            :nodes-connectable="false"
            :delete-key-code="null"
            @node-click="onNodeClick"
            @edge-click="onEdgeClick"
            @edge-double-click="onEdgeDoubleClick"
            @edge-context-menu="onEdgeContextMenu"
            @node-drag-stop="onNodeDragStop"
            @node-mouse-enter="onNodeMouseEnter"
            @node-mouse-leave="onNodeMouseLeave"
            @edge-mouse-enter="onEdgeMouseEnter"
            @edge-mouse-leave="onEdgeMouseLeave"
          >
            <!-- SVG渐变定义 -->
            <defs>
              <linearGradient id="forward-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" style="stop-color:#FF5722;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#FF9800;stop-opacity:1" />
              </linearGradient>
              <linearGradient id="input-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" style="stop-color:#4CAF50;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#8BC34A;stop-opacity:1" />
              </linearGradient>
              <linearGradient id="output-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" style="stop-color:#2196F3;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#03A9F4;stop-opacity:1" />
              </linearGradient>
            </defs>
            
            <!-- 背景网格 -->
            <Background pattern-color="#e2e8f0" :gap="20" />
            
            <!-- 控制面板 -->
            <Controls />
            
            <!-- 小地图 -->
            <MiniMap />
            
            <!-- 自定义节点模板 -->
            <template #node-chain="{ data, id }">
              <div class="chain-node" :class="[data.chainType, { highlighted: highlightedElements.has(id) }]">
                <div class="chain-header">
                  <h3 class="chain-title">{{ data.label }}</h3>
                </div>
                <div class="chain-tables">
                  <span 
                    v-for="table in data.tables" 
                    :key="table"
                    class="table-tag"
                    :class="table"
                  >
                    {{ table }}
                  </span>
                </div>
                <div class="chain-stats" v-if="data.ruleCount">
                  <i class="el-icon-document"></i>
                  {{ data.ruleCount }} 规则
                </div>
              </div>
            </template>
            
            <template #node-interface="{ data, id }">
              <div class="interface-node" :class="[data.interfaceType, { highlighted: highlightedElements.has(id) }]">
                <div class="interface-icon">
                  {{ getInterfaceIcon(data.interfaceType) }}
                </div>
                <div class="interface-label">{{ data.label }}</div>
                <div class="interface-status" v-if="highlightedElements.has(id)">
                  <div class="status-indicator active"></div>
                </div>
              </div>
            </template>
            
            <template #node-decision="{ data, id }">
              <div class="decision-node" :class="{ highlighted: highlightedElements.has(id) }">
                <div class="decision-icon">🔀</div>
                <div class="decision-label">{{ data.label }}</div>
              </div>
            </template>
            
            <template #node-process="{ data, id }">
              <div class="process-node" :class="{ highlighted: highlightedElements.has(id) }">
                <div class="process-icon">⚙️</div>
                <div class="process-label">{{ data.label }}</div>
                <div class="process-activity" v-if="highlightedElements.has(id)">
                  <div class="activity-dot"></div>
                  <div class="activity-dot"></div>
                  <div class="activity-dot"></div>
                </div>
              </div>
            </template>
          </VueFlow>
        </div>
      </div>
    </div>

    <!-- 节点详情对话框 -->
    <el-dialog
      v-model="nodeDetailVisible"
      :title="selectedNode ? `${selectedNode.label} - 详细信息` : '节点详情'"
      width="600px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedNode" class="node-detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ selectedNode.id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selectedNode.type }}</el-descriptions-item>
          <el-descriptions-item label="标签">{{ selectedNode.label }}</el-descriptions-item>
          <el-descriptions-item label="层级">{{ selectedNode.layer }}</el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.interface_name" label="接口名称">
            {{ selectedNode.interface_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.interface_type" label="接口类型">
            {{ selectedNode.interface_type }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.table_name" label="表名称">
            {{ selectedNode.table_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.chain_name" label="链名称">
            {{ selectedNode.chain_name }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.policy" label="策略">
            <el-tag :type="selectedNode.policy === 'ACCEPT' ? 'success' : 'danger'">
              {{ selectedNode.policy }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.rule_count" label="规则数量">
            {{ selectedNode.rule_count }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.rule_number" label="规则编号">
            {{ selectedNode.rule_number }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.packets" label="数据包">
            {{ selectedNode.packets }}
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.bytes" label="字节数">
            {{ selectedNode.bytes }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="selectedNode.properties" class="node-properties">
          <h4>属性信息</h4>
          <el-descriptions :column="1" border>
            <el-descriptions-item 
              v-for="(value, key) in selectedNode.properties" 
              :key="key"
              :label="key"
            >
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-dialog>

    <!-- 错误提示对话框 -->
    <el-dialog
      v-model="errorDialogVisible"
      title="数据加载错误"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="error-content">
        <el-alert
          :title="errorMessage"
          type="error"
          :description="errorDetails"
          show-icon
          :closable="false"
        />
        
        <div class="error-actions">
          <el-button @click="retryLoadData" type="primary">重试</el-button>
          <el-button @click="goToDashboard">返回首页</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, FullScreen, Download, Star, Position, Search, Close, Edit, DocumentCopy, Tools } from '@element-plus/icons-vue'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import type { Node, Edge, Elements } from '@vue-flow/core'
import { MarkerType } from '@vue-flow/core'

// 导入样式
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

// 响应式数据
const loading = ref(false)
const viewMode = ref<'flow' | 'chain'>('chain')
const selectedFlow = ref<string>('')
const selectedNodeInfo = ref<Node | null>(null)
const nodeDetailVisible = ref(false)
const selectedNode = ref<any>(null)
const errorDialogVisible = ref(false)
const errorMessage = ref('')
const errorDetails = ref('')

// 悬停状态管理
const hoveredNodeId = ref<string | null>(null)
const hoveredEdgeId = ref<string | null>(null)
const highlightedElements = ref<Set<string>>(new Set())

// 过滤控件
const protocolFilter = ref<string>('')
const portFilter = ref<string>('')

// Vue Flow 相关
const flowElements = ref<Elements>([])
const { fitView } = useVueFlow()

// 计算属性
const chainNodes = computed(() => {
  return flowElements.value.filter((el: any) => 
    'type' in el && el.type === 'chain'
  ) as Node[]
})

const interfaceNodes = computed(() => {
  return flowElements.value.filter((el: any) => 
    'type' in el && el.type === 'interface'
  ) as Node[]
})

// 生命周期
onMounted(() => {
  initializeFlowElements()
  nextTick(() => {
    loadNodePositions()
    loadLayoutConfiguration()
    loadArrowAdjustments()
    initializeConnectionQuality()
  })
})

// 初始化流程图元素
const initializeFlowElements = () => {
  const nodes: Node[] = [
    // 网络接口节点 - 优化布局，增加间距
    {
      id: 'interface-external',
      type: 'interface',
      position: { x: 50, y: 300 },
      data: { 
        label: '外部网络',
        interfaceType: 'external'
      },
      draggable: true
    },
    {
      id: 'interface-internal',
      type: 'interface', 
      position: { x: 1100, y: 300 },
      data: { 
        label: '内部网络',
        interfaceType: 'internal'
      },
      draggable: true
    },
    
    // IPTables链节点 - 重新布局，增加层次感
    {
      id: 'prerouting',
      type: 'chain',
      position: { x: 250, y: 300 },
      data: {
        label: 'PREROUTING',
        chainType: 'prerouting',
        tables: ['raw', 'mangle', 'nat'],
        ruleCount: 12
      },
      draggable: true
    },
    {
      id: 'routing-decision',
      type: 'decision',
      position: { x: 500, y: 300 },
      data: {
        label: '路由决策'
      },
      draggable: true
    },
    {
      id: 'input',
      type: 'chain',
      position: { x: 700, y: 150 },
      data: {
        label: 'INPUT',
        chainType: 'input',
        tables: ['mangle', 'filter'],
        ruleCount: 8
      },
      draggable: true
    },
    {
      id: 'forward',
      type: 'chain',
      position: { x: 700, y: 300 },
      data: {
        label: 'FORWARD',
        chainType: 'forward',
        tables: ['mangle', 'filter'],
        ruleCount: 15
      },
      draggable: true
    },
    {
      id: 'output',
      type: 'chain',
      position: { x: 700, y: 450 },
      data: {
        label: 'OUTPUT',
        chainType: 'output',
        tables: ['raw', 'mangle', 'nat', 'filter'],
        ruleCount: 6
      },
      draggable: true
    },
    {
      id: 'postrouting',
      type: 'chain',
      position: { x: 900, y: 350 },
      data: {
        label: 'POSTROUTING',
        chainType: 'postrouting',
        tables: ['mangle', 'nat'],
        ruleCount: 4
      },
      draggable: true
    },
    {
      id: 'local-process',
      type: 'process',
      position: { x: 900, y: 150 },
      data: {
        label: '本地进程'
      },
      draggable: true
    }
  ]

  const edges: Edge[] = [
    // 主要数据流路径 - 优化连接路径，使用最直接的直线连接
    {
      id: 'e1',
      source: 'interface-external',
      target: 'prerouting',
      type: 'straight', // 使用直线连接，避免不必要的拐点
      animated: selectedFlow.value === 'forward' || selectedFlow.value === 'input',
      style: { 
        stroke: '#409EFF', 
        strokeWidth: selectedFlow.value === 'forward' || selectedFlow.value === 'input' ? 6 : 4,
        filter: 'drop-shadow(0 3px 8px rgba(64, 158, 255, 0.4))',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 1000 // 确保连接线显示在节点上方
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: '#409EFF', 
        width: 24, // 增大箭头确保可见性
        height: 24,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse', // 使用绝对单位避免缩放问题
        orient: 'auto' // 自动箭头方向
      },
      label: '入站数据包',
      labelStyle: { 
        fill: '#409EFF', 
        fontWeight: 700,
        fontSize: '13px',
        textShadow: '0 1px 2px rgba(0,0,0,0.1)'
      },
      labelBgStyle: { 
        fill: 'rgba(255, 255, 255, 0.95)', 
        fillOpacity: 0.95,
        stroke: '#409EFF',
        strokeWidth: 1,
        strokeOpacity: 0.3
      },
      data: {
        protocol: 'tcp',
        bandwidth: 'high',
        flowType: 'inbound',
        connectionType: 'horizontal' // 标记连接类型
      }
    },
    {
      id: 'e2',
      source: 'prerouting',
      target: 'routing-decision',
      type: 'straight', // 使用直线连接，避免不必要的拐点
      animated: selectedFlow.value === 'forward' || selectedFlow.value === 'input',
      style: { 
        stroke: '#409EFF', 
        strokeWidth: selectedFlow.value === 'forward' || selectedFlow.value === 'input' ? 6 : 4,
        filter: 'drop-shadow(0 3px 8px rgba(64, 158, 255, 0.4))',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        strokeDasharray: selectedFlow.value === 'forward' || selectedFlow.value === 'input' ? '0' : '12,6',
        zIndex: 1000
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: '#409EFF', 
        width: 24,
        height: 24,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto'
      },
      data: {
        protocol: 'all',
        bandwidth: 'high',
        flowType: 'processing',
        connectionType: 'horizontal'
      }
    },
    {
      id: 'e3',
      source: 'routing-decision',
      target: 'input',
      type: 'smoothstep', // 使用智能步进避免对角线遮挡
      animated: selectedFlow.value === 'input',
      pathOptions: {
        borderRadius: 12,
        offset: 25, // 增大偏移避免交叉
        centerX: 0.3, // 调整连接点避免节点中心
        centerY: 0.3
      },
      style: { 
        stroke: selectedFlow.value === 'input' ? '#4CAF50' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'input' ? 7 : 4,
        strokeDasharray: selectedFlow.value === 'input' ? '0' : '10,5',
        filter: selectedFlow.value === 'input' ? 'drop-shadow(0 4px 12px rgba(76, 175, 80, 0.5))' : 'drop-shadow(0 1px 3px rgba(0,0,0,0.2))',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 1001 // 更高层级避免遮挡
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'input' ? '#4CAF50' : '#B0BEC5',
        width: selectedFlow.value === 'input' ? 26 : 22, // 增大箭头
        height: selectedFlow.value === 'input' ? 26 : 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      label: '本地处理',
      labelStyle: { 
        fill: selectedFlow.value === 'input' ? '#4CAF50' : '#666', 
        fontWeight: 700,
        fontSize: '13px',
        textShadow: '0 1px 2px rgba(0,0,0,0.1)'
      },
      labelBgStyle: { 
        fill: 'rgba(255, 255, 255, 0.95)', 
        fillOpacity: 0.95,
        rx: 6,
        ry: 6,
        stroke: selectedFlow.value === 'input' ? '#4CAF50' : '#B0BEC5',
        strokeWidth: 1,
        strokeOpacity: 0.4
      },
      data: {
        protocol: 'tcp',
        bandwidth: 'medium',
        flowType: 'input',
        priority: 'high',
        connectionType: 'diagonal-up' // 对角线向上连接
      }
    },
    {
      id: 'e4',
      source: 'routing-decision',
      target: 'forward',
      type: 'smoothstep', // 智能步进，水平直线连接
      animated: selectedFlow.value === 'forward',
      pathOptions: {
        borderRadius: 8,
        offset: 15, // 适中偏移
        centerX: 0.5, // 水平居中
        centerY: 0.5
      },
      style: { 
        stroke: selectedFlow.value === 'forward' ? '#FF5722' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'forward' ? 8 : 4,
        strokeDasharray: selectedFlow.value === 'forward' ? '0' : '12,6',
        filter: selectedFlow.value === 'forward' ? 'drop-shadow(0 4px 16px rgba(255, 87, 34, 0.6))' : 'drop-shadow(0 1px 3px rgba(0,0,0,0.2))',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 1002 // 最高层级，关键路径
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'forward' ? '#FF5722' : '#B0BEC5',
        width: selectedFlow.value === 'forward' ? 28 : 22, // 增大关键路径箭头
        height: selectedFlow.value === 'forward' ? 28 : 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      label: '转发处理',
      labelStyle: { 
        fill: selectedFlow.value === 'forward' ? '#FF5722' : '#666', 
        fontWeight: 700,
        fontSize: '14px',
        textShadow: '0 2px 4px rgba(0,0,0,0.2)'
      },
      labelBgStyle: { 
        fill: 'rgba(255, 255, 255, 0.95)', 
        fillOpacity: 0.95,
        rx: 8,
        ry: 8,
        stroke: selectedFlow.value === 'forward' ? '#FF5722' : '#B0BEC5',
        strokeWidth: 2,
        strokeOpacity: 0.5
      },
      data: {
        protocol: 'all',
        bandwidth: 'very-high',
        flowType: 'forward',
        priority: 'critical',
        connectionType: 'horizontal' // 水平连接
      }
    },
    {
      id: 'e5',
      source: 'input',
      target: 'local-process',
      type: 'smoothstep', // 智能步进连接
      animated: selectedFlow.value === 'input',
      pathOptions: {
        borderRadius: 10,
        offset: 20,
        centerX: 0.5,
        centerY: 0.5
      },
      style: { 
        stroke: selectedFlow.value === 'input' ? '#4CAF50' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'input' ? 5 : 3,
        filter: selectedFlow.value === 'input' ? 'drop-shadow(0 2px 6px rgba(76, 175, 80, 0.4))' : 'none',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 999
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'input' ? '#4CAF50' : '#B0BEC5',
        width: 22, // 增大箭头
        height: 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      data: {
        connectionType: 'horizontal'
      }
    },
    {
      id: 'e6',
      source: 'forward',
      target: 'postrouting',
      type: 'smoothstep', // 智能步进连接
      animated: selectedFlow.value === 'forward',
      pathOptions: {
        borderRadius: 8,
        offset: 18,
        centerX: 0.5,
        centerY: 0.7 // 调整垂直位置避免交叉
      },
      style: { 
        stroke: selectedFlow.value === 'forward' ? '#FF5722' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'forward' ? 5 : 3,
        filter: selectedFlow.value === 'forward' ? 'drop-shadow(0 2px 6px rgba(255, 87, 34, 0.4))' : 'none',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 998
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'forward' ? '#FF5722' : '#B0BEC5',
        width: 22,
        height: 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      data: {
        connectionType: 'diagonal-down'
      }
    },
    {
      id: 'e7',
      source: 'local-process',
      target: 'output',
      type: 'smoothstep', // 智能步进连接
      animated: selectedFlow.value === 'output',
      pathOptions: {
        borderRadius: 12,
        offset: 25,
        centerX: 0.3, // 调整连接点避免交叉
        centerY: 0.7
      },
      style: { 
        stroke: selectedFlow.value === 'output' ? '#2196F3' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'output' ? 5 : 3,
        filter: selectedFlow.value === 'output' ? 'drop-shadow(0 2px 6px rgba(33, 150, 243, 0.4))' : 'none',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 997
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'output' ? '#2196F3' : '#B0BEC5',
        width: 22,
        height: 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      label: '出站数据包',
      labelStyle: { 
        fill: selectedFlow.value === 'output' ? '#2196F3' : '#666', 
        fontWeight: 600,
        fontSize: '12px'
      },
      labelBgStyle: { 
        fill: 'rgba(255, 255, 255, 0.9)', 
        fillOpacity: 0.9,
        rx: 4,
        ry: 4
      },
      data: {
        connectionType: 'diagonal-down'
      }
    },
    {
      id: 'e8',
      source: 'output',
      target: 'postrouting',
      type: 'smoothstep', // 智能步进连接
      animated: selectedFlow.value === 'output',
      pathOptions: {
        borderRadius: 10,
        offset: 20,
        centerX: 0.5,
        centerY: 0.3 // 调整垂直位置
      },
      style: { 
        stroke: selectedFlow.value === 'output' ? '#2196F3' : '#B0BEC5', 
        strokeWidth: selectedFlow.value === 'output' ? 5 : 3,
        filter: selectedFlow.value === 'output' ? 'drop-shadow(0 2px 6px rgba(33, 150, 243, 0.4))' : 'none',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 996
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: selectedFlow.value === 'output' ? '#2196F3' : '#B0BEC5',
        width: 22,
        height: 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      data: {
        connectionType: 'diagonal-up'
      }
    },
    {
      id: 'e9',
      source: 'postrouting',
      target: 'interface-internal',
      type: 'smoothstep', // 智能步进连接
      animated: selectedFlow.value === 'forward' || selectedFlow.value === 'output',
      pathOptions: {
        borderRadius: 8,
        offset: 20,
        centerX: 0.5,
        centerY: 0.5
      },
      style: { 
        stroke: '#409EFF', 
        strokeWidth: 4,
        filter: 'drop-shadow(0 2px 4px rgba(64, 158, 255, 0.3))',
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        zIndex: 995
      },
      markerEnd: { 
        type: MarkerType.ArrowClosed, 
        color: '#409EFF', 
        width: 22,
        height: 22,
        strokeWidth: 2,
        markerUnits: 'userSpaceOnUse',
        orient: 'auto-start-reverse'
      },
      label: '出站数据包',
      labelStyle: { 
        fill: '#409EFF', 
        fontWeight: 600,
        fontSize: '12px'
      },
      labelBgStyle: { 
        fill: 'rgba(255, 255, 255, 0.9)', 
        fillOpacity: 0.9,
        rx: 4,
        ry: 4
      },
      data: {
        connectionType: 'horizontal'
      }
    }
  ]

  // 为边添加高亮类支持
  const enhancedEdges = edges.map((edge: any) => ({
    ...edge,
    class: highlightedElements.value.has(edge.id) ? 'highlighted' : ''
  }))

  flowElements.value = [...nodes, ...enhancedEdges]
}

// 事件处理
const onViewModeChange = (mode: 'flow' | 'chain') => {
  // 保存当前布局状态
  saveCurrentLayoutState()
  
  viewMode.value = mode
  if (mode === 'chain') {
    initializeFlowElements()
    // 恢复保存的布局状态
    nextTick(() => {
      restoreLayoutState()
    })
  }
}

const onFlowChange = (flow: string) => {
  // 保存当前布局状态
  saveCurrentLayoutState()
  
  selectedFlow.value = flow
  initializeFlowElements() // 重新初始化以更新动画状态
  
  // 恢复保存的布局状态
  nextTick(() => {
    restoreLayoutState()
  })
}

const onNodeClick = (event: any) => {
  selectedNodeInfo.value = event.node
  selectedNode.value = event.node.data
  nodeDetailVisible.value = true
  console.log('Node clicked:', event.node)
}

// 注意：这个函数已经在前面重新定义了，这里保留原有逻辑作为备份
const onEdgeClickOld = (event: any) => {
  console.log('Edge clicked:', event.edge)
  const edge = event.edge
  const edgeData = edge.data || {}
  
  // 显示详细的连接信息
  const protocol = edgeData.protocol || '未知'
  const bandwidth = edgeData.bandwidth || '未知'
  const flowType = edgeData.flowType || '未知'
  const priority = edgeData.priority || '普通'
  
  ElMessage({
    message: `
      <div style="text-align: left;">
        <strong>连接详情:</strong><br/>
        <span style="color: #409EFF;">路径:</span> ${edge.source} → ${edge.target}<br/>
        <span style="color: #67C23A;">协议:</span> ${protocol.toUpperCase()}<br/>
        <span style="color: #E6A23C;">带宽:</span> ${bandwidth}<br/>
        <span style="color: #F56C6C;">流类型:</span> ${flowType}<br/>
        <span style="color: #909399;">优先级:</span> ${priority}
      </div>
    `,
    dangerouslyUseHTMLString: true,
    type: 'info',
    duration: 5000,
    showClose: true
  })
  
  // 高亮显示该连接路径
  highlightConnectionPath(edge.id)
}

const onNodeDragStop = (event: any) => {
  console.log('Node drag stopped:', event.node)
  // 保存节点位置到本地存储
  const nodePositions = JSON.parse(localStorage.getItem('topology-node-positions') || '{}')
  nodePositions[event.node.id] = event.node.position
  localStorage.setItem('topology-node-positions', JSON.stringify(nodePositions))
  
  // 节点拖拽后自动重新计算最优连接路径
  recalculateOptimalPaths(event.node.id)
}

// 重新计算最优连接路径
const recalculateOptimalPaths = (movedNodeId: string) => {
  const edges = flowElements.value.filter((el: any) => 'source' in el) as any[]
  const nodes = flowElements.value.filter((el: any) => 'type' in el) as any[]
  
  // 找到与移动节点相关的所有边
  const affectedEdges = edges.filter(edge => 
    edge.source === movedNodeId || edge.target === movedNodeId
  )
  
  let optimizedCount = 0
  
  affectedEdges.forEach(edge => {
    const sourceNode = nodes.find((n: any) => n.id === edge.source)
    const targetNode = nodes.find((n: any) => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      const optimalPath = calculateOptimalPath(sourceNode, targetNode, nodes, edges)
      
      // 应用新的路径配置
      edge.type = optimalPath.connectionType
      edge.pathOptions = optimalPath.pathOptions
      edge.style = {
        ...edge.style,
        zIndex: optimalPath.zIndex
      }
      
      if (edge.markerEnd) {
        edge.markerEnd.width = optimalPath.arrowSize
        edge.markerEnd.height = optimalPath.arrowSize
      }
      
      optimizedCount++
    }
  })
  
  if (optimizedCount > 0) {
  ElMessage.success(`已重新优化 ${optimizedCount} 个连接路径`)
  }
}

// 应用差异化样式
const applyDifferentiatedStyles = (edges: any[]) => {
  edges.forEach(edge => {
    const flowType = edge.data?.flowType || 'default'
    const priority = edge.data?.priority || 'normal'
    
    // 根据流类型设置样式
    const styleMap = {
      'forward': {
        stroke: '#FF5722',
        strokeWidth: 5,
        strokeDasharray: '0',
        filter: 'drop-shadow(0 2px 6px rgba(255, 87, 34, 0.3))'
      },
      'input': {
        stroke: '#4CAF50',
        strokeWidth: 4,
        strokeDasharray: '0',
        filter: 'drop-shadow(0 2px 6px rgba(76, 175, 80, 0.3))'
      },
      'output': {
        stroke: '#2196F3',
        strokeWidth: 4,
        strokeDasharray: '0',
        filter: 'drop-shadow(0 2px 6px rgba(33, 150, 243, 0.3))'
      },
      'default': {
        stroke: '#666666',
        strokeWidth: 3,
        strokeDasharray: '0',
        filter: 'drop-shadow(0 1px 3px rgba(0, 0, 0, 0.2))'
      }
    }
    
    // 根据优先级调整样式
    const priorityAdjustments = {
      'critical': { strokeWidth: 6, filter: 'drop-shadow(0 0 8px currentColor)' },
      'high': { strokeWidth: 5, filter: 'drop-shadow(0 0 6px currentColor)' },
      'normal': { strokeWidth: 4 },
      'low': { strokeWidth: 3, opacity: 0.7 }
    }
    
    const baseStyle = styleMap[flowType] || styleMap.default
    const priorityStyle = priorityAdjustments[priority] || priorityAdjustments.normal
    
    edge.style = {
      ...edge.style,
      ...baseStyle,
      ...priorityStyle
    }
  })
}

// 减少不必要的连线弯曲和转折
const straightenUnnecessaryBends = (edges: any[], nodes: any[]) => {
  const optimizations = new Map()
  
  edges.forEach(edge => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      const straightenResult = analyzeBendNecessity(sourceNode, targetNode, nodes)
      if (straightenResult.canStraighten) {
        optimizations.set(edge.id, {
          type: 'straight',
          pathOptions: straightenResult.pathOptions
        })
      }
    }
  })
  
  return optimizations
}

// 分析弯曲的必要性
const analyzeBendNecessity = (sourceNode: any, targetNode: any, allNodes: any[]) => {
  const directPath = {
    x1: sourceNode.position.x + 50,
    y1: sourceNode.position.y + 40,
    x2: targetNode.position.x + 50,
    y2: targetNode.position.y + 40
  }
  
  // 检查直线路径是否会穿过其他节点
  const hasObstacles = allNodes.some(node => {
    if (node.id === sourceNode.id || node.id === targetNode.id) return false
    
    const nodeCenter = {
      x: node.position.x + 50,
      y: node.position.y + 40
    }
    
    const distance = pointToLineDistance(nodeCenter, directPath)
    return distance < 60 // 如果距离小于60px，认为有障碍
  })
  
  return {
    canStraighten: !hasObstacles,
    pathOptions: hasObstacles ? null : {
      type: 'straight',
      curvature: 0
    }
  }
}

// 为活跃连接添加动态视觉效果
const addDynamicVisualEffects = (edges: any[]) => {
  edges.forEach(edge => {
    const isActive = edge.data?.active || false
    const bandwidth = edge.data?.bandwidth || 'low'
    
    if (isActive) {
      // 添加流动动画
      edge.animated = true
      edge.class = (edge.class || '') + ' active-connection'
      
      // 根据带宽调整动画速度
      const animationSpeed = {
        'very-high': '0.5s',
        'high': '1s',
        'medium': '1.5s',
        'low': '2s'
      }[bandwidth] || '2s'
      
      edge.style = {
        ...edge.style,
        animationDuration: animationSpeed,
        strokeDasharray: '8,4',
        strokeDashoffset: '0'
      }
    }
  })
}

// 应用路径优化
const applyPathOptimization = (edge: any, optimization: any) => {
  edge.type = optimization.type
  
  if (optimization.controlPoints && optimization.controlPoints.length > 0) {
    edge.pathOptions = {
      curvature: 0.3,
      controlPoints: optimization.controlPoints
    }
  }
  
  // 设置连接质量评分
  connectionQuality.value.set(edge.id, optimization.quality)
  
  // 添加质量指示类
  const qualityClass = getQualityClass(optimization.quality)
  edge.class = (edge.class || '').replace(/quality-\w+/g, '') + ` ${qualityClass}`
}

// 应用箭头优化
const applyArrowOptimization = (edge: any, optimization: any) => {
  if (!edge.markerEnd) {
    edge.markerEnd = {
      type: 'arrowclosed',
      width: optimization.size,
      height: optimization.size
    }
  } else {
    edge.markerEnd.width = optimization.size
    edge.markerEnd.height = optimization.size
  }
  
  // 设置箭头样式
  edge.markerEnd.style = optimization.style
  edge.markerEnd.orient = 'auto'
  edge.markerEnd.markerUnits = 'strokeWidth'
  edge.markerEnd.refX = optimization.offset / optimization.size
}

// 计算连接质量评分
const calculateConnectionQuality = (edges: any[], nodes: any[]) => {
  edges.forEach(edge => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      let quality = 100
      
      // 距离因子
      const distance = Math.sqrt(
        Math.pow(targetNode.position.x - sourceNode.position.x, 2) +
        Math.pow(targetNode.position.y - sourceNode.position.y, 2)
      )
      
      if (distance < 100) quality -= 10 // 距离太近
      if (distance > 400) quality -= 15 // 距离太远
      
      // 角度因子
      const angle = Math.atan2(
        targetNode.position.y - sourceNode.position.y,
        targetNode.position.x - sourceNode.position.x
      ) * (180 / Math.PI)
      
      // 优先水平和垂直连接
      const normalizedAngle = Math.abs(angle % 90)
      if (normalizedAngle > 15 && normalizedAngle < 75) quality -= 5
      
      // 障碍因子
      const obstacles = findObstacleNodes(sourceNode, targetNode, nodes)
      quality -= obstacles.length * 10
      
      connectionQuality.value.set(edge.id, Math.max(quality, 0))
    }
  })
}

// 获取质量等级类名
const getQualityClass = (quality: number) => {
  if (quality >= 90) return 'quality-excellent'
  if (quality >= 75) return 'quality-good'
  if (quality >= 60) return 'quality-fair'
  return 'quality-poor'
}

// 保存优化状态
const saveOptimizationState = () => {
  const state = {
    timestamp: Date.now(),
    nodePositions: {},
    edgeStyles: {},
    connectionQualities: Object.fromEntries(connectionQuality.value)
  }
  
  // 保存节点位置
  flowElements.value.forEach((el: any) => {
    if ('position' in el) {
      state.nodePositions[el.id] = { ...el.position }
    } else if ('style' in el) {
      state.edgeStyles[el.id] = { ...el.style }
    }
  })
  
  optimizationHistory.value.push(state)
  
  // 只保留最近10次优化记录
  if (optimizationHistory.value.length > 10) {
    optimizationHistory.value.shift()
  }
}

// 保存布局配置
const saveLayoutConfiguration = () => {
  const config = {
    nodePositions: {},
    edgeStyles: {},
    connectionQualities: Object.fromEntries(connectionQuality.value),
    timestamp: Date.now()
  }
  
  // 保存当前节点位置
  flowElements.value.forEach((el: any) => {
    if ('position' in el) {
      config.nodePositions[el.id] = { ...el.position }
    } else if ('style' in el) {
      config.edgeStyles[el.id] = { ...el.style }
    }
  })
  
  // 保存到本地存储
  localStorage.setItem('topology-layout-config', JSON.stringify(config))
  layoutConfiguration.value = config
  
  ElMessage.success('布局配置已保存')
}

// 加载布局配置
const loadLayoutConfiguration = () => {
  try {
    const savedConfig = localStorage.getItem('topology-layout-config')
    if (savedConfig) {
      const config = JSON.parse(savedConfig)
      layoutConfiguration.value = config
      
      // 应用保存的连接质量评分
      if (config.connectionQualities) {
        connectionQuality.value = new Map(Object.entries(config.connectionQualities))
      }
      
      console.log('布局配置已加载')
    }
  } catch (error) {
    console.warn('加载布局配置失败:', error)
  }
}

// 保存当前布局状态
const saveCurrentLayoutState = () => {
  try {
    const layoutState = {
      nodePositions: {},
      edgeStyles: {},
      viewMode: viewMode.value,
      selectedFlow: selectedFlow.value,
      timestamp: Date.now()
    }
    
    // 保存节点位置
    flowElements.value.forEach((el: any) => {
      if ('position' in el) {
        layoutState.nodePositions[el.id] = { ...el.position }
      } else if ('style' in el) {
        layoutState.edgeStyles[el.id] = { ...el.style }
      }
    })
    
    localStorage.setItem('topology-layout-state', JSON.stringify(layoutState))
  } catch (error) {
    console.warn('保存布局状态失败:', error)
  }
}

// 恢复布局状态
const restoreLayoutState = () => {
  try {
    const savedState = localStorage.getItem('topology-layout-state')
    if (savedState) {
      const state = JSON.parse(savedState)
      
      // 恢复节点位置
      if (state.nodePositions) {
        flowElements.value.forEach((el: any) => {
          if ('position' in el && state.nodePositions[el.id]) {
            el.position = { ...state.nodePositions[el.id] }
          }
        })
      }
      
      // 恢复边样式
      if (state.edgeStyles) {
        flowElements.value.forEach((el: any) => {
          if ('source' in el && state.edgeStyles[el.id]) {
            el.style = { ...el.style, ...state.edgeStyles[el.id] }
          }
        })
      }
      
      console.log('布局状态已恢复')
    }
  } catch (error) {
    console.warn('恢复布局状态失败:', error)
  }
}

// 加载箭头调整设置
const loadArrowAdjustments = () => {
  try {
    const savedAdjustments = localStorage.getItem('arrow-adjustments')
    if (savedAdjustments) {
      const adjustments = JSON.parse(savedAdjustments)
      
      // 应用保存的箭头调整
      Object.entries(adjustments).forEach(([edgeId, adjustment]: [string, any]) => {
        const edge = flowElements.value.find((el: any) => el.id === edgeId)
        if (edge && edge.markerEnd) {
          if (adjustment.refX !== undefined) {
            edge.markerEnd.refX = adjustment.refX
          }
          if (adjustment.orient !== undefined) {
            edge.markerEnd.orient = adjustment.orient
          }
        }
      })
      
      console.log('箭头调整设置已加载')
    }
  } catch (error) {
    console.warn('加载箭头调整设置失败:', error)
  }
}

// 初始化连接质量评分
const initializeConnectionQuality = () => {
  const edges = flowElements.value.filter((el: any) => 'source' in el)
  const nodes = flowElements.value.filter((el: any) => 'type' in el)
  
  if (edges.length > 0 && nodes.length > 0) {
    calculateConnectionQuality(edges, nodes)
  }
}

// 导出拓扑配置
const exportTopology = () => {
  const config = {
    version: '1.0',
    timestamp: new Date().toISOString(),
    nodePositions: {},
    edgeStyles: {},
    connectionQualities: Object.fromEntries(connectionQuality.value),
    optimizationHistory: optimizationHistory.value,
    layoutConfiguration: layoutConfiguration.value
  }
  
  // 收集当前状态
  flowElements.value.forEach((el: any) => {
    if ('position' in el) {
      config.nodePositions[el.id] = { ...el.position }
    } else if ('style' in el) {
      config.edgeStyles[el.id] = { ...el.style }
    }
  })
  
  // 创建下载链接
  const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `topology-config-${Date.now()}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success('拓扑配置已导出')
}

// 导入拓扑配置
const importTopology = (configFile: File) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const config = JSON.parse(e.target?.result as string)
      
      // 验证配置格式
      if (!config.version || !config.nodePositions) {
        throw new Error('无效的配置文件格式')
      }
      
      // 应用配置
      if (config.nodePositions) {
        Object.entries(config.nodePositions).forEach(([nodeId, position]: [string, any]) => {
          const node = flowElements.value.find((el: any) => el.id === nodeId)
          if (node && 'position' in node) {
            node.position = { ...position }
          }
        })
      }
      
      if (config.edgeStyles) {
        Object.entries(config.edgeStyles).forEach(([edgeId, style]: [string, any]) => {
          const edge = flowElements.value.find((el: any) => el.id === edgeId)
          if (edge && 'style' in edge) {
            edge.style = { ...style }
          }
        })
      }
      
      if (config.connectionQualities) {
        connectionQuality.value = new Map(Object.entries(config.connectionQualities))
      }
      
      if (config.optimizationHistory) {
        optimizationHistory.value = config.optimizationHistory
      }
      
      if (config.layoutConfiguration) {
        layoutConfiguration.value = config.layoutConfiguration
      }
      
      ElMessage.success('拓扑配置已导入')
    } catch (error) {
      ElMessage.error('导入配置失败: ' + error.message)
    }
  }
  reader.readAsText(configFile)
}

// 连接线展开/折叠功能
const connectionExpanded = ref<Set<string>>(new Set())

const toggleConnectionExpansion = (edgeId: string) => {
  if (connectionExpanded.value.has(edgeId)) {
    connectionExpanded.value.delete(edgeId)
    collapseConnection(edgeId)
  } else {
    connectionExpanded.value.add(edgeId)
    expandConnection(edgeId)
  }
}

// 展开连接
const expandConnection = (edgeId: string) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId) as any
  if (edge && 'source' in edge) {
    // 增加连接的可视化强度
    edge.style = {
      ...edge.style,
      strokeWidth: (edge.style.strokeWidth || 4) + 2,
      filter: `${edge.style.filter || ''} drop-shadow(0 0 12px currentColor)`,
      zIndex: (edge.style.zIndex || 1000) + 100
    }
    
    // 增大箭头
    if (edge.markerEnd) {
      edge.markerEnd.width = (edge.markerEnd.width || 22) + 4
      edge.markerEnd.height = (edge.markerEnd.height || 22) + 4
    }
    
    // 添加展开类
    edge.class = (edge.class || '') + ' expanded'
  }
}

// 折叠连接
const collapseConnection = (edgeId: string) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId) as any
  if (edge && 'source' in edge) {
    // 恢复原始样式
    edge.style = {
      ...edge.style,
      strokeWidth: Math.max((edge.style.strokeWidth || 4) - 2, 2),
      filter: edge.style.filter?.replace(' drop-shadow(0 0 12px currentColor)', '') || '',
      zIndex: Math.max((edge.style.zIndex || 1000) - 100, 1000)
    }
    
    // 恢复箭头大小
    if (edge.markerEnd) {
      edge.markerEnd.width = Math.max((edge.markerEnd.width || 22) - 4, 18)
      edge.markerEnd.height = Math.max((edge.markerEnd.height || 22) - 4, 18)
    }
    
    // 移除展开类
    edge.class = (edge.class || '').replace(' expanded', '')
  }
}

// 手动微调连接路径
const manualAdjustMode = ref(false)
const adjustingEdgeId = ref<string | null>(null)
const edgeControlPoints = ref<Map<string, any[]>>(new Map())
const layoutConfiguration = ref<any>({})
const optimizationHistory = ref<any[]>([])
const connectionQuality = ref<Map<string, number>>(new Map())

const enableManualAdjust = (edgeId?: string) => {
  manualAdjustMode.value = true
  adjustingEdgeId.value = edgeId || null
  
  if (edgeId) {
    const edge = flowElements.value.find((el: any) => el.id === edgeId) as any
    if (edge) {
      edge.class = (edge.class || '') + ' manual-adjust'
      // 添加控制点
      addEdgeControlPoints(edge)
    }
  } else {
    // 为所有边添加控制点
    const edges = flowElements.value.filter((el: any) => 'source' in el)
    edges.forEach(edge => {
      edge.class = (edge.class || '') + ' manual-adjust'
      addEdgeControlPoints(edge)
    })
  }
  
  ElMessage.info('手动调整模式已启用，拖拽控制点调整连线路径，点击箭头调整位置')
}

const disableManualAdjust = () => {
  manualAdjustMode.value = false
  
  // 移除所有边的手动调整样式和控制点
  const edges = flowElements.value.filter((el: any) => 'source' in el)
  edges.forEach(edge => {
    edge.class = (edge.class || '').replace(' manual-adjust', '')
    removeEdgeControlPoints(edge.id)
  })
  
  adjustingEdgeId.value = null
  ElMessage.success('手动调整模式已关闭，所有调整已保存')
}

// 添加边控制点
const addEdgeControlPoints = (edge: any) => {
  const sourceNode = flowElements.value.find((el: any) => el.id === edge.source)
  const targetNode = flowElements.value.find((el: any) => el.id === edge.target)
  
  if (sourceNode && targetNode && 'position' in sourceNode && 'position' in targetNode) {
    const controlPoints = calculateControlPoints(sourceNode, targetNode)
    edgeControlPoints.value.set(edge.id, controlPoints)
    
    // 更新边的路径选项
    edge.pathOptions = {
      ...edge.pathOptions,
      controlPoints: controlPoints
    }
  }
}

// 计算控制点
const calculateControlPoints = (sourceNode: any, targetNode: any) => {
  const sx = sourceNode.position.x + 50
  const sy = sourceNode.position.y + 40
  const tx = targetNode.position.x + 50
  const ty = targetNode.position.y + 40
  
  // 计算中点
  const midX = (sx + tx) / 2
  const midY = (sy + ty) / 2
  
  // 计算垂直偏移
  const dx = tx - sx
  const dy = ty - sy
  const distance = Math.sqrt(dx * dx + dy * dy)
  
  // 根据距离调整控制点偏移
  const offset = Math.min(distance * 0.2, 50)
  
  return [
    {
      id: `${sourceNode.id}-${targetNode.id}-cp1`,
      x: midX - dy / distance * offset,
      y: midY + dx / distance * offset,
      type: 'control-point'
    },
    {
      id: `${sourceNode.id}-${targetNode.id}-cp2`,
      x: midX + dy / distance * offset,
      y: midY - dx / distance * offset,
      type: 'control-point'
    }
  ]
}

// 移除边控制点
const removeEdgeControlPoints = (edgeId: string) => {
  edgeControlPoints.value.delete(edgeId)
}

// 处理控制点拖拽
const onControlPointDrag = (controlPointId: string, newPosition: any) => {
  // 找到对应的边
  for (const [edgeId, controlPoints] of edgeControlPoints.value.entries()) {
    const controlPoint = controlPoints.find(cp => cp.id === controlPointId)
    if (controlPoint) {
      // 更新控制点位置
      controlPoint.x = newPosition.x
      controlPoint.y = newPosition.y
      
      // 更新边的路径
      const edge = flowElements.value.find((el: any) => el.id === edgeId)
      if (edge) {
        updateEdgePath(edge, controlPoints)
      }
      break
    }
  }
}

// 更新边路径
const updateEdgePath = (edge: any, controlPoints: any[]) => {
  edge.pathOptions = {
    ...edge.pathOptions,
    controlPoints: controlPoints,
    type: 'bezier'
  }
  
  // 重新计算连接质量
  const sourceNode = flowElements.value.find((el: any) => el.id === edge.source)
  const targetNode = flowElements.value.find((el: any) => el.id === edge.target)
  
  if (sourceNode && targetNode) {
    const quality = calculatePathQuality(sourceNode, targetNode, controlPoints)
    connectionQuality.value.set(edge.id, quality)
    
    // 更新质量指示类
    const qualityClass = getQualityClass(quality)
    edge.class = (edge.class || '').replace(/quality-\w+/g, '') + ` ${qualityClass}`
  }
}

// 计算路径质量
const calculatePathQuality = (sourceNode: any, targetNode: any, controlPoints: any[]) => {
  let quality = 100
  
  // 基础距离评分
  const directDistance = Math.sqrt(
    Math.pow(targetNode.position.x - sourceNode.position.x, 2) +
    Math.pow(targetNode.position.y - sourceNode.position.y, 2)
  )
  
  // 计算实际路径长度
  let pathLength = 0
  let prevX = sourceNode.position.x + 50
  let prevY = sourceNode.position.y + 40
  
  controlPoints.forEach(cp => {
    pathLength += Math.sqrt(Math.pow(cp.x - prevX, 2) + Math.pow(cp.y - prevY, 2))
    prevX = cp.x
    prevY = cp.y
  })
  
  pathLength += Math.sqrt(
    Math.pow(targetNode.position.x + 50 - prevX, 2) +
    Math.pow(targetNode.position.y + 40 - prevY, 2)
  )
  
  // 路径效率评分（实际长度与直线距离的比值）
  const efficiency = directDistance / pathLength
  quality *= efficiency
  
  // 弯曲度评分
  const bendCount = controlPoints.length
  quality -= bendCount * 5
  
  return Math.max(quality, 0)
}

// 自定义箭头位置调整
const adjustArrowPosition = (edgeId: string, offset: number) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId)
  if (edge && edge.markerEnd) {
    edge.markerEnd.refX = offset
    
    // 保存调整
    const adjustments = JSON.parse(localStorage.getItem('arrow-adjustments') || '{}')
    adjustments[edgeId] = { refX: offset }
    localStorage.setItem('arrow-adjustments', JSON.stringify(adjustments))
  }
}

// 自定义箭头角度调整
const adjustArrowAngle = (edgeId: string, angle: number) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId)
  if (edge && edge.markerEnd) {
    edge.markerEnd.orient = `${angle}deg`
    
    // 保存调整
    const adjustments = JSON.parse(localStorage.getItem('arrow-adjustments') || '{}')
    if (!adjustments[edgeId]) adjustments[edgeId] = {}
    adjustments[edgeId].orient = `${angle}deg`
    localStorage.setItem('arrow-adjustments', JSON.stringify(adjustments))
  }
}

// 检测密集连接区域
const detectDenseAreas = () => {
  const edges = flowElements.value.filter((el: any) => 'source' in el) as any[]
  const nodes = flowElements.value.filter((el: any) => 'type' in el) as any[]
  
  // 计算每个区域的连接密度
  const densityMap = new Map<string, number>()
  
  nodes.forEach((node: any) => {
    const connectedEdges = edges.filter(edge => 
      edge.source === node.id || edge.target === node.id
    )
    
    if (connectedEdges.length > 2) {
      densityMap.set(node.id, connectedEdges.length)
      
      // 为密集区域的边添加特殊样式
      connectedEdges.forEach(edge => {
        edge.class = (edge.class || '') + ' dense-area'
      })
    }
  })
  
  const denseCount = densityMap.size
  if (denseCount > 0) {
    ElMessage.info(`检测到 ${denseCount} 个密集连接区域，已应用优化样式`)
  }
  
  return densityMap
}

// 悬停事件处理
const onNodeMouseEnter = (event: any) => {
  hoveredNodeId.value = event.node.id
  highlightConnectedElements(event.node.id)
}

const onNodeMouseLeave = () => {
  hoveredNodeId.value = null
  highlightedElements.value.clear()
}

// 连线点击事件处理
const onEdgeClick = (event: any) => {
  if (manualAdjustMode.value) {
    // 在手动调整模式下，点击连线进行调整
    const edgeId = event.edge.id
    adjustingEdgeId.value = edgeId
    
    // 高亮选中的连线
    highlightConnectionPath(edgeId)
    
    ElMessage.info(`已选中连线 ${edgeId}，可以拖拽控制点调整路径`)
  } else {
    // 普通模式下显示连线信息
    showEdgeDetails(event.edge)
  }
}

// 显示连线详情
const showEdgeDetails = (edge: any) => {
  const sourceNode = flowElements.value.find((el: any) => el.id === edge.source)
  const targetNode = flowElements.value.find((el: any) => el.id === edge.target)
  
  if (sourceNode && targetNode) {
    const quality = connectionQuality.value.get(edge.id) || 0
    const qualityText = quality >= 90 ? '优秀' : quality >= 75 ? '良好' : quality >= 60 ? '一般' : '较差'
    
    ElMessageBox.alert(
      `源节点: ${sourceNode.data?.label || sourceNode.id}\n` +
      `目标节点: ${targetNode.data?.label || targetNode.id}\n` +
      `连接类型: ${edge.data?.flowType || '默认'}\n` +
      `连接质量: ${qualityText} (${quality.toFixed(1)}分)\n` +
      `优先级: ${edge.data?.priority || '普通'}`,
      '连接详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

// 连线双击事件处理
const onEdgeDoubleClick = (event: any) => {
  if (!manualAdjustMode.value) {
    // 双击连线展开/折叠
    const edgeId = event.edge.id
    if (connectionExpanded.value.has(edgeId)) {
      collapseConnection(edgeId)
      connectionExpanded.value.delete(edgeId)
    } else {
      expandConnection(edgeId)
      connectionExpanded.value.add(edgeId)
    }
  }
}

// 连线右键菜单
const onEdgeContextMenu = (event: any) => {
  event.preventDefault()
  
  const edgeId = event.edge.id
  const menuItems = [
    {
      label: '优化此连线',
      action: () => optimizeSingleConnection(edgeId)
    },
    {
      label: '调整箭头位置',
      action: () => showArrowAdjustmentDialog(edgeId)
    },
    {
      label: '设置连线优先级',
      action: () => showPriorityDialog(edgeId)
    },
    {
      label: '复制连线配置',
      action: () => copyConnectionConfig(edgeId)
    }
  ]
  
  // 这里可以显示自定义右键菜单
  console.log('连线右键菜单:', menuItems)
}

// 优化单个连接
const optimizeSingleConnection = (edgeId: string) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId)
  const nodes = flowElements.value.filter((el: any) => 'type' in el)
  
  if (edge && 'source' in edge) {
    const sourceNode = nodes.find((n: any) => n.id === edge.source)
    const targetNode = nodes.find((n: any) => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      const optimization = calculateBestPath(sourceNode, targetNode, nodes, [edge])
      applyPathOptimization(edge, optimization)
      
      const arrowOptimization = calculateOptimalArrowDirection(sourceNode, targetNode, edge)
      applyArrowOptimization(edge, arrowOptimization)
      
      ElMessage.success(`连线 ${edgeId} 已优化`)
    }
  }
}

// 显示箭头调整对话框
const showArrowAdjustmentDialog = (edgeId: string) => {
  ElMessageBox.prompt('请输入箭头偏移量 (0-20)', '调整箭头位置', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^\d+(\.\d+)?$/,
    inputErrorMessage: '请输入有效的数字'
  }).then(({ value }) => {
    const offset = parseFloat(value)
    adjustArrowPosition(edgeId, offset)
    ElMessage.success('箭头位置已调整')
  }).catch(() => {
    // 用户取消
  })
}

// 显示优先级设置对话框
const showPriorityDialog = (edgeId: string) => {
  const priorities = ['low', 'normal', 'high', 'critical']
  const priorityLabels = ['低', '普通', '高', '关键']
  
  ElMessageBox({
    title: '设置连线优先级',
    message: '请选择连线优先级',
    showCancelButton: true,
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(() => {
    // 这里可以显示优先级选择界面
    ElMessage.success('优先级已设置')
  }).catch(() => {
    // 用户取消
  })
}

// 复制连线配置
const copyConnectionConfig = (edgeId: string) => {
  const edge = flowElements.value.find((el: any) => el.id === edgeId)
  if (edge) {
    const config = {
      style: edge.style,
      type: edge.type,
      pathOptions: edge.pathOptions,
      markerEnd: edge.markerEnd,
      data: edge.data
    }
    
    navigator.clipboard.writeText(JSON.stringify(config, null, 2)).then(() => {
      ElMessage.success('连线配置已复制到剪贴板')
    }).catch(() => {
      ElMessage.error('复制失败')
    })
  }
}

const onEdgeMouseEnter = (event: any) => {
  hoveredEdgeId.value = event.edge.id
  highlightedElements.value.add(event.edge.source)
  highlightedElements.value.add(event.edge.target)
  highlightedElements.value.add(event.edge.id)
}

const onEdgeMouseLeave = () => {
  hoveredEdgeId.value = null
  highlightedElements.value.clear()
}

// 高亮连接的元素
const highlightConnectedElements = (nodeId: string) => {
  highlightedElements.value.clear()
  highlightedElements.value.add(nodeId)
  
  // 查找所有连接到该节点的边
  flowElements.value.forEach((element: any) => {
    if ('source' in element && (element.source === nodeId || element.target === nodeId)) {
      highlightedElements.value.add(element.id)
      highlightedElements.value.add(element.source)
      highlightedElements.value.add(element.target)
    }
  })
}

// 高亮特定连接路径 - 增强版
const highlightConnectionPath = (edgeId: string) => {
  highlightedElements.value.clear()
  highlightedElements.value.add(edgeId)
  
  // 查找该边的源节点和目标节点
  const edge = flowElements.value.find((el: any) => el.id === edgeId)
  if (edge && 'source' in edge) {
    highlightedElements.value.add(edge.source)
    highlightedElements.value.add(edge.target)
    
    // 动态调整箭头属性
    adjustArrowProperties(edge)
  }
  
  // 3秒后清除高亮
  setTimeout(() => {
    highlightedElements.value.clear()
  }, 3000)
}

// 动态调整箭头属性 - 防止遮挡和优化显示
const adjustArrowProperties = (edge: any) => {
  const sourceNode = flowElements.value.find((el: any) => el.id === edge.source)
  const targetNode = flowElements.value.find((el: any) => el.id === edge.target)
  
  if (sourceNode && targetNode && 'position' in sourceNode && 'position' in targetNode) {
    // 计算节点间距离
    const dx = targetNode.position.x - sourceNode.position.x
    const dy = targetNode.position.y - sourceNode.position.y
    const distance = Math.sqrt(dx * dx + dy * dy)
    
    // 根据距离动态调整箭头大小
    let arrowSize = 18 // 默认尺寸
    if (distance < 150) {
      arrowSize = 14 // 近距离使用小箭头
    } else if (distance > 300) {
      arrowSize = 22 // 远距离使用大箭头
    }
    
    // 更新边的箭头属性
    if (edge.markerEnd) {
      edge.markerEnd.width = arrowSize
      edge.markerEnd.height = arrowSize
      
      // 添加边缘检测，确保箭头不被节点遮挡
      const nodeRadius = 50 // 假设节点半径
      const offset = nodeRadius + 10 // 箭头偏移量
      edge.markerEnd.markerUnits = 'strokeWidth'
      edge.markerEnd.refX = offset / arrowSize // 动态调整箭头位置
    }
  }
}

// 检测箭头是否被节点遮挡
const detectArrowOverlap = (edge: any) => {
  const sourceNode = flowElements.value.find((el: any) => el.id === edge.source)
  const targetNode = flowElements.value.find((el: any) => el.id === edge.target)
  
  if (sourceNode && targetNode && 'position' in sourceNode && 'position' in targetNode) {
    const nodeSize = 80 // 平均节点尺寸
    const dx = targetNode.position.x - sourceNode.position.x
    const dy = targetNode.position.y - sourceNode.position.y
    const distance = Math.sqrt(dx * dx + dy * dy)
    
    // 如果距离太近，可能存在遮挡
    return distance < nodeSize * 1.5
  }
  
  return false
}

// 智能箭头位置调整 - 全面增强版
const optimizeArrowPositions = () => {
  const edges = flowElements.value.filter(el => 'source' in el) as any[]
  const nodes = flowElements.value.filter(el => 'type' in el) as any[]
  
  let optimizedCount = 0
  
  edges.forEach(edge => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      // 计算最优连接路径
      const optimalPath = calculateOptimalPath(sourceNode, targetNode, nodes, edges)
      
      // 应用路径优化
      if (optimalPath.needsOptimization) {
        edge.type = optimalPath.connectionType
        edge.pathOptions = optimalPath.pathOptions
        edge.style = {
          ...edge.style,
          zIndex: optimalPath.zIndex,
          strokeLinecap: 'round',
          strokeLinejoin: 'round'
        }
        
        // 优化箭头属性
        if (edge.markerEnd) {
          edge.markerEnd.width = optimalPath.arrowSize
          edge.markerEnd.height = optimalPath.arrowSize
          edge.markerEnd.markerUnits = 'userSpaceOnUse'
          edge.markerEnd.orient = 'auto-start-reverse'
          edge.markerEnd.strokeWidth = 2
        }
        
        optimizedCount++
      }
    }
  })
  
  // 重新计算边的层级避免交叉
  optimizeEdgeZIndex(edges)
  
  ElMessage.success(`箭头位置已优化，处理了 ${optimizedCount} 个连接，避免节点遮挡`)
}

// 计算最优连接路径
const calculateOptimalPath = (sourceNode: any, targetNode: any, allNodes: any[], allEdges: any[]) => {
  const dx = targetNode.position.x - sourceNode.position.x
  const dy = targetNode.position.y - sourceNode.position.y
  const distance = Math.sqrt(dx * dx + dy * dy)
  
  // 优先使用直线连接，确保一致性
  let connectionType = 'straight'
  let pathOptions: any = {}
  
  // 检查是否为关键连接（外部网络→PREROUTING，PREROUTING→路由决策）
  const isKeyConnection = (
    (sourceNode.id === 'interface-external' && targetNode.id === 'prerouting') ||
    (sourceNode.id === 'prerouting' && targetNode.id === 'routing-decision')
  )
  
  // 关键连接始终保持直线，不进行避让优化
  if (isKeyConnection) {
    connectionType = 'straight'
    pathOptions = {} // 清空路径选项，使用默认直线
  } else {
    // 检测是否需要避让其他节点
    const needsAvoidance = checkNodeAvoidance(sourceNode, targetNode, allNodes)
    
    // 只有在必须避让时才使用曲线连接
    if (needsAvoidance) {
      connectionType = 'smoothstep'
    
      // 水平连接（左右节点）
      if (Math.abs(dy) < 50 && Math.abs(dx) > 100) {
        pathOptions = {
          borderRadius: 6,
          offset: 25, // 增加偏移避让节点
          centerX: 0.5,
          centerY: 0.5
        }
      }
      // 垂直连接（上下节点）
      else if (Math.abs(dx) < 50 && Math.abs(dy) > 80) {
        pathOptions = {
          borderRadius: 10,
          offset: 30,
          centerX: 0.5,
          centerY: 0.5
        }
      }
      // 对角线连接
      else {
        pathOptions = {
          borderRadius: 12,
          offset: Math.max(25, distance / 8),
          centerX: dx > 0 ? 0.3 : 0.7,
          centerY: dy > 0 ? 0.3 : 0.7
        }
      }
    }
  }
  
  // 计算箭头大小 - 统一标准
  let arrowSize = 24 // 默认大小
  if (distance < 150) {
    arrowSize = 20
  } else if (distance > 300) {
    arrowSize = 28
  }
  
  // 计算Z-index层级
  const zIndex = calculateEdgeZIndex(sourceNode, targetNode, allEdges)
  
  return {
    needsOptimization: needsAvoidance, // 只有需要避让时才标记为需要优化
    connectionType,
    pathOptions,
    arrowSize,
    zIndex
  }
}

// 检查是否需要避让其他节点
const checkNodeAvoidance = (sourceNode: any, targetNode: any, allNodes: any[]) => {
  const path = {
    x1: sourceNode.position.x,
    y1: sourceNode.position.y,
    x2: targetNode.position.x,
    y2: targetNode.position.y
  }
  
  // 检查路径是否经过其他节点
  return allNodes.some(node => {
    if (node.id === sourceNode.id || node.id === targetNode.id) return false
    
    const nodeCenter = {
      x: node.position.x + 50, // 假设节点宽度100px
      y: node.position.y + 40   // 假设节点高度80px
    }
    
    // 计算点到线段的距离
    const distance = pointToLineDistance(nodeCenter, path)
    return distance < 60 // 如果距离小于60px，需要避让
  })
}

// 点到线段距离计算
const pointToLineDistance = (point: any, line: any) => {
  const A = point.x - line.x1
  const B = point.y - line.y1
  const C = line.x2 - line.x1
  const D = line.y2 - line.y1
  
  const dot = A * C + B * D
  const lenSq = C * C + D * D
  
  if (lenSq === 0) return Math.sqrt(A * A + B * B)
  
  let param = dot / lenSq
  param = Math.max(0, Math.min(1, param))
  
  const xx = line.x1 + param * C
  const yy = line.y1 + param * D
  
  const dx = point.x - xx
  const dy = point.y - yy
  
  return Math.sqrt(dx * dx + dy * dy)
}

// 计算边的Z-index层级
const calculateEdgeZIndex = (sourceNode: any, targetNode: any, allEdges: any[]) => {
  // 基础层级
  let baseZIndex = 1000
  
  // 关键路径获得更高层级
  if (sourceNode.data?.chainType === 'forward' || targetNode.data?.chainType === 'forward') {
    baseZIndex += 100
  }
  
  // 根据连接重要性调整
  if (sourceNode.type === 'interface' || targetNode.type === 'interface') {
    baseZIndex += 50
  }
  
  return baseZIndex
}

// 优化边的Z-index避免交叉
const optimizeEdgeZIndex = (edges: any[]) => {
  // 按重要性排序边
  edges.sort((a, b) => {
    const priorityA = getEdgePriority(a)
    const priorityB = getEdgePriority(b)
    return priorityB - priorityA
  })
  
  // 分配Z-index
  edges.forEach((edge, index) => {
    edge.style = {
      ...edge.style,
      zIndex: 1000 + (edges.length - index) * 10
    }
  })
}

// 获取边的优先级
const getEdgePriority = (edge: any) => {
  let priority = 0
  
  // 数据流类型优先级
  if (edge.data?.flowType === 'forward') priority += 100
  if (edge.data?.flowType === 'input') priority += 80
  if (edge.data?.flowType === 'output') priority += 60
  
  // 带宽优先级
  if (edge.data?.bandwidth === 'very-high') priority += 50
  if (edge.data?.bandwidth === 'high') priority += 30
  if (edge.data?.bandwidth === 'medium') priority += 20
  
  // 关键路径优先级
  if (edge.data?.priority === 'critical') priority += 200
  if (edge.data?.priority === 'high') priority += 150
  
  return priority
}

// 智能自动优化布局功能 - 全面增强版
const autoOptimizeLayout = () => {
  const nodes = flowElements.value.filter((el: any) => 'type' in el) as Node[]
  const edges = flowElements.value.filter((el: any) => 'source' in el) as Edge[]
  
  ElMessage.info('正在进行智能布局优化...')
  
  // 保存当前状态到历史记录
  saveOptimizationState()
  
  // 1. 基于当前手动调整的节点位置，计算最优连线路径
  const optimizedPaths = calculateOptimalConnectionPaths(nodes, edges)
  
  // 2. 智能调整箭头方向和位置，避免交叉和重叠
  const arrowOptimizations = optimizeArrowDirections(edges, nodes)
  
  // 3. 应用差异化样式
  applyDifferentiatedStyles(edges)
  
  // 4. 减少不必要的连线弯曲和转折
  const straightenedPaths = straightenUnnecessaryBends(edges, nodes)
  
  // 5. 为活跃连接添加动态视觉效果
  addDynamicVisualEffects(edges)
  
  // 应用所有优化
  let totalOptimizations = 0
  
  optimizedPaths.forEach((optimization, edgeId) => {
    const edge = edges.find(e => e.id === edgeId)
    if (edge) {
      applyPathOptimization(edge, optimization)
      totalOptimizations++
    }
  })
  
  arrowOptimizations.forEach((optimization, edgeId) => {
    const edge = edges.find(e => e.id === edgeId)
    if (edge) {
      applyArrowOptimization(edge, optimization)
    }
  })
  
  // 计算连接质量评分
  calculateConnectionQuality(edges, nodes)
  
  ElMessage.success(`智能优化完成！优化了 ${totalOptimizations} 个连接，提升了整体布局质量`)
}

// 检测两条边是否相交
const edgesIntersect = (edge1: Edge, edge2: Edge, nodes: Node[]): boolean => {
  const node1Start = nodes.find(n => n.id === edge1.source)
  const node1End = nodes.find(n => n.id === edge1.target)
  const node2Start = nodes.find(n => n.id === edge2.source)
  const node2End = nodes.find(n => n.id === edge2.target)
  
  if (!node1Start || !node1End || !node2Start || !node2End) return false
  
  // 简化的线段相交检测
  return lineSegmentsIntersect(
    node1Start.position, node1End.position,
    node2Start.position, node2End.position
  )
}

// 线段相交检测
const lineSegmentsIntersect = (p1: any, p2: any, p3: any, p4: any): boolean => {
  const ccw = (A: any, B: any, C: any) => {
    return (C.y - A.y) * (B.x - A.x) > (B.y - A.y) * (C.x - A.x)
  }
  return ccw(p1, p3, p4) !== ccw(p2, p3, p4) && ccw(p1, p2, p3) !== ccw(p1, p2, p4)
}

// 计算最优连接路径
const calculateOptimalConnectionPaths = (nodes: any[], edges: any[]) => {
  const optimizations = new Map()
  
  edges.forEach(edge => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      const optimization = calculateBestPath(sourceNode, targetNode, nodes, edges)
      optimizations.set(edge.id, optimization)
    }
  })
  
  return optimizations
}

// 计算最佳路径 - 统一标准化处理
const calculateBestPath = (sourceNode: any, targetNode: any, allNodes: any[], allEdges: any[]) => {
  const dx = targetNode.position.x - sourceNode.position.x
  const dy = targetNode.position.y - sourceNode.position.y
  const distance = Math.sqrt(dx * dx + dy * dy)
  
  // 统一使用直线连接作为默认选择
  let pathType = 'straight'
  let controlPoints: any[] = []
  let quality = 100
  
  // 检查是否需要避让其他节点
  const obstacles = findObstacleNodes(sourceNode, targetNode, allNodes)
  
  if (obstacles.length > 0) {
    // 只有在必须避让时才使用曲线路径
    const avoidancePath = calculateAvoidancePath(sourceNode, targetNode, obstacles)
    pathType = avoidancePath.type
    controlPoints = avoidancePath.controlPoints
    quality = avoidancePath.quality
  } else {
    // 所有无障碍连接统一使用直线，确保一致性
    pathType = 'straight'
    quality = 100 // 直线连接质量最高
  }
  
  return {
    type: pathType,
    controlPoints,
    quality,
    distance,
    needsOptimization: obstacles.length > 0 // 只有有障碍时才需要优化
  }
}

// 修复关键连接路径
const fixKeyConnections = () => {
  const edges = flowElements.value.filter((el: any) => 'source' in el)
  
  edges.forEach((edge: any) => {
    // 检查是否为关键连接（外部网络→PREROUTING，PREROUTING→路由决策）
    const isKeyConnection = (
      (edge.source === 'interface-external' && edge.target === 'prerouting') ||
      (edge.source === 'prerouting' && edge.target === 'routing-decision')
    )
    
    if (isKeyConnection) {
      // 强制设置为直线连接
      edge.type = 'straight'
      
      // 移除所有可能导致弯曲的路径选项
      delete edge.pathOptions
      delete edge.curvature
      
      // 确保箭头正确显示
      if (edge.markerEnd) {
        edge.markerEnd.orient = 'auto'
        edge.markerEnd.markerUnits = 'userSpaceOnUse'
        edge.markerEnd.refX = 0
        edge.markerEnd.refY = 0
      }
      
      // 重置样式
      edge.style = {
        ...edge.style,
        strokeLinecap: 'round',
        strokeLinejoin: 'round'
      }
    }
  })
  
  ElMessage.success('关键连接路径已修复为直线')
}

// 标准化连接路径处理
const standardizeConnectionPaths = () => {
  // 首先修复关键连接
  fixKeyConnections()
  
  const edges = flowElements.value.filter((el: any) => 'source' in el)
  const nodes = flowElements.value.filter((el: any) => 'position' in el)
  
  edges.forEach((edge: any) => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      // 检查是否为关键连接（外部网络→PREROUTING，PREROUTING→路由决策）
      const isKeyConnection = (
        (edge.source === 'interface-external' && edge.target === 'prerouting') ||
        (edge.source === 'prerouting' && edge.target === 'routing-decision')
      )
      
      if (isKeyConnection) {
        // 关键连接强制使用直线
        edge.type = 'straight'
        delete edge.pathOptions // 移除路径选项，使用默认直线
        
        // 确保箭头指向正确
        if (edge.markerEnd) {
          edge.markerEnd.orient = 'auto'
          edge.markerEnd.markerUnits = 'userSpaceOnUse'
          edge.markerEnd.refX = 0 // 重置箭头位置
        }
        
        // 重置样式确保直线显示
        edge.style = {
          ...edge.style,
          strokeLinecap: 'round',
          strokeLinejoin: 'round'
        }
      }
    }
  })
  
  ElMessage.success('连接路径已标准化，关键连接使用直线路径')
}

// 查找障碍节点
const findObstacleNodes = (sourceNode: any, targetNode: any, allNodes: any[]) => {
  const obstacles: any[] = []
  const path = {
    x1: sourceNode.position.x + 50, // 节点中心
    y1: sourceNode.position.y + 40,
    x2: targetNode.position.x + 50,
    y2: targetNode.position.y + 40
  }
  
  allNodes.forEach(node => {
    if (node.id !== sourceNode.id && node.id !== targetNode.id) {
      const nodeCenter = {
        x: node.position.x + 50,
        y: node.position.y + 40
      }
      
      const distance = pointToLineDistance(nodeCenter, path)
      if (distance < 70) { // 如果节点太接近连线路径
        obstacles.push({
          node,
          distance,
          center: nodeCenter
        })
      }
    }
  })
  
  return obstacles.sort((a, b) => a.distance - b.distance)
}

// 计算避让路径
const calculateAvoidancePath = (sourceNode: any, targetNode: any, obstacles: any[]) => {
  const controlPoints: any[] = []
  let quality = 80
  
  // 为每个障碍计算避让点
  obstacles.forEach((obstacle, index) => {
    const avoidancePoint = calculateAvoidancePoint(sourceNode, targetNode, obstacle.node)
    controlPoints.push(avoidancePoint)
    quality -= 5 // 每个避让点降低质量评分
  })
  
  return {
    type: 'bezier',
    controlPoints,
    quality: Math.max(quality, 50)
  }
}

// 计算避让点
const calculateAvoidancePoint = (sourceNode: any, targetNode: any, obstacleNode: any) => {
  const sx = sourceNode.position.x + 50
  const sy = sourceNode.position.y + 40
  const tx = targetNode.position.x + 50
  const ty = targetNode.position.y + 40
  const ox = obstacleNode.position.x + 50
  const oy = obstacleNode.position.y + 40
  
  // 计算垂直于连线的避让方向
  const dx = tx - sx
  const dy = ty - sy
  const perpX = -dy
  const perpY = dx
  const perpLength = Math.sqrt(perpX * perpX + perpY * perpY)
  
  if (perpLength === 0) return { x: ox, y: oy }
  
  // 标准化垂直向量
  const unitPerpX = perpX / perpLength
  const unitPerpY = perpY / perpLength
  
  // 避让距离
  const avoidanceDistance = 80
  
  return {
    x: ox + unitPerpX * avoidanceDistance,
    y: oy + unitPerpY * avoidanceDistance
  }
}

// 优化箭头方向
const optimizeArrowDirections = (edges: any[], nodes: any[]) => {
  const optimizations = new Map()
  
  edges.forEach(edge => {
    const sourceNode = nodes.find(n => n.id === edge.source)
    const targetNode = nodes.find(n => n.id === edge.target)
    
    if (sourceNode && targetNode) {
      const arrowOptimization = calculateOptimalArrowDirection(sourceNode, targetNode, edge)
      optimizations.set(edge.id, arrowOptimization)
    }
  })
  
  return optimizations
}

// 计算最优箭头方向
const calculateOptimalArrowDirection = (sourceNode: any, targetNode: any, edge: any) => {
  const dx = targetNode.position.x - sourceNode.position.x
  const dy = targetNode.position.y - sourceNode.position.y
  const distance = Math.sqrt(dx * dx + dy * dy)
  
  // 计算箭头角度
  const angle = Math.atan2(dy, dx) * (180 / Math.PI)
  
  // 根据距离调整箭头大小
  let size = 20
  if (distance < 150) {
    size = 16
  } else if (distance > 300) {
    size = 24
  }
  
  // 计算箭头位置偏移，避免被节点遮挡
  const nodeRadius = 50
  const offset = nodeRadius + 8
  
  return {
    angle,
    size,
    offset,
    visible: true,
    style: {
      fill: getArrowColor(edge),
      stroke: getArrowStrokeColor(edge),
      strokeWidth: 1
    }
  }
}

// 获取箭头颜色
const getArrowColor = (edge: any) => {
  const flowType = edge.data?.flowType || 'default'
  const colorMap = {
    'forward': '#FF5722',
    'input': '#4CAF50',
    'output': '#2196F3',
    'default': '#666666'
  }
  return colorMap[flowType] || colorMap.default
}

// 获取箭头描边颜色
const getArrowStrokeColor = (edge: any) => {
  const flowType = edge.data?.flowType || 'default'
  const colorMap = {
    'forward': '#D84315',
    'input': '#388E3C',
    'output': '#1976D2',
    'default': '#444444'
  }
  return colorMap[flowType] || colorMap.default
}

// 调整节点位置以减少边交叉
const adjustNodesForBetterLayout = (nodes: Node[], edges: Edge[]) => {
  // 使用改进的力导向算法
  const iterations = 100
  const repulsionStrength = 1000
  const attractionStrength = 0.1
  const dampening = 0.9
  
  for (let iter = 0; iter < iterations; iter++) {
    nodes.forEach(node => {
      let fx = 0, fy = 0
      
      // 节点间排斥力
      nodes.forEach(otherNode => {
        if (node.id !== otherNode.id) {
          const dx = node.position.x - otherNode.position.x
          const dy = node.position.y - otherNode.position.y
          const distance = Math.sqrt(dx * dx + dy * dy) || 1
          
          const force = repulsionStrength / (distance * distance)
          fx += (dx / distance) * force
          fy += (dy / distance) * force
        }
      })
      
      // 连接边的吸引力
      edges.forEach(edge => {
        if (edge.source === node.id || edge.target === node.id) {
          const connectedNodeId = edge.source === node.id ? edge.target : edge.source
          const connectedNode = nodes.find(n => n.id === connectedNodeId)
          
          if (connectedNode) {
            const dx = connectedNode.position.x - node.position.x
            const dy = connectedNode.position.y - node.position.y
            const distance = Math.sqrt(dx * dx + dy * dy) || 1
            
            const idealDistance = 200
            const force = attractionStrength * (distance - idealDistance)
            fx += (dx / distance) * force
            fy += (dy / distance) * force
          }
        }
      })
      
      // 应用力并添加阻尼
      node.position.x += fx * dampening
      node.position.y += fy * dampening
      
      // 边界约束
      node.position.x = Math.max(100, Math.min(1100, node.position.x))
      node.position.y = Math.max(100, Math.min(500, node.position.y))
    })
  }
}

const resetView = () => {
  // 询问用户是否确认重置布局
  ElMessageBox.confirm(
    '重置视图将清除所有自定义布局设置，是否继续？',
    '确认重置',
    {
      confirmButtonText: '确认重置',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    // 清除保存的布局状态
    localStorage.removeItem('topology-layout-state')
    localStorage.removeItem('topology-node-positions')
    localStorage.removeItem('topology-layout-config')
    
    selectedFlow.value = ''
    selectedNodeInfo.value = null
    protocolFilter.value = ''
    portFilter.value = ''
    initializeFlowElements()
    nextTick(() => {
      fitView()
    })
    
    ElMessage.success('视图已重置到默认状态')
  }).catch(() => {
    ElMessage.info('已取消重置操作')
  })
}

const applyFilters = () => {
  // 根据过滤条件重新初始化元素
  initializeFlowElements()
}

// 重复的exportTopology函数已删除，使用前面定义的增强版本

// 错误处理方法
const retryLoadData = () => {
  errorDialogVisible.value = false
  loading.value = true
  setTimeout(() => {
    initializeFlowElements()
    loading.value = false
  }, 1000)
}

const goToDashboard = () => {
  errorDialogVisible.value = false
  // 这里可以添加路由跳转逻辑
  ElMessage.info('返回首页功能待实现')
}

// 加载保存的节点位置
const loadNodePositions = () => {
  try {
    const savedPositions = JSON.parse(localStorage.getItem('topology-node-positions') || '{}')
    flowElements.value.forEach((element: any) => {
      if ('type' in element && savedPositions[element.id]) {
        element.position = savedPositions[element.id]
      }
    })
  } catch (error) {
    console.warn('Failed to load saved node positions:', error)
  }
}

// 自动布局优化
const optimizeLayout = () => {
  // 简单的力导向布局算法
  const nodes = flowElements.value.filter(el => 'type' in el) as Node[]
  const edges = flowElements.value.filter(el => 'source' in el) as Edge[]
  
  // 计算节点间的理想距离
  const idealDistance = 200
  const iterations = 50
  
  for (let i = 0; i < iterations; i++) {
    nodes.forEach(node => {
      let fx = 0, fy = 0
      
      // 排斥力
      nodes.forEach(otherNode => {
        if (node.id !== otherNode.id) {
          const dx = node.position.x - otherNode.position.x
          const dy = node.position.y - otherNode.position.y
          const distance = Math.sqrt(dx * dx + dy * dy) || 1
          
          if (distance < idealDistance) {
            const force = (idealDistance - distance) / distance
            fx += dx * force * 0.1
            fy += dy * force * 0.1
          }
        }
      })
      
      // 吸引力（连接的节点）
      edges.forEach(edge => {
        if (edge.source === node.id || edge.target === node.id) {
          const connectedNodeId = edge.source === node.id ? edge.target : edge.source
          const connectedNode = nodes.find(n => n.id === connectedNodeId)
          
          if (connectedNode) {
            const dx = connectedNode.position.x - node.position.x
            const dy = connectedNode.position.y - node.position.y
            const distance = Math.sqrt(dx * dx + dy * dy) || 1
            
            const force = Math.log(distance / idealDistance) * 0.05
            fx += dx * force
            fy += dy * force
          }
        }
      })
      
      // 应用力
      node.position.x += fx
      node.position.y += fy
      
      // 边界约束
      node.position.x = Math.max(50, Math.min(1200, node.position.x))
      node.position.y = Math.max(50, Math.min(600, node.position.y))
    })
  }
  
  ElMessage.success('布局已优化')
}

// 工具方法
const getInterfaceIcon = (interfaceType: string): string => {
  switch (interfaceType) {
    case 'external': return '🌐'
    case 'internal': return '🏠'
    case 'docker': return '🐳'
    case 'wifi': return '📡'
    default: return '🖧'
  }
}

const getTableTagType = (table: string): string => {
  switch (table) {
    case 'raw': return 'danger'
    case 'mangle': return 'warning'
    case 'nat': return 'info'
    case 'filter': return 'success'
    default: return 'default'
  }
}
</script>

<style scoped>
.topology-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.topology-header {
  padding: 20px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.topology-header h2 {
  margin: 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.topology-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.topology-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.topology-sidebar {
  width: 320px;
  padding: 20px;
  background: white;
  border-right: 1px solid #e4e7ed;
  overflow-y: auto;
  box-shadow: 2px 0 8px rgba(0,0,0,0.1);
}

.topology-main {
  flex: 1;
  position: relative;
}

.vue-flow-container {
  height: 100%;
  width: 100%;
}

/* Vue Flow 自定义样式 */
.iptables-flow {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

/* 图例样式 */
.legend-card, .node-info-card, .stats-card {
  margin-bottom: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.legend-section {
  border-bottom: 1px solid #e4e7ed;
  padding-bottom: 12px;
}

.legend-section:last-child {
  border-bottom: none;
}

.legend-section h4 {
  margin: 0 0 10px 0;
  font-size: 13px;
  color: #606266;
  font-weight: 600;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 12px;
  color: #303133;
}

.legend-icon {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-line {
  width: 24px;
  height: 3px;
  border-radius: 2px;
  flex-shrink: 0;
}

/* 图例图标样式 */
.interface-icon {
  background: linear-gradient(45deg, #409EFF, #337ecc);
}

.chain-icon {
  background: linear-gradient(45deg, #67C23A, #529b2e);
}

.decision-icon {
  background: linear-gradient(45deg, #FFC107, #FF9800);
}

.process-icon {
  background: linear-gradient(45deg, #9C27B0, #673AB7);
}

/* 图例连接线样式 */
.forward-flow {
  background: linear-gradient(90deg, #FF5722, #D84315);
}

.input-flow {
  background: linear-gradient(90deg, #4CAF50, #388E3C);
}

.output-flow {
  background: linear-gradient(90deg, #2196F3, #1976D2);
}

.return-flow {
  background: linear-gradient(90deg, #9C27B0, #7B1FA2);
}

/* 统计卡片样式 */
.stats-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid #f0f0f0;
}

.stat-item:last-child {
  border-bottom: none;
}

.stat-label {
  font-size: 13px;
  color: #606266;
}

.stat-value {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

/* 节点信息卡片样式 */
.node-info-content {
  font-size: 13px;
}

/* 自定义节点样式 */
:deep(.chain-node) {
  background: white;
  border: 3px solid #e1e5e9;
  border-radius: 16px;
  padding: 20px;
  min-width: 160px;
  min-height: 80px;
  box-shadow: 0 6px 16px rgba(0,0,0,0.12);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

:deep(.chain-node::before) {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #409EFF, #67C23A, #E6A23C, #F56C6C);
  opacity: 0;
  transition: opacity 0.3s ease;
}

:deep(.chain-node:hover) {
  transform: translateY(-4px) scale(1.02);
  box-shadow: 0 12px 32px rgba(0,0,0,0.2);
  border-color: #409EFF;
}

:deep(.chain-node:hover::before) {
  opacity: 1;
}

:deep(.chain-node.highlighted) {
  border-color: #409EFF;
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.3), 0 12px 32px rgba(0,0,0,0.2);
  transform: translateY(-2px) scale(1.05);
}

:deep(.chain-node.prerouting) {
  border-color: #FF9800;
  background: linear-gradient(135deg, #fff8e1 0%, #ffecb3 100%);
}

:deep(.chain-node.input) {
  border-color: #4CAF50;
  background: linear-gradient(135deg, #f1f8e9 0%, #c8e6c9 100%);
}

:deep(.chain-node.forward) {
  border-color: #FF5722;
  background: linear-gradient(135deg, #fbe9e7 0%, #ffab91 100%);
}

:deep(.chain-node.output) {
  border-color: #2196F3;
  background: linear-gradient(135deg, #e3f2fd 0%, #90caf9 100%);
}

:deep(.chain-node.postrouting) {
  border-color: #9C27B0;
  background: linear-gradient(135deg, #f3e5f5 0%, #ce93d8 100%);
}

:deep(.chain-header) {
  margin-bottom: 8px;
}

:deep(.chain-title) {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

:deep(.chain-tables) {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

:deep(.table-tag) {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.table-tag:hover) {
  transform: scale(1.05);
}

:deep(.table-tag.raw) {
  background: #f44336;
}

:deep(.table-tag.mangle) {
  background: #ff9800;
}

:deep(.table-tag.nat) {
  background: #2196f3;
}

:deep(.table-tag.filter) {
  background: #4caf50;
}

:deep(.chain-stats) {
  font-size: 11px;
  color: #666;
  text-align: center;
}

/* 接口节点样式 */
:deep(.interface-node) {
  background: white;
  border: 4px solid #409EFF;
  border-radius: 50%;
  width: 100px;
  height: 100px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 16px rgba(0,0,0,0.12);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  position: relative;
}

:deep(.interface-node::after) {
  content: '';
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  border-radius: 50%;
  border: 2px solid transparent;
  background: linear-gradient(45deg, #409EFF, #67C23A) border-box;
  mask: linear-gradient(#fff 0 0) padding-box, linear-gradient(#fff 0 0);
  mask-composite: exclude;
  opacity: 0;
  transition: opacity 0.3s ease;
}

:deep(.interface-node:hover) {
  transform: scale(1.1);
  box-shadow: 0 8px 24px rgba(0,0,0,0.2);
}

:deep(.interface-node:hover::after) {
  opacity: 1;
}

:deep(.interface-node.highlighted) {
  border-color: #67C23A;
  box-shadow: 0 0 0 6px rgba(103, 194, 58, 0.3), 0 8px 24px rgba(0,0,0,0.2);
  transform: scale(1.15);
}

:deep(.interface-node.external) {
  border-color: #FF5722;
  background: linear-gradient(135deg, #ffebee 0%, #ffcdd2 100%);
}

:deep(.interface-node.internal) {
  border-color: #4CAF50;
  background: linear-gradient(135deg, #e8f5e8 0%, #c8e6c9 100%);
}

:deep(.interface-icon) {
  font-size: 24px;
  margin-bottom: 4px;
}

:deep(.interface-label) {
  font-size: 10px;
  font-weight: 600;
  color: #303133;
  text-align: center;
}

/* 决策节点样式 */
:deep(.decision-node) {
  background: linear-gradient(135deg, #FFC107 0%, #FF9800 100%);
  border: 3px solid #F57C00;
  border-radius: 12px;
  width: 120px;
  height: 80px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 16px rgba(0,0,0,0.12);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  transform: rotate(45deg);
  position: relative;
}

:deep(.decision-node::before) {
  content: '';
  position: absolute;
  top: -3px;
  left: -3px;
  right: -3px;
  bottom: -3px;
  background: linear-gradient(135deg, #FFD54F, #FF8F00);
  border-radius: 12px;
  z-index: -1;
  opacity: 0;
  transition: opacity 0.3s ease;
}

:deep(.decision-node:hover) {
  transform: rotate(45deg) scale(1.1);
  box-shadow: 0 8px 24px rgba(0,0,0,0.2);
}

:deep(.decision-node:hover::before) {
  opacity: 1;
}

:deep(.decision-node.highlighted) {
  box-shadow: 0 0 0 4px rgba(255, 193, 7, 0.4), 0 8px 24px rgba(0,0,0,0.2);
  transform: rotate(45deg) scale(1.15);
}

:deep(.decision-icon) {
  font-size: 20px;
  transform: rotate(-45deg);
}

:deep(.decision-label) {
  font-size: 10px;
  font-weight: 600;
  color: white;
  text-align: center;
  transform: rotate(-45deg);
  margin-top: 2px;
}

/* 进程节点样式 */
:deep(.process-node) {
  background: linear-gradient(135deg, #9C27B0 0%, #673AB7 100%);
  border: 3px solid #7B1FA2;
  border-radius: 16px;
  width: 110px;
  height: 80px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 16px rgba(0,0,0,0.12);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

:deep(.process-node::before) {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
  transition: left 0.6s ease;
}

:deep(.process-node:hover) {
  transform: translateY(-4px) scale(1.05);
  box-shadow: 0 12px 32px rgba(0,0,0,0.25);
}

:deep(.process-node:hover::before) {
  left: 100%;
}

:deep(.process-node.highlighted) {
  border-color: #E91E63;
  box-shadow: 0 0 0 4px rgba(156, 39, 176, 0.3), 0 12px 32px rgba(0,0,0,0.25);
  transform: translateY(-2px) scale(1.1);
}

:deep(.process-icon) {
  font-size: 20px;
  margin-bottom: 4px;
}

:deep(.process-label) {
  font-size: 11px;
  font-weight: 600;
  color: white;
  text-align: center;
}

/* Vue Flow 控制面板样式 */
:deep(.vue-flow__controls) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  backdrop-filter: blur(10px);
}

:deep(.vue-flow__controls-button) {
  border: none;
  background: transparent;
  color: #4a5568;
  transition: all 0.2s ease;
  border-radius: 4px;
}

:deep(.vue-flow__controls-button:hover) {
  background: #e2e8f0;
  color: #2d3748;
  transform: scale(1.05);
}

/* Vue Flow 小地图样式 */
:deep(.vue-flow__minimap) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  backdrop-filter: blur(10px);
}

/* Vue Flow 边标签样式 */
:deep(.vue-flow__edge-label) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: #2d3748;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
  z-index: 2000; /* 确保标签始终在最上层 */
}

/* 智能连接路径样式 - 消除节点遮挡 */
:deep(.vue-flow__edge) {
  pointer-events: stroke; /* 只有线条部分可点击 */
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.vue-flow__edge path) {
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  /* 确保连线始终在节点上方 */
  z-index: inherit;
}

/* 水平连接优化 */
:deep(.vue-flow__edge[data-connection-type="horizontal"]) {
  z-index: 1000;
}

:deep(.vue-flow__edge[data-connection-type="horizontal"] path) {
  stroke-width: 4px;
  filter: drop-shadow(0 2px 6px rgba(0,0,0,0.2));
}

/* 垂直连接优化 */
:deep(.vue-flow__edge[data-connection-type="vertical"]) {
  z-index: 1001;
}

:deep(.vue-flow__edge[data-connection-type="vertical"] path) {
  stroke-width: 4px;
  filter: drop-shadow(0 2px 6px rgba(0,0,0,0.2));
}

/* 对角线连接优化 */
:deep(.vue-flow__edge[data-connection-type="diagonal-up"]),
:deep(.vue-flow__edge[data-connection-type="diagonal-down"]) {
  z-index: 1002;
}

:deep(.vue-flow__edge[data-connection-type="diagonal-up"] path),
:deep(.vue-flow__edge[data-connection-type="diagonal-down"] path) {
  stroke-width: 4px;
  filter: drop-shadow(0 3px 8px rgba(0,0,0,0.25));
  /* 对角线连接使用更明显的阴影 */
}

/* 智能避让样式 */
:deep(.vue-flow__edge.smart-avoidance) {
  z-index: 1100;
}

:deep(.vue-flow__edge.smart-avoidance path) {
  stroke-dasharray: 0;
  opacity: 1;
  filter: drop-shadow(0 4px 12px rgba(0,0,0,0.3)) drop-shadow(0 0 8px currentColor);
}

/* 连接密集区域的展开效果 */
:deep(.vue-flow__edge.dense-area) {
  animation: dense-area-pulse 3s infinite;
}

@keyframes dense-area-pulse {
  0%, 100% {
    opacity: 0.8;
  }
  50% {
    opacity: 1;
    filter: drop-shadow(0 0 12px currentColor);
  }
}

/* 手动调整模式增强样式 */
:deep(.vue-flow__edge.manual-adjust) {
  cursor: grab;
  stroke-dasharray: 8,4;
  opacity: 0.9;
}

:deep(.vue-flow__edge.manual-adjust:hover) {
  stroke-dasharray: none;
  opacity: 1;
  cursor: grabbing;
}

:deep(.vue-flow__edge.manual-adjust path) {
  stroke-dasharray: 8,4;
  opacity: 0.9;
  cursor: pointer;
}

:deep(.vue-flow__edge.manual-adjust:hover path) {
  stroke-dasharray: none;
  opacity: 1;
  stroke-width: 6px;
  filter: drop-shadow(0 0 8px currentColor);
}

/* 连接线margin优化 - 避免紧贴节点边缘 */
:deep(.vue-flow__edge path) {
  stroke-width: 4px;
  stroke-linecap: round;
  stroke-linejoin: round;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 智能优化后的连线样式 */
:deep(.vue-flow__edge.optimized-path) {
  z-index: 1050;
}

:deep(.vue-flow__edge.optimized-path path) {
  filter: drop-shadow(0 2px 8px rgba(0,0,0,0.15));
}

/* 连接质量指示样式 */
:deep(.vue-flow__edge.quality-excellent path) {
  stroke: #4CAF50;
  filter: drop-shadow(0 0 6px rgba(76, 175, 80, 0.4));
}

:deep(.vue-flow__edge.quality-good path) {
  stroke: #8BC34A;
  filter: drop-shadow(0 0 4px rgba(139, 195, 74, 0.3));
}

:deep(.vue-flow__edge.quality-fair path) {
  stroke: #FF9800;
  filter: drop-shadow(0 0 4px rgba(255, 152, 0, 0.3));
}

:deep(.vue-flow__edge.quality-poor path) {
  stroke: #F44336;
  stroke-dasharray: 6,3;
  filter: drop-shadow(0 0 4px rgba(244, 67, 54, 0.3));
}

/* 活跃连接动画 */
:deep(.vue-flow__edge.active-connection path) {
  animation: connection-flow 2s linear infinite;
}

@keyframes connection-flow {
  0% {
    stroke-dashoffset: 0;
  }
  100% {
    stroke-dashoffset: -12;
  }
}

/* 控制点样式 */
.control-point {
  position: absolute;
  width: 12px;
  height: 12px;
  background: #409EFF;
  border: 2px solid white;
  border-radius: 50%;
  cursor: grab;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  z-index: 2000;
  transition: all 0.2s ease;
}

.control-point:hover {
  transform: scale(1.3);
  background: #66B2FF;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.control-point:active {
  cursor: grabbing;
  transform: scale(1.1);
}

/* 箭头调整指示器 */
.arrow-adjustment-indicator {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #E6A23C;
  border-radius: 50%;
  cursor: pointer;
  z-index: 1500;
  animation: arrow-indicator-pulse 2s infinite;
}

@keyframes arrow-indicator-pulse {
  0%, 100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

/* 展开连接样式 */
:deep(.vue-flow__edge.expanded) {
  z-index: 1200 !important;
}

:deep(.vue-flow__edge.expanded path) {
  stroke-width: 8px !important;
  filter: drop-shadow(0 0 16px currentColor) drop-shadow(0 4px 12px rgba(0,0,0,0.4)) !important;
  animation: expanded-pulse 2s infinite;
}

:deep(.vue-flow__edge.expanded .vue-flow__edge-marker) {
  transform: scale(1.3);
  filter: drop-shadow(0 0 12px currentColor);
}

@keyframes expanded-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.8;
    transform: scale(1.02);
  }
}

/* 智能路径计算指示器 */
.path-calculation-indicator {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(64, 158, 255, 0.9);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  z-index: 3000;
  animation: calculation-fade 2s ease-out forwards;
}

@keyframes calculation-fade {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
  80% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.05);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.95);
  }
}

/* 连接质量指示器 */
:deep(.vue-flow__edge.high-quality) {
  filter: drop-shadow(0 0 8px #4CAF50);
}

:deep(.vue-flow__edge.medium-quality) {
  filter: drop-shadow(0 0 6px #FF9800);
}

:deep(.vue-flow__edge.low-quality) {
  filter: drop-shadow(0 0 4px #F44336);
  stroke-dasharray: 6,3;
}

/* 连接优化建议提示 */
.connection-suggestion {
  position: absolute;
  background: rgba(255, 193, 7, 0.95);
  color: #333;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  z-index: 2500;
  animation: suggestion-bounce 0.5s ease-out;
  pointer-events: none;
}

@keyframes suggestion-bounce {
  0% {
    opacity: 0;
    transform: translateY(10px) scale(0.8);
  }
  60% {
    opacity: 1;
    transform: translateY(-2px) scale(1.05);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 路径优化成功指示 */
:deep(.vue-flow__edge.optimized) {
  animation: optimization-success 1s ease-out;
}

@keyframes optimization-success {
  0% {
    filter: drop-shadow(0 0 0px currentColor);
  }
  50% {
    filter: drop-shadow(0 0 20px currentColor) drop-shadow(0 0 40px currentColor);
  }
  100% {
    filter: drop-shadow(0 2px 6px rgba(0,0,0,0.2));
  }
}

/* 智能避让成功指示 */
:deep(.vue-flow__edge.avoidance-success) {
  stroke-dasharray: 0 !important;
  opacity: 1 !important;
  animation: avoidance-highlight 1.5s ease-out;
}

@keyframes avoidance-highlight {
  0% {
    stroke-width: 2px;
  }
  50% {
    stroke-width: 8px;
    filter: drop-shadow(0 0 16px currentColor);
  }
  100% {
    stroke-width: 4px;
    filter: drop-shadow(0 2px 6px rgba(0,0,0,0.2));
  }
}

/* 连接路径质量评分显示 */
.connection-quality-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: bold;
  color: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  z-index: 2000;
}

.connection-quality-badge.excellent {
  background: #4CAF50;
}

.connection-quality-badge.good {
  background: #8BC34A;
}

.connection-quality-badge.fair {
  background: #FF9800;
}

.connection-quality-badge.poor {
  background: #F44336;
}

/* 响应式优化 - 移动端连接线调整 */
@media (max-width: 768px) {
  :deep(.vue-flow__edge path) {
    stroke-width: 3px !important;
  }
  
  :deep(.vue-flow__edge-marker) {
    transform: scale(0.8);
  }
  
  :deep(.vue-flow__edge-label) {
    font-size: 10px;
    padding: 4px 8px;
  }
}

@media (max-width: 480px) {
  :deep(.vue-flow__edge path) {
    stroke-width: 2px !important;
  }
  
  :deep(.vue-flow__edge-marker) {
    transform: scale(0.6);
  }
  
  :deep(.vue-flow__edge-label) {
    display: none; /* 小屏幕隐藏标签 */
  }
}

/* 边的高亮样式 - 全面增强 */
:deep(.vue-flow__edge.highlighted) {
  z-index: 1000;
}

:deep(.vue-flow__edge.highlighted path) {
  stroke-width: 8px !important;
  filter: drop-shadow(0 0 12px currentColor) drop-shadow(0 0 24px currentColor) !important;
  animation: pulse-edge 2s infinite, flow-animation 3s linear infinite;
}

:deep(.vue-flow__edge.highlighted .vue-flow__edge-label) {
  background: rgba(64, 158, 255, 0.95);
  color: white;
  transform: scale(1.15);
  box-shadow: 0 8px 24px rgba(64, 158, 255, 0.5);
  border: 2px solid rgba(255, 255, 255, 0.3);
  animation: label-glow 2s infinite;
}

/* 协议类型特定样式 */
:deep(.vue-flow__edge[data-protocol="tcp"]) {
  stroke-dasharray: none;
}

:deep(.vue-flow__edge[data-protocol="udp"]) {
  stroke-dasharray: 8,4;
}

:deep(.vue-flow__edge[data-protocol="icmp"]) {
  stroke-dasharray: 2,2;
}

/* 带宽指示样式 */
:deep(.vue-flow__edge[data-bandwidth="very-high"] path) {
  stroke-width: 8px;
  filter: drop-shadow(0 2px 8px rgba(0,0,0,0.3));
}

:deep(.vue-flow__edge[data-bandwidth="high"] path) {
  stroke-width: 6px;
  filter: drop-shadow(0 2px 6px rgba(0,0,0,0.25));
}

:deep(.vue-flow__edge[data-bandwidth="medium"] path) {
  stroke-width: 4px;
  filter: drop-shadow(0 1px 4px rgba(0,0,0,0.2));
}

/* 流类型颜色增强 */
:deep(.vue-flow__edge[data-flow-type="inbound"] path) {
  stroke: #409EFF;
  background: linear-gradient(90deg, #409EFF, #67C23A);
}

:deep(.vue-flow__edge[data-flow-type="forward"] path) {
  stroke: #FF5722;
  background: linear-gradient(90deg, #FF5722, #FF9800);
}

:deep(.vue-flow__edge[data-flow-type="input"] path) {
  stroke: #4CAF50;
  background: linear-gradient(90deg, #4CAF50, #8BC34A);
}

:deep(.vue-flow__edge[data-flow-type="output"] path) {
  stroke: #2196F3;
  background: linear-gradient(90deg, #2196F3, #03A9F4);
}

/* 边的脉冲动画 - 增强版 */
@keyframes pulse-edge {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.02);
  }
}

/* 数据流动动画 */
@keyframes flow-animation {
  0% {
    stroke-dasharray: 20,10;
    stroke-dashoffset: 0;
  }
  100% {
    stroke-dasharray: 20,10;
    stroke-dashoffset: 30;
  }
}

/* 标签发光动画 */
@keyframes label-glow {
  0%, 100% {
    box-shadow: 0 8px 24px rgba(64, 158, 255, 0.5);
  }
  50% {
    box-shadow: 0 12px 32px rgba(64, 158, 255, 0.7), 0 0 20px rgba(64, 158, 255, 0.5);
  }
}

/* 跳线效果动画 */
@keyframes jump-line {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

/* 活跃连接的脉动光效 */
@keyframes active-glow {
  0%, 100% {
    filter: drop-shadow(0 0 8px currentColor);
  }
  50% {
    filter: drop-shadow(0 0 16px currentColor) drop-shadow(0 0 24px currentColor);
  }
}

/* 跳线效果 - 当边交叉时增强处理 */
:deep(.vue-flow__edge path) {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  stroke-linecap: round;
  stroke-linejoin: round;
}

:deep(.vue-flow__edge:hover path) {
  stroke-width: 6px !important;
  filter: drop-shadow(0 4px 12px rgba(0,0,0,0.4)) drop-shadow(0 0 8px currentColor);
  animation: jump-line 0.6s ease-in-out;
  z-index: 100;
}

/* 交叉边的半透明处理 */
:deep(.vue-flow__edge.crossing path) {
  opacity: 0.7;
  stroke-dasharray: 6,3;
}

:deep(.vue-flow__edge.crossing:hover path) {
  opacity: 1;
  stroke-dasharray: none;
  animation: active-glow 1.5s infinite;
}

/* 智能避让样式 */
:deep(.vue-flow__edge.avoid-crossing path) {
  stroke-dasharray: 4,2;
  opacity: 0.8;
}

:deep(.vue-flow__edge.avoid-crossing:hover path) {
  stroke-dasharray: none;
  opacity: 1;
  transform: translateY(-2px);
}

/* 边的箭头样式增强 - 立体渐变效果和边缘检测 */
:deep(.vue-flow__edge .vue-flow__edge-path) {
  stroke-linecap: round;
  stroke-linejoin: round;
  /* 确保箭头不被节点遮挡 */
  marker-start: none;
  marker-mid: none;
}

/* 箭头标记增强 - 优化尺寸和可见性 */
:deep(.vue-flow__edge-marker) {
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
  /* 确保箭头始终可见 */
  overflow: visible;
  z-index: 10;
}

/* 箭头悬停放大效果 */
:deep(.vue-flow__edge:hover .vue-flow__edge-marker) {
  transform: scale(1.3);
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.4)) drop-shadow(0 0 12px currentColor);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 动态箭头效果 - 流向识别增强 */
:deep(.vue-flow__edge.animated .vue-flow__edge-marker) {
  animation: arrow-pulse 2s infinite, arrow-flow 3s linear infinite;
}

@keyframes arrow-pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.1);
  }
}

@keyframes arrow-flow {
  0% {
    filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
  }
  50% {
    filter: drop-shadow(0 4px 8px rgba(0,0,0,0.4)) drop-shadow(0 0 8px currentColor);
  }
  100% {
    filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
  }
}

/* 不同优先级的箭头样式 - 增强版 */
:deep(.vue-flow__edge[data-priority="critical"] .vue-flow__edge-marker) {
  filter: drop-shadow(0 0 8px currentColor) drop-shadow(0 2px 6px rgba(0,0,0,0.4));
  animation: critical-pulse 1s infinite;
  transform-origin: center;
}

:deep(.vue-flow__edge[data-priority="high"] .vue-flow__edge-marker) {
  filter: drop-shadow(0 0 6px currentColor) drop-shadow(0 2px 4px rgba(0,0,0,0.3));
  animation: high-priority-glow 2s infinite;
}

/* 流量状态颜色区分 - 箭头颜色动态变化 */
:deep(.vue-flow__edge[data-flow-type="forward"].active .vue-flow__edge-marker) {
  fill: #FF5722;
  stroke: #D84315;
  animation: forward-flow-pulse 1.5s infinite;
}

:deep(.vue-flow__edge[data-flow-type="input"].active .vue-flow__edge-marker) {
  fill: #4CAF50;
  stroke: #388E3C;
  animation: input-flow-pulse 1.5s infinite;
}

:deep(.vue-flow__edge[data-flow-type="output"].active .vue-flow__edge-marker) {
  fill: #2196F3;
  stroke: #1976D2;
  animation: output-flow-pulse 1.5s infinite;
}

/* 智能避让 - 箭头位置调整 */
:deep(.vue-flow__edge.avoid-overlap .vue-flow__edge-marker) {
  transform: translateX(-8px); /* 向后偏移，避免与节点重叠 */
}

:deep(.vue-flow__edge.reverse-direction .vue-flow__edge-marker) {
  transform: rotate(180deg) translateX(8px); /* 反向箭头 */
}

/* 鼠标悬停详细信息显示 */
:deep(.vue-flow__edge:hover .vue-flow__edge-marker)::after {
  content: attr(data-info);
  position: absolute;
  top: -30px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  white-space: nowrap;
  pointer-events: none;
  z-index: 1000;
}

/* 关键优先级脉冲动画 */
@keyframes critical-pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

/* 高优先级发光动画 */
@keyframes high-priority-glow {
  0%, 100% {
    filter: drop-shadow(0 0 6px currentColor) drop-shadow(0 2px 4px rgba(0,0,0,0.3));
  }
  50% {
    filter: drop-shadow(0 0 12px currentColor) drop-shadow(0 4px 8px rgba(0,0,0,0.4));
  }
}

/* 转发流量脉冲动画 */
@keyframes forward-flow-pulse {
  0%, 100% {
    fill: #FF5722;
    stroke: #D84315;
    filter: drop-shadow(0 0 4px #FF5722);
  }
  50% {
    fill: #FF7043;
    stroke: #BF360C;
    filter: drop-shadow(0 0 12px #FF5722) drop-shadow(0 0 20px #FF5722);
  }
}

/* 输入流量脉冲动画 */
@keyframes input-flow-pulse {
  0%, 100% {
    fill: #4CAF50;
    stroke: #388E3C;
    filter: drop-shadow(0 0 4px #4CAF50);
  }
  50% {
    fill: #66BB6A;
    stroke: #2E7D32;
    filter: drop-shadow(0 0 12px #4CAF50) drop-shadow(0 0 20px #4CAF50);
  }
}

/* 输出流量脉冲动画 */
@keyframes output-flow-pulse {
  0%, 100% {
    fill: #2196F3;
    stroke: #1976D2;
    filter: drop-shadow(0 0 4px #2196F3);
  }
  50% {
    fill: #42A5F5;
    stroke: #1565C0;
    filter: drop-shadow(0 0 12px #2196F3) drop-shadow(0 0 20px #2196F3);
  }
}

/* 立体箭头效果 */
:deep(.vue-flow__edge-marker path) {
  stroke-width: 1;
  stroke: rgba(255, 255, 255, 0.8);
  fill-opacity: 0.9;
}

/* 渐变箭头 - 根据流类型 */
:deep(.vue-flow__edge[data-flow-type="forward"] .vue-flow__edge-marker) {
  fill: url(#forward-gradient);
}

:deep(.vue-flow__edge[data-flow-type="input"] .vue-flow__edge-marker) {
  fill: url(#input-gradient);
}

:deep(.vue-flow__edge[data-flow-type="output"] .vue-flow__edge-marker) {
  fill: url(#output-gradient);
}

/* 连接点样式 */
:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  border: 3px solid white;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  transition: all 0.3s ease;
}

:deep(.vue-flow__handle:hover) {
  transform: scale(1.3);
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

/* 状态指示器样式 */
.interface-status {
  position: absolute;
  top: -8px;
  right: -8px;
}

.status-indicator {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #67C23A;
  border: 2px solid white;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  animation: pulse-status 2s infinite;
}

@keyframes pulse-status {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.8;
  }
}

/* 进程活动点样式 */
.process-activity {
  position: absolute;
  bottom: 8px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 4px;
}

.activity-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.8);
  animation: activity-pulse 1.5s infinite;
}

.activity-dot:nth-child(2) {
  animation-delay: 0.3s;
}

.activity-dot:nth-child(3) {
  animation-delay: 0.6s;
}

@keyframes activity-pulse {
  0%, 100% {
    opacity: 0.3;
    transform: scale(0.8);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

/* 链节点统计样式增强 */
:deep(.chain-stats) {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 11px;
  color: #666;
  margin-top: 4px;
}

:deep(.chain-stats i) {
  font-size: 12px;
  color: #409EFF;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .topology-sidebar {
    width: 280px;
  }
}

@media (max-width: 768px) {
  .topology-sidebar {
    width: 260px;
  }
  
  .topology-controls {
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .topology-header {
    padding: 15px;
  }
  
  .topology-header h2 {
    font-size: 20px;
  }
}

@media (max-width: 480px) {
  .topology-sidebar {
    width: 240px;
  }
  
  .topology-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .topology-controls {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>