import type { ContentPush } from "../lib/messages";

/**
 * Handle "Create Page Task" from page context menu.
 * Shows task creation overlay in the tab.
 */
export async function handleCreateTaskFromPage(tab: chrome.tabs.Tab | undefined): Promise<void> {
  if (!tab?.url || !tab?.title || !tab?.id) {
    console.warn("Cannot create task: missing tab URL, title, or ID");
    return;
  }

  // Send message to content script to show task creation overlay
  chrome.tabs.sendMessage(tab.id, {
    type: "SHOW_TASK_CREATION",
    title: tab.title,
    url: tab.url,
  } satisfies ContentPush);
}

/**
 * Handle "Create Link Task" context menu.
 * Shows task creation overlay with the link URL and text.
 */
export async function handleCreateTaskFromLink(
  info: chrome.contextMenus.OnClickData,
  tab: chrome.tabs.Tab | undefined,
): Promise<void> {
  if (!info.linkUrl || !tab?.id) {
    console.warn("Cannot create task: missing link URL or tab ID");
    return;
  }

  // Use selection text as title if available, otherwise extract from link URL
  const title = info.selectionText || new URL(info.linkUrl).hostname;

  // Send message to content script to show task creation overlay
  chrome.tabs.sendMessage(tab.id, {
    type: "SHOW_TASK_CREATION",
    title,
    url: info.linkUrl,
  } satisfies ContentPush);
}
