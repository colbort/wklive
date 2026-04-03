import type { RespBase, BaseService } from '@/services'
import {
  apiSysConfigList,
  apiSysConfigCreate,
  apiSysConfigUpdate,
  apiSysConfigDelete,
  apiSysConfigKeys,
} from '@/api/system/config'

// ===== 系统配置类型定义 =====

export type SysConfigItem = {
  id: number
  configKey: string
  configValue: string
  remark?: string
  createdAt: number
}

export type SysConfigListReq = {
  keyword?: string
  cursor?: string | null
  limit?: number
}

export type SysConfigListResp = RespBase<SysConfigItem[]>

export type SysConfigCreateReq = {
  configKey: string
  configValue: string
  remark?: string
}

export type SysConfigUpdateReq = {
  id: number
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
  region: string
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

  async getKeys(): Promise<RespBase<string[]>> {
    return apiSysConfigKeys()
  }
}

export const configService = new ConfigService()
