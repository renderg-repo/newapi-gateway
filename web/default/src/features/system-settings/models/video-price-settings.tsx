import { memo, useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useUpdateOption } from '../hooks/use-update-option'

const OPTION_KEY = 'video_price_table.table'

type VideoPriceRow = {
  id: number
  model: string
  resolution: string
  hasVideo: boolean
  price: number
}

function parseTable(raw: string | undefined): Record<string, VideoPriceRow[]> {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as Record<string, Omit<VideoPriceRow, 'id'>[]>
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as unknown as Record<string, VideoPriceRow[]>
    }
  } catch {
    // fall through
  }
  return {}
}

function tableToRows(table: Record<string, Omit<VideoPriceRow, 'id'>[]>): VideoPriceRow[] {
  const rows: VideoPriceRow[] = []
  let id = 0
  for (const [model, variants] of Object.entries(table)) {
    for (const v of variants) {
      rows.push({ id: id++, model, resolution: v.resolution, hasVideo: v.hasVideo, price: v.price })
    }
  }
  return rows
}

function rowsToTable(rows: VideoPriceRow[]): Record<string, { resolution: string; hasVideo: boolean; price: number }[]> {
  const table: Record<string, { resolution: string; hasVideo: boolean; price: number }[]> = {}
  for (const row of rows) {
    if (!row.model?.trim()) continue
    const model = row.model.trim()
    if (!table[model]) table[model] = []
    table[model].push({
      resolution: row.resolution?.trim() || '480p',
      hasVideo: !!row.hasVideo,
      price: Number(row.price) || 0,
    })
  }
  return table
}

type VideoPriceSettingsProps = {
  defaultValue: string
}

