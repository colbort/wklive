import type { RespBase, BaseService } from '@/services'
import {
  apiUserList,
  apiUserDetail,
  apiUserCreate,
  apiUserUpdate,
  apiUserDelete,
  apiChangeUserStatus,
  apiResetUserPwd,
  apiAssignUserRoles,
  apiGoogle2faInit,
  apiGoogle2faEnable,
  apiGoogle2faDisable,
  apiGoogle2faReset
} from '@/api/system/users'

// ===== 用户相关类型定义 =====

export type SysUserItem = {
  id: number
  username: string
  nickname: string
  status: number
  roleIds: number[]
  createdAt: number
  google2faEnabled: number
}

export type Google2FABindInitResp = {
  secret: string
  otpauthUrl: string
  qrCode: string
}

// 用户接口定义（复用现有的类型）
export interface User extends SysUserItem {}

export interface CreateUserRequest {
  username: string
  password: string
  nickname?: string
  status?: number
  roleIds?: number[]
}

export interface UpdateUserRequest {
  nickname?: string
  status?: number
  roleIds?: number[]
}

export interface UserQueryParams {
  keyword?: string
  status?: number
  page?: number
  size?: number
}

/**
 * 用户服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class UserService implements BaseService {

  /**
   * 获取用户列表（支持分页和筛选）
   */
  async getList(params?: UserQueryParams): Promise<RespBase<SysUserItem[]>> {
    return apiUserList(params || {})
  }

  /**
   * 获取用户详情
   */
  async getDetail(id: string | number): Promise<RespBase<SysUserItem>> {
    return apiUserDetail(Number(id))
  }

  /**
   * 创建用户
   */
  async create(data: CreateUserRequest): Promise<RespBase<SysUserItem>> {
    return apiUserCreate(data)
  }

  /**
   * 更新用户
   */
  async update(id: string | number, data: UpdateUserRequest): Promise<RespBase<SysUserItem>> {
    return apiUserUpdate({ ...data, id: Number(id) })
  }

  /**
   * 删除用户
   */
  async delete(id: string | number): Promise<RespBase<null>> {
    return apiUserDelete(Number(id))
  }

  /**
   * 批量删除用户
   */
  async batchDelete(ids: (string | number)[]): Promise<RespBase<null>> {
    // 现有 API 不支持批量删除，这里实现为逐个删除
    for (const id of ids) {
      const result = await this.delete(id)
      if (result.code !== 200) {
        return result
      }
    }
    return { code: 200, msg: '批量删除成功' }
  }

  /**
   * 重置用户密码
   */
  async resetPassword(id: number, newPassword: string): Promise<RespBase<null>> {
    return apiResetUserPwd({ id, password: newPassword })
  }

  /**
   * 更新用户状态
   */
  async updateUserStatus(id: number, status: number): Promise<RespBase<User>> {
    return apiChangeUserStatus({ id, status })
  }

  /**
   * 分配用户角色
   */
  async assignUserRoles(id: number, roleIds: number[]): Promise<RespBase<null>> {
    return apiAssignUserRoles({ userId: id, roleIds })
  }

  /**
   * 初始化 Google 2FA
   */
  async initGoogle2FA(userId: number): Promise<RespBase<Google2FABindInitResp>> {
    return apiGoogle2faInit({ userId })
  }

  /**
   * 启用 Google 2FA
   */
  async enableGoogle2FA(userId: number, code: string): Promise<RespBase<null>> {
    return apiGoogle2faEnable({ userId, code })
  }

  /**
   * 禁用 Google 2FA
   */
  async disableGoogle2FA(userId: number, code?: string): Promise<RespBase<null>> {
    return apiGoogle2faDisable({ userId, code })
  }

  /**
   * 重置 Google 2FA
   */
  async resetGoogle2FA(userId: number): Promise<RespBase<null>> {
    return apiGoogle2faReset({ userId })
  }
}

// 导出单例实例
export const userService = new UserService()