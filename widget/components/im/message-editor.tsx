"use client";

import { useEffect, useRef, useState } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Image from "@tiptap/extension-image";
import Placeholder from "@tiptap/extension-placeholder";
import { ImageIcon, PaperclipIcon, SendHorizonalIcon } from "lucide-react";

type UploadedImage = {
  url: string;
  filename?: string;
};

type MessageEditorProps = {
  disabled?: boolean;
  uploadingAsset?: boolean;
  onSend: (html: string) => Promise<void>;
  onUploadImage: (file: File) => Promise<UploadedImage | null>;
  onSendAttachment: (file: File) => Promise<void>;
};

export function MessageEditor({
  disabled = false,
  uploadingAsset = false,
  onSend,
  onUploadImage,
  onSendAttachment,
}: MessageEditorProps) {
  const [localUploading, setLocalUploading] = useState(false);
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);
  const onSendRef = useRef(onSend);
  const onUploadImageRef = useRef(onUploadImage);
  const onSendAttachmentRef = useRef(onSendAttachment);
  const shouldRestoreFocusRef = useRef(false);
  const isUploading = uploadingAsset || localUploading;

  useEffect(() => {
    onSendRef.current = onSend;
  }, [onSend]);

  useEffect(() => {
    onUploadImageRef.current = onUploadImage;
  }, [onUploadImage]);

  useEffect(() => {
    onSendAttachmentRef.current = onSendAttachment;
  }, [onSendAttachment]);

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        heading: false,
        blockquote: false,
        codeBlock: false,
        bulletList: false,
        orderedList: false,
        horizontalRule: false,
      }),
      Image,
      Placeholder.configure({
        placeholder: "输入消息，Enter 发送，Shift + Enter 换行",
      }),
    ],
    content: "",
    editorProps: {
      attributes: {
        class:
          "cs-agent-scrollbar min-h-12 max-h-40 overflow-y-auto px-1.5 py-1 text-[13px] leading-6 text-slate-900 outline-none [&_p]:m-0 [&_p+*]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-xl [&_img]:object-contain",
      },
      handleKeyDown: (_view, event) => {
        if (event.key === "Enter" && !event.shiftKey) {
          event.preventDefault();
          void handleSend();
          return true;
        }
        return false;
      },
      handlePaste: (_view, event) => {
        if (disabled || isUploading) {
          return false;
        }
        const imageFile = getClipboardImageFile(event.clipboardData);
        if (!imageFile) {
          return false;
        }
        event.preventDefault();
        void insertUploadedImage(imageFile);
        return true;
      },
    },
  });

  useEffect(() => {
    if (!editor) {
      return;
    }
    editor.setEditable(!disabled && !isUploading);
  }, [disabled, editor, isUploading]);

  useEffect(() => {
    if (!editor || disabled || isUploading || !shouldRestoreFocusRef.current) {
      return;
    }
    requestAnimationFrame(() => {
      editor.commands.focus();
    });
  }, [disabled, editor, isUploading]);

  async function handleSend() {
    if (!editor || disabled || isUploading) {
      return;
    }
    const html = editor.getHTML();
    if (!isMeaningfulHTML(html)) {
      return;
    }
    await onSendRef.current(html);
    editor.commands.clearContent(true);
  }

  async function handleSelectImage(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || !editor || disabled || isUploading) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus();
        });
      }
      return;
    }
    await insertUploadedImage(file);
  }

  async function insertUploadedImage(file: File) {
    if (!editor || disabled || isUploading) {
      return;
    }
    shouldRestoreFocusRef.current = true;
    const objectUrl = URL.createObjectURL(file);
    const placeholderId = `uploading-${crypto.randomUUID()}`;
    editor
      .chain()
      .focus()
      .setImage({
        src: objectUrl,
        alt: file.name || "uploading-image",
        title: placeholderId,
      })
      .run();

    try {
      setLocalUploading(true);
      const uploaded = await onUploadImageRef.current(file);
      if (!uploaded?.url) {
        removeImageByTitle(editor, placeholderId);
        return;
      }
      replaceImageSourceByTitle(
        editor,
        placeholderId,
        uploaded.url,
        uploaded.filename || "image",
      );
    } finally {
      setLocalUploading(false);
      URL.revokeObjectURL(objectUrl);
      requestAnimationFrame(() => {
        if (!disabled && shouldRestoreFocusRef.current) {
          editor.commands.focus();
        }
      });
    }
  }

  async function handleSelectAttachment(
    event: React.ChangeEvent<HTMLInputElement>,
  ) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || disabled || isUploading) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus();
        });
      }
      return;
    }
    shouldRestoreFocusRef.current = editor?.isFocused ?? true;
    setLocalUploading(true);
    try {
      await onSendAttachmentRef.current(file);
    } finally {
      setLocalUploading(false);
      requestAnimationFrame(() => {
        if (editor && !disabled && shouldRestoreFocusRef.current) {
          editor.commands.focus();
        }
      });
    }
  }

  return (
    <div className="px-3 pb-3 pt-2">
      <div className="rounded-3xl border border-white/60 bg-white/78 p-2 shadow-[0_10px_24px_rgba(15,23,42,0.05)] backdrop-blur">
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleSelectImage}
        />
        <input
          ref={attachmentInputRef}
          type="file"
          className="hidden"
          onChange={handleSelectAttachment}
        />
        <div className="min-h-10">
          <EditorContent editor={editor} />
        </div>
        <div className="mt-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true;
                imageInputRef.current?.click();
              }}
              disabled={disabled || isUploading}
              aria-label={isUploading ? "图片上传中" : "发送图片"}
              className="inline-flex size-8 shrink-0 items-center justify-center rounded-xl border border-slate-200/80 bg-white/90 text-slate-500 shadow-[0_8px_18px_rgba(15,23,42,0.05)] transition duration-200 hover:-translate-y-0.5 hover:text-slate-700 disabled:translate-y-0 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-300"
            >
              <ImageIcon className="size-4" />
            </button>
            <button
              type="button"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true;
                attachmentInputRef.current?.click();
              }}
              disabled={disabled || isUploading}
              aria-label={isUploading ? "附件上传中" : "发送附件"}
              className="inline-flex size-8 shrink-0 items-center justify-center rounded-xl border border-slate-200/80 bg-white/90 text-slate-500 shadow-[0_8px_18px_rgba(15,23,42,0.05)] transition duration-200 hover:-translate-y-0.5 hover:text-slate-700 disabled:translate-y-0 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-300"
            >
              <PaperclipIcon className="size-4" />
            </button>
          </div>
          <div className="flex items-center gap-1">
            <p className="text-[10px] text-slate-400">Enter 发送</p>
            <button
              type="button"
              onClick={() => void handleSend()}
              disabled={disabled || isUploading}
              aria-label="发送"
              className="inline-flex size-8 shrink-0 items-center justify-center rounded-xl bg-[linear-gradient(135deg,var(--primary),color-mix(in_srgb,var(--primary)_75%,white_25%))] text-white shadow-[0_10px_20px_color-mix(in_srgb,var(--primary)_28%,transparent)] transition duration-200 hover:-translate-y-0.5 hover:brightness-105 disabled:translate-y-0 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:shadow-none"
            >
              <SendHorizonalIcon className="size-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function isMeaningfulHTML(html: string) {
  const normalized = html
    .replace(/<p><\/p>/g, "")
    .replace(/<p><br><\/p>/g, "")
    .replace(/\s+/g, "");
  if (/<img[\s\S]*?>/i.test(normalized)) {
    return true;
  }
  const plainText = normalized.replace(/<[^>]+>/g, "").trim();
  return plainText !== "";
}

