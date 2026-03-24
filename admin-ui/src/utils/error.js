/**
 * 错误处理工具
 */
import { ElMessage } from 'element-plus'
import { logger } from '@/utils/logger'
class ErrorHandler {
  /**
   * 处理异常并显示用户友好的消息
   */
  handle(error, defaultMsg) {
    const errorInfo = this.parse(error)
    this.log(errorInfo)
    this.notify(errorInfo)
    return errorInfo
  }
  /**
   * 解析异常对象
   */
  parse(error) {
    // 如果是已格式化的错误对象
    if (typeof error === 'object' && error !== null && 'message' in error) {
      return {
        ...error,
        message: error.message || 'Unknown error',
      }
    }
    // 如果是 Error 实例
    if (error instanceof Error) {
      return {
        message: error.message,
      }
    }
    // 如果是字符串
    if (typeof error === 'string') {
      return { message: error }
    }
    // 默认
    return { message: 'An unexpected error occurred' }
  }
  /**
   * 记录异常
   */
  log(error) {
    logger.error(`Error [${error.code || error.status || 'unknown'}]`, {
      message: error.message,
      data: error.data,
    })
  }
  /**
   * 通知用户
   */
  notify(error) {
    ElMessage.error({
      message: error.message,
      duration: 3000,
    })
  }
  /**
   * 静默处理（仅记录，不通知）
   */
  silent(error) {
    const errorInfo = this.parse(error)
    this.log(errorInfo)
    return errorInfo
  }
  /**
   * 自定义处理
   */
  custom(error, callback) {
    const errorInfo = this.parse(error)
    callback(errorInfo)
    return errorInfo
  }
}
export const errorHandler = new ErrorHandler()
export default errorHandler
