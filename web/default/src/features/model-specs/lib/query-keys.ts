import type { GetModelSpecsParams } from '../types'

/**
 * React Query cache keys for model specs
 */
export const modelSpecsQueryKeys = {
  all: ['model-specs'] as const,
  lists: () => [...modelSpecsQueryKeys.all, 'list'] as const,
  list: (filters: GetModelSpecsParams) =>
    [...modelSpecsQueryKeys.lists(), filters] as const,
  detail: (id: number) => [...modelSpecsQueryKeys.all, 'detail', id] as const,
}
