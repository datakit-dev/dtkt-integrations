import { useCallback, useEffect, useRef, useState } from 'react';
import { Box, Divider, Flex, LoadingOverlay, ScrollArea, Tabs, Text, Title } from '@mantine/core';
import { IconBrandChrome, IconClockHour4, IconStack2 } from '@tabler/icons-react';
import { AppProviders } from '../components/AppProviders';
import { QueueTab } from '../components/QueueTab';
import { HistoryTab } from '../components/HistoryTab';
import type { Task, TaskView } from '../lib/types';
import type { PopupRequest, PopupResponse } from '../lib/messages';
import { type Transport } from '@connectrpc/connect';
import { ListTasksRequestSchema, TaskService, TaskState } from '../proto/integration/browser/v1beta/task_pb';
import { create, fromJsonString } from '@bufbuild/protobuf';
import { ContextSchema } from '@/proto/integration/browser/v1beta/context_pb';
import { getTransport as _getTransport } from '@/lib/transport';
import { createStrictClient } from '@/lib/client';

function sendMessage(req: PopupRequest): Promise<PopupResponse> {
  return chrome.runtime.sendMessage(req);
}

// Active task IDs known from this popup session (taskId → tabId).
// Populated by START_TASK responses; used to enrich task list with active state.
const activeTabIds = new Map<string, number>();

function buildTaskViews(tasks: Task[]): TaskView[] {
  return tasks.map((task) => {
    const tabId = activeTabIds.get(task.id);
    return tabId !== undefined
      ? { ...task, active: true as const, tabId }
      : { ...task, active: false as const };
  });
}

export const App = () => {
  const [tasks, setTasks] = useState<TaskView[]>([]);
  const [loading, setLoading] = useState(true);
  const transportRef = useRef<Transport | null>(null);

  // ── Initialize transport from background config ───────────────────────────

  const getTransport = useCallback(async (): Promise<Transport | null> => {
    if (transportRef.current) return transportRef.current;
    const res = await sendMessage({ type: 'GET_CONTEXT' });
    if (res.type === 'GET_CONTEXT' && res.contextJson) {
      const context = fromJsonString(ContextSchema, res.contextJson);
      if (!context) {
        console.error('Failed to parse context JSON:', res.contextJson);
        return null;
      }
      transportRef.current = _getTransport(context);
    }
    return transportRef.current;
  }, []);

  // ── Fetch task list directly from server ──────────────────────────────────

  const refresh = useCallback(async () => {
    const transport = await getTransport();
    if (!transport) { setLoading(false); return; }
    const client = createStrictClient(TaskService, transport);
    const res = await (client.listTasks(create(ListTasksRequestSchema, {})));
    setTasks(buildTaskViews(res.tasks));
    setLoading(false);
  }, [getTransport]);

  useEffect(() => { void refresh(); }, [refresh]);

  const handleStart = useCallback(async (taskId: string, url: string) => {
    const res = await sendMessage({ type: 'START_TASK', taskId, url });
    if (res.type === 'START_TASK' && res.ok && res.tabId !== undefined) {
      activeTabIds.set(taskId, res.tabId);
    }
    void refresh();
  }, [refresh]);

  const handleFocus = useCallback(async (taskId: string) => {
    await sendMessage({ type: 'FOCUS_TASK_TAB', taskId });
  }, []);

  const handleDismiss = useCallback(async (taskId: string) => {
    await sendMessage({ type: 'DISMISS_TASK', taskId });
    activeTabIds.delete(taskId);
    void refresh();
  }, [refresh]);

  const queueTasks = tasks.filter(
    (t) => t.active || t.state === TaskState.PENDING || t.state === TaskState.ACTIVE,
  );
  const historyTasks = tasks.filter(
    (t) => t.state === TaskState.COMPLETED || t.state === TaskState.DISMISSED,
  ) as Task[];

  const pendingCount = queueTasks.filter((t) => !t.active).length;
  const activeCount = queueTasks.filter((t) => t.active).length;

  return (
    <AppProviders>
      <Box style={{ position: 'relative', width: '100%' }}>
        <LoadingOverlay visible={loading} zIndex={1000} overlayProps={{ blur: 2 }} />

        {/* Header */}
        <Flex align="center" gap="sm" px="md" pt="md" pb="sm">
          <IconBrandChrome size={22} color="var(--mantine-color-blue-5)" />
          <Box style={{ flex: 1, minWidth: 0 }}>
            <Title order={5} lh={1.2}>DataKit</Title>
            <Text size="xs" c="dimmed">Task Queue</Text>
          </Box>
          {(pendingCount + activeCount) > 0 && (
            <Flex align="center" gap={4}>
              {activeCount > 0 && (
                <Text size="xs" c="blue" fw={600}>{activeCount} active</Text>
              )}
              {pendingCount > 0 && activeCount > 0 && (
                <Text size="xs" c="dimmed">·</Text>
              )}
              {pendingCount > 0 && (
                <Text size="xs" c="dimmed">{pendingCount} pending</Text>
              )}
            </Flex>
          )}
        </Flex>

        <Divider />

        <Tabs defaultValue="queue" keepMounted={false}>
          <Tabs.List px="md" style={{ borderBottom: '1px solid var(--mantine-color-dark-4)' }}>
            <Tabs.Tab
              value="queue"
              leftSection={<IconStack2 size={14} />}
              py="xs"
            >
              <Text size="sm">Queue</Text>
            </Tabs.Tab>
            <Tabs.Tab
              value="history"
              leftSection={<IconClockHour4 size={14} />}
              py="xs"
            >
              <Text size="sm">History</Text>
            </Tabs.Tab>
          </Tabs.List>

          <ScrollArea.Autosize mah={420} type="hover">
            <Box p="sm">
              <Tabs.Panel value="queue">
                <QueueTab
                  tasks={queueTasks}
                  onStart={handleStart}
                  onFocus={handleFocus}
                  onDismiss={handleDismiss}
                />
              </Tabs.Panel>
              <Tabs.Panel value="history">
                <HistoryTab actions={historyTasks} />
              </Tabs.Panel>
            </Box>
          </ScrollArea.Autosize>
        </Tabs>
      </Box>
    </AppProviders>
  );
};
