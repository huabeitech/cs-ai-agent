"use client"

import type { ReactNode } from "react"
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"

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
import { cn } from "@/lib/utils"

export type ComboboxOption = {
  value: string
  label: string
}

type OptionComboboxProps = {
  value: string
  options: ComboboxOption[]
  placeholder: string
  searchPlaceholder?: string
  emptyText?: string
  disabled?: boolean
  onChange: (value: string) => void
  renderOptionAction?: (option: ComboboxOption) => ReactNode
}

export function OptionCombobox({
  value,
  options,
  placeholder,
  searchPlaceholder = "请输入关键字搜索",
  emptyText = "没有可选项",
  disabled = false,
  onChange,
  renderOptionAction,
}: OptionComboboxProps) {
  const selectedLabel =
    options.find((option) => option.value === value)?.label ?? placeholder

  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            className="w-full justify-between font-normal"
            disabled={disabled}
          />
        }
      >
        <span className="truncate">{selectedLabel}</span>
        <ChevronsUpDownIcon className="ml-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-(--radix-popover-trigger-width) p-0" align="start">
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyText}</CommandEmpty>
            <CommandGroup>
              {options.map((option) => (
                <CommandItem
                  key={option.value}
                  value={`${option.label} ${option.value}`}
                  onSelect={() => onChange(option.value)}
                >
                  <div className="flex min-w-0 flex-1 items-center justify-between gap-2">
                    <div className="flex min-w-0 items-center">
                      <CheckIcon
                        className={cn(
                          "mr-2 size-4 shrink-0",
                          option.value === value ? "opacity-100" : "opacity-0"
                        )}
                      />
                      <span className="truncate">{option.label}</span>
                    </div>
                    {renderOptionAction ? (
                      <div
                        className="shrink-0"
                        onMouseDown={(event) => event.preventDefault()}
                        onClick={(event) => event.stopPropagation()}
                      >
                        {renderOptionAction(option)}
                      </div>
                    ) : null}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
