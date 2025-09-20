import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
  Board,
  BoardItem,
  BoardConnection,
  CreateBoardRequest,
  UpdateBoardRequest,
  CreateBoardItemRequest,
  UpdateBoardItemRequest,
  CreateBoardConnectionRequest,
  ShareBoardRequest,
  PaginatedResponse,
} from '../types';

// Create axios instance
const createApiInstance = (baseURL: string): AxiosInstance => {
  const instance = axios.create({
    baseURL,
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Add auth token to requests
  instance.interceptors.request.use((config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  // Handle auth errors
  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
      }
      return Promise.reject(error);
    }
  );

  return instance;
};

// API instances
const authApiInstance = createApiInstance(
  process.env.REACT_APP_API_BASE_URL || 'http://localhost:8001'
);

const boardsApiInstance = createApiInstance(
  process.env.REACT_APP_BOARDS_API_URL || 'http://localhost:8002'
);

// Auth API
export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/api/v1/auth/login',
      data
    );
    return response.data;
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/api/v1/auth/register',
      data
    );
    return response.data;
  },

  refreshToken: async (token: string): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/api/v1/auth/refresh',
      { token }
    );
    return response.data;
  },

  getProfile: async (): Promise<User> => {
    const response: AxiosResponse<User> = await authApiInstance.get('/api/v1/me');
    return response.data;
  },

  updateProfile: async (data: Partial<User>): Promise<User> => {
    const response: AxiosResponse<User> = await authApiInstance.put('/api/v1/me', data);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await authApiInstance.post('/api/v1/logout');
  },

  getGoogleLoginUrl: async (): Promise<{ url: string; state: string }> => {
    const response = await authApiInstance.get('/api/v1/auth/google');
    return response.data;
  },
};

// Boards API
export const boardsApi = {
  // Board CRUD
  getBoards: async (page = 1, limit = 20): Promise<PaginatedResponse<Board>> => {
    const response: AxiosResponse<PaginatedResponse<Board>> = await boardsApiInstance.get(
      `/api/v1/boards?page=${page}&limit=${limit}`
    );
    return response.data;
  },

  getBoard: async (id: string): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.get(`/api/v1/boards/${id}`);
    return response.data;
  },

  getPublicBoard: async (id: string): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.get(
      `/api/v1/public/boards/${id}`
    );
    return response.data;
  },

  createBoard: async (data: CreateBoardRequest): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.post('/api/v1/boards', data);
    return response.data;
  },

  updateBoard: async (id: string, data: UpdateBoardRequest): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.put(
      `/api/v1/boards/${id}`,
      data
    );
    return response.data;
  },

  deleteBoard: async (id: string): Promise<void> => {
    await boardsApiInstance.delete(`/api/v1/boards/${id}`);
  },

  // Board sharing
  shareBoard: async (id: string, data: ShareBoardRequest): Promise<void> => {
    await boardsApiInstance.post(`/api/v1/boards/${id}/share`, data);
  },

  unshareBoard: async (boardId: string, userId: string): Promise<void> => {
    await boardsApiInstance.delete(`/api/v1/boards/${boardId}/share/${userId}`);
  },

  updateUserPermission: async (
    boardId: string,
    userId: string,
    permission: string
  ): Promise<void> => {
    await boardsApiInstance.put(`/api/v1/boards/${boardId}/users/${userId}/permission`, {
      permission,
    });
  },

  // Board items
  getBoardItems: async (boardId: string): Promise<BoardItem[]> => {
    const response: AxiosResponse<BoardItem[]> = await boardsApiInstance.get(
      `/api/v1/boards/${boardId}/items`
    );
    return response.data;
  },

  createBoardItem: async (boardId: string, data: CreateBoardItemRequest): Promise<BoardItem> => {
    const response: AxiosResponse<BoardItem> = await boardsApiInstance.post(
      `/api/v1/boards/${boardId}/items`,
      data
    );
    return response.data;
  },

  updateBoardItem: async (
    boardId: string,
    itemId: string,
    data: UpdateBoardItemRequest
  ): Promise<BoardItem> => {
    const response: AxiosResponse<BoardItem> = await boardsApiInstance.put(
      `/api/v1/boards/${boardId}/items/${itemId}`,
      data
    );
    return response.data;
  },

  deleteBoardItem: async (boardId: string, itemId: string): Promise<void> => {
    await boardsApiInstance.delete(`/api/v1/boards/${boardId}/items/${itemId}`);
  },

  // Board connections
  getBoardConnections: async (boardId: string): Promise<BoardConnection[]> => {
    const response: AxiosResponse<BoardConnection[]> = await boardsApiInstance.get(
      `/api/v1/boards/${boardId}/connections`
    );
    return response.data;
  },

  createBoardConnection: async (
    boardId: string,
    data: CreateBoardConnectionRequest
  ): Promise<BoardConnection> => {
    const response: AxiosResponse<BoardConnection> = await boardsApiInstance.post(
      `/api/v1/boards/${boardId}/connections`,
      data
    );
    return response.data;
  },

  updateBoardConnection: async (
    boardId: string,
    connectionId: string,
    data: any
  ): Promise<BoardConnection> => {
    const response: AxiosResponse<BoardConnection> = await boardsApiInstance.put(
      `/api/v1/boards/${boardId}/connections/${connectionId}`,
      data
    );
    return response.data;
  },

  deleteBoardConnection: async (boardId: string, connectionId: string): Promise<void> => {
    await boardsApiInstance.delete(`/api/v1/boards/${boardId}/connections/${connectionId}`);
  },
};

export default {
  auth: authApi,
  boards: boardsApi,
};


