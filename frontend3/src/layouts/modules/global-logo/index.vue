<script setup lang="ts">
import { computed } from 'vue';
import { useAuthStore } from '@/store/modules/auth';

defineOptions({ name: 'GlobalLogo' });

const authStore = useAuthStore();

/**
 * The home path for the current logged-in user.
 * Falls back to "/" if user is not logged in or homePath is not configured.
 */
const homePath = computed(() => {
  const path = authStore.userInfo?.homePath;
  return path && path.trim() ? path : '/';
});
</script>

<template>
  <RouterLink :to="homePath" class="w-full flex-center nowrap-hidden">
    <SystemLogo class="size-48px" />
  </RouterLink>
</template>

<style scoped></style>
