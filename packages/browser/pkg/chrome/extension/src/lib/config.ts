import { ChromeConfig } from "@/proto/integration/browser/v1beta/chrome_pb";
import { Context } from "@/proto/integration/browser/v1beta/context_pb";

let context: Context | null = null;

export function handleConfig(config: ChromeConfig): void {
  console.log("Received chrome config:", { config });
  if (context === null) {
    context = config.context ?? null;
  }
}

export function getContext(): Context | null {
  return context;
}
