<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { localStg } from '@/utils/storage';
import 'xterm/css/xterm.css';

interface Props {
  sessionId: number;
  serverId: number;
  serverName: string;
  serverIp: string;
  loginAccount: string;
}

const props = defineProps<Props>();

let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let ws: WebSocket | null = null;
const terminalRef = ref<HTMLDivElement>();

// 防止重复初始化
let isInitializing = false;
let currentSessionId: number | null = null;

// 初始化终端
function initTerminal() {
  console.log('=== initTerminal 开始 ===');
  console.log('isInitializing:', isInitializing);
  console.log('currentSessionId:', currentSessionId);
  console.log('props.sessionId:', props.sessionId);
  console.log('serverId:', props.serverId);
  console.log('serverName:', props.serverName);
  console.log('terminal exists:', !!terminal);
  console.log('ws exists:', !!ws);

  // 防止重复初始化同一个会话
  if (isInitializing && currentSessionId === props.sessionId) {
    console.warn('=== 防止重复初始化，跳过 ===');
    return;
  }

  // 如果终端实例已存在，先清理
  if (terminal) {
    console.log('=== 清理已存在的终端实例 ===');
    if (ws) {
      ws.close();
      ws = null;
    }
    terminal.dispose();
    terminal = null;
    fitAddon = null;
  }

  isInitializing = true;
  currentSessionId = props.sessionId;

  console.log('terminalRef.value:', terminalRef.value);

  if (!terminalRef.value) {
    console.error('terminalRef.value 不存在！');
    return;
  }

  console.log('开始创建 Terminal 实例...');

  // 创建终端实例
  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#cccccc',
      cursor: '#cccccc',
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
    }
  });

  console.log('Terminal 实例创建完成:', terminal);

  // 添加插件
  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.loadAddon(new WebLinksAddon());

  console.log('插件加载完成');

  // 挂载终端
  terminal.open(terminalRef.value);
  console.log('终端已挂载到 DOM');

  fitAddon.fit();
  console.log('终端已调整大小');

  // 欢迎信息
  terminal.writeln(`\x1b[1;32m欢迎使用 OneOps 终端\x1b[0m`);
  terminal.writeln(`\x1b[1;36m连接到: ${props.serverName} (${props.serverIp})\x1b[0m`);
  terminal.writeln(`\x1b[1;33m登录用户: ${props.loginAccount}\x1b[0m`);
  terminal.writeln(``);
  terminal.writeln(`正在连接到服务器...`);
  terminal.writeln(``);

  console.log('欢迎信息已输出');

  // 监听用户输入
  terminal.onData(data => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'input',
        data: data
      }));
    }
  });

  // 建立WebSocket连接
  connectWebSocket();

  // 标记初始化完成
  setTimeout(() => {
    isInitializing = false;
    console.log('=== initTerminal 完成 ===');
  }, 100);
}

