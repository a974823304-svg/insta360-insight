// 入口文件
// 1. 注册 Vue 3
// 2. 注入 Pinia (全局筛选状态)
// 3. 注册 Element Plus + 暗色主题
// 4. 挂载 Vue Router (目前只有数据洞察一个页面, 预留多页面扩展)
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as echarts from 'echarts'

import App from './App.vue'
import router from './router'
import './styles/theme.scss'

const app = createApp(App)

// 关键:开启 Element Plus 暗色主题(需要在 <html> 上加 class="dark")
document.documentElement.classList.add('dark')

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

// 全局挂载 ECharts 供组件内部使用,避免每个组件重复 import
app.config.globalProperties.$echarts = echarts

app.mount('#app')
