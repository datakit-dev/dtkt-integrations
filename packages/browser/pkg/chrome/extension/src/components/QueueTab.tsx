import { Box, Center, Stack, Text } from "@mantine/core";
import { IconInboxOff } from "@tabler/icons-react";
import { TaskCard } from "./TaskCard";
import type { TaskView } from "../lib/types";
import { TaskState } from "@/proto/integration/browser/v1beta/task_pb";

type Props = {
  tasks: TaskView[];
  onStart: (taskId: string, url: string) => void;
  onFocus: (taskId: string) => void;
  onDismiss: (taskId: string) => void;
};

export const QueueTab = ({ tasks, onStart, onFocus, onDismiss }: Props) => {
  const active = tasks.filter((a) => a.active || a.state === TaskState.ACTIVE);
  const pending = tasks.filter((a) => !a.active && a.state === TaskState.PENDING);

  if (active.length === 0 && pending.length === 0) {
    return (
      <Center py="xl">
        <Stack align="center" gap="xs">
          <IconInboxOff size={32} color="var(--mantine-color-dark-3)" />
          <Text size="sm" c="dimmed">No pending tasks</Text>
          <Text size="xs" c="dimmed">Tasks queued from DataKit will appear here.</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="xs">
      {active.length > 0 && (
        <>
          <Text size="xs" fw={600} tt="uppercase" c="dimmed">Active</Text>
          {active.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              onStart={onStart}
              onFocus={onFocus}
              onDismiss={onDismiss}
            />
          ))}
          {pending.length > 0 && (
            <Box mt={4}>
              <Text size="xs" fw={600} tt="uppercase" c="dimmed">Pending</Text>
            </Box>
          )}
        </>
      )}
      {pending.map((task) => (
        <TaskCard
          key={task.id}
          task={task}
          onStart={onStart}
          onFocus={onFocus}
          onDismiss={onDismiss}
        />
      ))}
    </Stack>
  );
};
