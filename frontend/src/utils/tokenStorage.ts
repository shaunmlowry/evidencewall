/**
 * Secure token storage utilities
 * Provides secure storage and retrieval of authentication tokens
 */

const TOKEN_KEY = 'auth_token';
const TOKEN_EXPIRY_KEY = 'auth_token_expiry';

/**
 * Store authentication token with expiry
 * @param token JWT token
 * @param expiryTime Expiry time in milliseconds
 */
export const storeToken = (token: string, expiryTime?: number): void => {
  try {
    // Store token in localStorage (consider using httpOnly cookies in production)
    localStorage.setItem(TOKEN_KEY, token);
    
    // Store expiry time if provided
    if (expiryTime) {
      localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
    }
  } catch (error) {
    console.error('Failed to store token:', error);
  }
};

/**
 * Retrieve authentication token
 * @returns JWT token or null if not found/expired
 */
export const getToken = (): string | null => {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);
    
    // Check if token is expired
    if (token && expiryTime) {
      const now = Date.now();
      const expiry = parseInt(expiryTime, 10);
      
      if (now >= expiry) {
        // Token is expired, remove it
        removeToken();
        return null;
      }
    }
    
    return token;
  } catch (error) {
    console.error('Failed to retrieve token:', error);
    return null;
  }
};

/**
 * Remove authentication token
 */
export const removeToken = (): void => {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(TOKEN_EXPIRY_KEY);
  } catch (error) {
    console.error('Failed to remove token:', error);
  }
};

/**
 * Check if token exists and is valid
 * @returns true if token exists and is not expired
 */
export const hasValidToken = (): boolean => {
  return getToken() !== null;
};

/**
 * Get token expiry time
 * @returns Expiry time in milliseconds or null
 */
export const getTokenExpiry = (): number | null => {
  try {
    const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);
    return expiryTime ? parseInt(expiryTime, 10) : null;
  } catch (error) {
    console.error('Failed to get token expiry:', error);
    return null;
  }
};

/**
 * Check if token will expire soon (within 5 minutes)
 * @returns true if token expires within 5 minutes
 */
export const isTokenExpiringSoon = (): boolean => {
  const expiryTime = getTokenExpiry();
  if (!expiryTime) return false;
  
  const now = Date.now();
  const fiveMinutes = 5 * 60 * 1000; // 5 minutes in milliseconds
  
  return (expiryTime - now) <= fiveMinutes;
};

/**
 * Parse JWT token to extract payload (without verification)
 * @param token JWT token
 * @returns Decoded payload or null if invalid
 */
export const parseTokenPayload = (token: string): any | null => {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded);
  } catch (error) {
    console.error('Failed to parse token payload:', error);
    return null;
  }
};

/**
 * Get token expiry from JWT payload
 * @param token JWT token
 * @returns Expiry time in milliseconds or null
 */
export const getTokenExpiryFromPayload = (token: string): number | null => {
  const payload = parseTokenPayload(token);
  if (!payload || !payload.exp) return null;
  
  // JWT exp is in seconds, convert to milliseconds
  return payload.exp * 1000;
};
