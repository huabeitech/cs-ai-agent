"use client";

import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, Resolver, useForm } from "react-hook-form";
import { z } from "zod/v4";

import { ProjectDialog } from "@/components/project-dialog";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  type CreateTagPayload,
  fetchTag,
  fetchTagsAll,
  type Tag,
} from "@/lib/api/admin";

type TagFormDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateTagPayload) => Promise<void>;
};

const emptyForm: EditForm = {
  parentId: "0",
  name: "",
  remark: "",
};

const tagFormSchema = z.object({
  parentId: z.string(),
  name: z.string().trim().min(1, "标签名称不能为空"),
  remark: z.string(),
});

type EditForm = z.infer<typeof tagFormSchema>;
const editFormResolver = zodResolver(tagFormSchema as never) as Resolver<
  z.input<typeof tagFormSchema>,
  undefined,
  z.output<typeof tagFormSchema>
>;

function buildForm(item: Tag | null): EditForm {
  if (!item) {
    return emptyForm;
  }

  return {
    parentId: String(item.parentId),
    name: item.name,
    remark: item.remark,
  };
}

function buildPayload(form: EditForm): CreateTagPayload {
  return {
    parentId: Number(form.parentId),
    name: form.name.trim(),
    remark: form.remark.trim(),
    status: 0,
  };
}

type TagTreeNode = Tag & {
  children: TagTreeNode[];
  depth: number;
};

function buildTagTree(tags: Tag[]): TagTreeNode[] {
  const tagMap = new Map<number, TagTreeNode>();
  const roots: TagTreeNode[] = [];

  tags.forEach((tag) => {
    tagMap.set(tag.id, { ...tag, children: [], depth: 0 });
  });

  tags.forEach((tag) => {
    const node = tagMap.get(tag.id)!;
    if (tag.parentId === 0 || !tagMap.has(tag.parentId)) {
      roots.push(node);
    } else {
      const parent = tagMap.get(tag.parentId);
      if (parent) {
        parent.children.push(node);
      }
    }
  });

  function setDepth(nodes: TagTreeNode[], depth: number) {
    nodes.forEach((node) => {
      node.depth = depth;
      setDepth(node.children, depth + 1);
    });
  }
  setDepth(roots, 0);

  return roots;
}

function flattenTreeForSelect(
  nodes: TagTreeNode[],
  excludeId?: number,
): { id: number; name: string; depth: number }[] {
  const result: { id: number; name: string; depth: number }[] = [];
  function traverse(node: TagTreeNode) {
    if (node.id !== excludeId) {
      result.push({ id: node.id, name: node.name, depth: node.depth });
      node.children.forEach(traverse);
    }
  }
  nodes.forEach(traverse);
  return result;
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: TagFormDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <TagFormDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type TagFormDialogBodyProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateTagPayload) => Promise<void>;
};

function TagFormDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: TagFormDialogBodyProps) {
  const formId = "tag-edit-form";
  const [loading, setLoading] = useState(false);
  const [parentTags, setParentTags] = useState<
    { id: number; name: string; depth: number }[]
  >([]);
  const form = useForm<
    z.input<typeof tagFormSchema>,
    undefined,
    z.output<typeof tagFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  });
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form;

  useEffect(() => {
    async function loadParentTags() {
      try {
        const data = await fetchTagsAll();
        const tree = buildTagTree(data);
        const flatList = flattenTreeForSelect(tree, itemId ?? undefined);
        setParentTags(flatList);
      } catch (error) {
        console.error("Failed to load parent tags:", error);
      }
    }
    void loadParentTags();
  }, [itemId]);

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm);
        return;
      }
      setLoading(true);
      try {
        const data = await fetchTag(itemId);
        reset(buildForm(data));
      } catch (error) {
        console.error("Failed to load tag:", error);
      } finally {
        setLoading(false);
      }
    }
    void loadDetail();
  }, [itemId, reset]);

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values);
    await onSubmit(payload);
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑标签" : "新建标签"}
      size="md"
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
        <form
          id={formId}
          onSubmit={handleSubmit(onFormSubmit)}
          className="space-y-4"
        >
          <Field data-invalid={!!errors.parentId}>
            <FieldLabel htmlFor="tag-parent-id">父标签</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="parentId"
                render={({ field }) => (
                  <select
                    id="tag-parent-id"
                    value={field.value}
                    onChange={(e) => field.onChange(e.target.value)}
                    className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
                  >
                    <option value="0">无（顶级标签）</option>
                    {parentTags.map((tag) => (
                      <option key={tag.id} value={String(tag.id)}>
                        {"　".repeat(tag.depth)}
                        {tag.name}
                      </option>
                    ))}
                  </select>
                )}
              />
              <FieldError errors={[errors.parentId]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="tag-name">标签名称</FieldLabel>
            <FieldContent>
              <Input
                id="tag-name"
                placeholder="请输入标签名称"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="tag-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="tag-remark"
                placeholder="请输入备注（可选）"
                rows={3}
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  );
}
