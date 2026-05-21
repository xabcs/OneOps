<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import 'xterm/css/xterm.css';
import { $t } from '@/locales';

interface Props {
  sessionId: number;
  websocketUrl: string;
  serverName: string;
  serverIp: string;
  rows?: number;
  cols?: number;
}

const props = withDefaults(defineProps<Props>(), {
  rows: 40,
  cols: 80
});

const emit = defineEmits<{
  (e: 'connected'): void;
  (e: 'disconnected', reason: string): void;
  (e: 'error', error: string): void;
}>();

const terminalRef = ref<HTMLElement>();
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let ws: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

// 初始化终端
function initTerminal() {
  if (!terminalRef.value) return;

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#ffffff',
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
    rows: props.rows,
    cols: props.cols
  });

  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.loadAddon(new WebLinksAddon());

  terminal.open(terminalRef.value);
  fitAddon.fit();

  // 欢迎信息
  terminal.writeln(`\x1b[1;32m正在连接到 ${props.serverName} (${props.serverIp})...\x1b[0m\r\n`);

  // 监听终端尺寸变化
  terminal.onResize(({ rows, cols }) => {
    resizeTerminal(rows, cols);
  });
}

// 连接 WebSocket
function connect() {
  if (!props.websocketUrl) {
    emit('error', 'WebSocket URL 为空');
    return;
  }

  const wsUrl = props.websocketUrl.startsWith('ws')
    ? props.websocketUrl
    : `ws://${window.location.host}${props.websocketUrl}`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    if (terminal) {
      terminal.writeln(`\x1b[1;32m连接成功！\x1b[0m\r\n`);
      emit('connected');
    }
  };

  ws.onmessage = (event) => {
    if (terminal) {
      terminal.write(event.data);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    if (terminal) {
      terminal.writeln(`\x1b[1;31m连接错误\x1b[0m\r\n`);
    }
    emit('error', 'WebSocket 连接错误');
  };

  ws.onclose = (event) => {
    if (terminal) {
      terminal.writeln(`\r\n\x1b[1;33m连接已断开 (code: ${event.code})\x1b[0m\r\n`);
    }
    emit('disconnected', `连接断开: ${event.code}`);

    // 自动重连（仅非正常断开）
    if (event.code !== 1000 && reconnectTimer === null) {
      terminal?.writeln(`\x1b[1;33m3秒后尝试重连...\x1b[0m\r\n`);
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 3000);
    }
  };
}

// 调整终端大小
function resizeTerminal(rows: number, cols: number) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    // 调用后端接口调整终端大小
    fetch(`/api/cmdb/sessions/${props.sessionId}/resize`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify({ rows, cols })
    }).catch(err => console.error('Failed to resize terminal:', err));
  }
}

// 断开连接
function disconnect() {
  if (ws) {
    ws.close(1000, '用户主动断开');
    ws = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

// 清理
onMounted(() => {
  initTerminal();
  connect();

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  disconnect();
  window.removeEventListener('resize', handleResize);
  if (terminal) {
    terminal.dispose();
    terminal = null;
  }
});

function handleResize() {
  if (fitAddon) {
    fitAddon.fit();
  }
}

// 监听 sessionId 变化
watch(() => props.sessionId, (newId) => {
  if (newId && newId !== 0) {
    disconnect();
    if (terminal) {
      terminal.reset();
      terminal.writeln(`\r\n\x1b[1;32m正在连接到新会话...\x1b[0m\r\n`);
    }
    connect();
  }
});
</script>

<template>
  <div class="ssh-terminal-wrapper">
    <div ref="terminalRef" class="ssh-terminal" />
  </div>
</template>

<style scoped>
.ssh-terminal-wrapper {
  width: 100%;
  height: 100%;
  background: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}

.ssh-terminal {
  width: 100%;
  height: 100%;
  padding: 8px;
}

:deep(.xterm) {
  padding: 0;
}

:deep(.xterm .xterm-viewport) {
  background-color: #1e1e1e;
}
</style>
