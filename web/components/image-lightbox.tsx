"use client";

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import {
  ExternalLinkIcon,
  RefreshCwIcon,
  RotateCcwIcon,
  RotateCwIcon,
  XIcon,
  ZoomInIcon,
  ZoomOutIcon,
} from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import type { ReactZoomPanPinchContentRef } from "react-zoom-pan-pinch";
import {
  TransformComponent,
  TransformWrapper,
} from "react-zoom-pan-pinch";

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
  pinchRef,
  rotationDeg,
}: {
  src: string;
  alt?: string;
  pinchRef: React.RefObject<ReactZoomPanPinchContentRef | null>;
  rotationDeg: number;
}) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const showOpenTab = canOpenInNewTab(src);

  useEffect(() => {
    requestAnimationFrame(() => {
      pinchRef.current?.centerView(1, 0);
    });
  }, [rotationDeg, pinchRef]);

  return (
    <div className="relative h-full min-h-0 w-full min-w-0 flex-1">
      {loading && !error ? (
        <div
          className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center"
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
        <TransformWrapper
          ref={pinchRef}
          initialScale={1}
          minScale={0.35}
          maxScale={8}
          centerOnInit
          centerZoomedOut
          limitToBounds
          wheel={{ step: 0.12 }}
          pinch={{ step: 5 }}
          panning={{ velocityDisabled: false }}
          doubleClick={{ mode: "reset", step: 0.7 }}
        >
          <TransformComponent
            wrapperClass="!h-full !w-full !max-h-full !max-w-full"
            contentClass="!flex !h-full !min-h-0 !w-full !min-w-0 !items-center !justify-center !p-4 sm:!p-6"
          >
            {/* eslint-disable-next-line @next/next/no-img-element -- 外链与任意尺寸大图预览 */}
            <img
              src={src}
              alt={alt || "预览图片"}
              draggable={false}
              style={{ transform: `rotate(${rotationDeg}deg)` }}
              className={cn(
                "max-h-[min(85vh,calc(100dvh-3rem))] max-w-full origin-center object-contain transition-transform duration-200 ease-out select-none",
                loading ? "opacity-0" : "opacity-100",
              )}
              onLoad={() => {
                setLoading(false);
                setError(false);
                requestAnimationFrame(() => {
                  pinchRef.current?.centerView(1, 0);
                });
              }}
              onError={() => {
                setLoading(false);
                setError(true);
              }}
            />
          </TransformComponent>
        </TransformWrapper>
      )}
    </div>
  );
}

/** 按 src 作为 key 挂载，切换图片时旋转角自动回到 0 */
function ImageLightboxDialogContent({
  src,
  alt,
}: {
  src: string;
  alt?: string;
}) {
  const pinchRef = useRef<ReactZoomPanPinchContentRef | null>(null);
  const [rotationDeg, setRotationDeg] = useState(0);
  const showOpenTab = canOpenInNewTab(src);
  const titleText = alt?.trim() || "图片预览";

  const rotateLeft = useCallback(() => {
    setRotationDeg((d) => (d - 90 + 360) % 360);
  }, []);

  const rotateRight = useCallback(() => {
    setRotationDeg((d) => (d + 90) % 360);
  }, []);

  return (
    <DialogPortal>
      <DialogOverlay className="z-100 bg-black/85 supports-backdrop-filter:backdrop-blur-xs" />
      <DialogPrimitive.Popup
        data-slot="image-lightbox-popup"
        className={cn(
          "fixed inset-0 z-100 flex max-h-dvh min-h-0 flex-col outline-none",
          "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0 duration-100",
        )}
      >
        <div className="flex h-12 shrink-0 items-center gap-2 border-b border-white/10 bg-black/55 px-2 py-2 text-white sm:gap-3 sm:px-4">
          <DialogTitle className="min-w-0 flex-1 truncate text-left text-sm font-medium leading-snug text-white">
            {titleText}
          </DialogTitle>
          <div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="text-white hover:bg-white/10"
              aria-label="放大"
              onClick={() => pinchRef.current?.zoomIn(0.25)}
            >
              <ZoomInIcon className="size-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="text-white hover:bg-white/10"
              aria-label="缩小"
              onClick={() => pinchRef.current?.zoomOut(0.25)}
            >
              <ZoomOutIcon className="size-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="text-white hover:bg-white/10"
              aria-label="向左旋转"
              onClick={rotateLeft}
            >
              <RotateCcwIcon className="size-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="text-white hover:bg-white/10"
              aria-label="向右旋转"
              onClick={rotateRight}
            >
              <RotateCwIcon className="size-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="text-white hover:bg-white/10"
              aria-label="重置缩放、位置与旋转"
              onClick={() => {
                setRotationDeg(0);
                pinchRef.current?.resetTransform(200);
              }}
            >
              <RefreshCwIcon className="size-4" />
            </Button>
            {showOpenTab ? (
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="text-white hover:bg-white/10"
                aria-label="在新标签页打开"
                onClick={() => {
                  window.open(src, "_blank", "noopener,noreferrer");
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
        <div className="relative flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          <LightboxImageBody
            pinchRef={pinchRef}
            rotationDeg={rotationDeg}
            src={src}
            alt={alt}
          />
        </div>
        <p className="sr-only">
          使用滚轮或双指缩放，按住拖拽可平移图片；工具栏可向左或向右旋转。
        </p>
      </DialogPrimitive.Popup>
    </DialogPortal>
  );
}

export function ImageLightboxView({
  open,
  onOpenChange,
  src,
  alt,
}: ImageLightboxProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {src ? (
        <ImageLightboxDialogContent key={src} src={src} alt={alt} />
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
