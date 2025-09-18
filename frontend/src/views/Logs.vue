<template>
  <div class="logs-page">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <div class="operation-bar">
        <div class="operation-left">
          <el-button type="success" @click="refreshLogs">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
          <el-button type="warning" @click="exportLogs">
            <el-icon><Download /></el-icon>
            导出日志
          </el-button>
        </div>
        <div class="operation-right">
          <el-input
            v-model="searchText"
            placeholder="搜索日志..."
            style="width: 300px"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </div>
    </el-card>

    <!-- 日志表格 -->
    <el-card class="table-card">
      <el-table
        :data="filteredLogs"
        v-loading="loading"
        stripe
        border
        style="width: 100%"
        :default-sort="{ prop: 'timestamp', order: 'descending' }"
      >
        <el-table-column prop="id" label="ID" width="80" sortable />
        <el-table-column prop="username" label="用户" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.username }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operation" label="操作" width="150">
          <template #default="{ row }">
            <el-tag :type="getOperationTagType(row.operation)">
              {{ row.operation }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="details" label="详情" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.details" placement="top">
              <span class="details-text">{{ row.details }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="150" />
        <el-table-column prop="timestamp" label="时间" width="180" sortable>
          <template #default="{ row }">
            {{ formatDate(row.timestamp) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="totalLogs"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { apiService, type OperationLog } from '../api'

const loading = ref(false)
const searchText = ref('')
const logs = ref<OperationLog[]>([])
const currentPage = ref(1)
const pageSize = ref(20)

const filteredLogs = computed(() => {
  let filtered = logs.value
  
  if (searchText.value) {
    filtered = filtered.filter(log => 
      log.username.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.operation.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.details.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.ip_address.toLowerCase().includes(searchText.value.toLowerCase())
    )
  }
  
  // 分页
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filtered.slice(start, end)
})

const totalLogs = computed(() => {
  if (searchText.value) {
    return logs.value.filter(log => 
      log.username.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.operation.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.details.toLowerCase().includes(searchText.value.toLowerCase()) ||
      log.ip_address.toLowerCase().includes(searchText.value.toLowerCase())
    ).length
  }
  return logs.value.length
})

const getOperationTagType = (operation: string) => {
  if (operation.includes('登录')) return 'success'
  if (operation.includes('创建') || operation.includes('添加')) return 'primary'
  if (operation.includes('更新') || operation.includes('编辑')) return 'warning'
  if (operation.includes('删除')) return 'danger'
  return 'info'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const loadLogs = async () => {
  loading.value = true
  try {
    const response = await apiService.getLogs()
    logs.value = response.data
  } catch (error) {
    ElMessage.error('加载日志失败')
  } finally {
    loading.value = false
  }
}

const refreshLogs = () => {
  loadLogs()
}

const exportLogs = () => {
  // 模拟导出功能
  const csvContent = [
    ['ID', '用户', '操作', '详情', 'IP地址', '时间'].join(','),
    ...logs.value.map(log => [
      log.id,
      log.username,
      log.operation,
      `"${log.details}"`,
      log.ip_address,
      formatDate(log.timestamp)
    ].join(','))
  ].join('\n')
  
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `operation_logs_${new Date().toISOString().split('T')[0]}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  ElMessage.success('日志导出成功')
}

const handleSearch = () => {
  currentPage.value = 1
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  currentPage.value = 1
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.logs-page {
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

.table-card {
  margin-bottom: 20px;
}

.details-text {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination-container {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

@media (max-width: 768px) {
  .operation-bar {
    flex-direction: column;
    gap: 16px;
  }
  
  .operation-right {
    width: 100%;
  }
  
  .operation-right .el-input {
    width: 100% !important;
  }
}
</style>