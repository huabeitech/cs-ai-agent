"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { ProjectDialog } from "@/components/project-dialog"
import { ContentEditor } from "@/components/content-editor"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  type KnowledgeDocument,
  type CreateKnowledgeDocumentPayload,
  fetchKnowledgeDocument,
} from "@/lib/api/admin"
import {
  KnowledgeDocumentContentType,
} from "@/lib/generated/enums"

type DocumentEditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  knowledgeBaseId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateKnowledgeDocumentPayload) => Promise<void>
}

const emptyForm: EditForm = {
  title: "",
  contentType: KnowledgeDocumentContentType.HTML,
  content: "",
}

const knowledgeDocumentFormSchema = z.object({
  title: z.string().trim().min(1, "标题不能为空").max(255, "标题最多255个字符"),
  contentType: z.string().trim().min(1, "请选择内容类型"),
  content: z.string().trim().min(1, "内容不能为空"),
})

type EditForm = z.infer<typeof knowledgeDocumentFormSchema>
const editFormResolver = zodResolver(knowledgeDocumentFormSchema as never) as Resolver<
  z.input<typeof knowledgeDocumentFormSchema>,
  undefined,
  z.output<typeof knowledgeDocumentFormSchema>
>

function buildForm(item: KnowledgeDocument | null): EditForm {
  if (!item) {
    return emptyForm
  }

  return {
    title: item.title,
    contentType: item.contentType || KnowledgeDocumentContentType.HTML,
    content: item.content || "",
  }
}

function buildPayload(form: EditForm, knowledgeBaseId: number): CreateKnowledgeDocumentPayload {
  return {
    knowledgeBaseId,
    title: form.title.trim(),
    contentType: form.contentType,
    content: form.content.trim(),
  }
}

export function DocumentEditDialog({
  open,
  saving,
  itemId,
  knowledgeBaseId,
  onOpenChange,
  onSubmit,
}: DocumentEditDialogProps) {
  if (!open || !knowledgeBaseId) {
    return null
  }

  return (
    <DocumentFormDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      itemId={itemId}
      knowledgeBaseId={knowledgeBaseId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type DocumentFormDialogBodyProps = {
  saving: boolean
  itemId: number | null
  knowledgeBaseId: number
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateKnowledgeDocumentPayload) => Promise<void>
}

function DocumentFormDialogBody({
  saving,
  itemId,
  knowledgeBaseId,
  onOpenChange,
  onSubmit,
}: DocumentFormDialogBodyProps) {
  const formId = "knowledge-document-edit-form"
  const [loading, setLoading] = useState(false)
  const form = useForm<
    z.input<typeof knowledgeDocumentFormSchema>,
    undefined,
    z.output<typeof knowledgeDocumentFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    setValue,
    watch,
    formState: { errors },
  } = form

  const contentType = watch("contentType")
  const content = watch("content")

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchKnowledgeDocument(itemId)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load knowledge document:", error)
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload({ ...values, contentType, content }, knowledgeBaseId)
    await onSubmit(payload)
  }

  return (
    <ProjectDialog
      open={true}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑文档" : "新建文档"}
      size="xl"
      allowFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : itemId ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <Field data-invalid={!!errors.title}>
            <FieldLabel htmlFor="doc-title">标题</FieldLabel>
            <FieldContent>
              <Input
                id="doc-title"
                placeholder="文档标题"
                aria-invalid={!!errors.title}
                {...register("title")}
              />
              <FieldError errors={[errors.title]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.content}>
            <FieldLabel htmlFor="doc-content">内容</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="content"
                render={({ field }) => (
                  <ContentEditor
                    value={{
                      mode:
                        contentType === KnowledgeDocumentContentType.Markdown
                          ? KnowledgeDocumentContentType.Markdown
                          : KnowledgeDocumentContentType.HTML,
                      raw: field.value ?? "",
                    }}
                    onChange={(next) => {
                      field.onChange(next.raw)
                      setValue("contentType", next.mode, {
                        shouldDirty: true,
                        shouldValidate: true,
                      })
                    }}
                    placeholder={
                      contentType === KnowledgeDocumentContentType.Markdown
                        ? "输入 Markdown 内容..."
                        : "输入 HTML 内容..."
                    }
                    disabled={saving}
                  />
                )}
              />
              <FieldError errors={[errors.content]} />
            </FieldContent>
          </Field>

        </form>
      )}
    </ProjectDialog>
  )
}
