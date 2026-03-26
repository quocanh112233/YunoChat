'use client';

import React from 'react';
import { Message } from '@/types/api';
import { cn } from '@/lib/utils';
import dayjs from 'dayjs';
import { Check, CheckCheck, Loader2, RefreshCw, FileText, Download } from 'lucide-react';

interface AttachmentInfo {
  id: string;
  file_type: 'IMAGE' | 'FILE' | 'VIDEO';
  url: string;
  thumbnail_url?: string;
  original_name: string;
  mime_type: string;
  size_bytes: number;
}

// Extend Message type to include attachment if present in the data from WebSocket/API
interface ExtendedMessage extends Message {
  attachment?: AttachmentInfo;
}

interface MessageBubbleProps {
  message: ExtendedMessage;
  isMe: boolean;
  senderName: string;
  onRetry: (msg: Message) => void;
  onImageClick?: (url: string) => void;
}

export default function MessageBubble({ 
  message, 
  isMe, 
  senderName, 
  onRetry,
  onImageClick 
}: MessageBubbleProps) {
  const isSending = message.status === 'SENDING';
  const isDeleted = !!message.deleted_at;
  const isFailed = isSending && dayjs().diff(dayjs(message.created_at), 'second') > 10;

  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const renderContent = () => {
    if (isDeleted) {
      return <p className="flex items-center gap-2 italic text-slate-500">🚫 Tin nhắn đã bị xóa</p>;
    }

    if (message.type === 'TEXT') {
      return <p className="whitespace-pre-wrap wrap-break-word font-sans leading-relaxed">{message.body}</p>;
    }

    if (message.type === 'IMAGE' || (message.type === 'FILE' && message.attachment?.file_type === 'IMAGE')) {
      const imageUrl = message.attachment?.url || message.body;
      return (
        <div 
          className="relative rounded-lg overflow-hidden cursor-pointer hover:opacity-90 transition-opacity"
          onClick={() => imageUrl && onImageClick?.(imageUrl)}
        >
          <img 
            src={imageUrl} 
            alt={message.attachment?.original_name || 'Image'} 
            className="max-w-full max-h-[300px] object-cover"
          />
        </div>
      );
    }

    if (message.type === 'FILE' || message.attachment) {
      const att = message.attachment;
      if (!att) return <p className="text-rose-400">Lỗi: Không tìm thấy tệp đính kèm</p>;

      if (att.file_type === 'VIDEO') {
        return (
          <div className="relative rounded-lg overflow-hidden bg-black/20 group">
            <video 
              src={att.url} 
              className="max-w-full max-h-[300px]"
              controls
            />
          </div>
        );
      }

      // Default File Card
      return (
        <div className="flex items-center gap-3 bg-slate-800/40 p-3 rounded-xl border border-slate-700/50 hover:bg-slate-800/60 transition-colors group">
          <div className="p-2 bg-indigo-500/20 rounded-lg text-indigo-400 group-hover:scale-110 transition-transform">
            <FileText className="w-6 h-6" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-slate-100 truncate">{att.original_name}</p>
            <p className="text-[10px] text-slate-400">{formatSize(att.size_bytes)}</p>
          </div>
          <a 
            href={att.url} 
            download={att.original_name}
            target="_blank"
            rel="noopener noreferrer"
            className="p-2 text-slate-400 hover:text-indigo-400 transition-colors"
          >
            <Download className="w-5 h-5" />
          </a>
        </div>
      );
    }

    return null;
  };

  return (
    <div className={cn(
      "flex flex-col group animate-in fade-in slide-in-from-bottom-2 duration-300",
      isMe ? "items-end" : "items-start"
    )}>
      {!isMe && (
        <span className="text-[10px] text-slate-500 ml-2 mb-1">
          {senderName}
        </span>
      )}

      <div className="flex items-end gap-2 max-w-[85%] md:max-w-[70%]">
        <div className={cn(
          "flex-1 px-4 py-2.5 rounded-2xl text-sm relative transition-all duration-200 shadow-sm",
          isMe 
            ? "bg-indigo-600 text-white rounded-br-none hover:bg-indigo-500" 
            : "bg-slate-700/90 text-slate-100 rounded-bl-none hover:bg-slate-700",
          isSending && "opacity-60",
          isDeleted && "italic text-slate-500 bg-slate-700/40 border border-dashed border-slate-600",
          isFailed && "border-rose-500/50 border",
          (message.type === 'IMAGE' || message.attachment?.file_type === 'IMAGE') && "p-1 bg-transparent border-none shadow-none hover:bg-transparent"
        )}>
          {renderContent()}
          
          <div className={cn(
            "flex items-center gap-1 mt-1 text-[10px]",
            isMe ? "text-indigo-200 justify-end" : "text-slate-400 justify-start",
            (message.type === 'IMAGE' || message.attachment?.file_type === 'IMAGE') && "bg-black/40 px-2 py-0.5 rounded-full backdrop-blur-md inline-flex absolute bottom-2 right-2 m-0"
          )}>
            <span>{dayjs(message.created_at).format('HH:mm')}</span>
            {isMe && (
              <span className="ml-0.5">
                {isSending && !isFailed ? (
                  <Loader2 className="w-3 h-3 animate-spin" />
                ) : isFailed ? (
                  <span className="text-rose-400 font-bold" title="Gửi thất bại">!</span>
                ) : message.status === 'READ' ? (
                  <CheckCheck className="w-3 h-3 text-indigo-300" />
                ) : (
                  <Check className="w-3 h-3 text-slate-400" />
                )}
              </span>
            )}
          </div>
        </div>

        {isFailed && (
          <button 
            onClick={() => onRetry(message)}
            className="mb-2 p-1.5 bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white rounded-full transition-all border border-rose-500/20"
            title="Thử lại"
          >
            <RefreshCw className="w-3 h-3" />
          </button>
        )}
      </div>
    </div>
  );
}
