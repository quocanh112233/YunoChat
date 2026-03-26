'use client';

import React, { useState, useRef } from 'react';
import { useAuthStore } from '@/store/auth';
import { useUpload } from '@/hooks/useUpload';
import api from '@/lib/axios';
import { toast } from 'sonner';
import { Camera, Loader2, User, Mail, Shield, Save } from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

export default function SettingsPage() {
  const { user, setUser } = useAuthStore();
  const { isUploading, uploadProgress, uploadFile } = useUpload();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [displayName, setDisplayName] = useState(user?.display_name || '');
  const [bio, setBio] = useState(user?.bio || '');
  const [isSaving, setIsSaving] = useState(false);

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Client validation
    if (!file.type.startsWith('image/')) {
      toast.error('Vui lòng chọn tệp hình ảnh');
      return;
    }
    if (file.size > 10 * 1024 * 1024) {
      toast.error('Ảnh quá lớn (tối đa 10MB)');
      return;
    }

    try {
      const result = await uploadFile(file, '', 'IMAGE');
      if (result) {
        // Cập nhật URL avatar vào profile
        const response = await api.put('/users/me', {
          avatar_url: result.url,
          display_name: displayName,
          bio: bio,
        });
        
        if (response.data.status === 'success') {
          setUser(response.data.data);
          toast.success('Cập nhật ảnh đại diện thành công');
        }
      }
    } catch (error) {
      console.error('Avatar upload error:', error);
      toast.error('Không thể cập nhật ảnh đại diện');
    }
  };

  const handleSaveProfile = async () => {
    if (!displayName.trim()) {
      toast.error('Tên hiển thị không được để trống');
      return;
    }

    setIsSaving(true);
    try {
      const response = await api.put('/users/me', {
        display_name: displayName.trim(),
        bio: bio.trim(),
      });

      if (response.data.status === 'success') {
        setUser(response.data.data);
        toast.success('Cập nhật hồ sơ thành công');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Có lỗi xảy ra khi lưu hồ sơ');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Cài đặt</h1>
        <p className="text-slate-400">Quản lý tài khoản và thiết lập ứng dụng của bạn.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Sidebar Mini */}
        <div className="space-y-2">
          <div className="p-3 bg-indigo-500/10 text-indigo-400 rounded-xl border border-indigo-500/20 flex items-center gap-3">
            <User className="w-5 h-5" />
            <span className="font-medium text-sm">Hồ sơ cá nhân</span>
          </div>
          <div className="p-3 hover:bg-slate-800/50 text-slate-400 rounded-xl transition-colors flex items-center gap-3 cursor-not-allowed opacity-50">
            <Shield className="w-5 h-5" />
            <span className="font-medium text-sm">Bảo mật</span>
          </div>
        </div>

        {/* Main Content */}
        <div className="md:col-span-2 space-y-6">
          {/* Avatar Section */}
          <section className="bg-slate-900 border border-slate-800 rounded-2xl p-6 overflow-hidden relative">
            {isUploading && (
              <div className="absolute top-0 left-0 right-0 h-1 bg-slate-800">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: `${uploadProgress}%` }}
                  className="h-full bg-indigo-500 shadow-[0_0_10px_rgba(99,102,241,0.5)]"
                />
              </div>
            )}
            
            <h2 className="text-sm font-bold uppercase tracking-wider text-slate-500 mb-6">Ảnh đại diện</h2>
            <div className="flex flex-col sm:flex-row items-center gap-6">
              <div className="relative group">
                <div 
                  className={cn(
                    "w-24 h-24 rounded-full overflow-hidden border-4 border-slate-800 bg-slate-800 flex items-center justify-center transition-all group-hover:border-indigo-500/50",
                    isUploading && "opacity-50"
                  )}
                >
                  {user?.avatar_url ? (
                    <img src={user.avatar_url} alt="Avatar" className="w-full h-full object-cover" />
                  ) : (
                    <User className="w-10 h-10 text-slate-600" />
                  )}
                </div>
                <button 
                  onClick={handleAvatarClick}
                  disabled={isUploading}
                  className="absolute bottom-0 right-0 p-2 bg-indigo-600 text-white rounded-full shadow-lg hover:bg-indigo-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed group-hover:scale-110"
                >
                  {isUploading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Camera className="w-4 h-4" />}
                </button>
                <input 
                  type="file" 
                  ref={fileInputRef} 
                  className="hidden" 
                  accept="image/*" 
                  onChange={handleFileChange} 
                />
              </div>
              <div className="flex-1 text-center sm:text-left">
                <p className="text-slate-200 font-medium mb-1">Thay đổi ảnh đại diện</p>
                <p className="text-xs text-slate-500 mb-4">Hỗ trợ JPG, PNG hoặc GIF. Tối đa 10MB.</p>
                <button 
                  onClick={handleAvatarClick}
                  disabled={isUploading}
                  className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-100 text-sm font-medium rounded-xl transition-all"
                >
                  Chọn thư mục
                </button>
              </div>
            </div>
          </section>

          {/* Profile Info Section */}
          <section className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6">
            <h2 className="text-sm font-bold uppercase tracking-wider text-slate-500">Thông tin cá nhân</h2>
            
            {/* Display Name */}
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Tên hiển thị</label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                <input 
                  type="text" 
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Họ và tên của bạn"
                  className="w-full bg-slate-950 border-slate-800 rounded-xl pl-10 py-3 text-sm text-slate-100 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder:text-slate-600"
                />
              </div>
            </div>

            {/* Email (Read Only) */}
            <div className="space-y-2 opacity-60">
              <label className="text-sm font-medium text-slate-300">Email (Không thể thay đổi)</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                <input 
                  type="email" 
                  value={user?.email || ''} 
                  readOnly 
                  className="w-full bg-slate-900/50 border-slate-800 rounded-xl pl-10 py-3 text-sm text-slate-400 cursor-not-allowed"
                />
              </div>
            </div>

            {/* Bio */}
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Tiểu sử</label>
              <textarea 
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                placeholder="Một chút về bản thân bạn..."
                rows={3}
                className="w-full bg-slate-950 border-slate-800 rounded-xl p-3 text-sm text-slate-100 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder:text-slate-600 resize-none"
              />
            </div>

            <div className="pt-4 border-t border-slate-800 flex justify-end">
              <button 
                onClick={handleSaveProfile}
                disabled={isSaving || isUploading || displayName === user?.display_name && bio === user?.bio}
                className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl shadow-lg shadow-indigo-500/20 transition-all active:scale-95 disabled:opacity-50"
              >
                {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                Lưu thay đổi
              </button>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
