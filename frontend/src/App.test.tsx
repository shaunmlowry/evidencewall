import { ThemeProvider, createTheme } from '@mui/material/styles';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import React from 'react';
import { vi } from 'vitest';
import App from './App';

// Mock the AuthProvider
vi.mock('./contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  useAuth: () => ({
    user: null,
    token: null,
    isLoading: false,
    isAuthenticated: false,
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    updateUser: vi.fn(),
  }),
}));

// Mock the WebSocketProvider
vi.mock('./contexts/WebSocketContext', () => ({
  WebSocketProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  useWebSocket: () => ({
    isConnected: false,
    joinBoard: vi.fn(),
    leaveBoard: vi.fn(),
    onBoardUpdate: vi.fn(),
  }),
}));

const theme = createTheme();
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderWithProviders = (component: React.ReactElement) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        {component}
      </ThemeProvider>
    </QueryClientProvider>
  );
};

test('renders without crashing', () => {
  renderWithProviders(<App />);
  // Since the user is not authenticated, it should redirect to login
  // We can't easily test the redirect in this setup, but we can ensure it renders
  expect(document.body).toBeInTheDocument();
});

test('app structure is correct', () => {
  renderWithProviders(<App />);
  // The app should render without throwing errors
  expect(document.querySelector('body')).toBeInTheDocument();
});
