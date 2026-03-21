# 5. UI Wireframes & Design Specification

> **Version:** 1.0.0
> **Role:** Senior UX/UI Designer + Frontend Architect
> **Stack:** Next.js 15 App Router · Tailwind CSS · shadcn/ui
> **Dựa trên:** `1_MVP_Requirements.md` v1.1.0 · `4_System_Architecture.md` v1.0.0

---

## 1. Design System — Quy Chuẩn UI/UX Chung

### 1.1 Color Palette (Tailwind CSS Reference)

| Role | Tailwind Token | Hex | Dùng cho |
|------|---------------|-----|---------|
| **Primary** | `indigo-600` | #4F46E5 | Bubble tin nhắn mình gửi, CTA button, active tab |
| **Primary Hover** | `indigo-700` | #4338CA | Button hover state |
| **Primary Light** | `indigo-50` | #EEF2FF | Highlight, selected row |
| **Surface** | `slate-900` | #0F172A | App background (dark) |
| **Surface-2** | `slate-800` | #1E293B | Sidebar background |
| **Surface-3** | `slate-700` | #334155 | Input background, card |
| **Bubble-Other** | `slate-700` | #334155 | Bubble tin nhắn người khác |
| **Border** | `slate-600` | #475569 | Dividers, input border |
| **Text-Primary** | `slate-50` | #F8FAFC | Tiêu đề, nội dung chính |
| **Text-Secondary** | `slate-400` | #94A3B8 | Timestamp, placeholder, metadata |
| **Text-Muted** | `slate-500` | #64748B | Disabled state, hint text |
| **Online** | `emerald-500` | #10B981 | Presence indicator dot |
| **Unread Badge** | `indigo-500` | #6366F1 | Notification badge, unread count |
| **Error** | `red-500` | #EF4444 | Inline validation, error toast |
| **Error Surface** | `red-950` | #450A0A | Error alert background |
| **Warning** | `amber-500` | #F59E0B | File size warning |
| **Tick-Sent** | `slate-400` | #94A3B8 | Single tick (Sent) |
| **Tick-Read** | `indigo-400` | #818CF8 | Double tick (Read) |

### 1.2 Typography

| Scale | Tailwind | Dùng cho |
|-------|----------|---------|
| `text-xl font-bold` | 20px/700 | Page title, modal heading |
| `text-base font-semibold` | 16px/600 | Conversation name, section header |
| `text-sm font-medium` | 14px/500 | Message body, input text |
| `text-xs` | 12px/400 | Timestamp, metadata, badge |
| `text-xs font-semibold` | 12px/600 | Badge number, status label |

### 1.3 Spacing & Layout Constants

```text
Sidebar width   : 320px  (fixed, desktop) → full-width trên mobile
Header height   : 64px   (app shell top bar) → 56px trên mobile
Input area      : 72px   (min-height, flex-shrink-0)
Bubble max-width: 70%    (of message area width) → 85% trên mobile
Avatar size SM  : 32px   (w-8 h-8)
Avatar size MD  : 40px   (w-10 h-10)
Border radius   : rounded-2xl (bubbles), rounded-lg (cards), rounded-full (avatars)
```

### 1.4 Mobile Responsive Strategy

> **Approach:** Mobile-first, progressive enhancement lên desktop.
> **Breakpoints:** Dùng Tailwind breakpoints chuẩn.

| Breakpoint | Width | Layout |
|------------|-------|--------|
| **Mobile** (default) | `< 768px` | Single panel — sidebar HOẶC chat, toggle bằng navigation |
| **Tablet** `md:` | `768px – 1023px` | Sidebar 280px + Chat panel (sidebar ẩn khi mở chat) |
| **Desktop** `lg:` | `≥ 1024px` | Sidebar 320px + Chat panel side-by-side |

#### Mobile Layout Flow

```text
┌──────────────────────────────────────┐
│ MOBILE (< 768px) — Single view       │
│                                      │
│ ┌──────────────────────────────────┐ │
│ │ View A: Conversations List       │ │
│ │ (sidebar full-width)             │ │
│ │                                  │ │
│ │ ┌────────────────────────────┐   │ │
│ │ │ Header: Avatar + "Chats" + 🔔│ │ │
│ │ ├────────────────────────────┤   │ │
│ │ │ 🔍 Tìm kiếm...             │   │ │
│ │ ├────────────────────────────┤   │ │
│ │ │ [Chats] [Friends]          │   │ │
│ │ ├────────────────────────────┤   │ │
│ │ │ ┌──┐ Alice (2 phút trước)  │   │ │
│ │ │ └──┘ Xin chào!        (3) │   │ │   ← Tap → navigate to View B
│ │ │ ─────────────────────────  │   │ │
│ │ │ ┌──┐ Team Alpha             │   │ │
│ │ │ └──┘ Bob: File đã gửi      │   │ │
│ │ └────────────────────────────┘   │ │
│ └──────────────────────────────────┘ │
│                                      │
│ ┌──────────────────────────────────┐ │
│ │ View B: Chat (full-width)        │ │
│ │                                  │ │
│ │ ┌────────────────────────────┐   │ │
│ │ │ ← Alice          ● Online │   │ │   ← Back button → View A
│ │ ├────────────────────────────┤   │ │
│ │ │                            │   │ │
│ │ │    ┌──────────────────┐    │   │ │
│ │ │    │ Xin chào!       │    │   │ │   ← Bubble max-width: 85%
│ │ │    │           09:05  │    │   │ │
│ │ │    └──────────────────┘    │   │ │
│ │ │  ┌──────────────────┐     │   │ │
│ │ │  │ Chào bạn! 🙌     │     │   │ │
│ │ │  │           09:06 ✓│     │   │ │
│ │ │  └──────────────────┘     │   │ │
│ │ │                            │   │ │
│ │ ├────────────────────────────┤   │ │
│ │ │ 📎 Aa...          [Send] │   │ │   ← Input cố định dưới cùng
│ │ └────────────────────────────┘   │ │
│ └──────────────────────────────────┘ │
└──────────────────────────────────────┘
```

#### Responsive Implementation Notes