export const VideoPriceSettings = memo(function VideoPriceSettings({
  defaultValue,
}: VideoPriceSettingsProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [rows, setRows] = useState<VideoPriceRow[]>([])
  const [editMode, setEditMode] = useState<'visual' | 'json'>('visual')
  const [jsonText, setJsonText] = useState('')
  const [jsonError, setJsonError] = useState('')
  const [nextId, setNextId] = useState(0)

  useEffect(() => {
    const table = parseTable(defaultValue)
    const initial = tableToRows(table)
    setRows(initial)
    setJsonText(JSON.stringify(table, null, 2))
    setNextId(initial.length)
  }, [defaultValue])

  const currentTable = useMemo(() => rowsToTable(rows), [rows])

  const updateRow = useCallback(
    (id: number, field: keyof VideoPriceRow, value: string | number | boolean) => {
      setRows((prev) => prev.map((r) => (r.id === id ? { ...r, [field]: value } : r)))
      setJsonText(JSON.stringify(rowsToTable(rows.map((r) => (r.id === id ? { ...r, [field]: value } : r))), null, 2))
      setJsonError('')
    },
    [rows]
  )

  const addRow = useCallback(() => {
    const newRow: VideoPriceRow = {
      id: nextId,
      model: '',
      resolution: '480p',
      hasVideo: false,
      price: 0,
    }
    setNextId((prev) => prev + 1)
    const next = [...rows, newRow]
    setRows(next)
    setJsonText(JSON.stringify(rowsToTable(next), null, 2))
    setJsonError('')
  }, [nextId, rows])

  const removeRow = useCallback(
    (id: number) => {
      const next = rows.filter((r) => r.id !== id)
      setRows(next)
      setJsonText(JSON.stringify(rowsToTable(next), null, 2))
      setJsonError('')
    },
    [rows]
  )

  const handleJsonChange = useCallback(
    (text: string) => {
      setJsonText(text)
      try {
        const parsed = JSON.parse(text) as unknown
        if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
          setJsonError(t('JSON must be an object'))
          return
        }
        const nextRows = tableToRows(parsed as Record<string, Omit<VideoPriceRow, 'id'>[]>)
        setRows(nextRows)
        setNextId(nextRows.length)
        setJsonError('')
      } catch (error) {
        setJsonError(error instanceof Error ? error.message : t('Invalid JSON'))
      }
    },
    [t]
  )

  const handleSave = useCallback(async () => {
    if (editMode === 'json' && jsonError) {
      toast.error(t('Please fix JSON errors before saving'))
      return
    }
    await updateOption.mutateAsync({
      key: OPTION_KEY,
      value: JSON.stringify(currentTable),
    })
  }, [currentTable, editMode, jsonError, t, updateOption])

  return (
    <div className='space-y-4'>
      <div className='text-muted-foreground text-sm'>
        {t(
          'Configure per-model video dimension pricing. Model ID, resolution, has-video flag, and price are all editable. Prices are in per million tokens.'
        )}
      </div>

      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex flex-wrap items-center gap-2'>
          {editMode === 'visual' ? (
            <Button variant='outline' size='sm' onClick={addRow}>
              <Plus className='mr-2 h-4 w-4' />
              {t('Add tier')}
            </Button>
          ) : null}
        </div>
        <Button
          variant='outline'
          size='sm'
          onClick={() => {
            setEditMode((prev) => (prev === 'visual' ? 'json' : 'visual'))
            if (editMode === 'visual') {
              setJsonText(JSON.stringify(currentTable, null, 2))
            }
          }}
        >
          {editMode === 'visual' ? t('Switch to JSON') : t('Switch to Visual')}
        </Button>
      </div>

      {editMode === 'visual' ? (
        <div className='overflow-hidden rounded-md border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Model ID')}</TableHead>
                <TableHead>{t('Resolution')}</TableHead>
                <TableHead>{t('Has video input')}</TableHead>
                <TableHead className='w-[180px]'>
                  {t('Price (/1M tokens)')}
                </TableHead>
                <TableHead className='w-[60px] text-right'>
                  {t('Actions')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className='text-muted-foreground py-8 text-center'
                  >
                    {t('No video models configured')}
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell>
                      <Input
                        value={row.model}
                        placeholder='doubao-seedance-2-0-260128'
                        onChange={(e) => updateRow(row.id, 'model', e.target.value)}
                      />
                    </TableCell>
                    <TableCell>
                      <Input
                        value={row.resolution}
                        placeholder='1080p / 720p / 480p'
                        onChange={(e) => updateRow(row.id, 'resolution', e.target.value)}
                      />
                    </TableCell>
                    <TableCell>
                      <div className='flex items-center gap-2'>
                        <Checkbox
                          id={`hasVideo-${row.id}`}
                          checked={row.hasVideo}
                          onCheckedChange={(checked) =>
                            updateRow(row.id, 'hasVideo', !!checked)
                          }
                        />
                        <label
                          htmlFor={`hasVideo-${row.id}`}
                          className='cursor-pointer text-sm'
                        >
                          {row.hasVideo ? t('Yes') : t('No')}
                        </label>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Input
                        type='number'
                        min={0}
                        step={0.5}
                        value={row.price}
                        onChange={(e) =>
                          updateRow(row.id, 'price', Number(e.target.value) || 0)
                        }
                      />
                    </TableCell>
                    <TableCell className='text-right'>
                      <Button
                        variant='ghost'
                        size='icon'
                        onClick={() => removeRow(row.id)}
                        aria-label={t('Delete')}
                      >
                        <Trash2 className='text-destructive h-4 w-4' />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className='space-y-2'>
          <Textarea
            value={jsonText}
            onChange={(e) => handleJsonChange(e.target.value)}
            className='font-mono text-sm'
            rows={12}
            spellCheck={false}
          />
          {jsonError && <p className='text-destructive text-sm'>{jsonError}</p>}
        </div>
      )}

      <div className='flex justify-end'>
        <Button
          onClick={handleSave}
          disabled={
            updateOption.isPending || (editMode === 'json' && !!jsonError)
          }
        >
          {t('Save video prices')}
        </Button>
      </div>
    </div>
  )
})