<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue';
import { ElButton, ElMessage } from 'element-plus';
import { useAuthStore } from '@/store/modules/auth';

interface Props {
  clusterId: number;
  namespace: string;
  podName: string;
  containerName?: string;
}

const props = defineProps<Props>();

const message = ElMessage;
const authStore = useAuthStore();

const terminalRef = ref<HTMLElement | null>(null);
const connected = ref(false);
const ws = ref<WebSocket | null>(null);
const outputBuffer = ref<string[]>([]);

const connectTerminal = () => {
  const token = authStore.token;
  const wsUrl = `${import.meta.env.VITE_SERVICE_BASE_URL}/k8s/terminal/ws?clusterId=${props.clusterId}&namespace=${props.namespace}&podName=${props.podName}&containerName=${props.containerName || ''}&token=${token}`;

  try {
    ws.value = new WebSocket(wsUrl);

    ws.value.onopen = () => {
      connected.value = true;
      outputBuffer.value = ['已连接到 Pod 终端'];
      renderOutput();
    };

    ws.value.onmessage = event => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output') {
          outputBuffer.value.push(msg.data);
          renderOutput();
          scrollToBottom();
        }
      } catch (error) {
        // 处理纯文本消息
        outputBuffer.value.push(event.data);
        renderOutput();
        scrollToBottom();
      }
    };

    ws.value.onerror = error => {
      console.error('WebSocket error:', error);
      message.error('终端连接错误');
    };

    ws.value.onclose = () => {
      connected.value = false;
      outputBuffer.value.push('连接已关闭');
      renderOutput();
    };
  } catch (error: any) {
    message.error(error.message || '连接失败');
  }
};

const renderOutput = () => {
  if (!terminalRef.value) return;
  terminalRef.value.innerHTML = outputBuffer.value
    .map(line => `<div class="whitespace-pre-wrap text-gray-300">${escapeHtml(line)}</div>`)
    .join('');
};

const scrollToBottom = () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight;
    }
  });
};

const escapeHtml = (text: string): string => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

const handleDisconnect = () => {
  if (ws.value) {
    ws.value.close();
  }
};

onMounted(() => {
  connectTerminal();

  // 添加键盘监听
  window.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
  if (ws.value) {
    ws.value.close();
  }
  window.removeEventListener('keydown', handleKeyDown);
});

const handleKeyDown = (event: KeyboardEvent) => {
  if (!connected.value || !ws.value) return;

  // 发送键盘输入到终端
  const data = JSON.stringify({
    type: 'stdin',
    data: event.key
  });

  ws.value.send(data);
};
</script>

<template>
  <div class="h-full flex flex-col bg-gray-900">
    <div class="flex items-center justify-between border-b border-gray-700 bg-gray-800 px-4 py-2">
      <div class="flex items-center space-x-2">
        <span class="text-green-400">●</span>
        <span class="text-sm text-white">
          {{ namespace }}/{{ podName }}{{ containerName ? `:${containerName}` : '' }}
        </span>
      </div>
      <div class="flex items-center space-x-2">
        <ElButton size="small" @click="handleDisconnect">断开连接</ElButton>
      </div>
    </div>
    <div ref="terminalRef" class="flex-1 overflow-auto bg-black p-2 text-sm font-mono">
      <div v-if="!connected" class="text-gray-500">正在连接终端...</div>
      <div v-else class="text-green-400">已连接到 Pod 终端</div>
    </div>
  </div>
</template>
