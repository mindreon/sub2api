<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('videoGeneration.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('videoGeneration.description') }}</p>
      </div>

      <div class="card p-6">
        <form class="space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="input-label">{{ t('videoGeneration.apiKey') }}</label>
            <input
              v-model="apiKey"
              type="password"
              class="input mt-1 font-mono"
              :placeholder="t('videoGeneration.apiKeyPlaceholder')"
              required
            />
            <p class="input-hint">{{ t('videoGeneration.apiKeyHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('videoGeneration.model') }}</label>
            <Select
              v-model="model"
              :options="modelOptions"
              class="mt-1 w-full"
              searchable
              creatable
              :creatable-prefix="t('videoGeneration.useModel')"
            />
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="input-label">{{ t('videoGeneration.resolution') }}</label>
              <Select v-model="resolution" :options="resolutionOptions" class="mt-1 w-full" />
            </div>
            <div>
              <label class="input-label">{{ t('videoGeneration.ratio') }}</label>
              <Select v-model="ratio" :options="ratioOptions" class="mt-1 w-full" />
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('videoGeneration.duration') }}</label>
            <input v-model.number="duration" type="number" min="1" max="15" step="1" class="input mt-1" />
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <label class="input-label">{{ t('videoGeneration.referenceImage') }}</label>
              <input v-model="referenceImageUrl" type="url" class="input mt-1" placeholder="https://example.com/frame.jpg" />
            </div>
            <div>
              <label class="input-label">{{ t('videoGeneration.referenceVideo') }}</label>
              <input v-model="referenceVideoUrl" type="url" class="input mt-1" placeholder="https://example.com/reference.mp4" />
            </div>
            <div>
              <label class="input-label">{{ t('videoGeneration.referenceAudio') }}</label>
              <input v-model="referenceAudioUrl" type="url" class="input mt-1" placeholder="https://example.com/audio.mp3" />
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('videoGeneration.prompt') }}</label>
            <textarea
              v-model="prompt"
              rows="4"
              class="input mt-1"
              :placeholder="t('videoGeneration.promptPlaceholder')"
              required
            />
          </div>

          <button type="submit" class="btn btn-primary w-full" :disabled="submitting">
            {{ submitting ? t('videoGeneration.submitting') : t('videoGeneration.submit') }}
          </button>
        </form>
      </div>

      <div v-if="currentTask" class="card p-6">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('videoGeneration.taskStatus') }}</h2>
        <dl class="mt-4 space-y-2 text-sm">
          <div class="flex justify-between gap-4">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('videoGeneration.taskId') }}</dt>
            <dd class="font-mono text-gray-900 dark:text-white">{{ currentTask.task_id }}</dd>
          </div>
          <div class="flex justify-between gap-4">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('videoGeneration.status') }}</dt>
            <dd>
              <span class="badge" :class="statusBadgeClass(currentTask.status)">{{ displayStatus(currentTask.status) }}</span>
            </dd>
          </div>
          <div v-if="currentTask.quota != null" class="flex justify-between gap-4">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('videoGeneration.reservedCost') }}</dt>
            <dd>${{ Number(currentTask.quota).toFixed(4) }}</dd>
          </div>
          <div v-if="currentTask.actual_cost != null" class="flex justify-between gap-4">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('videoGeneration.actualCost') }}</dt>
            <dd>${{ Number(currentTask.actual_cost).toFixed(4) }}</dd>
          </div>
          <div v-if="currentTask.model" class="flex justify-between gap-4">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('videoGeneration.model') }}</dt>
            <dd class="font-mono text-gray-900 dark:text-white">{{ currentTask.model }}</dd>
          </div>
          <div v-if="currentTask.fail_reason || currentTask.error_message" class="text-red-600 dark:text-red-400">
            {{ currentTask.fail_reason || currentTask.error_message }}
          </div>
        </dl>
        <div v-if="currentTask.result_url" class="mt-5 space-y-3">
          <video
            class="aspect-video w-full rounded-lg bg-black"
            controls
            playsinline
            :src="currentTask.result_url"
          />
          <a
            :href="currentTask.result_url"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm inline-flex items-center gap-2"
          >
            <Icon name="externalLink" size="sm" />
            {{ t('videoGeneration.openResult') }}
          </a>
        </div>
        <p v-if="polling" class="mt-4 text-sm text-gray-500 dark:text-gray-400">{{ t('videoGeneration.polling') }}</p>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { mediaAPI, type CommonVideoGenerationTask, type CommonVideoGenerationMetadata } from '@/api/media'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const apiKey = ref('')
