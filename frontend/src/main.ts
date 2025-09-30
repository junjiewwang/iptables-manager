import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import 'element-plus/dist/index.css'
import '@/assets/index.css'

import App from './App.vue'

// 使用动态导入实现路由级别的代码分割
const Tables = () => import('./views/Tables.vue')
const Topology = () => import('./views/Topology.vue')
const Interfaces = () => import('./views/Interfaces.vue')
const Logs = () => import('./views/Logs.vue')
const Login = () => import('./views/Login.vue')
const ChainTableView = () => import('./views/ChainTableView.vue')
const TunnelAnalysis = () => import('./components/tunnel-analysis/TunnelAnalysis.vue')

const routes = [
  { 
    path: '/login', 
    component: Login,
    // 为路由添加元数据，用于预加载
    meta: { preload: true } 
  },
  { 
    path: '/', 
    redirect: '/chain-table-view' 
  },
  { 
    path: '/tables', 
    component: Tables,
    // 为路由添加名称，便于导航
    name: 'tables' 
  },
  { 
    path: '/topology', 
    component: Topology,
    name: 'topology' 
  },
  { 
    path: '/interfaces', 
    component: Interfaces,
    name: 'interfaces' 
  },
  { 
    path: '/chain-table-view', 
    component: ChainTableView,
    name: 'chain-table-view',
    meta: { preload: true } 
  },
  { 
    path: '/tunnel-analysis', 
    component: TunnelAnalysis,
    name: 'tunnel-analysis' 
  },
  { 
    path: '/logs', 
    component: Logs,
    name: 'logs' 
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 添加路由预加载功能
router.beforeResolve((to, from, next) => {
  // 预加载当前路由
  if (typeof to.matched[0]?.components.default === 'function') {
    to.matched[0].components.default()
  }
  
  // 预加载标记为preload的路由
  routes
    .filter(route => route.meta?.preload)
    .forEach(route => {
      if (typeof route.component === 'function') {
        route.component()
      }
    })
  
  next()
})

// 创建应用实例
const app = createApp(App)

// 使用插件
app.use(router)
app.use(createPinia())
app.mount('#app')