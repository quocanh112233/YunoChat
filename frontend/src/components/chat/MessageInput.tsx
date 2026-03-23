'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { Send, Smile, Paperclip, Loader2, WifiOff } from 'lucide-react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useSocketStore } from '@/store/socket';
import { useAuthStore } from '@/store/auth';
import { cn } from '@/lib/utils';

interface MessageInputProps {
  conversationId: string;
  onSendMessage: (text: string) => void;
}

export default function MessageInput({ conversationId, onSendMessage }: MessageInputProps) {
  const [text, setText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const { sendMessage } = useWebSocket();
  const { user } = useAuthStore();
  const { typingIndicators, isConnected, isReconnecting } = useSocketStore();
  
  const conversationTyping = typingIndicators[conversationId] || [];
  const otherTyping = conversationTyping.filter(t => t.user_id !== user?.id);

  useEffect(() => {
    if (typeof window === 'undefined' || !window.visualViewport) return;

    const handleResize = () => {
      if (!containerRef.current || !window.visualViewport) return;
      const offset = window.innerHeight - window.visualViewport.height;
      containerRef.current.style.bottom = `${offset}px`;
    };

    window.visualViewport.addEventListener('resize', handleResize);
    window.visualViewport.addEventListener('scroll', handleResize);
    return () => {
      window.visualViewport?.removeEventListener('resize', handleResize);
      window.visualViewport?.removeEventListener('scroll', handleResize);
    };
  }, []);

  const handleSendTyping = useCallback((typing: boolean) => {
    if (!isConnected) return;
    sendMessage({
      type: typing ? 'typing_start' : 'typing_stop',
      payload: { conversation_id: conversationId },
    });
  }, [conversationId, sendMessage, isConnected]);

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setText(value);

    // Typing indicator logic
    if (!isTyping && value.trim() && isConnected) {
      setIsTyping(true);
      handleSendTyping(true);
    }

    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    
    typingTimeoutRef.current = setTimeout(() => {
      setIsTyping(false);
      handleSendTyping(false);
    }, 2000);
  };

  const handleSend = () => {
    if (!text.trim() || !isConnected) return;
    
    onSendMessage(text.trim());
    setText('');
    
    if (isTyping) {
      setIsTyping(false);
      handleSendTyping(false);
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    }

    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div 
      ref={containerRef}
      className="p-4 bg-slate-900 border-t border-slate-800 transition-all duration-300"
    >
      {/* Typing Indicator & Connection Status */}
      <div className="h-6 mb-1 px-2 flex items-center justify-between">
        <div className="flex-1">
          {otherTyping.length > 0 && isConnected && (
            <p className="text-[10px] text-slate-400 animate-pulse italic">
              {otherTyping.length === 1 
                ? `${otherTyping[0].display_name} đang nhập...`
                : `${otherTyping.length} người đang nhập...`}
            </p>
          )}
        </div>
        
        {isReconnecting && (
          <div className="flex items-center gap-2 text-amber-500 animate-pulse">
            <Loader2 className="w-3 h-3 animate-spin" />
            <span className="text-[10px] font-bold uppercase tracking-wider">Đang kết nối lại...</span>
          </div>
        )}
        
        {!isConnected && !isReconnecting && (
          <div className="flex items-center gap-2 text-rose-500">
            <WifiOff className="w-3 h-3" />
            <span className="text-[10px] font-bold uppercase tracking-wider">Mất kết nối</span>
          </div>
        )}
      </div>

      <div className={cn(
        "max-w-4xl mx-auto flex items-end gap-2 bg-slate-800/50 p-2 rounded-2xl border transition-all shadow-inner",
        !isConnected ? "opacity-50 border-slate-700/30" : "border-slate-700/50"
      )}>
        <button 
          disabled={!isConnected}
          className="p-2 text-slate-400 hover:text-indigo-400 hover:bg-slate-700/50 rounded-xl transition-all disabled:opacity-50"
        >
          <Paperclip className="w-5 h-5" />
        </button>

        <textarea
          ref={inputRef}
          value={text}
          onChange={handleTextChange}
          onKeyDown={handleKeyDown}
          disabled={!isConnected}
          placeholder={isConnected ? "Viết tin nhắn..." : "Đang chờ kết nối..."}
          className="flex-1 max-h-32 min-h-[40px] bg-transparent border-none focus:ring-0 text-slate-100 placeholder-slate-500 resize-none py-2 px-1 text-sm scrollbar-hide disabled:cursor-not-allowed"
          rows={1}
          onInput={(e) => {
            const target = e.target as HTMLTextAreaElement;
            target.style.height = 'auto';
            target.style.height = `${target.scrollHeight}px`;
          }}
        />

        <button 
          disabled={!isConnected}
          className="p-2 text-slate-400 hover:text-yellow-400 hover:bg-slate-700/50 rounded-xl transition-all disabled:opacity-50"
        >
          <Smile className="w-5 h-5" />
        </button>

        <button
          onClick={handleSend}
          disabled={!text.trim() || !isConnected}
          className="p-2.5 bg-indigo-600 text-white rounded-xl hover:bg-indigo-500 disabled:opacity-50 disabled:hover:bg-indigo-600 transition-all shadow-lg active:scale-95"
        >
          <Send className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
}
