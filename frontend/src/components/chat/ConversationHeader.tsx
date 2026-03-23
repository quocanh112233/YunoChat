'use client';

import React from 'react';
import { Conversation } from '@/types/api';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { ChevronLeft, Info, Phone, Video } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';

interface ConversationHeaderProps {
  conversation: Conversation & {
    other_user?: {
      display_name: string;
      avatar_url: string;
      status: string;
    };
  };
}

export default function ConversationHeader({ conversation }: ConversationHeaderProps) {
  const router = useRouter();

  const displayName = conversation.type === 'DM' 
    ? conversation.other_user?.display_name 
    : conversation.name;
    
  const avatarUrl = conversation.type === 'DM'
    ? conversation.other_user?.avatar_url
    : conversation.avatar_url;

  const status = conversation.type === 'DM' ? conversation.other_user?.status : `${conversation.member_count} thành viên`;
  const isOnline = conversation.type === 'DM' && conversation.other_user?.status === 'ONLINE';

  return (
    <header className="h-16 px-4 flex items-center justify-between border-b border-slate-700 bg-slate-800/50 backdrop-blur-md sticky top-0 z-10">
      <div className="flex items-center gap-3 min-w-0">
        <Button 
          variant="ghost" 
          size="icon" 
          className="md:hidden h-9 w-9 text-slate-400 hover:bg-slate-700"
          onClick={() => router.push('/conversations')}
        >
          <ChevronLeft className="w-6 h-6" />
        </Button>

        <div className="relative shrink-0">
          <Avatar className="w-10 h-10 border border-slate-700">
            <AvatarImage src={avatarUrl} />
            <AvatarFallback className="bg-slate-700 font-bold">
              {displayName?.charAt(0) || 'G'}
            </AvatarFallback>
          </Avatar>
          {isOnline && (
            <div className="absolute bottom-0 right-0 w-3 h-3 bg-emerald-500 rounded-full border-2 border-slate-800" />
          )}
        </div>

        <div className="flex flex-col min-w-0 overflow-hidden">
          <h1 className="text-sm font-bold text-slate-100 truncate">{displayName}</h1>
          <span className={cn(
            "text-[10px] font-medium transition-colors",
            isOnline ? "text-emerald-400" : "text-slate-500"
          )}>
            {status === 'ONLINE' ? 'Đang hoạt động' : status === 'OFFLINE' ? 'Ngoại tuyến' : status}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-1 text-slate-400">
        <Button variant="ghost" size="icon" className="h-9 w-9 hover:bg-slate-700 hidden sm:flex">
          <Phone className="w-4.5 h-4.5" />
        </Button>
        <Button variant="ghost" size="icon" className="h-9 w-9 hover:bg-slate-700 hidden sm:flex">
          <Video className="w-5 h-5" />
        </Button>
        <Button variant="ghost" size="icon" className="h-9 w-9 hover:bg-slate-700">
          <Info className="w-5 h-5" />
        </Button>
      </div>
    </header>
  );
}
