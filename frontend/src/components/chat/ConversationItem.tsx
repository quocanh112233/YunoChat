'use client';

import React from 'react';
import { Conversation, Message } from '@/types/api';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAuthStore } from '@/store/auth';
import { cn } from '@/lib/utils';
import dayjs from 'dayjs';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Check, CheckCheck } from 'lucide-react';

interface ConversationItemProps {
  conversation: Conversation & {
    last_message?: Message;
    unread_count?: number;
    other_user?: {
      display_name: string;
      avatar_url: string;
      status: string;
    };
  };
}

export default function ConversationItem({ conversation }: ConversationItemProps) {
  const { user } = useAuthStore();
  const pathname = usePathname();
  const isActive = pathname === `/conversations/${conversation.id}`;

  const displayName = conversation.type === 'DM' 
    ? conversation.other_user?.display_name 
    : conversation.name;
  
  const avatarUrl = conversation.type === 'DM'
    ? conversation.other_user?.avatar_url
    : conversation.avatar_url;

  const lastMsg = conversation.last_message;
  const isOnline = conversation.type === 'DM' && conversation.other_user?.status === 'ONLINE';

  return (
    <Link 
      href={`/conversations/${conversation.id}`}
      className={cn(
        "flex items-center gap-3 px-3 py-3 rounded-lg transition-all cursor-pointer group mb-1",
        isActive ? "bg-slate-700" : "hover:bg-slate-700/50"
      )}
    >
      <div className="relative shrink-0">
        <Avatar className="w-12 h-12 border-2 border-transparent group-hover:border-slate-600 transition-all">
          <AvatarImage src={avatarUrl} />
          <AvatarFallback className="bg-slate-600 text-white font-bold select-none">
            {displayName?.charAt(0) || 'G'}
          </AvatarFallback>
        </Avatar>
        {isOnline && (
          <div className="absolute bottom-0 right-0 w-3.5 h-3.5 bg-emerald-500 rounded-full border-2 border-slate-800" />
        )}
      </div>

      <div className="flex-1 min-w-0 flex flex-col gap-0.5">
        <div className="flex items-center justify-between">
          <span className={cn(
            "text-sm font-semibold truncate",
            conversation.unread_count && conversation.unread_count > 0 ? "text-slate-50" : "text-slate-200"
          )}>
            {displayName}
          </span>
          <span className="text-[10px] text-slate-500 shrink-0">
            {lastMsg ? dayjs(lastMsg.created_at).format('HH:mm') : dayjs(conversation.created_at).format('DD/MM')}
          </span>
        </div>
        
        <div className="flex items-center justify-between gap-2">
          <p className={cn(
            "text-xs truncate",
            conversation.unread_count && conversation.unread_count > 0 ? "text-slate-300 font-medium" : "text-slate-400"
          )}>
            {lastMsg?.body || (lastMsg?.id ? '📎 File đính kèm' : 'Bắt đầu cuộc trò chuyện')}
          </p>
          
          {conversation.unread_count && conversation.unread_count > 0 ? (
            <div className="bg-indigo-500 text-white text-[10px] font-bold min-w-5 h-5 px-1 rounded-full flex items-center justify-center shrink-0">
              {conversation.unread_count > 9 ? '9+' : conversation.unread_count}
            </div>
          ) : lastMsg?.sender_id === user?.id && (
            <div className="shrink-0 flex text-indigo-400">
               {lastMsg?.status === 'READ' ? <CheckCheck className="w-3.5 h-3.5" /> : <Check className="w-3.5 h-3.5 text-slate-500" />}
            </div>
          )}
        </div>
      </div>
    </Link>
  );
}
