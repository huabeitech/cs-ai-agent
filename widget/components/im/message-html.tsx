"use client";

import { memo, useEffect, useRef } from "react";

type MessageHTMLProps = {
  html: string;
  className?: string;
  onImageSettled?: () => void;
  onImageClick?: (src: string, alt?: string) => void;
};

function MessageHTMLComponent({
  html,
  className = "",
  onImageSettled,
  onImageClick,
}: MessageHTMLProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const onImageSettledRef = useRef(onImageSettled);
  const onImageClickRef = useRef(onImageClick);

  useEffect(() => {
    onImageSettledRef.current = onImageSettled;
  }, [onImageSettled]);

  useEffect(() => {
    onImageClickRef.current = onImageClick;
  }, [onImageClick]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }
    const images = Array.from(container.querySelectorAll("img"));
    if (images.length === 0) {
      return;
    }
    const cleanups = images.map((image) => {
      const handleSettled = () => onImageSettledRef.current?.();
      const handleClick = () => {
        const src = image.getAttribute("src");
        if (src) {
          const alt = image.getAttribute("alt") ?? undefined;
          onImageClickRef.current?.(src, alt);
        }
      };
      image.addEventListener("load", handleSettled);
      image.addEventListener("error", handleSettled);
      image.addEventListener("click", handleClick);
      if (image.complete) {
        onImageSettledRef.current?.();
      }
      image.classList.add("cursor-zoom-in");
      return () => {
        image.removeEventListener("load", handleSettled);
        image.removeEventListener("error", handleSettled);
        image.removeEventListener("click", handleClick);
      };
    });
    return () => {
      cleanups.forEach((cleanup) => cleanup());
    };
  }, [html, onImageSettled, onImageClick]);

  return (
    <div
      ref={containerRef}
      className={`break-words [&_p]:m-0 [&_p+*]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-xl [&_img]:object-contain [&_img]:max-w-full [&_.im-attachment]:min-w-0 [&_.im-attachment-link]:flex [&_.im-attachment-link]:min-w-0 [&_.im-attachment-link]:items-center [&_.im-attachment-link]:gap-3 [&_.im-attachment-link]:rounded-2xl [&_.im-attachment-link]:no-underline [&_.im-attachment-link]:transition-colors hover:[&_.im-attachment-link]:bg-black/5 [&_.im-attachment-icon]:flex [&_.im-attachment-icon]:size-10 [&_.im-attachment-icon]:shrink-0 [&_.im-attachment-icon]:items-center [&_.im-attachment-icon]:justify-center [&_.im-attachment-icon]:rounded-2xl [&_.im-attachment-icon]:bg-black/5 [&_.im-attachment-icon_svg]:size-5 [&_.im-attachment-content]:flex [&_.im-attachment-content]:min-w-0 [&_.im-attachment-content]:flex-col [&_.im-attachment-title]:truncate [&_.im-attachment-title]:font-medium [&_.im-attachment-meta]:text-xs [&_.im-attachment-meta]:opacity-70 ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

export const MessageHTML = memo(
  MessageHTMLComponent,
  (prevProps, nextProps) =>
    prevProps.html === nextProps.html &&
    prevProps.className === nextProps.className &&
    prevProps.onImageSettled === nextProps.onImageSettled &&
    prevProps.onImageClick === nextProps.onImageClick,
);
