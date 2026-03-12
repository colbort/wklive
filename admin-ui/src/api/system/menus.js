import { del, get, post, put } from '@/utils/request';
export async function apiMenuTree() {
    return await get('/admin/menus/tree');
}
export async function apiPermList() {
    return await get('/admin/perms');
}
/** 新增菜单 */
export async function sysMenuCreate(data) {
    return await post('/admin/menus', data);
}
/** 更新菜单 */
export async function sysMenuUpdate(data) {
    return await put('/admin/menus', data);
}
/** 删除菜单 */
export async function sysMenuDelete(id) {
    return await del(`/admin/menus/${id}`);
}
/** 菜单列表 */
export async function sysMenuList(data) {
    return await get('/admin/menus', data);
}
