"use client";

import { memo, useEffect, useRef } from "react";

type MessageHTMLProps = {
  html: string;
  className?: string;
  onImageSettled?: () => void;
};

function MessageHTMLComponent({
  html,
  className = "",
  onImageSettled,
}: MessageHTMLProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const onImageSettledRef = useRef(onImageSettled);

  useEffect(() => {
    onImageSettledRef.current = onImageSettled;
  }, [onImageSettled]);

  useEffect(() => {
    const container = containerRef.current;
    const handleSettled = onImageSettledRef.current;
    if (!container || !handleSettled) {
      return;
    }
    const images = Array.from(container.querySelectorAll("img"));
    if (images.length === 0) {
      return;
    }
    const cleanups = images.map((image) => {
      const handler = () => onImageSettledRef.current?.();
      image.addEventListener("load", handler);
      image.addEventListener("error", handler);
      if (image.complete) {
        handleSettled();
      }
      return () => {
        image.removeEventListener("load", handler);
        image.removeEventListener("error", handler);
      };
    });
    return () => {
      cleanups.forEach((cleanup) => cleanup());
    };
  }, [html, onImageSettled]);

  return (
    <div
      ref={containerRef}
      className={`break-words [&_p]:m-0 [&_p+*]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-xl [&_img]:object-contain ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

export const MessageHTML = memo(
  MessageHTMLComponent,
  (prevProps, nextProps) =>
    prevProps.html === nextProps.html &&
    prevProps.className === nextProps.className,
);
