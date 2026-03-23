'use client';

import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { friendService } from '@/services/friend';
import { conversationService } from '@/services/conversation';
import { RelationshipStatus } from '@/types/api';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Search, UserPlus, Check, X, MessageSquare, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';

interface UserSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function UserSearchModal({ isOpen, onClose }: UserSearchModalProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);


  const { data: searchResults = [], isLoading } = useQuery({
    queryKey: ['user-search', debouncedQuery],
    queryFn: () => friendService.searchUsers(debouncedQuery),
    enabled: debouncedQuery.length > 0,
  });

  const { data: friends = [] } = useQuery({
    queryKey: ['friends'],
    queryFn: friendService.getFriends,
    enabled: isOpen,
  });

  const { data: pendingReceived = [] } = useQuery({
    queryKey: ['pending_received'],
    queryFn: friendService.getPendingRequests,
    enabled: isOpen,
  });

  const { data: pendingSent = [] } = useQuery({
    queryKey: ['pending_sent'],
    queryFn: friendService.getSentRequests,
    enabled: isOpen,
  });

  const sendRequestMutation = useMutation({
    mutationFn: friendService.sendFriendRequest,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending_sent'] });
    },
  });

  const respondMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: 'ACCEPT' | 'DECLINE' }) => 
      friendService.respondToRequest(id, action),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending_received'] });
      queryClient.invalidateQueries({ queryKey: ['friends'] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });

  const startChatMutation = useMutation({
    mutationFn: (userId: string) => conversationService.createConversation(userId),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      if (data?.id) {
        router.push(`/conversations/${data.id}`);
      }
      onClose();
    },
  });

  const getRelationshipStatus = (userId: string): RelationshipStatus => {
    if (friends.some(f => f.id === userId)) return 'FRIEND';
    if (pendingReceived.some(r => r.sender.id === userId)) return 'PENDING_RECEIVED';
    if (pendingSent.some((r) => r.addressee_id === userId)) return 'PENDING_SENT';
    return 'STRANGER';
  };

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-end md:items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200 p-0 md:p-4"
      onClick={onClose}
    >
      <div 
        className="w-full md:max-w-lg bg-slate-900 border-t md:border border-slate-800 rounded-t-3xl md:rounded-2xl shadow-2xl flex flex-col h-[90vh] md:max-h-[80vh] overflow-hidden animate-in slide-in-from-bottom md:slide-in-from-bottom-0 md:zoom-in-95 duration-300 md:duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 border-b border-slate-800 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-indigo-600/20 flex items-center justify-center">
              <UserPlus className="w-4 h-4 text-indigo-400" />
            </div>
            <h2 className="text-lg font-bold text-slate-100">Tìm kiếm bạn mới</h2>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="h-9 w-9 hover:bg-slate-800 rounded-full">
            <XCircle className="w-5 h-5 text-slate-500" />
          </Button>
        </div>

        <div className="p-4 shrink-0 bg-slate-900/50">
          <div className="relative group">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
            <Input 
              autoFocus
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Nhập tên hoặc username..."
              className="h-11 bg-slate-800/50 border-slate-700/50 pl-10 text-sm focus-visible:ring-indigo-500 rounded-xl"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2 scrollbar-thin scrollbar-thumb-slate-800">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-20 gap-3">
              <div className="animate-spin rounded-full h-8 w-8 border-2 border-slate-700 border-t-indigo-500" />
              <p className="text-xs text-slate-500 animate-pulse font-medium">Đang tìm kiếm...</p>
            </div>
          ) : debouncedQuery && searchResults.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-500 italic">
              <div className="w-16 h-16 rounded-full bg-slate-800/50 flex items-center justify-center mb-4">
                <Search className="w-8 h-8 opacity-20" />
              </div>
              <p className="text-sm">Không tìm thấy người dùng nào</p>
            </div>
          ) : (
            <div className="flex flex-col gap-1.5 px-1">
              {searchResults.map((user) => {
                const status = getRelationshipStatus(user.id);
                const isPendingAction = sendRequestMutation.isPending && sendRequestMutation.variables === user.id;
                
                return (
                  <div 
                    key={user.id}
                    className="flex items-center gap-3 p-3 rounded-2xl hover:bg-slate-800/40 transition-all border border-transparent hover:border-slate-800/60 group"
                  >
                    <div className="relative">
                      <Avatar className="w-12 h-12 border border-slate-700/50 shadow-sm">
                        <AvatarImage src={user.avatar_url} />
                        <AvatarFallback className="bg-slate-800 text-indigo-400 font-bold text-lg">
                          {user.display_name.charAt(0)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 bg-slate-900 rounded-full flex items-center justify-center border border-slate-800">
                        <div className={cn(
                          "w-2 h-2 rounded-full",
                          user.status === 'ONLINE' ? "bg-emerald-500" : "bg-slate-600"
                        )} />
                      </div>
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-bold text-slate-100 truncate group-hover:text-indigo-400 transition-colors">
                        {user.display_name}
                      </p>
                      <p className="text-[10px] text-slate-500 font-mono truncate">@{user.username}</p>
                    </div>

                    <div className="shrink-0">
                      {status === 'FRIEND' ? (
                        <Button 
                          size="sm" 
                          variant="ghost" 
                          className="text-indigo-400 hover:text-indigo-300 hover:bg-indigo-500/10 h-9 rounded-xl gap-2 font-bold px-4"
                          onClick={() => startChatMutation.mutate(user.id)}
                          disabled={startChatMutation.isPending}
                        >
                          <MessageSquare className="w-4 h-4" />
                          <span className="hidden sm:inline">Nhắn tin</span>
                        </Button>
                      ) : status === 'PENDING_SENT' ? (
                        <div className="flex items-center gap-2 px-3 py-1.5 bg-slate-800/50 rounded-xl border border-slate-700/50">
                           <span className="text-[10px] font-bold text-slate-400">Đã gửi</span>
                           <div className="w-4 h-4 rounded-full bg-indigo-500/20 flex items-center justify-center">
                              <Check className="w-2.5 h-2.5 text-indigo-400" />
                           </div>
                        </div>
                      ) : status === 'PENDING_RECEIVED' ? (
                        <div className="flex gap-1.5">
                          <Button 
                            size="icon" 
                            className="h-9 w-9 bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600 hover:text-white transition-all rounded-xl"
                            onClick={() => {
                              const req = pendingReceived.find(r => r.sender.id === user.id);
                              if (req) respondMutation.mutate({ id: req.id, action: 'ACCEPT' });
                            }}
                            disabled={respondMutation.isPending}
                          >
                            <Check className="w-4 h-4" />
                          </Button>
                          <Button 
                            size="icon" 
                            variant="ghost" 
                            className="h-9 w-9 text-rose-400 hover:bg-rose-500/10 rounded-xl"
                            onClick={() => {
                              const req = pendingReceived.find(r => r.sender.id === user.id);
                              if (req) respondMutation.mutate({ id: req.id, action: 'DECLINE' });
                            }}
                            disabled={respondMutation.isPending}
                          >
                            <X className="w-4 h-4" />
                          </Button>
                        </div>
                      ) : (
                        <Button 
                          size="sm" 
                          variant="secondary"
                          className="h-9 bg-indigo-600/10 text-indigo-400 hover:bg-indigo-600 hover:text-white border border-indigo-500/20 transition-all gap-2 rounded-xl font-bold px-4 shadow-sm"
                          onClick={() => sendRequestMutation.mutate(user.id)}
                          disabled={isPendingAction}
                        >
                          {isPendingAction ? (
                            <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-indigo-500" />
                          ) : (
                            <>
                              <UserPlus className="w-4 h-4" />
                              <span>Kết bạn</span>
                            </>
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
