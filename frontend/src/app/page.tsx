'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';

export default function RootPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.push('/conversations');
    } else {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  return (
    <div className="flex h-screen items-center justify-center bg-slate-900">
      <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-indigo-500" />
    </div>
  );
}
