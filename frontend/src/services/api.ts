import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
    AuthResponse,
    Board,
    BoardConnection,
    BoardItem,
    CreateBoardConnectionRequest,
    CreateBoardItemRequest,
    CreateBoardRequest,
    LoginRequest,
    PaginatedResponse,
    RegisterRequest,
    ShareBoardRequest,
    UpdateBoardItemRequest,
    UpdateBoardRequest,
    User,
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
    if (import.meta.env.MODE !== 'production') {
      // Verbose client-side request log
      // eslint-disable-next-line no-console
      console.debug('[api:req]', config.method?.toUpperCase(), config.baseURL, config.url, {
        params: config.params,
        hasAuth: !!token,
      });
    }
    return config;
  });

  // Handle auth errors
  instance.interceptors.response.use(
    (response) => {
      if (import.meta.env.MODE !== 'production') {
        // eslint-disable-next-line no-console
        console.debug('[api:res]', response.status, response.config.url);
      }
      return response;
    },
    (error) => {
      if (import.meta.env.MODE !== 'production') {
        // eslint-disable-next-line no-console
        console.debug('[api:err]', {
          url: error.config?.url,
          method: error.config?.method,
          status: error.response?.status,
          data: error.response?.data,
          message: error.message,
        });
      }
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
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8001'
);

const boardsApiInstance = createApiInstance(
  import.meta.env.VITE_BOARDS_API_URL || 'http://localhost:8002'
);

// Auth API
export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/v1/auth/login',
      data
    );
    return response.data;
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/v1/auth/register',
      data
    );
    return response.data;
  },

  refreshToken: async (token: string): Promise<AuthResponse> => {
    const response: AxiosResponse<AuthResponse> = await authApiInstance.post(
      '/v1/auth/refresh',
      { token }
    );
    return response.data;
  },

  getProfile: async (): Promise<User> => {
    const response: AxiosResponse<User> = await authApiInstance.get('/v1/me');
    return response.data;
  },

  updateProfile: async (data: Partial<User>): Promise<User> => {
    const response: AxiosResponse<User> = await authApiInstance.put('/v1/me', data);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await authApiInstance.post('/v1/logout');
  },

  getGoogleLoginUrl: async (): Promise<{ url: string; state: string }> => {
    const response = await authApiInstance.get('/v1/auth/google');
    return response.data;
  },
};

// Boards API
export const boardsApi = {
  // Board CRUD
  getBoards: async (page = 1, limit = 20): Promise<PaginatedResponse<Board>> => {
    const response: AxiosResponse<any> = await boardsApiInstance.get(
      `/v1/boards?page=${page}&limit=${limit}`
    );
    const data = response.data;
    // Normalize backend response { boards, total, page, limit } -> { data, total, page, limit, has_more }
    if (data && Array.isArray(data.data)) {
      return data as PaginatedResponse<Board>;
    }
    if (data && Array.isArray(data.boards)) {
      const total = typeof data.total === 'number' ? data.total : data.boards.length;
      const currentPage = typeof data.page === 'number' ? data.page : page;
      const currentLimit = typeof data.limit === 'number' ? data.limit : limit;
      return {
        data: data.boards,
        total,
        page: currentPage,
        limit: currentLimit,
        has_more: currentPage * currentLimit < total,
      } as PaginatedResponse<Board>;
    }
    return { data: [], total: 0, page, limit, has_more: false } as PaginatedResponse<Board>;
  },

  getBoard: async (id: string): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.get(`/v1/boards/${id}`);
    return response.data;
  },

  getPublicBoard: async (id: string): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.get(
      `/v1/public/boards/${id}`
    );
    return response.data;
  },

  createBoard: async (data: CreateBoardRequest): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.post('/v1/boards', data);
    return response.data;
  },

  updateBoard: async (id: string, data: UpdateBoardRequest): Promise<Board> => {
    const response: AxiosResponse<Board> = await boardsApiInstance.put(
      `/v1/boards/${id}`,
      data
    );
    return response.data;
  },

  deleteBoard: async (id: string): Promise<void> => {
    await boardsApiInstance.delete(`/v1/boards/${id}`);
  },

  // Board sharing
  shareBoard: async (id: string, data: ShareBoardRequest): Promise<void> => {
    await boardsApiInstance.post(`/v1/boards/${id}/share`, data);
  },

  unshareBoard: async (boardId: string, userId: string): Promise<void> => {
    await boardsApiInstance.delete(`/v1/boards/${boardId}/share/${userId}`);
  },

  updateUserPermission: async (
    boardId: string,
    userId: string,
    permission: string
  ): Promise<void> => {
    await boardsApiInstance.put(`/v1/boards/${boardId}/users/${userId}/permission`, {
      permission,
    });
  },

  // Board items
  getBoardItems: async (boardId: string): Promise<BoardItem[]> => {
    const response: AxiosResponse<BoardItem[]> = await boardsApiInstance.get(
      `/v1/boards/${boardId}/items`
    );
    return response.data;
  },

  createBoardItem: async (boardId: string, data: CreateBoardItemRequest): Promise<BoardItem> => {
    const response: AxiosResponse<BoardItem> = await boardsApiInstance.post(
      `/v1/boards/${boardId}/items`,
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
      `/v1/boards/${boardId}/items/${itemId}`,
      data
    );
    return response.data;
  },

  deleteBoardItem: async (boardId: string, itemId: string): Promise<void> => {
    await boardsApiInstance.delete(`/v1/boards/${boardId}/items/${itemId}`);
  },

  // Board connections
  getBoardConnections: async (boardId: string): Promise<BoardConnection[]> => {
    const response: AxiosResponse<BoardConnection[]> = await boardsApiInstance.get(
      `/v1/boards/${boardId}/connections`
    );
    return response.data;
  },

  createBoardConnection: async (
    boardId: string,
    data: CreateBoardConnectionRequest
  ): Promise<BoardConnection> => {
    const response: AxiosResponse<BoardConnection> = await boardsApiInstance.post(
      `/v1/boards/${boardId}/connections`,
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
      `/v1/boards/${boardId}/connections/${connectionId}`,
      data
    );
    return response.data;
  },

  deleteBoardConnection: async (boardId: string, connectionId: string): Promise<void> => {
    await boardsApiInstance.delete(`/v1/boards/${boardId}/connections/${connectionId}`);
  },
};

export default {
  auth: authApi,
  boards: boardsApi,
};