```text
Navigation:
  - Mobile: dùng Next.js router.push('/conversations/[id]') → full-page transition
  - Desktop: sidebar + chat panel side-by-side, không cần page transition
  - Back button (←) chỉ hiện trên mobile (hidden md:block)

Touch Targets:
  - Minimum 44×44px cho tất cả interactive elements (Apple HIG)
  - Conversation item: min-height 72px (avatar 40px + padding)
  - Action buttons (Accept/Decline friend): min-width 80px, height 40px

Keyboard:
  - Input area: position sticky bottom, tránh bị keyboard che
  - iOS Safari: dùng visualViewport API để detect keyboard height
  - Android: CSS env(keyboard-inset-height) hoặc resize observer

Modals trên Mobile:
  - UserSearchModal → Full-screen sheet (slide up từ bottom, rounded-t-2xl)
  - CreateGroupModal → Full-screen sheet
  - NotificationPanel → Full-screen overlay thay vì dropdown
  - Image Lightbox → Full-screen overlay, swipe to dismiss

CSS Strategy:
  - Container queries cho components cần tự responsive
  - Tailwind: mobile-first → md: → lg: progressive enhancement
  - Sidebar toggle: dùng CSS translate-x + transition, không mount/unmount
```

### 1.4 Loading Skeleton — Sidebar

```text
┌─────────────────────────────────────┐
│ ████████████  [320px Sidebar]       │  ← slate-800 bg
│─────────────────────────────────────│
│  ┌──┐  ████████████████   ██████   │  ← Header skeleton
│  └──┘  ██████████                  │    avatar + name + bell
│─────────────────────────────────────│
│  ┌──────────────────────────────┐   │  ← Search bar skeleton
│  │ ░░░░░░░░░░░░░░░░░░░░░░░░░░ │   │    animate-pulse bg-slate-700
│  └──────────────────────────────┘   │
│─────────────────────────────────────│
│  ████████    ████████               │  ← Tab skeleton
│─────────────────────────────────────│
│  ┌──┐  ████████████████            │  ╮
│  └──┘  ████████████                │  │
│        ████████                    │  │  Conversation item skeletons
│─────────────────────────────────────│  │  (3 items, staggered width)
│  ┌──┐  ██████████████              │  │
│  └──┘  █████████                   │  │
│        ██████████████              │  │
│─────────────────────────────────────│  │
│  ┌──┐  ████████████                │  │
│  └──┘  ███████████████             │  ╯
│        ██████                      │
└─────────────────────────────────────┘
  ░░░ = bg-slate-700 animate-pulse rounded
  ███ = bg-slate-600 animate-pulse rounded
```

### 1.5 Loading Skeleton — Message List

```text
┌──────────────────────────────────────────────────────────┐
│                  [Message Area]                          │
│                                                          │
│  ┌──┐                                                    │
│  └──┘  ┌───────────────────────┐                        │  ← Message người khác
│        │ ████████████████████  │                        │    avatar trái
│        │ █████████████         │                        │
│        └───────────────────────┘                        │
│        ████████                                         │
│                                                          │
│                    ┌─────────────────────────────┐      │  ← Message mình
│                    │ ████████████████████████    │      │    align phải
│                    │ ████████████████            │      │
│                    └─────────────────────────────┘      │
│                    ██████████████                       │
│                                                          │
│  ┌──┐                                                    │
│  └──┘  ┌──────────────────────────────┐                 │
│        │ █████████████████████████    │                 │
│        └──────────────────────────────┘                 │
│        ████████████                                      │
│                                                          │
│                    ┌──────────────────┐                 │
│                    │ ████████████████ │                 │
│                    └──────────────────┘                 │
│                    ██████                               │
└──────────────────────────────────────────────────────────┘
  Skeleton: bg-slate-700 animate-pulse rounded-2xl
  Timestamp skeleton: bg-slate-600, w-16, centered dưới bubble
```

### 1.6 Empty States

**Empty State — Chưa có hội thoại (Conversations tab):**

```text
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                                                         │
│                    ╭─────────────╮                     │
│                    │             │                     │
│                    │  💬         │                     │
│                    │             │                     │
│                    ╰─────────────╯                     │
│                                                         │
│              Chưa có cuộc trò chuyện nào               │
│         text-slate-400 · text-base · text-center        │
│                                                         │
│        Kết bạn với ai đó để bắt đầu trò chuyện         │
│         text-slate-500 · text-sm · text-center          │
│                                                         │
│              ┌──────────────────────────┐               │
│              │  + Tìm kiếm bạn bè       │               │
│              └──────────────────────────┘               │
│               bg-indigo-600 · text-white                │
│               rounded-lg · px-4 py-2                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Empty State — Chưa có bạn bè (Friends tab):**

```text
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                    ╭─────────────╮                     │
│                    │  👥         │                     │
│                    ╰─────────────╯                     │
│                                                         │
│                  Chưa có bạn bè nào                    │
│              text-slate-400 · text-base                 │
│                                                         │
│       Tìm kiếm người dùng và gửi lời mời kết bạn       │
│              text-slate-500 · text-sm                   │
│                                                         │
│              ┌──────────────────────────┐               │
│              │  🔍 Tìm kiếm người dùng  │               │
│              └──────────────────────────┘               │
│                    bg-indigo-600                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Empty State — Chưa có lời mời kết bạn (Pending tab):**

