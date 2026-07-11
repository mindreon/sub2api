/**
 * 多模态异步任务 API（JWT 用户控制台）
 */

import { apiClient } from './client'
import { buildGatewayUrl } from './url'
import type { PaginatedResponse } from '@/types'

export interface MediaTask {
  task_id: string
  upstream_task_id?: string
  model: string
  media_type: string
  status: string
  billing_metric?: string
  reserved_cost: number
  actual_cost?: number | null
  billing_currency: string
  result_url?: string
  error_message?: string
  expires_at: string
  settled_at?: string | null
  created_at: string
  updated_at: string
}

export interface MediaTaskListParams {
  page?: number
  page_size?: number
  status?: string
  media_type?: string
  model?: string
  created_from?: string
  created_to?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export async function listTasks(
  params: MediaTaskListParams = {},
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<MediaTask>> {
  const { data } = await apiClient.get<PaginatedResponse<MediaTask>>('/media/tasks', {
    params,
    signal: options?.signal
  })
  return data
}

export interface CommonVideoGenerationMetadata {
  resolution?: string
  ratio?: string
  duration?: number
  content?: unknown[]
  [key: string]: unknown
}

export interface SubmitCommonVideoGenerationPayload {
  model: string
  prompt: string
  metadata?: CommonVideoGenerationMetadata
  task_id?: string
}

export interface CommonVideoGenerationTask {
  id?: string | number
  task_id: string
  object?: string
  model?: string
  status: string
  progress?: number | string
  created_at?: number
  updated_at?: number
  quota?: number
  actual_cost?: number | null
  result_url?: string
  fail_reason?: string
  error_message?: string
  data?: Record<string, unknown>
}

interface CommonVideoGenerationQueryResponse {
  code: string
  message: string
  data: CommonVideoGenerationTask
}

function commonStyleError(body: any, fallback: string): string {
  if (body?.error?.message) return String(body.error.message)
  if (body?.message) return String(body.message)
  if (typeof body?.error === 'string') return body.error
  return fallback
}

async function readJSON(res: Response): Promise<any> {
  const text = await res.text()
  if (!text) return null
  try {
    return JSON.parse(text)
  } catch {
    return null
  }
}

export async function submitCommonVideoGeneration(
  apiKey: string,
  payload: SubmitCommonVideoGenerationPayload
): Promise<CommonVideoGenerationTask> {
  const res = await fetch(buildGatewayUrl('/v1/video/generations'), {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${apiKey.trim()}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })
  const body = await readJSON(res)
  if (!res.ok) {
    throw new Error(commonStyleError(body, `HTTP ${res.status}`))
  }
  return body as CommonVideoGenerationTask
}

export async function getCommonVideoGeneration(
  apiKey: string,
  taskId: string
): Promise<CommonVideoGenerationTask> {
  const res = await fetch(buildGatewayUrl(`/v1/video/generations/${encodeURIComponent(taskId)}`), {
    headers: { Authorization: `Bearer ${apiKey.trim()}` }
  })
  const body = await readJSON(res)
  if (!res.ok) {
    throw new Error(commonStyleError(body, `HTTP ${res.status}`))
  }
  const wrapped = body as CommonVideoGenerationQueryResponse
  return wrapped.data ?? (body as CommonVideoGenerationTask)
}

export const mediaAPI = { listTasks, submitCommonVideoGeneration, getCommonVideoGeneration }
