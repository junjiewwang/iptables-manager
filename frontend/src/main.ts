import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'

import App from './App.vue'

import Tables from './views/Tables.vue'
import Topology from './views/Topology.vue'
import Interfaces from './views/Interfaces.vue'
import Logs from './views/Logs.vue'
import Login from './views/Login.vue'
import ChainTableView from './views/ChainTableView.vue'

const routes = [
  { path: '/login', component: Login },
  { path: '/', redirect: '/chain-table-view' },

  { path: '/tables', component: Tables },
  { path: '/topology', component: Topology },
  { path: '/interfaces', component: Interfaces },
  { path: '/chain-table-view', component: ChainTableView },
  { path: '/logs', component: Logs }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)

// 注册所有Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(router)
app.use(createPinia())
app.mount('#app')