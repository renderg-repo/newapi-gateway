import { type ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { formatTimestampToDate } from '@/lib/format'
import { getLobeIcon } from '@/lib/lobe-icon'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { StatusBadge } from '@/components/status-badge'
import type { ModelSpec } from '../types'
import { parseCapabilities } from '../lib'
import { DataTableRowActions } from './data-table-row-actions'

/**
 * Render limited items with "and X more" indicator
 */
function renderLimitedItems(
  items: React.ReactNode[],
  maxDisplay: number = 2
): React.ReactNode {
  if (items.length === 0)
    return <span className='text-muted-foreground text-xs'>-</span>

  const displayed = items.slice(0, maxDisplay)
  const remaining = items.length - maxDisplay

  return (
    <div className='flex max-w-full items-center gap-1 overflow-x-auto'>
      {displayed}
      {remaining > 0 && (
        <StatusBadge
          label={`+${remaining}`}
          variant='neutral'
          size='sm'
          copyable={false}
          className='flex-shrink-0'
        />
      )}
    </div>
  )
}

/**
 * Generate model specs columns configuration
 */
interface UseModelSpecsColumnsOptions {
  onEdit?: (spec: ModelSpec) => void
}

export function useModelSpecsColumns(
  options: UseModelSpecsColumnsOptions = {}
): ColumnDef<ModelSpec>[] {
  const { t } = useTranslation()

  return [
    // Checkbox column
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          indeterminate={table.getIsSomePageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label='Select all'
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label='Select row'
        />
      ),
      enableSorting: false,
      enableHiding: false,
      size: 40,
    },

    // ID column
    {
      accessorKey: 'id',
      meta: { label: t('ID'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='ID' />
      ),
      cell: ({ row }) => {
        const id = row.getValue('id') as number
        return (
          <StatusBadge
            label={String(id)}
            variant='neutral'
            copyText={String(id)}
            size='sm'
            className='font-mono'
          />
        )
      },
      size: 80,
    },

    // Icon column
    {
      accessorKey: 'icon',
      meta: { label: t('Icon'), mobileHidden: true },
      header: t('Icon'),
      cell: ({ row }) => {
        const iconKey = row.getValue('icon') as string
        const modelName = row.getValue('model_name') as string
        const icon = getLobeIcon(iconKey || modelName?.[0] || 'N', 20)
        return <div className='flex items-center justify-center'>{icon}</div>
      },
      size: 70,
      enableSorting: false,
    },

    // Model Name column
    {
      accessorKey: 'model_name',
      meta: { label: t('Model Name'), mobileTitle: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Model Name')} />
      ),
      cell: ({ row }) => {
        const name = row.getValue('model_name') as string
        return (
          <StatusBadge
            label={name}
            variant='neutral'
            copyText={name}
            size='sm'
            className='font-mono'
          />
        )
      },
      minSize: 200,
    },

    // Description column
    {
      accessorKey: 'description',
      meta: { label: t('Description'), mobileHidden: true },
      header: t('Description'),
      cell: ({ row }) => {
        const description = row.getValue('description') as string
        if (!description) {
          return <span className='text-muted-foreground text-xs'>-</span>
        }
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger render={<div className='cursor-help' />}>
                <span className='text-muted-foreground line-clamp-1 max-w-[200px] text-sm'>
                  {description}
                </span>
              </TooltipTrigger>
              <TooltipContent>
                <p className='max-w-xs'>{description}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )
      },
      size: 180,
      enableSorting: false,
    },

    // Context Length column
    {
      accessorKey: 'context_length',
      meta: { label: t('Context Length'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Context')} />
      ),
      cell: ({ row }) => {
        const val = row.getValue('context_length') as number
        if (!val) {
          return <span className='text-muted-foreground text-xs'>-</span>
        }
        return (
          <span className='font-mono text-sm'>
            {val >= 1000 ? `${(val / 1000).toFixed(0)}k` : val}
          </span>
        )
      },
      size: 120,
    },

    // Max Output Tokens column
    {
      accessorKey: 'max_output_tokens',
      meta: { label: t('Max Output'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Max Output')} />
      ),
      cell: ({ row }) => {
        const val = row.getValue('max_output_tokens') as number
        if (!val) {
          return <span className='text-muted-foreground text-xs'>-</span>
        }
        return (
          <span className='font-mono text-sm'>
            {val >= 1000 ? `${(val / 1000).toFixed(0)}k` : val}
          </span>
        )
      },
      size: 120,
    },

    // Capabilities column
    {
      accessorKey: 'capabilities',
      meta: { label: t('Capabilities'), mobileHidden: true },
      header: t('Capabilities'),
      cell: ({ row }) => {
        const capsStr = row.getValue('capabilities') as string
        const caps = parseCapabilities(capsStr)

        if (caps.length === 0) {
          return <span className='text-muted-foreground text-xs'>-</span>
        }

        const capBadges = caps.map((cap, idx) => (
          <StatusBadge key={idx} label={cap} autoColor={cap} size='sm' />
        ))

        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger render={<div />}>
                {renderLimitedItems(capBadges, 2)}
              </TooltipTrigger>
              {caps.length > 2 && (
                <TooltipContent
                  side='top'
                  className='border-border bg-popover max-h-48 max-w-[320px] overflow-y-auto p-2'
                >
                  <div className='flex flex-wrap gap-1'>{capBadges}</div>
                </TooltipContent>
              )}
            </Tooltip>
          </TooltipProvider>
        )
      },
      size: 160,
      enableSorting: false,
    },

    // Release Date column
    {
      accessorKey: 'release_date',
      meta: { label: t('Release'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Release')} />
      ),
      cell: ({ row }) => {
        const date = row.getValue('release_date') as string
        if (!date) {
          return <span className='text-muted-foreground text-xs'>-</span>
        }
        return <span className='text-sm'>{date}</span>
      },
      size: 120,
    },

    // Status column
    {
      accessorKey: 'status',
      meta: { label: t('Status'), mobileBadge: true },
      header: t('Status'),
      cell: ({ row }) => {
        const status = row.getValue('status') as number
        const isEnabled = status === 1
        return (
          <StatusBadge
            label={isEnabled ? t('Enabled') : t('Disabled')}
            variant={isEnabled ? 'success' : 'neutral'}
            showDot={isEnabled}
            size='sm'
            copyable={false}
          />
        )
      },
      size: 100,
      enableSorting: false,
    },

    // Created Time column
    {
      accessorKey: 'created_time',
      meta: { label: t('Created'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Created')} />
      ),
      cell: ({ row }) => {
        const timestamp = row.getValue('created_time') as number
        return (
          <div className='min-w-[140px] font-mono text-sm'>
            {formatTimestampToDate(timestamp)}
          </div>
        )
      },
      size: 160,
    },

    // Updated Time column
    {
      accessorKey: 'updated_time',
      meta: { label: t('Updated'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Updated')} />
      ),
      cell: ({ row }) => {
        const timestamp = row.getValue('updated_time') as number
        return (
          <div className='min-w-[140px] font-mono text-sm'>
            {formatTimestampToDate(timestamp)}
          </div>
        )
      },
      size: 160,
    },

    // Actions column
    {
      id: 'actions',
      cell: ({ row }) => {
        return <DataTableRowActions row={row} onEdit={options.onEdit ?? (() => {})} />
      },
      size: 100,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
