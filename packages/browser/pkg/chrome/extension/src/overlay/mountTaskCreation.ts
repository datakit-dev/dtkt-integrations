import { createElement } from 'react';
import { createRoot } from 'react-dom/client';
import { TaskCreationOverlay } from './TaskCreationOverlay';

/**
 * Mount the task creation overlay into the main document.
 */
export function mountTaskCreationOverlay(title: string, url: string): void {
  const host = document.createElement('div');
  host.id = 'dtkt-task-creation-overlay-host';
  host.setAttribute('data-dtkt-extension', 'true');
  host.style.cssText = 'all: initial !important; position: static !important; display: block !important; visibility: visible !important; opacity: 1 !important;';
  document.body.appendChild(host);

  const root = createRoot(host);

  const unmount = () => {
    root.unmount();
    host.remove();
  };

  root.render(createElement(TaskCreationOverlay, {
    initialTitle: title,
    initialUrl: url,
    onClose: unmount,
  }));
}
