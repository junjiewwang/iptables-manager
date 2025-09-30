import { describe, it, expect, vi } from 'vitest'
import { useErrorHandler, ErrorType } from '../useErrorHandler'
import { ElMessage } from 'element-plus'

// Mock Element Plus
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn()
  }
}))

describe('useErrorHandler', () => {
  const { handleError, wrapAsync } = useErrorHandler()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('handleError', () => {
    it('应该正确处理API错误', () => {
      const error = {
        isAxiosError: true,
        response: {
          data: {
            message: 'API错误'
          }
        }
      }

      const result = handleError(error, '测试')
      
      expect(result.success).toBe(false)
      expect(result.message).toBe('API错误')
      expect(result.type).toBe(ErrorType.API)
      expect(ElMessage.error).toHaveBeenCalledWith('API错误')
    })

    it('应该正确处理验证错误', () => {
      const error = {
        name: 'ValidationError',
        message: '验证失败'
      }

      const result = handleError(error, '测试')
      
      expect(result.success).toBe(false)
      expect(result.message).toBe('测试验证失败: 验证失败')
      expect(result.type).toBe(ErrorType.VALIDATION)
      expect(ElMessage.error).toHaveBeenCalledWith('测试验证失败: 验证失败')
    })

    it('应该正确处理网络错误', () => {
      const error = {
        message: 'network error'
      }

      const result = handleError(error, '测试')
      
      expect(result.success).toBe(false)
      expect(result.message).toBe('测试网络错误: network error')
      expect(result.type).toBe(ErrorType.NETWORK)
      expect(ElMessage.error).toHaveBeenCalledWith('测试网络错误: network error')
    })

    it('应该正确处理未知错误', () => {
      const error = {
        message: '未知错误'
      }

      const result = handleError(error, '测试')
      
      expect(result.success).toBe(false)
      expect(result.message).toBe('测试错误: 未知错误')
      expect(result.type).toBe(ErrorType.UNKNOWN)
      expect(ElMessage.error).toHaveBeenCalledWith('测试错误: 未知错误')
    })

    it('应该尊重配置选项', () => {
      const error = {
        message: '错误'
      }

      // 不显示消息
      handleError(error, '测试', { showMessage: false })
      expect(ElMessage.error).not.toHaveBeenCalled()

      // 不记录到控制台
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      handleError(error, '测试', { logToConsole: false })
      expect(consoleSpy).not.toHaveBeenCalled()
      consoleSpy.mockRestore()

      // 重新抛出错误
      expect(() => {
        handleError(error, '测试', { rethrow: true })
      }).toThrow()
    })
  })

  describe('wrapAsync', () => {
    it('应该正确包装异步函数', async () => {
      const fn = vi.fn().mockResolvedValue('成功')
      
      const result = await wrapAsync(fn, '测试')
      
      expect(result).toBe('成功')
      expect(fn).toHaveBeenCalled()
    })

    it('应该处理异步函数中的错误', async () => {
      const error = new Error('异步错误')
      const fn = vi.fn().mockRejectedValue(error)
      
      const result = await wrapAsync(fn, '测试')
      
      expect(result).toBeNull()
      expect(fn).toHaveBeenCalled()
      expect(ElMessage.error).toHaveBeenCalled()
    })

    it('应该尊重配置选项', async () => {
      const error = new Error('异步错误')
      const fn = vi.fn().mockRejectedValue(error)
      
      // 重新抛出错误
      await expect(wrapAsync(fn, '测试', { rethrow: true })).rejects.toThrow('异步错误')
    })
  })
})