const model = ref('dreamina-seedance-2-0-260128')
const resolution = ref('720p')
const ratio = ref('16:9')
const duration = ref(5)
const prompt = ref('')
const referenceImageUrl = ref('')
const referenceVideoUrl = ref('')
const referenceAudioUrl = ref('')
const submitting = ref(false)
const polling = ref(false)
const currentTask = ref<CommonVideoGenerationTask | null>(null)

let pollTimer: ReturnType<typeof setTimeout> | null = null

const resolutionOptions = [
  { value: '480p', label: '480p' },
  { value: '720p', label: '720p' },
  { value: '1080p', label: '1080p' },
  { value: '4k', label: '4k' }
]

const ratioOptions = [
  { value: '16:9', label: '16:9' },
  { value: '9:16', label: '9:16' },
  { value: '1:1', label: '1:1' },
  { value: '4:3', label: '4:3' },
  { value: '3:4', label: '3:4' }
]

const modelOptions = [
  { value: 'dreamina-seedance-2-0-260128', label: 'dreamina-seedance-2-0-260128' },
  { value: 'dreamina-seedance-2-0-fast-260128', label: 'dreamina-seedance-2-0-fast-260128' }
]

function statusBadgeClass(status: string) {
  const normalized = normalizeStatus(status)
  if (normalized === 'success') return 'badge-success'
  if (normalized === 'failed') return 'badge-danger'
  if (normalized === 'in_progress') return 'badge-warning'
  return 'badge-secondary'
}

function isTerminal(status: string) {
  return ['success', 'failed'].includes(normalizeStatus(status))
}

function normalizeStatus(status: string) {
  const value = status.toLowerCase()
  if (value === 'success' || value === 'succeeded' || value === 'completed') return 'success'
  if (value === 'failed' || value === 'expired' || value === 'cancelled') return 'failed'
  if (value === 'in_progress' || value === 'running' || value === 'pending' || value === 'queued') return 'in_progress'
  return value
}

function displayStatus(status: string) {
  const normalized = normalizeStatus(status)
  if (normalized === 'success') return 'SUCCESS'
  if (normalized === 'failed') return 'FAILED'
  if (normalized === 'in_progress') return 'IN_PROGRESS'
  return status
}

function buildMetadata(): CommonVideoGenerationMetadata {
  const content: unknown[] = []
  const imageURL = referenceImageUrl.value.trim()
  const videoURL = referenceVideoUrl.value.trim()
  const audioURL = referenceAudioUrl.value.trim()

  if (videoURL) {
    content.push({ type: 'video_url', role: 'reference_video', video_url: { url: videoURL } })
  }
  if (imageURL) {
    content.push({ type: 'image_url', role: 'reference_image', image_url: { url: imageURL } })
  }
  if (audioURL) {
    content.push({ type: 'audio_url', role: 'reference_audio', audio_url: { url: audioURL } })
  }

  const metadata: CommonVideoGenerationMetadata = {
    resolution: resolution.value,
    ratio: ratio.value,
    duration: duration.value
  }
  if (content.length > 0) {
    metadata.content = content
  }
  return metadata
}

async function fetchTask(taskId: string): Promise<CommonVideoGenerationTask> {
  return mediaAPI.getCommonVideoGeneration(apiKey.value, taskId)
}

function stopPolling() {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

async function pollTask(taskId: string) {
  polling.value = true
  try {
    const task = await fetchTask(taskId)
    currentTask.value = task
    if (!isTerminal(task.status)) {
      pollTimer = setTimeout(() => pollTask(taskId), 3000)
      return
    }
    stopPolling()
  } catch (err) {
    stopPolling()
    appStore.showError(err instanceof Error ? err.message : t('videoGeneration.pollFailed'))
  }
}

async function handleSubmit() {
  submitting.value = true
  stopPolling()
  currentTask.value = null
  try {
    const task = await mediaAPI.submitCommonVideoGeneration(apiKey.value, {
      model: model.value.trim(),
      prompt: prompt.value.trim(),
      metadata: buildMetadata()
    })
    currentTask.value = task
    appStore.showSuccess(t('videoGeneration.submitted'))
    if (!isTerminal(task.status)) {
      pollTask(task.task_id)
    }
  } catch (err) {
    appStore.showError(err instanceof Error ? err.message : t('videoGeneration.submitFailed'))
  } finally {
    submitting.value = false
  }
}

onUnmounted(stopPolling)
</script>
