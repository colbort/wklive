import { post } from '@/utils/request'
import type { RespBase, UploadFileResp } from '@/services'

function buildUploadFormData(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return formData
}

export function apiUploadAvatar(file: File): Promise<RespBase<UploadFileResp>> {
  const formData = buildUploadFormData(file)

  return post<UploadFileResp>('/admin/upload/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function apiUploadFile(file: File): Promise<RespBase<UploadFileResp>> {
  const formData = buildUploadFormData(file)

  return post<UploadFileResp>('/admin/upload/file', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}
