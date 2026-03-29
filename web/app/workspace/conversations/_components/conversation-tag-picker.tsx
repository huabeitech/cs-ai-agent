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
import { type TagTree } from "@/lib/api/admin"
import { cn } from "@/lib/utils"

type TagNode = TagTree & {
  depth: number
}

function flattenTagTree(nodes: TagTree[], depth = 0): TagNode[] {
  const result: TagNode[] = []
  nodes.forEach((item) => {
    result.push({ ...item, depth })
    if (item.children.length > 0) {
      result.push(...flattenTagTree(item.children, depth + 1))
    }
  })
  return result
}

type ConversationTagPickerProps = {
  conversation: AgentConversation
  availableTags: TagTree[]
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

  const flattenedTags = useMemo(() => flattenTagTree(availableTags), [availableTags])
  const activeTags = useMemo(
    () => flattenedTags.filter((item) => item.status === 0),
    [flattenedTags]
  )
  const selectedTagIds = useMemo(
    () => new Set((conversation.tags ?? []).map((item) => item.id)),
    [conversation.tags]
  )

  async function handleToggle(tag: TagNode) {
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
            size="sm"
            className="h-7 shrink-0 gap-1 px-2 text-xs"
            aria-label="编辑会话标签"
          />
        }
      >
        <TagIcon className="size-3.5 text-muted-foreground" />
        编辑
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
                      value={`${tag.id} ${tag.name} ${tag.remark}`}
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
    <div className="flex flex-wrap items-center gap-1.5">
      {tags.map((tag) => (
        <Badge
          key={tag.id}
          variant="outline"
          className="max-w-full truncate px-2 text-[11px] font-normal"
        >
          {tag.name}
        </Badge>
      ))}
    </div>
  )
}
