import { apiUserList, apiUserDetail, apiUserCreate, apiUserUpdate, apiUserDelete, apiChangeUserStatus, apiResetUserPwd, apiAssignUserRoles, apiGoogle2faInit, apiGoogle2faEnable, apiGoogle2faDisable, apiGoogle2faReset } from '@/api/system/users';
/**
 * 用户服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class UserService {
    /**
     * 获取用户列表（支持分页和筛选）
     */
    async getList(params) {
        return apiUserList(params || {});
    }
    /**
     * 获取用户详情
     */
    async getDetail(id) {
        return apiUserDetail(Number(id));
    }
    /**
     * 创建用户
     */
    async create(data) {
        return apiUserCreate(data);
    }
    /**
     * 更新用户
     */
    async update(id, data) {
        return apiUserUpdate({ ...data, id: Number(id) });
    }
    /**
     * 删除用户
     */
    async delete(id) {
        return apiUserDelete(Number(id));
    }
    /**
     * 批量删除用户
     */
    async batchDelete(ids) {
        // 现有 API 不支持批量删除，这里实现为逐个删除
        for (const id of ids) {
            const result = await this.delete(id);
            if (result.code !== 200) {
                return result;
            }
        }
        return { code: 200, msg: '批量删除成功' };
    }
    /**
     * 重置用户密码
     */
    async resetPassword(id, newPassword) {
        return apiResetUserPwd({ id, password: newPassword });
    }
    /**
     * 更新用户状态
     */
    async updateUserStatus(id, status) {
        return apiChangeUserStatus({ id, status });
    }
    /**
     * 分配用户角色
     */
    async assignUserRoles(id, roleIds) {
        return apiAssignUserRoles({ userId: id, roleIds });
    }
    /**
     * 初始化 Google 2FA
     */
    async initGoogle2FA(userId) {
        return apiGoogle2faInit({ userId });
    }
    /**
     * 启用 Google 2FA
     */
    async enableGoogle2FA(userId, code) {
        return apiGoogle2faEnable({ userId, code });
    }
    /**
     * 禁用 Google 2FA
     */
    async disableGoogle2FA(userId, code) {
        return apiGoogle2faDisable({ userId, code });
    }
    /**
     * 重置 Google 2FA
     */
    async resetGoogle2FA(userId) {
        return apiGoogle2faReset({ userId });
    }
}
// 导出单例实例
export const userService = new UserService();
