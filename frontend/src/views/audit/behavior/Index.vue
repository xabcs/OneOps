<template>
    <div class="behavior-container">
        <div class="content-wrapper">
            <el-tabs v-model="activeTab" class="behavior-tabs" @tab-click="handleTabClick">
                <el-tab-pane label="登录日志" name="login" />
                <el-tab-pane label="操作日志" name="operation" />
            </el-tabs>
            <div class="tab-content">
                <router-view v-slot="{ Component }">
                    <transition name="fade-transform" mode="out-in">
                        <component :is="Component" />
                    </transition>
                </router-view>
            </div>
        </div>
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
        gap: 0;
        background: var(--bg-primary);
        min-height: calc(100vh - 60px);
    }

    .content-wrapper {
        display: flex;
        flex-direction: column;
        flex: 1;
    }

    :deep(.behavior-tabs .el-tabs__header) {
        margin-bottom: 0;
        padding: 0 24px;
        background: var(--bg-primary);
        border-bottom: 1px solid var(--border);
    }

    :deep(.behavior-tabs .el-tabs__nav-wrap::after) {
        display: none;
    }

    .tab-content {
        padding: 0;
    }

    /* 覆盖子组件的 header */
    :deep(.logs-container .page-header) {
        display: none;
    }
</style>
