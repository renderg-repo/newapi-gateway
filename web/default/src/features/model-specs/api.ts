import { api } from '@/lib/api'
import type {
  GetModelSpecsParams,
  GetModelSpecsResponse,
  GetModelSpecResponse,
  ModelSpec,
} from './types'

/**
 * Get paginated list of model specs
 */
export async function getModelSpecs(
  params: GetModelSpecsParams = {}
): Promise<GetModelSpecsResponse> {
  const res = await api.get('/api/model-catalog/admin/specs', { params })
  return res.data
}

/**
 * Get single model spec by ID
 */
export async function getModelSpec(
  id: number
): Promise<GetModelSpecResponse> {
  const res = await api.get(`/api/model-catalog/admin/specs/${id}`)
  return res.data
}

/**
 * Create new model spec
 */
export async function createModelSpec(
  data: Partial<ModelSpec>
): Promise<{ success: boolean; message?: string; data?: ModelSpec }> {
  const res = await api.post('/api/model-catalog/admin/specs', data)
  return res.data
}

/**
 * Update existing model spec
 */
export async function updateModelSpec(
  data: Partial<ModelSpec> & { id: number }
): Promise<{ success: boolean; message?: string; data?: ModelSpec }> {
  const res = await api.put('/api/model-catalog/admin/specs', data)
  return res.data
}

/**
 * Delete model spec
 */
export async function deleteModelSpec(
  id: number
): Promise<{ success: boolean; message?: string }> {
  const res = await api.delete(`/api/model-catalog/admin/specs/${id}`)
  return res.data
}
