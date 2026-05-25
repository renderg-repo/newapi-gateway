import { z } from 'zod'

// ============================================================================
// Model Spec Types
// ============================================================================

/**
 * Model spec entity from API (sidecar model_specs table)
 */
export interface ModelSpec {
  id: number
  model_name: string
  context_length: number
  max_output_tokens: number
  capabilities?: string
  description?: string
  icon?: string
  release_date?: string
  knowledge_cutoff?: string
  parameter_count?: string
  status: number
  created_time: number
  updated_time: number
}

/**
 * Get model specs list parameters
 */
export interface GetModelSpecsParams {
  keyword?: string
  p?: number
  page_size?: number
}

/**
 * Get model specs response
 */
export interface GetModelSpecsResponse {
  success: boolean
  message?: string
  data?: {
    items: ModelSpec[]
    total: number
    page: number
    page_size: number
  }
}

/**
 * Get model spec detail response
 */
export interface GetModelSpecResponse {
  success: boolean
  message?: string
  data?: ModelSpec
}

// ============================================================================
// Form Data Types
// ============================================================================

/**
 * Model spec form schema
 */
export const modelSpecFormSchema = z.object({
  id: z.number().optional(),
  model_name: z.string().min(1, 'Model name is required'),
  context_length: z.number().min(0),
  max_output_tokens: z.number().min(0),
  capabilities: z.array(z.string()),
  description: z.string(),
  icon: z.string(),
  release_date: z.string(),
  knowledge_cutoff: z.string(),
  parameter_count: z.string(),
  status: z.boolean(),
})

export type ModelSpecFormValues = z.infer<typeof modelSpecFormSchema>

// ============================================================================
// Capability Options
// ============================================================================

export const CAPABILITY_OPTIONS = [
  { value: 'vision', label: 'Vision' },
  { value: 'reasoning', label: 'Reasoning' },
  { value: 'code', label: 'Code' },
  { value: 'tool-use', label: 'Tool Use' },
  { value: 'audio', label: 'Audio' },
  { value: 'image', label: 'Image' },
  { value: 'video', label: 'Video' },
  { value: 'search', label: 'Search' },
  { value: 'long-context', label: 'Long Context' },
] as const