```text
┌─────────────────────────────────────────────────────────┐
│                                                         │
│                    ╭─────────────╮                     │
│                    │  📭         │                     │
│                    ╰─────────────╯                     │
│                                                         │
│              Không có lời mời kết bạn nào              │
│                   text-slate-400                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Màn Hình Auth

### 2.1 Login Page — `/login`

```text
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│                                                                      │
│                    ┌─────────────────────────────┐                  │
│                    │                             │                  │
│                    │    💬  ChatApp              │                  │
│                    │   text-2xl font-bold        │                  │
│                    │   text-indigo-400           │                  │
│                    │                             │                  │
│                    │   Chào mừng trở lại         │                  │
│                    │   text-xl text-slate-50     │                  │
│                    │   font-semibold             │                  │
│                    │                             │                  │
│                    │   Đăng nhập để tiếp tục     │                  │
│                    │   text-sm text-slate-400    │                  │
│                    │                             │                  │
│  ── GLOBAL ERROR (chỉ hiện khi server trả 401) ──────────────────   │
│                    │ ┌─────────────────────────┐ │                  │
│                    │ │ ⚠  Email hoặc mật khẩu  │ │ ← bg-red-950    │
│                    │ │    không đúng            │ │   border-red-500│
│                    │ └─────────────────────────┘ │   text-red-400  │
│                    │                             │   rounded-lg    │
│  ── FIELDS ───────────────────────────────────────────────────────   │
│                    │   Email                     │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ alice@example.com   │   │ ← bg-slate-700  │
│                    │   └─────────────────────┘   │   border-slate-600│
│                    │                             │   focus:border-indigo-500│
│                    │   ← (không lỗi → không có gì)               │  │
│                    │                             │                  │
│                    │   Mật khẩu                  │                  │
│                    │   ┌───────────────────┬───┐ │                  │
│                    │   │ ••••••••••••      │ 👁 │ │ ← toggle show  │
│                    │   └───────────────────┴───┘ │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ ⚠ Mật khẩu tối thiểu│   │ ← INLINE ERROR  │
│                    │   │   8 ký tự           │   │   text-red-400  │
│                    │   └─────────────────────┘   │   text-xs mt-1  │
│                    │                             │   flex gap-1    │
│  ── SUBMIT ────────────────────────────────────────────────────────  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │    Đăng nhập        │   │ ← bg-indigo-600 │
│                    │   └─────────────────────┘   │   w-full py-2.5 │
│                    │   (loading state):          │   rounded-lg    │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │  ◌  Đang đăng nhập  │   │ ← spinner icon  │
│                    │   └─────────────────────┘   │   opacity-80    │
│                    │                             │                  │
│                    │   Chưa có tài khoản?        │                  │
│                    │   Đăng ký ngay →            │                  │
│                    │   text-indigo-400           │                  │
│                    │                             │                  │
│                    └─────────────────────────────┘                  │
│                     max-w-sm · bg-slate-800 · rounded-2xl           │
│                     p-8 · shadow-2xl · mx-auto · mt-16              │
└──────────────────────────────────────────────────────────────────────┘
  Page bg: bg-slate-900  (full viewport)
