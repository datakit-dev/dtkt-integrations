// ─── Task Types ───────────────────────────────────────────────────────────────

import type { Task } from "../proto/integration/browser/v1beta/task_pb";
import type { Timestamp } from "../proto/google/protobuf/timestamp_pb";

export type { Task } from "../proto/integration/browser/v1beta/task_pb";
export { TaskSchema } from "../proto/integration/browser/v1beta/task_pb";
export type { FieldDef, ExtractionSchema, ExtractionRecord } from "../proto/integration/browser/v1beta/extraction_pb";
export { ExtractionSchemaSchema, ExtractionRecordSchema, ElementCaptureSchema } from "../proto/integration/browser/v1beta/extraction_pb";

/**
 * Runtime view of a task enriched with ephemeral active-tab state.
 * Used only in the popup — never persisted.
 */
export type TaskView =
  | ({ active: false } & Task)
  | ({ active: true; tabId: number } & Task);

/** Convert a proto Timestamp to epoch milliseconds for display/comparison. */
export function tsToMs(ts: Timestamp | undefined): number {
  if (!ts) return 0;
  return Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000);
}
