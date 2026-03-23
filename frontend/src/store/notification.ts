import { create } from 'zustand';
import { Notification } from '@/types/api';

interface NotificationState {
  unreadCount: number;
  notifications: Notification[];
  setUnreadCount: (count: number) => void;
  addNotification: (notification: Notification) => void;
  setNotifications: (notifications: Notification[]) => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  unreadCount: 0,
  notifications: [],
  setUnreadCount: (count) => set({ unreadCount: count }),
  addNotification: (notification) =>
    set((state) => ({
      notifications: [notification, ...state.notifications],
      unreadCount: state.unreadCount + 1,
    })),
  setNotifications: (notifications) => set({ notifications }),
}));