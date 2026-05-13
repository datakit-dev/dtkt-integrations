import {
  type PopupRequest,
  type PopupResponse,
  type ContentRequest,
  type ContentResponse,
  type ContentPush,
} from "../lib/messages";
import { ChromeEventSchema, RuntimeEventSchema } from "../proto/integration/browser/v1beta/chrome_pb";
import {
  TaskService,
  TaskSchema,
  TaskState,
  type Task,
} from "../proto/integration/browser/v1beta/task_pb";
import {
  ExtractionSchemaService,
  ExtractionSchemaSchema,
  ExtractionRecordService,
  ExtractionRecordSchema,
  ElementCaptureSchema,
  ListExtractionRecordsRequestSchema,
  ExtractionTaskSchema,
  ListExtractionSchemasRequestSchema,
  ListExtractionSchemasResponseSchema,
} from "../proto/integration/browser/v1beta/extraction_pb";
import { create, toJsonString } from "@bufbuild/protobuf";
import { handleEvent } from "./native-port";
import { getContext } from "../lib/config";
import { getTransport } from "../lib/transport";
import { ContextSchema } from "@/proto/integration/browser/v1beta/context_pb";
import { createStrictClient } from "@/lib/client";

// ─── Connect client helper ────────────────────────────────────────────────────

function makeTaskClient() {
  const context = getContext();
  if (!context) return null;
  return createStrictClient(TaskService, getTransport(context));
}

function makeExtractionSchemaClient() {
  const context = getContext();
  if (!context) return null;
  return createStrictClient(ExtractionSchemaService, getTransport(context));
}

function makeExtractionRecordClient() {
  const context = getContext();
  if (!context) return null;
  return createStrictClient(ExtractionRecordService, getTransport(context));
}

function activateTask(taskId: string): void {
  const client = makeTaskClient();
  if (!client) return;
  void (client.updateTask({
    task: create(TaskSchema, { id: taskId, state: TaskState.ACTIVE }),
    updateMask: { paths: ['state'] },
  }));
}

// ─── Active-tab state ─────────────────────────────────────────────────────────
//
// Ephemeral — lives only for the lifetime of the service worker.
// Maps taskId → tabId for currently-active tasks.

const activeTasks = new Map<string, number>();

export function getTaskIdForTab(tabId: number): string | undefined {
  for (const [taskId, t] of activeTasks) {
    if (t === tabId) return taskId;
  }
  return undefined;
}

// ─── Tab-removal listener ─────────────────────────────────────────────────────

chrome.tabs.onRemoved.addListener((tabId) => {
  for (const [taskId, activeTabId] of activeTasks) {
    if (activeTabId === tabId) {
      activeTasks.delete(taskId);
      // Tab closed before completion — return to pending so it re-appears in the queue.
      const client = makeTaskClient();
      if (client) {
        void (client.getTask({ id: taskId }))
          .then((res) => {
            if (res.task?.state === TaskState.ACTIVE) {
              void (client.updateTask({
                task: create(TaskSchema, { id: taskId, state: TaskState.PENDING }),
                updateMask: { paths: ['state'] },
              }));
            }
          });
      }
    }
  }
});

// ─── Message handler ──────────────────────────────────────────────────────────

type AnyRequest = PopupRequest | ContentRequest;
type AnyResponse = PopupResponse | ContentResponse;

