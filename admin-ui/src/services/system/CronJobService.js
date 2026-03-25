import { apiSysCronJobList, apiSysCronJobCreate, apiSysCronJobUpdate, apiSysCronJobDelete, apiSysCronJobRun, apiSysCronJobStart, apiSysCronJobStop, apiSysCronJobHandlers, apiSysCronJobLogList, } from '@/api/system/cronjob';
class CronJobService {
    async getList(params) {
        return apiSysCronJobList(params);
    }
    async create(data) {
        return apiSysCronJobCreate(data);
    }
    async update(id, data) {
        return apiSysCronJobUpdate({ id: Number(id), ...data });
    }
    async delete(id) {
        return apiSysCronJobDelete(id);
    }
    async run(id) {
        return apiSysCronJobRun(id);
    }
    async start(id) {
        return apiSysCronJobStart(id);
    }
    async stop(id) {
        return apiSysCronJobStop(id);
    }
    async handlers() {
        return apiSysCronJobHandlers();
    }
    async getLogList(params) {
        return apiSysCronJobLogList(params);
    }
}
export { CronJobService };
export const cronJobService = new CronJobService();
