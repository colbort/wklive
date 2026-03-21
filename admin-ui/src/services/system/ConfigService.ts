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
  page?: number
  size?: number
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

// ===== 系统配置服务 =====

class ConfigService implements BaseService {
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

export { ConfigService }
export const configService = new ConfigService()