```

**Inline Validation Rules:**

| Field | Trigger | Message |
|-------|---------|---------|
| Email | `onBlur` + submit | "Email không hợp lệ" |
| Password | `onBlur` + submit | "Mật khẩu tối thiểu 8 ký tự" |
| Global | Server 401 | "Email hoặc mật khẩu không đúng" — box đỏ trên cùng |

---

### 2.2 Register Page — `/register`

```text
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│                    ┌─────────────────────────────┐                  │
│                    │                             │                  │
│                    │    💬  ChatApp              │                  │
│                    │   Tạo tài khoản mới         │                  │
│                    │   text-xl font-semibold     │                  │
│                    │                             │                  │
│  ── GLOBAL ERROR ─────────────────────────────────────────────────   │
│                    │ ┌─────────────────────────┐ │                  │
│                    │ │ ⚠  Email đã được sử dụng│ │ ← bg-red-950    │
│                    │ └─────────────────────────┘ │   (chỉ server)  │
│                    │                             │                  │
│  ── FIELDS ────────────────────────────────────────────────────────  │
│                    │   Tên hiển thị              │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ Alice               │   │                  │
│                    │   └─────────────────────┘   │                  │
│                    │                             │                  │
│                    │   Username                  │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ @alice_dev          │   │ ← prefix "@"    │
│                    │   └─────────────────────┘   │   text-slate-400│
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ ⚠ Chỉ dùng a-z, 0-9,│   │ ← INLINE ERROR  │
│                    │   │   dấu gạch dưới     │   │                  │
│                    │   └─────────────────────┘   │                  │
│                    │                             │                  │
│                    │   Email                     │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ alice@example.com   │   │                  │
│                    │   └─────────────────────┘   │                  │
│                    │   ✓ ← (valid: icon xanh,    │                  │
│                    │         text-emerald-400)   │                  │
│                    │                             │                  │
│                    │   Mật khẩu                  │                  │
│                    │   ┌───────────────────┬───┐ │                  │
│                    │   │ ••••••••••        │ 👁 │ │                  │
│                    │   └───────────────────┴───┘ │                  │
│                    │                             │                  │
│                    │   Xác nhận mật khẩu         │                  │
│                    │   ┌───────────────────┬───┐ │                  │
│                    │   │ ••••••••••        │ 👁 │ │                  │
│                    │   └───────────────────┴───┘ │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │ ⚠ Mật khẩu không    │   │ ← INLINE ERROR  │
│                    │   │   khớp nhau         │   │                  │
│                    │   └─────────────────────┘   │                  │
│                    │                             │                  │
│                    │   ┌─────────────────────┐   │                  │
│                    │   │    Đăng ký          │   │ ← bg-indigo-600 │
│                    │   └─────────────────────┘   │                  │
│                    │                             │                  │
│                    │   Đã có tài khoản?          │                  │
│                    │   Đăng nhập →               │                  │
│                    │                             │                  │
│                    └─────────────────────────────┘                  │
└──────────────────────────────────────────────────────────────────────┘
```

**Inline Validation Rules:**

| Field | Trigger | Thành công | Lỗi |
|-------|---------|-----------|-----|
| Display name | `onBlur` | — | "Tối thiểu 2 ký tự" |
| Username | `onBlur` | icon ✓ emerald | "Chỉ dùng a-z, 0-9, gạch dưới (3-30 ký tự)" |
| Email | `onBlur` | icon ✓ emerald | "Email không hợp lệ" |
| Password | `onBlur` | strength bar | "Tối thiểu 8 ký tự" |
| Confirm | `onChange` | icon ✓ | "Mật khẩu không khớp" |
| Global | Server 409 | — | Box đỏ: "Email/Username đã được sử dụng" |

---

## 3. Main App Shell — Desktop Layout

### 3.1 Tổng Thể Desktop (≥ 768px)

```text
┌────────────────────────────────────────────────────────────────────────────────┐
│                          FULL VIEWPORT (100vw × 100vh)                        │
│                          bg-slate-900                                          │
│                                                                                │
│ ┌──────────────────────┐ ┌──────────────────────────────────────────────────┐ │
│ │   SIDEBAR [320px]    │ │              MAIN CONTENT [flex-1]               │ │
│ │   bg-slate-800       │ │              bg-slate-900                        │ │
│ │   border-r           │ │                                                  │ │
│ │   border-slate-700   │ │  ┌────────────────────────────────────────────┐  │ │
│ │                      │ │  │         CONVERSATION HEADER [64px]         │  │ │
│ │  ┌────────────────┐  │ │  │         border-b border-slate-700          │  │ │
│ │  │  PERSONAL      │  │ │  └────────────────────────────────────────────┘  │ │
│ │  │  HEADER [64px] │  │ │                                                  │ │
│ │  └────────────────┘  │ │  ┌────────────────────────────────────────────┐  │ │
│ │                      │ │  │                                            │  │ │
│ │  ┌────────────────┐  │ │  │           MESSAGE LIST                     │  │ │
│ │  │  SEARCH BAR    │  │ │  │           [flex-1, overflow-y-auto]        │  │ │
│ │  └────────────────┘  │ │  │           px-4 py-4                        │  │ │
│ │                      │ │  │                                            │  │ │
│ │  ┌────────────────┐  │ │  │                                            │  │ │
│ │  │ TABS           │  │ │  │                                            │  │ │
│ │  │ Chats | Friends│  │ │  │                                            │  │ │
│ │  └────────────────┘  │ │  │                                            │  │ │
│ │                      │ │  └────────────────────────────────────────────┘  │ │
│ │  ┌────────────────┐  │ │                                                  │ │
│ │  │                │  │ │  ┌────────────────────────────────────────────┐  │ │
│ │  │  CONVERSATION  │  │ │  │         MESSAGE INPUT [72px min]           │  │ │
│ │  │  LIST          │  │ │  │         border-t border-slate-700          │  │ │
│ │  │  [flex-1,      │  │ │  └────────────────────────────────────────────┘  │ │
│ │  │  overflow-y]   │  │ │                                                  │ │
│ │  │                │  │ │                                                  │ │
│ │  └────────────────┘  │ │              EMPTY STATE (no conversation)       │ │
│ │                      │ │         Chọn một cuộc trò chuyện để bắt đầu     │ │
│ └──────────────────────┘ └──────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────────┘
```

---

### 3.2 Sidebar Chi Tiết

```text
┌────────────────────────────────────┐
│  PERSONAL HEADER  [h-16]           │  bg-slate-800, border-b border-slate-700
│                                    │
│  ┌────┐  Alice               🔔 3  │  ← Avatar w-10 h-10 rounded-full
│  │ A  │  @alice_dev               │    ● online dot: emerald-500, absolute
│  │ ●  │  text-sm text-slate-400   │    Bell icon: lucide BellIcon
│  └────┘                           │    Badge "3": bg-indigo-500, rounded-full
│                           ┌──────┐ │              text-xs w-5 h-5
│                           │ ⚙  ↩ │ │  ← Settings + Logout icons
│                           └──────┘ │    text-slate-400 hover:text-slate-200
│                                    │
├────────────────────────────────────┤
│  SEARCH BAR                        │
│                                    │
│  ┌──────────────────────────────┐  │
│  │ 🔍  Tìm kiếm hoặc thêm bạn  │  │  bg-slate-700, rounded-lg
│  └──────────────────────────────┘  │  placeholder: text-slate-500
│                                    │  focus: border border-indigo-500
├────────────────────────────────────┤
│  TABS                              │
│                                    │
│  ┌─────────────┬─────────────┐    │
│  │   Chats     │   Friends   │    │  active tab: text-indigo-400
│  │  (active)   │             │    │              border-b-2 border-indigo-400
│  └─────────────┴─────────────┘    │  inactive: text-slate-400
│                                    │
│  ← Sub-tabs (dưới Friends tab) →  │
│  ┌──────────┬────────────────┐     │
│  │  Bạn bè  │  Lời mời  [2] │     │  Badge "2": bg-indigo-500 text-xs
│  └──────────┴────────────────┘     │
│                                    │
├────────────────────────────────────┤
│  CONVERSATION LIST  [flex-1]       │
│                                    │
│  ┌────────────────────────────┐    │
│  │ ┌────┐                    │    │  ← Conversation Item
│  │ │ B  │ Bob Smith          │    │    hover: bg-slate-700/50
│  │ │    │ text-sm font-medium│    │    active: bg-slate-700
│  │ └────┘ Ok nhé, gặp lúc 3h │    │    rounded-lg mx-2 px-3 py-3
│  │        text-xs text-slate-400   │    cursor-pointer
│  │        08:28 AM         ●3│    │    ● = unread badge indigo-500
│  └────────────────────────────┘    │
│                                    │
│  ┌────────────────────────────┐    │
│  │ ┌────┐                    │    │  ← GROUP conversation
│  │ │TA  │ Team Alpha         │    │    Group avatar: initials
│  │ │    │ Charlie: Nhìn này  │    │    prefix sender name
│  │ └────┘ 07:00 AM           │    │
│  └────────────────────────────┘    │
│                                    │
│  ┌────────────────────────────┐    │
│  │ ┌────┐                    │    │
│  │ │ D  │ Dave K             │    │
│  │ │    │ 📎 File đính kèm   │    │    attachment preview text
│  │ └────┘ Hôm qua        ✓✓ │    │    ✓✓ = read receipt (indigo)
│  └────────────────────────────┘    │
│                                    │
│  [+ Tạo nhóm mới]                  │  ← floating at bottom
│  text-indigo-400 text-sm           │    or sticky bottom button
│  hover:underline cursor-pointer    │
└────────────────────────────────────┘
```

---

### 3.3 Main Content — Conversation View

```text
┌──────────────────────────────────────────────────────────────┐
│  CONVERSATION HEADER  [h-16]          border-b border-slate-700│
│                                                               │
│  ←  ┌────┐  Bob Smith          ● Online                      │
│  (mob│    │  text-base font-semibold  text-xs text-emerald-400│
│  back│    │                                                   │
│  btn)└────┘                       👥  ⋮                      │
│              (DM: chỉ hiện online status)   Members · More   │
│              (Group: hiện số thành viên)                     │
│                                                               │
│  GROUP header variant:                                        │
│  ←  ┌────┐  Team Alpha      👥 4 thành viên                  │
│     │ TA │  text-base font-semibold   text-xs text-slate-400 │
│     └────┘                  👤+ ⋮ (Add member · More)        │
├──────────────────────────────────────────────────────────────┤
│  MESSAGE LIST  [flex-1, overflow-y-auto, px-4 py-4]          │
│                                                              │
│                  ── Thứ Hai, 15 Tháng 1 ──                  │  ← date divider
│              text-xs text-slate-500 · text-center           │
│                                                              │
│  ┌────┐  Bob Smith                                          │  ← GROUP: show name
│  │ B  │  ┌──────────────────────────────────────────────┐  │
│  └────┘  │ Mọi người ơi, họp lúc 3h nha!               │  │
│          └──────────────────────────────────────────────┘  │
│          08:28 AM                                           │  ← timestamp
│          bg-slate-700 · rounded-2xl rounded-tl-sm          │
│          text-slate-50 · px-4 py-2.5 · max-w-[70%]        │
│                                                              │
│          (consecutive from same sender — no avatar repeat): │
│          ┌──────────────────────────────────────┐          │
│          │ Ok luôn, có agenda chưa?             │          │
│          └──────────────────────────────────────┘          │
│                                                              │
│                    ┌──────────────────────────────────────┐ │  ← MY message
│                    │ Rồi, mình sẽ gửi ngay               │ │    align RIGHT
│                    └──────────────────────────────────────┘ │
│                    08:29 AM  ✓✓                              │    bg-indigo-600
│                    text-right text-xs text-slate-400        │    rounded-2xl rounded-tr-sm
│                    ✓✓ = double tick indigo-400              │    text-white
│                                                              │
│                    ┌──────────────────────────────────────┐ │  ← OPTIMISTIC (sending)
│                    │ Ok mình sẽ vào sớm hơn              │ │    opacity-60
│                    └──────────────────────────────────────┘ │    spinner ◌ thay tick
│                    ◌ Đang gửi...                            │
│                                                              │
│          ┌─────────────────────────────────────────┐        │  ← TYPING indicator
│          │  Bob đang gõ...  ● ● ●                  │        │    bg-slate-700/60
│          └─────────────────────────────────────────┘        │    animate-bounce dots
├──────────────────────────────────────────────────────────────┤
│  MESSAGE INPUT  [min-h-[72px], border-t border-slate-700]   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 📎  Nhập tin nhắn...                            😊 ➤ │   │
│  └──────────────────────────────────────────────────────┘   │
│  bg-slate-700 · rounded-2xl · px-4 py-3 · mx-4 my-3       │
│  📎 = Attach (paperclip icon, text-slate-400)              │
│  😊 = Emoji picker toggle                                   │
│  ➤  = Send button (bg-indigo-600, disabled if empty)       │
└──────────────────────────────────────────────────────────────┘
```

---

### 3.4 Responsive Note — Mobile (< 768px)

```text
MOBILE LAYOUT  [< 768px]
══════════════════════════════════════════════

