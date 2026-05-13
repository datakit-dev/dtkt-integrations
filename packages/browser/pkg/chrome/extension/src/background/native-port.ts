import type { ChromeAction, ChromeConfig } from "../proto/integration/browser/v1beta/chrome_pb";
import {
  ChromeEventSchema,
  ChromeActionSchema,
  ChromeConfigSchema,
} from "../proto/integration/browser/v1beta/chrome_pb";
import type { ChromeEvent } from "../proto/integration/browser/v1beta/chrome_pb";
import { fromJson, toJson } from "@bufbuild/protobuf";

let nativePort: chrome.runtime.Port | null = null;
let config: ChromeConfig | null = null;
let onConfigCallback: ((config: ChromeConfig) => void) | null = null;
let onActionCallback: ((action: ChromeAction) => void) | null = null;

/**
 * Register the callback that will be invoked whenever the native host sends
 * a ChromeAction. Must be called before any events are emitted.
 */
export function initNativePort(
  onConfig: (context: ChromeConfig) => void,
  onAction: (action: ChromeAction) => void,
): void {
  onConfigCallback = onConfig;
  onActionCallback = onAction;
}

/**
 * Emit a ChromeEvent to the native host over the messaging bridge.
 */
export function handleEvent(event: ChromeEvent): void {
  const port = getPort();
  if (!port) {
    if (chrome.runtime.lastError) {
      console.error(chrome.runtime.lastError.message);
    } else {
      console.warn("Native port unavailable, event dropped:", { event });
    }
    return;
  }
  port.postMessage(toJson(ChromeEventSchema, event));
}

function getPort(): chrome.runtime.Port | null {
  if (!nativePort) {
    nativePort = chrome.runtime.connectNative("com.datakit.chrome_bridge");

    nativePort.onDisconnect.addListener(() => {
      console.warn("Native port disconnected.");
      nativePort = null;
    });

    nativePort.onMessage.addListener((message) => {
      // The first message from the native host should be the config.
      // Start accepting actions after that.
      if (!config) {
        try {
          config = fromJson(ChromeConfigSchema, message);
          if (onConfigCallback) {
            onConfigCallback(config);
          }
        } catch (e) {
          console.error("Failed to parse config from native message:", e);
        }
        return;
      }

      if (onActionCallback) {
        onActionCallback(fromJson(ChromeActionSchema, message));
      }
    });
  }
  return nativePort;
}
