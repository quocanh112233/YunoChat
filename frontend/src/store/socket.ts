import { create } from 'zustand';
import { Message } from '@/types/api';

interface TypingIndicator {
  user_id: string;
  display_name: string;
}

interface SocketState {
  isConnected: boolean;
  isReconnecting: boolean;
  typingIndicators: Record<string, TypingIndicator[]>; // conversation_id -> users typing
  pendingMessages: Record<string, Message>; // client_temp_id -> Message
  setConnected: (connected: boolean) => void;
  setReconnecting: (reconnecting: boolean) => void;
  setTyping: (conversationId: string, userId: string, displayName: string, isTyping: boolean) => void;
  clearTyping: (conversationId: string) => void;
  addPendingMessage: (message: Message) => void;
  removePendingMessage: (tempId: string) => void;
}

export const useSocketStore = create<SocketState>((set) => ({
  isConnected: false,
  isReconnecting: false,
  typingIndicators: {},
  pendingMessages: {},
  setConnected: (connected) => set({ isConnected: connected }),
  setReconnecting: (reconnecting) => set({ isReconnecting: reconnecting }),
  setTyping: (conversationId, userId, displayName, isTyping) =>
    set((state) => {
      const currentTyping = state.typingIndicators[conversationId] || [];
      if (isTyping) {
        if (currentTyping.some((t) => t.user_id === userId)) return state;
        return {
          typingIndicators: {
            ...state.typingIndicators,
            [conversationId]: [...currentTyping, { user_id: userId, display_name: displayName }],
          },
        };
      } else {
        return {
          typingIndicators: {
            ...state.typingIndicators,
            [conversationId]: currentTyping.filter((t) => t.user_id !== userId),
          },
        };
      }
    }),
  clearTyping: (conversationId) =>
    set((state) => {
      const newTyping = { ...state.typingIndicators };
      delete newTyping[conversationId];
      return { typingIndicators: newTyping };
    }),
  addPendingMessage: (message) =>
    set((state) => ({
      pendingMessages: { ...state.pendingMessages, [message.client_temp_id || message.id]: message },
    })),
  removePendingMessage: (tempId) =>
    set((state) => {
      const newPending = { ...state.pendingMessages };
      delete newPending[tempId];
      return { pendingMessages: newPending };
    }),
}));