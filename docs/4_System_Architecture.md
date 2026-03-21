# 4. System Architecture

> **Version:** 1.0.0
> **Dựa trên:** `3_API_WebSocket_Specs.md` v1.0.0
> **Stack:** Next.js 15 (App Router) · Golang 1.23 · PostgreSQL 16 · Cloudinary · Cloudflare R2

---

## 1. Tổng Quan Hệ Thống

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT (Vercel)                             │
│                    Next.js 15 — App Router                          │
│           Browser ◄──── HTTPS/WSS ────► fly.io                     │
└──────────────────────────────┬──────────────────────────────────────┘
                               │  REST (HTTPS) + WebSocket (WSS)
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      BACKEND (fly.io)                               │
│                   Golang — Single Instance                          │
│                                                                     │
│   ┌────────────┐   ┌────────────┐   ┌────────────────────────┐     │
│   │ HTTP Router│   │   WS Hub   │   │  pg LISTEN goroutine   │     │
│   │  (chi/v5)  │   │            │◄──│  chat_events channel   │     │
│   └─────┬──────┘   └─────┬──────┘   └────────────────────────┘     │
│         │                │                                          │
│   ┌─────▼────────────────▼──────┐                                  │
│   │      Use Cases / Services   │                                   │
│   └─────────────┬───────────────┘                                  │
│                 │                                                    │
│   ┌─────────────▼───────────────┐                                  │
│   │    Repository (sqlc/pgx)    │                                   │
│   └─────────────┬───────────────┘                                  │
└─────────────────┼───────────────────────────────────────────────────┘
                  │  pgxpool (queries) + listenConn (LISTEN)
     ┌────────────┼────────────────────┐
     ▼            ▼                    ▼
