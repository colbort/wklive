import { post } from '@/utils/request'
import { ApiResp } from '../types'

export interface UploadFileResp {
  url: string
}

export function apiUploadAvatar(file: File): Promise<ApiResp<UploadFileResp>> {
  const formData = new FormData()
  formData.append('file', file)

  return post<UploadFileResp>('/admin/upload/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}