'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { friendService } from '@/services/friend';
import { conversationService } from '@/services/conversation';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Search, MessageSquare, UserMinus, MoreVertical } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';

export default function FriendList() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  
  const { data: friends = [], isLoading } = useQuery({
    queryKey: ['friends'],
    queryFn: friendService.getFriends,
  });

  const startChatMutation = useMutation({
    mutationFn: (userId: string) => conversationService.createConversation(userId),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      if (data?.id) {
        router.push(`/conversations/${data.id}`);
      }
    },
  });

  const filteredFriends = friends.filter(friend => 
    friend.display_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    friend.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleStartChat = (userId: string) => {
    startChatMutation.mutate(userId);
  };

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-2">
        {[1, 2, 3].map(i => (
          <div key={i} className="flex items-center gap-3 p-3 rounded-xl bg-slate-800/20 animate-pulse">
            <div className="w-10 h-10 rounded-full bg-slate-700" />
            <div className="flex-1 space-y-2">
              <div className="h-3 bg-slate-700 rounded w-1/2" />
              <div className="h-2 bg-slate-700 rounded w-1/4" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-slate-900/50 rounded-xl overflow-hidden border border-slate-800/50">
      <div className="p-3 border-b border-slate-800/50 bg-slate-900/30">
        <div className="relative group">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
          <Input 
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Tìm bạn bè..."
            className="h-9 bg-slate-800/50 border-slate-700/50 pl-9 text-xs focus-visible:ring-indigo-500/50"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-2 scrollbar-thin scrollbar-thumb-slate-800">
        {filteredFriends.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 opacity-40">
            <UserMinus className="w-10 h-10 mb-2" />
            <p className="text-xs font-medium">Không tìm thấy bạn bè</p>
          </div>
        ) : (
          <div className="flex flex-col gap-1">
            {filteredFriends.map((friend) => (
              <div 
                key={friend.id}
                className="group flex items-center gap-3 p-3 rounded-xl transition-all hover:bg-slate-800/60 border border-transparent hover:border-slate-700/30"
              >
                <div className="relative">
                  <Avatar className="w-10 h-10 border border-slate-700/50">
                    <AvatarImage src={friend.avatar_url} />
                    <AvatarFallback className="bg-indigo-600/20 text-indigo-400 font-bold">
                      {friend.display_name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className={cn(
                    "absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-slate-900",
                    friend.status === 'ONLINE' ? "bg-emerald-500" : "bg-slate-600"
                  )} />
                </div>
                
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-bold text-slate-100 truncate">{friend.display_name}</p>
                  <p className="text-[10px] text-slate-500 font-mono truncate">@{friend.username}</p>
                </div>

                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-all">
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 rounded-lg text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/10"
                    onClick={() => handleStartChat(friend.id)}
                  >
                    <MessageSquare className="w-4 h-4" />
                  </Button>
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10"
                  >
                    <MoreVertical className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
