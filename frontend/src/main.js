import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import store from './store'
import './mock'

const app = createApp(App)

// 全局错误处理，防止系统彻底崩溃
app.config.errorHandler = (err, vm, info) => {
  const errorMessage = err instanceof Error ? err.message : String(err)
  console.error('Vue Global Error:', errorMessage)
  // 避免打印完整的 vm 或 info 对象，因为它们可能包含循环引用
  console.error('Error Context:', String(info))
}

// 全局警告处理
app.config.warnHandler = (msg, vm, trace) => {
  // 仅打印消息和追踪，避免打印完整的组件实例
  console.warn('Vue Warning:', msg)
  if (trace) console.warn('Vue Warning Trace:', trace)
}

// 初始化认证状态
store.dispatch('initializeAuth')

app.use(router).use(ElementPlus, { locale: zhCn }).use(store).mount('#app')
