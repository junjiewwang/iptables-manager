import { ref } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { apiService, rulesAPI } from '@/api'
import type { RuleForm, IPTablesRule, CleanupResult } from '@/types/ChainTable/types'

/**
 * 表格操作逻辑组合式函数
 * 处理所有与表格操作相关的状态和方法
 */
export function useTableActions() {
  // 规则管理相关状态
  const rulesLoading = ref(false)
  const ruleDialogVisible = ref(false)
  const isEditRule = ref(false)
  const rules = ref<IPTablesRule[]>([])
  const ruleFormRef = ref<FormInstance>()
  
  // 清除无效规则相关状态
  const cleaningRules = ref(false)
  const showCleanupDialog = ref(false)
  const cleanupResult = ref<CleanupResult | null>(null)
  
  // 规则表单数据
  const ruleForm = ref<RuleForm>({
    id: undefined,
    chain_name: '',
    table: '',
    target: '',
    protocol: 'all',
    source_ip: '',
    destination_ip: '',
    source_port: '',
    destination_port: '',
    in_interface: '',
    out_interface: '',
    options: ''
  })
  
  // 规则表单验证规则
  const ruleFormRules = {
    chain_name: [
      { required: true, message: '请选择链名', trigger: 'change' }
    ],
    table: [
      { required: true, message: '请选择表', trigger: 'change' }
    ],
    target: [
      { required: true, message: '请选择目标', trigger: 'change' }
    ],
    protocol: [
      { required: true, message: '请选择协议', trigger: 'change' }
    ]
  }
  
  // 加载规则
  const loadRules = async () => {
    rulesLoading.value = true
    try {
      const response = await apiService.getRules()
      rules.value = response.data
      return response.data
    } catch (error) {
      console.error('Failed to load rules:', error)
      ElMessage.error('加载规则失败')
      return []
    } finally {
      rulesLoading.value = false
    }
  }
  
  // 打开添加规则对话框
  const openAddRuleDialog = (chainName?: string) => {
    isEditRule.value = false
    ruleDialogVisible.value = true
    resetRuleForm()
    
    // 如果指定了链名，自动设置
    if (chainName) {
      ruleForm.value.chain_name = chainName
      // 触发链变化处理
      handleChainChange()
    }
  }
  
  // 编辑规则
  const editRule = (rule: IPTablesRule) => {
    isEditRule.value = true
    ruleDialogVisible.value = true
    
    // 将规则数据转换为表单数据
    ruleForm.value = {
      id: rule.id,
      chain_name: rule.chain_name || '',
      table: rule.table || '',
      target: rule.target || '',
      protocol: rule.protocol || 'all',
      source_ip: rule.source_ip || rule.source || '',
      destination_ip: rule.destination_ip || rule.destination || '',
      source_port: rule.source_port || '',
      destination_port: rule.destination_port || '',
      in_interface: rule.in_interface || '',
      out_interface: rule.out_interface || '',
      options: rule.options || ''
    }
  }
  
  // 删除规则
  const deleteRule = async (rule: IPTablesRule) => {
    try {
      await ElMessageBox.confirm('确定要删除这条规则吗？', '确认删除', {
        type: 'warning'
      })
      
      if (rule.id) {
        await apiService.deleteRule(rule.id)
        ElMessage.success('删除成功')
        await loadRules()
        return true
      } else {
        ElMessage.warning('无法删除：规则ID不存在')
        return false
      }
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
      return false
    }
  }
  
  // 提交规则表单
  const submitRuleForm = async () => {
    if (!ruleFormRef.value) return false
    
    try {
      await ruleFormRef.value.validate()
      
      if (isEditRule.value) {
        await apiService.updateRule(ruleForm.value.id!, ruleForm.value)
        ElMessage.success('更新成功')
      } else {
        await apiService.addRule(ruleForm.value)
        ElMessage.success('添加成功')
      }
      
      ruleDialogVisible.value = false
      await loadRules()
      return true
    } catch (error) {
      ElMessage.error(isEditRule.value ? '更新失败' : '添加失败')
      return false
    }
  }
  
  // 重置规则表单
  const resetRuleForm = () => {
    Object.assign(ruleForm.value, {
      id: undefined,
      chain_name: '',
      table: '',
      target: '',
      protocol: 'all',
      source_ip: '',
      destination_ip: '',
      source_port: '',
      destination_port: '',
      in_interface: '',
      out_interface: '',
      options: ''
    })
    ruleFormRef.value?.clearValidate()
  }
  
  // 处理链变化
  const handleChainChange = () => {
    // 如果当前选择的表不在可用表中，清空
    if (ruleForm.value.table && !getAvailableTables(ruleForm.value.chain_name).includes(ruleForm.value.table)) {
      ruleForm.value.table = ''
    }
  }
  
  // 获取可用表
  const getAvailableTables = (chainName: string): string[] => {
    if (!chainName) return []
    
    const chainTableMap: Record<string, string[]> = {
      'PREROUTING': ['raw', 'mangle', 'nat'],
      'INPUT': ['mangle', 'filter', 'nat'],
      'FORWARD': ['mangle', 'filter'],
      'OUTPUT': ['raw', 'mangle', 'nat', 'filter'],
      'POSTROUTING': ['mangle', 'nat']
    }
    
    return chainTableMap[chainName] || []
  }
  
  // 清除无效规则
  const cleanInvalidRules = async () => {
    try {
      // 首先进行dry-run检查
      cleaningRules.value = true
      const dryRunResponse = await rulesAPI.cleanInvalidRules(true)
      
      if (dryRunResponse.data.total_cleaned === 0) {
        ElMessage.info('未发现无效规则')
        return false
      }

      // 显示确认对话框
      const confirmResult = await ElMessageBox.confirm(
        `发现 ${dryRunResponse.data.total_cleaned} 条无效规则：\n` +
        `• 重复规则: ${dryRunResponse.data.duplicate_rules} 条\n` +
        `• 无效网桥规则: ${dryRunResponse.data.invalid_bridges} 条\n` +
        `• 无效链规则: ${dryRunResponse.data.invalid_chains} 条\n` +
        `• 无效目标规则: ${dryRunResponse.data.invalid_targets} 条\n\n` +
        `确定要清除这些无效规则吗？`,
        '清除无效规则确认',
        {
          confirmButtonText: '确定清除',
          cancelButtonText: '取消',
          type: 'warning',
          dangerouslyUseHTMLString: true
        }
      )

      if (confirmResult === 'confirm') {
        // 执行实际清除
        const cleanResponse = await rulesAPI.cleanInvalidRules(false)
        cleanupResult.value = cleanResponse.data

        // 显示清除结果
        ElNotification({
          title: '清除完成',
          message: `成功清除 ${cleanResponse.data.total_cleaned} 条无效规则`,
          type: 'success',
          duration: 5000
        })
        
        return true
      }
      
      return false
    } catch (error) {
      console.error('清除无效规则失败:', error)
      ElMessage.error('清除无效规则失败: ' + (error.response?.data?.error || error.message || '未知错误'))
      return false
    } finally {
      cleaningRules.value = false
    }
  }
  
  // 辅助函数：从规则文本中提取目标
  const extractTarget = (ruleText: string): string => {
    const targetMatch = ruleText.match(/-j\s+(\w+)/)
    if (targetMatch) return targetMatch[1]
    
    // 检查常见的目标关键词
    if (ruleText.includes('ACCEPT')) return 'ACCEPT'
    if (ruleText.includes('DROP')) return 'DROP'
    if (ruleText.includes('REJECT')) return 'REJECT'
    if (ruleText.includes('RETURN')) return 'RETURN'
    if (ruleText.includes('MASQUERADE')) return 'MASQUERADE'
    if (ruleText.includes('SNAT')) return 'SNAT'
    if (ruleText.includes('DNAT')) return 'DNAT'
    
    return '-'
  }

  // 辅助函数：从规则文本中提取协议
  const extractProtocol = (ruleText: string): string => {
    const protocolMatch = ruleText.match(/-p\s+(\w+)/)
    return protocolMatch ? protocolMatch[1] : 'all'
  }
  
  return {
    // 状态
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
    
    // 方法
    loadRules,
    openAddRuleDialog,
    editRule,
    deleteRule,
    submitRuleForm,
    resetRuleForm,
    handleChainChange,
    getAvailableTables,
    cleanInvalidRules,
    extractTarget,
    extractProtocol
  }
}