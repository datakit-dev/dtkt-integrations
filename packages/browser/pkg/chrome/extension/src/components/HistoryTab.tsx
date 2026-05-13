import { Center, Stack, Text } from "@mantine/core";
import { IconClockOff } from "@tabler/icons-react";
import { HistoryCard } from "./HistoryCard";
import { tsToMs } from "../lib/types";
import type { Task } from "../lib/types";
import { TaskState } from "../proto/integration/browser/v1beta/task_pb";

type Props = {
  actions: Task[];
};

export const HistoryTab = ({ actions }: Props) => {
  const history = actions
    .filter((t) => t.state === TaskState.COMPLETED || t.state === TaskState.DISMISSED)
    .sort((a, b) => tsToMs(b.completeTime ?? b.createTime) - tsToMs(a.completeTime ?? a.createTime));

  if (history.length === 0) {
    return (
      <Center py="xl">
        <Stack align="center" gap="xs">
          <IconClockOff size={32} color="var(--mantine-color-dark-3)" />
          <Text size="sm" c="dimmed">No history yet</Text>
          <Text size="xs" c="dimmed">Completed and dismissed tasks will appear here.</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="xs">
      {history.map((action) => (
        <HistoryCard key={action.id} action={action} />
      ))}
    </Stack>
  );
};
