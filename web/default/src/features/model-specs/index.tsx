import { useState } from 'react'
import { Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { ModelSpecsTable } from './components/model-specs-table'
import { ModelSpecMutateDrawer } from './components/drawers/model-spec-mutate-drawer'
import type { ModelSpec } from './types'

interface ModelSpecsProps {
  onCreate?: () => void
}

export function ModelSpecs({ onCreate }: ModelSpecsProps = {}) {
  const { t } = useTranslation()
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [currentRow, setCurrentRow] = useState<ModelSpec | null>(null)

  const handleCreate = () => {
    setCurrentRow(null)
    setDrawerOpen(true)
    onCreate?.()
  }

  const handleEdit = (spec: ModelSpec) => {
    setCurrentRow(spec)
    setDrawerOpen(true)
  }

  return (
    <>
      <div className='flex justify-end'>
        <Button onClick={handleCreate} size='sm'>
          <Plus className='h-4 w-4' />
          {t('Create Spec')}
        </Button>
      </div>
      <ModelSpecsTable onEdit={handleEdit} />

      <ModelSpecMutateDrawer
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
        currentRow={currentRow}
      />
    </>
  )
}
