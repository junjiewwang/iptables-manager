<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon total">
              <el-icon><List /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ statistics.total_rules }}</div>
              <div class="stat-label">总规则数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon active">
              <el-icon><Check /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ Object.keys(statistics.rules_by_chain).length }}</div>
              <div class="stat-label">活跃链数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon operations">
              <el-icon><Operation /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ statistics.recent_operations }}</div>
              <div class="stat-label">今日操作</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon status" :class="statusClass">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ statistics.system_status }}</div>
              <div class="stat-label">系统状态</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>规则分布</span>
              <el-button type="text" @click="refreshCharts">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div ref="pieChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>操作趋势</span>
              <el-button type="text" @click="refreshCharts">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div ref="lineChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card class="quick-actions-card">
          <template #header>
            <span>快速操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/rules')">
              <el-icon><Plus /></el-icon>
              添加规则
            </el-button>
            <el-button type="success" @click="$router.push('/topology')">
              <el-icon><Share /></el-icon>
              查看拓扑
            </el-button>
            <el-button type="info" @click="$router.push('/logs')">
              <el-icon><Document /></el-icon>
              查看日志
            </el-button>
            <el-button type="warning" @click="handleBackup">
              <el-icon><Download /></el-icon>
              备份配置
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts/core'
import { PieChart, LineChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { apiService, type Statistics } from '../api'

echarts.use([
  PieChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  CanvasRenderer
])

const pieChartRef = ref<HTMLElement>()
const lineChartRef = ref<HTMLElement>()

const statistics = reactive<Statistics>({
  total_rules: 0,
  rules_by_chain: {},
  recent_operations: 0,
  system_status: '正常'
})

const statusClass = computed(() => {
  return statistics.system_status === '正常' ? 'normal' : 'error'
})

let pieChart: echarts.ECharts | null = null
let lineChart: echarts.ECharts | null = null

const loadStatistics = async () => {
  try {
    const response = await apiService.getStatistics()
    Object.assign(statistics, response.data)
    await nextTick()
    initCharts()
  } catch (error) {
    ElMessage.error('加载统计数据失败')
    console.error('Failed to load statistics:', error)
  }
}

const initCharts = () => {
  initPieChart()
  initLineChart()
}

const initPieChart = () => {
  if (!pieChartRef.value) return
  
  pieChart = echarts.init(pieChartRef.value)
  
  const data = Object.entries(statistics.rules_by_chain).map(([name, value]) => ({
    name,
    value
  }))
  
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '规则分布',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data
      }
    ]
  }
  
  pieChart.setOption(option)
}

const initLineChart = () => {
  if (!lineChartRef.value) return
  
  lineChart = echarts.init(lineChartRef.value)
  
  // 模拟操作趋势数据
  const days = []
  const operations = []
  for (let i = 6; i >= 0; i--) {
    const date = new Date()
    date.setDate(date.getDate() - i)
    days.push(date.toLocaleDateString())
    operations.push(Math.floor(Math.random() * 20) + 5)
  }
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: days
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '操作次数',
        type: 'line',
        stack: 'Total',
        smooth: true,
        areaStyle: {
          opacity: 0.3
        },
        data: operations
      }
    ]
  }
  
  lineChart.setOption(option)
}

const refreshCharts = () => {
  loadStatistics()
}

const handleBackup = async () => {
  try {
    await apiService.backup()
    ElMessage.success('备份成功')
  } catch (error) {
    ElMessage.error('备份失败')
  }
}

onMounted(() => {
  loadStatistics()
  
  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    pieChart?.resize()
    lineChart?.resize()
  })
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.active {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.operations {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.status.normal {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-icon.status.error {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.charts-row {
  margin-bottom: 20px;
}

.chart-card {
  height: 400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 320px;
}

.quick-actions-card {
  margin-bottom: 20px;
}

.quick-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.quick-actions .el-button {
  flex: 1;
  min-width: 120px;
}

@media (max-width: 768px) {
  .quick-actions .el-button {
    flex: none;
    width: 100%;
  }
}
</style>