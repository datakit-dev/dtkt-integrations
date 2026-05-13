import {
  Anchor,
  Badge,
  Card,
  Code,
  Container,
  Divider,
  Flex,
  Group,
  List,
  Stack,
  Stepper,
  Text,
  ThemeIcon,
  Title,
} from '@mantine/core';
import {
  IconBrandChrome,
  IconCircleCheckFilled,
  IconCode,
  IconPlug,
  IconTerminal2,
} from '@tabler/icons-react';
import { AppProviders } from '@/components/AppProviders';

const App = () => {
  return (
    <AppProviders>
      <Container size="sm" py="xl">
        <Stack gap="xl">

          {/* Header */}
          <Stack gap={4}>
            <Flex align="center" gap="sm">
              <IconBrandChrome size={36} color="var(--mantine-color-blue-5)" />
              <Title order={1}>DataKit Browser</Title>
            </Flex>
            <Text c="dimmed">
              Your browser is now a secure, programmable runtime. Follow the steps below
              to connect it to the DataKit platform.
            </Text>
          </Stack>

          {/* What this is */}
          <Card withBorder padding="lg" radius="md">
            <Stack gap="sm">
              <Flex align="center" gap="xs">
                <Title order={3}>What is DataKit Browser?</Title>
                <Badge color="blue" variant="light" size="sm">v1beta</Badge>
              </Flex>
              <Text size="sm">
                DataKit Browser is a Chrome extension that connects your browser to the DataKit
                platform via a secure native messaging bridge. Once connected, DataKit flows can
                observe browser activity, extract structured data from any page, trigger
                navigations, show notifications, and run scripts.
              </Text>
            </Stack>
          </Card>

          {/* Setup steps */}
          <Stack gap="sm">
            <Title order={3}>Setup</Title>
            <Stepper orientation="vertical" active={-1} color="blue">

              <Stepper.Step
                label="Install the DataKit CLI"
                description={
                  <Stack gap={4} mt={4} mb="md">
                    <Text size="sm" c="dimmed">
                      The CLI ships the native messaging host that this extension communicates with.
                    </Text>
                    <Code block fz="xs">
                      {`# macOS / Linux\nbrew install datakit-dev/tap/dtkt\n\n# Or via Go\ngo install github.com/datakit-dev/dtkt-cli@latest`}
                    </Code>
                  </Stack>
                }
                icon={<ThemeIcon radius="xl" size="md" color="blue" variant="light"><IconTerminal2 size={16} /></ThemeIcon>}
              />

              <Stepper.Step
                label="Install the native messaging bridge"
                description={
                  <Stack gap={4} mt={4} mb="md">
                    <Text size="sm" c="dimmed">
                      This registers the native host so Chrome can launch it. Run once per machine.
                    </Text>
                    <Code block fz="xs">
                      {`dtkt browser install --ext-id <your-extension-id>`}
                    </Code>
                    <Text size="xs" c="dimmed">
                      Your extension ID is shown in{' '}
                      <Anchor href="chrome://extensions" size="xs">chrome://extensions</Anchor>
                      {' '}with Developer Mode enabled.
                    </Text>
                  </Stack>
                }
                icon={<ThemeIcon radius="xl" size="md" color="blue" variant="light"><IconPlug size={16} /></ThemeIcon>}
              />

              <Stepper.Step
                label="Connect from DataKit"
                description={
                  <Stack gap={4} mt={4} mb="md">
                    <Text size="sm" c="dimmed">
                      Add a Browser connection to any DataKit flow. The extension will connect
                      automatically the next time a browser event occurs.
                    </Text>
                  </Stack>
                }
                icon={<ThemeIcon radius="xl" size="md" color="blue" variant="light"><IconCode size={16} /></ThemeIcon>}
              />

              <Stepper.Step
                label="You're ready"
                description={
                  <Stack gap={4} mt={4} mb="md">
                    <Flex align="center" gap={6}>
                      <IconCircleCheckFilled size={16} color="var(--mantine-color-green-6)" />
                      <Text size="sm">The extension is installed and active.</Text>
                    </Flex>
                  </Stack>
                }
                icon={<ThemeIcon radius="xl" size="md" color="green" variant="light"><IconCircleCheckFilled size={16} /></ThemeIcon>}
              />

            </Stepper>
          </Stack>

          <Divider />

          <Flex justify="space-between" align="center">
            <Text size="xs" c="dimmed">You can close this tab — the extension is already active.</Text>
            <Anchor
              href="https://datakit.dev/docs/integrations/browser"
              target="_blank"
              rel="noopener noreferrer"
              size="xs"
            >
              Documentation &rarr;
            </Anchor>
          </Flex>

        </Stack>
      </Container>
    </AppProviders>
  );
};

export { App };
