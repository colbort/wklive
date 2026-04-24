import type { RouteRecordRaw } from 'vue-router'

export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true, titleKey: 'route.login' },
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layout/index.vue'),
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/home/index.vue'),
        meta: { title: '工作台', affix: true },
      },
      {
        path: 'categories',
        name: 'TenantCategories',
        component: () => import('@/views/tenant/categories.vue'),
        meta: { title: '分类管理' },
      },
      {
        path: 'products',
        name: 'TenantProducts',
        component: () => import('@/views/tenant/products.vue'),
        meta: { title: '产品管理' },
      },
      {
        path: 'users',
        name: 'TenantUsers',
        component: () => import('@/views/tenant/users.vue'),
        meta: { title: '用户管理' },
      },
      {
        path: 'recharge-orders',
        name: 'RechargeOrders',
        component: () => import('@/views/tenant/recharge-orders.vue'),
        meta: { title: '充值订单' },
      },
      {
        path: 'withdraw-orders',
        name: 'WithdrawOrders',
        component: () => import('@/views/tenant/withdraw-orders.vue'),
        meta: { title: '提现订单' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { public: true, titleKey: 'route.notFound' },
  },
]
