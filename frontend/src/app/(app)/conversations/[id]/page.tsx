'use client';

import React, { useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'next/navigation';
import api from '@/lib/axios';
import { Conversation, Message } from '@/types/api';
import ConversationHeader from '@/components/chat/ConversationHeader';
import MessageList from '@/components/chat/MessageList';
import MessageInput from '@/components/chat/MessageInput';
import { useAuthStore } from '@/store/auth';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useSocketStore } from '@/store/socket';

export default function ConversationPage() {
  const { id: conversationId } = useParams() as { id: string };
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const { sendMessage } = useWebSocket();
  const { addPendingMessage } = useSocketStore();

  // 1. Fetch conversation details
  const { data: conversation } = useQuery({
    queryKey: ['conversations', conversationId],
    queryFn: async () => {
      const response = await api.get(`/conversations/${conversationId}`);
      return response.data.data as Conversation;
    },
  });

  // 2. Message sending function with Optimistic UI & WebSocket
  const handleSendMessage = useCallback((text: string) => {
    if (!user) return;

    const tempId = crypto.randomUUID();
    const optimisticMessage: Message = {
      id: tempId,
      client_temp_id: tempId,
      conversation_id: conversationId,
      sender_id: user.id,
      body: text,
      type: 'TEXT',
      status: 'SENDING',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    // Add to pending messages store
    addPendingMessage(optimisticMessage);

    // Optimistically update the query cache
    queryClient.setQueryData(['messages', conversationId], (old: unknown) => {
      const data = old as { pages: Message[][], pageParams: unknown[] } | undefined;
      // If no data, initialize with the optimistic message
      if (!data) return { pages: [[optimisticMessage]], pageParams: [undefined] };
      
      const newPages = [...data.pages];
      newPages[0] = [optimisticMessage, ...newPages[0]];
      return { ...data, pages: newPages };
    });

    // Send via WebSocket
    sendMessage({
      type: 'send_message',
      payload: {
        conversation_id: conversationId,
        body: text,
        client_temp_id: tempId,
      },
    });

    // Timeout handled by the fact that it stays in 'SENDING' status
    // A more advanced implementation would start a timer here
  }, [conversationId, user, sendMessage, addPendingMessage, queryClient]);

  if (!conversation) return null;

  return (
    <div className="flex flex-col h-full bg-slate-900 shadow-2xl overflow-hidden">
      <ConversationHeader conversation={conversation} />
      
      <div className="flex-1 flex flex-col relative overflow-hidden bg-[url('/chat-bg-dark.png')] bg-repeat bg-fixed opacity-95">
        <MessageList conversationId={conversationId} />
      </div>

      <MessageInput 
        conversationId={conversationId} 
        onSendMessage={handleSendMessage} 
      />
    </div>
  );
}
