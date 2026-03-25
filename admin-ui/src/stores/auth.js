import { defineStore } from 'pinia';
import { get, post } from '@/utils/request';
export const useAuthStore = defineStore('auth', {
    state: () => ({
        token: localStorage.getItem('token') || '',
        exp: Number(localStorage.getItem('exp') || 0),
        user: null,
        menus: [],
        perms: [],
        isProfileLoaded: false,
    }),
    getters: {
        hasPerm: (s) => (p) => s.perms.includes(p),
    },
    actions: {
        async login(payload) {
            const res = await post('/admin/auth/login', payload);
            if (res.code !== 200)
                throw new Error(res.msg || 'login failed');
            // payload is stored at top level since RespBase strips `data`
            this.token = res.data.token;
            this.exp = res.data.exp;
            localStorage.setItem('token', res.data.token);
            localStorage.setItem('exp', String(res.data.exp));
        },
        async fetchProfile() {
            const res = await get('/admin/auth/profile');
            if (res.code !== 200)
                throw new Error(res.msg || 'profile failed');
            this.user = res.data.user;
            this.menus = res.data.menus || [];
            this.perms = res.data.perms || [];
            this.isProfileLoaded = true;
        },
        logout() {
            this.token = '';
            this.exp = 0;
            this.user = null;
            this.menus = [];
            this.perms = [];
            this.isProfileLoaded = false;
            localStorage.removeItem('token');
            localStorage.removeItem('exp');
        },
    },
});
export function apiUpdateProfile(data) {
    return post('/admin/auth/profile', data);
}
