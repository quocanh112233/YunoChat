export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

export type UserStatus = 'ONLINE' | 'OFFLINE' | 'AWAY' | 'BUSY';

export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url?: string;
  status: UserStatus;
  last_seen_at?: string;
  created_at: string;
}

export interface Conversation {
  id: string;
  type: 'DM' | 'GROUP';
  name?: string;
  avatar_url?: string;
  last_message_id?: string;
  last_activity_at: string;
  member_count?: number; // Added for group info
  created_at: string;
}

export type MessageStatus = 'SENDING' | 'SENT' | 'DELIVERED' | 'READ';

export interface Message {
  id: string;
  client_temp_id?: string; // For optimistic UI tracking
  conversation_id: string;
  sender_id: string;
  body?: string;
  type: 'TEXT' | 'FILE' | 'IMAGE';
  status: MessageStatus;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface Notification {
  id: string;
  recipient_id: string;
  actor_id: string;
  type: string;
  entity_id: string;
  is_read: boolean;
  created_at: string;
}
