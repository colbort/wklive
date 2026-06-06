import { http } from './http'
import type { RespBase } from '../types/api'

export interface UploadFileData {
  url: string
}

export function apiUploadFile(file: File): Promise<RespBase & { data: UploadFileData }> {
  const formData = new FormData()
  formData.append('file', file)

  return http
    .post('/user/upload/file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    .then((res: { data: any }) => res.data)
}
