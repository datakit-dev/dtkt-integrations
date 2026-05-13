import { useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import type { Task } from '../lib/types';
import type { ExtractionSchema, ExtractionRecord } from '../proto/integration/browser/v1beta/extraction_pb';
import type { ContentRequest } from '../lib/messages';
import { Value, ValueSchema } from '@/proto/google/protobuf/struct_pb';
import { create, toJsonString } from "@bufbuild/protobuf";

// ── Element selector utilities ────────────────────────────────────────────────

function getCssSelector(el: Element): string {
  const parts: string[] = [];
  let cur: Element | null = el;
  while (cur && cur !== document.body) {
    const tag = cur.tagName.toLowerCase();
    if (cur.id) {
      parts.unshift(`#${CSS.escape(cur.id)}`);
      break;
    }
    const ariaLabel = cur.getAttribute('aria-label');
    if (ariaLabel) { parts.unshift(`${tag}[aria-label=${JSON.stringify(ariaLabel)}]`); break; }
    const dataTestId = cur.getAttribute('data-testid') ?? cur.getAttribute('data-test-id');
    if (dataTestId) { parts.unshift(`${tag}[data-testid=${JSON.stringify(dataTestId)}]`); break; }
    const stableClasses = Array.from(cur.classList)
      .filter((c) => /^[a-zA-Z_-][a-zA-Z0-9_-]*$/.test(c) && !/^[_\d]/.test(c))
      .slice(0, 2);
    let seg = tag + stableClasses.map((c) => `.${c}`).join('');
    const parent = cur.parentElement;
    if (parent) {
      const matches = Array.from(parent.querySelectorAll(`:scope > ${seg}`));
      if (matches.length > 1) {
        const idx = Array.from(parent.children).indexOf(cur as HTMLElement) + 1;
        seg += `:nth-child(${idx})`;
      }
    }
    parts.unshift(seg);
    cur = cur.parentElement;
  }
  return parts.join(' > ') || el.tagName.toLowerCase();
}

function getXPath(el: Element): string {
  const parts: string[] = [];
  let cur: Element | null = el;
  while (cur && cur.parentElement !== null) {
    const parent: Element | null = cur.parentElement;
    if (!parent) break;
    const tag = cur.tagName.toLowerCase();
    const sameTag = Array.from(parent.children).filter((c) => c.tagName === cur!.tagName);
    const idx = sameTag.indexOf(cur) + 1;
    parts.unshift(sameTag.length > 1 ? `${tag}[${idx}]` : tag);
    cur = parent;
  }
  return '/' + parts.join('/');
}

interface Props {
  taskId: string;
  task: Task;
  schema: ExtractionSchema;
  record: ExtractionRecord | null;
  onUnmount: () => void;
}

type PendingConfirm = {
  fieldName: string;
  /** Editable draft value shown in the confirm dialog. */
  draft: string;
  /** Selector metadata for the source element, captured at click-time. */
  capture: { cssSelector: string; xpath: string };
};

export function Overlay({ taskId, task, schema, record, onUnmount }: Props) {
  const fields = schema.fields;
  const [values, setValues] = useState<Record<string, Value>>(
    { ...(record?.values ?? {}) }
  );
  const [pickingField, setPickingField] = useState<string | null>(null);
  const [highlightRect, setHighlightRect] = useState<DOMRect | null>(null);
  const [pendingConfirm, setPendingConfirm] = useState<PendingConfirm | null>(null);
  const [minimized, setMinimized] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  // Draggable position: null = default (right:16, top:50%)
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);
  const dragRef = useRef<{ startX: number; startY: number; startPanelX: number; startPanelY: number } | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);
  const hostRef = useRef<HTMLElement | null>(null);
  const fieldListRef = useRef<HTMLDivElement | null>(null);
  const confirmTextareaRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    hostRef.current = document.getElementById('dtkt-overlay-host');
  }, []);

  // ── Drag logic ────────────────────────────────────────────────────────────

  const onDragStart = useCallback((e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('button')) return;
    e.preventDefault();
    const panel = panelRef.current;
    if (!panel) return;
    const rect = panel.getBoundingClientRect();
    dragRef.current = { startX: e.clientX, startY: e.clientY, startPanelX: rect.left, startPanelY: rect.top };

    const onMove = (ev: MouseEvent) => {
      if (!dragRef.current) return;
      const dx = ev.clientX - dragRef.current.startX;
      const dy = ev.clientY - dragRef.current.startY;
      setPos({ x: dragRef.current.startPanelX + dx, y: dragRef.current.startPanelY + dy });
    };
    const onUp = () => {
      dragRef.current = null;
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }, []);

  // ── Auto-activate first incomplete field on mount ──────────────────────────

  useEffect(() => {
    const first = fields.find((f) => !values[f.name]);
    if (first) setPickingField(first.name);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // ── Auto-scroll active field into view ─────────────────────────────────────

  useEffect(() => {
    if (!pickingField || !fieldListRef.current) return;
    const idx = fields.findIndex((f) => f.name === pickingField);
    const row = fieldListRef.current.children[idx] as HTMLElement | undefined;
    row?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  }, [pickingField, fields]);

  // ── Focus confirm textarea when it appears ─────────────────────────────────

  useEffect(() => {
    if (pendingConfirm) {
      // Small delay lets React finish rendering before we grab focus.
      setTimeout(() => confirmTextareaRef.current?.select(), 30);
    }
  }, [pendingConfirm]);

  // ── Crosshair cursor while picking ────────────────────────────────────────

  useEffect(() => {
    if (!pickingField) return;
    const style = document.createElement('style');
    style.textContent = '* { cursor: crosshair !important; } [data-dtkt-overlay], [data-dtkt-overlay] * { cursor: auto !important; }';
    document.head.appendChild(style);
    return () => style.remove();
  }, [pickingField]);

  // ── Pick-mode event listeners ──────────────────────────────────────────────

  useEffect(() => {
    if (!pickingField) return;

    const targetThrough = (x: number, y: number): Element | null => {
      const host = hostRef.current;
      if (host) host.style.pointerEvents = 'none';
      const el = document.elementFromPoint(x, y);
      if (host) host.style.pointerEvents = '';
      return el;
    };

    let lastEl: Element | null = null;

    // Returns true if the point is over the overlay host itself.
    const isOverOverlay = (x: number, y: number): boolean => {
      const top = document.elementFromPoint(x, y);
      return !!(top && panelRef.current?.contains(top));
    };

    const onMove = (e: MouseEvent) => {
      // Don't highlight elements while cursor is over the overlay panel.
      if (isOverOverlay(e.clientX, e.clientY)) {
        if (lastEl !== null) {
          lastEl = null;
          setHighlightRect(null);
        }
        return;
      }
      const el = targetThrough(e.clientX, e.clientY);
      if (el && el !== lastEl) {
        lastEl = el;
        setHighlightRect(el.getBoundingClientRect());
      }
    };

    const onClick = (e: MouseEvent) => {
      // Don't intercept clicks on the overlay itself (buttons, fields, etc.).
      if (isOverOverlay(e.clientX, e.clientY)) return;
      // Shift+click: suspend picking and let the page handle the click normally
      // (e.g. open a modal or follow a link so the user can pick from it).
      if (e.shiftKey) return;
      // Prevent default and stop propagation — Shift+click (above) is the
      // dedicated escape hatch to let the page handle a click normally.
      e.preventDefault();
      e.stopPropagation();
      const el = targetThrough(e.clientX, e.clientY);
      const draft = el ? ((el as HTMLElement).innerText ?? el.textContent ?? '').trim() : '';
      const capture = el
        ? { cssSelector: getCssSelector(el), xpath: getXPath(el) }
        : { cssSelector: '', xpath: '' };
      // Exit pick mode, transition to confirm dialog.
      setPickingField(null);
      setHighlightRect(null);
      setPendingConfirm({ fieldName: pickingField, draft, capture });
    };

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setPickingField(null);
        setHighlightRect(null);
      }
    };

    document.addEventListener('mousemove', onMove, true);
    document.addEventListener('click', onClick, true);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousemove', onMove, true);
      document.removeEventListener('click', onClick, true);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [pickingField]);

  // ── Confirm dialog handlers ────────────────────────────────────────────────

  const handleConfirm = useCallback(async (fieldName: string, strValue: string, capture: { cssSelector: string; xpath: string }) => {
    const value = create(ValueSchema, { kind: { case: 'stringValue', value: strValue } })
    const newValues = { ...values, [fieldName]: value };
    setValues(newValues);
    setPendingConfirm(null);

    // Persist immediately — don't wait for "Complete Task".
    void chrome.runtime.sendMessage({
      type: 'SAVE_FIELD_VALUE',
      taskId,
      fieldName,
      value,
      capture,
    } satisfies ContentRequest);

    // Auto-advance to the next incomplete field.
    const nextField = fields.find((f) => f.name !== fieldName && !newValues[f.name]);
    if (nextField) setPickingField(nextField.name);
  }, [values, taskId, fields]);

  const handleCancelConfirm = useCallback((fieldName: string) => {
    setPendingConfirm(null);
    // Re-activate the same field so the user can try again.
    setPickingField(fieldName);
  }, []);

  // ── Final submit (complete action) ────────────────────────────────────────

  const handleSubmit = useCallback(async () => {
    setSubmitting(true);
    await chrome.runtime.sendMessage({
      type: 'SUBMIT_FIELDS',
      taskId,
      values,
    } satisfies ContentRequest);
    setSubmitted(true);
    setTimeout(onUnmount, 3000);
  }, [taskId, values, onUnmount]);

  const filledCount = fields.filter((f) => values[f.name]).length;
  const allFilled = filledCount === fields.length && fields.length > 0;

  // ── Submitted toast ────────────────────────────────────────────────────────

  if (submitted) {
    return createPortal(
      <div style={S.overlayRoot}>
        <div style={S.successToast}>✓ DataKit: Task completed</div>
      </div>,
      document.body,
    );
  }

  // ── Main render ────────────────────────────────────────────────────────────

  const bodyContent = (() => {
    // ── Confirm dialog (replaces body while pending) ───────────────────────
    if (pendingConfirm) {
      const fieldDef = fields.find((f) => f.name === pendingConfirm.fieldName);
      return (
        <div style={S.confirmBody}>
          <div style={S.confirmFieldName}>{pendingConfirm.fieldName}</div>
          {fieldDef?.description && (
            <div style={S.confirmDescription}>{fieldDef.description}</div>
          )}
          <textarea
            ref={confirmTextareaRef}
            style={S.confirmTextarea}
            rows={4}
            value={pendingConfirm.draft}
            onChange={(e) => setPendingConfirm((p) => p ? { ...p, draft: e.target.value } : p)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                void handleConfirm(pendingConfirm.fieldName, pendingConfirm.draft, pendingConfirm.capture);
              }
              if (e.key === 'Escape') handleCancelConfirm(pendingConfirm.fieldName);
            }}
          />
          <div style={S.confirmHint}>⌘↵ to confirm · Esc to re-pick</div>
          <div style={S.confirmActions}>
            <button
              style={{ ...S.btn, ...S.btnGhost }}
              onClick={() => handleCancelConfirm(pendingConfirm.fieldName)}
            >
              Re-pick
            </button>
            <button
              style={{ ...S.btn, ...S.btnPrimary }}
              onClick={() => { void handleConfirm(pendingConfirm.fieldName, pendingConfirm.draft, pendingConfirm.capture); }}
            >
              Confirm ✓
            </button>
          </div>
        </div>
      );
    }

    // ── Normal field list ──────────────────────────────────────────────────
    return (
      <>
        <div style={S.progressTrack}>
          <div style={{
            height: '100%',
            width: `${fields.length > 0 ? (filledCount / fields.length) * 100 : 0}%`,
            backgroundColor: allFilled ? '#2f9e44' : '#1971c2',
            borderRadius: 2,
            transition: 'width 0.25s ease',
          }} />
        </div>
        <div style={S.progressLabel}>{filledCount} / {fields.length} captured</div>

        <div style={S.fieldList} ref={fieldListRef}>
          {fields.map((field) => {
            const value = values[field.name];
            const isPicking = pickingField === field.name;
            return (
              <div
                key={field.name}
                style={{
                  ...S.fieldRow,
                  ...(isPicking ? S.fieldRowActive : {}),
                }}
              >
                <span style={{
                  fontSize: 12,
                  color: value ? '#51cf66' : isPicking ? '#339af0' : '#5c5f66',
                  flexShrink: 0,
                  lineHeight: 1,
                  width: 12,
                  textAlign: 'center',
                }}>
                  {isPicking ? '◎' : value ? '✓' : '○'}
                </span>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ ...S.fieldName, color: isPicking ? '#e9ecef' : '#c1c2c5' }}>
                    {field.name}
                  </div>
                  {value && <div style={S.fieldValue}>{toJsonString(ValueSchema, value)}</div>}
                </div>
                <button
                  style={{ ...S.pickBtn, ...(isPicking ? S.pickBtnActive : {}) }}
                  onClick={() => {
                    setPickingField(isPicking ? null : field.name);
                    if (isPicking) setHighlightRect(null);
                  }}
                >
                  {isPicking ? 'Cancel' : value ? 'Re-pick' : 'Pick'}
                </button>
              </div>
            );
          })}
        </div>

        <div style={{ padding: '8px 12px 12px' }}>
          <button
            style={{
              ...S.submitBtn,
              opacity: submitting || filledCount === 0 ? 0.45 : 1,
              cursor: submitting || filledCount === 0 ? 'not-allowed' : 'pointer',
            }}
            onClick={() => { void handleSubmit(); }}
            disabled={submitting || filledCount === 0}
          >
            {submitting ? 'Saving…' : allFilled ? 'Complete Task ✓' : `Complete (${filledCount}/${fields.length})`}
          </button>
        </div>
      </>
    );
  })();

  return createPortal(
    <div style={S.overlayRoot}>
      {/* Element highlight box */}
      {pickingField && highlightRect && (
        <div
          style={{
            position: 'fixed',
            top: highlightRect.top - 2,
            left: highlightRect.left - 2,
            width: highlightRect.width + 4,
            height: highlightRect.height + 4,
            border: '2px solid #339af0',
            borderRadius: 3,
            backgroundColor: 'rgba(51,154,240,0.08)',
            pointerEvents: 'none',
            zIndex: 2147483646,
            boxSizing: 'border-box',
          }}
        />
      )}

      <div
        ref={panelRef}
        data-dtkt-overlay
        style={{
          ...S.panel,
          opacity: pickingField ? 0.9 : 1,
          ...(pos ? { right: 'unset' as const, top: pos.y, transform: 'none', left: pos.x } : {}),
        }}
        onClick={(e) => e.stopPropagation()}
      >

        {/* Header */}
        <div style={{ ...S.header, cursor: 'grab' }} onMouseDown={onDragStart}>
          <div style={{ ...S.dot, backgroundColor: allFilled ? '#2f9e44' : '#339af0', boxShadow: `0 0 6px ${allFilled ? '#2f9e44' : '#339af0'}` }} />
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={S.title}>{task.title}</div>
            <div style={S.subtitle}>
              {pickingField
                ? `Picking: ${pickingField} · ⇧ to click through`
                : pendingConfirm
                  ? `Confirm: ${pendingConfirm.fieldName}`
                  : `DataKit · ${fields.length} fields`}
            </div>
          </div>
          {!pendingConfirm && (
            <button
              style={S.iconBtn}
              onClick={() => setMinimized((v) => !v)}
              title={minimized ? 'Expand' : 'Collapse'}
            >
              {minimized ? '▲' : '▼'}
            </button>
          )}
        </div>

        {!minimized && bodyContent}
      </div>
    </div>,
    document.body,
  );
}

