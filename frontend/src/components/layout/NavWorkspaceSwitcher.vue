<template>
  <div
    class="nav-workspace-switcher"
    :class="{ 'nav-workspace-switcher-collapsed': collapsed }"
    role="tablist"
    :aria-label="t('nav.workspaceSwitcherLabel')"
  >
    <button
      type="button"
      class="nav-workspace-switcher-btn"
      :class="{ 'nav-workspace-switcher-btn-active': workspace === 'consumer' }"
      role="tab"
      :aria-selected="workspace === 'consumer'"
      :title="collapsed ? consumerLabel : undefined"
      @click="switchWorkspace('consumer')"
    >
      <component :is="consumerIcon" class="h-4 w-4 flex-shrink-0" />
      <span class="nav-workspace-switcher-label" :aria-hidden="collapsed ? 'true' : 'false'">
        {{ consumerLabel }}
      </span>
    </button>
    <button
      type="button"
      class="nav-workspace-switcher-btn"
      :class="{ 'nav-workspace-switcher-btn-active': workspace === 'distribution' }"
      role="tab"
      :aria-selected="workspace === 'distribution'"
      :title="collapsed ? distributionLabel : undefined"
      @click="switchWorkspace('distribution')"
    >
      <component :is="distributionIcon" class="h-4 w-4 flex-shrink-0" />
      <span class="nav-workspace-switcher-label" :aria-hidden="collapsed ? 'true' : 'false'">
        {{ distributionLabel }}
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useNavWorkspace, type NavWorkspaceScope } from '@/composables/useNavWorkspace'
import type { NavWorkspace } from '@/nav/types'

const props = withDefaults(
  defineProps<{
    collapsed?: boolean
    scope?: NavWorkspaceScope
  }>(),
  {
    scope: 'user',
  },
)

const { t } = useI18n()
const appStore = useAppStore()
const { switchWorkspace: navigateWorkspace } = useNavWorkspace(props.scope)

const workspace = computed(() => (props.scope === 'admin' ? appStore.adminNavWorkspace : appStore.navWorkspace))

const consumerLabel = computed(() =>
  props.scope === 'admin' ? t('nav.workspaceAdminPlatform') : t('nav.workspaceConsumer'),
)
const distributionLabel = computed(() =>
  props.scope === 'admin' ? t('nav.workspaceAdminDistribution') : t('nav.workspaceDistribution'),
)

const ApiWorkspaceIcon = {
  render: () =>
    h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z',
      }),
    ]),
}

const AdminPlatformIcon = {
  render: () =>
    h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z',
      }),
    ]),
}

const DistributionWorkspaceIcon = {
  render: () =>
    h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0l4.179 2.25L12 17.25 2.25 12m15.321-2.25l4.179 2.25L12 17.25l-9.75-5.25',
      }),
    ]),
}

const consumerIcon = computed(() => (props.scope === 'admin' ? AdminPlatformIcon : ApiWorkspaceIcon))
const distributionIcon = DistributionWorkspaceIcon

async function switchWorkspace(next: NavWorkspace): Promise<void> {
  await navigateWorkspace(next)
}
</script>

<style scoped>
.nav-workspace-switcher {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.25rem;
  margin: 0 0.75rem 0.75rem;
  padding: 0.25rem;
  border-radius: 0.875rem;
  background: rgb(243 244 246);
}

.dark .nav-workspace-switcher {
  background: rgb(31 41 55);
}

.nav-workspace-switcher-btn {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  border-radius: 0.625rem;
  padding: 0.5rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1rem;
  color: rgb(75 85 99);
  transition: background-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
}

.dark .nav-workspace-switcher-btn {
  color: rgb(156 163 175);
}

.nav-workspace-switcher-btn-active {
  background: white;
  color: rgb(17 24 39);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.06);
}

.dark .nav-workspace-switcher-btn-active {
  background: rgb(17 24 39);
  color: white;
}

.nav-workspace-switcher-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-workspace-switcher-collapsed {
  grid-template-columns: 1fr;
}

.nav-workspace-switcher-collapsed .nav-workspace-switcher-btn {
  padding: 0.625rem;
}

.nav-workspace-switcher-collapsed .nav-workspace-switcher-label {
  display: none;
}
</style>
