<script setup>
    import { ref, computed, watch, onMounted } from 'vue'
    import { useRouter, useRoute } from 'vue-router'
    import { useStore } from 'vuex'
    import * as ElementPlusIconsVue from '@element-plus/icons-vue'
    import { 
        House, Monitor, Timer, DataLine, User, SwitchButton, 
        Expand, Fold, Bell, Search, Setting, QuestionFilled,
        Close, Moon, Sunny, Menu, UserFilled, Plus
    } from '@element-plus/icons-vue'
    import { loginApi } from './api/index.js'

    const router = useRouter()
    const route = useRoute()
    const store = useStore()

    const isAuthenticated = computed(() => store.getters.isAuthenticated)
    const user = computed(() => store.getters.user)
    const permissions = computed(() => store.getters.permissions)
    const currentTheme = computed(() => store.getters.currentTheme)
    const isCollapse = ref(false)

    const fetchUserInfo = async () => {
        if (!isAuthenticated.value) return
        try {
            const res = await loginApi.getUserInfo()
            if (res.code === 200) {
                const oldUserId = user.value?.id
                store.commit('SET_USER', res.data)
                store.commit('SET_PERMISSIONS', res.data.permissions)
                store.commit('SET_MENU_TREE', res.data.menuTree)

                // 用户切换或登录时，清理无权限的标签页
                if (res.data.id && res.data.id !== oldUserId) {
                    cleanInvalidTabs()
                }
            }
        } catch (error) {
            console.error('Error fetching user info:', error)
        }
    }

    onMounted(fetchUserInfo)

    watch(isAuthenticated, (val) => {
        if (val) {
            fetchUserInfo()
        } else {
            // 登出时，重置标签页为首页
            visitedViews.value = [
                { name: 'Home', path: '/', title: '控制台', closable: false }
            ]
        }
    })

    const menuItems = computed(() => {
        const rawMenuTree = store.getters.menuTree || []
        
        // 将图标字符串映射为组件
        const mapIcons = (list) => {
            return list.map(item => {
                const node = { ...item }
                if (typeof item.icon === 'string') {
                    node.icon = ElementPlusIconsVue[item.icon] || item.icon
                }
                if (item.children) {
                    node.children = mapIcons(item.children)
                }
                return node
            })
        }

        return mapIcons(rawMenuTree)
    })

    const toggleTheme = (event) => {
        const isAppearanceTransition = document.startViewTransition &&
            !window.matchMedia('(prefers-reduced-motion: reduce)').matches

        if (!isAppearanceTransition) {
            const newTheme = currentTheme.value === 'light' ? 'dark' : 'light'
            store.commit('SET_THEME', newTheme)
            return
        }

        const x = event.clientX
        const y = event.clientY
        const endRadius = Math.hypot(
            Math.max(x, innerWidth - x),
            Math.max(y, innerHeight - y)
        )

        const transition = document.startViewTransition(() => {
            const newTheme = currentTheme.value === 'light' ? 'dark' : 'light'
            store.commit('SET_THEME', newTheme)
        })

        transition.ready.then(() => {
            const clipPath = [
                `circle(0px at ${x}px ${y}px)`,
                `circle(${endRadius}px at ${x}px ${y}px)`,
            ]
            document.documentElement.animate(
                {
                    clipPath: clipPath,
                },
                {
                    duration: 500,
                    easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
                    pseudoElement: '::view-transition-new(root)',
                }
            )
        })
    }

    // Tabs logic
    const visitedViews = ref([
        { name: 'Home', path: '/', title: '控制台', closable: false }
    ])

    const activeTab = computed({
        get: () => route.path,
        set: (val) => {
            router.push(val)
        }
    })

    const addTab = (route) => {
        if (!route.name || route.name === 'Login') return

        // 检查用户是否有访问该路由的权限
        if (route.meta.permission) {
            const hasPermission = permissions.value.includes('*:*:*') || permissions.value.includes(route.meta.permission)
            if (!hasPermission) {
                console.warn(`无权访问标签页: ${route.path}`)
                return
            }
        }

        const isExist = visitedViews.value.some(v => v.path === route.path)
        if (!isExist) {
            visitedViews.value.push({
                name: route.name,
                path: route.path,
                title: route.meta.title || route.name,
                closable: route.path !== '/'
            })
        }
    }

    const removeTab = (targetPath) => {
        const tabs = visitedViews.value
        let activePath = activeTab.value
        if (activePath === targetPath) {
            tabs.forEach((tab, index) => {
                if (tab.path === targetPath) {
                    const nextTab = tabs[index + 1] || tabs[index - 1]
                    if (nextTab) {
                        activePath = nextTab.path
                    }
                }
            })
        }
        activeTab.value = activePath
        visitedViews.value = tabs.filter(tab => tab.path !== targetPath)
    }

    // 清理无权限的标签页
    const cleanInvalidTabs = () => {
        visitedViews.value = visitedViews.value.filter(tab => {
            // 首页始终保留
            if (tab.path === '/') return true

            // 对于其他标签页，检查权限
            // 这里我们需要找到对应的路由配置
            const matchedRoute = router.resolve(tab.path).matched.find(r => r.meta.permission)
            if (!matchedRoute) return true

            const hasPermission = permissions.value.includes('*:*:*') || permissions.value.includes(matchedRoute.meta.permission)
            return hasPermission
        })

        // 确保至少保留首页
        if (visitedViews.value.length === 0 || !visitedViews.value.some(t => t.path === '/')) {
            visitedViews.value = [
                { name: 'Home', path: '/', title: '控制台', closable: false }
            ]
        }
    }

    watch(() => route.path, () => {
        addTab(route)
    }, { immediate: true })

    const activeMenu = computed(() => {
        return route.path
    })

    const handleLogout = () => {
        store.dispatch('logout')
        router.push('/login')
    }

    const toggleSidebar = () => {
        isCollapse.value = !isCollapse.value
    }