STATE 1: Hiển thị Sidebar (không có conversation đang mở)
┌─────────────────────────┐
│ SIDEBAR [100vw]         │   ← chiếm toàn màn hình
│ (Main content ẩn)       │   → dùng CSS: md:flex hidden
└─────────────────────────┘

STATE 2: Mở một Conversation
┌─────────────────────────┐
│ MAIN CONTENT [100vw]    │   ← chiếm toàn màn hình
│ ← back button ở header  │   → Sidebar ẩn (translateX(-100%))
│   để quay lại sidebar   │   → animate: slide-in-from-right
└─────────────────────────┘

IMPLEMENTATION NOTE:
- Dùng Zustand store: activeMobileView: 'sidebar' | 'chat'
- Router: khi chọn conversation → setActiveMobileView('chat')
- Back button header → setActiveMobileView('sidebar')
- Tailwind: sidebar có class `hidden md:flex md:w-[320px] w-full`
- Main content: `hidden md:flex flex-1` → khi mobile active: `flex`
- NO slide animation ở MVP: chỉ cần toggle visibility
```

---

## 4. Micro-Components

### 4.1 Notification Dropdown (Bell Icon)

```text
                                         Bell icon tapped
                                                 │
                              ┌──────────────────▼──────────────────────┐
                              │  NOTIFICATION PANEL                      │
                              │  w-80 · bg-slate-800 · rounded-xl        │
                              │  shadow-2xl · border border-slate-700    │
                              │  absolute top-14 right-4 z-50           │
                              │                                          │
                              │  ┌──────────────────────────────────┐   │
                              │  │ 🔔 Thông báo          Đọc tất ✓ │   │
                              │  │ text-base font-semibold          │   │
                              │  └──────────────────────────────────┘   │
                              │  border-b border-slate-700               │
                              │                                          │
                              │  ── FRIEND_REQUEST item (unread) ──      │
                              │  ┌──────────────────────────────────┐   │
                              │  │ bg-indigo-950/40 (unread bg)     │   │
                              │  │ ┌────┐ Charlie đã gửi lời mời    │   │
                              │  │ │ C  │ kết bạn cho bạn           │   │
                              │  │ └────┘ text-sm text-slate-200    │   │
                              │  │        5 phút trước              │   │
                              │  │        text-xs text-slate-500    │   │
                              │  │                                  │   │
                              │  │  ┌──────────┐  ┌─────────────┐  │   │
                              │  │  │ ✓ Chấp   │  │  ✗ Từ chối  │  │   │
                              │  │  │  nhận    │  │             │  │   │
                              │  │  └──────────┘  └─────────────┘  │   │
                              │  │  bg-indigo-600  bg-slate-600     │   │
                              │  │  text-xs px-3 py-1 rounded-md   │   │
                              │  └──────────────────────────────────┘   │
                              │                                          │
                              │  ── FRIEND_ACCEPTED item (read) ──       │
                              │  ┌──────────────────────────────────┐   │
                              │  │ ┌────┐ Bob đã chấp nhận lời mời  │   │
                              │  │ │ B  │ kết bạn của bạn           │   │
                              │  │ └────┘ Hôm qua · 09:15           │   │
                              │  │        → Click: navigate to DM   │   │
                              │  └──────────────────────────────────┘   │
                              │  hover: bg-slate-700/50 cursor-pointer   │
                              │                                          │
                              │  ── GROUP_ADDED item (unread) ──         │
                              │  ┌──────────────────────────────────┐   │
                              │  │ bg-indigo-950/40                 │   │
                              │  │ ┌────┐ Alice đã thêm bạn vào     │   │
                              │  │ │ TA │ nhóm Team Alpha           │   │
                              │  │ └────┘ 2 giờ trước               │   │
                              │  │        → Click: navigate to group│   │
                              │  └──────────────────────────────────┘   │
                              │                                          │
                              │  ┌──────────────────────────────────┐   │
                              │  │     Xem tất cả thông báo →       │   │
                              │  └──────────────────────────────────┘   │
                              │  text-indigo-400 text-sm text-center    │
                              └──────────────────────────────────────────┘

  Overlay (click outside để đóng):
  fixed inset-0 z-40 (transparent)
