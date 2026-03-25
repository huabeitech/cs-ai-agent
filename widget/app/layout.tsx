import type { Metadata } from "next";
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
      <body>{children}</body>
    </html>
  );
}