function getClipboardImageFile(clipboardData: DataTransfer | null) {
  if (!clipboardData) {
    return null;
  }
  for (const item of Array.from(clipboardData.items)) {
    if (item.kind === "file" && item.type.startsWith("image/")) {
      return item.getAsFile();
    }
  }
  return null;
}

function removeImageByTitle(editor: NonNullable<ReturnType<typeof useEditor>>, title: string) {
  const { state } = editor;
  let targetPos: number | null = null;
  state.doc.descendants((node, pos) => {
    if (node.type.name === "image" && node.attrs.title === title) {
      targetPos = pos;
      return false;
    }
    return true;
  });
  if (targetPos === null) {
    return;
  }
  editor.chain().focus().deleteRange({ from: targetPos, to: targetPos + 1 }).run();
}

function replaceImageSourceByTitle(
  editor: NonNullable<ReturnType<typeof useEditor>>,
  title: string,
  src: string,
  alt: string,
) {
  const { state, view } = editor;
  let targetPos: number | null = null;
  state.doc.descendants((node, pos) => {
    if (node.type.name === "image" && node.attrs.title === title) {
      targetPos = pos;
      return false;
    }
    return true;
  });
  if (targetPos === null) {
    return;
  }
  const transaction = view.state.tr.setNodeMarkup(targetPos, undefined, {
    ...view.state.doc.nodeAt(targetPos)?.attrs,
    src,
    alt,
    title: "",
  });
  view.dispatch(transaction);
}
