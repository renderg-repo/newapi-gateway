import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { type Row } from '@tanstack/react-table'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { deleteModelSpec } from '../api'
import { modelSpecsQueryKeys } from '../lib/query-keys'
import type { ModelSpec } from '../types'

interface DataTableRowActionsProps {
  row: Row<ModelSpec>
  onEdit: (spec: ModelSpec) => void
}

export function DataTableRowActions({ row, onEdit }: DataTableRowActionsProps) {
  const { t } = useTranslation()
  const spec = row.original
  const queryClient = useQueryClient()
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)

  const handleDelete = async () => {
    try {
      const res = await deleteModelSpec(spec.id)
      if (res.success) {
        toast.success(t('Model spec deleted successfully'))
        queryClient.invalidateQueries({ queryKey: modelSpecsQueryKeys.lists() })
      } else {
        toast.error(res.message || t('Failed to delete model spec'))
      }
    } catch (error: unknown) {
      toast.error((error as Error)?.message || t('Failed to delete model spec'))
    }
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button
              variant='ghost'
              className='data-popup-open:bg-muted flex h-8 w-8 p-0'
            />
          }
        >
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>{t('Open menu')}</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-48'>
          <DropdownMenuItem onClick={() => onEdit(spec)}>
            {t('Edit')}
            <DropdownMenuShortcut>
              <Pencil size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={(e) => {
              e.preventDefault()
              setDeleteConfirmOpen(true)
            }}
            className='text-destructive focus:text-destructive'
          >
            {t('Delete')}
            <DropdownMenuShortcut>
              <Trash2 size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ConfirmDialog
        open={deleteConfirmOpen}
        onOpenChange={setDeleteConfirmOpen}
        title={t('Delete Model Spec')}
        desc={`${t('Are you sure you want to delete')} "${spec.model_name}"? ${t('This action cannot be undone.')}`}
        confirmText={t('Delete')}
        destructive
        handleConfirm={() => {
          void handleDelete()
          setDeleteConfirmOpen(false)
        }}
      />
    </>
  )
}
