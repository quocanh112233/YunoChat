'use client';

import React, { createContext, useContext, useEffect, useRef, useCallback } from 'react';
import { Message, MessageStatus, Notification } from '@/types/api';
import { useSocketStore } from '@/store/socket';
import { useNotificationStore } from '@/store/notification';
import { useAuthStore } from '@/store/auth';
import { useQueryClient, InfiniteData } from '@tanstack/react-query';
import { toast } from 'sonner';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/v1/ws';
const RECONNECT_INITIAL_DELAY = 1000;
const RECONNECT_MAX_DELAY = 30000;

interface WebSocketContextType {
  sendMessage: (data: { type: string; payload: unknown }) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocketContext = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider = ({ children }: { children: React.ReactNode }) => {
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectDelayRef = useRef(RECONNECT_INITIAL_DELAY);
  const { accessToken, isAuthenticated, user } = useAuthStore();
  const { setConnected, setReconnecting, setTyping, removePendingMessage } = useSocketStore();
  const queryClient = useQueryClient();

  // Use a ref for handleWsEvent to avoid it being a dependency of connect
  const handleWsEventRef = useRef<((event: unknown) => void) | null>(null);

  const handleWsEvent = useCallback((event: unknown) => {
    const { type, payload } = event as { type: string; payload: Record<string, unknown> };

    switch (type) {
      case 'new_message':
        const message = payload as unknown as Message;
        // Skip if sender is current user (handled by optimistic UI)
        if (message.sender_id === user?.id) break;

        // Direct Cache Update for new message
        queryClient.setQueryData<InfiniteData<Message[]>>(['messages', message.conversation_id], (old) => {
          if (!old) return old;
          const newPages = [...old.pages];
          if (newPages.length > 0) {
            newPages[0] = [message, ...newPages[0]];
          } else {
            newPages[0] = [message];
          }
          return { ...old, pages: newPages };
        });

        queryClient.invalidateQueries({ queryKey: ['conversations'] });
        break;

      case 'message_sent':
        const sentPayload = payload as { client_temp_id: string; message_id: string; status: string; conversation_id?: string };
        if (sentPayload.client_temp_id) {
          removePendingMessage(sentPayload.client_temp_id);

          // Direct Cache Update for message sent ack
          const convId = sentPayload.conversation_id;
          if (convId) {
            queryClient.setQueryData<InfiniteData<Message[]>>(['messages', convId], (old) => {
              if (!old) return old;
              const newPages = old.pages.map((page: Message[]) => 
                page.map(m => m.client_temp_id === sentPayload.client_temp_id 
                  ? { ...m, id: sentPayload.message_id, status: sentPayload.status as MessageStatus } 
                  : m
                )
              );
              return { ...old, pages: newPages };
            });
          } else {
            // Fallback: invalidate if convId is missing in payload
            queryClient.invalidateQueries({ queryKey: ['messages'] });
          }
        }
        queryClient.invalidateQueries({ queryKey: ['conversations'] });
        break;

      case 'user_typing':
        const typingPayload = payload as { conversation_id: string; user_id: string; display_name: string; is_typing: boolean };
        setTyping(
          typingPayload.conversation_id,
          typingPayload.user_id,
          typingPayload.display_name,
          typingPayload.is_typing
        );
        break;

      case 'presence_update':
        queryClient.invalidateQueries({ queryKey: ['users'] });
        queryClient.invalidateQueries({ queryKey: ['conversations'] });
        break;

      case 'notification_new':
        useNotificationStore.getState().addNotification(payload as unknown as Notification);
        break;

      case 'member_removed': {
        // When we are kicked or any member is removed, invalidate conversation list
        // so the sidebar reflects the change automatically
        const removedPayload = payload as { conversation_id: string; removed_user_id: string };
        queryClient.invalidateQueries({ queryKey: ['conversations'] });
        queryClient.invalidateQueries({ queryKey: ['messages', removedPayload.conversation_id] });

        // If current user is the one removed, redirect to home and show toast
        if (removedPayload.removed_user_id === user?.id) {
          toast.error('Bạn đã bị xóa khỏi cuộc trò chuyện');
          window.location.href = '/conversations';
        }
        break;
      }

      case 'error': {
        // Handle server-sent WS errors (e.g. FORBIDDEN when unfriended)
        const errPayload = payload as { code: string; message: string; ref_event?: string };
        if (errPayload.code === 'FORBIDDEN' && errPayload.ref_event === 'send_message') {
          toast.error('Bạn không thể gửi tin nhắn cho người này');
        } else {
          console.warn('WS error event:', errPayload);
        }
        break;
      }

      default:
        console.warn('Unhandled WS event type:', type);
    }
  }, [queryClient, setTyping, user?.id, removePendingMessage]);


  const connectRef = useRef<() => void | undefined>(undefined);

  const messageQueueRef = useRef<{ type: string; payload: unknown }[]>([]);

  const connect = useCallback(() => {
    if (!accessToken || !isAuthenticated) return;
    if (socketRef.current?.readyState === WebSocket.OPEN) return;

    console.log('Attempting to connect to WebSocket...');
    const url = `${WS_URL}?token=${accessToken}`;
    try {
      const socket = new WebSocket(url);

      socket.onopen = () => {
        console.log('WebSocket connected');
        setConnected(true);
        setReconnecting(false);
        reconnectDelayRef.current = RECONNECT_INITIAL_DELAY;
        queryClient.invalidateQueries({ queryKey: ['conversations'] });

        // Flush message queue
        if (messageQueueRef.current.length > 0) {
          console.log(`Flushing ${messageQueueRef.current.length} queued messages...`);
          messageQueueRef.current.forEach(msg => {
            socket.send(JSON.stringify(msg));
          });
          messageQueueRef.current = [];
        }
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (handleWsEventRef.current) {
            handleWsEventRef.current(data);
          }
        } catch (err) {
          console.error('Failed to parse WS message:', err);
        }
      };

      socket.onclose = () => {
        console.log('WebSocket disconnected');
        setConnected(false);
        socketRef.current = null;

        const nextDelay = reconnectDelayRef.current;
        reconnectDelayRef.current = Math.min(reconnectDelayRef.current * 2, RECONNECT_MAX_DELAY);

        if (isAuthenticated) {
          setReconnecting(true);
          console.log(`Reconnecting in ${nextDelay}ms...`);
          setTimeout(() => {
            if (useAuthStore.getState().isAuthenticated && connectRef.current) {
              connectRef.current();
            }
          }, nextDelay);
        }
      };

      socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        socket.close();
      };

      socketRef.current = socket;
    } catch (err) {
      console.error('WebSocket connection error:', err);
      setReconnecting(true);
    }
  }, [accessToken, isAuthenticated, setConnected, setReconnecting, queryClient]);

  useEffect(() => {
    handleWsEventRef.current = handleWsEvent;
  }, [handleWsEvent]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  useEffect(() => {
    connect();
    return () => {
      socketRef.current?.close();
    };
  }, [connect]);

  const sendMessage = useCallback((data: { type: string; payload: unknown }) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(data));
    } else {
      console.log('WS not connected, queuing message:', data.type);
      messageQueueRef.current.push(data);
    }
  }, []);

  return (
    <WebSocketContext.Provider value={{ sendMessage }}>
      {children}
    </WebSocketContext.Provider>
  );
};
