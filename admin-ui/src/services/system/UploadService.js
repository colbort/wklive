import { apiUploadAvatar } from '@/api/system/upload'
/**
 * 文件上传服务类
 * 使用现有的 API 函数
 */
export class UploadService {
  /**
   * 上传头像
   */
  async uploadAvatar(file) {
    return apiUploadAvatar(file)
  }
  /**
   * 通用文件上传（如果需要扩展）
   */
  async uploadFile(file, type = 'general') {
    // 这里可以根据 type 调用不同的上传接口
    // 目前只有头像上传，未来可以扩展其他类型
    switch (type) {
      case 'avatar':
        return this.uploadAvatar(file)
      default:
        return this.uploadAvatar(file) // 默认使用头像上传
    }
  }
}
// 导出单例实例
export const uploadService = new UploadService()
