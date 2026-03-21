import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import Cookies from 'js-cookie';

export type UserStatus = 'ONLINE' | 'OFFLINE' | 'AWAY' | 'BUSY';

export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url?: string;
  status: UserStatus;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, accessToken: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      setAuth: (user, accessToken) => {
        // Lưu vào cookie cho middleware (7 ngày)
        Cookies.set('access_token', accessToken, { expires: 7, path: '/' });
        set({
          user,
          accessToken,
          isAuthenticated: true,
        });
      },
      clearAuth: () => {
        Cookies.remove('access_token', { path: '/' });
        set({
          user: null,
          accessToken: null,
          isAuthenticated: false,
        });
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => sessionStorage),
    }
  )
);
