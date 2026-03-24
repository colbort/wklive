import {
  apiRoleList,
  apiRoleCreate,
  apiRoleUpdate,
  apiRoleDelete,
  apiRoleGrant,
  apiRoleGrantDetail,
} from '@/api/system/roles'
/**
 * 角色服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class RoleService {
  /**
   * 获取角色列表（支持分页和筛选）
   */
  async getList(params) {
    return apiRoleList(params || {})
  }
  /**
   * 创建角色
   */
  async create(data) {
    return apiRoleCreate(data)
  }
  /**
   * 更新角色
   */
  async update(id, data) {
    return apiRoleUpdate({ ...data, id: Number(id) })
  }
  /**
   * 删除角色
   */
  async delete(id) {
    return apiRoleDelete(Number(id))
  }
  /**
   * 角色授权
   */
  async grantRole(data) {
    return apiRoleGrant(data)
  }
  /**
   * 获取角色授权详情
   */
  async getRoleGrantDetail(roleId) {
    return apiRoleGrantDetail(roleId)
  }
}
// 导出单例实例
export const roleService = new RoleService()
