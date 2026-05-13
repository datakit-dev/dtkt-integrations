import { Badge, Box, Flex, Stack, Text } from "@mantine/core";
import { IconCircleCheckFilled, IconCircleDashedX, IconExternalLink } from "@tabler/icons-react";
import { tsToMs } from "../lib/types";
import type { Task } from "../lib/types";
import { TaskState } from "../proto/integration/browser/v1beta/task_pb";

type Props = {
  action: Task;
};

function timeAgo(ms: number): string {
  const diff = Date.now() - ms;
  const m = Math.floor(diff / 60_000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

export const HistoryCard = ({ action }: Props) => {
  const isCompleted = action.state === TaskState.COMPLETED;
  const extraction = action.payload.case === "extraction" ? action.payload.value : null;
  const hostname = (() => {
    try { return new URL(action.url).hostname; } catch { return action.url; }
  })();
  const ts = tsToMs(action.completeTime ?? action.createTime);
  const hasExtraction = extraction !== null;

  return (
    <Box
      p="sm"
      style={{
        borderRadius: "var(--mantine-radius-sm)",
        border: "1px solid var(--mantine-color-dark-5)",
        background: "var(--mantine-color-dark-9)",
        opacity: 0.8,
      }}
    >
      <Stack gap={4}>
        {/* Header row */}
        <Flex align="center" gap="xs" justify="space-between">
          <Flex align="center" gap={6} style={{ minWidth: 0 }}>
            {isCompleted ? (
              <IconCircleCheckFilled size={14} color="var(--mantine-color-teal-5)" style={{ flexShrink: 0 }} />
            ) : (
              <IconCircleDashedX size={14} color="var(--mantine-color-gray-6)" style={{ flexShrink: 0 }} />
            )}
            <Text size="sm" fw={500} truncate c={isCompleted ? undefined : "dimmed"}>
              {action.title}
            </Text>
          </Flex>
          <Text size="xs" c="dimmed" style={{ flexShrink: 0 }}>
            {timeAgo(ts)}
          </Text>
        </Flex>

        {/* URL */}
        <Flex align="center" gap={4} pl={20}>
          <IconExternalLink size={11} color="var(--mantine-color-dimmed)" style={{ flexShrink: 0 }} />
          <Text size="xs" c="dimmed" truncate>
            {hostname}
          </Text>
        </Flex>

        {/* Badge */}
        {isCompleted && hasExtraction && (
          <Flex pl={20} mt={2}>
            <Badge size="xs" color="teal" variant="dot">
              Data extraction completed
            </Badge>
          </Flex>
        )}
      </Stack>
    </Box>
  );
};
