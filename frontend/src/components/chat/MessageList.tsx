'use client';

import React, { useEffect, useRef, useState, useMemo, useCallback } from 'react';
import { useInfiniteQuery, useQuery, useQueryClient, InfiniteData } from '@tanstack/react-query';
import api from '@/lib/axios';
import { Message, Conversation, MessageStatus } from '@/types/api';
import { cn } from '@/lib/utils';
import dayjs from 'dayjs';
import 'dayjs/locale/vi';
import calendar from 'dayjs/plugin/calendar';
import { Loader2 } from 'lucide-react';
import { useAuthStore } from '@/store/auth';
import * as ReactWindow from 'react-window';
import { AutoSizer } from 'react-virtualized-auto-sizer';

const VariableSizeList = (ReactWindow as any).VariableSizeList || (ReactWindow as any).default?.VariableSizeList;
import { useWebSocket } from '@/hooks/useWebSocket';
import MessageBubble from './MessageBubble';
import ImageLightbox from './ImageLightbox';

dayjs.extend(calendar);
dayjs.locale('vi');

interface MessageListProps {
  conversationId: string;
}

export default function MessageList({ conversationId }: MessageListProps) {
  const { user } = useAuthStore();
  const { sendMessage } = useWebSocket();
  const queryClient = useQueryClient();
  const listRef = useRef<any>(null);
  const outerRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const [selectedImageUrl, setSelectedImageUrl] = useState<string | null>(null);
  const [, setTick] = useState(0);

  // N2: Force re-render periodically to update 'isFailed' status (reactive timeout)
  useEffect(() => {
    const timer = setInterval(() => {
      setTick(t => t + 1);
    }, 5000); // Check every 5s
    return () => clearInterval(timer);
  }, []);

  // Fetch conversation metadata for sender names (DM)
  const { data: conversation } = useQuery({
    queryKey: ['conversations', conversationId],
    queryFn: async () => {
      const response = await api.get(`/conversations/${conversationId}`);
      return response.data.data as Conversation & { other_user?: { display_name: string } };
    },
  });

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
    queryKey: ['messages', conversationId],
    queryFn: async ({ pageParam }) => {
      const response = await api.get(`/conversations/${conversationId}/messages`, {
        params: {
          cursor: pageParam,
          limit: 50,
        },
      });
      return response.data.data as Message[];
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => {
      if (!lastPage || lastPage.length < 50) return undefined;
      return lastPage[lastPage.length - 1].id;
    },
  });

  const allMessages = useMemo(() => {
    if (!data) return [];
    return data.pages.flatMap((page) => page).reverse();
  }, [data]);

  const handleRetry = useCallback((msg: Message) => {
    const now = new Date().toISOString();

    // N1: Update existing message in cache instead of creating duplicate
    queryClient.setQueryData<InfiniteData<Message[]>>(['messages', conversationId], (old) => {
      if (!old) return old;
      const newPages = old.pages.map(page =>
        page.map(m => (m.client_temp_id === msg.client_temp_id || m.id === msg.id)
          ? { ...m, status: 'SENDING' as MessageStatus, created_at: now }
          : m
        )
      );
      return { ...old, pages: newPages };
    });

    sendMessage({
      type: 'send_message',
      payload: {
        conversation_id: conversationId,
        body: msg.body,
        client_temp_id: msg.client_temp_id || msg.id,
      },
    });
  }, [conversationId, sendMessage, queryClient]);

  const getItemSize = useCallback((index: number) => {
    const msg = allMessages[index];
    if (!msg) return 0;

    // Heuristic for height calculation
    const lines = Math.ceil((msg.body?.length || 0) / 40) || 1;
    let height = lines * 24 + 40; // Base text height + padding/time

    // Add space for date divider
    const prevMsg = allMessages[index - 1];
    if (!prevMsg || !dayjs(msg.created_at).isSame(dayjs(prevMsg.created_at), 'day')) {
      height += 60;
    }

    // Add space for sender name
    if (msg.sender_id !== user?.id) {
      height += 20;
    }

    return height;
  }, [allMessages, user?.id]);

  useEffect(() => {
    if (isAtBottom && listRef.current && allMessages.length > 0) {
      listRef.current.scrollToItem(allMessages.length - 1, 'end');
    }
  }, [allMessages, isAtBottom]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
    );
  }

  const MessageRow = ({ index, style }: { index: number; style: React.CSSProperties }) => {
    const msg = allMessages[index];
    if (!msg) return null;
    const isMe = msg.sender_id === user?.id;
    const prevMsg = allMessages[index - 1];
    const isSameDay = prevMsg && dayjs(msg.created_at).isSame(dayjs(prevMsg.created_at), 'day');

    // Determine sender display name
    const senderName = isMe
      ? 'Bạn'
      : conversation?.other_user?.display_name || `Người dùng ${msg.sender_id.slice(0, 4)}`;

    return (
      <div style={style} className="px-4">
        {!isSameDay && (
          <div className="flex justify-center my-6">
            <span className="bg-slate-800/80 px-3 py-1 rounded-full text-[10px] uppercase tracking-wider font-bold text-slate-500 border border-slate-700/50 backdrop-blur-sm shadow-sm select-none">
              {dayjs(msg.created_at).calendar(null, {
                sameDay: '[Hôm nay]',
                lastDay: '[Hôm qua]',
                lastWeek: 'dddd',
                sameElse: 'DD/MM/YYYY'
              })}
            </span>
          </div>
        )}

        <MessageBubble
          message={msg}
          isMe={isMe}
          senderName={senderName}
          onRetry={handleRetry}
          onImageClick={(url) => setSelectedImageUrl(url)}
        />
      </div>
    );
  };

  return (
    <div className="flex-1 relative min-h-0">
      {/* @ts-expect-error - Fixed library type mismatch */}
      <AutoSizer>
        {({ height, width }: { height: number; width: number }) => (
          <VariableSizeList
            ref={listRef}
            outerRef={outerRef}
            height={height}
            width={width}
            itemCount={allMessages.length}
            itemSize={getItemSize}
            onScroll={({ scrollOffset, scrollHeight, clientHeight }: { scrollOffset: number; scrollHeight: number; clientHeight: number }) => {
              const isBottom = Math.abs(scrollHeight - clientHeight - scrollOffset) < 100;
              setIsAtBottom(isBottom);

              if (scrollOffset < 50 && hasNextPage && !isFetchingNextPage) {
                fetchNextPage();
              }
            }}
            className="scrollbar-thin scrollbar-thumb-slate-800 hover:scrollbar-thumb-slate-700 transition-all"
          >
            {MessageRow}
          </VariableSizeList>
        )}
      </AutoSizer>

      <ImageLightbox
        url={selectedImageUrl}
        onClose={() => setSelectedImageUrl(null)}
      />
    </div>
  );
}
