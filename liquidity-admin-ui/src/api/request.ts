import axios from "axios";
import { ElMessage } from "element-plus";

export const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/liquidity/admin",
  timeout: 15_000,
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem("liquidity_admin_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.message || error.message || "请求失败";
    ElMessage.error(message);
    if (error.response?.status === 401) {
      localStorage.removeItem("liquidity_admin_token");
      location.href = "/login";
    }
    return Promise.reject(error);
  },
);
