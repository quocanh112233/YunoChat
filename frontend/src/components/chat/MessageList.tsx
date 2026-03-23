'use client';

import React, { useEffect, useRef, useState, useMemo, useCallback } from 'react';
import { useInfiniteQuery, useQuery, useQueryClient, InfiniteData } from '@tanstack/react-query';
import api from '@/lib/axios';
import { Message, Conversation, MessageStatus } from '@/types/api';
import { cn } from '@/lib/utils';
import dayjs from 'dayjs';
import 'dayjs/locale/vi';
import calendar from 'dayjs/plugin/calendar';
import { Check, CheckCheck, Loader2, RefreshCw } from 'lucide-react';
import { useAuthStore } from '@/store/auth';
// @ts-expect-error - Fixed library type mismatch
import { VariableSizeList } from 'react-window';
// @ts-expect-error - Fixed library type mismatch
import AutoSizer from 'react-virtualized-auto-sizer';
import { useWebSocket } from '@/hooks/useWebSocket';

dayjs.extend(calendar);
dayjs.locale('vi');

interface MessageListProps {
  conversationId: string;
}

export default function MessageList({ conversationId }: MessageListProps) {
  const { user } = useAuthStore();
  const { sendMessage } = useWebSocket();
  const queryClient = useQueryClient();
  const listRef = useRef<VariableSizeList>(null);
  const outerRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);
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
    const isSending = msg.status === 'SENDING';
    const isDeleted = !!msg.deleted_at;
    
    const isFailed = isSending && dayjs().diff(dayjs(msg.created_at), 'second') > 10;

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

        <div className={cn(
          "flex flex-col group transition-all duration-300",
          isMe ? "items-end" : "items-start"
        )}>
          {!isMe && (
            <span className="text-[10px] text-slate-500 ml-2 mb-1">
              {senderName}
            </span>
          )}

          <div className="flex items-end gap-2 max-w-[85%] md:max-w-[70%]">
            <div className={cn(
              "flex-1 px-4 py-2.5 rounded-2xl text-sm relative transition-all duration-200 shadow-sm",
              isMe 
                ? "bg-indigo-600 text-white rounded-br-none hover:bg-indigo-500" 
                : "bg-slate-700/90 text-slate-100 rounded-bl-none hover:bg-slate-700",
              isSending && "opacity-60",
              isDeleted && "italic text-slate-500 bg-slate-700/40 border border-dashed border-slate-600",
              isFailed && "border-rose-500/50 border"
            )}>
              {isDeleted ? (
                <p className="flex items-center gap-2">🚫 Tin nhắn đã bị xóa</p>
              ) : (
                <p className="whitespace-pre-wrap wrap-break-word leading-relaxed">{msg.body}</p>
              )}
              
              <div className={cn(
                "flex items-center gap-1 mt-1 text-[10px]",
                isMe ? "text-indigo-200 justify-end" : "text-slate-400 justify-start"
              )}>
                <span>{dayjs(msg.created_at).format('HH:mm')}</span>
                {isMe && (
                  <span className="ml-0.5">
                    {isSending && !isFailed ? (
                      <Loader2 className="w-3 h-3 animate-spin" />
                    ) : isFailed ? (
                      <span className="text-rose-400 font-bold" title="Gửi thất bại">!</span>
                    ) : msg.status === 'READ' ? (
                      <CheckCheck className="w-3 h-3 text-indigo-300" />
                    ) : (
                      <Check className="w-3 h-3 text-slate-400" />
                    )}
                  </span>
                )}
              </div>
            </div>

            {isFailed && (
              <button 
                onClick={() => handleRetry(msg)}
                className="mb-2 p-1.5 bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white rounded-full transition-all border border-rose-500/20"
                title="Thử lại"
              >
                <RefreshCw className="w-3 h-3" />
              </button>
            )}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="flex-1 relative min-h-0">
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
    </div>
  );
}
