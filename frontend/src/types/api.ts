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
  friendship_status?: 'PENDING' | 'ACCEPTED' | 'DECLINED' | 'BLOCKED' | null;
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

export type RelationshipStatus = 'STRANGER' | 'PENDING_SENT' | 'PENDING_RECEIVED' | 'FRIEND';

export interface Friendship {
  id: string;
  requester_id: string;
  addressee_id: string;
  status: 'PENDING' | 'ACCEPTED' | 'DECLINED' | 'BLOCKED';
  created_at: string;
  updated_at: string;
}

export interface FriendRequest {
  id: string;
  sender: User;
  status: 'PENDING';
  created_at: string;
}

export type NotificationType = 'FRIEND_REQUEST' | 'FRIEND_ACCEPTED' | 'NEW_MESSAGE';

export interface Notification {
  id: string;
  recipient_id: string;
  actor_id: string;
  actor?: User; // Details of the person who triggered the notification
  type: NotificationType;
  entity_id: string; // ID of the related object (conversation_id, friendship_id, etc.)
  is_read: boolean;
  created_at: string;
}
