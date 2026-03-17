import { apiLoginLogList, apiOpLogList } from '@/api/system/logs';
/**
 * 日志服务类
 * 使用现有的 API 函数
 */
export class LogService {
    baseURL;
    constructor() {
        this.baseURL = '/admin/logs';
    }
    /**
     * 获取登录日志列表
     */
    async getLoginLogs(params) {
        return apiLoginLogList(params || {});
    }
    /**
     * 获取操作日志列表
     */
    async getOperationLogs(params) {
        return apiOpLogList(params || {});
    }
}
// 导出单例实例
export const logService = new LogService();
