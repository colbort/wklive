/**
 * HTTP 请求工具（增强版）
 * 支持：请求拦截、响应拦截、错误处理、日志
 */
import axios from 'axios';
import { useAuthStore } from '@/stores';
import { logger } from '@/utils/logger';
export const http = axios.create({
    baseURL: 'http://localhost:8888', // ✅ admin-api:8888
    timeout: 15000,
    headers: {
        'Content-Type': 'application/json',
    },
});
// ==================== 请求拦截器 ====================
http.interceptors.request.use((config) => {
    const auth = useAuthStore();
    if (auth.token) {
        config.headers.Authorization = `Bearer ${auth.token}`;
    }
    logger.debug(`📤 [${config.method?.toUpperCase()}] ${config.url}`);
    return config;
}, (error) => {
    logger.error('Request error:', error);
    return Promise.reject(error);
});
// ==================== 响应拦截器 ====================
http.interceptors.response.use((resp) => {
    logger.debug(`✅ Response from ${resp.config.url}`);
    return resp.data;
}, (error) => {
    const { response } = error;
    if (error?.response?.status === 401) {
        const auth = useAuthStore();
        auth.logout();
        logger.warn('Token expired, redirect to login');
        location.href = '/login';
    }
    logger.error(`❌ [${response?.status}] ${error.config?.url}`);
    return Promise.reject(error);
});
// ==================== 通用 request ================== 
function request(method, url, options) {
    const { params, data, config } = options ?? {};
    return http.request({
        method,
        url,
        params,
        data,
        ...(config ?? {}),
    });
}
// ==================== 基础方法 ====================
export function get(url, params, config) {
    return request('GET', url, { params, config });
}
export function post(url, data, config) {
    return request('POST', url, { data, config });
}
export function put(url, data, config) {
    return request('PUT', url, { data, config });
}
export function del(url, params, config) {
    return request('DELETE', url, { params, config });
}
export function patch(url, data, config) {
    return request('PATCH', url, { data, config });
}
export default http;
