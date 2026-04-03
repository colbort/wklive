import { post } from '@/utils/request'
import type { RespBase, UploadFileResp } from '@/services'

export function apiUploadAvatar(file: File): Promise<RespBase<UploadFileResp>> {
  const formData = new FormData()
  formData.append('file', file)

  return post<UploadFileResp>('/admin/system/upload/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}
