import api from '@/lib/axios';
import { Conversation, ApiResponse } from '@/types/api';

export const conversationService = {
  getConversations: async () => {
    const response = await api.get<ApiResponse<Conversation[]>>('/conversations');
    return response.data.data || [];
  },

  createConversation: async (userId: string) => {
    const response = await api.post<ApiResponse<Conversation>>('/conversations', {
      user_id: userId
    });
    return response.data.data;
  },

  createGroup: async (name: string, userIds: string[]) => {
    const response = await api.post<ApiResponse<Conversation>>('/conversations/groups', {
      name,
      user_ids: userIds
    });
    return response.data.data;
  }
};