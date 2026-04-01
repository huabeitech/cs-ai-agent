import type { ComponentType } from "react"

export type ContentMode = "markdown" | "html"

export type ContentValue = {
  mode: ContentMode
  raw: string
}

export type UploadImageResult = {
  url: string
  alt?: string
  title?: string
}

export type UploadImageHandler = (file: File) => Promise<UploadImageResult | null>

export type EditorToolbarButtonAction = {
  key: string
  label: string
  icon: ComponentType<{ className?: string }>
  onClick: () => void
  disabled?: boolean
  pressed?: boolean
}

export type EditorToolbarSeparatorAction = {
  key: string
  type: "separator"
}

export type EditorToolbarAction =
  | EditorToolbarButtonAction
  | EditorToolbarSeparatorAction
