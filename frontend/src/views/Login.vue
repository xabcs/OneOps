<template>
  <div class="login-page">
    <!-- Left Decorative Column -->
    <div class="visual-column">
      <div class="gradient-overlay"></div>
      <div class="pattern-overlay"></div>
      
      <!-- Curve Separator -->
      <div class="curve-divider">
        <svg preserveAspectRatio="none" viewBox="0 0 100 100" class="curve-svg">
          <path d="M100,0 C30,25 30,75 100,100 Z" fill="#ffffff"></path>
        </svg>
      </div>

      <div class="visual-content">
        <div class="brand-logo">
          <img src="/logo.svg" alt="NexOps" class="logo-image" />
          <span class="brand-name">NexOps</span>
        </div>
        
        <div class="hero-text-area">
          <h1 class="hero-title">
            高效运维<br />
            智领未来
          </h1>
        </div>
        
        <div class="visual-footer">
          <p class="copyright">&copy; 2026 NexOps Team.</p>
        </div>
      </div>
    </div>

    <!-- Right Form Column -->
    <div class="form-column">
      <div class="form-wrapper">
        <header class="form-intro">
          <h2 class="welcome-title">欢迎使用</h2>
          <p class="welcome-sub">请登录您的账号以继续</p>
        </header>

        <el-form 
          :model="loginForm" 
          :rules="rules" 
          ref="loginFormRef" 
          label-position="top" 
          class="login-main-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item label="用户名" prop="username">
            <el-input 
              v-model="loginForm.username" 
              placeholder="请输入您的用户名"
              :prefix-icon="User"
              size="large"
            />
          </el-form-item>
          
          <el-form-item label="密码" prop="password">
            <el-input 
              v-model="loginForm.password" 
              type="password" 
              placeholder="请输入您的密码"
              :prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>

          <div class="form-actions-meta">
            <el-checkbox v-model="rememberMe">记住登录状态</el-checkbox>
            <el-button link type="primary" class="forgot-btn">忘记密码？</el-button>
          </div>

          <el-button 
            type="primary" 
            class="submit-action-btn" 
            :loading="loading"
            @click="handleLogin"
          >
            {{ loading ? '验证中...' : '立即登录' }}
          </el-button>
        </el-form>

        <footer class="form-extras">
          <p>还没有账号？ <el-button link class="contact-admin">联系管理员进行注册</el-button></p>
        </footer>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { loginApi } from '../api/index.js'

const router = useRouter()
const store = useStore()
const loginFormRef = ref(null)
const loading = ref(false)
const rememberMe = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  try {
    await loginFormRef.value.validate()
    loading.value = true
    
    const response = await loginApi.login(loginForm)
    
    if (response.success) {
      store.commit('SET_TOKEN', response.token)
      store.commit('SET_USER', response.user)
      store.commit('SET_PERMISSIONS', response.user.permissions || [])
      store.commit('SET_MENU_TREE', response.user.menuTree || [])
      
      ElMessage.success('欢迎回来, ' + response.user.username)
      router.push(response.user.homePath || '/')
    } else {
       ElMessage.error(response.message || '用户名或密码错误')
    }
  } catch (error) {
    console.error('Login error:', error)
    ElMessage.error('无法连接到系统服务，请稍后再试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  background-color: #fff;
}

/* Left Column - Visuals */
.visual-column {
  flex: 1.1;
  position: relative;
  background: linear-gradient(135deg, #334155 0%, #1e293b 100%);
  display: flex;
  flex-direction: column;
  padding: 60px;
  color: #fff;
  overflow: hidden;
}

.curve-divider {
  position: absolute;
  top: 0;
  right: -1px;
  bottom: 0;
  width: 120px;
  z-index: 5;
}

.curve-svg {
  width: 100%;
  height: 100%;
}

.gradient-overlay {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 10% 10%, rgba(255, 255, 255, 0.1) 0%, transparent 40%);
}

.pattern-overlay {
  position: absolute;
  inset: 0;
  opacity: 0.1;
  background-image: radial-gradient(#fff 1px, transparent 1px);
  background-size: 30px 30px;
}

.visual-content {
  position: relative;
  z-index: 10;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: auto;
}

.logo-image {
  height: 40px;
  width: auto;
}

.brand-name {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.hero-text-area {
  margin-bottom: auto;
}

.hero-title {
  font-size: clamp(2.5rem, 4.5vw, 3.5rem);
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 24px;
}

.hero-sub {
  font-size: 1.125rem;
  line-height: 1.7;
  max-width: 480px;
  opacity: 0.9;
  font-weight: 300;
}

.visual-footer {
  font-size: 0.875rem;
  opacity: 0.6;
}

/* Right Column - Form */
.form-column {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background-color: #fff;
}

.form-wrapper {
  width: 100%;
  max-width: 400px;
  animation: slide-in-fade 0.8s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes slide-in-fade {
  from { opacity: 0; transform: translateX(20px); }
  to { opacity: 1; transform: translateX(0); }
}

.form-intro {
  margin-bottom: 40px;
}

.welcome-title {
  font-size: 2.25rem;
  font-weight: 700;
  color: #111827;
  margin-bottom: 8px;
}

.welcome-sub {
  color: #6b7280;
  font-size: 1rem;
}

.login-main-form {
  margin-bottom: 32px;
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: #374151;
  padding-bottom: 8px;
}

:deep(.el-input__wrapper) {
  padding: 8px 16px;
  border-radius: 12px;
  box-shadow: 0 0 0 1px #e5e7eb inset;
  transition: all 0.2s;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px #4f46e5 inset, 0 0 0 4px rgba(79, 70, 229, 0.1) !important;
}

.form-actions-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.forgot-btn {
  font-size: 13px;
  color: #6b7280;
}

.submit-action-btn {
  width: 100%;
  height: 52px;
  font-size: 1.125rem;
  font-weight: 600;
  border-radius: 14px;
  background-color: #4f46e5;
  border: none;
  box-shadow: 0 4px 6px -1px rgba(79, 70, 229, 0.2), 0 2px 4px -1px rgba(79, 70, 229, 0.1);
  transition: all 0.2s;
}

.submit-action-btn:hover {
  background-color: #4338ca;
  transform: translateY(-1px);
  box-shadow: 0 10px 15px -3px rgba(79, 70, 229, 0.3);
}

.form-extras {
  text-align: center;
  font-size: 0.875rem;
  color: #6b7280;
}

.contact-admin {
  font-weight: 600;
}

/* Responsiveness */
@media (max-width: 1024px) {
  .visual-column {
    display: none;
  }
  
  .form-column {
    background-color: #f9fafb;
  }
  
  .form-wrapper {
    background: #fff;
    padding: 48px;
    border-radius: 24px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
}

@media (max-width: 480px) {
  .form-column {
    padding: 20px;
  }
  
  .form-wrapper {
    padding: 32px 20px;
    box-shadow: none;
    border-radius: 0;
    background: transparent;
  }
  
  .welcome-title {
    font-size: 1.75rem;
  }
}
</style>
