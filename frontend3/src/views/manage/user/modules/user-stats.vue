<script setup lang="ts">
defineProps<{
  userStats: {
    total: number;
    active: number;
    inactive: number;
  };
}>();
</script>

<template>
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-value">{{ userStats.total }}</div>
      <div class="stat-label">用户总数</div>
      <div class="stat-desc">系统注册用户总量</div>
    </div>
    <div class="stat-card success-card">
      <div class="stat-value">{{ userStats.active }}</div>
      <div class="stat-label">启用用户</div>
      <div class="stat-desc">当前状态为启用</div>
    </div>
    <div class="stat-card warning-card">
      <div class="stat-value">{{ userStats.inactive }}</div>
      <div class="stat-label">禁用用户</div>
      <div class="stat-desc">当前状态为禁用</div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 6px;
}

.stat-card {
  cursor: default;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  text-align: center;
  background: var(--sx-stat-default-bg);
  border: 1px solid var(--sx-stat-default-border);
  border-radius: var(--sx-stat-radius);
  box-shadow: var(--sx-stat-shadow);
  padding: var(--sx-content-card-padding);
  display: flex;
  flex-direction: column;
  gap: 8px;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  }

  &::after {
    content: '';
    position: absolute;
    right: -24px;
    bottom: -30px;
    width: 108px;
    height: 108px;
    border-radius: 50%;
    opacity: 0.16;
  }

  &.success-card {
    background: var(--sx-stat-success-bg);
  }

  &.warning-card {
    background: var(--sx-stat-warning-bg);
  }
}

.stat-value {
  font-size: 28px;
  line-height: 1.05;
  color: var(--sx-text-primary);
  font-weight: 700;
  margin-bottom: 4px;
}

.stat-label {
  color: var(--sx-text-secondary);
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stat-desc {
  color: var(--sx-text-muted);
  font-size: 12px;
  line-height: 1.4;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
