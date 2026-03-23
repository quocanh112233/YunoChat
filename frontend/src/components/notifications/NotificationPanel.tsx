'use client';

import React from 'react';
import { useQueryClient, useMutation } from '@tanstack/react-query';
import { friendService } from '@/services/friend';
import { useNotifications } from '@/hooks/useNotifications';
import { Notification } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Bell, Check, X, MessageSquare, UserPlus, CheckCircle2 } from 'lucide-react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/vi';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/navigation';

dayjs.extend(relativeTime);
dayjs.locale('vi');

interface NotificationPanelProps {
  onClose?: () => void;
}

export default function NotificationPanel({ onClose }: NotificationPanelProps) {
  const { notifications, markAsRead, markAllAsRead, isLoading } = useNotifications();
  const queryClient = useQueryClient();
  const router = useRouter();

  const respondMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: 'ACCEPT' | 'DECLINE' }) => 
      friendService.respondToRequest(id, action),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending_received'] });
      queryClient.invalidateQueries({ queryKey: ['friends'] });
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });

  const handleNotificationClick = (notif: Notification) => {
    if (!notif.is_read) {
      markAsRead(notif.id);
    }
    
    if (notif.type === 'NEW_MESSAGE' || notif.type === 'FRIEND_ACCEPTED') {
      router.push(`/conversations/${notif.entity_id}`);
      onClose?.();
    }
  };

  const handleAccept = (e: React.MouseEvent, notif: Notification) => {
    e.stopPropagation();
    // In notifications, the entity_id is usually a link to the related object.
    // However, for FRIEND_REQUEST, we need the request_id.
    // If the backend sends the friendship_id in entity_id, we use it.
    respondMutation.mutate({ id: notif.entity_id, action: 'ACCEPT' });
  };

  const handleDecline = (e: React.MouseEvent, notif: Notification) => {
    e.stopPropagation();
    respondMutation.mutate({ id: notif.entity_id, action: 'DECLINE' });
  };

  const renderIcon = (type: string) => {
    switch (type) {
      case 'FRIEND_REQUEST': return <UserPlus className="w-4 h-4 text-indigo-400" />;
      case 'FRIEND_ACCEPTED': return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
      case 'NEW_MESSAGE': return <MessageSquare className="w-4 h-4 text-sky-400" />;
      default: return <Bell className="w-4 h-4 text-slate-400" />;
    }
  };

  return (
    <div className="flex flex-col h-full bg-slate-900/95 backdrop-blur-xl border border-slate-800/50 shadow-2xl rounded-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
      <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50 shrink-0">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-indigo-500/10 flex items-center justify-center">
            <Bell className="w-4.5 h-4.5 text-indigo-400" />
          </div>
          <h3 className="font-bold text-slate-100 text-sm">Thông báo</h3>
        </div>
        <Button 
          variant="ghost" 
          size="sm" 
          className="text-[10px] font-bold uppercase tracking-wider text-indigo-400 hover:text-indigo-300 hover:bg-slate-800 px-3 h-7 rounded-lg"
          onClick={() => markAllAsRead()}
        >
          Đọc tất cả
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-800">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 gap-3">
             <div className="animate-spin rounded-full h-8 w-8 border-2 border-slate-700 border-t-indigo-500" />
             <p className="text-[10px] text-slate-500 font-medium">Đang tải...</p>
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 px-10 text-center">
            <div className="w-16 h-16 rounded-full bg-slate-800/30 flex items-center justify-center mb-4">
              <Bell className="w-8 h-8 text-slate-700" />
            </div>
            <p className="text-slate-400 text-xs font-medium">Bạn chưa có thông báo nào</p>
          </div>
        ) : (
          <div className="flex flex-col">
            {notifications.map((notif) => (
              <div 
                key={notif.id}
                onClick={() => handleNotificationClick(notif)}
                className={cn(
                  "p-4 border-b border-slate-800/30 cursor-pointer transition-all hover:bg-slate-800/40 group relative",
                  !notif.is_read && "bg-indigo-500/5"
                )}
              >
                {!notif.is_read && (
                  <div className="absolute left-0 top-0 bottom-0 w-1 bg-indigo-500" />
                )}
                <div className="flex gap-3">
                  <div className="relative shrink-0">
                    <Avatar className="w-10 h-10 border border-slate-700/50 shadow-sm">
                      <AvatarImage src={notif.actor?.avatar_url} />
                      <AvatarFallback className="bg-slate-800 text-indigo-400 font-bold">
                        {notif.actor?.display_name?.charAt(0) || '?'}
                      </AvatarFallback>
                    </Avatar>
                    <div className="absolute -bottom-1 -right-1 w-5 h-5 rounded-full bg-slate-900 border-2 border-slate-900 flex items-center justify-center shadow-sm">
                      {renderIcon(notif.type)}
                    </div>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-slate-200 leading-snug">
                      <span className="font-bold text-slate-100 group-hover:text-indigo-400 transition-colors">
                        {notif.actor?.display_name}
                      </span>
                      {' '}
                      <span className="text-slate-400 text-xs">
                        {notif.type === 'FRIEND_REQUEST' && 'đã gửi lời mời kết bạn.'}
                        {notif.type === 'FRIEND_ACCEPTED' && 'đã chấp nhận lời mời kết bạn.'}
                        {notif.type === 'NEW_MESSAGE' && 'đã gửi tin nhắn mới.'}
                      </span>
                    </p>
                    <span className="text-[10px] text-slate-600 mt-1 block font-medium">
                      {dayjs(notif.created_at).fromNow()}
                    </span>

                    {notif.type === 'FRIEND_REQUEST' && !notif.is_read && (
                      <div className="flex gap-2 mt-3">
                        <Button 
                          size="sm" 
                          className="flex-1 bg-indigo-600 hover:bg-indigo-500 text-white text-[10px] font-bold h-8 gap-1.5 rounded-lg shadow-lg shadow-indigo-500/10"
                          onClick={(e) => handleAccept(e, notif)}
                          disabled={respondMutation.isPending}
                        >
                          <Check className="w-3.5 h-3.5" /> Chấp nhận
                        </Button>
                        <Button 
                          size="sm" 
                          variant="secondary"
                          className="flex-1 bg-slate-800 hover:bg-slate-700 text-slate-300 text-[10px] font-bold h-8 gap-1.5 rounded-lg"
                          onClick={(e) => handleDecline(e, notif)}
                          disabled={respondMutation.isPending}
                        >
                          <X className="w-3.5 h-3.5" /> Từ chối
                        </Button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
