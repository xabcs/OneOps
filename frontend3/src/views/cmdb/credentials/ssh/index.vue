<script setup lang="ts">
  import CredentialManager from '../modules/credential-manager.vue';

  defineOptions({ name: 'CmdbSshCredentials' });
</script>

<template>
  <!-- 复用凭证管理共享面板：原实现与 Access 页逐字重复，改一处生效两处 -->
  <CredentialManager default-tab="all" />
</template>
