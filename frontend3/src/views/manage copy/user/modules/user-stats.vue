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
  <ElRow :gutter="16">
    <ElCol :span="8">
      <ElCard shadow="hover">
        <ElStatistic :value="userStats.total" title="用户总数" />
      </ElCard>
    </ElCol>
    <ElCol :span="8">
      <ElCard shadow="hover">
        <ElStatistic :value="userStats.active" title="启用用户">
          <template #suffix>
            <ElTag type="success" size="small">活跃</ElTag>
          </template>
        </ElStatistic>
      </ElCard>
    </ElCol>
    <ElCol :span="8">
      <ElCard shadow="hover">
        <ElStatistic :value="userStats.inactive" title="禁用用户">
          <template #suffix>
            <ElTag type="warning" size="small">停用</ElTag>
          </template>
        </ElStatistic>
      </ElCard>
    </ElCol>
  </ElRow>
</template>
