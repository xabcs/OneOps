<template>
  <div class="behavior-container">
    <!-- 使用新的 PageHeader 组件 -->
    <PageHeader
      title="行为日志"
      subtitle="审计系统内所有用户的登录行为及业务操作记录"
    />

    <!-- 标签卡片 -->
    <el-card shadow="never" class="tabs-card">
      <el-tabs v-model="activeTab" class="behavior-tabs" @tab-click="handleTabClick">
        <el-tab-pane label="登录审计" name="login" />
        <el-tab-pane label="操作审计" name="operation" />
      </el-tabs>

      <div class="tab-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PageHeader } from '@/components'

const route = useRoute()
const router = useRouter()
const activeTab = ref('login')

const updateTabFromRoute = () => {
  const path = route.path
  if (path.includes('operation')) {
    activeTab.value = 'operation'
  } else {
    activeTab.value = 'login'
  }
}

const handleTabClick = (tab) => {
  router.push(`/audit/behavior/${tab.props.name}`)
}

watch(() => route.path, updateTabFromRoute)
</script>

<style scoped>
.behavior-container {
  display: flex;
  flex-direction: column;
  gap: 0; /* 移除 gap，因为 PageHeader 组件已经有了 margin-bottom */
}

.tabs-card {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-card__body) {
  padding: 0;
}

:deep(.behavior-tabs .el-tabs__header) {
  margin-bottom: 0;
  padding: 0 20px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
}

:deep(.behavior-tabs .el-tabs__nav-wrap::after) {
  display: none;
}

.tab-content {
  padding: 0;
}

.tab-content :deep(.logs-container) {
  padding: 0;
  border: none;
  border-radius: 0 0 8px 8px;
}

/* 隐藏子组件的页面头部 */
:deep(.logs-container .page-header) {
  display: none;
}
</style>
