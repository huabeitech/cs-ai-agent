"use client";

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { ExternalLinkIcon, XIcon } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { Button, buttonVariants } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

type ImageLightboxItem = {
  src: string;
  alt?: string;
};

type ImageLightboxContextValue = {
  open: (src: string, alt?: string) => void;
  close: () => void;
};

const ImageLightboxContext = createContext<ImageLightboxContextValue | null>(
  null,
);

export function useImageLightbox(): ImageLightboxContextValue {
  const ctx = useContext(ImageLightboxContext);
  if (!ctx) {
    throw new Error("useImageLightbox 必须在 ImageLightboxProvider 内使用");
  }
  return ctx;
}

/** 未包裹 Provider 时返回 null，便于渐进接入 */
export function useImageLightboxOptional(): ImageLightboxContextValue | null {
  return useContext(ImageLightboxContext);
}

export type ImageLightboxProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  src: string | null;
  alt?: string;
};

function canOpenInNewTab(url: string): boolean {
  if (!url) {
    return false;
  }
  if (url.startsWith("/")) {
    return true;
  }
  try {
    const parsed = new URL(url);
    return (
      parsed.protocol === "http:" ||
      parsed.protocol === "https:" ||
      parsed.protocol === "blob:"
    );
  } catch {
    return false;
  }
}

function LightboxImageBody({
  src,
  alt,
}: {
  src: string;
  alt?: string;
}) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const showOpenTab = canOpenInNewTab(src);

  return (
    <>
      {loading && !error ? (
        <div
          className="pointer-events-none absolute inset-0 flex items-center justify-center"
          aria-hidden
        >
          <div className="size-10 animate-pulse rounded-full bg-white/25" />
        </div>
      ) : null}
      {error ? (
        <div className="flex min-h-[min(50vh,320px)] flex-col items-center justify-center gap-4 px-6 py-12 text-center text-sm text-white/90">
          <p>图片加载失败</p>
          {showOpenTab ? (
            <a
              href={src}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(buttonVariants({ variant: "secondary", size: "sm" }))}
            >
              在新标签页打开
            </a>
          ) : null}
        </div>
      ) : (
        // eslint-disable-next-line @next/next/no-img-element -- 外链与任意尺寸大图预览
        <img
          src={src}
          alt={alt || "预览图片"}
          className={cn(
            "mx-auto block max-h-[min(85vh,calc(100dvh-3rem))] w-auto max-w-full object-contain p-4 sm:p-6",
            loading ? "opacity-0" : "opacity-100",
          )}
          onLoad={() => {
            setLoading(false);
            setError(false);
          }}
          onError={() => {
            setLoading(false);
            setError(true);
          }}
        />
      )}
    </>
  );
}

export function ImageLightboxView({
  open,
  onOpenChange,
  src,
  alt,
}: ImageLightboxProps) {
  const showOpenTab = src ? canOpenInNewTab(src) : false;
  const titleText = alt?.trim() || "图片预览";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {src ? (
        <DialogPortal>
          <DialogOverlay className="z-100 bg-black/85 supports-backdrop-filter:backdrop-blur-xs" />
          <DialogPrimitive.Popup
            data-slot="image-lightbox-popup"
            className={cn(
              "fixed inset-0 z-100 flex max-h-dvh flex-col outline-none",
              "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0 duration-100",
            )}
          >
            <div className="flex h-12 shrink-0 items-center gap-3 border-b border-white/10 bg-black/55 px-3 py-2 text-white sm:px-4">
              <DialogTitle className="min-w-0 flex-1 truncate text-left text-sm font-medium leading-snug text-white">
                {titleText}
              </DialogTitle>
              <div className="flex shrink-0 items-center gap-1">
                {showOpenTab ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className="text-white hover:bg-white/10"
                    aria-label="在新标签页打开"
                    onClick={() => {
                      if (src) {
                        window.open(src, "_blank", "noopener,noreferrer");
                      }
                    }}
                  >
                    <ExternalLinkIcon className="size-4" />
                  </Button>
                ) : null}
                <DialogClose
                  render={
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="text-white hover:bg-white/10"
                      aria-label="关闭"
                    />
                  }
                >
                  <XIcon className="size-4" />
                  <span className="sr-only">关闭</span>
                </DialogClose>
              </div>
            </div>
            <div className="relative min-h-0 flex-1 overflow-auto">
              <LightboxImageBody key={src} src={src} alt={alt} />
            </div>
          </DialogPrimitive.Popup>
        </DialogPortal>
      ) : null}
    </Dialog>
  );
}

export function ImageLightboxProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<ImageLightboxItem | null>(null);

  const open = useCallback((src: string, alt?: string) => {
    const trimmed = src?.trim();
    if (!trimmed) {
      return;
    }
    setState({ src: trimmed, alt });
  }, []);

  const close = useCallback(() => {
    setState(null);
  }, []);

  const contextValue = useMemo(
    () => ({
      open,
      close,
    }),
    [open, close],
  );

  return (
    <ImageLightboxContext.Provider value={contextValue}>
      {children}
      <ImageLightboxView
        open={state !== null}
        onOpenChange={(next) => {
          if (!next) {
            setState(null);
          }
        }}
        src={state?.src ?? null}
        alt={state?.alt}
      />
    </ImageLightboxContext.Provider>
  );
}
