import './assets/main.css'

// main.ts 文件
import { createApp } from 'vue'
import App from './App.vue'
import {createPinia} from 'pinia'
// 创建 Pinia 实例
const pinia = createPinia()
// 创建 Vue 实例
const app = createApp(App)
// 挂载到 Vue 根实例
app.use(pinia)
app.mount('#app')
