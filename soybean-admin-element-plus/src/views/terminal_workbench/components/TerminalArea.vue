<script setup lang="ts">
    import { Icon } from "@iconify/vue";
    import XTermTerminal from "./XTermTerminal.vue";
    import SessionList from "../session_list.vue";
    import { computed } from "vue";

    interface Props {
        activeSession: any;
        sessions: any[];
    }

    interface Emits {
        (e: "toggleFullscreen"): void;
    }

    const props = defineProps<Props>();
    defineEmits<Emits>();

    // 判断当前是否为会话列表视图
    const isSessionListView = computed(() => {
        return props.activeSession?.isSessionListView === true;
    });
</script>

<template>
    <div class="terminal-area">
        <!-- 会话列表视图 -->
        <div v-show="isSessionListView" class="session-list-view">
            <SessionList />
        </div>

        <!-- 终端视图 -->
        <div v-show="!isSessionListView" class="terminal-view-container">
            <!-- 终端工具栏 -->
            <div class="terminal-toolbar">
                <div class="toolbar-left">
                    <button class="toolbar-btn" title="监控">
                        <Icon icon="lucide:line-chart" class="btn-icon" />
                    </button>
                </div>
                <div class="toolbar-center">
                    <div class="terminal-mode">
                        <button class="mode-btn">Shell</button>
                        <button class="mode-btn mode-active">Agent</button>
                    </div>
                    <button class="toolbar-btn" title="帮助">
                        <Icon icon="lucide:circle-help" class="btn-icon" />
                    </button>
                </div>
                <div class="toolbar-right">
                    <button class="toolbar-btn" title="安全连接">
                        <Icon icon="lucide:shield-check" class="btn-icon" />
                    </button>
                </div>
            </div>

            <!-- 终端内容区 -->
            <div class="terminal-content">
                <div v-if="!activeSession" class="terminal-placeholder">
                    <Icon icon="lucide:terminal" class="placeholder-icon" />
                    <p class="placeholder-title">终端工作台</p>
                    <p class="placeholder-desc">请从左侧主机资产中选择主机进行连接</p>
                    <p class="placeholder-hint">当前会话数: {{ sessions.length }}</p>
                </div>

                <div v-else class="terminal-sessions">
                    <template v-for="session in sessions" :key="session.id">
                        <XTermTerminal
                            v-if="session.connected && session.status === 'connected'"
                            v-show="activeSession && session.id === activeSession.id"
                            :session-id="session.id"
                            :server-id="session.serverId"
                            :server-name="session.serverName"
                            :server-ip="session.serverIp"
                            :login-account="session.loginAccount"
                            :websocket-url="session.websocketUrl"
                        />
                    </template>
                </div>
            </div>
        </div>
    </div>
</template>

<style lang="scss" scoped>
.terminal-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: #171717;
    overflow: hidden;
    min-width: 0;
    min-height: 0;
    position: relative;
}

.session-list-view,
.terminal-view-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
    min-height: 0;
    position: absolute;
    inset: 0;
}

.terminal-view-container {
    background: #171717;
}

.terminal-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    height: 36px;
    background: #252526;
    border-bottom: 1px solid #000;
    flex-shrink: 0;
}

.toolbar-left,
.toolbar-center,
.toolbar-right {
    display: flex;
    align-items: center;
    gap: 4px;
}

.toolbar-btn {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    cursor: pointer;
    border-radius: 2px;
    color: #858585;

    &:hover {
        background: rgba(0, 0, 0, 0.2);
        color: #aaa;
    }
}

.btn-icon {
    width: 14px;
    height: 14px;
}

.terminal-mode {
    display: flex;
    background: #1e1e1e;
    border-radius: 2px;
    padding: 1px;
}

.mode-btn {
    padding: 2px 8px;
    font-size: 11px;
    background: transparent;
    border: none;
    color: #858585;
    cursor: pointer;
    border-radius: 1px;

    &:hover {
        color: #ccc;
    }

    &.mode-active {
        background: #007acc;
        color: #fff;
    }
}

.terminal-content {
    flex: 1;
    overflow: hidden;
    min-height: 0;
}

.terminal-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #858585;
    gap: 8px;
}

.placeholder-icon {
    width: 48px;
    height: 48px;
    opacity: 0.3;
}

.placeholder-title {
    font-size: 14px;
    font-weight: 500;
    color: #ccc;
    margin: 0;
}

.placeholder-desc {
    font-size: 12px;
    color: #6e6e6e;
    margin: 0;
}

.placeholder-hint {
    font-size: 11px;
    color: #4a4a4a;
    margin: 0;
}

.terminal-sessions {
    width: 100%;
    height: 100%;
    overflow: hidden;
}

.terminal-session {
    width: 100%;
    height: 100%;
    overflow: hidden;
}
</style>
