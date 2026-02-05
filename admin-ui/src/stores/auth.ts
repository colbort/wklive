import { defineStore } from 'pinia'
import { http } from '@/utils/request'

export type ProfileUser = {
  id: number
  username: string
  nickname?: string
  avatar?: string
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
  status?: number
  perms?: string
  children?: MenuNode[]
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    exp: Number(localStorage.getItem('exp') || 0),
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
      const res = await http.post('/admin/auth/login', payload)
      if (res.code !== 200) throw new Error(res.msg || 'login failed')
      this.token = res.token
      this.exp = res.exp
      localStorage.setItem('token', res.token)
      localStorage.setItem('exp', String(res.exp))
    },
    async fetchProfile() {
      const res = await http.get('/admin/auth/profile')
      if (res.code !== 200) throw new Error(res.msg || 'profile failed')
      this.user = res.user
      this.menus = res.menus || []
      this.perms = res.perms || []
      this.isProfileLoaded = true
    },
    logout() {
      this.token = ''
      this.exp = 0
      this.user = null
      this.menus = []
      this.perms = []
      this.isProfileLoaded = false
      localStorage.removeItem('token')
      localStorage.removeItem('exp')
    },
  },
})
