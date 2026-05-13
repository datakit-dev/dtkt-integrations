import type { ContentRequest, ContentResponse, ContentPush } from './lib/messages';
import { mountOverlay } from './overlay/mount';
import { mountTaskCreationOverlay } from './overlay/mountTaskCreation';
import { fromJson } from '@bufbuild/protobuf';
import { TaskSchema } from './proto/integration/browser/v1beta/task_pb';
import { ExtractionSchemaSchema, ExtractionRecordSchema } from './proto/integration/browser/v1beta/extraction_pb';
import type { ExtractionSchema, ExtractionRecord } from './proto/integration/browser/v1beta/extraction_pb';

// Wait for DOM to be ready to avoid interfering with React hydration on sites like LinkedIn
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}

// Listen for push messages from background
chrome.runtime.onMessage.addListener((msg: ContentPush) => {
  if (msg.type === 'SHOW_TASK_CREATION') {
    mountTaskCreationOverlay(msg.title, msg.url);
  }
});

async function init(): Promise<void> {
  // Ask background for the active task ID. Background also holds the server
  // address but content scripts must not fetch localhost directly — Chrome's
  // Private Network Access policy triggers a permission dialog for web-page
  // origins. Instead we ask background to proxy the task fetch (GET_TASK).
  let activeRes: ContentResponse;
  try {
    activeRes = await chrome.runtime.sendMessage({ type: 'GET_ACTIVE_TASK' } satisfies ContentRequest);
  } catch (err) {
    return;
  }

  if (!activeRes || activeRes.type !== 'GET_ACTIVE_TASK' || activeRes.taskId === null) {
    return;
  }

  let taskRes: ContentResponse;
  try {
    taskRes = await chrome.runtime.sendMessage({ type: 'GET_TASK', taskId: activeRes.taskId } satisfies ContentRequest);
  } catch (err) {
    return;
  }

  if (!taskRes || taskRes.type !== 'GET_TASK' || taskRes.taskJson === null) {
    return;
  }

  const task = fromJson(TaskSchema, JSON.parse(taskRes.taskJson));

  if (task.payload.case !== 'extraction') {
    return;
  }
  const schemaId = task.payload.value.schemaId;
  if (!schemaId) {
    return;
  }

  // Fetch the schema (field definitions) and any existing record (current values).
  const [schemaRes, recordRes] = await Promise.all([
    chrome.runtime.sendMessage({ type: 'GET_EXTRACTION_SCHEMA', schemaId } satisfies ContentRequest).catch(() => null) as Promise<ContentResponse | null>,
    chrome.runtime.sendMessage({ type: 'GET_EXTRACTION_RECORD', taskId: activeRes.taskId } satisfies ContentRequest).catch(() => null) as Promise<ContentResponse | null>,
  ]);

  let schema: ExtractionSchema | null = null;
  if (schemaRes && schemaRes.type === 'GET_EXTRACTION_SCHEMA' && schemaRes.schemaJson) {
    schema = fromJson(ExtractionSchemaSchema, JSON.parse(schemaRes.schemaJson));
  }
  if (!schema) {
    return;
  }

  let record: ExtractionRecord | null = null;
  if (recordRes && recordRes.type === 'GET_EXTRACTION_RECORD' && recordRes.recordJson) {
    record = fromJson(ExtractionRecordSchema, JSON.parse(recordRes.recordJson));
  }

  const unmount = mountOverlay(activeRes.taskId, task, schema, record);

  // Listen for background push (e.g. task dismissed from popup).
  chrome.runtime.onMessage.addListener((msg: ContentPush) => {
    if (msg.type === 'DISMISS_OVERLAY') unmount();
  });
}
