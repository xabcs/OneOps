<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { fetchGetAuditStats } from '@/service/api';

defineOptions({ name: 'AuditCenter' });

const router = useRouter();

// 审计统计数据
interface AuditStats {
  login: {
    total: number;
    success: number;
    failed: number;
    today: number;
    thisWeek: number;
    thisMonth: number;
  };
  operation: {
    total: number;
    success: number;
    failed: number;
  };
  system: {
    total: number;
    info: number;
    warning: number;
    error: number;
    critical: number;
  };
}

const stats = ref<AuditStats>({
  login: {
    total: 0,
    success: 0,
    failed: 0,
    today: 0,
    thisWeek: 0,
    thisMonth: 0
  },
  operation: {
    total: 0,
    success: 0,
    failed: 0
  },
  system: {
    total: 0,
    info: 0,
    warning: 0,
    error: 0,
    critical: 0
  }
});

const loading = ref(false);

// 获取审计统计数据
async function getAuditStats() {
  loading.value = true;
  try {
    const { data, error } = await fetchGetAuditStats();
    if (!error && data) {
      stats.value = data as AuditStats;
    }
  } catch (err) {
    console.error('获取审计统计失败:', err);
  } finally {
    loading.value = false;
  }
}

// 计算成功率
const loginSuccessRate = computed(() => {
  if (stats.value.login.total === 0) return '0';
  return ((stats.value.login.success / stats.value.login.total) * 100).toFixed(1);
});

const operationSuccessRate = computed(() => {
  if (stats.value.operation.total === 0) return '0';
  return ((stats.value.operation.success / stats.value.operation.total) * 100).toFixed(1);
});

// 导航到详细页面
function navigateToDetail(route: string) {
  router.push(route);
}

// 初始化
onMounted(() => {
  getAuditStats();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <template #header>
        <div class="flex items-center justify-between">
          <p>{{ $t('route.audit') }}</p>
          <ElButton @click="getAuditStats" :loading="loading">
            刷新
          </ElButton>
        </div>
      </template>

      <!-- 统计概览 -->
      <div v-loading="loading" class="grid grid-cols-1 gap-16px sm:grid-cols-2 lg:grid-cols-3">
        <!-- 登录统计 -->
        <div class="stat-card" @click="navigateToDetail('/audit-login-logs')">
          <div class="stat-header">
            <span class="stat-icon">🔐</span>
            <span class="stat-title">登录统计</span>
          </div>
          <div class="stat-content">
            <div class="stat-item">
              <span class="stat-label">总计</span>
              <span class="stat-value">{{ stats.login.total }}</span>
            </div>
            <div class="stat-row">
              <div class="stat-item mini">
                <span class="stat-label">成功</span>
                <span class="stat-value success">{{ stats.login.success }}</span>
              </div>
              <div class="stat-item mini">
                <span class="stat-label">失败</span>
                <span class="stat-value error">{{ stats.login.failed }}</span>
              </div>
            </div>
            <div class="stat-row">
              <div class="stat-item mini">
                <span class="stat-label">今日</span>
                <span class="stat-value">{{ stats.login.today }}</span>
              </div>
              <div class="stat-item mini">
                <span class="stat-label">本周</span>
                <span class="stat-value">{{ stats.login.thisWeek }}</span>
              </div>
              <div class="stat-item mini">
                <span class="stat-label">本月</span>
                <span class="stat-value">{{ stats.login.thisMonth }}</span>
              </div>
            </div>
            <div class="stat-footer">
              <span>成功率: {{ loginSuccessRate }}%</span>
              <ElProgress :percentage="parseFloat(loginSuccessRate)" :color="loginSuccessRate > 80 ? 'success' : loginSuccessRate > 50 ? 'warning' : 'danger'" />
            </div>
            <div class="stat-action">
              <ElButton type="primary" size="small" text>查看详情 →</ElButton>
            </div>
          </div>
        </div>

        <!-- 操作统计 -->
        <div class="stat-card" @click="navigateToDetail('/audit-operation-logs')">
          <div class="stat-header">
            <span class="stat-icon">⚙️</span>
            <span class="stat-title">操作统计</span>
          </div>
          <div class="stat-content">
            <div class="stat-item">
              <span class="stat-label">总计</span>
              <span class="stat-value">{{ stats.operation.total }}</span>
            </div>
            <div class="stat-row">
              <div class="stat-item">
                <span class="stat-label">成功</span>
                <span class="stat-value success">{{ stats.operation.success }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">失败</span>
                <span class="stat-value error">{{ stats.operation.failed }}</span>
              </div>
            </div>
            <div class="stat-footer">
              <span>成功率: {{ operationSuccessRate }}%</span>
              <ElProgress :percentage="parseFloat(operationSuccessRate)" :color="operationSuccessRate > 80 ? 'success' : operationSuccessRate > 50 ? 'warning' : 'danger'" />
            </div>
            <div class="stat-action">
              <ElButton type="primary" size="small" text>查看详情 →</ElButton>
            </div>
          </div>
        </div>

        <!-- 系统事件统计 -->
        <div class="stat-card" @click="navigateToDetail('/audit-system-events')">
          <div class="stat-header">
            <span class="stat-icon">⚠️</span>
            <span class="stat-title">系统事件</span>
          </div>
          <div class="stat-content">
            <div class="stat-item">
              <span class="stat-label">总计</span>
              <span class="stat-value">{{ stats.system.total }}</span>
            </div>
            <div class="stat-row">
              <div class="stat-item mini">
                <span class="stat-label">信息</span>
                <span class="stat-value info">{{ stats.system.info }}</span>
              </div>
              <div class="stat-item mini">
                <span class="stat-label">警告</span>
                <span class="stat-value warning">{{ stats.system.warning }}</span>
              </div>
            </div>
            <div class="stat-row">
              <div class="stat-item mini">
                <span class="stat-label">错误</span>
                <span class="stat-value error">{{ stats.system.error }}</span>
              </div>
              <div class="stat-item mini">
                <span class="stat-label">严重</span>
                <span class="stat-value critical">{{ stats.system.critical }}</span>
              </div>
            </div>
            <div class="stat-action">
              <ElButton type="primary" size="small" text>查看详情 →</ElButton>
            </div>
          </div>
        </div>
      </div>
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
.stat-card {
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  padding: 16px;
  background-color: var(--el-bg-color);
  cursor: pointer;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
    border-color: var(--el-color-primary);
  }

  .stat-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
    font-size: 16px;
    font-weight: 500;
  }

  .stat-icon {
    font-size: 24px;
  }

  .stat-content {
    .stat-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;

      .stat-label {
        color: var(--el-text-color-secondary);
        font-size: 14px;
      }

      .stat-value {
        font-size: 18px;
        font-weight: 600;

        &.success { color: var(--el-color-success); }
        &.error { color: var(--el-color-danger); }
        &.warning { color: var(--el-color-warning); }
        &.info { color: var(--el-color-info); }
        &.critical { color: #f56c6c; }
      }

      &.mini {
        .stat-value { font-size: 14px; }
      }
    }

    .stat-row {
      display: flex;
      gap: 8px;
      margin-bottom: 8px;

      .stat-item {
        flex: 1;
      }
    }
  }

  .stat-footer {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color);

    span {
      display: block;
      margin-bottom: 8px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .stat-action {
    margin-top: 12px;
    text-align: center;
  }
}
</style>