'use client';

import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { friendService } from '@/services/friend';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Check, X, UserPlus } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function PendingRequests() {
  const queryClient = useQueryClient();
  
  const { data: requests = [], isLoading } = useQuery({
    queryKey: ['pending_requests'],
    queryFn: friendService.getPendingRequests,
  });

  const respondMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: 'ACCEPT' | 'DECLINE' }) => 
      friendService.respondToRequest(id, action),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending_requests'] });
      queryClient.invalidateQueries({ queryKey: ['friends'] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-2">
        {[1, 2].map(i => (
          <div key={i} className="h-20 rounded-xl bg-slate-800/20 animate-pulse" />
        ))}
      </div>
    );
  }

  if (requests.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-10 opacity-40">
        <UserPlus className="w-10 h-10 mb-2" />
        <p className="text-xs font-medium">Không có lời mời nào</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 p-2 max-h-[400px] overflow-y-auto scrollbar-thin scrollbar-thumb-slate-800">
      {requests.map((request) => (
        <div 
          key={request.id}
          className="flex flex-col gap-3 p-3 rounded-xl bg-slate-800/40 border border-slate-700/30"
        >
          <div className="flex items-center gap-3">
            <Avatar className="w-10 h-10 border border-slate-700/50">
              <AvatarImage src={request.sender.avatar_url} />
              <AvatarFallback className="bg-slate-700 text-slate-300">
                {request.sender.display_name.charAt(0)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-bold text-slate-100 truncate">
                {request.sender.display_name}
              </p>
              <p className="text-[10px] text-slate-500 font-mono truncate">
                @{request.sender.username}
              </p>
            </div>
          </div>
          
          <div className="flex gap-2">
            <Button 
              size="sm" 
              className="flex-1 bg-indigo-600 hover:bg-indigo-500 text-white text-xs h-8 gap-1"
              onClick={() => respondMutation.mutate({ id: request.id, action: 'ACCEPT' })}
              disabled={respondMutation.isPending}
            >
              <Check className="w-3.5 h-3.5" /> Chấp nhận
            </Button>
            <Button 
              size="sm" 
              variant="secondary"
              className="flex-1 bg-slate-700 hover:bg-slate-600 text-slate-300 text-xs h-8 gap-1"
              onClick={() => respondMutation.mutate({ id: request.id, action: 'DECLINE' })}
              disabled={respondMutation.isPending}
            >
              <X className="w-3.5 h-3.5" /> Từ chối
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
