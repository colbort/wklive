import { defineStore } from 'pinia'
import { get, post } from '@/utils/request'
import type { RespBase } from '@/services'

// response payload returned by login endpoint
export type LoginResp = {
  token: string
  exp: number
}

export type ProfileUser = {
  id: number
  username: string
  nickname?: string
  avatar?: string
  tenantId: number // 所属租户ID：0=系统侧，>0=租户ID
  userType: number // 用户类型：1系统管理员 2租户主账号 3租户管理员
  isOwner: number // 是否租户主账号：1是 0否
}

export type MenuNode = {
  id: number
  parentId: number
  name: string
  menuType: 1 | 2 | 3
  path?: string
  component?: string
  icon?: string
  sort: number
  visible?: number
  enabled?: number
  perms?: string
  children?: MenuNode[]
}

export type ProfileResp = {
  user: ProfileUser
  menus: MenuNode[]
  perms: string[]
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    exp: Number(localStorage.getItem('exp') || 0),
    tenantId: Number(localStorage.getItem('tenantId') || 0),
    user: null as ProfileUser | null,
    menus: [] as MenuNode[],
    perms: [] as string[],
    isProfileLoaded: false,
  }),
  getters: {
    hasPerm: (s) => (p: string) => s.perms.includes(p),
  },
  actions: {
    async login(payload: { username: string; password: string; googleCode?: string }) {
      const res = await post<LoginResp>('/admin/system/auth/login', payload)
      if (res.code !== 200) {
        throw new Error(res.msg || 'login failed')
      }
      // payload is stored at top level since RespBase strips `data`
      this.token = res.data!.token
      this.exp = res.data!.exp
      localStorage.setItem('token', res.data!.token)
      localStorage.setItem('exp', String(res.data!.exp))
    },
    async fetchProfile() {
      const res = await get<ProfileResp>('/admin/system/auth/profile')
      if (res.code !== 200) throw new Error(res.msg || 'profile failed')
      this.user = res.data!.user
      this.tenantId = Number(res.data!.user?.tenantId || 0)
      this.menus = res.data!.menus || []
      this.perms = res.data!.perms || []
      this.isProfileLoaded = true
      localStorage.setItem('tenantId', String(this.tenantId))
    },
    logout() {
      this.token = ''
      this.exp = 0
      this.tenantId = 0
      this.user = null
      this.menus = []
      this.perms = []
      this.isProfileLoaded = false
      localStorage.removeItem('token')
      localStorage.removeItem('exp')
      localStorage.removeItem('tenantId')
    },
  },
})

export function apiUpdateProfile(data: {
  nickname?: string
  avatar?: string
  password?: string
}): Promise<RespBase> {
  return post<RespBase>('/admin/system/auth/profile', data)
}
