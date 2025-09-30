<template>
  <el-dialog
    :title="isEditRule ? '编辑规则' : '添加规则'"
    v-model="dialogVisible"
    width="600px"
    :close-on-click-modal="false"
    @closed="handleDialogClosed"
  >
    <el-form
      ref="ruleFormRef"
      :model="ruleForm"
      :rules="ruleFormRules"
      label-width="100px"
      label-position="right"
      size="default"
    >
      <el-form-item label="链名" prop="chain_name">
        <el-select v-model="ruleForm.chain_name" placeholder="请选择链" @change="handleChainChange" style="width: 100%">
          <el-option label="PREROUTING" value="PREROUTING" />
          <el-option label="INPUT" value="INPUT" />
          <el-option label="FORWARD" value="FORWARD" />
          <el-option label="OUTPUT" value="OUTPUT" />
          <el-option label="POSTROUTING" value="POSTROUTING" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="表" prop="table">
        <el-select v-model="ruleForm.table" placeholder="请选择表" style="width: 100%">
          <el-option
            v-for="table in availableTables"
            :key="table"
            :label="table"
            :value="table"
          />
        </el-select>
      </el-form-item>
      
      <el-form-item label="目标" prop="target">
        <el-select v-model="ruleForm.target" placeholder="请选择目标" style="width: 100%">
          <el-option label="ACCEPT" value="ACCEPT" />
          <el-option label="DROP" value="DROP" />
          <el-option label="REJECT" value="REJECT" />
          <el-option label="RETURN" value="RETURN" />
          <el-option label="MASQUERADE" value="MASQUERADE" />
          <el-option label="SNAT" value="SNAT" />
          <el-option label="DNAT" value="DNAT" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="协议" prop="protocol">
        <el-select v-model="ruleForm.protocol" placeholder="请选择协议" style="width: 100%">
          <el-option label="all" value="all" />
          <el-option label="tcp" value="tcp" />
          <el-option label="udp" value="udp" />
          <el-option label="icmp" value="icmp" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="源IP" prop="source_ip">
        <el-input v-model="ruleForm.source_ip" placeholder="例如: 192.168.1.0/24" />
      </el-form-item>
      
      <el-form-item label="目标IP" prop="destination_ip">
        <el-input v-model="ruleForm.destination_ip" placeholder="例如: 10.0.0.1" />
      </el-form-item>
      
      <el-form-item label="源端口" prop="source_port">
        <el-input v-model="ruleForm.source_port" placeholder="例如: 1024:65535" />
      </el-form-item>
      
      <el-form-item label="目标端口" prop="destination_port">
        <el-input v-model="ruleForm.destination_port" placeholder="例如: 80,443" />
      </el-form-item>
      
      <el-form-item label="入接口" prop="in_interface">
        <el-select v-model="ruleForm.in_interface" placeholder="请选择入接口" clearable style="width: 100%">
          <el-option
            v-for="iface in interfaces"
            :key="iface.name"
            :label="iface.name"
            :value="iface.name"
          />
        </el-select>
      </el-form-item>
      
      <el-form-item label="出接口" prop="out_interface">
        <el-select v-model="ruleForm.out_interface" placeholder="请选择出接口" clearable style="width: 100%">
          <el-option
            v-for="iface in interfaces"
            :key="iface.name"
            :label="iface.name"
            :value="iface.name"
          />
        </el-select>
      </el-form-item>
      
      <el-form-item label="其他选项" prop="options">
        <el-input
          v-model="ruleForm.options"
          type="textarea"
          :rows="3"
          placeholder="其他iptables选项，例如: --log-level debug"
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">
          {{ isEditRule ? '更新' : '添加' }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { FormInstance } from 'element-plus'
import type { RuleForm, NetworkInterface } from '../types'

// 定义组件属性
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  isEditRule: {
    type: Boolean,
    default: false
  },
  ruleForm: {
    type: Object as () => RuleForm,
    required: true
  },
  ruleFormRules: {
    type: Object,
    required: true
  },
  interfaces: {
    type: Array as () => NetworkInterface[],
    default: () => []
  }
})

// 定义事件
const emit = defineEmits([
  'update:visible',
  'submit',
  'chain-change',
  'reset'
])

// 表单引用
const ruleFormRef = ref<FormInstance>()

// 对话框可见性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

// 提交状态
const submitting = ref(false)

// 可用表
const availableTables = computed(() => {
  if (!props.ruleForm.chain_name) return []
  
  const chainTableMap: Record<string, string[]> = {
    'PREROUTING': ['raw', 'mangle', 'nat'],
    'INPUT': ['mangle', 'filter', 'nat'],
    'FORWARD': ['mangle', 'filter'],
    'OUTPUT': ['raw', 'mangle', 'nat', 'filter'],
    'POSTROUTING': ['mangle', 'nat']
  }
  
  return chainTableMap[props.ruleForm.chain_name] || []
})

// 处理链变化
const handleChainChange = () => {
  emit('chain-change')
}

// 提交表单
const submitForm = async () => {
  if (!ruleFormRef.value) return
  
  submitting.value = true
  try {
    await ruleFormRef.value.validate()
    emit('submit', ruleFormRef.value)
  } finally {
    submitting.value = false
  }
}

// 处理对话框关闭
const handleDialogClosed = () => {
  emit('reset')
}

// 暴露方法给父组件
defineExpose({
  ruleFormRef
})
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>