```

---

### 4.2 Message Bubbles

**Text messages:**

```text
── Tin nhắn người khác (LEFT align) ──────────────────────────────────────

  ┌────┐
  │ B  │  ┌──────────────────────────────────────────────────────┐
  └────┘  │  Mọi người ơi, họp lúc 3h nha! Ai không đến được   │
          │  nhắn mình trước nha 🙏                              │
          └──────────────────────────────────────────────────────┘
          bg-slate-700 · text-slate-50 · rounded-2xl rounded-tl-sm
          px-4 py-2.5 · max-w-[70%] · w-fit
          08:28 AM   ← text-xs text-slate-500

── Tin nhắn mình (RIGHT align) ───────────────────────────────────────────

                              ┌─────────────────────────────────────┐
                              │ Ok mình vào sớm được, mấy giờ      │
                              │ bắt đầu?                            │
                              └─────────────────────────────────────┘
                              bg-indigo-600 · text-white · rounded-2xl rounded-tr-sm
                              px-4 py-2.5 · max-w-[70%] · w-fit · ml-auto
                              08:29 AM  ✓    ← SENT (single, text-slate-400)
                              08:29 AM  ✓✓   ← DELIVERED (double, text-slate-400)
                              08:29 AM  ✓✓   ← READ (double, text-indigo-400)
                              08:29 AM  ◌    ← SENDING (spinner, text-slate-500)

── Soft deleted message ──────────────────────────────────────────────────

          ┌───────────────────────────────┐
          │ 🚫 Tin nhắn đã bị xóa        │
          └───────────────────────────────┘
          bg-slate-700/40 · italic · text-slate-500
          border border-slate-600 border-dashed
```

**Image attachment:**

```text
── Ảnh đính kèm (người khác) ─────────────────────────────────────────────

  ┌────┐
  │ B  │  ┌────────────────────────────────────┐
  └────┘  │                                    │
          │   [IMAGE THUMBNAIL 280×200px]      │  ← object-cover
          │   rounded-xl overflow-hidden       │    cursor-pointer
          │   bg-slate-600 (loading bg)        │    hover: brightness-90
          │                                    │
          └────────────────────────────────────┘
          08:30 AM                              Click → lightbox modal

── Upload progress (mình đang gửi ảnh) ──────────────────────────────────

                    ┌────────────────────────────────────┐
                    │                                    │
                    │   [IMAGE PREVIEW - opacity-50]     │
                    │                                    │
                    └────────────────────────────────────┘
                    ┌────────────────────────────────────┐
                    │ ████████████████░░░░░░░░░░░  72%  │  ← progress bar
                    └────────────────────────────────────┘
                    bg-slate-600 · rounded-full · h-1.5
                    fill: bg-indigo-500
                    text-xs text-slate-400 text-right mt-1
```

**File attachment:**

```text
── File đính kèm (PDF / DOCX / ZIP) ─────────────────────────────────────

  ┌────┐
  │ B  │  ┌──────────────────────────────────────────────┐
  └────┘  │  ┌──────┐                                   │
          │  │      │  Q4_Report.pdf                    │
          │  │  📄  │  5.0 MB · PDF                     │
          │  │      │                         ⬇ Tải về  │
          │  └──────┘                                   │
          └──────────────────────────────────────────────┘
          bg-slate-700 · rounded-2xl · p-3
          file icon bg: bg-slate-600 rounded-lg w-10 h-10
          filename: text-sm font-medium text-slate-50
          meta: text-xs text-slate-400
          download btn: text-indigo-400 hover:text-indigo-300

── File upload in progress (mình gửi) ───────────────────────────────────

                    ┌──────────────────────────────────────────────┐
                    │  ┌──────┐                                   │
                    │  │      │  project_brief.docx               │
                    │  │  📄  │  2.3 MB · DOCX                    │
                    │  │      │                      ◌ Đang gửi  │
                    │  └──────┘                                   │
                    │  ┌────────────────────────────────────────┐ │
                    │  │ █████████████████░░░░░░░░░░░░░░  54%  │ │
                    │  └────────────────────────────────────────┘ │
                    └──────────────────────────────────────────────┘
                    opacity-80 (uploading state)
```

---

### 4.3 Modal Tìm Kiếm & Kết Bạn

```text
  ← Mở bằng: Click Search bar → trigger modal, hoặc Ctrl+K

