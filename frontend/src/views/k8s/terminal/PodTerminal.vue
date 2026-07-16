<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { ElButton, ElMessage } from 'element-plus';
import { useAuthStore } from '@/store/modules/auth';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

interface Props {
  clusterId: number;
  namespace: string;
  podName: string;
  containerName?: string;
}

const props = defineProps<Props>();

const message = ElMessage;
const authStore = useAuthStore();

const terminalContainer = ref<HTMLElement | null>(null);
const connected = ref(false);
const ws = ref<WebSocket | null>(null);
const terminal = ref<Terminal | null>(null);
const fitAddon = ref<FitAddon | null>(null);
const isComponentMounted = ref(true);

// 监听组件卸载
watch(() => isComponentMounted.value, (newVal) => {
  if (!newVal && ws.value) {
    ws.value.close();
  }
});

const connectTerminal = () => {
  if (!isComponentMounted.value || !terminalContainer.value) return;

  const token = authStore.token;

  // 构造 WebSocket URL
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = import.meta.env.VITE_SERVICE_BASE_URL?.replace(/^https?:\/\//, '').replace(/\/api$/, '') || window.location.host;
  const wsUrl = `${protocol}//${host}/api/k8s/terminal/ws?clusterId=${props.clusterId}&namespace=${props.namespace}&podName=${props.podName}&containerName=${props.containerName || ''}&token=${token}`;

  console.log('Connecting to WebSocket:', wsUrl);

  // 初始化 xterm.js
  terminal.value = new Terminal({
    fontSize: 14,
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", Consolas, "source-code-pro", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#ffffff',
      cursor: '#ffffff',
      cursorAccent: '#000000',
      selection: 'rgba(255, 255, 255, 0.3)',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#ffffff'
    },
    cursorBlink: true,
    cursorStyle: 'block',
    scrollback: 1000,
    allowProposedApi: true
  });

  // 初始化 FitAddon
  fitAddon.value = new FitAddon();
  terminal.value.loadAddon(fitAddon.value);

  // 打开终端
  terminal.value.open(terminalContainer.value);
  fitAddon.value.fit();

  // 监听终端输入
  terminal.value.onData((data) => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'stdin',
        data: data
      }));
    }
  });

  // 监听窗口大小变化
  const handleResize = () => {
    if (fitAddon.value && terminal.value) {
      fitAddon.value.fit();

      // 发送终端大小到后端
      if (ws.value && ws.value.readyState === WebSocket.OPEN) {
        ws.value.send(JSON.stringify({
          type: 'resize',
          cols: terminal.value.cols,
          rows: terminal.value.rows
        }));
      }
    }
  };

  window.addEventListener('resize', handleResize);

  try {
    ws.value = new WebSocket(wsUrl);

    ws.value.onopen = () => {
      if (!isComponentMounted.value) {
        ws.value?.close();
        return;
      }
      connected.value = true;
      terminal.value?.writeln('\x1b[32m✓ 已连接到 Pod 终端\x1b[0m');
      terminal.value?.writeln(`\x1b[36mPod: ${props.namespace}/${props.podName}\x1b[0m`);
      if (props.containerName) {
        terminal.value?.writeln(`\x1b[36mContainer: ${props.containerName}\x1b[0m`);
      }
      terminal.value?.writeln('\x1b[33m提示: 点击上方"设置提示符"按钮显示当前目录\x1b[0m');
      terminal.value?.writeln('');

      // 发送初始终端大小
      if (terminal.value) {
        ws.value?.send(JSON.stringify({
          type: 'resize',
          cols: terminal.value.cols,
          rows: terminal.value.rows
        }));
      }
    };

    ws.value.onmessage = event => {
      if (!isComponentMounted.value) return;
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output') {
          terminal.value?.write(msg.data);
        }
      } catch (error) {
        if (!isComponentMounted.value) return;
        terminal.value?.write(event.data);
      }
    };

    ws.value.onerror = error => {
      console.error('WebSocket error:', error);
      if (!isComponentMounted.value) return;
      terminal.value?.writeln('\x1b[31m✗ 终端连接错误\x1b[0m');
      message.error('终端连接错误');
    };

    ws.value.onclose = (event) => {
      if (!isComponentMounted.value) return;
      connected.value = false;
      if (event.wasClean) {
        terminal.value?.writeln(`\x1b[33m连接已关闭 (code: ${event.code})\x1b[0m`);
      } else {
        terminal.value?.writeln('\x1b[31m连接异常关闭\x1b[0m');
      }
      window.removeEventListener('resize', handleResize);
    };
  } catch (error: any) {
    if (!isComponentMounted.value) return;
    terminal.value?.writeln(`\x1b[31m✗ 连接失败: ${error.message}\x1b[0m`);
    message.error(error.message || '连接失败');
  }
};

const handleDisconnect = () => {
  if (ws.value) {
    ws.value.close();
  }
};

// 设置友好的提示符
const setFriendlyPrompt = () => {
  if (ws.value && ws.value.readyState === WebSocket.OPEN) {
    // 一次性发送完整的命令，使用分号分隔避免多行输入
    const cmd = 'if [ -n "$BASH_VERSION" ]; then export PS1=\'[\\u@\\h \\w]\\$ \'; else export PS1=\'[$(pwd)]\\$ \'; fi\n';
    ws.value.send(JSON.stringify({
      type: 'stdin',
      data: cmd
    }));
  }
};

onMounted(() => {
  isComponentMounted.value = true;
  connectTerminal();
});

onUnmounted(() => {
  isComponentMounted.value = false;
  if (ws.value) {
    ws.value.close();
  }
  if (terminal.value) {
    terminal.value.dispose();
  }
});
</script>

<template>
  <div class="h-full flex flex-col bg-gray-900">
    <!-- 终端标题栏 -->
    <div class="flex items-center justify-between border-b border-gray-700 bg-gray-800 px-4 py-2">
      <div class="flex items-center space-x-3">
        <div class="flex items-center space-x-2">
          <span :class="connected ? 'text-green-400 animate-pulse' : 'text-red-400'">●</span>
          <span class="text-sm font-mono text-white">
            {{ namespace }}/{{ podName }}{{ containerName ? `:${containerName}` : '' }}
          </span>
        </div>
        <div v-if="connected" class="text-xs text-gray-400">
          按 Ctrl+C 发送中断信号
        </div>
      </div>
      <div class="flex items-center space-x-2">
        <ElButton v-if="connected" size="small" @click="setFriendlyPrompt">
          设置提示符
        </ElButton>
        <ElButton v-if="connected" size="small" type="danger" @click="handleDisconnect">
          断开连接
        </ElButton>
        <ElButton v-if="!connected" size="small" type="primary" @click="connectTerminal">
          重新连接
        </ElButton>
      </div>
    </div>

    <!-- 终端容器 -->
    <div ref="terminalContainer" class="flex-1 bg-black p-2"></div>
  </div>
</template>

<style scoped>
/* 动画效果 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>
