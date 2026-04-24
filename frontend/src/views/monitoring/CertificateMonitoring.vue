<template>
    <div class="certificate-monitoring-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">证书监控</h2>
                <p class="page-subtitle">实时监控DNS证书状态，及时发现即将过期或异常的证书。</p>
            </div>
            <div class="header-actions">
                <el-button :icon="FullScreen" @click="toggleFullscreen">{{ isFullscreen ? '退出全屏' : '全屏显示' }}</el-button>
            </div>
        </header>

        <div class="dashboard-wrapper" :class="{ 'fullscreen': isFullscreen }">
            <div class="iframe-container">
                <iframe
                    ref="grafanaIframe"
                    :src="grafanaUrl"
                    class="grafana-iframe"
                    frameborder="0"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                    allowfullscreen
                    @load="onIframeLoad"
                ></iframe>
                <div v-if="loading" class="loading-overlay">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>正在加载Grafana面板...</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { FullScreen, Loading } from '@element-plus/icons-vue'

// Grafana配置 - 使用kiosk模式（全屏无导航栏，无需登录）
const grafanaConfig = {
    baseUrl: 'http://192.168.4.168:3000',
    dashboardUrl: '/d/fdwecevaqo7wge/cloud-dns-record-info',
    orgId: 1,
    // kiosk模式参数：隐藏UI，但保留过滤器
    params: {
        kiosk: true,           // kiosk模式（隐藏顶部/侧边栏，保留过滤器）
        autofitpanels: 'true', // 自动调整面板大小
        refresh: '30s',        // 自动刷新间隔
        showTimestamp: 'true'  // 显示时间戳
    }
}

// 构建完整的Grafana URL（使用kiosk模式，无需登录）
const buildGrafanaUrl = () => {
    const params = new URLSearchParams()
    params.append('orgId', grafanaConfig.orgId.toString())

    // kiosk 参数（无值，保留过滤器）
    params.append('kiosk', '')

    // 添加其他参数
    if (grafanaConfig.params.autofitpanels) {
        params.append('autofitpanels', grafanaConfig.params.autofitpanels)
    }
    if (grafanaConfig.params.refresh) {
        params.append('refresh', grafanaConfig.params.refresh)
    }
    if (grafanaConfig.params.showTimestamp) {
        params.append('showTimestamp', grafanaConfig.params.showTimestamp)
    }

    return `${grafanaConfig.baseUrl}${grafanaConfig.dashboardUrl}?${params.toString()}`
}

const grafanaUrl = ref(buildGrafanaUrl())

// 状态管理
const loading = ref(true)
const isFullscreen = ref(false)
const grafanaIframe = ref(null)

// iframe加载完成
const onIframeLoad = () => {
    loading.value = false
}

// 全屏切换
const toggleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value

    if (isFullscreen.value) {
        // 进入全屏
        const element = document.querySelector('.dashboard-wrapper')
        if (element.requestFullscreen) {
            element.requestFullscreen()
        } else if (element.webkitRequestFullscreen) {
            element.webkitRequestFullscreen()
        } else if (element.mozRequestFullScreen) {
            element.mozRequestFullScreen()
        }
    } else {
        // 退出全屏
        if (document.exitFullscreen) {
            document.exitFullscreen()
        } else if (document.webkitExitFullscreen) {
            document.webkitExitFullscreen()
        } else if (document.mozCancelFullScreen) {
            document.mozCancelFullScreen()
        }
    }
}

// 监听全屏变化事件
const handleFullscreenChange = () => {
    isFullscreen.value = !!(
        document.fullscreenElement ||
        document.webkitFullscreenElement ||
        document.mozFullScreenElement
    )
}

onMounted(() => {
    // 添加全屏变化监听
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
    document.addEventListener('mozfullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
    // 移除事件监听
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
    document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
    document.removeEventListener('mozfullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
.certificate-monitoring-container {
    display: flex;
    flex-direction: column;
    gap: 0;
    height: calc(100vh - 130px); /* 计算视口高度，减去头部和tabs的高度 */
    min-height: 600px; /* 最小高度 */
    position: relative;
    z-index: 1; /* 确保不被侧边栏遮挡 */
    margin: -4px; /* 抵消全局的 app-main padding */
    width: calc(100% + 8px); /* 补偿负边距造成的宽度损失 */
}

.page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-shrink: 0;
    padding: 12px 16px;
    background: var(--bg-primary);
    border-radius: 8px 8px 0 0; /* 只保留顶部圆角 */
    margin-bottom: 0;
}

.header-content {
    flex: 1;
}

.page-title {
    font-size: 20px;
    font-weight: 700;
    font-style: normal;
    color: var(--text-primary);
    letter-spacing: 0;
    margin: 0;
    line-height: 1.2;
}

.page-subtitle {
    font-size: 12px;
    color: var(--text-secondary);
    margin: 4px 0 0 0;
    line-height: 1.4;
}

.header-actions {
    display: flex;
    gap: 8px;
}

.dashboard-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--bg-primary);
    border-radius: 0 0 8px 8px; /* 只保留底部圆角 */
    overflow: hidden;
    margin-top: 0;
}

.dashboard-wrapper.fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 9999;
    background: var(--bg-primary);
    padding: 20px;
    border-radius: 0;
}

.iframe-container {
    position: relative;
    width: 100%;
    height: 100%;
}

.grafana-iframe {
    width: 100%;
    height: 100%;
    border: none;
    display: block;
}

.loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    font-size: 16px;
    color: var(--text-secondary);
}

.loading-overlay .el-icon {
    font-size: 32px;
    color: var(--primary);
}
</style>
