const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"

export function createWebSocketBaseUrl() {
  return API_BASE_URL.replace(/^http/, "ws").replace(/\/$/, "")
}
