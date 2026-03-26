import { useState, useCallback } from 'react';
import axios from 'axios';
import api from '@/lib/axios';
import { toast } from 'sonner';

export type UploadType = 'IMAGE' | 'VIDEO' | 'FILE';

interface PresignAvatarResponse {
  signature: string;
  timestamp: number;
  api_key: string;
  public_id: string;
}

interface PresignFileResponse {
  url: string;
  method: string;
  headers: Record<string, string>;
  public_url: string;
}

export interface UploadResult {
  url: string;
  file_type: UploadType;
  original_name: string;
  mime_type: string;
  size_bytes: number;
}

export const useUpload = () => {
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  const uploadAvatar = useCallback(async (file: File): Promise<string | null> => {
    setIsUploading(true);
    setUploadProgress(0);

    try {
      // 1. Get presigned signature from backend
      const { data: presignData } = await api.post<any>('/upload/avatar/presign', {
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
      });

      const { signature, timestamp, api_key, public_id } = presignData.data as PresignAvatarResponse;

      // 2. Upload to Cloudinary
      const formData = new FormData();
      formData.append('file', file);
      formData.append('api_key', api_key);
      formData.append('timestamp', timestamp.toString());
      formData.append('signature', signature);
      formData.append('public_id', public_id);
      formData.append('folder', 'avatars');

      const cloudName = process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME;
      if (!cloudName) throw new Error('Cloudinary cloud name not configured');

      const uploadRes = await axios.post(
        `https://api.cloudinary.com/v1_1/${cloudName}/image/upload`,
        formData,
        {
          onUploadProgress: (progressEvent) => {
            const progress = progressEvent.total
              ? Math.round((progressEvent.loaded * 100) / progressEvent.total)
              : 0;
            setUploadProgress(progress);
          },
        }
      );

      return uploadRes.data.secure_url;
    } catch (error: any) {
      console.error('Avatar upload failed:', error);
      toast.error(error.response?.data?.error?.message || 'Tải ảnh đại diện thất bại');
      return null;
    } finally {
      setIsUploading(false);
      setUploadProgress(0);
    }
  }, []);

  const uploadFile = useCallback(async (
    file: File, 
    conversationId: string, 
    type: UploadType
  ): Promise<UploadResult | null> => {
    setIsUploading(true);
    setUploadProgress(0);

    try {
      // 1. Get presigned URL from backend
      const { data: presignData } = await api.post<any>('/upload/file/presign', {
        conversation_id: conversationId,
        file_type: type,
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
      });

      const { url, method, headers, public_url } = presignData.data as PresignFileResponse;

      // 2. Upload to R2 via presigned URL
      await axios({
        url,
        method: method || 'PUT',
        data: file,
        headers: {
          ...headers,
          'Content-Type': file.type,
        },
        onUploadProgress: (progressEvent) => {
          const progress = progressEvent.total
            ? Math.round((progressEvent.loaded * 100) / progressEvent.total)
            : 0;
          setUploadProgress(progress);
        },
      });

      return {
        url: public_url,
        file_type: type,
        original_name: file.name,
        mime_type: file.type,
        size_bytes: file.size,
      };
    } catch (error: any) {
      console.error('File upload failed:', error);
      toast.error(error.response?.data?.error?.message || 'Tải tệp tin thất bại');
      return null;
    } finally {
      setIsUploading(false);
      setUploadProgress(0);
    }
  }, []);

  return {
    isUploading,
    uploadProgress,
    uploadAvatar,
    uploadFile,
  };
};