import { initNativePort } from "./background/native-port";
import { handleAction } from "./background/action-dispatcher";
import { handleConfig } from "./lib/config";

// ─── Wire native port → action dispatcher ────────────────────────────────────
initNativePort(handleConfig, handleAction);

// ─── Register all Chrome event listeners ─────────────────────────────────────
import "./background/chrome-listeners";

// ─── Register action queue (storage, popup messaging, tab-removal) ───────────
import "./background/task-queue";