┌─────────┐ ┌──────────┐      ┌───────────────┐
│ Neon DB │ │Cloudinary│      │ Cloudflare R2 │
│ (Pgsql) │ │ (images) │      │ (files/video) │
└─────────┘ └──────────┘      └───────────────┘
```

---

## 2. Folder Structure — Backend (Golang)

> **Triết lý:** Clean Architecture + DDD. Dependency chỉ đi từ ngoài vào trong. `domain` không import bất kỳ package nào ngoài stdlib.

```
backend/
├── cmd/
│   └── server/
│       └── main.go                  # Entrypoint: wire DI, start HTTP server
│
├── internal/
│   │
│   ├── domain/                      # ★ Lõi DDD — không import gì ngoài stdlib
│   │   ├── user/
│   │   │   ├── entity.go            # User struct, value objects
│   │   │   ├── repository.go        # Interface: UserRepository
│   │   │   └── errors.go            # Domain errors: ErrUserNotFound, etc.
│   │   ├── friendship/
│   │   │   ├── entity.go            # Friendship struct, Status enum
│   │   │   └── repository.go        # Interface: FriendshipRepository
│   │   ├── conversation/
│   │   │   ├── entity.go            # Conversation, Participant structs
│   │   │   └── repository.go        # Interface: ConversationRepository
│   │   ├── message/
│   │   │   ├── entity.go            # Message, Attachment structs
│   │   │   └── repository.go        # Interface: MessageRepository
│   │   └── notification/
│   │       ├── entity.go            # Notification struct, Type enum
│   │       └── repository.go        # Interface: NotificationRepository
│   │
│   ├── usecase/                     # Application logic — orchestrate domain
│   │   ├── auth/
│   │   │   ├── register.go          # RegisterUseCase
│   │   │   ├── login.go             # LoginUseCase
│   │   │   ├── refresh.go           # RefreshTokenUseCase
│   │   │   └── logout.go            # LogoutUseCase
│   │   ├── friendship/
│   │   │   ├── send_request.go      # SendFriendRequestUseCase
│   │   │   ├── respond_request.go   # AcceptUseCase, DeclineUseCase
│   │   │   └── unfriend.go          # UnfriendUseCase
│   │   ├── conversation/
│   │   │   ├── list.go              # ListConversationsUseCase
│   │   │   ├── create_group.go      # CreateGroupUseCase
│   │   │   ├── manage_members.go    # AddMemberUseCase, RemoveMemberUseCase
│   │   │   └── mark_read.go         # MarkReadUseCase
│   │   ├── message/
│   │   │   ├── send.go              # SendMessageUseCase
│   │   │   ├── list.go              # ListMessagesUseCase (pagination)
│   │   │   └── delete.go            # SoftDeleteMessageUseCase
│   │   ├── notification/
│   │   │   ├── list.go              # ListNotificationsUseCase
│   │   │   └── mark_read.go         # MarkNotificationReadUseCase
│   │   └── upload/
│   │       ├── avatar_presign.go    # CloudinaryPresignUseCase
│   │       └── file_presign.go      # R2PresignUseCase
│   │
│   ├── handler/                     # Delivery layer — HTTP + WebSocket
│   │   ├── http/
│   │   │   ├── router.go            # chi.Router setup, middleware mount
│   │   │   ├── auth_handler.go      # POST /auth/*
│   │   │   ├── user_handler.go      # PATCH /users/me, GET /users/search
│   │   │   ├── friend_handler.go    # /friends/*
│   │   │   ├── conversation_handler.go
│   │   │   ├── message_handler.go
│   │   │   ├── notification_handler.go
│   │   │   ├── upload_handler.go
│   │   │   └── middleware/
│   │   │       ├── auth.go          # JWT parse, inject userID vào context
│   │   │       ├── require_member.go
│   │   │       ├── require_admin.go
│   │   │       ├── rate_limiter.go
│   │   │       └── cors.go
│   │   └── ws/
│   │       ├── hub.go               # Hub: clients map, dispatch, ListenLoop
│   │       ├── client.go            # Client: conn, send chan, read/write pumps
│   │       ├── handler.go           # HTTP Upgrade → WS, auth check
│   │       └── events.go            # Event type constants + payload structs
│   │
│   ├── repository/                  # Infrastructure — implements domain interfaces
│   │   ├── postgres/
│   │   │   ├── db.go                # pgxpool.Pool init
│   │   │   ├── user_repo.go         # implements domain/user/repository.go
│   │   │   ├── friendship_repo.go
│   │   │   ├── conversation_repo.go
│   │   │   ├── message_repo.go
│   │   │   ├── notification_repo.go
│   │   │   └── refresh_token_repo.go
│   │   └── sqlc/                    # Auto-generated bởi sqlc
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── querier.go
│   │       └── *.sql.go             # Generated query functions
│   │
│   ├── pkg/                         # Shared utilities — không chứa business logic
│   │   ├── jwt/
│   │   │   └── jwt.go               # GenerateAccessToken, ParseToken
│   │   ├── password/
│   │   │   └── bcrypt.go            # Hash, Compare
│   │   ├── cloudinary/
│   │   │   └── client.go            # GenerateSignature, DeleteAsset
│   │   ├── r2/
│   │   │   └── client.go            # GeneratePresignedPutURL
│   │   ├── validator/
│   │   │   └── validator.go         # go-playground/validator wrapper
│   │   └── response/
│   │       └── response.go          # JSON envelope helpers: OK(), Err()
│   │
│   └── config/
│       └── config.go                # Load .env → Config struct (viper/godotenv)
│
├── db/
│   ├── migrations/                  # golang-migrate files
│   │   ├── 001_create_users.up.sql
│   │   ├── 001_create_users.down.sql
│   │   └── ...
│   └── queries/                     # sqlc input SQL files
│       ├── users.sql
│       ├── messages.sql
│       └── ...
│
├── tests/                           # Test suites
│   ├── integration/                 # API tests (httptest)
│   │   ├── auth_test.go
│   │   ├── friend_test.go
│   │   └── message_test.go
│   ├── unit/                        # UseCase tests (mock repo)
│   │   ├── auth_usecase_test.go
│   │   └── ...
│   └── fixtures/                    # Seed data, test helpers
│       └── seed.go
│
├── sqlc.yaml                        # sqlc config
├── Makefile                         # make migrate-up, make sqlc, make run, make test
├── Dockerfile
├── fly.toml
├── .env.example
└── go.mod
```

### Dependency Rule (Clean Architecture)

```
handler  →  usecase  →  domain  ←  repository
   ↓            ↓                        ↓
  pkg          pkg                      pkg
