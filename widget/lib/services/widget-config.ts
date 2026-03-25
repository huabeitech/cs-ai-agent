import { requestJson } from "@/lib/services/http";
import type { JsonResult, WidgetConfigResponse } from "@/lib/services/types";
import { readWidgetConfig } from "@/lib/widget/config";

export async function fetchWidgetConfig() {
  const config = readWidgetConfig();
  const result = await requestJson<JsonResult<WidgetConfigResponse>>(
    `/api/open/im/widget/config?appId=${encodeURIComponent(config.appId)}`,
  );
  return result.data ?? {};
}

export { readWidgetConfig };