┌──────────────────────────────────────────────────────────────────┐
│  OVERLAY: fixed inset-0 bg-black/60 z-50                        │
│                                                                  │
│         ┌────────────────────────────────────────────┐          │
│         │  SEARCH MODAL                              │          │
│         │  w-[480px] · bg-slate-800 · rounded-2xl    │          │
│         │  shadow-2xl · mx-auto · mt-24              │          │
│         │                                            │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ 🔍  Tìm theo tên hoặc @username  ✕  │ │          │
│         │  └──────────────────────────────────────┘ │          │
│         │  bg-slate-700 · rounded-xl · px-4 py-3   │          │
│         │  autofocus · text-slate-50               │          │
│         │  ✕ = close modal (Escape key also works) │          │
│         │                                            │          │
│         │  border-b border-slate-700                │          │
│         │                                            │          │
│         │  ── RESULTS ──────────────────────────── │          │
│         │                                            │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ ┌────┐  Charlie Nguyen               │ │          │
│         │  │ │ C  │  @charlie_99                  │ │          │
│         │  │ └────┘                  [+ Kết bạn]  │ │  ← NONE
│         │  └──────────────────────────────────────┘ │          │
│         │  hover: bg-slate-700/50                  │          │
│         │  [+ Kết bạn] = bg-indigo-600 text-xs     │          │
│         │  rounded-md px-3 py-1                    │          │
│         │                                            │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ ┌────┐  Dave Kim                     │ │          │
│         │  │ │ D ●│  @dave_k  · Online            │ │          │
│         │  │ └────┘           [Đã gửi lời mời ↩] │ │  ← PENDING_SENT
│         │  └──────────────────────────────────────┘ │          │
│         │  [Đã gửi...↩] = bg-slate-600 text-slate-400│         │
│         │  cursor-pointer (click để hủy)            │          │
│         │                                            │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ ┌────┐  Eve Tran                     │ │          │
│         │  │ │ E  │  @eve_design                  │ │          │
│         │  │ └────┘             [Phản hồi lời mời]│ │  ← PENDING_RECEIVED
│         │  └──────────────────────────────────────┘ │          │
│         │  → click mở inline Accept/Decline buttons │          │
│         │                                            │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ ┌────┐  Frank Le                     │ │          │
│         │  │ │ F ●│  @frank_l  · Bạn bè           │ │          │
│         │  │ └────┘                  [Nhắn tin →] │ │  ← ACCEPTED
│         │  └──────────────────────────────────────┘ │          │
│         │  [Nhắn tin] = bg-slate-600 text-slate-200 │          │
│         │                                            │          │
│         │  ── EMPTY RESULT STATE ─────────────────  │          │
│         │  (khi gõ nhưng không có kết quả)          │          │
│         │                                            │          │
│         │      🔍 Không tìm thấy "@unknow_user"     │          │
│         │      text-slate-400 · text-sm · text-center│         │
│         │                                            │          │
│         │  ── LOADING STATE ──────────────────────  │          │
│         │  (khi đang fetch)                          │          │
│         │  ┌──────────────────────────────────────┐ │          │
│         │  │ ┌────┐ ██████████████████            │ │          │
│         │  │ └────┘ ████████████                  │ │          │
│         │  └──────────────────────────────────────┘ │          │
│         │  (animate-pulse skeleton)                 │          │
│         │                                            │          │
│         └────────────────────────────────────────────┘          │
└──────────────────────────────────────────────────────────────────┘

  Debounce: 300ms sau khi user ngừng gõ mới gọi API
  Min query: 2 ký tự (dưới 2: không gọi API, không hiện results)
```

---

### 4.4 Modal Tạo Nhóm

```text
┌──────────────────────────────────────────────────────────────┐
│  OVERLAY: fixed inset-0 bg-black/60 z-50                    │
│                                                              │
│       ┌────────────────────────────────────────────┐        │
│       │  TẠO NHÓM MỚI                          ✕  │        │
│       │  text-base font-semibold                   │        │
│       │  border-b border-slate-700                 │        │
│       │                                            │        │
│       │  Tên nhóm                                  │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │ VD: Team Alpha                       │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │                                            │        │
│       │  Thêm thành viên (từ bạn bè)              │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │ 🔍  Tìm bạn bè...                    │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │                                            │        │
│       │  Đã chọn (2/N):                           │        │
│       │  ┌─────────┐  ┌─────────┐                 │        │
│       │  │ Bob  ✕  │  │ Charlie✕│                 │        │
│       │  └─────────┘  └─────────┘                 │        │
│       │  bg-indigo-600/20 border-indigo-500        │        │
│       │  rounded-full text-xs px-3 py-1            │        │
│       │                                            │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │ ┌──┐ Bob Smith          ☑ selected   │ │        │
│       │  │ └──┘ text-sm                          │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │ ┌──┐ Charlie Nguyen     ☑ selected   │ │        │
│       │  │ └──┘                                  │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │ ┌──┐ Dave Kim           ☐            │ │        │
│       │  │ └──┘                                  │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │                                            │        │
│       │  ⚠ Cần ít nhất 2 thành viên khác          │        │
│       │  text-amber-400 text-xs (khi < 2 selected)│        │
│       │                                            │        │
│       │  ┌──────────────────────────────────────┐ │        │
│       │  │          Tạo nhóm                    │ │        │
│       │  └──────────────────────────────────────┘ │        │
│       │  bg-indigo-600 · disabled: opacity-40      │        │
│       │  (disabled khi tên rỗng hoặc < 2 members) │        │
│       │                                            │        │
│       └────────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────────────┘
```

---

## 4.5 Settings Page

```text
┌────────────────────────────────────────────────────────────────┐
│  ← Cài đặt                                              │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│           ┌──────┐                                             │
│           │      │    ← Avatar hiện tại (w-24 h-24)            │
│           │  👤  │       rounded-full                          │
│           │      │                                             │
│           └──────┘                                             │
│         [Đổi ảnh đại diện]   ← text-indigo-400, text-sm       │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Username                                                │  │
│  │ @alice_dev                           (read-only, muted)  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Display Name                                            │  │
│  │ Alice Updated                                   ✏️ edit │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Bio                                                     │  │
│  │ Full-stack developer | Coffee addict ☕                  │  │
│  │                                        135/160 ✏️ edit  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Email                                                   │  │
│  │ alice@example.com                    (read-only, muted)  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           Lưu thay đổi                                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  bg-indigo-600 · disabled khi chưa thay đổi gì                │
│                                                                │
│  ──────── Nguy hiểm ────────                                   │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           Đăng xuất                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│  bg-red-600/10 · text-red-400 · border border-red-600/20       │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

> **Responsive:** Settings page render full-width trên mobile (có nút ← quay lại).  
> Trên desktop: render trong content area bên phải sidebar (max-w-lg mx-auto).  
> Avatar upload: tap vào ảnh → mở File Picker → gọi presign → upload Cloudinary → update URL.

