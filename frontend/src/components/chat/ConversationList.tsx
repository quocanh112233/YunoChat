'use client';

import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/axios';
import { Conversation, Message } from '@/types/api';
import ConversationItem from './ConversationItem';
import { Skeleton } from '@/components/ui/skeleton';
import { MessageSquare, Search } from 'lucide-react';
import { Button } from '../ui/button';

// Extend Conversation with last_message and unread_count for the UI
type ConversationWithMeta = Conversation & {
  last_message?: Message;
  unread_count?: number;
  other_user?: {
    display_name: string;
    avatar_url: string;
    status: string;
  };
};

export default function ConversationList() {
  const { data: conversations, isLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const response = await api.get('/conversations');
      return response.data.data as ConversationWithMeta[];
    },
    refetchOnWindowFocus: true,
  });

  const sortedConversations = useMemo(() => {
    if (!conversations) return [];
    return [...conversations].sort((a, b) => 
      new Date(b.last_activity_at).getTime() - new Date(a.last_activity_at).getTime()
    );
  }, [conversations]);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-1">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="flex items-center gap-3 px-3 py-3">
            <Skeleton className="w-12 h-12 rounded-full shrink-0 bg-slate-700" />
            <div className="flex-1 space-y-2 py-1">
              <Skeleton className="h-4 w-3/4 bg-slate-700" />
              <Skeleton className="h-3 w-1/2 bg-slate-700" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (!conversations || conversations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-4 py-12 text-center">
        <div className="w-16 h-16 bg-slate-800/50 rounded-full flex items-center justify-center mb-4 border border-slate-700/50">
          <MessageSquare className="w-8 h-8 text-slate-600 opacity-50" />
        </div>
        <p className="text-slate-400 text-sm font-medium mb-1">Chưa có cuộc trò chuyện nào</p>
        <p className="text-slate-500 text-[10px] mb-6 max-w-[200px]">Hãy bắt đầu kết nối với bạn bè và người thân ngay bây giờ!</p>
        <Button variant="outline" size="sm" className="bg-indigo-600/10 border-indigo-500/20 text-indigo-400 hover:bg-indigo-600 hover:text-white rounded-xl transition-all gap-2 text-[10px] h-8 font-bold">
          <Search className="w-3 h-3" />
          Tìm kiếm bạn bè
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-0.5">
      {sortedConversations.map((conv) => (
        <ConversationItem key={conv.id} conversation={conv} />
      ))}
    </div>
  );
}
