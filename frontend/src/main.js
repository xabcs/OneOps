import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import store from './store'
import './mock'

const app = createApp(App)

// 全局错误处理，防止系统彻底崩溃
app.config.errorHandler = (err, vm, info) => {
  console.error('Vue Global Error:', err)
  console.error('Error Info:', info)
}

// 初始化认证状态
store.dispatch('initializeAuth')

app.use(router).use(ElementPlus).use(store).mount('#app')
