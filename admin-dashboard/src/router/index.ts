import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '首页看板', icon: 'Odometer' }
        },
        {
          path: 'orders',
          name: 'Orders',
          component: () => import('@/views/orders/list.vue'),
          meta: { title: '订单列表', icon: 'List' }
        },
        {
          path: 'orders/:id',
          name: 'OrderDetail',
          component: () => import('@/views/orders/detail.vue'),
          meta: { title: '订单详情', hidden: true }
        },
        {
          path: 'inventory',
          name: 'Inventory',
          component: () => import('@/views/inventory/index.vue'),
          meta: { title: '库存概览', icon: 'Box' }
        },
        {
          path: 'products',
          name: 'Products',
          component: () => import('@/views/products/index.vue'),
          meta: { title: '商品管理', icon: 'Goods' }
        }
      ]
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && authStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router
