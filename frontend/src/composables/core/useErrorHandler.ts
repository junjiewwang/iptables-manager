/**
 * 错误处理 Composable
 * 提供统一的错误处理机制
 * 
 * 该组合式函数提供了一个统一的错误处理机制，包括：
 * - 错误类型分类（API、验证、网络、未知）
 * - 错误消息格式化
 * - 错误日志记录
 * - 错误消息提示
 * - 异步函数错误处理包装
 * 
 * 该机制可以在整个应用中提供一致的错误处理体验，并简化错误处理代码
 * 
 * @example
 * // 基本用法
 * const { handleError, wrapAsync } = useErrorHandler()
 * 
 * // 直接处理错误
 * try {
 *   // 可能出错的代码
 * } catch (error) {
 *   handleError(error, '操作名称')
 * }
 * 
 * // 包装异步函数
 * const result = await wrapAsync(
 *   async () => await someAsyncFunction(),
 *   '操作名称'
 * )
 */

import { ElMessage } from 'element-plus'

/**
 * 错误类型枚举
 * 
 * @enum {string}
 * @description 定义了不同类型的错误，用于错误分类和处理
 */
export enum ErrorType {
  /** API 调用错误，如服务器响应错误、状态码错误等 */
  API = 'api',
  /** 数据验证错误，如表单验证失败、数据格式错误等 */
  VALIDATION = 'validation',
  /** 网络连接错误，如连接超时、断网等 */
  NETWORK = 'network',
  /** 未知错误，无法分类的错误 */
  UNKNOWN = 'unknown'
}

/**
 * 错误处理配置
 * 
 * @interface
 * @description 错误处理器的配置选项，控制错误处理的行为
 */
export interface ErrorHandlerConfig {
  /** 是否显示错误消息提示，默认为 true */
  showMessage?: boolean
  /** 是否将错误记录到控制台，默认为 true */
  logToConsole?: boolean
  /** 是否重新抛出错误，默认为 false */
  rethrow?: boolean
}

/**
 * 错误处理结果
 * 
 * @interface
 * @description 错误处理器返回的结果对象，包含错误处理的相关信息
 */
export interface ErrorHandlerResult {
  /** 是否成功，错误处理时始终为 false */
  success: boolean
  /** 错误消息文本 */
  message: string
  /** 错误类型 */
  type: ErrorType
}

/**
 * 错误处理 Hook
 * 
 * @returns 错误处理相关的方法
 * @description 返回错误处理器的方法，包括直接处理错误和包装异步函数
 */
export function useErrorHandler() {
  /**
   * 处理错误
   * 
   * @param {any} error 错误对象
   * @param {string} context 错误上下文，用于描述错误发生的位置或操作
   * @param {ErrorHandlerConfig} config 错误处理配置
   * @returns {ErrorHandlerResult} 错误处理结果
   * @description 处理给定的错误对象，根据配置进行日志记录、消息提示和错误重抛
   * @example
   * try {
   *   // 可能出错的代码
   * } catch (error) {
   *   handleError(error, '保存数据', { showMessage: true, rethrow: false })
   * }
   */
  const handleError = (
    error: any,
    context: string,
    config: ErrorHandlerConfig = { showMessage: true, logToConsole: true, rethrow: false }
  ): ErrorHandlerResult => {
    // 确定错误类型
    const errorType = determineErrorType(error)
    
    // 获取错误消息
    const errorMessage = getErrorMessage(error, context, errorType)
    
    // 记录到控制台
    if (config.logToConsole) {
      console.error(`[${context}] ${errorMessage}`, error)
    }
    
    // 显示消息
    if (config.showMessage) {
      ElMessage.error(errorMessage)
    }
    
    // 重新抛出错误
    if (config.rethrow) {
      throw error
    }
    
    return {
      success: false,
      message: errorMessage,
      type: errorType
    }
  }
  
  /**
   * 确定错误类型
   * 
   * @param {any} error 错误对象
   * @returns {ErrorType} 错误类型
   * @description 根据错误对象的特征确定错误类型，用于后续的错误处理
   * @private 内部方法，不应直接在外部使用
   */
  const determineErrorType = (error: any): ErrorType => {
    if (error.isAxiosError) {
      return ErrorType.API
    }
    
    if (error.name === 'ValidationError') {
      return ErrorType.VALIDATION
    }
    
    if (error.message && error.message.includes('network')) {
      return ErrorType.NETWORK
    }
    
    return ErrorType.UNKNOWN
  }
  
  /**
   * 获取错误消息
   * 
   * @param {any} error 错误对象
   * @param {string} context 错误上下文
   * @param {ErrorType} type 错误类型
   * @returns {string} 格式化后的错误消息
   * @description 根据错误类型和上下文格式化错误消息
   * @private 内部方法，不应直接在外部使用
   */
  const getErrorMessage = (error: any, context: string, type: ErrorType): string => {
    switch (type) {
      case ErrorType.API:
        return error.response?.data?.message || `${context}请求失败: ${error.message || '未知错误'}`
      case ErrorType.VALIDATION:
        return `${context}验证失败: ${error.message || '数据验证错误'}`
      case ErrorType.NETWORK:
        return `${context}网络错误: ${error.message || '网络连接问题'}`
      default:
        return `${context}错误: ${error.message || '未知错误'}`
    }
  }
  
  /**
   * 包装异步函数，自动处理错误
   * 
   * @template T 异步函数的返回类型
   * @param {() => Promise<T>} fn 要执行的异步函数
   * @param {string} context 错误上下文，用于描述操作
   * @param {ErrorHandlerConfig} config 错误处理配置
   * @returns {Promise<T | null>} 异步函数的结果，如果出错则返回 null
   * @description 包装异步函数，在出错时自动调用 handleError 进行错误处理
   * @example
   * const result = await wrapAsync(
   *   async () => await api.getData(),
   *   '获取数据',
   *   { showMessage: true, rethrow: false }
   * )
   * 
   * if (result) {
   *   // 处理成功的结果
   * }
   */
  const wrapAsync = async <T>(
    fn: () => Promise<T>,
    context: string,
    config: ErrorHandlerConfig = { showMessage: true, logToConsole: true, rethrow: false }
  ): Promise<T | null> => {
    try {
      return await fn()
    } catch (error) {
      handleError(error, context, config)
      return null
    }
  }
  
  return {
    handleError,
    wrapAsync
  }
}