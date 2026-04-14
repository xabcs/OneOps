<template>
  <div class="login-page">
    <!-- Left Side: Decorative & Brand -->
    <div class="login-aside">
      <div class="aside-content">
        <div class="brand">
          <div class="logo-box">
            <el-icon :size="32" color="#fff"><Monitor /></el-icon>
          </div>
          <h1 class="brand-name">NexOps</h1>
        </div>
        <div class="hero-text">
          <h2>全栈自动化运维新标杆</h2>
          <p>NexOps 为企业提供高效、安全、智能的 IT 基础设施管理方案，助力业务数字化转型。</p>
        </div>
        <div class="aside-footer">
          <p>&copy; 2026 NexOps Team. All rights reserved.</p>
        </div>
      </div>
      <div class="aside-bg"></div>
    </div>

    <!-- Right Side: Login Form -->
    <div class="login-main">
      <div class="form-container">
        <div class="form-header">
          <h2>欢迎回来</h2>
          <p>请输入您的凭据以访问控制台</p>
        </div>

        <el-form 
          :model="loginForm" 
          :rules="rules" 
          ref="loginFormRef" 
          label-position="top" 
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item label="用户名" prop="username">
            <el-input 
              v-model="loginForm.username" 
              placeholder="请输入用户名"
              :prefix-icon="User"
              size="large"
            />
          </el-form-item>
          
          <el-form-item label="密码" prop="password">
            <el-input 
              v-model="loginForm.password" 
              type="password" 
              placeholder="请输入密码"
              :prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>

          <div class="form-options">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
            <el-button link type="primary">忘记密码？</el-button>
          </div>

          <el-button 
            type="primary" 
            class="submit-btn" 
            :loading="loading"
            @click="handleLogin"
            size="large"
          >
            {{ loading ? '正在登录...' : '立即登录' }}
          </el-button>
        </el-form>

        <div class="form-footer">
          <p>还没有账号？ <el-button link type="primary">联系管理员注册</el-button></p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { ElMessage } from 'element-plus'
import { User, Lock, Monitor } from '@element-plus/icons-vue'
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
      store.dispatch('login', {
        token: response.token,
        user: response.user
      })
      ElMessage.success('欢迎回来, ' + response.user.username)
      router.push('/')
    } else {
      ElMessage.error(response.message || '登录失败')
    }
  } catch (error) {
    console.error('Login error:', error)
    if (error.response) {
      ElMessage.error(error.response.data.message || '登录失败')
    } else {
      ElMessage.error('网络连接异常，请检查后端服务')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  background-color: #ffffff;
}

/* Left Side Styles */
.login-aside {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 60px;
  background-color: #0f172a;
  color: #ffffff;
  overflow: hidden;
}

.aside-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: 
    radial-gradient(circle at 20% 30%, rgba(37, 99, 235, 0.15) 0%, transparent 50%),
    radial-gradient(circle at 80% 70%, rgba(16, 185, 129, 0.1) 0%, transparent 50%);
  z-index: 1;
}

.aside-content {
  position: relative;
  z-index: 2;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.brand {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: auto;
}

.logo-box {
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.3);
}

.brand-name {
  font-size: 2.25rem;
  font-weight: 800;
  color: #ffffff;
  letter-spacing: -0.02em;
  margin: 0;
}

.hero-text {
  margin-bottom: auto;
  max-width: 480px;
}

.hero-text h2 {
  font-size: 3rem;
  font-weight: 800;
  line-height: 1.2;
  margin-bottom: 24px;
  background: linear-gradient(to bottom right, #ffffff, #94a3b8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-text p {
  font-size: 1.125rem;
  color: #94a3b8;
  line-height: 1.6;
}

.aside-footer p {
  font-size: 0.875rem;
  color: #475569;
}

/* Right Side Styles */
.login-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px;
}

.form-container {
  width: 100%;
  max-width: 400px;
}

.form-header {
  margin-bottom: 40px;
}

.form-header h2 {
  font-size: 2rem;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 8px;
}

.form-header p {
  color: #64748b;
  font-size: 1rem;
}

.login-form {
  margin-bottom: 32px;
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: #475569;
  padding-bottom: 8px;
}

:deep(.el-input__wrapper) {
  padding: 4px 12px;
  border-radius: 12px;
  box-shadow: 0 0 0 1px #e2e8f0 inset;
  transition: all 0.2s;
  background-color: #f8fafc;
}

:deep(.el-input__wrapper.is-focus) {
  background-color: #ffffff;
  box-shadow: 0 0 0 1px #2563eb inset, 0 0 0 4px rgba(37, 99, 235, 0.1) !important;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.submit-btn {
  width: 100%;
  height: 54px;
  border-radius: 12px;
  font-size: 1.125rem;
  font-weight: 600;
  background: #2563eb;
  border: none;
  transition: all 0.2s;
}

.submit-btn:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 10px 15px -3px rgba(37, 99, 235, 0.3);
}

.form-footer {
  text-align: center;
  color: #64748b;
  font-size: 0.875rem;
}

@media (max-width: 1024px) {
  .login-aside {
    display: none;
  }
  
  .login-main {
    background-color: #f8fafc;
  }
  
  .form-container {
    background: #ffffff;
    padding: 40px;
    border-radius: 24px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }
}

@media (max-width: 480px) {
  .login-main {
    padding: 20px;
  }
  
  .form-container {
    padding: 30px 20px;
    box-shadow: none;
    background: transparent;
  }
}
</style>
