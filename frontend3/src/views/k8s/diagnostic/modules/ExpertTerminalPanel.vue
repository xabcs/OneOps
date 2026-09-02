<script setup lang="ts">
  import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { useAuthStore } from '@/store/modules/auth';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import '@xterm/xterm/css/xterm.css';

  defineOptions({ name: 'DiagnosticExpertTerminalPanel' });

  const props = defineProps<{ autoCommand?: string }>();
  const emit = defineEmits<{ 'update:autoCommand': [value: string] }>();

  const authStore = useAuthStore();
  const diagStore = useDiagnosticStore();

  const terminalContainer = ref<HTMLElement | null>(null);
  const terminal = ref<Terminal | null>(null);
  const fitAddon = ref<FitAddon | null>(null);
  const ws = ref<WebSocket | null>(null);
  const connected = ref(false);
  const connecting = ref(false);
  /** 待连接建立后自动下发的命令（场景跳转） */
  const pendingCommand = ref('');

  let mounted = true;
  let pingTimer: ReturnType<typeof setInterval> | null = null;
  let resizeHandler: (() => void) | null = null;

  // 目标变化时断开当前会话（每 agent 互斥一个会话）
  watch(
    () => diagStore.currentAgent?.agentId,
    (val, old) => {
      if (val !== old && connected.value) {
        disconnect();
        writeSystem(`\r\n\x1B[33m已切换诊断目标，会话已断开\x1B[0m\r\n`);
      }
    }
  );

  // 场景跳转预填命令
  watch(
    () => props.autoCommand,
    cmd => {
      if (!cmd) return;
      if (connected.value) {
        sendCommand(cmd);
        emit('update:autoCommand', '');
      } else {
        pendingCommand.value = cmd;
        emit('update:autoCommand', '');
        connect();
      }
    }
  );

  function writeSystem(text: string) {
    terminal.value?.write(text);
  }

  function sendCommand(command: string) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'input', data: `${command}\n` }));
    }
  }

  function startPing() {
    stopPing();
    pingTimer = setInterval(() => {
      if (ws.value?.readyState === WebSocket.OPEN) {
        ws.value.send(JSON.stringify({ type: 'ping' }));
      }
    }, 60_000);
  }

  function stopPing() {
    if (pingTimer) {
      clearInterval(pingTimer);
      pingTimer = null;
    }
  }

  function initTerminal() {
    if (terminal.value || !terminalContainer.value) return;
    terminal.value = new Terminal({
      fontSize: 13,
      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", Consolas, "source-code-pro", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#e5e5e5',
        cursor: '#ffffff',
        selectionBackground: 'rgba(255, 255, 255, 0.25)'
      },
      cursorBlink: true,
      scrollback: 5000,
      allowProposedApi: true,
      convertEol: false
    });
    fitAddon.value = new FitAddon();
    terminal.value.loadAddon(fitAddon.value);
    terminal.value.open(terminalContainer.value);

    // 输入 → WS
    terminal.value.onData(data => {
      if (ws.value?.readyState === WebSocket.OPEN) {
        ws.value.send(JSON.stringify({ type: 'input', data }));
      }
    });

    resizeHandler = () => {
      fitAddon.value?.fit();
      if (ws.value?.readyState === WebSocket.OPEN && terminal.value) {
        ws.value.send(JSON.stringify({ type: 'resize', cols: terminal.value.cols, rows: terminal.value.rows }));
      }
    };
    window.addEventListener('resize', resizeHandler);
    nextTick(() => fitAddon.value?.fit());
  }

  function connect() {
    const agent = diagStore.currentAgent;
    if (!agent) {
      ElMessage.warning('请先在左侧选择诊断实例');
      return;
    }
    if (!agent.online) {
      ElMessage.error('该实例 Agent 离线，无法建立会话');
      return;
    }
    if (connected.value || connecting.value) return;

    initTerminal();

    const token = authStore.token;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host =
      import.meta.env.VITE_SERVICE_BASE_URL?.replace(/^https?:\/\//, '').replace(/\/api$/, '') || window.location.host;
    const wsUrl = `${protocol}//${host}/api/k8s/diagnostic/session/ws?agentId=${encodeURIComponent(agent.agentId)}&clusterId=${agent.clusterId}&token=${token}`;

    connecting.value = true;
    writeSystem(`\r\n\x1B[36m正在连接 ${agent.podName || agent.agentId} ...\x1B[0m\r\n`);

    const socket = new WebSocket(wsUrl);
    ws.value = socket;

    socket.onopen = () => {
      if (!mounted) {
        socket.close();
        return;
      }
      connected.value = true;
      connecting.value = false;
      startPing();
      writeSystem('\x1B[32m✓ 已建立 Arthas 诊断会话（tunnel 透传，全量命令可用）\x1B[0m\r\n');
      writeSystem('\x1B[90m提示：q 或 Ctrl+C 退出流式命令；stop 结束会话；会话空闲 10 分钟自动断开\x1B[0m\r\n\r\n');
      nextTick(() => {
        fitAddon.value?.fit();
        if (terminal.value) {
          socket.send(JSON.stringify({ type: 'resize', cols: terminal.value.cols, rows: terminal.value.rows }));
        }
        // 场景跳转命令自动下发
        if (pendingCommand.value) {
          sendCommand(pendingCommand.value);
          pendingCommand.value = '';
        }
      });
    };

    socket.onmessage = event => {
      if (!mounted) return;
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output' && typeof msg.data === 'string') {
          terminal.value?.write(msg.data);
        } else if (msg.type === 'closed') {
          writeSystem('\r\n\x1B[33m诊断会话已关闭\x1B[0m\r\n');
        }
      } catch {
        // 兼容非 JSON 帧透传
        terminal.value?.write(event.data);
      }
    };

    socket.onerror = () => {
      if (!mounted) return;
      writeSystem('\r\n\x1B[31m✗ 连接错误\x1B[0m\r\n');
    };

    socket.onclose = ev => {
      if (!mounted) return;
      connected.value = false;
      connecting.value = false;
      stopPing();
      if (ev.wasClean) {
        writeSystem(`\r\n\x1B[33m连接已关闭 (code: ${ev.code})\x1B[0m\r\n`);
      } else {
        writeSystem('\r\n\x1B[31m连接异常关闭（可能被管理员终止或 agent 掉线）\x1B[0m\r\n');
      }
    };
  }

  function disconnect() {
    stopPing();
    ws.value?.close();
    ws.value = null;
    connected.value = false;
    connecting.value = false;
  }

  function sendStop() {
    // 会话收尾：还原字节码增强后断开
    sendCommand('reset');
    setTimeout(() => sendCommand('stop'), 300);
  }

  onMounted(() => {
    mounted = true;
    initTerminal();
  });

  onBeforeUnmount(() => {
    mounted = false;
    disconnect();
    if (resizeHandler) {
      window.removeEventListener('resize', resizeHandler);
      resizeHandler = null;
    }
    terminal.value?.dispose();
    terminal.value = null;
  });
