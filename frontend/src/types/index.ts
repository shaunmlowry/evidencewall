// User types
export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  verified: boolean;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

// Board types
export type PermissionLevel = 'read' | 'read_write' | 'admin';
export type BoardVisibility = 'private' | 'shared' | 'public';

export interface Board {
  id: string;
  title: string;
  description: string;
  visibility: BoardVisibility;
  owner_id: string;
  permission?: PermissionLevel;
  created_at: string;
  updated_at: string;
  items?: BoardItem[];
  connections?: BoardConnection[];
  users?: BoardUserResponse[];
}

export interface BoardUserResponse {
  user: User;
  permission: PermissionLevel;
  created_at: string;
}

export interface BoardItem {
  id: string;
  board_id: string;
  type: 'post-it' | 'suspect-card';
  x: number;
  y: number;
  width: number;
  height: number;
  rotation: number;
  z_index: number;
  content: string;
  style: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BoardConnection {
  id: string;
  board_id: string;
  from_item_id: string;
  to_item_id: string;
  style: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// API request types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  name: string;
  password: string;
}

export interface CreateBoardRequest {
  title: string;
  description?: string;
  visibility?: BoardVisibility;
}

export interface UpdateBoardRequest {
  title?: string;
  description?: string;
  visibility?: BoardVisibility;
}

export interface CreateBoardItemRequest {
  type: 'post-it' | 'suspect-card';
  x: number;
  y: number;
  width?: number;
  height?: number;
  rotation?: number;
  content?: string;
  style?: string;
}

export interface UpdateBoardItemRequest {
  x?: number;
  y?: number;
  width?: number;
  height?: number;
  rotation?: number;
  content?: string;
  style?: string;
  z_index?: number;
}

export interface CreateBoardConnectionRequest {
  from_item_id: string;
  to_item_id: string;
  style?: string;
}

export interface ShareBoardRequest {
  user_email: string;
  permission: PermissionLevel;
}

// WebSocket message types
export interface SocketMessage {
  type: string;
  payload: any;
  board_id?: string;
  user_id?: string;
  timestamp: string;
}

export interface ItemUpdateMessage {
  type: 'item_update';
  payload: {
    item: BoardItem;
    action: 'create' | 'update' | 'delete';
  };
  board_id: string;
  user_id: string;
  timestamp: string;
}

export interface ConnectionUpdateMessage {
  type: 'connection_update';
  payload: {
    connection: BoardConnection;
    action: 'create' | 'update' | 'delete';
  };
  board_id: string;
  user_id: string;
  timestamp: string;
}

export interface UserCursorMessage {
  type: 'user_cursor';
  payload: {
    x: number;
    y: number;
    user: User;
  };
  board_id: string;
  user_id: string;
  timestamp: string;
}

// API response types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
}

// Component prop types
export interface BoardItemProps {
  item: BoardItem;
  isSelected: boolean;
  isConnecting: boolean;
  onSelect: (item: BoardItem) => void;
  onUpdate: (item: BoardItem, updates: UpdateBoardItemRequest) => void;
  onDelete: (item: BoardItem) => void;
  onConnect: (item: BoardItem) => void;
}

export interface BoardConnectionProps {
  connection: BoardConnection;
  fromItem: BoardItem;
  toItem: BoardItem;
  onUpdate: (connection: BoardConnection, updates: any) => void;
  onDelete: (connection: BoardConnection) => void;
}

// Utility types
export interface Position {
  x: number;
  y: number;
}

export interface Size {
  width: number;
  height: number;
}

export interface Bounds {
  x: number;
  y: number;
  width: number;
  height: number;
}


