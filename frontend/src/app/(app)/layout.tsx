'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';
import Sidebar from '@/components/layout/Sidebar';
import { WebSocketProvider } from '@/components/providers/WebSocketProvider';

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) return null;

  return (
    <WebSocketProvider>
      <div className="flex h-screen w-full bg-slate-900 text-slate-50 overflow-hidden">
        <Sidebar />
        <main className="flex-1 flex flex-col min-w-0 bg-slate-900 border-l border-slate-700 md:border-l-0">
          {children}
        </main>
      </div>
    </WebSocketProvider>
  );
}
