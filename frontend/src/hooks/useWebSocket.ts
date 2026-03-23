'use client';

import { useWebSocketContext } from '@/components/providers/WebSocketProvider';

export const useWebSocket = () => {
  const { sendMessage } = useWebSocketContext();
  return { sendMessage };
};