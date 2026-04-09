import type { NextConfig } from "next"
import { PHASE_DEVELOPMENT_SERVER } from "next/constants"

const backendBaseUrl =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || ""

export default function nextConfig(phase: string): NextConfig {
  const config: NextConfig = {
    output: "export",
    trailingSlash: true,
  }

  if (phase !== PHASE_DEVELOPMENT_SERVER) {
    return config
  }

  return {
    ...config,
    async rewrites() {
      return [
        {
          source: "/api/:path*",
          destination: `${backendBaseUrl}/api/:path*`,
        },
        {
          source: "/storage/:path*",
          destination: `${backendBaseUrl}/storage/:path*`,
        },
      ]
    },
  }
}