---

## 4.6 Image Lightbox

```text
┌────────────────────────────────────────────────────────────────┐
│  ┌────────────────────────────────────────────────────────┐    │
│  │                                              ✕ close  │    │
│  │                                                        │    │
│  │                                                        │    │
│  │              ┌──────────────────────┐                  │    │
│  │              │                      │                  │    │
│  │              │    Full-size image    │                  │    │
│  │              │    object-contain     │                  │    │
│  │              │    pinch-to-zoom      │                  │    │
│  │              │                      │                  │    │
│  │              └──────────────────────┘                  │    │
│  │                                                        │    │
│  │  bg-black/90 backdrop                                  │    │
│  │                                                        │    │
│  │  ┌────────────────────────────────────────────────┐    │    │
│  │  │ alice_photo.jpg · 2.4 MB · [⬇ Download]       │    │    │
│  │  └────────────────────────────────────────────────┘    │    │
│  │  text-slate-300 text-xs, bottom bar                    │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────────┘
```

> **Interactions:**  
> - Tap image trong chat bubble → mở Lightbox (full-screen overlay)  
> - Desktop: click ✕ hoặc click bên ngoài ảnh → đóng  
> - Mobile: swipe down → đóng (với opacity transition)  
> - Pinch-to-zoom trên mobile (dùng CSS `touch-action: manipulation`)  
> - ESC key → đóng (desktop)  
> - z-index: 50 (trên tất cả UI elements)

---

## 5. Component Mapping — Next.js App Router

| Khu vực trên Wireframe | File Component | Ghi chú |
|------------------------|---------------|---------|
| **AUTH** | | |
| Login form | `components/auth/LoginForm.tsx` | `'use client'`, react-hook-form + zod |
| Register form | `components/auth/RegisterForm.tsx` | `'use client'`, zod schema validation |
| Global error box (đỏ) | Inline trong `LoginForm` / `RegisterForm` | Không tách component riêng |
| Inline field error | Inline dùng `react-hook-form` `formState.errors` | |
| **APP SHELL** | | |
| Root layout (auth guard) | `app/(app)/layout.tsx` | Server Component, kiểm tra cookie |
| Full sidebar | `components/layout/Sidebar.tsx` | `'use client'` (WS events, tabs) |
| Personal header (avatar + bell) | `components/layout/Sidebar.tsx` (top section) | Bell badge từ `useNotifications` hook |
| Bell icon + badge | `components/layout/NotificationBell.tsx` | `'use client'`, Zustand store |
| Search bar (sidebar) | `components/layout/Sidebar.tsx` (inline) | Trigger mở `UserSearchModal` |
| Tabs Chats / Friends | `components/layout/Sidebar.tsx` (inline) | Controlled tab state |
| Conversation list item | Inline trong `Sidebar.tsx` | Lặp qua conversation list |
| Mobile back button | `components/chat/ConversationHeader.tsx` | Ẩn trên `md:` breakpoint |
| **CONVERSATION** | | |
| Conversation header (DM/Group) | `components/chat/ConversationHeader.tsx` | `'use client'`, online status |
| Message list (virtualized) | `components/chat/MessageList.tsx` | `'use client'`, react-window |
| Single message bubble | `components/chat/MessageBubble.tsx` | Xử lý TEXT, ATTACHMENT, deleted |
| Image attachment preview | `components/chat/AttachmentPreview.tsx` | Lightbox modal cho ảnh |
| File attachment card | `components/chat/MessageBubble.tsx` (sub-render) | Nội tuyến trong bubble |
| Upload progress bar | `components/chat/MessageBubble.tsx` (optimistic) | Từ `useUpload` hook |
| Typing indicator | `components/chat/TypingIndicator.tsx` | Nhận WS event `user_typing` |
| Message input bar | `components/chat/MessageInput.tsx` | `'use client'`, file attach, emoji |
| Date divider | Inline trong `MessageList.tsx` | Logic grouping theo ngày |
| **MICRO-COMPONENTS** | | |
| Notification dropdown panel | `components/notifications/NotificationPanel.tsx` | `'use client'`, Zustand |
| Notification item (3 types) | `components/notifications/NotificationItem.tsx` | Render theo `type` prop |
| Accept/Decline inline buttons | `components/notifications/NotificationItem.tsx` | Gọi `friend.service.ts` |
| User search modal | `components/friends/UserSearchModal.tsx` | `'use client'`, debounce 300ms |
| User result item + states | `components/friends/UserSearchModal.tsx` (inline) | 4 relationship states |
| Create group modal | `components/friends/CreateGroupModal.tsx` | `'use client'`, multi-select |
| Friend list | `components/friends/FriendList.tsx` | Server Component + client refetch |
| Friend request card | `components/friends/FriendRequestCard.tsx` | Accept/Decline buttons |
| User avatar with online dot | `components/layout/UserAvatar.tsx` | Props: `size`, `showStatus` |
| **PAGES** | | |
| Login page | `app/(auth)/login/page.tsx` | Server Component wrapper |
| Register page | `app/(auth)/register/page.tsx` | Server Component wrapper |
| Conversations list | `app/(app)/conversations/page.tsx` | Server Component |
| Chat view | `app/(app)/conversations/[id]/page.tsx` | Server → Client handoff |
| Friends page | `app/(app)/friends/page.tsx` | Server Component |
| Settings page | `app/(app)/settings/page.tsx` | Client Component (form) |
| **HOOKS** | | |
| WS connection & dispatch | `hooks/useWebSocket.ts` | Khởi tạo ở App layout |
| Infinite scroll messages | `hooks/useMessages.ts` | TanStack Query `useInfiniteQuery` |
| Online/offline indicator | `hooks/usePresence.ts` | Subscribe WS `presence_update` |
| Bell badge count | `hooks/useNotifications.ts` | Zustand + WS `notification_new` |
| File upload flow | `hooks/useUpload.ts` | Presign → upload → confirm |

---

*Wireframes này là nguồn sự thật cho giao diện MVP. Mọi deviation trong implementation cần ghi chú lại.*
