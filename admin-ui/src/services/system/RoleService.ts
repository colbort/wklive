import type { RespBase, BaseService, OptionGroup } from '@/services'
import {
  apiRoleList,
  apiRoleCreate,
  apiRoleUpdate,
  apiRoleDelete,
  apiRoleGrant,
  apiRoleGrantDetail,
} from '@/api/system/roles'
import { getCoreOptions } from '@/stores/core'

// ===== 角色相关类型定义 =====

export type SysRole = {
  id: number
  name: string
  code: string
  remark?: string
  enabled?: number
  tenantId?: number
  isSuper?: boolean // 可选：如果后端有的话
}

export type Role = SysRole

export type RoleListResp = {
  list: SysRole[]
  total: number
}

export type RoleItem = { id: number; name: string; code: string; enabled: number; remark?: string }

export interface CreateRoleRequest {
  name: string
  code: string
  description?: string
  enabled?: number
  menuIds?: number[]
  permIds?: number[]
}

export interface UpdateRoleRequest {
  id: number
  name?: string
  code?: string
  description?: string
  enabled?: number
  menuIds?: number[]
  permIds?: number[]
}

export type RoleQueryParams = {
  keyword?: string
  enabled?: number
  cursor?: number
  limit?: number
}

export type RoleGrantRequest = {
  roleId: number
  menuIds: number[]
  permKeys: string[]
}

/**
 * 角色服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class RoleService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  /**
   * 获取角色列表（支持分页和筛选）
   */
  async getList(params?: RoleQueryParams): Promise<RespBase<SysRole[]>> {
    return apiRoleList(params || {})
  }

  /**
   * 创建角色
   */
  async create(data: CreateRoleRequest): Promise<RespBase<SysRole>> {
    return apiRoleCreate(data)
  }

  /**
   * 更新角色
   */
  async update(id: string | number, data: UpdateRoleRequest): Promise<RespBase<SysRole>> {
    return apiRoleUpdate({ ...data, id: Number(id) })
  }

  /**
   * 删除角色
   */
  async delete(id: string | number): Promise<RespBase<null>> {
    return apiRoleDelete(Number(id))
  }

  /**
   * 角色授权
   */
  async grantRole(data: RoleGrantRequest): Promise<RespBase<null>> {
    return apiRoleGrant(data)
  }

  /**
   * 获取角色授权详情
   */
  async getRoleGrantDetail(roleId: number): Promise<RespBase<RoleGrantRequest>> {
    return apiRoleGrantDetail(roleId)
  }
}

// 导出单例实例
export const roleService = new RoleService()
