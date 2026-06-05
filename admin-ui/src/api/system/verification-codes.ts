import { get, post } from '@/utils/request'
import type {
  RespBase,
  TestVerificationCodeReq,
  VerificationCodeRecordListReq,
  VerificationCodeRecordItem,
} from '@/services'

export function apiVerificationCodeRecordList(
  params: VerificationCodeRecordListReq,
): Promise<RespBase<VerificationCodeRecordItem[]>> {
  return get<VerificationCodeRecordItem[]>('/admin/system/verification-codes', params)
}

export function apiVerificationCodeRecordDetail(
  id: number,
): Promise<RespBase<VerificationCodeRecordItem>> {
  return get<VerificationCodeRecordItem>(`/admin/system/verification-codes/${id}`)
}

export function apiTestVerificationCode(data: TestVerificationCodeReq): Promise<RespBase> {
  return post('/admin/system/verification-codes/test', data)
}
