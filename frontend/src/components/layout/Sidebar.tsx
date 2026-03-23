'use client';

import React, { useState } from 'react';
import { useAuthStore } from '@/store/auth';
import { Bell, LogOut, Search, MessageSquare, Users, Settings, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useNotificationStore } from '@/store/notification';
import { cn } from '@/lib/utils';
import { usePathname } from 'next/navigation';
import ConversationList from '../chat/ConversationList';

export default function Sidebar() {
  const { user, clearAuth } = useAuthStore();
  const { unreadCount } = useNotificationStore();
  const [activeTab, setActiveTab] = useState<'chats' | 'friends'>('chats');
  const pathname = usePathname();

  // Mobile visibility logic: hide on mobile if we are in a conversation
  const isChatOpen = pathname.startsWith('/conversations/');

  return (
    <aside className={cn(
      "w-full md:w-[320px] bg-slate-900 border-r border-slate-800/50 flex flex-col h-full transition-all shrink-0",
      isChatOpen ? "hidden md:flex" : "flex"
    )}>
      {/* Personal Header */}
      <div className="h-16 px-4 flex items-center justify-between border-b border-slate-800 shrink-0 bg-slate-900/50 backdrop-blur-md">
        <div className="flex items-center gap-3">
          <div className="relative group cursor-pointer">
            <Avatar className="w-10 h-10 border-2 border-slate-700 transition-all group-hover:border-indigo-500">
              <AvatarImage src={user?.avatar_url} />
              <AvatarFallback className="bg-indigo-600 text-white font-bold">
                {user?.display_name?.charAt(0) || 'U'}
              </AvatarFallback>
            </Avatar>
            <div className="absolute bottom-0 right-0 w-3 h-3 bg-emerald-500 rounded-full border-2 border-slate-900" />
          </div>
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-semibold truncate text-slate-100">{user?.display_name}</span>
            <span className="text-[10px] text-slate-500 truncate font-mono">@{user?.username}</span>
          </div>
        </div>
        <div className="flex items-center gap-0.5 text-slate-400">
          <Button variant="ghost" size="icon" className="relative h-8 w-8 hover:bg-slate-800 hover:text-indigo-400 transition-colors">
            <Bell className="w-4.5 h-4.5" />
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 w-3.5 h-3.5 bg-rose-500 rounded-full text-[10px] font-bold text-white flex items-center justify-center border-2 border-slate-900">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-slate-800 hover:text-slate-100 transition-colors">
            <Settings className="w-4.5 h-4.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-slate-800 hover:text-rose-400 transition-colors" onClick={() => clearAuth()}>
            <LogOut className="w-4.5 h-4.5" />
          </Button>
        </div>
      </div>

      {/* Tabs & Search */}
      <div className="p-4 flex flex-col gap-3 shrink-0">
        <div className="flex items-center gap-2">
          <div className="relative flex-1 group">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
            <input
              type="text"
              placeholder="Tìm kiếm..."
              className="w-full bg-slate-800/80 border border-slate-700/50 rounded-xl py-2 pl-10 pr-4 text-sm focus:ring-1 focus:ring-indigo-500 text-slate-100 placeholder:text-slate-500 outline-none transition-all"
            />
          </div>
          <Button variant="secondary" size="icon" className="h-9 w-9 bg-indigo-600/10 text-indigo-400 hover:bg-indigo-600 hover:text-white rounded-xl transition-all border border-indigo-500/20">
            <Plus className="w-5 h-5" />
          </Button>
        </div>

        <div className="flex bg-slate-800/50 p-1 rounded-xl border border-slate-700/30">
          <button
            onClick={() => setActiveTab('chats')}
            className={cn(
              "flex-1 flex items-center justify-center gap-2 py-2 text-xs font-bold rounded-lg transition-all",
              activeTab === 'chats' ? "bg-slate-700 text-indigo-400 shadow-sm" : "text-slate-500 hover:text-slate-300"
            )}
          >
            <MessageSquare className="w-3.5 h-3.5" />
            Đoạn chat
          </button>
          <button
            onClick={() => setActiveTab('friends')}
            className={cn(
              "flex-1 flex items-center justify-center gap-2 py-2 text-xs font-bold rounded-lg transition-all",
              activeTab === 'friends' ? "bg-slate-700 text-indigo-400 shadow-sm" : "text-slate-500 hover:text-slate-300"
            )}
          >
            <Users className="w-3.5 h-3.5" />
            Danh bạ
          </button>
        </div>
      </div>

      {/* Conversation List Area */}
      <div className="flex-1 overflow-y-auto px-2 pb-2 scrollbar-thin scrollbar-thumb-slate-800 hover:scrollbar-thumb-slate-700 transition-all flex flex-col">
        <div className="flex-1">
          {activeTab === 'chats' ? (
            <ConversationList />
          ) : (
            <div className="flex flex-col items-center justify-center py-10 text-center">
              <Users className="w-12 h-12 text-slate-700 mb-3 opacity-20" />
              <p className="text-slate-500 text-xs font-semibold">Tính năng đang phát triển</p>
            </div>
          )}
        </div>

        {/* Create Group Button */}
        {activeTab === 'chats' && (
          <div className="mt-auto pt-2 sticky bottom-0 bg-slate-900 pb-2">
            <Button 
              variant="outline" 
              className="w-full justify-start gap-3 bg-slate-800/30 border-slate-700/50 hover:bg-indigo-600/10 hover:border-indigo-500/50 hover:text-indigo-400 rounded-xl transition-all group"
            >
              <div className="w-10 h-10 rounded-full bg-slate-700/50 flex items-center justify-center group-hover:bg-indigo-500/20 transition-all">
                <Users className="w-5 h-5" />
              </div>
              <span className="text-xs font-bold">Tạo nhóm mới</span>
            </Button>
          </div>
        )}
      </div>
    </aside>
  );
}
