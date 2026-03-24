import {
  apiMenuTree,
  apiPermList,
  sysMenuCreate,
  sysMenuUpdate,
  sysMenuDelete,
  sysMenuList,
} from '@/api/system/menus'
/**
 * 菜单服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class MenuService {
  /**
   * 获取菜单树
   */
  async getMenuTree() {
    return apiMenuTree()
  }
  /**
   * 获取权限列表
   */
  async getPermissionList() {
    return apiPermList()
  }
  /**
   * 获取菜单列表
   */
  async getList(params) {
    return sysMenuList(params || {})
  }
  /**
   * 创建菜单
   */
  async create(data) {
    return sysMenuCreate(data)
  }
  /**
   * 更新菜单
   */
  async update(id, data) {
    return sysMenuUpdate({ ...data, id: Number(id) })
  }
  /**
   * 删除菜单
   */
  async delete(id) {
    return sysMenuDelete(Number(id))
  }
}
// 导出单例实例
export const menuService = new MenuService()
