import type { ReactElement } from 'react';
import { render, type RenderOptions } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { vi } from 'vitest';

// Create a test theme (light mode by default)
const testTheme = createTheme({
  colorSchemes: {
    light: true,
    dark: true,
  },
});

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  initialRoute?: string;
}

/**
 * Custom render function that wraps components with necessary providers
 * (Theme, Router, etc.)
 */
export function renderWithProviders(
  ui: ReactElement,
  { initialRoute = '/', ...renderOptions }: CustomRenderOptions = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <BrowserRouter>
        <ThemeProvider theme={testTheme}>
          <CssBaseline />
          {children}
        </ThemeProvider>
      </BrowserRouter>
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}

/**
 * Mock API client for testing
 */
export const mockApiClient = {
  getConfig: vi.fn(),
  updateConfig: vi.fn(),
  getServers: vi.fn(),
  addServer: vi.fn(),
  updateServer: vi.fn(),
  deleteServer: vi.fn(),
  testServer: vi.fn(),
  getLogs: vi.fn(),
  clearLogs: vi.fn(),
  exportLogs: vi.fn(),
  triggerAutomation: vi.fn(),
  getAutomationStatus: vi.fn(),
  getStats: vi.fn(),
  getServerStats: vi.fn(),
};

/**
 * Mock WebSocket client for testing
 */
export class MockWebSocketClient {
  public onMessage = vi.fn();
  public onStatusChange = vi.fn();

  connect = vi.fn();
  disconnect = vi.fn();
  send = vi.fn();

  // Helper to simulate receiving a message
  simulateMessage(data: any) {
    this.onMessage(data);
  }

  // Helper to simulate status change
  simulateStatusChange(connected: boolean) {
    this.onStatusChange(connected);
  }
}

// Re-export everything from React Testing Library
export * from '@testing-library/react';
export { default as userEvent } from '@testing-library/user-event';
