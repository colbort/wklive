import type { BaseService, RespBase } from '@/services'
import {
  apiTestVerificationCode,
  apiVerificationCodeRecordDetail,
  apiVerificationCodeRecordList,
} from '@/api/system/verification-codes'

export type VerificationCodeRecordItem = {
  id: number
  tenantId: number
  channel: number
  target: string
  scene: number
  code: string
  status: number
  provider?: string
  errorMessage?: string
  createTimes: number
  updateTimes: number
}

export type VerificationCodeRecordListReq = {
  tenantId?: number
  channel?: number
  target?: string
  scene?: number
  status?: number
  cursor?: number
  limit?: number
}

export type TestVerificationCodeReq = {
  tenantId?: number
  channel: number
  email?: string
  phone?: string
  scene?: number
}

export class VerificationCodeService implements BaseService {
  async getList(
    params: VerificationCodeRecordListReq,
  ): Promise<RespBase<VerificationCodeRecordItem[]>> {
    return apiVerificationCodeRecordList(params)
  }

  async getDetail(id: string | number): Promise<RespBase<VerificationCodeRecordItem>> {
    return apiVerificationCodeRecordDetail(Number(id))
  }

  async testSend(data: TestVerificationCodeReq): Promise<RespBase> {
    return apiTestVerificationCode(data)
  }
}

export const verificationCodeService = new VerificationCodeService()
