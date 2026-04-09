import { generateUUID } from "@/lib/utils";

const EXTERNAL_ID_KEY = "cs-agent:external-id";

export function getOrCreateExternalId() {
  if (typeof window === "undefined") {
    return "";
  }
  const current = window.localStorage.getItem(EXTERNAL_ID_KEY);
  if (current) {
    return current;
  }
  const visitorId = `visitor_${generateUUID()}`;
  window.localStorage.setItem(EXTERNAL_ID_KEY, visitorId);
  return visitorId;
}
