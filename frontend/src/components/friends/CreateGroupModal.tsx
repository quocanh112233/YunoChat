'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { friendService } from '@/services/friend';
import { conversationService } from '@/services/conversation';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { XCircle, Users, Check, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function CreateGroupModal({ isOpen, onClose }: CreateGroupModalProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [groupName, setGroupName] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  const { data: friends = [], isLoading } = useQuery({
    queryKey: ['friends'],
    queryFn: friendService.getFriends,
    enabled: isOpen,
  });

  const createGroupMutation = useMutation({
    mutationFn: () => conversationService.createGroup(groupName, selectedUserIds),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      if (data?.id) {
        router.push(`/conversations/${data.id}`);
      }
      onClose();
      // Reset state
      setGroupName('');
      setSelectedUserIds([]);
    },
  });

  const filteredFriends = friends.filter(friend => 
    friend.display_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    friend.username.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const toggleUser = (userId: string) => {
    setSelectedUserIds(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };


  if (!isOpen) return null;

  const isFormValid = groupName.trim().length > 0 && selectedUserIds.length >= 2;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200"
      onClick={onClose}
    >
      <div 
        className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl flex flex-col max-h-[85vh] overflow-hidden animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-indigo-400" />
            <h2 className="text-lg font-bold text-slate-100">Tạo nhóm mới</h2>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="h-8 w-8 hover:bg-slate-800">
            <XCircle className="w-5 h-5 text-slate-500" />
          </Button>
        </div>

        <div className="p-4 space-y-4">
          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Tên nhóm</label>
            <Input 
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
              placeholder="Nhập tên nhóm..."
              className="bg-slate-800/50 border-slate-700/50 text-sm h-11 focus-visible:ring-indigo-500"
            />
          </div>

          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Chọn thành viên ({selectedUserIds.length})
            </label>
            <div className="relative group">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 group-focus-within:text-indigo-400 transition-colors" />
              <Input 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Tìm bạn bè..."
                className="h-10 bg-slate-800/50 border-slate-700/50 pl-10 text-sm focus-visible:ring-indigo-500"
              />
            </div>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-4 pb-4 scrollbar-thin scrollbar-thumb-slate-800">
          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500" />
            </div>
          ) : (
            <div className="space-y-1">
              {filteredFriends.map((friend) => (
                <div 
                  key={friend.id}
                  onClick={() => toggleUser(friend.id)}
                  className={cn(
                    "flex items-center gap-3 p-3 rounded-xl cursor-pointer transition-all border border-transparent",
                    selectedUserIds.includes(friend.id) 
                      ? "bg-indigo-600/10 border-indigo-500/30" 
                      : "hover:bg-slate-800/40"
                  )}
                >
                  <Avatar className="w-10 h-10 border border-slate-700/50">
                    <AvatarImage src={friend.avatar_url} />
                    <AvatarFallback className="bg-slate-800 text-slate-400">
                      {friend.display_name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-bold text-slate-100 truncate">{friend.display_name}</p>
                    <p className="text-[10px] text-slate-500 truncate">@{friend.username}</p>
                  </div>
                  <Checkbox 
                    checked={selectedUserIds.includes(friend.id)}
                    onCheckedChange={() => toggleUser(friend.id)}
                    className="rounded-full w-5 h-5 border-slate-700 data-[state=checked]:bg-indigo-600 data-[state=checked]:border-indigo-600"
                  />
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="p-4 border-t border-slate-800 bg-slate-900/50">
          <Button 
            className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-11 rounded-xl shadow-lg disabled:opacity-50 disabled:cursor-not-allowed group"
            disabled={!isFormValid || createGroupMutation.isPending}
            onClick={() => createGroupMutation.mutate()}
          >
            {createGroupMutation.isPending ? (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white" />
            ) : (
              <div className="flex items-center gap-2">
                <Check className="w-5 h-5 group-hover:scale-110 transition-transform" />
                <span>Tạo nhóm hội thoại</span>
              </div>
            )}
          </Button>
          {!isFormValid && groupName.trim() && (
            <p className="text-[10px] text-slate-500 text-center mt-2 italic">
              Cần chọn ít nhất 2 thành viên khác để tạo nhóm
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

