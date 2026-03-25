export type EditorMode = "markdown" | "html"

export type EditorValue = {
  mode: EditorMode
  raw: string
}

export type EditorFeatures = {
  image?: boolean
  link?: boolean
  table?: boolean
  codeBlock?: boolean
}

export type BaseEditorProps = {
  placeholder?: string
  disabled?: boolean
  features?: EditorFeatures
  className?: string
}

