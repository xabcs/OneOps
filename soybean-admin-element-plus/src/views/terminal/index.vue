<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';

defineOptions({
  name: 'TerminalIndex'
});

const router = useRouter();
const route = useRoute();

onMounted(() => {
  // 重定向到终端工作台，保留所有查询参数
  router.replace({
    path: '/terminal/workbench',
    query: route.query
  });
});
</script>

<template>
  <div class="terminal-redirect">
    <div class="loading-container">
      <div class="loading-spinner"></div>
      <p>正在打开终端工作台...</p>
    </div>
  </div>
</template>

<style scoped>
.terminal-redirect {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #050505;
}

.loading-container {
  text-align: center;
  color: #fff;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 20px;
  border: 3px solid #333;
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
