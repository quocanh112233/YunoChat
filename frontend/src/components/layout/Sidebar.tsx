'use client';

import React, { useState } from 'react';
import { useAuthStore } from '@/store/auth';
import { Bell, LogOut, Search, MessageSquare, Users, Settings, Plus, UserPlus, Inbox } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useNotificationStore } from '@/store/notification';
import { cn } from '@/lib/utils';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import ConversationList from '../chat/ConversationList';
import FriendList from '../friends/FriendList';
import PendingRequests from '../friends/PendingRequests';
import NotificationPanel from '../notifications/NotificationPanel';
import UserSearchModal from '../friends/UserSearchModal';
import CreateGroupModal from '../friends/CreateGroupModal';

export default function Sidebar() {
  const { user, clearAuth } = useAuthStore();
  const { unreadCount } = useNotificationStore();
  const [activeTab, setActiveTab] = useState<'chats' | 'friends'>('chats');
  const [friendSubTab, setFriendSubTab] = useState<'list' | 'pending'>('list');
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);
  const [isCreateGroupOpen, setIsCreateGroupOpen] = useState(false);
  
  const pathname = usePathname();

  // Mobile visibility logic: hide on mobile if we are in a conversation
  const isChatOpen = pathname.startsWith('/conversations/');

  return (
    <>
      <aside className={cn(
        "w-full md:w-[320px] bg-slate-900 border-r border-slate-800/50 flex flex-col h-full transition-all shrink-0 relative",
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
            <Button 
              variant="ghost" 
              size="icon" 
              className={cn(
                "relative h-8 w-8 transition-colors",
                isNotificationsOpen ? "bg-indigo-600/10 text-indigo-400" : "hover:bg-slate-800 hover:text-indigo-400"
              )}
              onClick={() => setIsNotificationsOpen(!isNotificationsOpen)}
            >
              <Bell className="w-4.5 h-4.5" />
              {unreadCount > 0 && (
                <span className="absolute top-1 right-1 min-w-[14px] h-[14px] px-0.5 bg-rose-500 rounded-full text-[8px] font-black text-white flex items-center justify-center border-2 border-slate-900">
                  {unreadCount > 9 ? '9+' : unreadCount}
                </span>
              )}
            </Button>
            <Link href="/settings">
              <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-slate-800 hover:text-slate-100 transition-colors">
                <Settings className="w-4.5 h-4.5" />
              </Button>
            </Link>
            <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-slate-800 hover:text-rose-400 transition-colors" onClick={() => clearAuth()}>
              <LogOut className="w-4.5 h-4.5" />
            </Button>
          </div>
        </div>

        {/* Search Modal Trigger (Visual Bar) */}
        <div className="p-4 pb-0 shrink-0">
          <div 
            onClick={() => setIsSearchOpen(true)}
            className="flex items-center gap-3 px-3 py-2.5 bg-slate-800/50 border border-slate-700/50 rounded-xl cursor-pointer hover:bg-slate-800 hover:border-indigo-500/30 transition-all group"
          >
            <Search className="w-4 h-4 text-slate-500 group-hover:text-indigo-400 transition-colors" />
            <span className="text-xs text-slate-500 group-hover:text-slate-400">Tìm kiếm người dùng...</span>
          </div>
        </div>

        {/* Tab Selection */}
        <div className="p-4 shrink-0 space-y-3">
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

          {/* Sub-tabs for Friends */}
          {activeTab === 'friends' && (
            <div className="flex gap-2">
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => setFriendSubTab('list')}
                className={cn(
                  "flex-1 text-[10px] font-bold h-7 rounded-lg transition-all",
                  friendSubTab === 'list' ? "bg-indigo-600/10 text-indigo-400" : "text-slate-500 hover:bg-slate-800"
                )}
              >
                <Users className="w-3 h-3 mr-1.5" /> Bạn bè
              </Button>
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => setFriendSubTab('pending')}
                className={cn(
                  "flex-1 text-[10px] font-bold h-7 rounded-lg transition-all",
                  friendSubTab === 'pending' ? "bg-indigo-600/10 text-indigo-400" : "text-slate-500 hover:bg-slate-800"
                )}
              >
                <Inbox className="w-3 h-3 mr-1.5" /> Lời mời
              </Button>
            </div>
          )}
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-2 pb-2 scrollbar-thin scrollbar-thumb-slate-800 flex flex-col">
          <div className="flex-1">
            {activeTab === 'chats' ? (
              <ConversationList />
            ) : (
              <div className="flex flex-col gap-2">
                {friendSubTab === 'list' ? <FriendList /> : <PendingRequests />}
              </div>
            )}
          </div>
        </div>

        {/* Bottom Actions */}
        <div className="p-2 border-t border-slate-800 shrink-0 bg-slate-900/80 backdrop-blur-md">
          {activeTab === 'chats' ? (
            <Button 
              onClick={() => setIsCreateGroupOpen(true)}
              variant="outline" 
              className="w-full justify-start gap-3 bg-slate-800/30 border-slate-700/50 hover:bg-indigo-600/10 hover:border-indigo-500/50 hover:text-indigo-400 rounded-xl transition-all group py-6"
            >
              <div className="w-10 h-10 rounded-full bg-slate-700/50 flex items-center justify-center group-hover:bg-indigo-500/20 transition-all">
                <Plus className="w-5 h-5" />
              </div>
              <span className="text-xs font-bold">Tạo nhóm mới</span>
            </Button>
          ) : (
            <Button 
              onClick={() => setIsSearchOpen(true)}
              variant="outline" 
              className="w-full justify-start gap-3 bg-slate-800/30 border-slate-700/50 hover:bg-indigo-600/10 hover:border-indigo-500/50 hover:text-indigo-400 rounded-xl transition-all group py-6"
            >
              <div className="w-10 h-10 rounded-full bg-slate-700/50 flex items-center justify-center group-hover:bg-indigo-500/20 transition-all">
                <UserPlus className="w-5 h-5" />
              </div>
              <span className="text-xs font-bold">Tìm bạn mới</span>
            </Button>
          )}
        </div>

        {/* Notification Overlay/Panel */}
        {isNotificationsOpen && (
          <div className="absolute top-16 left-4 right-4 z-40 h-[70%] animate-in slide-in-from-top-2 duration-300">
            <NotificationPanel onClose={() => setIsNotificationsOpen(false)} />
          </div>
        )}
      </aside>

      {/* Modals */}
      <UserSearchModal 
        key={isSearchOpen ? 'search-open' : 'search-closed'}
        isOpen={isSearchOpen} 
        onClose={() => setIsSearchOpen(false)} 
      />
      
      <CreateGroupModal 
        key={isCreateGroupOpen ? 'group-open' : 'group-closed'}
        isOpen={isCreateGroupOpen} 
        onClose={() => setIsCreateGroupOpen(false)} 
      />

      {/* Click outside to close notification panel */}
      {isNotificationsOpen && (
        <div 
          className="fixed inset-0 z-30" 
          onClick={() => setIsNotificationsOpen(false)} 
        />
      )}
    </>
  );
}
