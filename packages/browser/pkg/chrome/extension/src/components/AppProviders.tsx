import { MantineProvider } from '@mantine/core';

export const AppProviders = ({ children }: { children: React.ReactNode }) => {
  return <MantineProvider defaultColorScheme="dark">{children}</MantineProvider>;
};
