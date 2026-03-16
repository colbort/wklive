import { post } from '@/utils/request';
export function apiUploadAvatar(file) {
    const formData = new FormData();
    formData.append('file', file);
    return post('/admin/upload/avatar', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
}