</script>

<template>
  <div class="expert-terminal">
    <!-- 终端工具栏 -->
    <div class="terminal-toolbar">
      <div class="toolbar-left">
        <span :class="['conn-dot', { online: connected }]"></span>
        <span class="target-name">
          {{ diagStore.currentAgent?.podName || diagStore.currentAgent?.agentId || '未选择诊断目标' }}
        </span>
        <ElTag v-if="connected" size="small" type="success">会话中</ElTag>
        <ElTag v-else-if="connecting" size="small" type="warning">连接中</ElTag>
        <span v-else class="toolbar-hint">全量 Arthas 命令（含 ognl / redefine 等高危命令，全程审计录制）</span>
      </div>
      <div class="toolbar-right">
        <PermissionButton
          v-if="!connected"
          code="k8s.diagnostic.execute"
          size="small"
          type="primary"
          :loading="connecting"
          :disabled="!diagStore.currentAgent?.online"
          @click="connect"
        >
          连接
        </PermissionButton>
        <template v-else>
          <ElButton size="small" @click="sendStop">还原增强(stop)</ElButton>
          <ElButton size="small" type="danger" @click="disconnect">断开</ElButton>
        </template>
      </div>
    </div>

    <!-- 终端 -->
    <div v-show="connected || terminal" ref="terminalContainer" class="terminal-body"></div>
    <div v-if="!terminal" class="terminal-placeholder">
      <ElEmpty description="连接后进入 Arthas 专家终端，支持全部诊断命令" :image-size="90" />
    </div>
  </div>
</template>

<style scoped lang="scss">
  .expert-terminal {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background: #1e1e1e;

    .terminal-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: #252526;
      border-bottom: 1px solid #3c3c3c;
      flex-shrink: 0;

      .toolbar-left {
        display: flex;
        align-items: center;
        gap: 8px;
        color: #cccccc;
        font-size: 13px;
        min-width: 0;

        .conn-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: #f14c4c;

          &.online {
            background: #23d18b;
          }
        }

        .target-name {
          font-family: Monaco, Menlo, Consolas, monospace;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .toolbar-hint {
          color: #858585;
          font-size: 12px;
        }
      }
    }

    .terminal-body {
      flex: 1;
      min-height: 0;
      padding: 4px;
    }

    .terminal-placeholder {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--el-bg-color);
    }
  }
</style>
