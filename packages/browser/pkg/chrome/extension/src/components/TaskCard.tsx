import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Flex,
  Group,
  Stack,
  Text,
  Tooltip,
} from "@mantine/core";
import {
  IconExternalLink,
  IconPlayerPlay,
  IconFocus2,
  IconX,
  IconList,
} from "@tabler/icons-react";
import type { TaskView } from "../lib/types";
import { TaskState } from "@/proto/integration/browser/v1beta/task_pb";

type Props = {
  task: TaskView;
  onStart: (taskId: string, url: string) => void;
  onFocus: (taskId: string) => void;
  onDismiss: (taskId: string) => void;
};

export const TaskCard = ({ task, onStart, onFocus, onDismiss }: Props) => {
  const isActive = task.active || task.state === TaskState.ACTIVE;
  const isExtraction = task.payload.case === "extraction";
  const hostname = (() => {
    try { return new URL(task.url).hostname; } catch { return task.url; }
  })();

  return (
    <Box
      p="sm"
      style={{
        borderRadius: "var(--mantine-radius-sm)",
        border: `1px solid ${isActive ? "var(--mantine-color-blue-7)" : "var(--mantine-color-dark-4)"}`,
        background: isActive ? "var(--mantine-color-dark-7)" : "var(--mantine-color-dark-8)",
        position: "relative",
        overflow: "hidden",
      }}
    >
      {/* Active indicator stripe */}
      {isActive && (
        <Box
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            width: 3,
            height: "100%",
            background: "var(--mantine-color-blue-5)",
            borderRadius: "var(--mantine-radius-sm) 0 0 var(--mantine-radius-sm)",
          }}
        />
      )}

      <Stack gap={6} pl={isActive ? 8 : 0}>
        {/* Header row */}
        <Flex align="center" justify="space-between" gap="xs">
          <Flex align="center" gap={6} style={{ minWidth: 0 }}>
            <Text size="sm" fw={600} truncate style={{ flex: 1 }}>
              {task.title}
            </Text>
            {isActive && (
              <Badge size="xs" color="blue" variant="light" style={{ flexShrink: 0 }}>
                Active
              </Badge>
            )}
          </Flex>

          <Tooltip label="Dismiss" position="left" withArrow>
            <ActionIcon
              size="xs"
              variant="subtle"
              color="gray"
              onClick={() => onDismiss(task.id)}
            >
              <IconX size={12} />
            </ActionIcon>
          </Tooltip>
        </Flex>

        {/* URL */}
        <Flex align="center" gap={4}>
          <IconExternalLink size={11} color="var(--mantine-color-dimmed)" style={{ flexShrink: 0 }} />
          <Text size="xs" c="dimmed" truncate style={{ flex: 1 }}>
            {hostname}
          </Text>
        </Flex>

        {/* Task type */}
        {isExtraction && (
          <Flex align="center" gap={4} style={{ minWidth: 0 }}>
            <IconList size={11} color="var(--mantine-color-dimmed)" style={{ flexShrink: 0 }} />
            <Text size="xs" c="dimmed" style={{ flexShrink: 0 }}>Data extraction</Text>
          </Flex>
        )}

        {/* Tasks */}
        <Group gap="xs" mt={4}>
          {task.active ? (
            <Button
              size="xs"
              variant="light"
              color="blue"
              leftSection={<IconFocus2 size={13} />}
              onClick={() => onFocus(task.id)}
              style={{ flex: 1 }}
            >
              Go to Tab
            </Button>
          ) : isActive ? (
            <Button
              size="xs"
              variant="light"
              color="blue"
              leftSection={<IconPlayerPlay size={13} />}
              onClick={() => onStart(task.id, task.url)}
              style={{ flex: 1 }}
            >
              Resume
            </Button>
          ) : (
            <Button
              size="xs"
              variant="light"
              color="teal"
              leftSection={<IconPlayerPlay size={13} />}
              onClick={() => onStart(task.id, task.url)}
              style={{ flex: 1 }}
            >
              Start
            </Button>
          )}
        </Group>
      </Stack>
    </Box>
  );
};
