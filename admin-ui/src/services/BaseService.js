// ===== 通用类型定义 =====
/**
 * 基础服务实现类
 * 提供一些通用功能
 */
export class BaseServiceImpl {
  baseURL
  timeout
  constructor(baseURL, options = {}) {
    this.baseURL = baseURL
    this.timeout = options.timeout || 10000
  }
}
