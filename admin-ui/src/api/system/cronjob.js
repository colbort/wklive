import { get, post, put, del } from '@/utils/request';
export function apiSysCronJobList(params) {
    return get('/admin/jobs', params);
}
export function apiSysCronJobCreate(data) {
    return post('/admin/jobs', data);
}
export function apiSysCronJobUpdate(data) {
    return put('/admin/jobs', data);
}
export function apiSysCronJobDelete(id) {
    return del(`/admin/jobs/${id}`);
}
export function apiSysCronJobRun(id) {
    return post(`/admin/jobs/${id}/run`);
}
export function apiSysCronJobStart(id) {
    return post(`/admin/jobs/${id}/start`);
}
export function apiSysCronJobStop(id) {
    return post(`/admin/jobs/${id}/stop`);
}
export function apiSysCronJobHandlers() {
    return get('/admin/jobs/handlers');
}
export function apiSysCronJobLogList(params) {
    return get('/admin/logs/job', params);
}