</script>

<template>
    <div v-if="isAuthenticated" class="app-wrapper">
        <el-container class="app-container">
            <!-- Sidebar -->
            <el-aside :width="isCollapse ? '64px' : '240px'" class="app-sidebar">
                <div class="logo-container">
                    <img src="/logo.svg" alt="NexOps" class="logo-img" :class="{ 'collapsed': isCollapse }" />
                </div>
                
                <el-scrollbar>
                    <el-menu
                        :default-active="activeMenu"
                        class="sidebar-menu"
                        :collapse="isCollapse"
                        router
                    >
                        <template v-for="item in menuItems" :key="item.path">
                            <template v-if="item.children && item.children.length > 0">
                                <el-sub-menu :index="item.path">
                                    <template #title>
                                        <el-icon><component :is="item.icon" /></el-icon>
                                        <span>{{ item.name }}</span>
                                    </template>
                                    <template v-for="child in item.children" :key="child.path">
                                        <el-sub-menu v-if="child.children && child.children.length > 0" :index="child.path">
                                            <template #title>
                                                <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
                                                <span>{{ child.name }}</span>
                                            </template>
                                            <el-menu-item v-for="grandChild in child.children" :key="grandChild.path" :index="grandChild.path">
                                                <el-icon v-if="grandChild.icon"><component :is="grandChild.icon" /></el-icon>
                                                {{ grandChild.name }}
                                            </el-menu-item>
                                        </el-sub-menu>
                                        <el-menu-item v-else :index="child.path">
                                            <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
                                            {{ child.name }}
                                        </el-menu-item>
                                    </template>
                                </el-sub-menu>
                            </template>
                            <el-menu-item v-else :index="item.path">
                                <el-icon><component :is="item.icon" /></el-icon>
                                <template #title>{{ item.name }}</template>
                            </el-menu-item>
                        </template>
                    </el-menu>
                </el-scrollbar>

                <div class="sidebar-footer" v-show="!isCollapse">
                    <el-button link :icon="QuestionFilled">帮助文档</el-button>
                </div>
            </el-aside>

            <el-container class="main-container">
                <!-- Header -->
                <el-header class="app-header">
                    <div class="header-left">
                        <el-button link @click="toggleSidebar" class="collapse-btn">
                            <el-icon :size="20">
                                <component :is="isCollapse ? Expand : Fold" />
                            </el-icon>
                        </el-button>
                        <el-breadcrumb separator="/" class="breadcrumb">
                            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
                            <el-breadcrumb-item v-if="route.meta.parent">{{ route.meta.parent }}</el-breadcrumb-item>
                            <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title }}</el-breadcrumb-item>
                        </el-breadcrumb>
                    </div>

                    <div class="header-right">
                        <div class="header-search">
                            <el-input
                                placeholder="搜索资源、任务..."
                                :prefix-icon="Search"
                                size="default"
                                clearable
                            />
                        </div>

                        <el-button 
                            link 
                            :icon="currentTheme === 'light' ? Moon : Sunny" 
                            class="header-icon-btn"
                            @click="toggleTheme($event)"
                        />
                        
                        <el-badge :value="3" class="notice-badge" is-dot>
                            <el-button link :icon="Bell" class="header-icon-btn" />
                        </el-badge>

                        <el-divider direction="vertical" />

                        <el-dropdown trigger="click">
                            <div class="user-profile">
                                <el-avatar :size="32" class="user-avatar">
                                    {{ user?.username?.charAt(0).toUpperCase() }}
                                </el-avatar>
                                <div class="user-meta">
                                    <div style="display: flex; align-items: center; gap: 6px">
                                        <span class="user-name">{{ user?.username }}</span>
                                        <span class="accent-dot" style="width: 4px; height: 4px; box-shadow: 0 0 4px var(--accent);"></span>
                                    </div>
                                    <span class="user-role">{{ user?.roleNames?.join(', ') || '普通用户' }}</span>
                                </div>
                            </div>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item :icon="User">个人中心</el-dropdown-item>
                                    <el-dropdown-item :icon="Setting">账号设置</el-dropdown-item>
                                    <el-dropdown-item divided :icon="SwitchButton" @click="handleLogout" class="logout-item">
                                        退出登录
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                </el-header>

                <!-- Tabs View -->
                <div class="app-tabs-container">
                    <el-tabs
                        v-model="activeTab"
                        type="card"
                        class="app-tabs"
                        @tab-remove="removeTab"
                    >
                        <el-tab-pane
                            v-for="item in visitedViews"
                            :key="item.path"
                            :label="item.title"
                            :name="item.path"
                            :closable="item.closable"
                        />
                    </el-tabs>
                </div>

                <!-- Main Content -->
                <el-main class="app-main">
                    <router-view v-slot="{ Component }">
                        <transition name="fade-transform" mode="out-in">
                            <component :is="Component" />
                        </transition>
                    </router-view>
                </el-main>
            </el-container>
        </el-container>
    </div>
    <div v-else class="auth-wrapper">
        <router-view></router-view>
    </div>
