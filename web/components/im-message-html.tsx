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
      className={`break-words text-sm [&_p]:m-0 [&_p+*]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-md [&_img]:object-contain ${className}`}
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
