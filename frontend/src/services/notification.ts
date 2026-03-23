import api from '@/lib/axios';
import { Notification, ApiResponse } from '@/types/api';

export const notificationService = {
  getNotifications: async () => {
    const response = await api.get<ApiResponse<Notification[]>>('/notifications');
    return response.data.data || [];
  },

  markAsRead: async (id: string) => {
    await api.patch(`/notifications/${id}/read`);
  },

  markAllAsRead: async () => {
    await api.patch('/notifications/read-all');
  },
};