</template>

<style scoped>
    .app-wrapper {
        width: 100%;
        height: 100vh;
        overflow: hidden;
    }

    .app-container {
        height: 100%;
    }

    /* Sidebar Styles */
    .app-sidebar {
        background-color: var(--bg-sidebar);
        display: flex;
        flex-direction: column;
        transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        border-right: 1px solid var(--sidebar-border);
        z-index: 100;
        box-shadow: 4px 0 24px rgba(0, 0, 0, 0.02);
    }

    .logo-container {
        height: 80px;
        display: flex;
        align-items: center;
        padding-left: 32px;
        background: var(--bg-sidebar);
        position: relative;
    }

    .logo-container::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 24px;
        right: 24px;
        height: 1px;
        background: linear-gradient(to right, var(--border), transparent);
        opacity: 0.5;
    }

    .logo-img {
        height: 36px;
        width: auto;
        display: block;
        transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        filter: drop-shadow(0 4px 8px rgba(0, 102, 255, 0.1));
    }

    .logo-img.collapsed {
        height: 32px;
        width: 32px;
        object-fit: cover;
        object-position: 0 50%;
        margin-left: -4px;
    }

    .sidebar-menu {
        border-right: none !important;
        background-color: transparent !important;
        padding-top: 8px;
    }

    :deep(.el-menu) {
        border-right: none;
    }

    :deep(.el-menu-item), :deep(.el-sub-menu__title) {
        color: var(--sidebar-text) !important;
        height: 48px !important;
        line-height: 48px !important;
        margin: 4px 12px !important;
        border-radius: 12px !important;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
    }

    :deep(.el-menu-item:hover), :deep(.el-sub-menu__title:hover) {
        background-color: var(--sidebar-hover) !important;
        color: var(--primary) !important;
        transform: translateX(4px);
    }

    :deep(.el-menu-item.is-active) {
        background-color: var(--sidebar-active-bg) !important;
        color: var(--sidebar-active-text) !important;
        font-weight: 600;
        box-shadow: 0 4px 12px rgba(0, 102, 255, 0.08);
    }

    :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
        color: var(--sidebar-active-text) !important;
        font-weight: 600;
    }

    /* 二级菜单背景及样式优化 */
    :deep(.el-sub-menu .el-menu) {
        background-color: var(--sidebar-submenu-bg) !important;
        border: none;
    }

    :deep(.el-sub-menu .el-menu-item) {
        padding-left: 48px !important;
        font-size: 12px;
    }

    :deep(.el-sub-menu .el-menu-item.is-active) {
        background-color: var(--sidebar-active-bg) !important;
        color: var(--sidebar-active-text) !important;
    }

    .sidebar-footer {
        padding: 12px 16px;
        border-top: 1px solid var(--sidebar-border);
    }

    /* Header Styles */
    .app-header {
        background-color: var(--bg-primary);
        border-bottom: 1px solid var(--border);
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0 20px;
        height: 52px !important;
        flex-shrink: 0;
    }

    .app-tabs-container {
        background-color: var(--bg-primary);
        padding: 0 20px;
        border-bottom: 1px solid var(--border);
        flex-shrink: 0;
        height: 34px;
        display: flex;
        align-items: center;
    }

    :deep(.el-tabs--card > .el-tabs__header) {
        border-bottom: none;
        margin: 0;
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__nav) {
        border: none;
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__item) {
        border: none;
        height: 34px;
        line-height: 34px;
        font-size: 12px;
        color: var(--text-secondary);
        padding: 0 16px !important;
        transition: all 0.2s;
        position: relative;
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__item::after) {
        content: "";
        position: absolute;
        bottom: 0;
        left: 20px;
        right: 20px;
        height: 2px;
        background-color: transparent;
        transition: all 0.2s;
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__item:hover) {
        color: var(--primary);
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
        color: var(--primary);
        font-weight: 600;
    }

    :deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active::after) {
        background-color: var(--primary);
    }

    :deep(.el-tabs__item .el-icon-close) {
        width: 12px;
        height: 12px;
        margin-left: 6px;
    }

    .header-left {
        display: flex;
        align-items: center;
        gap: 16px;
    }

    .breadcrumb {
        font-size: 13px;
    }

    :deep(.el-breadcrumb__inner) {
        color: var(--text-tertiary) !important;
        font-weight: 400 !important;
    }

    :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
        color: var(--text-primary) !important;
        font-weight: 500 !important;
    }

    .collapse-btn {
        color: var(--text-secondary);
    }

    .header-right {
        display: flex;
        align-items: center;
        gap: 20px;
    }

    .header-search {
        width: 240px;
    }

    :deep(.el-input__wrapper) {
        background-color: var(--bg-tertiary);
        box-shadow: none !important;
        border-radius: 8px;
    }

    .header-icon-btn {
        font-size: 20px;
        color: var(--text-secondary);
    }

    .user-profile {
        display: flex;
        align-items: center;
        gap: 12px;
        cursor: pointer;
        padding: 4px 8px;
        border-radius: 8px;
        transition: background 0.2s;
    }

    .user-profile:hover {
        background-color: var(--bg-tertiary);
    }

    .user-avatar {
        background-color: #2563eb;
        font-weight: 700;
    }

    .user-meta {
        display: flex;
        flex-direction: column;
        line-height: 1.2;
    }

    .user-name {
        font-size: 0.875rem;
        font-weight: 600;
        color: var(--text-primary);
    }

    .user-role {
        font-size: 0.75rem;
        color: var(--text-tertiary);
    }

    .logout-item {
        color: #ef4444 !important;
    }

    /* Main Content Styles */
    .app-main {
        background-color: var(--bg-secondary);
        padding: 12px;
        overflow-y: auto;
    }

    /* Transitions */
    .fade-transform-enter-active,
    .fade-transform-leave-active {
        transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
    }

    .fade-transform-enter-from {
        opacity: 0;
        transform: translateX(-20px);
    }

    .fade-transform-leave-to {
        opacity: 0;
        transform: translateX(20px);
    }

    .auth-wrapper {
        background-color: #f8fafc;
    }
</style>
