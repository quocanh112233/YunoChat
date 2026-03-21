'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useRouter } from 'next/navigation';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import Link from 'next/link';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Field, FieldLabel, FieldError } from '@/components/ui/field';
import api from '@/lib/axios';
import { useAuthStore } from '@/store/auth';

const loginSchema = z.object({
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(8, 'Mật khẩu tối thiểu 8 ký tự'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const setAuth = useAuthStore((state: any) => state.setAuth);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const onSubmit = async (values: LoginFormValues) => {
    setError(null);
    try {
      const response = await api.post('/auth/login', values);
      const { user, access_token } = response.data.data;
      
      setAuth(user, access_token);
      router.push('/conversations');
    } catch (err: any) {
      if (err.response?.status === 401) {
        setError('Email hoặc mật khẩu không đúng');
      } else {
        setError(err.response?.data?.error?.message || 'Đã có lỗi xảy ra. Vui lòng thử lại sau.');
      }
    }
  };

  return (
    <Card className="border-slate-700 bg-slate-800 shadow-2xl">
      <CardHeader className="space-y-1">
        <div className="flex items-center justify-center space-x-2 text-indigo-400">
          <span className="text-2xl font-bold">💬 YunoChat</span>
        </div>
        <CardTitle className="text-center text-2xl font-bold text-slate-50">
          Chào mừng trở lại
        </CardTitle>
        <CardDescription className="text-center text-slate-400">
          Đăng nhập để tiếp tục trò chuyện
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive" className="border-red-500 bg-red-950 text-red-400">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Field>
            <FieldLabel htmlFor="email" className="text-slate-300">Email</FieldLabel>
            <Input
              id="email"
              type="email"
              placeholder="name@example.com"
              className="border-slate-600 bg-slate-700 text-slate-50 focus:border-indigo-500"
              {...register('email')}
            />
            <FieldError errors={[errors.email]} />
          </Field>

          <Field>
            <FieldLabel htmlFor="password" className="text-slate-300">Mật khẩu</FieldLabel>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? 'text' : 'password'}
                className="border-slate-600 bg-slate-700 pr-10 text-slate-50 focus:border-indigo-500"
                {...register('password')}
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            <FieldError errors={[errors.password]} />
          </Field>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <Button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-indigo-600 py-6 text-base font-semibold hover:bg-indigo-700"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Đang đăng nhập...
              </>
            ) : (
              'Đăng nhập'
            )}
          </Button>
          <div className="text-center text-sm text-slate-400">
            Chưa có tài khoản?{' '}
            <Link href="/register" className="text-indigo-400 hover:underline">
              Đăng ký ngay →
            </Link>
          </div>
        </CardFooter>
      </form>
    </Card>
  );
}
