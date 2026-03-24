import {
  apiSysConfigList,
  apiSysConfigCreate,
  apiSysConfigUpdate,
  apiSysConfigDelete,
  apiSysConfigKeys,
} from '@/api/system/config'
// ===== 系统配置服务 =====
class ConfigService {
  async getList(params) {
    return apiSysConfigList(params)
  }
  async create(data) {
    return apiSysConfigCreate(data)
  }
  async update(id, data) {
    return apiSysConfigUpdate({ id: Number(id), ...data })
  }
  async delete(id) {
    return apiSysConfigDelete(Number(id))
  }
  async getKeys() {
    return apiSysConfigKeys()
  }
}
export { ConfigService }
export const configService = new ConfigService()
