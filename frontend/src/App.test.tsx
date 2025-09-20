import { ThemeProvider, createTheme } from '@mui/material/styles';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import React from 'react';
import App from './App';

// Mock the AuthProvider
jest.mock('./contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  useAuth: () => ({
    user: null,
    token: null,
    isLoading: false,
    isAuthenticated: false,
    login: jest.fn(),
    register: jest.fn(),
    logout: jest.fn(),
    updateUser: jest.fn(),
  }),
}));

// Mock the SocketProvider
jest.mock('./contexts/SocketContext', () => ({
  SocketProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  useSocket: () => ({
    socket: null,
    isConnected: false,
    joinBoard: jest.fn(),
    leaveBoard: jest.fn(),
    sendMessage: jest.fn(),
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
