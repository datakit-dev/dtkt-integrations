/**
 * DOM field extractor — injected into page context via
 * chrome.scripting.executeScript. Must be a named module-level function with
 * no closures over background scope.
 */

export type PlainField = {
  strategyType: string;   // "cssSelector" | "xpath" | "metaName" | "jsonLd"
  strategyValue: string;  // the selector / expression / meta name / JSON-LD @type
  jsonLdKey: string;      // property key when strategyType === "jsonLd"
  extract: string;        // "text" | "html" | attribute name
  all: boolean;           // collect all matches vs first only
};

export function extractPageFields(
  fields: Record<string, PlainField>,
): Record<string, string | string[]> {
  function val(el: Element, extract: string): string {
    if (!extract || extract === "text") return (el as HTMLElement).innerText ?? el.textContent ?? "";
    if (extract === "html") return el.innerHTML;
    return el.getAttribute(extract) ?? "";
  }

  const result: Record<string, string | string[]> = {};

  for (const [name, f] of Object.entries(fields)) {
    try {
      if (f.strategyType === "cssSelector") {
        if (f.all) {
          result[name] = Array.from(document.querySelectorAll(f.strategyValue)).map(el => val(el, f.extract));
        } else {
          const el = document.querySelector(f.strategyValue);
          result[name] = el ? val(el, f.extract) : "";
        }
      } else if (f.strategyType === "xpath") {
        if (f.all) {
          const xr = document.evaluate(f.strategyValue, document, null, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, null);
          const items: string[] = [];
          for (let i = 0; i < xr.snapshotLength; i++) {
            const node = xr.snapshotItem(i);
            items.push(node instanceof Element ? val(node, f.extract) : (node?.textContent ?? ""));
          }
          result[name] = items;
        } else {
          const xr = document.evaluate(f.strategyValue, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);
          const node = xr.singleNodeValue;
          result[name] = node instanceof Element ? val(node, f.extract) : (node?.textContent ?? "");
        }
      } else if (f.strategyType === "metaName") {
        const el = document.querySelector(`meta[name="${CSS.escape(f.strategyValue)}"]`);
        result[name] = el?.getAttribute("content") ?? "";
      } else if (f.strategyType === "jsonLd") {
        result[name] = "";
        for (const script of Array.from(document.querySelectorAll('script[type="application/ld+json"]'))) {
          try {
            const data = JSON.parse(script.textContent ?? "{}");
            const items = Array.isArray(data) ? data : [data];
            for (const item of items) {
              if (item["@type"] === f.strategyValue) {
                result[name] = f.jsonLdKey ? String(item[f.jsonLdKey] ?? "") : "";
                break;
              }
            }
          } catch { /* ignore malformed JSON-LD */ }
        }
      }
    } catch {
      result[name] = "";
    }
  }

  return result;
}
