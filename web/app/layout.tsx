import type { Metadata } from "next"
import { Geist, Geist_Mono } from "next/font/google"

import { AuthProvider } from "@/components/auth-provider"
import { ImageLightboxProvider } from "@/components/image-lightbox"
import { ThemeProvider } from "@/components/theme-provider"
import { TooltipProvider } from "@/components/ui/tooltip"
import { Toaster } from "@/components/ui/sonner"

import "./globals.css"
import "md-editor-rt/lib/style.css"
import "@/styles/main.scss"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "AI 客服后台管理系统",
  description: "AI 客服后台管理系统",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider>
          <AuthProvider>
            <ImageLightboxProvider>
              <TooltipProvider>
                {children}
                <Toaster position="top-center" richColors />
              </TooltipProvider>
            </ImageLightboxProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
