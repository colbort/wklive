import { get, post, put, del } from '@/utils/request';
// ===== types =====
export function apiRoleList(params) {
    return get('/admin/roles', { params });
}
export async function apiRoleCreate(data) {
    // POST /roles
    return await post('/admin/roles', data);
}
export async function apiRoleUpdate(data) {
    // PUT /roles
    return await put('/admin/roles', data);
}
export async function apiRoleDelete(id) {
    // DELETE /roles/:id
    return await del(`/admin/roles/${id}`);
}
export async function apiRoleGrant(data) {
    // POST /roles/grant
    return await post('/admin/roles/grant', data);
}
export async function apiRoleGrantDetail(roleId) {
    return await get(`/admin/roles/${roleId}/grant`);
}