```

> Mũi tên = "được phép import". `domain` không biết `repository` tồn tại — chỉ định nghĩa interface. `repository` implement interface đó. `usecase` nhận interface qua constructor injection.

### Graceful Shutdown

`cmd/server/main.go` phải xử lý OS signals (`SIGTERM`, `SIGINT`) để:

1. **Đóng HTTP server** — đợi in-flight requests hoàn thành (timeout 10s)
2. **Đóng WS Hub** — gửi close frame đến tất cả connected clients
3. **Đóng `listenConn`** — hủy LISTEN/NOTIFY loop
4. **Đóng pgxpool** — giải phóng database connections
5. **Flush logger** — đảm bảo logs được ghi hết

```go
// main.go pseudo-code
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
<-quit
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
httpServer.Shutdown(ctx)
hub.Close()
listenConn.Close(ctx)
pool.Close()
```

---

## 3. Folder Structure — Frontend (Next.js 15)

> **Triết lý:** App Router + Server Components mặc định. Client Components (`'use client'`) chỉ khi cần interactivity hoặc WebSocket. State management: Zustand cho global state (auth, socket, notifications).

```
frontend/
├── app/                             # Next.js App Router
│   ├── layout.tsx                   # Root layout: font, providers, toaster
│   ├── page.tsx                     # Landing / redirect logic
│   │
│   ├── (auth)/                      # Route group — không có layout chung
│   │   ├── login/
│   │   │   └── page.tsx             # Login form (Client Component)
│   │   └── register/
│   │       └── page.tsx             # Register form (Client Component)
│   │
│   └── (app)/                       # Route group — yêu cầu đăng nhập
│       ├── layout.tsx               # App shell: Sidebar + Notification bell
│       ├── page.tsx                 # Redirect → /conversations
│       │
│       ├── conversations/
│       │   ├── page.tsx             # Danh sách conversations (Server Component)
│       │   └── [id]/
│       │       └── page.tsx         # Chat view (Client Component — cần WS)
│       │
│       ├── friends/
│       │   └── page.tsx             # Danh sách bạn bè + pending requests
│       │
│       └── settings/
│           └── page.tsx             # Cập nhật profile, avatar
│
├── components/
│   ├── ui/                          # shadcn/ui primitives (Button, Input, Avatar…)
│   │   └── ...
│   │
│   ├── auth/
│   │   ├── LoginForm.tsx
│   │   └── RegisterForm.tsx
│   │
│   ├── layout/
│   │   ├── Sidebar.tsx              # Conversation list sidebar
│   │   ├── NotificationBell.tsx     # Bell icon + dropdown panel
│   │   └── UserAvatar.tsx           # Avatar với online indicator
│   │
│   ├── chat/
│   │   ├── MessageList.tsx          # Virtualized message list (react-window)
│   │   ├── MessageBubble.tsx        # Single message: text, attachment, deleted state
│   │   ├── MessageInput.tsx         # Textarea + file attach + send button
│   │   ├── TypingIndicator.tsx      # "Bob đang gõ..."
│   │   ├── AttachmentPreview.tsx    # Image lightbox, file card
│   │   └── ConversationHeader.tsx   # Tên, avatar, online status, group members
│   │
│   ├── friends/
│   │   ├── FriendList.tsx
│   │   ├── FriendRequestCard.tsx    # Accept/Decline inline
│   │   └── UserSearchModal.tsx
│   │
│   └── notifications/
│       ├── NotificationItem.tsx     # Render theo type: FRIEND_REQUEST / GROUP_ADDED…
│       └── NotificationPanel.tsx
│
├── hooks/                           # Custom hooks
│   ├── useWebSocket.ts              # WS connect, reconnect, event dispatcher
│   ├── useMessages.ts               # Infinite scroll + optimistic update
│   ├── usePresence.ts               # Online/offline indicator
│   ├── useNotifications.ts          # Bell badge count, realtime push
│   └── useUpload.ts                 # Presign → upload → confirm flow
│
├── lib/
│   ├── api.ts                       # Axios/fetch instance, interceptors (token refresh)
│   ├── queryClient.ts               # TanStack Query client config
│   └── utils.ts                     # cn(), formatDate(), formatFileSize()…
│
├── services/                        # API call functions (dùng với TanStack Query)
│   ├── auth.service.ts
│   ├── friend.service.ts
│   ├── conversation.service.ts
│   ├── message.service.ts
│   ├── notification.service.ts
│   └── upload.service.ts
│
├── store/                           # Zustand global state
│   ├── auth.store.ts                # currentUser, access_token, setAuth, clearAuth
│   ├── socket.store.ts              # wsInstance, connectionStatus
│   └── notification.store.ts        # unreadCount, notifications[]
│
├── types/
│   ├── api.types.ts                 # Response envelope, shared types
│   ├── chat.types.ts                # Message, Conversation, Attachment
│   ├── user.types.ts                # User, Friendship
│   ├── notification.types.ts        # Notification, NotificationType
│   └── ws.types.ts                  # WS event payloads
│
├── middleware.ts                    # Next.js middleware: redirect nếu chưa login
├── next.config.ts
├── tailwind.config.ts
├── .env.local.example
└── package.json
```

### State & Data Flow (Frontend)

```
Server Component (RSC)              Client Component
       │                                   │
       │ fetch (server-side,               │ TanStack Query
       │ có cache)                         │ (client-side fetch + cache)
       ▼                                   ▼
  services/*.ts  ◄──────────────  lib/api.ts (Axios + refresh interceptor)
                                           │
                              Zustand ◄────┤◄──── useWebSocket.ts
                              (global)     │      (WS events → store update
                                           │       → Query invalidate)
```

---

## 4. Message Flow — User A gửi tin nhắn đến User B

> Luồng đầy đủ từ khi A nhấn Enter đến khi B thấy tin nhắn trên màn hình.

```
USER A (Browser)                    GOLANG SERVER                    USER B (Browser)
      │                                    │                                │
      │  1. Nhấn Enter                     │                                │
      │                                    │                                │
      │  2. Optimistic UI:                 │                                │
      │     render message tạm             │                                │
      │     với client_temp_id             │                                │
      │                                    │                                │
      │──3. WS: send_message ─────────────►│                                │
      │   { conversation_id,               │                                │
      │     body, client_temp_id }         │                                │
      │                                    │                                │
      │                          4. WS handler nhận event                   │
      │                             → gọi SendMessageUseCase                │
      │                                    │                                │
      │                          5. BEGIN TRANSACTION                       │
      │                             INSERT INTO messages                    │
      │                             UPDATE conversations                    │
      │                               SET last_message_id,                  │
      │                                   last_activity_at                  │
      │                             COMMIT                                  │
      │                                    │                                │
      │                          6. pg_notify('chat_events', {              │
      │                               type: "new_message",                  │
      │                               conversation_id: "...",               │
      │                               recipient_ids: [A, B],                │
      │                               data: { message object }              │
      │                             })                                      │
      │                                    │                                │
      │                          7. Hub.ListenLoop()                        │
      │                             nhận Notification từ Postgres           │
      │                             dispatch() theo recipient_ids           │
      │                                    │                                │
      │◄──8a. WS: message_sent ───────────│                                │
      │    { client_temp_id,              │──8b. WS: new_message ─────────►│
      │      message_id (real),           │    { message object,            │
      │      status: DELIVERED }          │      conversation_id }          │
      │                                   │                                 │
      │  9a. Replace optimistic           │                        9b. Append message
      │      message bằng real            │                            vào MessageList
      │      (khớp client_temp_id)        │                            Tăng unread_count
      │      Tick: ✓ (DELIVERED)          │                            nếu conversation
      │                                   │                            không đang focus
      │                                   │                                │
      │                         10. UPDATE messages                        │
      │                             SET status='DELIVERED'                 │
      │                             (khi B's WS ack nhận được)             │
      │                                   │                                │
      │                                   │         11. B mở conversation  │
      │                                   │◄── WS: mark_read ─────────────│
      │                                   │    { last_message_id }         │
      │                                   │                                │
      │                         12. UPDATE messages                        │
      │                             SET status='READ'                      │
      │                             UPDATE conversation_participants        │
      │                             SET last_read_message_id               │
      │                                   │                                │
      │◄──13. WS: read_receipt ──────────│                                │
      │    { reader_id: B,               │                                │
      │      last_read_message_id }      │                                │
      │                                  │                                │
      │  14. Update tick: ✓✓ (READ)      │                                │
```

### Ghi chú từng bước

| Bước | Layer | Chi tiết |
|------|-------|---------|
| 2 | Frontend | Optimistic UI render ngay — không chờ server. Message có class `opacity-70` và spinner nhỏ |
| 5 | UseCase | Transaction đảm bảo `messages` và `conversations.last_message_id` luôn đồng bộ |
| 6 | Repository | `pg_notify` gọi **sau khi COMMIT** — tránh notify trước khi data thật sự trong DB |
| 7 | WS Hub | `dispatch()` chạy trong goroutine riêng — `ListenLoop` không bao giờ bị block |
| 8a vs 8b | WS Hub | Hub gửi 2 event khác nhau: `message_sent` (chỉ A, kèm `client_temp_id`) và `new_message` (cả A lẫn B) |
| 9a | Frontend | `useMessages` hook tìm message có `client_temp_id` trùng → replace bằng server data |
| 10 | WS Hub | `DELIVERED` cập nhật khi WS frame tới B thành công (B online). Nếu B offline: giữ `SENT` |

---

## 5. Biến Môi Trường

### 5.1 Backend — `backend/.env`

```bash
# ─── Server ───────────────────────────────────────────────────────────────────
APP_ENV=development                        # development | production
APP_PORT=8080
APP_BASE_URL=http://localhost:8080

# ─── Database (Neon PostgreSQL) ───────────────────────────────────────────────
DATABASE_URL=postgresql://user:password@ep-xxx.neon.tech/chatdb?sslmode=require
DATABASE_MAX_CONNS=20                      # pgxpool max connections
DATABASE_MIN_CONNS=2
# Dedicated connection cho LISTEN (không dùng pool):
DATABASE_LISTEN_URL=postgresql://user:password@ep-xxx.neon.tech/chatdb?sslmode=require

# ─── JWT ──────────────────────────────────────────────────────────────────────
JWT_ACCESS_SECRET=your-256-bit-secret-here
JWT_ACCESS_TTL_SECONDS=900                 # 15 phút
JWT_REFRESH_TTL_DAYS=7

# ─── Cookies ──────────────────────────────────────────────────────────────────
COOKIE_SECURE=false                        # true ở production (HTTPS only)
COOKIE_DOMAIN=localhost
COOKIE_SAME_SITE=Strict

# ─── CORS ─────────────────────────────────────────────────────────────────────
CORS_ALLOWED_ORIGINS=http://localhost:3000

# ─── Cloudinary ───────────────────────────────────────────────────────────────
CLOUDINARY_CLOUD_NAME=your-cloud-name
CLOUDINARY_API_KEY=123456789012345
CLOUDINARY_API_SECRET=your-api-secret
CLOUDINARY_UPLOAD_PRESET=chat_avatars     # Upload preset trên Cloudinary dashboard

# ─── Cloudflare R2 ────────────────────────────────────────────────────────────
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key
R2_BUCKET_NAME=chat-files
R2_PUBLIC_URL=https://pub-xxx.r2.dev       # Custom domain hoặc R2 public URL
R2_PRESIGN_TTL_SECONDS=300                 # 5 phút

# ─── Rate Limiting ────────────────────────────────────────────────────────────
RATE_LIMIT_RPS=100                         # Request per second per IP
RATE_LIMIT_BURST=20
```

### 5.2 Backend — `backend/.env.production` (fly.io Secrets)

```bash
# Các biến giống .env nhưng override cho production:
APP_ENV=production
APP_BASE_URL=https://api.yourdomain.com
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://your-app.vercel.app
COOKIE_SECURE=true
COOKIE_DOMAIN=.yourdomain.com
COOKIE_SAME_SITE=None                      # None vì cross-origin (Vercel ↔ fly.io)

# DATABASE_URL trỏ đến Neon production branch
# Các secrets còn lại set qua: flyctl secrets set KEY=VALUE
```

> **Không commit file `.env.production`.** Set secrets trên fly.io bằng:
>
> ```bash
> flyctl secrets set JWT_ACCESS_SECRET="..." DATABASE_URL="..." -a your-app-name
> ```

---

### 5.3 Frontend — `frontend/.env.local`

```bash
# ─── API ──────────────────────────────────────────────────────────────────────
NEXT_PUBLIC_API_URL=http://localhost:8080/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/v1/ws

# ─── App ──────────────────────────────────────────────────────────────────────
NEXT_PUBLIC_APP_NAME=ChatApp
NEXT_PUBLIC_APP_URL=http://localhost:3000

# ─── Cloudinary (public — dùng để build URL phía client) ─────────────────────
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=your-cloud-name

# ─── Feature Flags (optional, dùng để bật/tắt tính năng theo env) ────────────
NEXT_PUBLIC_ENABLE_VIDEO_UPLOAD=false      # true khi R2 video sẵn sàng
```

### 5.4 Frontend — `frontend/.env.production` (Vercel Environment Variables)

```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/v1
NEXT_PUBLIC_WS_URL=wss://api.yourdomain.com/v1/ws
NEXT_PUBLIC_APP_URL=https://app.yourdomain.com
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=your-cloud-name
NEXT_PUBLIC_ENABLE_VIDEO_UPLOAD=true
```

> Set trên Vercel Dashboard → Project Settings → Environment Variables. Biến `NEXT_PUBLIC_*` sẽ được bundle vào client bundle — **không đặt secret vào đây**.

---

### 5.5 Bảng Tổng Hợp Biến Môi Trường

| Biến | Backend | Frontend | Public | Bắt buộc |
|------|---------|----------|--------|---------|
| `DATABASE_URL` | ✓ | ✗ | ✗ | ✓ |
| `DATABASE_LISTEN_URL` | ✓ | ✗ | ✗ | ✓ |
| `JWT_ACCESS_SECRET` | ✓ | ✗ | ✗ | ✓ |
| `CLOUDINARY_API_SECRET` | ✓ | ✗ | ✗ | ✓ |
| `R2_SECRET_ACCESS_KEY` | ✓ | ✗ | ✗ | ✓ |
| `CLOUDINARY_CLOUD_NAME` | ✓ | ✓ (public) | ✓ | ✓ |
| `NEXT_PUBLIC_API_URL` | ✗ | ✓ | ✓ | ✓ |
| `NEXT_PUBLIC_WS_URL` | ✗ | ✓ | ✓ | ✓ |
| `NEXT_PUBLIC_ENABLE_VIDEO_UPLOAD` | ✗ | ✓ | ✓ | ✗ |

> **Quy tắc vàng:** Bất kỳ biến nào có `SECRET`, `KEY`, `PASSWORD` trong tên → **không bao giờ** đặt vào frontend, không bao giờ commit lên git.

---

## 6. Key Packages

### Backend (Go)

| Package | Mục đích |
|---------|---------|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/jackc/pgx/v5` | PostgreSQL driver + LISTEN/NOTIFY |
| `github.com/jackc/pgx/v5/pgxpool` | Connection pool |
| `github.com/sqlc-dev/sqlc` | Generate type-safe Go từ SQL |
| `github.com/golang-jwt/jwt/v5` | JWT access token |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/gorilla/websocket` | WebSocket upgrade |
| `github.com/cloudinary/cloudinary-go/v2` | Cloudinary signature |
| `github.com/aws/aws-sdk-go-v2/service/s3` | R2 presigned URL (S3-compat) |
| `github.com/go-playground/validator/v10` | Input validation |
| `github.com/golang-migrate/migrate/v4` | DB migrations |
| `github.com/spf13/viper` | Config / .env loading |

### Frontend (Next.js)

| Package | Mục đích |
|---------|---------|
| `@tanstack/react-query` | Server state, caching, pagination |
| `zustand` | Global client state (auth, socket, notifs) |
| `axios` | HTTP client + refresh token interceptor |
| `shadcn/ui` + `tailwindcss` | UI components |
| `react-window` | Virtualized message list (performance) |
| `react-hook-form` + `zod` | Form + validation |
| `react-dropzone` | Drag & drop file upload |
| `dayjs` | Date formatting (lightweight moment.js) |
| `lucide-react` | Icons |

---

*Architecture document này là nguồn tham chiếu cho toàn bộ team khi onboard. Mọi thay đổi cấu trúc thư mục hoặc tech decision mới cần được cập nhật vào đây trước khi code.*
