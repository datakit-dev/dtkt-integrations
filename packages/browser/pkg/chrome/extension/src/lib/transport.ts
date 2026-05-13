/**
 * Connect transport singletons.
 *
 * Two separate cached instances are kept because the background service worker
 * and the content-script IIFE are different bundles with different lifecycles:
 *
 * - Background: `getBackgroundTransport(addr)` — pass the address each time;
 *   the transport is created once and reused for the life of the SW.
 * - Content script: call `initContentTransport(addr)` once at startup, then
 *   `getContentTransport()` anywhere in the IIFE bundle (e.g. Overlay.tsx).
 */
import { Context } from "@/proto/integration/browser/v1beta/context_pb";
import { type Transport } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

const transports = new Map<string, Transport>();

export function getTransport(context: Context): Transport {
  let transport = transports.get(context.address);
  if (!transport) {
    transport = createConnectTransport({
      baseUrl: context.address,
      interceptors: [
        (next) => async (req) => {
          req.header.set("dtkt-addr-name", context.connection);
          return next(req);
        },
      ],
    });
    transports.set(context.address, transport);
  }
  return transport;
}
