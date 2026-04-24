"use client"

import { memo, useEffect, useRef } from "react"

type ImMessageHTMLProps = {
  html: string
  className?: string
  onImageSettled?: () => void
  onImageClick?: (src: string, alt?: string) => void
}

function ImMessageHTMLComponent({
  html,
  className = "",
  onImageSettled,
  onImageClick,
}: ImMessageHTMLProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const onImageSettledRef = useRef(onImageSettled)
  const onImageClickRef = useRef(onImageClick)

  useEffect(() => {
    onImageSettledRef.current = onImageSettled
  }, [onImageSettled])

  useEffect(() => {
    onImageClickRef.current = onImageClick
  }, [onImageClick])

  useEffect(() => {
    const container = containerRef.current
    if (!container) {
      return
    }
    const images = Array.from(container.querySelectorAll("img"))
    if (images.length === 0) {
      return
    }
    const cleanups = images.map((image) => {
      const handleSettled = () => onImageSettledRef.current?.()
      const handleClick = () => {
        const src = image.getAttribute("src")
        if (src) {
          const alt = image.getAttribute("alt") ?? undefined
          onImageClickRef.current?.(src, alt)
        }
      }

      image.addEventListener("load", handleSettled)
      image.addEventListener("error", handleSettled)
      image.addEventListener("click", handleClick)

      if (image.complete) {
        onImageSettledRef.current?.()
      }

      image.classList.add("cursor-zoom-in")

      return () => {
        image.removeEventListener("load", handleSettled)
        image.removeEventListener("error", handleSettled)
        image.removeEventListener("click", handleClick)
      }
    })
    return () => {
      cleanups.forEach((cleanup) => cleanup())
    }
  }, [html, onImageClick, onImageSettled])

  return (
    <div
      ref={containerRef}
      className={`break-words text-sm [&_p]:m-0 [&_p+*]:mt-2 [&_h1]:m-0 [&_h1]:text-base [&_h1]:font-semibold [&_h1+*]:mt-2 [&_h2]:m-0 [&_h2]:text-[15px] [&_h2]:font-semibold [&_h2+*]:mt-2 [&_h3]:m-0 [&_h3]:font-semibold [&_h3+*]:mt-2 [&_h4]:m-0 [&_h4]:font-medium [&_h4+*]:mt-2 [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_li]:my-1 [&_blockquote]:my-2 [&_blockquote]:border-l-2 [&_blockquote]:border-current/20 [&_blockquote]:pl-3 [&_blockquote]:opacity-90 [&_pre]:my-2 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:bg-black/6 [&_pre]:px-3 [&_pre]:py-2 [&_pre]:text-[13px] [&_pre]:leading-6 [&_code]:rounded [&_code]:bg-black/6 [&_code]:px-1 [&_code]:py-0.5 [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_hr]:my-3 [&_hr]:border-current/10 [&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_th]:border [&_th]:border-current/10 [&_th]:px-2 [&_th]:py-1 [&_th]:text-left [&_th]:font-medium [&_td]:border [&_td]:border-current/10 [&_td]:px-2 [&_td]:py-1 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-md [&_img]:object-contain [&_.im-attachment]:min-w-0 [&_.im-attachment-link]:flex [&_.im-attachment-link]:min-w-0 [&_.im-attachment-link]:items-center [&_.im-attachment-link]:gap-3 [&_.im-attachment-link]:rounded-xl [&_.im-attachment-link]:no-underline [&_.im-attachment-link]:transition-colors hover:[&_.im-attachment-link]:bg-black/5 [&_.im-attachment-icon]:flex [&_.im-attachment-icon]:size-10 [&_.im-attachment-icon]:shrink-0 [&_.im-attachment-icon]:items-center [&_.im-attachment-icon]:justify-center [&_.im-attachment-icon]:rounded-xl [&_.im-attachment-icon]:bg-black/5 [&_.im-attachment-icon_svg]:size-5 [&_.im-attachment-content]:flex [&_.im-attachment-content]:min-w-0 [&_.im-attachment-content]:flex-col [&_.im-attachment-title]:truncate [&_.im-attachment-title]:font-medium [&_.im-attachment-meta]:text-xs [&_.im-attachment-meta]:opacity-70 ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

export const ImMessageHTML = memo(
  ImMessageHTMLComponent,
  (prevProps, nextProps) =>
    prevProps.html === nextProps.html &&
    prevProps.className === nextProps.className &&
    prevProps.onImageSettled === nextProps.onImageSettled &&
    prevProps.onImageClick === nextProps.onImageClick
)
