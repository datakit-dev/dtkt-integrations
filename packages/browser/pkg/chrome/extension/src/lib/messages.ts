// ─── Background ↔ Popup/Content Message Protocol ─────────────────────────────
//
// All messages flow through chrome.runtime.sendMessage / onMessage.
// Task data is fetched directly from the server via Connect — background
// only handles tab lifecycle and config distribution.
//
// NOTE: Content scripts must NEVER make direct fetch() calls to localhost.
// Chrome's Private Network Access policy treats those as cross-origin requests
// from the web-page origin and shows a permission dialog. Instead, content
// scripts ask the background (extension origin, exempt from that policy) to
// proxy any Connect calls via GET_TASK / similar messages.

import { Value } from "@/proto/google/protobuf/struct_pb";

// ── Popup → Background requests ───────────────────────────────────────────────

/** Open a new tab for the task and register it as active. */
export type StartTaskRequest = { type: "START_TASK"; taskId: string; url: string };

/** Dismiss an active task: push DISMISS_OVERLAY and clean up the tab mapping. */
export type DismissTaskRequest = { type: "DISMISS_TASK"; taskId: string };

/** Focus the tab associated with an active task. */
export type FocusTaskTabRequest = { type: "FOCUS_TASK_TAB"; taskId: string };

export type PopupRequest =
  | StartTaskRequest
  | DismissTaskRequest
  | FocusTaskTabRequest
  | GetContextRequest;

// ── Background → Popup responses ──────────────────────────────────────────────

export type StartTaskResponse = { type: "START_TASK"; ok: boolean; tabId?: number };
export type DismissTaskResponse = { type: "DISMISS_TASK"; ok: boolean };
export type FocusTaskTabResponse = { type: "FOCUS_TASK_TAB"; ok: boolean };

export type PopupResponse =
  | StartTaskResponse
  | DismissTaskResponse
  | FocusTaskTabResponse
  | GetContextResponse;

// ── Content Script ↔ Background messages ─────────────────────────────────────

/** Ask background which task is active in this tab (returns taskId only). */
export type GetActiveTaskRequest = { type: "GET_ACTIVE_TASK" };
export type GetActiveTaskResponse = { type: "GET_ACTIVE_TASK"; taskId: string | null };

/**
 * Ask background to fetch a specific task from the Connect server.
 * Background makes the request (extension origin — no Private Network Access
 * dialog) and returns the raw task JSON to the content script.
 */
export type GetTaskRequest = { type: "GET_TASK"; taskId: string };
export type GetTaskResponse = { type: "GET_TASK"; taskJson: string | null };

/** Ask background for the Connect transport context so content can init its transport. */
export type GetContextRequest = { type: "GET_CONTEXT" };
export type GetContextResponse = { type: "GET_CONTEXT"; contextJson: string | null };

/**
 * Ask background to fetch an ExtractionSchema from the Connect server.
 */
export type GetExtractionSchemaRequest = { type: "GET_EXTRACTION_SCHEMA"; schemaId: string };
export type GetExtractionSchemaResponse = { type: "GET_EXTRACTION_SCHEMA"; schemaJson: string | null };

/**
 * Ask background to fetch an ExtractionRecord by task ID.
 * Returns the first record associated with the task, or null if none exists yet.
 */
export type GetExtractionRecordRequest = { type: "GET_EXTRACTION_RECORD"; taskId: string };
export type GetExtractionRecordResponse = { type: "GET_EXTRACTION_RECORD"; recordJson: string | null };

/**
 * Persist a captured field value. Background forwards to the native bridge;
 * will become a direct UpdateTask Connect call once that RPC is wired.
 */
export type SaveFieldValueRequest = {
  type: "SAVE_FIELD_VALUE";
  taskId: string;
  fieldName: string;
  value: Value;
  capture: { cssSelector: string; xpath: string };
};

/**
 * Submit all captured field values as a completed task. Background calls
 * UpdateTask with state=COMPLETED via the Connect server.
 */
export type SubmitFieldsRequest = {
  type: "SUBMIT_FIELDS";
  taskId: string;
  values: Record<string, Value>;
};

export type SubmitFieldsResponse = { type: "SUBMIT_FIELDS"; ok: boolean };

/**
 * Request to create a new task. Background forwards to TaskService.CreateTask.
 */
export type CreateTaskRequest = {
  type: "CREATE_TASK";
  title: string;
  url: string;
  schemaId: string;
};

export type CreateTaskResponse = { type: "CREATE_TASK"; ok: boolean; taskId?: string; error?: string };

/**
 * Request to list all extraction schemas.
 */
export type ListSchemasRequest = { type: "LIST_SCHEMAS" };
export type ListSchemasResponse = { type: "LIST_SCHEMAS"; schemasJson: string | null };

export type ContentRequest =
  | GetActiveTaskRequest
  | GetTaskRequest
  | GetContextRequest
  | GetExtractionSchemaRequest
  | GetExtractionRecordRequest
  | SaveFieldValueRequest
  | SubmitFieldsRequest
  | CreateTaskRequest
  | ListSchemasRequest;
export type ContentResponse =
  | GetActiveTaskResponse
  | GetTaskResponse
  | GetContextResponse
  | GetExtractionSchemaResponse
  | GetExtractionRecordResponse
  | SubmitFieldsResponse
  | CreateTaskResponse
  | ListSchemasResponse;

/** Push from background → content script (no sendResponse). */
export type ContentPush =
  | { type: "DISMISS_OVERLAY" }
  | { type: "SHOW_TASK_CREATION"; title: string; url: string };
