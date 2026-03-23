import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationService } from '@/services/notification';
import { useNotificationStore } from '@/store/notification';
import { useEffect } from 'react';

export function useNotifications() {
  const queryClient = useQueryClient();
  const { setNotifications, setUnreadCount, markAsRead: markReadInStore, markAllAsRead: markAllReadInStore } = useNotificationStore();

  const { data: notifications = [], isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: notificationService.getNotifications,
  });

  useEffect(() => {
    if (notifications) {
      setNotifications(notifications);
      const unread = notifications.filter(n => !n.is_read).length;
      setUnreadCount(unread);
    }
  }, [notifications, setNotifications, setUnreadCount]);

  const markAsReadMutation = useMutation({
    mutationFn: notificationService.markAsRead,
    onSuccess: (_, id) => {
      markReadInStore(id);
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const markAllAsReadMutation = useMutation({
    mutationFn: notificationService.markAllAsRead,
    onSuccess: () => {
      markAllReadInStore();
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  return {
    notifications,
    isLoading,
    markAsRead: markAsReadMutation.mutate,
    markAllAsRead: markAllAsReadMutation.mutate,
  };
}