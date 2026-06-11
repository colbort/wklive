import type { RespBase, BaseService, OptionItem } from '@/services'
import {
  apiSysConfigList,
  apiSysConfigCreate,
  apiSysConfigUpdate,
  apiSysConfigDelete,
  } from '@/api/system/config'
import { getCoreOptions } from '@/stores/core'

// ===== 系统配置类型定义 =====

export type SysConfigItem = {
  id: number
  tenantId: number
  configKey: string
  configValue: string
  remark?: string
  createTimes: number
  updateTimes: number
}

export type SysConfigListReq = {
  keyword?: string
  tenantId?: number
  cursor?: number
  limit?: number
}

export type SysConfigListResp = RespBase<SysConfigItem[]>

export type SysConfigCreateReq = {
  tenantId: number
  configKey: string
  configValue: string
  remark?: string
}

export type SysConfigUpdateReq = {
  id: number
  tenantId?: number
  configKey?: string
  configValue?: string
  remark?: string
}

export type SysConfigDeleteReq = {
  id: number
}

// 系统配置
export type SystemCore = {
  site_name: string // 网站名称
  site_logo: string // 网站LOGO
  is_captcha_enabled: number // 是否开启验证码
  is_register_enabled: number // 是否开启注册
  is_guest_enabled: number // 是否允许游客登录
  is_crypto_enabled: number // 是否加密接口提交数据
}

export type ItickConfig = {
  api_url: string // ITICK API地址
  api_token: string // ITICK API密钥
  ws_url: string // ITICK WebSocket地址
}

// 阿里云 OSS配置
export type AliyunOssConfig = {
  endpoint: string
  access_key_id: string
  access_key_secret: string
  bucket_name: string
  bucket_url: string
}

// 腾讯云 COS配置
export type TencentCosConfig = {
  market: string
  secret_id: string
  secret_key: string
  bucket_name: string
  bucket_url: string
}

// MINIO配置
export type MinioConfig = {
  endpoint: string
  access_key_id: string
  access_key_secret: string
  bucket_name: string
  bucket_url: string
}

// 对象存储配置
export type ObjectStorageConfig = {
  aliyun_oss: AliyunOssConfig
  tencent_cos: TencentCosConfig
  minio: MinioConfig
  oss_type: number // 1阿里云OSS 2腾讯云COS 3MINIO
  oss_domain: string // 对象存储访问域名（可选，优先使用bucket_url）
}

// 充值配置
export type RechargeConfig = {
  minAmount: number // 最小充值金额，单位：分
  maxAmount: number // 最大充值金额，单位：分
  feeRate: number // 手续费配置，单位：万分之几，例如：100表示1%
}

// 提现配置
export type WithdrawConfig = {
  minAmount: number // 最小提现金额，单位：分
  maxAmount: number // 最大提现金额，单位：分
  feeRate: number // 手续费配置，单位：万分之几，例如：100表示1%
  dailyLimitPerUser: number // 每人每天提现次数限制
  dailyAmountLimitPerUser: number // 每人每天提现金额限制，单位：分
  allowedTimeRange: string // 允许提现的时间段，例如：每天9:00-18:00，格式为"09:00-18:00"
  pendingWithdrawalLimitPerUser: number // 允许未审核在提现数量限制，单位：笔
  freeWithdrawTimesPerDay: number // 每日免费提现次数，0=没有免费提现
}

export type EmailConfig = {
  enabled: number
  smtp_host: string
  smtp_port: number
  username: string
  password: string
  from_email: string
  from_name: string
  subject_template: string
  body_template: string
}

export type PhoneConfig = {
  enabled: number
  provider: string
  endpoint: string
  method: string
  headers_json: string
  body_template: string
}

// ===== 系统配置服务 =====

export class ConfigService implements BaseService {
  async getList(params: SysConfigListReq): Promise<RespBase<SysConfigItem[]>> {
    return apiSysConfigList(params)
  }

  async create(data: SysConfigCreateReq): Promise<RespBase> {
    return apiSysConfigCreate(data)
  }

  async update(id: string | number, data: Partial<SysConfigUpdateReq>): Promise<RespBase> {
    return apiSysConfigUpdate({ id: Number(id), ...data })
  }

  async delete(id: string | number): Promise<RespBase> {
    return apiSysConfigDelete(Number(id))
  }

  async getKeys(): Promise<RespBase<OptionItem[]>> {
    const res = await getCoreOptions()
    const group = (res.data || []).find((item) => item.key === 'sysConfigType')
    return {
      ...res,
      data: group?.options ?? [],
    }
  }
}

export const configService = new ConfigService()
