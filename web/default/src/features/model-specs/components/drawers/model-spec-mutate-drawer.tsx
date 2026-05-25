import { useEffect, useState, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { createModelSpec, updateModelSpec } from '../../api'
import { modelSpecsQueryKeys } from '../../lib/query-keys'
import { serializeCapabilities, parseCapabilities } from '../../lib'
import { CAPABILITY_OPTIONS, modelSpecFormSchema } from '../../types'
import type { ModelSpecFormValues, ModelSpec } from '../../types'

interface ModelSpecMutateDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: ModelSpec | null
}

export function ModelSpecMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: ModelSpecMutateDrawerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEditing = Boolean(currentRow?.id)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<ModelSpecFormValues>({
    resolver: zodResolver(modelSpecFormSchema),
    defaultValues: {
      model_name: '',
      context_length: 0,
      max_output_tokens: 0,
      capabilities: [],
      description: '',
      icon: '',
      release_date: '',
      knowledge_cutoff: '',
      parameter_count: '',
      status: true,
    },
  })

  // Reset form when drawer opens/closes or currentRow changes
  useEffect(() => {
    if (open && currentRow) {
      form.reset({
        model_name: currentRow.model_name,
        context_length: currentRow.context_length,
        max_output_tokens: currentRow.max_output_tokens,
        capabilities: parseCapabilities(currentRow.capabilities),
        description: currentRow.description || '',
        icon: currentRow.icon || '',
        release_date: currentRow.release_date || '',
        knowledge_cutoff: currentRow.knowledge_cutoff || '',
        parameter_count: currentRow.parameter_count || '',
        status: currentRow.status === 1,
      })
    } else if (open && !currentRow) {
      form.reset({
        model_name: '',
        context_length: 0,
        max_output_tokens: 0,
        capabilities: [],
        description: '',
        icon: '',
        release_date: '',
        knowledge_cutoff: '',
        parameter_count: '',
        status: true,
      })
    }
  }, [open, currentRow, form])

  const onSubmit = useCallback(
    async (values: ModelSpecFormValues) => {
      setIsSubmitting(true)
      try {
        const submitData = {
          ...values,
          id: isEditing ? currentRow!.id : undefined,
          capabilities: serializeCapabilities(values.capabilities),
          status: values.status ? 1 : 0,
        }

        const response = isEditing
          ? await updateModelSpec({ ...submitData, id: currentRow!.id })
          : await createModelSpec(submitData)

        if (response.success) {
          toast.success(
            isEditing
              ? t('Model spec updated successfully')
              : t('Model spec created successfully')
          )
          queryClient.invalidateQueries({
            queryKey: modelSpecsQueryKeys.lists(),
          })
          onOpenChange(false)
        } else {
          toast.error(response.message || t('Operation failed'))
        }
      } catch (error: unknown) {
        toast.error((error as Error)?.message || t('Operation failed'))
      } finally {
        setIsSubmitting(false)
      }
    },
    [isEditing, currentRow, queryClient, onOpenChange, t]
  )

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='flex h-dvh w-full flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl'>
        <SheetHeader className='border-b px-4 py-3 text-start sm:px-6 sm:py-4'>
          <SheetTitle>
            {isEditing ? t('Edit Model Spec') : t('Create Model Spec')}
          </SheetTitle>
          <SheetDescription>
            {isEditing
              ? t("Update model spec and click save when you're done.")
              : t(
                  'Add a new model spec to enrich model catalog display.'
                )}
          </SheetDescription>
        </SheetHeader>

        <Form {...form}>
          <form
            id='model-spec-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='flex-1 space-y-4 overflow-y-auto px-3 py-3 pb-4 sm:space-y-6 sm:px-4'
          >
            {/* Basic Information */}
            <div className='space-y-4'>
              <h3 className='text-sm font-semibold'>{t('Basic Information')}</h3>

              <FormField
                control={form.control}
                name='model_name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Model Name *')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('gpt-4, claude-3-opus, etc.')}
                        {...field}
                        disabled={isEditing}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('The model name must match an existing pricing model')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='description'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Description')}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t('Describe this model...')}
                        rows={3}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='icon'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Icon')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('OpenAI, Anthropic, etc.')}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription className='text-xs'>
                      {t('@lobehub/icons key')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* Specifications */}
            <div className='space-y-4'>
              <h3 className='text-sm font-semibold'>{t('Specifications')}</h3>

              <div className='grid grid-cols-2 gap-4'>
                <FormField
                  control={form.control}
                  name='context_length'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Context Length')}</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          placeholder='0'
                          {...field}
                          onChange={(e) =>
                            field.onChange(parseInt(e.target.value) || 0)
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='max_output_tokens'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Max Output Tokens')}</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          placeholder='0'
                          {...field}
                          onChange={(e) =>
                            field.onChange(parseInt(e.target.value) || 0)
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name='capabilities'
                render={() => (
                  <FormItem>
                    <FormLabel>{t('Capabilities')}</FormLabel>
                    <div className='grid grid-cols-3 gap-3'>
                      {CAPABILITY_OPTIONS.map((option) => (
                        <FormField
                          key={option.value}
                          control={form.control}
                          name='capabilities'
                          render={({ field }) => {
                            return (
                              <FormItem
                                key={option.value}
                                className='flex flex-row items-start space-x-2 space-y-0'
                              >
                                <FormControl>
                                  <Checkbox
                                    checked={field.value?.includes(option.value)}
                                    onCheckedChange={(checked) => {
                                      const current = field.value || []
                                      if (checked) {
                                        field.onChange([...current, option.value])
                                      } else {
                                        field.onChange(
                                          current.filter(
                                            (v) => v !== option.value
                                          )
                                        )
                                      }
                                    }}
                                  />
                                </FormControl>
                                <Label className='cursor-pointer font-normal'>
                                  {option.label}
                                </Label>
                              </FormItem>
                            )
                          }}
                        />
                      ))}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* Metadata */}
            <div className='space-y-4'>
              <h3 className='text-sm font-semibold'>{t('Metadata')}</h3>

              <div className='grid grid-cols-2 gap-4'>
                <FormField
                  control={form.control}
                  name='release_date'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Release Date')}</FormLabel>
                      <FormControl>
                        <Input placeholder='2024-01' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='knowledge_cutoff'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Knowledge Cutoff')}</FormLabel>
                      <FormControl>
                        <Input placeholder='2024-06' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name='parameter_count'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Parameter Count')}</FormLabel>
                    <FormControl>
                      <Input placeholder='1.76T' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* Status */}
            <FormField
              control={form.control}
              name='status'
              render={({ field }) => (
                <FormItem className='flex items-center justify-between rounded-lg border p-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base'>{t('Enabled')}</FormLabel>
                    <FormDescription>
                      {t('Enable or disable this model spec')}
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          </form>
        </Form>

        <SheetFooter className='grid grid-cols-2 gap-2 border-t px-4 py-3 sm:flex sm:px-6 sm:py-4'>
          <SheetClose
            render={<Button variant='outline' disabled={isSubmitting} />
            }
          >
            {t('Cancel')}
          </SheetClose>
          <Button form='model-spec-form' type='submit' disabled={isSubmitting}>
            {isSubmitting && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {isEditing ? t('Update') : t('Create')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
