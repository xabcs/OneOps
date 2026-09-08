<script setup lang="ts">
  import { computed } from 'vue';
  import { CircleCheck, CircleClose, User } from '@element-plus/icons-vue';

  interface UserStatsData {
    total: number;
    active: number;
    inactive: number;
  }

  const props = defineProps<{
    userStats: UserStatsData;
  }>();

  const statsItems = computed(() => [
    {
      key: 'total',
      value: props.userStats.total,
      label: '用户总数',
      icon: User,
      color: 'var(--el-color-primary)'
    },
    {
      key: 'active',
      value: props.userStats.active,
      label: '活跃用户',
      icon: CircleCheck,
      color: 'var(--el-color-success)'
    },
    {
      key: 'inactive',
      value: props.userStats.inactive,
      label: '非活跃用户',
      icon: CircleClose,
      color: 'var(--el-color-info)'
    }
  ]);
</script>

<template>
  <ElRow :gutter="16">
    <ElCol v-for="item in statsItems" :key="item.key" :span="8">
      <ElCard shadow="hover" class="stat-card">
        <div class="flex items-center gap-12px">
          <ElIcon :size="24" :color="item.color">
            <component :is="item.icon" />
          </ElIcon>
          <ElStatistic :value="item.value" :title="item.label" />
        </div>
      </ElCard>
    </ElCol>
  </ElRow>
</template>
