<script setup lang="ts">
import { ref, watch } from 'vue';
import { Search, Refresh } from '@element-plus/icons-vue';
import { useRouteStore } from '@/store/modules/route';
import { $t } from '@/locales';

defineOptions({
  name: 'PermissionSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const routeStore = useRouteStore();

const searchParams = ref({
  name: '',
  code: '',
  module: '',
  level: undefined as number | undefined
});

// 监听路由查询参数变化
watch(
  () => routeStore.query,
  val => {
    if (val) {
      // 从路由查询参数中解析搜索条件
      Object.keys(val).forEach(key => {
        if (key in searchParams.value) {
          (searchParams.value as any)[key] = val[key];
        }
      });
    }
  },
  { immediate: true }
);

// 重置搜索
function handleReset() {
  searchParams.value = {
    name: '',
    code: '',
    module: '',
    level: undefined
  };
  emit('reset');
}

// 搜索
function handleSearch() {
  emit('search');
}
</script>

<template>
  <div class="permission-search-container">
    <div class="search-form">
      <div class="form-item">
        <label>{{ $t('page.manage.permission.permissionName') }}</label>
        <ElInput
          v-model="searchParams.name"
          :placeholder="$t('page.manage.permission.form.namePlaceholder')"
          clearable
          class="search-input"
        />
      </div>

      <div class="form-item">
        <label>{{ $t('page.manage.permission.permissionCode') }}</label>
        <ElInput
          v-model="searchParams.code"
          :placeholder="$t('page.manage.permission.form.codePlaceholder')"
          clearable
          class="search-input"
        />
      </div>

      <div class="form-item">
        <label>{{ $t('page.manage.permission.module') }}</label>
        <ElSelect
          v-model="searchParams.module"
          :placeholder="$t('page.manage.permission.form.modulePlaceholder')"
          clearable
          class="search-select"
        >
          <ElOption label="系统管理" value="system" />
          <ElOption label="授权中心" value="auth" />
          <ElOption label="资产管理" value="cmdb" />
          <ElOption label="监控中心" value="monitoring" />
          <ElOption label="K8s管理" value="k8s" />
        </ElSelect>
      </div>

      <div class="form-item">
        <label>{{ $t('page.manage.permission.level') }}</label>
        <ElSelect
          v-model="searchParams.level"
          :placeholder="$t('page.manage.permission.form.levelPlaceholder')"
          clearable
          class="search-select"
        >
          <ElOption label="模块级" :value="1" />
          <ElOption label="页面级" :value="2" />
          <ElOption label="按钮级" :value="3" />
          <ElOption label="API级" :value="4" />
        </ElSelect>
      </div>

      <div class="form-actions">
        <ElButton type="primary" @click="handleSearch">
          <ElIcon><Search /></ElIcon>
          {{ $t('common.search') }}
        </ElButton>
        <ElButton @click="handleReset">
          <ElIcon><Refresh /></ElIcon>
          {{ $t('common.reset') }}
        </ElButton>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.permission-search-container {
  padding: 16px;
  background: var(--el-bg-color-page);
}

.search-form {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 200px;
  flex: 1;

  label {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }
}

.search-input,
.search-select {
  width: 100%;
}

.form-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 2px;
}

@media (max-width: 768px) {
  .search-form {
    flex-direction: column;
  }

  .form-item {
    min-width: 100%;
  }

  .form-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
