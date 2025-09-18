<template>
  <div class="topology-page">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <div class="operation-bar">
        <div class="operation-left">
          <el-button type="success" @click="refreshTopology">
            <el-icon><Refresh /></el-icon>
            刷新拓扑
          </el-button>
          <el-button type="info" @click="resetView">
            <el-icon><Aim /></el-icon>
            重置视图
          </el-button>
        </div>
        <div class="operation-right">
          <el-select v-model="selectedChain" placeholder="选择链" @change="filterByChain">
            <el-option label="全部" value="" />
            <el-option label="INPUT" value="INPUT" />
            <el-option label="OUTPUT" value="OUTPUT" />
            <el-option label="FORWARD" value="FORWARD" />
          </el-select>
        </div>
      </div>
    </el-card>

    <!-- 拓扑图 -->
    <el-card class="topology-card">
      <div ref="topologyRef" class="topology-container" v-loading="loading">
        <!-- 网络拓扑图将在这里渲染 -->
      </div>
    </el-card>

    <!-- 规则详情面板 -->
    <el-card v-if="selectedRule" class="rule-detail-card">
      <template #header>
        <div class="card-header">
          <span>规则详情</span>
          <el-button type="text" @click="selectedRule = null">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="规则ID">{{ selectedRule.id }}</el-descriptions-item>
        <el-descriptions-item label="链名">
          <el-tag :type="getChainTagType(selectedRule.chain_name)">
            {{ selectedRule.chain_name }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="目标">
          <el-tag :type="getTargetTagType(selectedRule.target)">
            {{ selectedRule.target }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="协议">{{ selectedRule.protocol || '全部' }}</el-descriptions-item>
        <el-descriptions-item label="源IP">{{ selectedRule.source_ip || '任意' }}</el-descriptions-item>
        <el-descriptions-item label="目标IP">{{ selectedRule.destination_ip || '任意' }}</el-descriptions-item>
        <el-descriptions-item label="源端口">{{ selectedRule.source_port || '任意' }}</el-descriptions-item>
        <el-descriptions-item label="目标端口">{{ selectedRule.destination_port || '任意' }}</el-descriptions-item>
        <el-descriptions-item label="规则文本" :span="2">
          <el-input
            :value="selectedRule.rule_text"
            type="textarea"
            :rows="2"
            readonly
          />
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts/core'
import { GraphChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { apiService, type IPTablesRule } from '../api'

echarts.use([
  GraphChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer
])

const loading = ref(false)
const topologyRef = ref<HTMLElement>()
const selectedChain = ref('')
const selectedRule = ref<IPTablesRule | null>(null)
const rules = ref<IPTablesRule[]>([])

let topologyChart: echarts.ECharts | null = null

const getChainTagType = (chainName: string) => {
  const types: Record<string, string> = {
    'INPUT': 'success',
    'OUTPUT': 'warning',
    'FORWARD': 'info'
  }
  return types[chainName] || 'default'
}

const getTargetTagType = (target: string) => {
  const types: Record<string, string> = {
    'ACCEPT': 'success',
    'DROP': 'danger',
    'REJECT': 'warning'
  }
  return types[target] || 'default'
}

const loadRules = async () => {
  loading.value = true
  try {
    const response = await apiService.getRules()
    rules.value = response.data
    await nextTick()
    initTopology()
  } catch (error) {
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

const initTopology = () => {
  if (!topologyRef.value) return
  
  topologyChart = echarts.init(topologyRef.value)
  
  // 生成拓扑数据
  const nodes = generateNodes()
  const links = generateLinks()
  
  const option = {
    title: {
      text: 'IPTables 网络拓扑图',
      left: 'center',
      textStyle: {
        fontSize: 18,
        fontWeight: 'bold'
      }
    },
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        if (params.dataType === 'node') {
          return `<strong>${params.data.name}</strong><br/>类型: ${params.data.category}<br/>描述: ${params.data.description}`
        } else if (params.dataType === 'edge') {
          return `<strong>规则连接</strong><br/>规则: ${params.data.rule}<br/>动作: ${params.data.target}`
        }
        return ''
      }
    },
    legend: {
      data: ['网络接口', 'IPTables链', '规则节点'],
      bottom: 10
    },
    series: [
      {
        name: '网络拓扑',
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: links,
        categories: [
          { name: '网络接口', itemStyle: { color: '#91cc75' } },
          { name: 'IPTables链', itemStyle: { color: '#fac858' } },
          { name: '规则节点', itemStyle: { color: '#ee6666' } }
        ],
        roam: true,
        label: {
          show: true,
          position: 'right',
          formatter: '{b}'
        },
        labelLayout: {
          hideOverlap: true
        },
        scaleLimit: {
          min: 0.4,
          max: 2
        },
        lineStyle: {
          color: 'source',
          curveness: 0.3
        },
        emphasis: {
          focus: 'adjacency',
          lineStyle: {
            width: 10
          }
        },
        force: {
          repulsion: 1000,
          gravity: 0.1,
          edgeLength: [50, 200],
          layoutAnimation: true
        }
      }
    ]
  }
  
  topologyChart.setOption(option)
  
  // 添加点击事件
  topologyChart.on('click', (params) => {
    if (params.dataType === 'node' && params.data.ruleData) {
      selectedRule.value = params.data.ruleData
    }
  })
}

const generateNodes = () => {
  const nodes = []
  
  // 添加网络接口节点
  nodes.push(
    { id: 'internet', name: 'Internet', category: 0, description: '外部网络' },
    { id: 'localhost', name: 'Localhost', category: 0, description: '本地主机' },
    { id: 'lan', name: 'LAN', category: 0, description: '局域网' }
  )
  
  // 添加IPTables链节点
  nodes.push(
    { id: 'input', name: 'INPUT', category: 1, description: '输入链' },
    { id: 'output', name: 'OUTPUT', category: 1, description: '输出链' },
    { id: 'forward', name: 'FORWARD', category: 1, description: '转发链' }
  )
  
  // 添加规则节点
  const filteredRules = selectedChain.value 
    ? rules.value.filter(rule => rule.chain_name === selectedChain.value)
    : rules.value.slice(0, 10) // 限制显示数量以避免过于复杂
  
  filteredRules.forEach(rule => {
    nodes.push({
      id: `rule_${rule.id}`,
      name: `规则${rule.id}`,
      category: 2,
      description: `${rule.target} ${rule.protocol || 'ALL'} ${rule.destination_port || ''}`,
      ruleData: rule
    })
  })
  
  return nodes
}

const generateLinks = () => {
  const links = []
  
  // 基础网络连接
  links.push(
    { source: 'internet', target: 'input', rule: '外部流量', target: 'INPUT' },
    { source: 'output', target: 'internet', rule: '输出流量', target: 'OUTPUT' },
    { source: 'lan', target: 'forward', rule: '转发流量', target: 'FORWARD' },
    { source: 'localhost', target: 'input', rule: '本地输入', target: 'INPUT' },
    { source: 'output', target: 'localhost', rule: '本地输出', target: 'OUTPUT' }
  )
  
  // 规则连接
  const filteredRules = selectedChain.value 
    ? rules.value.filter(rule => rule.chain_name === selectedChain.value)
    : rules.value.slice(0, 10)
  
  filteredRules.forEach(rule => {
    const chainId = rule.chain_name.toLowerCase()
    links.push({
      source: chainId,
      target: `rule_${rule.id}`,
      rule: rule.rule_text || `${rule.target} 规则`,
      target: rule.target
    })
  })
  
  return links
}

const refreshTopology = () => {
  loadRules()
}

const resetView = () => {
  if (topologyChart) {
    topologyChart.dispatchAction({
      type: 'restore'
    })
  }
}

const filterByChain = () => {
  initTopology()
}

onMounted(() => {
  loadRules()
  
  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    topologyChart?.resize()
  })
})
</script>

<style scoped>
.topology-page {
  padding: 0;
}

.operation-card {
  margin-bottom: 20px;
}

.operation-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.operation-left {
  display: flex;
  gap: 12px;
}

.topology-card {
  margin-bottom: 20px;
}

.topology-container {
  height: 600px;
  width: 100%;
}

.rule-detail-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

@media (max-width: 768px) {
  .operation-bar {
    flex-direction: column;
    gap: 16px;
  }
  
  .topology-container {
    height: 400px;
  }
}
</style>