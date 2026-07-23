<script setup lang="ts">
import { ref } from 'vue';
import Applications from './applications/index.vue';
import RoleBindings from './role-bindings/index.vue';
import UserAuthorization from './user-authorization/index.vue';
import OperationLogs from './operation-logs/index.vue';

defineOptions({ name: 'AuthCenter' });

const activeTab = ref('applications');
</script>

<template>
  <div class="h-full flex flex-col p-4">
    <NCard title="授权中心" :bordered="false" class="flex-1 rounded-16px shadow-sm">
      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="applications" tab="外部应用">
          <Applications />
        </NTabPane>
        <NTabPane name="role-bindings" tab="角色绑定">
          <RoleBindings />
        </NTabPane>
        <NTabPane name="user-authorization" tab="用户授权">
          <UserAuthorization />
        </NTabPane>
        <NTabPane name="operation-logs" tab="操作日志">
          <OperationLogs />
        </NTabPane>
      </NTabs>
    </NCard>
  </div>
</template>
