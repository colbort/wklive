import { get } from '@/utils/request';
export function getSystemCore() {
    return get('/admin/system/core');
}
