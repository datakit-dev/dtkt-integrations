import type { ChromeAction } from "../proto/integration/browser/v1beta/chrome_pb";

export async function handleAction(action: ChromeAction): Promise<void> {
  switch (action.type.case) {
    case "tabs": {
      const tabs = action.type.value;
      switch (tabs.type.case) {
        case "create":
          chrome.tabs.create({
            url: tabs.type.value.url,
            active: tabs.type.value.active ?? undefined,
            pinned: tabs.type.value.pinned ?? undefined,
            index: tabs.type.value.index ?? undefined,
            windowId: tabs.type.value.windowId ?? undefined,
          });
          break;
        case "update":
          chrome.tabs.update(tabs.type.value.tabId, {
            url: tabs.type.value.url ?? undefined,
            active: tabs.type.value.active ?? undefined,
            pinned: tabs.type.value.pinned ?? undefined,
            highlighted: tabs.type.value.highlighted ?? undefined,
            muted: tabs.type.value.muted ?? undefined,
          });
          break;
        case "remove":
          chrome.tabs.remove(tabs.type.value.tabIds);
          break;
        case "reload":
          chrome.tabs.reload(
            tabs.type.value.tabId ?? (undefined as unknown as number),
            { bypassCache: tabs.type.value.bypassCache },
          );
          break;
        default:
          console.warn("Unknown tabs action:", tabs.type);
      }
      break;
    }

    case "notifications": {
      const notifs = action.type.value;
      switch (notifs.type.case) {
        case "create": {
          const v = notifs.type.value;
          chrome.notifications.create(v.notificationId ?? "", {
            type: "basic",
            title: v.title,
            message: v.message,
            iconUrl: v.iconUrl || chrome.runtime.getURL("assets/icon-color-on-blue.png"),
            priority: v.priority as chrome.notifications.NotificationOptions["priority"],
            buttons: v.buttons.map(b => ({ title: b.title, iconUrl: b.iconUrl })),
          });
          break;
        }
        case "update": {
          const v = notifs.type.value;
          chrome.notifications.update(v.notificationId, {
            title: v.title,
            message: v.message,
            iconUrl: v.iconUrl,
            priority: v.priority as chrome.notifications.NotificationOptions["priority"],
          });
          break;
        }
        case "clear":
          chrome.notifications.clear(notifs.type.value.notificationId);
          break;
        default:
          console.warn("Unknown notifications action:", notifs.type);
      }
      break;
    }

    case "scripting": {
      const scripting = action.type.value;
      switch (scripting.type.case) {
        case "executeScript": {
          const { target, source } = scripting.type.value;
          const t = target!;
          const injectionTarget = t.allFrames
            ? { tabId: t.tabId, allFrames: true as const }
            : { tabId: t.tabId, frameIds: t.frameIds };
          if (source.case === "file") {
            chrome.scripting.executeScript({ target: injectionTarget, files: [source.value] });
          } else if (source.case === "code") {
            // eslint-disable-next-line no-new-func
            chrome.scripting.executeScript({ target: injectionTarget, func: new Function(source.value) as () => void, args: [] });
          }
          break;
        }
        default:
          console.warn("Unknown scripting action:", scripting.type);
      }
      break;
    }

    case "contextMenus": {
      const menus = action.type.value;
      switch (menus.type.case) {
        case "create": {
          const v = menus.type.value;
          chrome.contextMenus.create({
            id: v.id,
            title: v.title,
            contexts: v.contexts as [chrome.contextMenus.ContextType, ...chrome.contextMenus.ContextType[]],
            parentId: v.parentId ?? undefined,
          });
          break;
        }
        case "remove":
          chrome.contextMenus.remove(menus.type.value.menuItemId);
          break;
        case "removeAll":
          chrome.contextMenus.removeAll();
          break;
        default:
          console.warn("Unknown contextMenus action:", menus.type);
      }
      break;
    }

    default:
      console.warn("Unknown action received from native:", action.type);
  }
}
