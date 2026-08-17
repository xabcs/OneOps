<script setup lang="ts">
  import { nextTick, onMounted, onUnmounted, ref } from 'vue';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import '@xterm/xterm/css/xterm.css';
  import { fetchResizeTerminal } from '@/service/api/cmdb';

  interface Props {
    sessionId: number | string;
    serverId: number;
    serverName: string;
    serverIp: string;
    loginAccount: string;
    websocketUrl?: string;
  }

  const props = defineProps<Props>();

  let terminal: Terminal | null = null;
  let fitAddon: FitAddon | null = null;
  let ws: WebSocket | null = null;
  const terminalRef = ref<HTMLDivElement>();
  const isReady = ref(false);

  // 防抖定时器
  let resizeTimeout: ReturnType<typeof setTimeout> | null = null;

  // 记录上次发送的尺寸，避免重复请求
  let lastSentCols = 0;
  let lastSentRows = 0;

  // 获取 token
  function getToken(): string {
    let token = localStorage.getItem('SOY_token') || localStorage.getItem('token') || '';
    token = token.replace(/^["']|["']$/g, '');
    return token;
  }

  // 初始化终端
  async function initTerminal() {
    if (!terminalRef.value) {
      console.error('终端容器不存在');
      return;
    }

    // 创建终端实例
    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0d0d0d',
        foreground: '#cccccc',
        cursor: '#2472c8',
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
      scrollback: 10000,
      tabStopWidth: 8
    });

    // 添加插件
    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());

    // 挂载终端
    terminal.open(terminalRef.value);

    // 等待 DOM 渲染
    await nextTick();

    // 显示欢迎信息
    terminal.writeln(`\x1B[1;36m${props.serverName}\x1B[0m (${props.serverIp})`);
    terminal.writeln(`\x1B[1;34m登录用户: ${props.loginAccount}\x1B[0m`);
    terminal.writeln('');

    // 标记准备就绪
    isReady.value = true;

    // 首次调整尺寸
    scheduleFit();

    // 监听用户输入
    terminal.onData(data => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });

    // 建立 WebSocket 连接
    connectWebSocket();
  }

  // 获取后端服务地址
  function getBackendHost(): string {
    const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL || 'http://localhost:8082/api';
    const url = new URL(baseUrl);
    return url.host;
  }

  // 连接 WebSocket
  function connectWebSocket() {
    const token = getToken();
    const backendHost = getBackendHost();

    let fullWsUrl: string;

    if (props.websocketUrl) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      fullWsUrl = `${protocol}//${backendHost}${props.websocketUrl}?token=${token}`;
    } else {
      const wsUrl = `/api/cmdb/sessions/${props.sessionId}/ws?token=${token}`;
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      fullWsUrl = `${protocol}//${backendHost}${wsUrl}`;
    }

    ws = new WebSocket(fullWsUrl);

    ws.onopen = () => {};

    ws.onmessage = event => {
      if (terminal) {
        terminal.write(event.data);
      }
    };

    ws.onerror = error => {
      console.error('WebSocket 错误:', error);
      if (terminal) {
        terminal.writeln(`\r\n\x1B[1;31m✗ 连接错误，会话可能已关闭\x1B[0m`);
        terminal.writeln(`\x1B[1;33m请重新连接主机\x1B[0m`);
      }
    };

    ws.onclose = event => {
      if (terminal) {
        terminal.writeln(`\r\n\x1B[1;31m✗ 连接已关闭 (code: ${event.code})\x1B[0m`);
      }
    };
  }

  // 调整终端尺寸
  function fitTerminal() {
    if (!fitAddon || !terminal || !terminalRef.value) return;

    try {
      const rect = terminalRef.value.getBoundingClientRect();

      if (rect.width > 0 && rect.height > 0) {
        fitAddon.fit();

        // 通知后端 PTY 调整窗口大小
        const cols = terminal.cols;
        const rows = terminal.rows;
        if (cols > 0 && rows > 0 && (cols !== lastSentCols || rows !== lastSentRows)) {
          lastSentCols = cols;
          lastSentRows = rows;
          fetchResizeTerminal(Number(props.sessionId), { cols, rows }).catch(() => {});
        }
      }
    } catch (error) {
      console.warn('调整终端尺寸失败:', error);
    }
  }

  // 延迟执行 fit
  function scheduleFit() {
    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }

    resizeTimeout = setTimeout(() => {
      fitTerminal();
    }, 100);
  }

  // 窗口 resize 处理
  function handleResize() {
    scheduleFit();
  }

  // 处理可见性变化
  function handleVisibilityChange() {
    if (isReady.value && document.visibilityState === 'visible') {
      scheduleFit();
    }
  }

  onMounted(() => {
    initTerminal();
    window.addEventListener('resize', handleResize);
    document.addEventListener('visibilitychange', handleVisibilityChange);

    // 初始调整
    setTimeout(() => {
      fitTerminal();
    }, 200);
  });

  onUnmounted(() => {
    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }

    window.removeEventListener('resize', handleResize);
    document.removeEventListener('visibilitychange', handleVisibilityChange);

    if (ws) {
      ws.close();
      ws = null;
    }

    if (terminal) {
      terminal.dispose();
      terminal = null;
    }

    if (fitAddon) {
      fitAddon = null;
    }
  });
</script>

<template>
  <div ref="terminalRef" class="xterm-terminal-container"></div>
</template>

<style lang="scss" scoped>
  .xterm-terminal-container {
    width: 100%;
    height: 100%;
    background-color: #0d0d0d;
    overflow: hidden;
    border: none;
  }

  .xterm-terminal-container :deep(.xterm) {
    width: 100%;
    height: 100%;
    background-color: #171717;
    padding: 16px;
    border: none !important;
  }

  .xterm-terminal-container :deep(.xterm-viewport) {
    background-color: #171717 !important;
    border: none !important;
    scrollbar-width: thin;
    scrollbar-color: #444 #171717;
  }

  .xterm-terminal-container :deep(.xterm-viewport::-webkit-scrollbar) {
    width: 8px;
    height: 8px;
  }

  .xterm-terminal-container :deep(.xterm-viewport::-webkit-scrollbar-track) {
    background: #171717;
  }

  .xterm-terminal-container :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
    background: #444;
    border-radius: 4px;
  }

  .xterm-terminal-container :deep(.xterm-viewport::-webkit-scrollbar-thumb:hover) {
    background: #555;
  }

  .xterm-terminal-container :deep(.xterm-screen) {
    background-color: #171717 !important;
    border: none !important;
  }
</style>
