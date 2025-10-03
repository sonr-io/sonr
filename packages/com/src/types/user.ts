/**
 * User Types
 * Types for user profiles, roles, and preferences
 */

import type { ID, Timestamp } from './common';

/**
 * User profile
 */
export interface User {
  id: ID;
  did: string;
  username?: string;
  email?: string;
  name?: string;
  avatar?: string;
  role?: UserRole;
  createdAt: Timestamp;
  lastLogin?: Timestamp;
  preferences?: UserPreferences;
}

/**
 * User role
 */
export type UserRole = 'owner' | 'admin' | 'developer' | 'viewer';

/**
 * User preferences
 */
export interface UserPreferences {
  theme?: 'light' | 'dark' | 'system';
  language?: string;
  timezone?: string;
  notifications?: NotificationPreferences;
}

/**
 * Notification preferences
 */
export interface NotificationPreferences {
  email?: boolean;
  push?: boolean;
  inApp?: boolean;
  frequency?: 'realtime' | 'daily' | 'weekly' | 'never';
}
