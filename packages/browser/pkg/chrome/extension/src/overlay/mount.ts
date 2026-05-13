import { createElement } from 'react';
import { createRoot } from 'react-dom/client';
import type { Task } from '../lib/types';
import type { ExtractionSchema, ExtractionRecord } from '../proto/integration/browser/v1beta/extraction_pb';
import { Overlay } from './Overlay';

/**
 * Mount the extraction overlay into the main document.
 * Returns an unmount function to remove it cleanly.
 */
export function mountOverlay(taskId: string, task: Task, schema: ExtractionSchema, record: ExtractionRecord | null): () => void {
  console.log('[DataKit] mountOverlay called', { taskId, task, schema, record });

  const host = document.createElement('div');
  host.id = 'dtkt-overlay-host';
  // Add inline styles and attributes to prevent page CSS from hiding us
  host.setAttribute('data-dtkt-extension', 'true');
  host.style.cssText = 'all: initial !important; position: static !important; display: block !important; visibility: visible !important; opacity: 1 !important;';
  document.body.appendChild(host);

  console.log('[DataKit] Host element appended to body:', host);

  // Keep host as the last child of body so it paints above same-z-index elements
  // (e.g. page modals that are appended to body after us).
  // Also re-add if removed (LinkedIn sometimes removes injected elements).
  const observer = new MutationObserver(() => {
    if (!document.body.contains(host)) {
      document.body.appendChild(host);
    } else if (document.body.lastChild !== host) {
      document.body.appendChild(host);
    }
  });
  observer.observe(document.body, { childList: true, subtree: false });

  const root = createRoot(host);

  const unmount = () => {
    observer.disconnect();
    root.unmount();
    host.remove();
  };

  root.render(createElement(Overlay, { taskId, task, schema, record, onUnmount: unmount }));

  console.log('[DataKit] Overlay component rendered');

  return unmount;
}
