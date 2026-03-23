import api from '@/lib/axios';
import { User, Friendship, FriendRequest, ApiResponse } from '@/types/api';

export const friendService = {
  getFriends: async () => {
    const response = await api.get<ApiResponse<User[]>>('/friends');
    return response.data.data || [];
  },

  getPendingRequests: async () => {
    const response = await api.get<ApiResponse<FriendRequest[]>>('/friends/requests/received');
    return response.data.data || [];
  },

  getSentRequests: async () => {
    const response = await api.get<ApiResponse<Friendship[]>>('/friends/requests/sent');
    return response.data.data || [];
  },

  sendFriendRequest: async (userId: string) => {
    const response = await api.post<ApiResponse<Friendship>>('/friends/requests', {
      addressee_id: userId
    });
    return response.data.data;
  },

  respondToRequest: async (requestId: string, action: 'ACCEPT' | 'DECLINE') => {
    const response = await api.patch<ApiResponse<Friendship>>(`/friends/requests/${requestId}`, {
      action: action
    });
    return response.data.data;
  },

  unfriend: async (userId: string) => {
    await api.delete(`/friends/${userId}`);
  },

  searchUsers: async (query: string) => {
    const response = await api.get<ApiResponse<User[]>>(`/users/search?q=${encodeURIComponent(query)}`);
    return response.data.data || [];
  }
};