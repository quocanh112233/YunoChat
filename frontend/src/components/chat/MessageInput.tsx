'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { Send, Smile, Paperclip, Loader2, WifiOff, X, File } from 'lucide-react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useSocketStore } from '@/store/socket';
import { useAuthStore } from '@/store/auth';
import { cn } from '@/lib/utils';
import { useDropzone } from 'react-dropzone';
import { useUpload, UploadType } from '@/hooks/useUpload';
import { toast } from 'sonner';
import { motion } from 'framer-motion';

interface MessageInputProps {
  conversationId: string;
  onSendMessage: (text: string) => void;
  isReadOnly?: boolean;
}

export default function MessageInput({ conversationId, onSendMessage, isReadOnly = false }: MessageInputProps) {
  const [text, setText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const { sendMessage } = useWebSocket();
  const { user } = useAuthStore();
  const { typingIndicators, isConnected, isReconnecting } = useSocketStore();

  const [files, setFiles] = useState<File[]>([]);
  const { isUploading, uploadProgress, uploadFile } = useUpload();

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

  const onDrop = useCallback((acceptedFiles: File[]) => {
    // Client-side validation
    const validFiles = acceptedFiles.filter(file => {
      if (file.size > 50 * 1024 * 1024) {
        toast.error(`File ${file.name} quá lớn (tối đa 50MB)`);
        return false;
      }
      return true;
    });

    setFiles(prev => [...prev, ...validFiles].slice(0, 5)); // Giới hạn 5 file
  }, []);

  const { getRootProps, getInputProps, isDragActive, open } = useDropzone({
    onDrop,
    noClick: true,
    noKeyboard: true,
    multiple: true
  });

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index));
  };

  const handleSend = async () => {
    if ((!text.trim() && files.length === 0) || !isConnected || isUploading) return;

    // 1. Nếu có tệp đính kèm, tải lên trước
    if (files.length > 0) {
      for (const file of files) {
        let uploadType: UploadType = 'FILE';
        if (file.type.startsWith('image/')) uploadType = 'IMAGE';
        else if (file.type.startsWith('video/')) uploadType = 'VIDEO';

        const result = await uploadFile(file, conversationId, uploadType);
        if (result) {
          // Gửi tin nhắn đính kèm qua WebSocket
          sendMessage({
            type: 'send_message',
            payload: {
              conversation_id: conversationId,
              type: 'ATTACHMENT',
              body: '', // Sẽ được backend điền từ attachment
              attachment: {
                storage_type: 'R2',
                file_type: result.file_type,
                url: result.url,
                original_name: result.original_name,
                mime_type: result.mime_type,
                size_bytes: result.size_bytes
              },
              client_temp_id: crypto.randomUUID(),
            },
          });
        }
      }
      setFiles([]);
    }

    // 2. Gửi tin nhắn văn bản (nếu có)
    if (text.trim()) {
      onSendMessage(text.trim());
      setText('');
    }

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

      <div
        {...getRootProps()}
        className={cn(
          "max-w-4xl mx-auto rounded-2xl border transition-all relative",
          isDragActive && !isReadOnly ? "bg-indigo-500/10 border-indigo-500 border-dashed ring-4 ring-indigo-500/20" : "bg-slate-900",
          (!isConnected || isReadOnly) ? "opacity-50 border-slate-800" : "border-slate-800"
        )}
      >
        <input {...getInputProps()} disabled={isReadOnly} />

        {/* Dropzone Overlay */}
        {isDragActive && !isReadOnly && (
          <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-slate-900/80 rounded-2xl animate-in zoom-in-95 duration-200">
            <div className="p-4 bg-indigo-500 text-white rounded-full shadow-lg shadow-indigo-500/40 mb-3 animate-bounce">
              <Paperclip className="w-8 h-8" />
            </div>
            <p className="text-sm font-bold text-slate-100">Thả tệp vào đây để gửi</p>
          </div>
        )}

        {/* File Previews */}
        {files.length > 0 && (
          <div className="flex flex-wrap gap-2 p-3 border-b border-slate-800 animate-in slide-in-from-top-2 duration-300">
            {files.map((file, i) => (
              <div key={i} className="group relative flex items-center gap-2 bg-slate-800 p-2 pr-1 rounded-xl border border-slate-700/50 min-w-[120px] max-w-[200px]">
                {file.type.startsWith('image/') ? (
                  <div className="w-8 h-8 rounded-lg overflow-hidden bg-slate-700">
                    <img src={URL.createObjectURL(file)} alt="preview" className="w-full h-full object-cover" />
                  </div>
                ) : (
                  <div className="w-8 h-8 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400">
                    <File className="w-4 h-4" />
                  </div>
                )}
                <div className="flex-1 min-w-0 pr-1">
                  <p className="text-[10px] font-medium text-slate-200 truncate">{file.name}</p>
                  <p className="text-[8px] text-slate-500">{(file.size / 1024).toFixed(0)} KB</p>
                </div>
                <button
                  onClick={() => removeFile(i)}
                  className="p-1 hover:text-rose-500 text-slate-500 transition-colors"
                >
                  <X className="w-3 h-3" />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Upload Progress Bar */}
        {isUploading && (
          <div className="absolute top-0 left-0 right-0 h-1 bg-slate-800 overflow-hidden rounded-t-2xl z-20">
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${uploadProgress}%` }}
              className="h-full bg-indigo-500 shadow-[0_0_10px_rgba(99,102,241,0.5)]"
            />
          </div>
        )}

        <div className={cn(
          "flex items-end gap-2 p-2 bg-slate-800/50 rounded-2xl transition-all shadow-inner border-transparent",
          (!isConnected || isReadOnly) ? "opacity-50" : ""
        )}>
          <button
            type="button"
            onClick={open}
            disabled={!isConnected || isUploading || isReadOnly}
            className="p-2 text-slate-400 hover:text-indigo-400 hover:bg-slate-700/50 rounded-xl transition-all disabled:opacity-50"
          >
            <Paperclip className="w-5 h-5" />
          </button>

          <textarea
            ref={inputRef}
            value={text}
            onChange={handleTextChange}
            onKeyDown={handleKeyDown}
            disabled={!isConnected || isUploading || isReadOnly}
            placeholder={
              isReadOnly
                ? "Kết bạn để gửi tin nhắn"
                : isUploading
                  ? `Đang tải lên... ${uploadProgress}%`
                  : isConnected ? "Viết tin nhắn..." : "Đang chờ kết nối..."
            }
            className="flex-1 max-h-32 min-h-[40px] bg-transparent border-none focus:ring-0 text-slate-100 placeholder-slate-500 resize-none py-2 px-1 text-sm scrollbar-hide disabled:cursor-not-allowed"
            rows={1}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = `${target.scrollHeight}px`;
            }}
          />

          <button
            disabled={!isConnected || isUploading || isReadOnly}
            className="p-2 text-slate-400 hover:text-yellow-400 hover:bg-slate-700/50 rounded-xl transition-all disabled:opacity-50"
          >
            <Smile className="w-5 h-5" />
          </button>

          <button
            onClick={handleSend}
            disabled={(!text.trim() && files.length === 0) || !isConnected || isUploading || isReadOnly}
            className={cn(
              "p-2.5 rounded-xl transition-all shadow-lg active:scale-95 flex items-center justify-center",
              (isUploading || isReadOnly) ? "bg-slate-700 text-indigo-400" : "bg-indigo-600 text-white hover:bg-indigo-500"
            )}
          >
            {isUploading ? <Loader2 className="w-5 h-5 animate-spin" /> : <Send className="w-5 h-5" />}
          </button>
        </div>
      </div>
    </div>
  );
}
