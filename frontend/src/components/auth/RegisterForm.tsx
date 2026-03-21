'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useRouter } from 'next/navigation';
import { Eye, EyeOff, Loader2, CheckCircle2 } from 'lucide-react';
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

const registerSchema = z.object({
  email: z.string().email('Email không hợp lệ'),
  username: z.string()
    .min(3, 'Username tối thiểu 3 ký tự')
    .max(30, 'Username tối đa 30 ký tự')
    .regex(/^[a-z0-9_]+$/, 'Chỉ dùng a-z, 0-9, gạch dưới'),
  display_name: z.string().min(2, 'Tên hiển thị tối thiểu 2 ký tự'),
  password: z.string().min(8, 'Mật khẩu tối thiểu 8 ký tự'),
  confirm_password: z.string(),
}).refine((data) => data.password === data.confirm_password, {
  message: 'Mật khẩu không khớp',
  path: ['confirm_password'],
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export default function RegisterForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const setAuth = useAuthStore((state: any) => state.setAuth);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      username: '',
      display_name: '',
      password: '',
      confirm_password: '',
    },
  });

  const emailValue = watch('email');
  const usernameValue = watch('username');

  const onSubmit = async (values: RegisterFormValues) => {
    setError(null);
    try {
      // 1. Register
      await api.post('/auth/register', {
        email: values.email,
        username: values.username,
        display_name: values.display_name,
        password: values.password,
      });

      // 2. Auto login
      const loginResponse = await api.post('/auth/login', {
        email: values.email,
        password: values.password,
      });

      const { user, access_token } = loginResponse.data.data;
      setAuth(user, access_token);
      router.push('/conversations');
    } catch (err: any) {
      if (err.response?.status === 409) {
        setError('Email hoặc Username đã được sử dụng');
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
          Tạo tài khoản mới
        </CardTitle>
        <CardDescription className="text-center text-slate-400">
          Tham gia cùng cộng đồng YunoChat
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-3">
          {error && (
            <Alert variant="destructive" className="border-red-500 bg-red-950 text-red-400">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Field>
            <FieldLabel htmlFor="display_name" className="text-slate-300">Tên hiển thị</FieldLabel>
            <Input
              id="display_name"
              placeholder="Alice"
              className="border-slate-600 bg-slate-700 text-slate-50 focus:border-indigo-500"
              {...register('display_name')}
            />
            <FieldError errors={[errors.display_name]} />
          </Field>

          <Field>
            <FieldLabel htmlFor="username" className="text-slate-300">Username</FieldLabel>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">@</span>
              <Input
                id="username"
                placeholder="alice_dev"
                className="border-slate-600 bg-slate-700 pl-8 text-slate-50 focus:border-indigo-500"
                {...register('username')}
              />
              {usernameValue && !errors.username && (
                <CheckCircle2 size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-emerald-500" />
              )}
            </div>
            <FieldError errors={[errors.username]} />
          </Field>

          <Field>
            <FieldLabel htmlFor="email" className="text-slate-300">Email</FieldLabel>
            <div className="relative">
              <Input
                id="email"
                type="email"
                placeholder="name@example.com"
                className="border-slate-600 bg-slate-700 text-slate-50 focus:border-indigo-500"
                {...register('email')}
              />
              {emailValue && !errors.email && (
                <CheckCircle2 size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-emerald-500" />
              )}
            </div>
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

          <Field>
            <FieldLabel htmlFor="confirm_password" className="text-slate-300">Xác nhận mật khẩu</FieldLabel>
            <div className="relative">
              <Input
                id="confirm_password"
                type={showPassword ? 'text' : 'password'}
                className="border-slate-600 bg-slate-700 pr-10 text-slate-50 focus:border-indigo-500"
                {...register('confirm_password')}
              />
            </div>
            <FieldError errors={[errors.confirm_password]} />
          </Field>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4 pt-4">
          <Button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-indigo-600 py-6 text-base font-semibold hover:bg-indigo-700"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Đang xử lý...
              </>
            ) : (
              'Đăng ký'
            )}
          </Button>
          <div className="text-center text-sm text-slate-400">
            Đã có tài khoản?{' '}
            <Link href="/login" className="text-indigo-400 hover:underline">
              Đăng nhập ngay →
            </Link>
          </div>
        </CardFooter>
      </form>
    </Card>
  );
}