// ── Styles ────────────────────────────────────────────────────────────────────

const FONT = '-apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif';

const S = {
  // Root container that resets all CSS to prevent page interference
  overlayRoot: {
    all: 'initial' as const,
    position: 'static' as const,
    display: 'block' as const,
    visibility: 'visible' as const,
    opacity: 1,
    pointerEvents: 'none' as const,
  },
  panel: {
    position: 'fixed' as const,
    right: 16,
    top: '50%',
    transform: 'translateY(-50%)',
    width: 264,
    backgroundColor: '#1a1b1e',
    border: '1px solid #373a40',
    borderRadius: 8,
    boxShadow: '0 8px 32px rgba(0,0,0,0.7)',
    fontFamily: FONT,
    fontSize: 13,
    color: '#c1c2c5',
    zIndex: 2147483647,
    overflow: 'hidden',
    userSelect: 'none' as const,
    pointerEvents: 'auto' as const,
  },
  header: {
    display: 'flex' as const,
    alignItems: 'center' as const,
    gap: 8,
    padding: '10px 10px 10px 12px',
    backgroundColor: '#25262b',
    borderBottom: '1px solid #373a40',
  },
  dot: {
    width: 8,
    height: 8,
    borderRadius: '50%',
    backgroundColor: '#339af0',
    flexShrink: 0,
    boxShadow: '0 0 6px #339af0',
    transition: 'background-color 0.3s, box-shadow 0.3s',
  },
  title: {
    fontWeight: 600,
    fontSize: 13,
    color: '#e9ecef',
    whiteSpace: 'nowrap' as const,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  },
  subtitle: { fontSize: 11, color: '#5c5f66', marginTop: 1 },
  iconBtn: {
    background: 'none',
    border: 'none',
    color: '#5c5f66',
    cursor: 'pointer',
    padding: '2px 4px',
    fontSize: 10,
    flexShrink: 0,
    lineHeight: 1,
  },
  progressTrack: {
    height: 3,
    backgroundColor: '#2c2e33',
    margin: '10px 12px 4px',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressLabel: { fontSize: 11, color: '#5c5f66', padding: '0 12px 6px' },
  fieldList: {
    borderTop: '1px solid #2c2e33',
    maxHeight: 280,
    overflowY: 'auto' as const,
  },
  fieldRow: {
    display: 'flex' as const,
    alignItems: 'center' as const,
    gap: 8,
    padding: '7px 10px 7px 12px',
    borderBottom: '1px solid #2c2e33',
    transition: 'background-color 0.15s',
  },
  fieldRowActive: {
    backgroundColor: 'rgba(51,154,240,0.06)',
  },
  fieldName: { fontSize: 12, fontWeight: 500, color: '#c1c2c5' },
  fieldValue: {
    fontSize: 11,
    color: '#868e96',
    marginTop: 2,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap' as const,
    maxWidth: 148,
  },
  pickBtn: {
    fontSize: 11,
    padding: '3px 8px',
    borderRadius: 4,
    border: '1px solid #373a40',
    backgroundColor: '#25262b',
    color: '#c1c2c5',
    cursor: 'pointer',
    flexShrink: 0,
    fontFamily: FONT,
  },
  pickBtnActive: {
    borderColor: '#339af0',
    color: '#339af0',
    backgroundColor: 'rgba(51,154,240,0.1)',
  },
  submitBtn: {
    width: '100%',
    padding: '8px 0',
    backgroundColor: '#1971c2',
    color: '#fff',
    border: 'none',
    borderRadius: 6,
    fontSize: 13,
    fontWeight: 600,
    fontFamily: FONT,
  },
  // ── Confirm dialog ──────────────────────────────────────────────────────────
  confirmBody: {
    padding: '12px',
    borderTop: '1px solid #2c2e33',
  },
  confirmFieldName: {
    fontWeight: 600,
    fontSize: 13,
    color: '#e9ecef',
    marginBottom: 4,
  },
  confirmDescription: {
    fontSize: 11,
    color: '#5c5f66',
    marginBottom: 8,
    lineHeight: 1.4,
  },
  confirmTextarea: {
    width: '100%',
    boxSizing: 'border-box' as const,
    backgroundColor: '#25262b',
    border: '1px solid #373a40',
    borderRadius: 5,
    color: '#c1c2c5',
    fontFamily: FONT,
    fontSize: 12,
    lineHeight: 1.5,
    padding: '7px 9px',
    resize: 'vertical' as const,
    outline: 'none',
  },
  confirmHint: {
    fontSize: 10,
    color: '#5c5f66',
    marginTop: 4,
    marginBottom: 10,
  },
  confirmActions: {
    display: 'flex' as const,
    gap: 8,
  },
  btn: {
    flex: 1,
    padding: '7px 0',
    borderRadius: 5,
    fontSize: 12,
    fontWeight: 600,
    fontFamily: FONT,
    cursor: 'pointer',
    border: '1px solid transparent',
  },
  btnGhost: {
    backgroundColor: 'transparent',
    borderColor: '#373a40',
    color: '#868e96',
  },
  btnPrimary: {
    backgroundColor: '#1971c2',
    color: '#fff',
    border: 'none',
  },
  successToast: {
    position: 'fixed' as const,
    bottom: 24,
    right: 24,
    backgroundColor: '#1a1b1e',
    border: '1px solid #2f9e44',
    color: '#51cf66',
    padding: '10px 18px',
    borderRadius: 8,
    fontSize: 14,
    fontWeight: 600,
    fontFamily: FONT,
    boxShadow: '0 4px 16px rgba(0,0,0,0.6)',
    zIndex: 2147483647,
    pointerEvents: 'auto' as const,
  },
} as const;
