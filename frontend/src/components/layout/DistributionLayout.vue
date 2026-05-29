<template>
  <AppLayout>
    <slot />
  </AppLayout>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from './AppLayout.vue'
import { useNavWorkspace } from '@/composables/useNavWorkspace'

const route = useRoute()
const { syncAllWorkspacesFromRoute } = useNavWorkspace()

watch(
  () => route.path,
  (path) => {
    syncAllWorkspacesFromRoute(path)
  },
  { immediate: true },
)
</script>
