/**
 * 格式化工具 Composable
 * 提供各种数据格式化功能
 */

/**
 * 格式化工具 Hook
 */
export function useFormatters() {
  /**
   * 格式化字节数
   * @param bytes 字节数
   * @returns 格式化后的字符串
   */
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  /**
   * 格式化数字，添加千分位分隔符
   * @param num 数字
   * @returns 格式化后的字符串
   */
  const formatNumber = (num: number): string => {
    return num.toLocaleString()
  }

  /**
   * 格式化IP地址，处理空值情况
   * @param ip IP地址
   * @returns 格式化后的IP地址
   */
  const formatIpAddress = (ip?: string): string => {
    return ip || '任意'
  }

  /**
   * 格式化端口，处理空值情况
   * @param port 端口
   * @returns 格式化后的端口
   */
  const formatPort = (port?: string): string => {
    return port || '任意'
  }

  /**
   * 格式化接口名称，处理空值情况
   * @param interfaceName 接口名称
   * @returns 格式化后的接口名称
   */
  const formatInterface = (interfaceName?: string): string => {
    return interfaceName || '任意'
  }

  /**
   * 截断长文本
   * @param text 文本
   * @param maxLength 最大长度
   * @returns 截断后的文本
   */
  const truncateText = (text: string, maxLength: number = 50): string => {
    if (text.length <= maxLength) return text
    return text.substring(0, maxLength) + '...'
  }

  /**
   * 格式化日期
   * @param dateString 日期字符串
   * @returns 格式化后的日期字符串
   */
  const formatDate = (dateString: string): string => {
    if (!dateString) return '-'
    return new Date(dateString).toLocaleString('zh-CN')
  }

  return {
    formatBytes,
    formatNumber,
    formatIpAddress,
    formatPort,
    formatInterface,
    truncateText,
    formatDate
  }
}