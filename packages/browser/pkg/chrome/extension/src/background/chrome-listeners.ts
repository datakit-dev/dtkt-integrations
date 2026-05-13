import {
  ChromeEventSchema,
  TabsEventSchema,
  RuntimeEventSchema,
  ContextMenusEventSchema,
  NotificationsEventSchema,
} from "../proto/integration/browser/v1beta/chrome_pb";
import { create } from "@bufbuild/protobuf";
import { handleEvent } from "./native-port";
import { handleCreateTaskFromPage, handleCreateTaskFromLink } from "./task-creation";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function mapTab(tab: chrome.tabs.Tab) {
  return {
    id: tab.id,
    windowId: tab.windowId,
    index: tab.index,
    url: tab.url,
    title: tab.title,
    status: tab.status ?? "",
    active: tab.active,
    pinned: tab.pinned,
    audible: tab.audible ?? false,
    muted: tab.mutedInfo?.muted ?? false,
    discarded: tab.discarded ?? false,
    incognito: tab.incognito,
    faviconUrl: tab.favIconUrl,
    groupId: tab.groupId ?? 0,
  };
}

// ─── Runtime ──────────────────────────────────────────────────────────────────

chrome.runtime.onInstalled.addListener(({ reason, previousVersion }) => {
  if (reason === "install") {
    chrome.tabs.create({ url: chrome.runtime.getURL("onboarding/index.html") });
  }

  // Create parent context menu item with DataKit branding
  chrome.contextMenus.create({
    id: "datakit-tasks",
    title: "DataKit Browser",
    contexts: ["page", "link"],
  });

  // Create context menu items for task creation
  chrome.contextMenus.create({
    id: "datakit-create-task-page",
    title: "Create Page Task",
    contexts: ["page"],
    parentId: "datakit-tasks",
  });

  chrome.contextMenus.create({
    id: "datakit-create-task-link",
    title: "Create Link Task",
    contexts: ["link"],
    parentId: "datakit-tasks",
  });

  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "runtime",
      value: create(RuntimeEventSchema, {
        type: { case: "onInstalled", value: { reason, previousVersion } },
      }),
    },
  }));
});

chrome.runtime.onStartup.addListener(() => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "runtime",
      value: create(RuntimeEventSchema, {
        type: { case: "onStartup", value: {} },
      }),
    },
  }));
});

// ─── Tabs ─────────────────────────────────────────────────────────────────────

chrome.tabs.onUpdated.addListener((tabId, info, tab) => {
  if (info.status === "complete" && tab.url) {
    handleEvent(create(ChromeEventSchema, {
      type: {
        case: "tabs",
        value: create(TabsEventSchema, {
          type: {
            case: "onUpdated",
            value: {
              tabId,
              url: tab.url,
              title: tab.title,
              status: info.status,
              pinned: info.pinned,
              audible: info.audible,
              muted: info.mutedInfo?.muted,
              faviconUrl: info.favIconUrl,
              tab: mapTab(tab),
            },
          },
        }),
      },
    }));
  }
});

chrome.tabs.onCreated.addListener((tab) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "tabs",
      value: create(TabsEventSchema, {
        type: { case: "onCreated", value: { tab: mapTab(tab) } },
      }),
    },
  }));
});

chrome.tabs.onActivated.addListener(({ tabId, windowId }) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "tabs",
      value: create(TabsEventSchema, {
        type: { case: "onActivated", value: { tabId, windowId } },
      }),
    },
  }));
});

chrome.tabs.onRemoved.addListener((tabId, { windowId, isWindowClosing }) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "tabs",
      value: create(TabsEventSchema, {
        type: {
          case: "onRemoved",
          value: { tabId, windowId, isWindowClosing },
        },
      }),
    },
  }));
});

// ─── Context Menus ────────────────────────────────────────────────────────────

chrome.contextMenus.onClicked.addListener((info, tab) => {
  // Handle DataKit task creation menu items
  if (info.menuItemId === "datakit-create-task-page") {
    handleCreateTaskFromPage(tab);
    return;
  }

  if (info.menuItemId === "datakit-create-task-link") {
    handleCreateTaskFromLink(info, tab);
    return;
  }

  // Generic fallthrough for dynamically-created context menu items.
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "contextMenus",
      value: create(ContextMenusEventSchema, {
        type: {
          case: "onClicked",
          value: {
            menuItemId: String(info.menuItemId),
            pageUrl: info.pageUrl,
            selectionText: info.selectionText,
            linkUrl: info.linkUrl,
            srcUrl: info.srcUrl,
            frameUrl: info.frameUrl,
            editable: info.editable,
            tabId: tab?.id,
          },
        },
      }),
    },
  }));
});

// ─── Notifications ────────────────────────────────────────────────────────────

chrome.notifications.onButtonClicked.addListener((notificationId, buttonIndex) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "notifications",
      value: create(NotificationsEventSchema, {
        type: { case: "onButtonClicked", value: { notificationId, buttonIndex } },
      }),
    },
  }));
});

chrome.notifications.onClicked.addListener((notificationId) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "notifications",
      value: create(NotificationsEventSchema, {
        type: { case: "onClicked", value: { notificationId } },
      }),
    },
  }));
});

chrome.notifications.onClosed.addListener((notificationId, byUser) => {
  handleEvent(create(ChromeEventSchema, {
    type: {
      case: "notifications",
      value: create(NotificationsEventSchema, {
        type: { case: "onClosed", value: { notificationId, byUser } },
      }),
    },
  }));
});
