export const testUsers = {
  validUser: {
    email: 'test@example.com',
    password: 'TestPassword123!',
    name: 'Test User'
  },
  adminUser: {
    email: 'admin@example.com',
    password: 'AdminPassword123!',
    name: 'Admin User'
  },
  collaborator: {
    email: 'collaborator@example.com',
    password: 'CollabPassword123!',
    name: 'Collaborator User'
  }
};

export const testBoards = {
  publicBoard: {
    title: 'Public Investigation Board',
    description: 'A public board for testing',
    isPublic: true
  },
  privateBoard: {
    title: 'Private Investigation Board',
    description: 'A private board for testing',
    isPublic: false
  },
  collaborativeBoard: {
    title: 'Collaborative Investigation Board',
    description: 'A board for testing collaboration features',
    isPublic: false
  }
};

export const testBoardItems = {
  postItNote: {
    type: 'post-it',
    content: 'Important evidence found at scene',
    color: '#ffeb3b',
    position: { x: 100, y: 100 }
  },
  suspectCard: {
    type: 'suspect',
    content: 'John Doe - Primary suspect',
    color: '#ffcdd2',
    position: { x: 300, y: 150 }
  },
  evidenceItem: {
    type: 'evidence',
    content: 'DNA sample collected',
    color: '#c8e6c9',
    position: { x: 200, y: 250 }
  }
};

export const apiEndpoints = {
  auth: {
    register: '/api/v1/auth/register',
    login: '/api/v1/auth/login',
    profile: '/api/v1/me'
  },
  boards: {
    list: '/api/v1/boards',
    create: '/api/v1/boards',
    get: (id: string) => `/api/v1/boards/${id}`,
    update: (id: string) => `/api/v1/boards/${id}`,
    delete: (id: string) => `/api/v1/boards/${id}`,
    share: (id: string) => `/api/v1/boards/${id}/share`,
    items: (boardId: string) => `/api/v1/boards/${boardId}/items`,
    public: (id: string) => `/api/v1/public/boards/${id}`
  }
};
