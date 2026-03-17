import { get, post, put, del } from '@/utils/request';
// ===== API 函数 =====
export function apiSysConfigList(params) {
    return get('/admin/configs', { params });
}
export function apiSysConfigCreate(data) {
    return post('/admin/configs', data);
}
export function apiSysConfigUpdate(data) {
    return put('/admin/configs', data);
}
export function apiSysConfigDelete(id) {
    return del(`/admin/configs/${id}`);
}
