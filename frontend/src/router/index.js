// 路由配置
// 当前 Demo 只展示「数据洞察」页面, 预留其它 Tab 的扩展点
import { createRouter, createWebHashHistory } from 'vue-router'

const InsightDashboard = () => import('../views/InsightDashboard.vue')

const routes = [
  { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true, title: '登录' } },
  { path: '/', redirect: '/insight' },
  { path: '/insight', component: InsightDashboard, meta: { title: '数据洞察' } },
  { path: '/profile', component: () => import('../views/ProfileView.vue'), meta: { title: '个人资料', requiresAuth: true } },
  { path: '/creator', component: () => import('../views/CreatorAnalysis.vue'), meta: { title: '达人分析' } },
  { path: '/content', component: () => import('../views/PlaceholderView.vue'), meta: { title: '内容分析' } },
  { path: '/market', component: () => import('../views/PlaceholderView.vue'), meta: { title: '市场洞察' } },
  { path: '/brand', component: () => import('../views/PlaceholderView.vue'), meta: { title: '品牌分析' } },
  { path: '/custom', component: () => import('../views/PlaceholderView.vue'), meta: { title: '自定义看板' } }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 看板公开: 未登录也能访问; 仅已登录访问登录页时弹回看板(避免重复登录)
// 个人资料等 requiresAuth 页面: 无 token 一律跳登录页
router.beforeEach((to) => {
  const token = localStorage.getItem('insta_token')
  // 已登录却访问登录页 → 弹回看板
  if (to.meta && to.meta.public && token) {
    return '/insight'
  }
  // 需要登录的页面(如个人资料)无 token → 跳登录
  if (to.meta && to.meta.requiresAuth && !token) {
    return '/login'
  }
  return true
})

export default router
