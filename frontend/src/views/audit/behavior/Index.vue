<template>
    <div class="behavior-container">
        <header class="page-header">
            <div class="header-content">
                <div style="display: flex; align-items: center; gap: 12px">
                    <h2 class="page-title">行为日志</h2>
                    <span class="accent-dot"></span>
                </div>
                <p class="page-subtitle">审计系统内所有用户的登录行为及业务操作记录。</p>
            </div>
        </header>

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
    import { ref, onMounted, watch } from 'vue'
    import { useRoute, useRouter } from 'vue-router'

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

    onMounted(() => {
        updateTabFromRoute()
    })
</script>

<style scoped>
    .behavior-container {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .tabs-card {
      border: 1px solid var(--border);
      border-radius: 2px;
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
      padding: 20px;
    }

    /* 覆盖子组件的 header */
    :deep(.logs-container .page-header) {
      display: none;
    }
</style>