// 连接WebSocket
function connectWebSocket() {
  // 获取 token - 添加详细调试
  console.log('=== 开始获取 token ===');
  console.log('localStg:', localStg);
  console.log('localStg.get:', typeof localStg.get);

  // 尝试多种方式获取 token
  const token1 = localStorage.getItem('SOY_token');
  const token2 = localStg.get('token');
  const token3 = localStorage.getItem('token');

  console.log('localStorage.getItem("SOY_token"):', token1 ? `${token1.substring(0, 20)}...` : 'null');
  console.log('localStg.get("token"):', token2 ? `${token2.substring(0, 20)}...` : 'null/empty');
  console.log('localStorage.getItem("token"):', token3 ? `${token3.substring(0, 20)}...` : 'null');

  const token = token2 || token1 || '';
  console.log('最终使用的 token:', token ? `${token.substring(0, 20)}...` : 'empty');

  // 构建 WebSocket URL - 使用后端服务器端口 8082
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  // 使用 Vite 代理配置中的后端地址，或者直接使用 localhost:8082
  const wsUrl = `ws://localhost:8082/api/cmdb/sessions/${props.sessionId}/ws?token=${token}`;

  console.log('连接 WebSocket:', wsUrl);

  try {
    console.log('=== 开始创建 WebSocket 连接 ===');
    console.log('当前时间:', new Date().toISOString());
    console.log('wsUrl:', wsUrl);

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('=== WebSocket onopen 触发 ===');
      console.log('当前时间:', new Date().toISOString());
      console.log('readyState:', ws?.readyState);
      if (terminal) {
        terminal.writeln(`\x1b[1;32m✓ 连接成功！\x1b[0m`);
        terminal.writeln(``);
      }
    };

    ws.onmessage = (event) => {
      console.log('=== WebSocket onmessage ===');
      console.log('数据长度:', event.data.length);
      console.log('数据内容:', event.data.substring(0, 100));
      if (terminal) {
        // 后端直接发送原始数据，不需要 JSON 解析
        terminal.write(event.data);
      }
    };

    ws.onerror = (error) => {
      console.error('=== WebSocket onerror ===');
      console.error('错误:', error);
      console.error('readyState:', ws?.readyState);
      if (terminal) {
        terminal.writeln(`\x1b[1;31m✗ 连接错误\x1b[0m`);
      }
    };

    ws.onclose = (event) => {
      console.error('=== WebSocket onclose ===');
      console.error('关闭码:', event.code);
      console.error('原因:', event.reason);
      console.error('wasClean:', event.wasClean);
      console.error('当前时间:', new Date().toISOString());
      console.error('readyState:', ws?.readyState);
      if (terminal) {
        terminal.writeln(`\r\n\x1b[1;31m✗ 连接已关闭 (code: ${event.code})\x1b[0m`);
      }
    };
  } catch (error) {
    console.error('=== WebSocket 创建失败 ===');
    console.error('错误:', error);
    if (terminal) {
      terminal.writeln(`\x1b[1;31m✗ 无法建立连接\x1b[0m`);
    }
  }
}

// 调整终端大小
function fitTerminal() {
  if (fitAddon) {
    fitAddon.fit();
  }
}

// 监听窗口大小变化
let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
  console.log('=== XTermTerminal 挂载 ===');
  console.log('terminalRef:', terminalRef.value);
  console.log('props:', props);

  initTerminal();

  // 监听容器大小变化
  if (terminalRef.value) {
    resizeObserver = new ResizeObserver(() => {
      fitTerminal();
    });
    resizeObserver.observe(terminalRef.value);
  }
});

onUnmounted(() => {
  // 关闭WebSocket
  if (ws) {
    ws.close();
  }

  // 销毁终端
  if (terminal) {
    terminal.dispose();
  }

  // 停止观察
  if (resizeObserver && terminalRef.value) {
    resizeObserver.unobserve(terminalRef.value);
    resizeObserver.disconnect();
  }
});

// 监听会话变化
watch(() => props.sessionId, (newSessionId, oldSessionId) => {
  console.log('=== XTermTerminal 会话变化 watch ===');
  console.log('oldSessionId:', oldSessionId);
  console.log('newSessionId:', newSessionId);
  console.log('currentSessionId:', currentSessionId);
  console.log('isInitializing:', isInitializing);
  console.log('ws exists:', !!ws);
  console.log('ws readyState:', ws?.readyState);

  // 如果新旧会话ID相同，跳过
  if (newSessionId === oldSessionId) {
    console.warn('=== sessionId 未变化，跳过 ===');
    return;
  }

  // 完全清理旧连接和终端实例
  console.log('=== 开始完全清理旧连接 ===');
  if (ws) {
    console.log('关闭旧 WebSocket');
    ws.close();
    ws = null;
  }
  if (terminal) {
    console.log('销毁旧终端实例');
    terminal.dispose();
    terminal = null;
  }
  if (fitAddon) {
    fitAddon = null;
  }

  // 重置状态
  isInitializing = false;
  currentSessionId = null;

  // 重新初始化
  console.log('=== 开始重新初始化 ===');
  initTerminal();
});
</script>

<template>
  <div ref="terminalRef" class="xterm-terminal"></div>
</template>

<style lang="scss">
.xterm-terminal {
  width: 100%;
  height: 100%;
  padding: 8px;
}

.xterm-terminal :deep(.xterm) {
  padding: 0;
}

.xterm-terminal :deep(.xterm .xterm-viewport) {
  background-color: #1e1e1e;
}
</style>
