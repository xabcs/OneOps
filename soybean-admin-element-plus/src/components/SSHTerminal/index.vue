<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import 'xterm/css/xterm.css';
import { localStg } from '@/utils/storage';

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
let wasConnected = false; // 本次连接是否成功过 onopen

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
  terminal.focus();

  // 将用户键盘输入通过 WebSocket 发送到后端 SSH 代理
  terminal.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data);
    }
  });

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

  // 获取 token（SoybeanAdmin 使用带前缀的 storage key）
  const token = localStg.get('token');

  // 构建 WebSocket URL
  // 开发环境：VITE_SERVICE_BASE_URL 是绝对地址（如 http://localhost:8082/api），直连后端
  //           避免通过 Vite 代理，因为 Vite HMR 也使用 WebSocket，两者会冲突
  // 生产环境：VITE_SERVICE_BASE_URL 通常为相对路径，使用当前域名（前后端同源）
  let wsUrl: string;
  const serviceUrl = import.meta.env.VITE_SERVICE_BASE_URL as string;

  if (serviceUrl && /^https?:\/\//.test(serviceUrl)) {
    // 绝对 URL（开发环境）：直接连接后端，去掉末尾 /api
    const wsBase = serviceUrl
      .replace(/^http:\/\//, 'ws://')
      .replace(/^https:\/\//, 'wss://')
      .replace(/\/api\/?$/, '');
    wsUrl = `${wsBase}${props.websocketUrl}`;
  } else {
    // 相对 URL（生产环境）：同源直连
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    wsUrl = `${protocol}://${window.location.host}${props.websocketUrl}`;
  }

  if (token) {
    wsUrl += `?token=${encodeURIComponent(token)}`;
  }

  wasConnected = false;
  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    wasConnected = true;
    if (terminal) {
      terminal.writeln(`\x1b[1;32m连接成功！\x1b[0m\r\n`);
      terminal.focus();
      emit('connected');
    }
  };

  ws.onmessage = async (event) => {
    if (!terminal) return;
    if (event.data instanceof Blob) {
      const text = await event.data.text();
      terminal.write(text);
    } else {
      terminal.write(event.data);
    }
  };

  ws.onerror = () => {
    if (terminal && !wasConnected) {
      terminal.writeln(`\x1b[1;31m无法建立连接，请检查：SSH 凭证是否已绑定、服务器是否可达\x1b[0m\r\n`);
    }
  };

  ws.onclose = (event) => {
    if (terminal) {
      terminal.writeln(`\r\n\x1b[1;33m连接已断开 (code: ${event.code})\x1b[0m\r\n`);
      if (event.reason) {
        terminal.writeln(`\x1b[1;31m原因: ${event.reason}\x1b[0m\r\n`);
      }
    }
    emit('disconnected', `连接断开: ${event.code}`);

    // 1000=正常关闭，1011=服务端错误（SSH失败等永久性错误）不重连
    // wasConnected=false 说明本次会话从未成功（凭证缺失/SSH失败/会话已关闭），重连无意义
    const isSessionDead = !wasConnected;
    const noRetry = event.code === 1000 || event.code === 1011 || isSessionDead;
    if (!noRetry && reconnectTimer === null) {
      terminal?.writeln(`\x1b[1;33m3秒后尝试重连...\x1b[0m\r\n`);
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 3000);
    } else if (isSessionDead && event.code !== 1000) {
      // 会话彻底失败，通知父组件
      emit('error', '连接失败，请检查 SSH 凭证配置或服务器可达性');
    }
  };
}

// 调整终端大小
function resizeTerminal(rows: number, cols: number) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    fetch(`/proxy-default/cmdb/sessions/${props.sessionId}/resize`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStg.get('token') || ''}`
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
