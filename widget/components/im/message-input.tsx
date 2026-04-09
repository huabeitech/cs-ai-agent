"use client";

import { useLayoutEffect, useRef, useState } from "react";
import { ImageIcon, SendHorizonalIcon } from "lucide-react";

type MessageInputProps = {
  disabled?: boolean;
  uploadingImage?: boolean;
  onSend: (content: string) => Promise<void>;
  onSendImage: (file: File) => Promise<void>;
};

export function MessageInput({
  disabled,
  uploadingImage = false,
  onSend,
  onSendImage,
}: MessageInputProps) {
  const [value, setValue] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const imageInputRef = useRef<HTMLInputElement | null>(null);

  useLayoutEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) {
      return;
    }
    const lineHeight = 20;
    const maxHeight = lineHeight * 8;
    textarea.style.height = "0px";
    textarea.style.height = `${Math.min(textarea.scrollHeight, maxHeight)}px`;
    textarea.style.overflowY =
      textarea.scrollHeight > maxHeight ? "auto" : "hidden";
  }, [value]);

  async function handleSubmit() {
    const content = value.trim();
    if (!content || disabled || submitting) {
      return;
    }
    setSubmitting(true);
    try {
      await onSend(content);
      setValue("");
    } finally {
      setSubmitting(false);
      requestAnimationFrame(() => {
        textareaRef.current?.focus();
      });
    }
  }

  async function handleSelectImage(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || disabled || uploadingImage || submitting) {
      return;
    }
    if (!file.type.startsWith("image/")) {
      return;
    }
    try {
      await onSendImage(file);
    } finally {
      requestAnimationFrame(() => {
        textareaRef.current?.focus();
      });
    }
  }

  return (
    <div className="px-3 pb-3 pt-2">
      <div className="rounded-3xl border border-white/60 bg-white/78 p-2.5 shadow-[0_10px_24px_rgba(15,23,42,0.05)] backdrop-blur">
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleSelectImage}
        />
        <div className="flex items-start gap-3">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(event) => setValue(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter" && !event.shiftKey) {
                event.preventDefault();
                void handleSubmit();
              }
            }}
            placeholder="输入消息，Enter 发送，Shift + Enter 换行"
            disabled={disabled || uploadingImage}
            rows={2}
            className="cs-agent-scrollbar min-h-12 flex-1 resize-none bg-transparent px-1.5 pt-1 text-[13px] leading-6 text-slate-900 outline-none placeholder:text-slate-400 disabled:cursor-not-allowed"
          />
          <button
            type="button"
            onClick={() => imageInputRef.current?.click()}
            disabled={disabled || uploadingImage || submitting}
            aria-label={uploadingImage ? "图片上传中" : "发送图片"}
            className="mt-1 inline-flex size-11 shrink-0 items-center justify-center rounded-2xl border border-slate-200/80 bg-white/90 text-slate-500 shadow-[0_10px_24px_rgba(15,23,42,0.05)] transition duration-200 hover:-translate-y-0.5 hover:text-slate-700 disabled:translate-y-0 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-300"
          >
            <ImageIcon className="size-4" />
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            disabled={disabled || submitting || uploadingImage}
            aria-label={submitting ? "发送中" : "发送"}
            className="mt-1 inline-flex size-11 shrink-0 items-center justify-center rounded-2xl bg-[linear-gradient(135deg,var(--primary),color-mix(in_srgb,var(--primary)_75%,white_25%))] text-white shadow-[0_12px_26px_color-mix(in_srgb,var(--primary)_28%,transparent)] transition duration-200 hover:-translate-y-0.5 hover:brightness-105 disabled:translate-y-0 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:shadow-none"
          >
            <SendHorizonalIcon className="size-4" />
          </button>
        </div>
        <div className="mt-2 px-1.5 text-[11px] text-slate-400">
          Enter 发送，支持图片上传
        </div>
      </div>
    </div>
  );
}
