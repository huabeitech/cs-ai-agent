"use client"

import { CheckIcon, Loader2Icon, TagIcon } from "lucide-react"
import { useMemo, useState } from "react"
import { toast } from "sonner"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  addConversationTag,
  removeConversationTag,
  type AgentConversation,
  type AgentConversationTag,
} from "@/lib/api/agent"
import { type Tag } from "@/lib/api/admin"
import { cn } from "@/lib/utils"

type TagNode = Tag & {
  children: TagNode[]
  depth: number
}

function buildTagTree(tags: Tag[]): TagNode[] {
  const sorted = [...tags].sort((a, b) => {
    if (a.sortNo !== b.sortNo) {
      return a.sortNo - b.sortNo
    }
    if (a.parentId !== b.parentId) {
      return a.parentId - b.parentId
    }
    return a.id - b.id
  })

  const tagMap = new Map<number, TagNode>()
  sorted.forEach((tag) => {
    tagMap.set(tag.id, { ...tag, children: [], depth: 0 })
  })

  const roots: TagNode[] = []
  sorted.forEach((tag) => {
    const node = tagMap.get(tag.id)
    if (!node) {
      return
    }
    if (tag.parentId === 0 || !tagMap.has(tag.parentId)) {
      roots.push(node)
      return
    }
    const parent = tagMap.get(tag.parentId)
    if (!parent) {
      roots.push(node)
      return
    }
    node.depth = parent.depth + 1
    parent.children.push(node)
  })

  return roots
}

function flattenTagTree(nodes: TagNode[]): TagNode[] {
  const result: TagNode[] = []
  const walk = (items: TagNode[]) => {
    items.forEach((item) => {
      result.push(item)
      if (item.children.length > 0) {
        walk(item.children)
      }
    })
  }
  walk(nodes)
  return result
}

type ConversationTagPickerProps = {
  conversation: AgentConversation
  availableTags: Tag[]
  loading?: boolean
  onTagsChange: (tags: AgentConversationTag[]) => void
}

export function ConversationTagPicker({
  conversation,
  availableTags,
  loading = false,
  onTagsChange,
}: ConversationTagPickerProps) {
  const [pendingTagId, setPendingTagId] = useState<number | null>(null)

  const activeTags = useMemo(
    () => availableTags.filter((item) => item.status === 0),
    [availableTags]
  )
  const flattenedTags = useMemo(
    () => flattenTagTree(buildTagTree(activeTags)),
    [activeTags]
  )
  const selectedTagIds = useMemo(
    () => new Set((conversation.tags ?? []).map((item) => item.id)),
    [conversation.tags]
  )

  async function handleToggle(tag: Tag) {
    if (pendingTagId !== null) {
      return
    }

    const exists = selectedTagIds.has(tag.id)
    const currentTags = conversation.tags ?? []
    const nextTags = exists
      ? currentTags.filter((item) => item.id !== tag.id)
      : [...currentTags, { id: tag.id, name: tag.name }]

    setPendingTagId(tag.id)
    try {
      if (exists) {
        await removeConversationTag({
          conversationId: conversation.id,
          tagId: tag.id,
        })
      } else {
        await addConversationTag({
          conversationId: conversation.id,
          tagId: tag.id,
        })
      }
      onTagsChange(nextTags)
      toast.success(exists ? "已移除会话标签" : "已添加会话标签")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新会话标签失败")
    } finally {
      setPendingTagId(null)
    }
  }

  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-6 shrink-0 text-muted-foreground"
            aria-label="编辑会话标签"
          />
        }
      >
        <TagIcon className="size-3.5" />
      </PopoverTrigger>
      <PopoverContent
        align="end"
        className="w-72 p-0"
        onClick={(event) => event.stopPropagation()}
      >
        <Command>
          <CommandInput placeholder="搜索标签" />
          <CommandList>
            {loading ? <CommandEmpty>加载标签中...</CommandEmpty> : null}
            {!loading && flattenedTags.length === 0 ? (
              <CommandEmpty>暂无可用标签</CommandEmpty>
            ) : null}
            {!loading ? (
              <CommandGroup heading="标签">
                {flattenedTags.map((tag) => {
                  const checked = selectedTagIds.has(tag.id)
                  const pending = pendingTagId === tag.id
                  return (
                    <CommandItem
                      key={tag.id}
                      value={`${tag.name} ${tag.remark}`}
                      disabled={pendingTagId !== null}
                      onSelect={() => void handleToggle(tag)}
                    >
                      {pending ? (
                        <Loader2Icon className="mr-2 size-4 animate-spin" />
                      ) : (
                        <CheckIcon
                          className={cn(
                            "mr-2 size-4",
                            checked ? "opacity-100" : "opacity-0"
                          )}
                        />
                      )}
                      <span
                        className="truncate"
                        style={{ paddingLeft: `${tag.depth * 12}px` }}
                      >
                        {tag.name}
                      </span>
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            ) : null}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

type ConversationTagBadgesProps = {
  tags?: AgentConversationTag[]
}

export function ConversationTagBadges({ tags }: ConversationTagBadgesProps) {
  if (!tags || tags.length === 0) {
    return null
  }

  return (
    <div className="mt-1 flex flex-wrap items-center gap-1">
      {tags.map((tag) => (
        <Badge key={tag.id} variant="outline" className="max-w-full truncate px-1.5">
          {tag.name}
        </Badge>
      ))}
    </div>
  )
}
