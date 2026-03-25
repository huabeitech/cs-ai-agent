const VISITOR_KEY = "cs-agent:visitor-id";

export function getOrCreateVisitorId() {
  if (typeof window === "undefined") {
    return "";
  }
  const current = window.localStorage.getItem(VISITOR_KEY);
  if (current) {
    return current;
  }
  const visitorId = `visitor_${crypto.randomUUID()}`;
  window.localStorage.setItem(VISITOR_KEY, visitorId);
  return visitorId;
}
