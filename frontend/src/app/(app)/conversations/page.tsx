'use client';

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/axios';
import { Conversation } from '@/types/api';
import { MessageSquare } from 'lucide-react';

export default function ConversationsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const response = await api.get('/conversations');
      return response.data.data as Conversation[];
    },
  });

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-slate-900">
        <div className="animate-pulse flex flex-col items-center gap-4">
          <div className="w-12 h-12 bg-slate-800 rounded-full" />
          <div className="h-4 w-32 bg-slate-800 rounded" />
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 bg-slate-900 text-center">
        <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mb-6">
          <MessageSquare className="w-8 h-8 text-slate-500" />
        </div>
        <h3 className="text-slate-100 text-lg font-semibold mb-2">Chưa có cuộc trò chuyện nào</h3>
        <p className="text-slate-400 text-sm max-w-xs mb-8">
          Kết bạn với ai đó hoặc tìm kiếm người dùng để bắt đầu trò chuyện.
        </p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex-col items-center justify-center p-8 bg-slate-900 text-center hidden md:flex">
      <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mb-6">
        <MessageSquare className="w-8 h-8 text-indigo-500" />
      </div>
      <h3 className="text-slate-200 text-lg font-semibold">Chọn một cuộc trò chuyện</h3>
      <p className="text-slate-500 text-sm mt-1">Để bắt đầu gửi tin nhắn</p>
    </div>
  );
}
