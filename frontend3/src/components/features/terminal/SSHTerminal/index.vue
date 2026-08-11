<script setup lang="ts">
    import { onMounted, onUnmounted, ref, watch } from 'vue';
    import { Terminal } from '@xterm/xterm';
    import { FitAddon } from '@xterm/addon-fit';
    import { WebLinksAddon } from '@xterm/addon-web-links';
    import '@xterm/xterm/css/xterm.css';
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
        cursorStyle: 'block',
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {
          background: '#000000',
          foreground: '#e8e8e8',
          cursor: '#0dbc79',
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
      terminal.onData(data => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(data);
        }
      });

      // 欢迎信息（简化为单行符号）
      terminal.writeln(`\x1B[90m⟳ 连接中...\x1B[0m`);

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

      // 如果已有连接且状态正常，不重新连接
      if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
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
          // 清除"连接中"消息并显示单行成功符号
          terminal.clear();
          terminal.writeln(`\x1B[92m✓ 已连接\x1B[0m\r\n`);
          terminal.focus();
          emit('connected');
        }
      };

      ws.onmessage = async event => {
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
          terminal.clear();
          terminal.writeln(`\x1B[91m✗ 连接失败：检查凭证和服务器可达性\x1B[0m\r\n`);
        }
      };

      ws.onclose = event => {
        if (terminal) {
          terminal.writeln(`\r\n\x1B[93m⚠ 连接断开 (${event.code})\x1B[0m\r\n`);
          if (event.reason) {
            terminal.writeln(`\x1B[91m${event.reason}\x1B[0m\r\n`);
          }
        }
        emit('disconnected', `连接断开: ${event.code}`);

        // 1000=正常关闭，1011=服务端错误（SSH失败等永久性错误）不重连
        // wasConnected=false 说明本次会话从未成功（凭证缺失/SSH失败/会话已关闭），重连无意义
        const isSessionDead = !wasConnected;
        const noRetry = event.code === 1000 || event.code === 1011 || isSessionDead;
        if (!noRetry && reconnectTimer === null) {
          terminal?.writeln(`\x1B[93m⟳ 3秒后重连...\x1B[0m`);
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
            Authorization: `Bearer ${localStg.get('token') || ''}`
          },
          body: JSON.stringify({ rows, cols })
        }).catch(err => console.error('Failed to resize terminal:', err));
      }
    }

    // 断开连接
    function disconnect() {
      if (ws) {
        // 移除所有事件监听器，避免内存泄漏
        ws.onopen = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.onclose = null;

        if (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN) {
          ws.close(1000, '组件卸载或切换会话');
        }
        ws = null;
      }
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      wasConnected = false;
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

    // 监听组件显示状态（通过v-show切换时重新聚焦）
    watch(
      () => [props.sessionId, props.websocketUrl],
      ([sessionId, websocketUrl]) => {
        // 当组件首次获得sessionId和websocketUrl时才连接
        if (sessionId && websocketUrl && !ws && terminal) {
          connect();
        } else if (sessionId && websocketUrl && terminal) {
          // 当切换回已存在的会话时，重新聚焦终端
          terminal.focus();
          if (fitAddon) {
            fitAddon.fit();
          }
        }
      }
    );
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
      background: #000000;
      border-radius: 0;
      overflow: hidden;
    }

    .ssh-terminal {
      width: 100%;
      height: 100%;
      padding: 6px;
    }

    :deep(.xterm) {
      padding: 0;
    }

    :deep(.xterm .xterm-viewport) {
      background-color: #000000;
    }
</style>
