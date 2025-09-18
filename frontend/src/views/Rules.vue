<template>
  <div class="rules-page">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <div class="operation-bar">
        <div class="operation-left">
          <el-button type="primary" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加规则
          </el-button>
          <el-button type="success" @click="refreshRules">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
        <div class="operation-right">
          <el-input
            v-model="searchText"
            placeholder="搜索规则..."
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

    <!-- 规则表格 -->
    <el-card class="table-card">
      <el-table
        :data="filteredRules"
        v-loading="loading"
        stripe
        border
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" sortable />
        <el-table-column prop="chain_name" label="链名" width="120">
          <template #default="{ row }">
            <el-tag :type="getChainTagType(row.chain_name)">
              {{ row.chain_name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="目标" width="120">
          <template #default="{ row }">
            <el-tag :type="getTargetTagType(row.target)">
              {{ row.target }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="100" />
        <el-table-column prop="source_ip" label="源IP" width="150" />
        <el-table-column prop="destination_ip" label="目标IP" width="150" />
        <el-table-column prop="destination_port" label="目标端口" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180" sortable>
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="editRule(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" size="small" @click="deleteRule(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑规则对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '添加规则'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="formRules"
        label-width="100px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="链名" prop="chain_name">
              <el-select v-model="ruleForm.chain_name" placeholder="选择链">
                <el-option label="INPUT" value="INPUT" />
                <el-option label="OUTPUT" value="OUTPUT" />
                <el-option label="FORWARD" value="FORWARD" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标" prop="target">
              <el-select v-model="ruleForm.target" placeholder="选择目标">
                <el-option label="ACCEPT" value="ACCEPT" />
                <el-option label="DROP" value="DROP" />
                <el-option label="REJECT" value="REJECT" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="协议" prop="protocol">
              <el-select v-model="ruleForm.protocol" placeholder="选择协议">
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
                <el-option label="ICMP" value="icmp" />
                <el-option label="ALL" value="" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标端口" prop="destination_port">
              <el-input v-model="ruleForm.destination_port" placeholder="如: 80, 443, 22" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="源IP" prop="source_ip">
              <el-input v-model="ruleForm.source_ip" placeholder="如: 192.168.1.0/24" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标IP" prop="destination_ip">
              <el-input v-model="ruleForm.destination_ip" placeholder="如: 192.168.1.100" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { apiService, type IPTablesRule } from '../api'

const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const searchText = ref('')
const rules = ref<IPTablesRule[]>([])
const ruleFormRef = ref<FormInstance>()

const ruleForm = reactive<IPTablesRule>({
  chain_name: '',
  target: '',
  protocol: '',
  source_ip: '',
  destination_ip: '',
  destination_port: ''
})

const formRules = {
  chain_name: [
    { required: true, message: '请选择链名', trigger: 'change' }
  ],
  target: [
    { required: true, message: '请选择目标', trigger: 'change' }
  ]
}

const filteredRules = computed(() => {
  if (!searchText.value) return rules.value
  return rules.value.filter(rule => 
    rule.chain_name.toLowerCase().includes(searchText.value.toLowerCase()) ||
    rule.target.toLowerCase().includes(searchText.value.toLowerCase()) ||
    rule.source_ip?.toLowerCase().includes(searchText.value.toLowerCase()) ||
    rule.destination_ip?.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

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

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const loadRules = async () => {
  loading.value = true
  try {
    const response = await apiService.getRules()
    rules.value = response.data
  } catch (error) {
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
  resetForm()
}

const editRule = (rule: IPTablesRule) => {
  isEdit.value = true
  dialogVisible.value = true
  Object.assign(ruleForm, rule)
}

const deleteRule = async (rule: IPTablesRule) => {
  try {
    await ElMessageBox.confirm('确定要删除这条规则吗？', '确认删除', {
      type: 'warning'
    })
    
    await apiService.deleteRule(rule.id!)
    ElMessage.success('删除成功')
    loadRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const submitForm = async () => {
  if (!ruleFormRef.value) return
  
  await ruleFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEdit.value) {
          await apiService.updateRule(ruleForm.id!, ruleForm)
          ElMessage.success('更新成功')
        } else {
          await apiService.addRule(ruleForm)
          ElMessage.success('添加成功')
        }
        dialogVisible.value = false
        loadRules()
      } catch (error) {
        ElMessage.error(isEdit.value ? '更新失败' : '添加失败')
      }
    }
  })
}

const resetForm = () => {
  Object.assign(ruleForm, {
    chain_name: '',
    target: '',
    protocol: '',
    source_ip: '',
    destination_ip: '',
    destination_port: ''
  })
  ruleFormRef.value?.clearValidate()
}

const refreshRules = () => {
  loadRules()
}

const handleSearch = () => {
  // 搜索逻辑已在computed中处理
}

onMounted(() => {
  loadRules()
})
</script>

<style scoped>
.rules-page {
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