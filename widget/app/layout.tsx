import type { Metadata } from "next";

import { ImageLightboxProvider } from "@/components/image-lightbox";

import "./globals.css";

export const metadata: Metadata = {
  title: "CS Agent Widget",
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
