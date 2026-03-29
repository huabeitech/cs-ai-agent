import type { NextConfig } from "next"
import { PHASE_DEVELOPMENT_SERVER } from "next/constants"

const backendBaseUrl =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"

export default function nextConfig(phase: string): NextConfig {
  if (phase === PHASE_DEVELOPMENT_SERVER) {
    return {
      reactStrictMode: true,
      trailingSlash: true,
      images: {
        unoptimized: true,
      },
      async rewrites() {
        return [
          {
            source: "/api/:path*",
            destination: `${backendBaseUrl}/api/:path*`,
            basePath: false,
          },
          {
            source: "/storage/:path*",
            destination: `${backendBaseUrl}/storage/:path*`,
            basePath: false,
          },
        ]
      },
    }
  }

  return {
    reactStrictMode: true,
    output: "export",
    basePath: "/widget",
    assetPrefix: "/widget/",
    trailingSlash: true,
    images: {
      unoptimized: true,
    },
  }
}
