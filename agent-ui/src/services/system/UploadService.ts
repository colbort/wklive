import type { RespBase } from '@/services'
import { apiUploadAvatar, apiUploadFile } from '@/api/system/upload'

// ===== 文件上传相关类型定义 =====

export interface UploadFileResp {
  url: string
}

/**
 * 文件上传服务类
 * 使用现有的 API 函数
 */
export class UploadService {
  /**
   * 上传头像
   */
  async uploadAvatar(file: File): Promise<RespBase<UploadFileResp>> {
    return apiUploadAvatar(file)
  }

  /**
   * 通用文件上传（如果需要扩展）
   */
  async uploadFile(file: File, type: string = 'general'): Promise<RespBase<UploadFileResp>> {
    switch (type) {
      case 'avatar':
        return this.uploadAvatar(file)
      default:
        return apiUploadFile(file)
    }
  }
}

// 导出单例实例
export const uploadService = new UploadService()
