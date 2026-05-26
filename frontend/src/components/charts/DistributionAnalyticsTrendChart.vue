<template>
  <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('distribution.analytics.trendTitle') }}</h3>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.trendDescription') }}</p>
      </div>
    </div>

    <div v-if="loading" class="flex h-64 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="chartData" class="h-64">
      <Line :data="chartData" :options="chartOptions" />
    </div>
    <div v-else class="flex h-64 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
      {{ t('distribution.analytics.emptyTrend') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import type { DistributionAnalyticsTrendPoint } from '@/api/distribution'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  trendData: DistributionAnalyticsTrendPoint[]
  loading?: boolean
}>()

const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const palette = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  recharge: '#0f766e',
  consumption: '#2563eb',
  commission: '#ea580c',
}))

const chartData = computed(() => {
  if (!props.trendData?.length) return null
  return {
    labels: props.trendData.map((point) => point.date),
    datasets: [
      {
        label: t('distribution.analytics.metrics.rechargeAmount'),
        data: props.trendData.map((point) => point.recharge_amount),
        borderColor: palette.value.recharge,
        backgroundColor: `${palette.value.recharge}20`,
        fill: true,
        tension: 0.3,
      },
      {
        label: t('distribution.analytics.metrics.consumptionAmount'),
        data: props.trendData.map((point) => point.consumption_amount),
        borderColor: palette.value.consumption,
        backgroundColor: `${palette.value.consumption}20`,
        fill: true,
        tension: 0.3,
      },
      {
        label: t('distribution.analytics.metrics.commissionAmount'),
        data: props.trendData.map((point) => point.commission_amount),
        borderColor: palette.value.commission,
        backgroundColor: `${palette.value.commission}20`,
        fill: true,
        tension: 0.3,
      },
    ],
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: palette.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
      },
    },
  },
  scales: {
    x: {
      grid: { color: palette.value.grid },
      ticks: { color: palette.value.text },
    },
    y: {
      grid: { color: palette.value.grid },
      ticks: {
        color: palette.value.text,
        callback: (value: string | number) => `$${Number(value).toFixed(2)}`,
      },
    },
  },
}))
</script>
