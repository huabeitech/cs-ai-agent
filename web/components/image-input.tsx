"use client"

import { useRef, useState } from "react"
import { UploadIcon, XIcon } from "lucide-react"
import { toast } from "sonner"

import { uploadAsset } from "@/lib/api/admin"
import { cn } from "@/lib/utils"

export type ImageInputProps = {
  value?: string
  onChange?: (value: string) => void
  disabled?: boolean
  accept?: string
  maxSize?: number
  prefix?: string
  placeholder?: string
  className?: string
}

export function ImageInput({
  value,
  onChange,
  disabled,
  accept = "image/*",
  maxSize = 5 * 1024 * 1024,
  prefix,
  placeholder = "点击上传图片",
  className,
}: ImageInputProps) {
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  function handleClick() {
    if (disabled || uploading) {
      return
    }
    fileInputRef.current?.click()
  }

  function handleClear(event: React.MouseEvent) {
    event.stopPropagation()
    onChange?.("")
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }

    if (!file.type.startsWith("image/")) {
      toast.error("请选择图片文件")
      return
    }

    if (file.size > maxSize) {
      const maxSizeMB = (maxSize / 1024 / 1024).toFixed(0)
      toast.error(`图片大小不能超过 ${maxSizeMB}MB`)
      return
    }

    setUploading(true)
    try {
      const result = await uploadAsset(file, prefix)
      onChange?.(result.url)
      toast.success("图片上传成功")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "上传图片失败")
    } finally {
      setUploading(false)
      if (fileInputRef.current) {
        fileInputRef.current.value = ""
      }
    }
  }

  const isDisabled = disabled || uploading

  return (
    <div className={cn("relative", className)}>
      <input
        ref={fileInputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={handleFileChange}
        disabled={isDisabled}
      />
      <div
        onClick={handleClick}
        className={cn(
          "group relative flex size-24 cursor-pointer items-center justify-center overflow-hidden rounded-lg border-2 border-dashed border-input bg-muted transition-colors",
          "hover:border-primary hover:bg-muted/50",
          "focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none",
          isDisabled && "cursor-not-allowed opacity-50"
        )}
        tabIndex={isDisabled ? -1 : 0}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            handleClick()
          }
        }}
        role="button"
        aria-label={value ? "更换图片" : placeholder}
      >
        {value ? (
          <>
            <img src={value} alt="已上传图片" className="size-full object-cover" />
            <div className="absolute inset-0 flex items-center justify-center bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
              <span className="text-sm text-white">更换图片</span>
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center gap-1 text-muted-foreground">
            <UploadIcon className="size-6" />
            <span className="text-xs">{uploading ? "上传中..." : placeholder}</span>
          </div>
        )}
        {uploading && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/50">
            <div className="size-6 animate-spin rounded-full border-2 border-white border-t-transparent" />
          </div>
        )}
      </div>
      {value && !isDisabled && (
        <button
          type="button"
          onClick={handleClear}
          className="absolute -right-2 -top-2 flex size-5 items-center justify-center rounded-full bg-destructive text-destructive-foreground shadow-sm transition-colors hover:bg-destructive/80"
          aria-label="删除图片"
        >
          <XIcon className="size-3" />
        </button>
      )}
    </div>
  )
}
