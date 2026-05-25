import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  getCoreRowModel,
  useReactTable,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table'
import { useMediaQuery } from '@/hooks'
import { useTranslation } from 'react-i18next'
import { DataTablePage } from '@/components/data-table'
import { getModelSpecs } from '../api'
import { modelSpecsQueryKeys } from '../lib/query-keys'
import type { ModelSpec } from '../types'
import { useModelSpecsColumns } from './model-specs-columns'

const DEFAULT_PAGE_SIZE = 10

interface ModelSpecsTableProps {
  onEdit: (spec: ModelSpec) => void
}

export function ModelSpecsTable({ onEdit }: ModelSpecsTableProps) {
  const { t } = useTranslation()
  const isMobile = useMediaQuery('(max-width: 640px)')

  // Table state
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({
    description: false,
    release_date: false,
    knowledge_cutoff: false,
    parameter_count: false,
    created_time: false,
  })
  const [rowSelection, setRowSelection] = useState({})
  const [globalFilter, setGlobalFilter] = useState('')
  const [pagination, setPagination] = useState({
    pageIndex: 0,
    pageSize: isMobile ? 10 : DEFAULT_PAGE_SIZE,
  })

  const shouldSearch = Boolean(globalFilter?.trim())

  // Fetch data
  const { data, isLoading, isFetching } = useQuery({
    queryKey: modelSpecsQueryKeys.list({
      keyword: globalFilter,
      p: pagination.pageIndex + 1,
      page_size: pagination.pageSize,
    }),
    queryFn: () =>
      getModelSpecs({
        keyword: shouldSearch ? globalFilter : undefined,
        p: pagination.pageIndex + 1,
        page_size: pagination.pageSize,
      }),
    placeholderData: (previousData) => previousData,
  })

  const specs = data?.data?.items || []
  const totalCount = data?.data?.total || 0

  // Columns configuration
  const columns = useModelSpecsColumns({ onEdit })

  // React Table instance
  const table = useReactTable({
    data: specs,
    columns,
    pageCount: Math.ceil(totalCount / pagination.pageSize),
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      pagination,
      globalFilter,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange: setPagination,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    manualFiltering: true,
  })

  return (
    <DataTablePage
      table={table}
      columns={columns}
      isLoading={isLoading}
      isFetching={isFetching}
      emptyTitle={t('No Model Specs Found')}
      emptyDescription={t(
        'No model specs available. Create your first model spec to get started.'
      )}
      skeletonKeyPrefix='model-spec-skeleton'
      applyHeaderSize
      toolbarProps={{
        searchPlaceholder: t('Filter by model name...'),
      }}
    />
  )
}