chrome.runtime.onMessage.addListener(
  (request: AnyRequest, _sender, sendResponse: (response: AnyResponse) => void) => {
    switch (request.type) {
      case "GET_CONTEXT": {
        const context = getContext();
        if (!context) {
          sendResponse({ type: "GET_CONTEXT", contextJson: null });
          return false;
        }

        sendResponse({ type: "GET_CONTEXT", contextJson: toJsonString(ContextSchema, context) });
        return false;
      }

      case "START_TASK": {
        // If the task already has an open tab, just focus it — no duplicate tabs.
        const existingTabId = activeTasks.get(request.taskId);
        if (existingTabId !== undefined) {
          chrome.tabs.get(existingTabId).then((tab) => {
            chrome.tabs.update(existingTabId, { active: true });
            if (tab.windowId !== undefined) {
              chrome.windows.update(tab.windowId, { focused: true });
            }
            sendResponse({ type: "START_TASK", ok: true, tabId: existingTabId });
          }).catch(() => {
            // Tab was closed without us knowing — fall through to open a new one.
            activeTasks.delete(request.taskId);
            chrome.tabs.create({ url: request.url, active: true }).then((tab) => {
              if (tab.id !== undefined) activeTasks.set(request.taskId, tab.id);
              activateTask(request.taskId);
              sendResponse({ type: "START_TASK", ok: true, tabId: tab.id });
            });
          });
          return true;
        }
        chrome.tabs.create({ url: request.url, active: true }).then((tab) => {
          if (tab.id !== undefined) activeTasks.set(request.taskId, tab.id);
          activateTask(request.taskId);
          sendResponse({ type: "START_TASK", ok: true, tabId: tab.id });
        });
        return true;
      }

      case "DISMISS_TASK": {
        const tabId = activeTasks.get(request.taskId);
        if (tabId !== undefined) {
          chrome.tabs.sendMessage(tabId, { type: "DISMISS_OVERLAY" } satisfies ContentPush).catch(() => { });
        }
        activeTasks.delete(request.taskId);
        sendResponse({ type: "DISMISS_TASK", ok: true });
        return false;
      }

      case "GET_TASK": {
        // Background (extension origin) makes the Connect call so content scripts
        // are never the fetch origin — avoids the Private Network Access dialog.
        const client = makeTaskClient();
        if (!client) {
          sendResponse({ type: "GET_TASK", taskJson: null });
          return false;
        }
        void (client.getTask({ id: request.taskId }))
          .then((res) => {
            sendResponse({ type: "GET_TASK", taskJson: toJsonString(TaskSchema, res.task as Task) });
          })
          .catch(() => {
            sendResponse({ type: "GET_TASK", taskJson: null });
          });
        return true;
      }

      case "GET_EXTRACTION_SCHEMA": {
        const client = makeExtractionSchemaClient();
        if (!client) {
          sendResponse({ type: "GET_EXTRACTION_SCHEMA", schemaJson: null });
          return false;
        }
        void client.getExtractionSchema({ id: request.schemaId })
          .then((res) => {
            sendResponse({
              type: "GET_EXTRACTION_SCHEMA",
              schemaJson: res.schema ? toJsonString(ExtractionSchemaSchema, res.schema) : null,
            });
          })
          .catch(() => {
            sendResponse({ type: "GET_EXTRACTION_SCHEMA", schemaJson: null });
          });
        return true;
      }

      case "GET_EXTRACTION_RECORD": {
        const client = makeExtractionRecordClient();
        if (!client) {
          sendResponse({ type: "GET_EXTRACTION_RECORD", recordJson: null });
          return false;
        }
        void client.listExtractionRecords(create(ListExtractionRecordsRequestSchema, { taskId: request.taskId }))
          .then((res) => {
            const record = res.records?.[0] ?? null;
            sendResponse({
              type: "GET_EXTRACTION_RECORD",
              recordJson: record ? toJsonString(ExtractionRecordSchema, record) : null,
            });
          })
          .catch(() => {
            sendResponse({ type: "GET_EXTRACTION_RECORD", recordJson: null });
          });
        return true;
      }

      // Persist a captured field value into an ExtractionRecord.
      // Creates a new record on the first save, then updates the task's record_id.
      case "SAVE_FIELD_VALUE": {
        const taskClient = makeTaskClient();
        const recordClient = makeExtractionRecordClient();
        if (!taskClient || !recordClient) return false;

        // const fieldDef = fields.find((f) => f.name === fieldName);
        // let value: Value;
        // switch (fieldDef?.type) {
        //   case 'string':
        //     value = create(ValueSchema, { kind: { case: 'stringValue', value: strValue } });
        //     break;
        //   case 'number':
        //     value = create(ValueSchema, { kind: { case: 'numberValue', value: parseFloat(strValue) || 0 } });
        //     break;
        //   case 'boolean':
        //     value = create(ValueSchema, { kind: { case: 'boolValue', value: strValue.toLowerCase() === 'true' } });
        //     break;
        //   default:
        // }

        void (taskClient.getTask({ id: request.taskId }))
          .then(async (res) => {
            const task = res.task;
            if (!task || task.payload.case !== 'extraction') return;
            const { schemaId, recordId } = task.payload.value;

            const captureMsg = create(ElementCaptureSchema, {
              cssSelector: request.capture.cssSelector,
              xpath: request.capture.xpath,
            });

            if (!recordId) {
              // First save — create the record then update the task with the new record_id.
              const created = await recordClient.createExtractionRecord({
                record: create(ExtractionRecordSchema, {
                  schemaId,
                  taskId: request.taskId,
                  values: {
                    [request.fieldName]: request.value,
                  },
                  captures: { [request.fieldName]: captureMsg },
                }),
              });
              const newRecordId = created.record?.id;
              if (newRecordId) {
                await taskClient.updateTask({
                  task: create(TaskSchema, {
                    id: request.taskId,
                    payload: { case: 'extraction', value: { schemaId, recordId: newRecordId } },
                  }),
                  updateMask: { paths: ['extraction'] },
                });
              }
            } else {
              // Subsequent saves — fetch current record, merge, then update.
              const existing = await recordClient.listExtractionRecords(create(ListExtractionRecordsRequestSchema, { taskId: request.taskId }));
              const current = existing.records?.[0];
              const mergedValues = {
                ...(current?.values ?? {}), [request.fieldName]: request.value,
              };
              const mergedCaptures = { ...(current?.captures ?? {}), [request.fieldName]: captureMsg };
              await recordClient.updateExtractionRecord({
                record: create(ExtractionRecordSchema, {
                  id: recordId,
                  values: mergedValues,
                  captures: mergedCaptures,
                }),
                updateMask: { paths: ['values', 'captures'] },
              });
            }
          })
          .catch(() => { });
        return false;
      }

      // Submit all captured field values and mark the task completed.
      case "SUBMIT_FIELDS": {
        const taskClient = makeTaskClient();
        const recordClient = makeExtractionRecordClient();
        if (!taskClient || !recordClient) {
          sendResponse({ type: "SUBMIT_FIELDS", ok: false });
          return false;
        }
        void (taskClient.getTask({ id: request.taskId }))
          .then(async (res) => {
            const task = res.task;
            if (!task || task.payload.case !== 'extraction') return;
            const { schemaId, recordId } = task.payload.value;

            if (recordId) {
              // Update the record with the final submitted values.
              await recordClient.updateExtractionRecord({
                record: create(ExtractionRecordSchema, {
                  id: recordId,
                  values: request.values,
                }),
                updateMask: { paths: ['values'] },
              });
            } else {
              // No record yet — create one with the final values.
              const created = await recordClient.createExtractionRecord({
                record: create(ExtractionRecordSchema, {
                  schemaId,
                  taskId: request.taskId,
                  values: request.values,
                }),
              });
              const newRecordId = created.record?.id;
              if (newRecordId) {
                await taskClient.updateTask({
                  task: create(TaskSchema, {
                    id: request.taskId,
                    payload: { case: 'extraction', value: { schemaId, recordId: newRecordId } },
                  }),
                  updateMask: { paths: ['extraction'] },
                });
              }
            }

            await taskClient.updateTask({
              task: create(TaskSchema, { id: request.taskId, state: TaskState.COMPLETED }),
              updateMask: { paths: ['state'] },
            });
          })
          .then(() => {
            sendResponse({ type: "SUBMIT_FIELDS", ok: true });
          })
          .catch(() => {
            sendResponse({ type: "SUBMIT_FIELDS", ok: false });
          });
        return true; // keep channel open until updateTask resolves
      }

      case "GET_ACTIVE_TASK": {
        const tabId = _sender.tab?.id;
        if (tabId === undefined) {
          sendResponse({ type: "GET_ACTIVE_TASK", taskId: null });
          return false;
        }
        sendResponse({ type: "GET_ACTIVE_TASK", taskId: getTaskIdForTab(tabId) ?? null });
        return false;
      }

      case "FOCUS_TASK_TAB": {
        const tabId = activeTasks.get(request.taskId);
        if (tabId === undefined) {
          sendResponse({ type: "FOCUS_TASK_TAB", ok: false });
          return false;
        }
        chrome.tabs.get(tabId).then((tab) => {
          chrome.tabs.update(tabId, { active: true });
          if (tab.windowId !== undefined) {
            chrome.windows.update(tab.windowId, { focused: true });
          }
          sendResponse({ type: "FOCUS_TASK_TAB", ok: true });
        }).catch(() => {
          activeTasks.delete(request.taskId);
          sendResponse({ type: "FOCUS_TASK_TAB", ok: false });
        });
        return true;
      }

      case "CREATE_TASK": {
        const client = makeTaskClient();
        if (!client) {
          sendResponse({ type: "CREATE_TASK", ok: false, error: "Extension not connected to DataKit server" });
          return false;
        }
        void client.createTask({
          task: create(TaskSchema, {
            title: request.title,
            url: request.url,
            state: TaskState.PENDING,
            payload: {
              case: "extraction",
              value: create(ExtractionTaskSchema, {
                schemaId: request.schemaId,
              }),
            },
          }),
        })
          .then((res) => {
            if (!res.task || !res.task.id) {
              sendResponse({ type: "CREATE_TASK", ok: false, error: "Invalid response from server" });
              return;
            }
            sendResponse({ type: "CREATE_TASK", ok: true, taskId: res.task.id });
            // Show success notification
            chrome.notifications.create({
              type: "basic",
              iconUrl: chrome.runtime.getURL("assets/icon-color-on-blue.png"),
              title: "Task Created",
              message: `"${res.task.title}" has been added to your DataKit queue`,
            });
          })
          .catch((error) => {
            sendResponse({
              type: "CREATE_TASK",
              ok: false,
              error: error instanceof Error ? error.message : "Unknown error occurred"
            });
          });
        return true;
      }

      case "LIST_SCHEMAS": {
        const client = makeExtractionSchemaClient();
        if (!client) {
          sendResponse({ type: "LIST_SCHEMAS", schemasJson: null });
          return false;
        }
        void client.listExtractionSchemas(create(ListExtractionSchemasRequestSchema, {}))
          .then((res) => {
            sendResponse({ type: "LIST_SCHEMAS", schemasJson: toJsonString(ListExtractionSchemasResponseSchema, res) });
          })
          .catch(() => {
            sendResponse({ type: "LIST_SCHEMAS", schemasJson: null });
          });
        return true;
      }

      default:
        // Forward to the native bridge as RuntimeEvent.onMessage.
        handleEvent(create(ChromeEventSchema, {
          type: {
            case: "runtime",
            value: create(RuntimeEventSchema, {
              type: {
                case: "onMessage",
                value: {
                  message: JSON.stringify(request),
                  senderUrl: _sender.url,
                  senderTabId: _sender.tab?.id,
                  senderExtId: _sender.id,
                },
              },
            }),
          },
        }));
        return false;
    }
  },
);
