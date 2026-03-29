import type { Metadata } from "next";

import { ImageLightboxProvider } from "@/components/image-lightbox";

import "./globals.css";

export const metadata: Metadata = {
  title: "贝壳AI客服插件",
  description: "Embedded customer service widget",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>
        <ImageLightboxProvider>{children}</ImageLightboxProvider>
      </body>
    </html>
  );
}
