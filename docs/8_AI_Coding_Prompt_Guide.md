# YunoChat — AI Coding Prompt Guide

> **Version:** 2.0.0 — Đã review-proof với docs v1.1.0
> **Tối ưu cho:** Claude 4.6 / Gemini 3.1 Pro (trong IDE agent mode)
> **Nguyên tắc:** Mỗi prompt tạo code có thể build + test được → commit Git → sang prompt tiếp

---

## ⚠️ Lỗi Trong Bộ Prompt Cũ — PHẢI SỬA

| # | Lỗi trong prompt gốc | Thực tế trong docs |
|---|----------------------|-------------------|
| 1 | ❌ Prompt 7 nói "Redis Pubsub" (`internal/redis/pubsub.go`, `online.go`) | ✅ Docs dùng **PostgreSQL `LISTEN/NOTIFY`** qua `pgx`. Không có Redis. Single instance. |
| 2 | ❌ Prompt 3 nói `internal/repository/user_repo.go` | ✅ Docs dùng `internal/repository/postgres/user_repo.go` (thêm subfolder `postgres/`) |
| 3 | ❌ Prompt 3 nói "raw queries với pgx" | ✅ Docs dùng **sqlc** (generate type-safe Go từ SQL). Repo gọi `sqlc.Queries`, không viết raw SQL trong Go |
| 4 | ❌ Prompt 3 nói `internal/utils/` | ✅ Docs dùng `internal/pkg/` (jwt, password, response, validator) |
| 5 | ❌ Prompt 4 nói `internal/service/auth_service.go` | ✅ Docs dùng `internal/usecase/auth/` (register.go, login.go, refresh.go, logout.go) |
| 6 | ❌ Prompt 4 nói `router/router.go` | ✅ Docs dùng `internal/handler/http/router.go` |
| 7 | ❌ Prompt 4 nói `internal/handler/middleware/auth.go` | ✅ Docs dùng `internal/handler/http/middleware/auth.go` |
| 8 | ❌ Prompt 5 FE nói `date-fns` | ✅ Docs dùng `dayjs`, và thiếu libs quan trọng: `@tanstack/react-query`, `react-hook-form`, `zod`, `react-window`, `react-dropzone`, `shadcn/ui` |
| 9 | ❌ Prompt 5 nói `lib/api.ts` | ✅ Docs dùng `lib/axios.ts` (interceptor + provider pattern) |
| 10 | ❌ Prompt 5 nói `store/auth-store.ts` | ✅ Docs dùng `store/auth.ts` |
| 11 | ❌ Prompt 9 nói `app/(app)/chat/[conversationId]/page.tsx` | ✅ Docs dùng `app/(app)/conversations/[id]/page.tsx` |
| 12 | ❌ Thiếu **hoàn toàn** prompt cho: Friend System, Group Chat, Notifications, Upload, Settings, Presence |
| 13 | ❌ Thiếu `internal/database/` | ✅ Docs dùng `internal/repository/postgres/db.go` + `internal/config/config.go` |
| 14 | ❌ Prompt 2 nói `2 file up.sql / down.sql` | ✅ Docs có **11 migration files** (000→010), mỗi file có pair up/down |

---

## Hướng Dẫn Sử Dụng

### Quy tắc vàng

```
1. MỖI PROMPT = 1 GIT COMMIT (build pass + test pass mới commit)
2. SAU MỖI PROMPT: review code AI viết 5-10 phút
3. NẾU BUG: copy log lỗi → paste cho AI → "Chỉ sửa dòng bị lỗi, KHÔNG rewrite file"
4. SAU 3 PROMPT: chạy go vet ./... và npx tsc --noEmit
```

### Anti-patterns cần tránh

```
❌ "Viết toàn bộ ứng dụng cho tôi"     → Quá rộng, AI sẽ hallucinate
❌ Bỏ qua review → commit              → Bug tích lũy, prompt sau build fail
❌ Chạy 5 prompt liên tục không test    → Đống code không chạy được
❌ Nói "làm giống Telegram"             → Mơ hồ, dùng docs thay vì assumption
✅ Luôn trỏ đến file docs cụ thể       → AI có source of truth
✅ Yêu cầu AI giải thích quyết định    → Bắt sớm nếu nó hiểu sai
```

---

## Giai Đoạn 1: Khởi Tạo Móng Backend (Golang)

### Chuẩn bị (bạn tự chạy)

```bash
mkdir backend && cd backend
go mod init backend
```

---

### Prompt 1 — Dựng Khung Thư Mục Backend

```
Đọc file docs/4_System_Architecture.md, phần "2. Folder Structure — Backend (Golang)".

Nhiệm vụ:
1. Tạo CHÍNH XÁC cây thư mục mô tả trong docs, bao gồm cả thư mục tests/.
2. Mỗi file .go chỉ chứa đúng `package <tên>` tương ứng — KHÔNG implement logic.
3. `go get` tất cả packages trong bảng "6. Key Packages → Backend (Go)" của cùng file docs.
4. Tạo file `backend/.env` theo template ở phần "5.1 Backend — backend/.env" (giữ placeholder values).
5. Tạo `backend/.env.example` copy từ `.env` nhưng XÓA các giá trị secret.
6. Tạo file `.gitignore` (bao gồm .env, .env.production, vendor/).
7. Thêm vào `.env` dòng `DATABASE_URL` và `DATABASE_LISTEN_URL` riêng biệt.
8. Tạo `sqlc.yaml` config cơ bản (input: db/queries/*.sql, output: internal/repository/sqlc/).

Lưu ý QUAN TRỌNG:
- Thư mục domain/ chia subdirectory: user/, friendship/, conversation/, message/, notification/
- Thư mục usecase/ (KHÔNG phải service/): auth/, friendship/, conversation/, message/, notification/, upload/
- Thư mục handler/ chia: http/ (chứa router.go + *_handler.go + middleware/) và ws/ (hub.go, client.go, handler.go, events.go)
- Thư mục repository/ chia: postgres/ và sqlc/
- Thư mục pkg/ (KHÔNG phải utils/): jwt/, password/, cloudinary/, r2/, validator/, response/
- Chạy `go mod tidy` ở cuối để verify không có import lỗi.

KHÔNG implement logic. Chỉ tạo skeleton files.
```

**✅ Verification:** `go build ./...` phải pass (vì chỉ có package declarations).

---

### Prompt 2 — Domain Entities & Database Migrations

```
Đọc file docs/2_Database_Design.md và docs/4_System_Architecture.md.

Nhiệm vụ:

A) DOMAIN ENTITIES — thư mục internal/domain/*/entity.go:
- Viết Go structs cho: User, Friendship, Conversation, ConversationParticipant, Message, Attachment, Notification, RefreshToken.
- Struct fields PHẢI khớp CHÍNH XÁC với cột trong bảng PostgreSQL (tên, kiểu dữ liệu).
- Dùng: uuid.UUID (google/uuid), time.Time cho TIMESTAMPTZ, *time.Time cho nullable timestamps.
- Thêm `json:"field_name"` và `db:"field_name"` tags đầy đủ.
- Trong mỗi thư mục domain/*, tạo thêm errors.go chứa domain errors (ErrUserNotFound, ErrDuplicateEmail, v.v.).

B) REPOSITORY INTERFACES — thư mục internal/domain/*/repository.go:
- Định nghĩa interface cho mỗi entity (UserRepository, FriendshipRepository, v.v.)
- Chỉ khai báo method signature (Create, FindByID, FindByEmail, v.v.) — KHÔNG implement.

C) CONFIG — file internal/config/config.go:
- Struct Config load từ .env bằng viper.
- Bao gồm: Server, Database, JWT, Cookie, CORS, Cloudinary, R2, RateLimit sub-structs.
- Func NewConfig() (*Config, error) đọc file .env.

D) DATABASE INIT — file internal/repository/postgres/db.go:
- Func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error)
- Config pgxpool với MaxConns, MinConns từ config.
- Return pool, KHÔNG tạo connection riêng cho LISTEN ở đây (sẽ làm trong WS prompt).

E) MIGRATIONS — thư mục db/migrations/:
- Tạo đúng các migration files theo thứ tự trong docs (000→010), mỗi file có pair .up.sql và .down.sql.
- ⚠️ Migration 000 = CREATE EXTENSION pg_trgm (PHẢI ĐẦU TIÊN, trước users table).
- Copy chính xác DDL từ docs, bao gồm indexes, constraints, triggers.
- Thêm Makefile targets: migrate-up, migrate-down, migrate-create.

Khi hoàn tất: chạy `go build ./...` để verify. KHÔNG implement repository.
```

**✅ Verification:** `go build ./...` pass. SQL migrations có thể `migrate -path db/migrations -database $DATABASE_URL up` thành công trên Neon test branch.

---

## Giai Đoạn 2: Lõi Tính Năng Auth

### Prompt 3 — Shared Packages (pkg)

```
Đọc docs/3_API_WebSocket_Specs.md phần 1 (Auth endpoints) và docs/7_Coding_Notes.md phần 4 (Token Refresh) và 10 (Password Security).

Implement các file trong internal/pkg/:

1. pkg/jwt/jwt.go:
   - GenerateAccessToken(userID uuid.UUID, secret string, ttl time.Duration) (string, error)
   - ParseToken(tokenString string, secret string) (uuid.UUID, error)
   - Dùng github.com/golang-jwt/jwt/v5. Claims chứa: sub (userID), exp, iat.

2. pkg/password/bcrypt.go:
   - Hash(password string) (string, error) — bcrypt cost=12
   - Compare(hashedPassword, password string) error
   - Tạo DummyHash constant (dùng cho timing attack prevention — xem docs/7_Coding_Notes.md phần 10)

3. pkg/response/response.go:
   - Chuẩn hóa JSON response theo format trong docs/3 phần 1.4:
     OK(w, statusCode, data interface{})
     Err(w, statusCode, code string, message string)
   - Format: {"success": true/false, "data": {...}, "error": {"code": "...", "message": "..."}}

4. pkg/validator/validator.go:
   - Wrapper quanh go-playground/validator/v10
   - Func ValidateStruct(s interface{}) map[string]string (trả field → error message)

Chạy go build ./... sau khi xong. KHÔNG viết test ở bước này (sẽ viết test sau khi có full flow).
```

---

### Prompt 4 — sqlc Queries & Repository (Auth)

```
Đọc docs/2_Database_Design.md (bảng users, refresh_tokens) và docs/4_System_Architecture.md (cấu trúc sqlc).

Nhiệm vụ:

A) SQLC QUERIES — tạo file db/queries/users.sql và db/queries/refresh_tokens.sql:

users.sql:
  -- name: CreateUser :one
  INSERT INTO users (id, email, username, display_name, password_hash, status)
  VALUES ($1, $2, $3, $4, $5, 'ONLINE')
  RETURNING *;

  -- name: FindUserByEmail :one
  SELECT * FROM users WHERE email = $1;

  -- name: FindUserByUsername :one
  SELECT * FROM users WHERE username = $1;

  -- name: FindUserByID :one
  SELECT * FROM users WHERE id = $1;

  -- name: UpdateUserStatus :exec
  UPDATE users SET status = $1, last_seen_at = NOW() WHERE id = $2;

refresh_tokens.sql:
  -- name: CreateRefreshToken :one
  INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
  VALUES ($1, $2, $3, $4) RETURNING *;

  -- name: FindRefreshTokenByHash :one
  SELECT * FROM refresh_tokens WHERE token_hash = $1 AND is_revoked = false AND expires_at > NOW();

  -- name: RevokeRefreshToken :exec
  UPDATE refresh_tokens SET is_revoked = true WHERE id = $1;

  -- name: RevokeAllUserTokens :exec
  UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1;

B) Chạy `sqlc generate` để tạo Go code trong internal/repository/sqlc/.

C) REPOSITORY IMPLEMENTATION — internal/repository/postgres/:
  - user_repo.go: implements domain/user/UserRepository, dùng sqlc.Queries.
  - refresh_token_repo.go: implements RefreshTokenRepository.
  - Repos nhận *sqlc.Queries qua constructor injection (NewUserRepo(q *sqlc.Queries)).

Chạy go build ./... để verify.
```

---

### Prompt 5 — UseCase & Handler (Auth)

```
Đọc docs/3_API_WebSocket_Specs.md phần 1 (Auth endpoints 1.1→1.6).
Đọc docs/7_Coding_Notes.md phần 4 (Token Refresh) và 10 (Password Security).

Nhiệm vụ:

A) USECASE — thư mục internal/usecase/auth/:
  - register.go: RegisterUseCase — validate input, check duplicate email/username, hash password, insert user, generate tokens.
  - login.go: LoginUseCase — find by email, compare password (⚠️ dùng DummyHash nếu email không tồn tại — chống timing attack), generate tokens.
  - refresh.go: RefreshTokenUseCase — validate cookie, find token, rotation (revoke cũ + tạo mới), detect reuse → revoke ALL.
  - logout.go: LogoutUseCase — revoke refresh token, update user status OFFLINE.
  - Mỗi UseCase là struct nhận repository interface qua constructor.

B) HANDLER — internal/handler/http/auth_handler.go:
  - Handler struct nhận UseCase interface (KHÔNG nhận repo trực tiếp).
  - POST /v1/auth/register → gọi RegisterUseCase
  - POST /v1/auth/login → gọi LoginUseCase → set cookie refresh_token (HttpOnly, Secure, SameSite)
  - POST /v1/auth/refresh → gọi RefreshTokenUseCase
  - POST /v1/auth/logout → gọi LogoutUseCase
  - GET /v1/users/me → trả profile user hiện tại
  - PATCH /v1/users/me → update display_name, bio
  - Response PHẢI dùng pkg/response envelope format.

C) MIDDLEWARE — internal/handler/http/middleware/auth.go:
  - RequireAuth middleware: parse Bearer token từ Authorization header, inject userID vào context.
  - CORS middleware tuỳ config.
  - Rate limiter middleware.

D) ROUTER — internal/handler/http/router.go:
  - Dùng chi/v5 router.
  - Mount middleware stack: Logger → CORS → RateLimiter.
  - Group /v1/auth/* (public routes, không cần RequireAuth).
  - Group /v1/* (protected routes, RequireAuth).

E) MAIN — cmd/server/main.go:
  - Load config → init DB pool → init repos → init usecases → init handlers → init router → start HTTP server.
  - ⚠️ Implement graceful shutdown (xem docs/4 phần Graceful Shutdown): OS signal → shutdown HTTP → close pool.
  - Wire DI thủ công (không dùng framework DI).

Chạy go build ./... → go run cmd/server/main.go. Server phải start trên port 8080.
```

**✅ Verification:** `curl -X POST localhost:8080/v1/auth/register -d '{"email":"test@test.com","username":"test_user","password":"Test@1234","display_name":"Test"}'` → 201. Check DB có row mới.

---

## Giai Đoạn 3: Khởi Tạo Móng Frontend (Next.js)

### Chuẩn bị (bạn tự chạy)

```bash
cd ..   # thoát khỏi backend/
npx create-next-app@latest frontend --typescript --tailwind --app --eslint --src-dir=false --import-alias="@/*" --use-npm
cd frontend
```

---

### Prompt 6 — Setup Frontend Base

```
Đọc docs/4_System_Architecture.md phần 3 (Frontend folder structure) và phần 6 (Key Packages → Frontend).

Nhiệm vụ:

A) CÀI ĐẶT PACKAGES (npm install):
  - @tanstack/react-query, zustand, axios, dayjs, lucide-react
  - react-hook-form, zod, @hookform/resolvers
  - react-window, @types/react-window
  - react-dropzone
  - Cài shadcn/ui: npx shadcn@latest init (chọn dark theme, slate color)

B) TẠO CÂY THƯ MỤC theo docs:
  - components/ (auth/, chat/, friends/, layout/, notifications/)
  - hooks/ (useWebSocket.ts, useMessages.ts, usePresence.ts, useNotifications.ts, useUpload.ts)
  - store/ (auth.ts, socket.ts, notification.ts)
  - lib/ (axios.ts, utils.ts)
  - types/ (api.ts, ws.ts)
  - services/ (auth.ts, friend.ts, conversation.ts, message.ts, notification.ts, upload.ts)
  - Mỗi file chỉ cần export trống hoặc comment "// TODO".

C) AXIOS INSTANCE — lib/axios.ts:
  - Base URL từ env NEXT_PUBLIC_API_URL.
  - Request interceptor: đọc access_token từ Zustand auth store → set Authorization header.
  - Response interceptor: bắt 401 → gọi refresh (queue pattern, xem docs/7_Coding_Notes.md phần 4) → retry request.
  - Nếu refresh cũng fail → clear store, redirect /login.

D) AUTH STORE — store/auth.ts:
  - Zustand store: { user, accessToken, setAuth, clearAuth, isAuthenticated }
  - User type match domain entity (id, email, username, display_name, bio, avatar_url, status)
  - persist middleware với sessionStorage.

E) TYPES — types/api.ts:
  - ApiResponse<T> = { success: boolean; data?: T; error?: { code: string; message: string } }
  - User, Conversation, Message, Notification types.

F) ENV — tạo .env.local theo docs/4 phần 5.3.

Chạy `npm run dev` → app phải chạy trên localhost:3000 không lỗi.
```

---

### Prompt 7 — Giao Diện Auth

```
Đọc docs/5_UI_Wireframes.md (phần 2.1 Login, 2.2 Register, 1.1→1.4 Design System).
Đọc docs/3_API_WebSocket_Specs.md phần 1 (Auth API contract).
Đọc docs/6_Use_Cases_Test_Cases.md phần 2 (Auth test cases — để hiểu UX flows).

Nhiệm vụ:

A) LOGIN FORM — components/auth/LoginForm.tsx:
  - 'use client', dùng react-hook-form + zod validation.
  - Fields: email, password.
  - Submit → gọi POST /v1/auth/login → set Zustand store → router.push('/conversations').
  - Error handling: hiện global error box (ĐỎ) phía trên form. Cùng 1 message cho sai email / sai password.
  - Loading: spinner trên button, disable button khi submitting.
  - Tuân thủ dark theme color palette (Surface=slate-900, Primary=indigo-600).

B) REGISTER FORM — components/auth/RegisterForm.tsx:
  - Fields: email, username, display_name, password, confirm_password.
  - Zod schema: email format, username chỉ a-z0-9_, password min 8 chars, confirm match.
  - Inline validation errors dưới mỗi field.
  - Submit → POST /v1/auth/register → auto login → redirect /conversations.

C) PAGES:
  - app/(auth)/login/page.tsx — render LoginForm.
  - app/(auth)/register/page.tsx — render RegisterForm.
  - Link navigate giữa login ↔ register.

D) MIDDLEWARE — middleware.ts (root frontend):
  - Protected routes: /conversations/*, /friends/*, /settings/*
  - Nếu không có access_token (check cookie hoặc header) → redirect /login.
  - Public routes: /login, /register.

Design PHẢI theo wireframe: dark theme, centered card, rounded-lg, slate-900 background.
```

**✅ Verification:** Mở browser, register user → auto redirect. Login → xem conversations page (trống). Inspect Network: tokens đúng.

---

## Giai Đoạn 4: Chat & WebSocket (Phần Khó Nhất)

### Prompt 8 — Backend WebSocket Hub & pg LISTEN/NOTIFY

```
Đọc docs/4_System_Architecture.md phần 4 (Message Flow Realtime).
Đọc docs/3_API_WebSocket_Specs.md phần 6 (Connection), 8 (Client→Server Events), 9 (Server→Client Events).
Đọc docs/7_Coding_Notes.md phần 2 (pg_notify limit), 6 (WS Connection Management), 11 (Presence Grace Period).

⚠️ LƯU Ý: Dự án này dùng PostgreSQL LISTEN/NOTIFY, KHÔNG dùng Redis.

Nhiệm vụ:

A) WS HUB — internal/handler/ws/hub.go:
  - Hub struct: clients map[uuid.UUID][]*Client, register/unregister channels, gracePeriods map.
  - ListenLoop(): tạo DEDICATED pgx connection (TÁCH BIỆT khỏi pgxpool), LISTEN 'chat_events'.
  - dispatch(): parse notification payload → tìm recipient clients → gửi qua client.send channel.
  - ⚠️ listenConn phải là connection RIÊNG (xem docs/7 phần 6 giải thích tại sao).
  - ⚠️ Handle payload > 7500 bytes: chỉ gửi message_id, Hub tự query DB (xem docs/7 phần 2).

B) WS CLIENT — internal/handler/ws/client.go:
  - Client struct: conn *websocket.Conn, userID uuid.UUID, send chan []byte.
  - readPump(): đọc message từ WS, parse event type, route to handler.
  - writePump(): đọc từ send channel, write to WS conn.
  - Ping/Pong: server gửi ping mỗi 30s, expect pong trong 10s, không nhận → close connection.

C) WS HANDLER — internal/handler/ws/handler.go:
  - HTTP upgrade endpoint: GET /v1/ws?token=<JWT>
  - Parse JWT từ query param (KHÔNG dùng Authorization header cho WS).
  - Upgrade → tạo Client → register vào Hub.
  - Gửi welcome event: { event: "connected", payload: { user_id, server_time } }

D) WS EVENTS — internal/handler/ws/events.go:
  - Client→Server: ping, join_conversation, leave_conversation, send_message, typing_start, typing_stop, mark_read.
  - Server→Client: pong, new_message, message_sent, user_typing, presence_update, notification_new, conversation_updated, member_removed, message_deleted.

E) PRESENCE — Grace Period Pattern:
  - Khi client disconnect: kiểm tra có connection khác không (multi-tab).
  - Nếu không: bắt đầu 60s timer. Sau 60s → broadcast OFFLINE.
  - Khi client reconnect trong 60s → cancel timer (xem docs/7 phần 11).
  - Support nhiều connections từ cùng userID (multi-tab).

F) INTEGRATE — đăng ký WS handler vào router, khởi tạo Hub trong main.go.
  - Graceful shutdown phải đóng Hub + listenConn.

Chạy backend → test WS bằng wscat: wscat -c "ws://localhost:8080/v1/ws?token=<JWT>" → nhận "connected" event.
```

---

### Prompt 9 — Backend Message & Conversation API

```
Đọc docs/3_API_WebSocket_Specs.md phần 4 (Conversations), 5 (Messages).
Đọc docs/2_Database_Design.md phần 4.3 (Query Patterns — cursor pagination).
Đọc docs/7_Coding_Notes.md phần 1 (Concurrency), 7 (Soft Delete), 8 (Cursor Pagination).

Nhiệm vụ:

A) SQLC QUERIES — db/queries/:
  conversations.sql:
  - ListConversationsByUser (JOIN conversation_participants, ORDER BY last_activity_at DESC, cursor pagination)
  - FindDMConversation (tìm DM giữa 2 users — dùng cho reuse khi re-friend)
  - CreateConversation, CreateParticipant
  - UpdateLastActivity, UpdateLastMessage

  messages.sql:
  - CreateMessage (INSERT + pg_notify trigger)
  - ListMessages (WHERE conversation_id, cursor (created_at, id), LIMIT, ORDER DESC)
  - SoftDeleteMessage (SET deleted_at=NOW(), body=NULL — xem docs/7 phần 7)

B) SQLC GENERATE → Chạy sqlc generate.

C) REPOSITORY — internal/repository/postgres/:
  - conversation_repo.go
  - message_repo.go

D) USECASE:
  - conversation/list.go: ListConversationsUseCase (cursor pagination)
  - conversation/create_group.go: CreateGroupUseCase (min 3 participants, creator = ADMIN)
  - conversation/mark_read.go: MarkReadUseCase (update last_read_at trên conversation_participants)
  - message/send.go: SendMessageUseCase
    ⚠️ PHẢI dùng DB transaction:
    1. Check friendship status (nếu DM)
    2. INSERT message
    3. UPDATE conversations.last_message_id + last_activity_at
    4. COMMIT
    5. pg_notify SAU commit (không trước!)
  - message/list.go: ListMessagesUseCase
  - message/delete.go: SoftDeleteUseCase (chỉ sender mới được xóa tin nhắn của mình)

E) HANDLER — internal/handler/http/:
  - conversation_handler.go: GET /v1/conversations, POST /v1/conversations/groups, PATCH /v1/conversations/:id
  - message_handler.go: GET /v1/conversations/:id/messages, POST /v1/conversations/:id/messages, DELETE /v1/messages/:id

Chạy go build. Test bằng Postman: tạo conversation → gửi message → list messages.
```

---

### Prompt 10 — Frontend Chat UI & WebSocket Hook

```
Đọc docs/5_UI_Wireframes.md phần 3 (Conversation View), phần 1.4 (Mobile Responsive Strategy).
Đọc docs/7_Coding_Notes.md phần 3 (Optimistic UI), 8 (Cursor Pagination), 12 (Typing Indicator).
Đọc docs/3_API_WebSocket_Specs.md phần 8-9 (WS events).

Nhiệm vụ:

A) WS HOOK — hooks/useWebSocket.ts:
  - Kết nối: ws://API_URL/v1/ws?token=<accessToken>
  - Auto-reconnect: exponential backoff 1s → 2s → 4s → ... → max 30s.
  - Sau reconnect: invalidate TanStack Query cache, re-join active conversations.
  - Dispatch events to Zustand stores (notification, presence).
  - Lắng nghe: new_message, message_sent, user_typing, presence_update, notification_new.

B) CONVERSATION LIST — app/(app)/conversations/page.tsx:
  - TanStack useQuery cho API GET /v1/conversations.
  - Render danh sách trong Sidebar — reuse component trong components/layout/Sidebar.tsx.
  - Mỗi item: avatar, name, last message preview, timestamp, unread badge.
  - Mobile: full-width. Desktop: 320px sidebar.

C) CHAT VIEW — app/(app)/conversations/[id]/page.tsx:
  - TanStack useInfiniteQuery cho messages (cursor pagination).
  - Scroll up → load more (prepend cũ, giữ scroll position — xem docs/7 phần 8).
  - Append realtime khi nhận new_message qua WS.

D) MESSAGE LIST — components/chat/MessageList.tsx:
  - Virtualized rendering bằng react-window (performance khi nhiều messages).
  - Date dividers giữa các ngày.
  - MessageBubble: indigo-600 cho tin mình gửi, slate-700 cho tin người khác.
  - Bubble max-width: 70% desktop, 85% mobile.
  - Tick status: ✓ sent (slate-400), ✓✓ read (indigo-400).
  - Typing indicator ở dưới cùng.

E) MESSAGE INPUT — components/chat/MessageInput.tsx:
  - Optimistic UI pattern (xem docs/7 phần 3):
    1. User nhấn Enter → tạo client_temp_id (UUID v4)
    2. Render message ngay (opacity-60, spinner)
    3. Gửi WS send_message
    4. Nhận ack message_sent → match client_temp_id → replace
    5. Timeout 10s → hiện retry button
  - Typing indicator: throttle 3s (chỉ gửi typing_start 1 lần / 3s).
  - File attach button (📎) — bước này chỉ UI, logic upload ở prompt sau.
  - Sticky bottom, handle keyboard trên mobile (visualViewport).

F) CONVERSATION HEADER — components/chat/ConversationHeader.tsx:
  - DM: avatar + name + online status dot.
  - Group: group avatar + name + member count.
  - Mobile: nút ← back (router.back()).

Chạy npm run dev → login → chọn conversation → gửi tin nhắn → xem realtime trên 2 tab browser.
```

---

## Giai Đoạn 5: Friend System & Notifications

### Prompt 11 — Backend Friend System

```
Đọc docs/3_API_WebSocket_Specs.md phần 3 (Friends endpoints 3.1→3.5).
Đọc docs/7_Coding_Notes.md phần 1 (Concurrency — Accept + Unfriend race conditions).
Đọc docs/2_Database_Design.md (bảng friendships, idx_friendships_canonical).

Nhiệm vụ:

A) SQLC QUERIES — db/queries/friendships.sql, notifications.sql
B) REPOSITORY — friendship_repo.go, notification_repo.go
C) USECASE — internal/usecase/friendship/:
  - send_request.go: ⚠️ Handle "2 user gửi request đồng thời" (unique constraint → 409)
  - respond_request.go: Accept = 1 DB TRANSACTION: update friendship + check/create DM conversation + insert notification + pg_notify
  - unfriend.go: Hard delete friendship, KHÔNG set conversation_participants.left_at, block new messages ở API layer

D) USECASE — internal/usecase/notification/:
  - list.go, mark_read.go

E) HANDLER — friend_handler.go, notification_handler.go
  - POST /v1/friends/requests
  - PATCH /v1/friends/requests/:id (body: { action: "ACCEPT" | "DECLINE" })
  - DELETE /v1/friends/:id (unfriend)
  - GET /v1/friends (list friends)
  - GET /v1/friends/requests/received (pending)
  - GET /v1/friends/requests/sent (sent — để user hủy)
  - GET /v1/users/search?q=... (tìm kiếm user, trả relationship status)
  - GET /v1/notifications, PATCH mark read, PATCH mark all read

⚠️ Tham chiếu docs/1_MVP_Requirements.md: FRD-01 search phải trả relationship status.
```

---

### Prompt 12 — Frontend Friends & Notifications UI

```
Đọc docs/5_UI_Wireframes.md phần 3.5 (Friend List), 3.4 (User Search Modal), 3.3 (Create Group), 3.6 (Notification Panel).
Đọc docs/6_Use_Cases_Test_Cases.md phần 3 (Friend test cases).

Nhiệm vụ:

A) SIDEBAR — components/layout/Sidebar.tsx:
  - Header: avatar user + tên + bell icon (NotificationBell).
  - Search bar: click → mở UserSearchModal.
  - Tabs: [Chats] [Friends] (controlled state).
  - Tab Chats: conversation list items.
  - Tab Friends: friend list + pending requests.

B) USER SEARCH MODAL — components/friends/UserSearchModal.tsx:
  - Full-screen sheet trên mobile, modal trên desktop.
  - Debounce 300ms input → GET /v1/users/search.
  - Mỗi kết quả hiển thị relationship status:
    - Stranger → button "+ Kết bạn"
    - Pending (sent) → "Đã gửi lời mời ↩"
    - Pending (received) → "Chấp nhận / Từ chối"
    - Friend → "💬 Nhắn tin"

C) NOTIFICATION PANEL — components/notifications/NotificationPanel.tsx:
  - Dropdown trên desktop, full-screen overlay trên mobile.
  - 3 types: FRIEND_REQUEST (có Accept/Decline buttons), FRIEND_ACCEPTED, NEW_MESSAGE.
  - Badge count từ Zustand notification store (cập nhật qua WS notification_new).
  - "Đánh dấu tất cả đã đọc" button.

D) CREATE GROUP MODAL — components/friends/CreateGroupModal.tsx:
  - Input tên nhóm + multi-select friends (min 2 thành viên khác).
  - Button "Tạo nhóm" disabled khi tên rỗng hoặc < 2 members.

Chạy dev → test: search user → gửi friend request → accept → DM tự tạo.
```

---

## Giai Đoạn 6: File Upload & Polish

### Prompt 13 — Backend File Upload (Presigned URLs)

```
Đọc docs/3_API_WebSocket_Specs.md phần 7 (Upload endpoints).
Đọc docs/7_Coding_Notes.md phần 5 (File Upload Safety).

Implement:
- pkg/cloudinary/client.go: GenerateSignature cho avatar upload.
- pkg/r2/client.go: GeneratePresignedPutURL cho file attachments.
- usecase/upload/: avatar_presign.go, file_presign.go.
- handler/http/upload_handler.go:
  - POST /v1/upload/avatar/presign → Cloudinary signature (public_id = avatars/{user_id})
  - POST /v1/upload/file/presign → R2 presigned PUT URL (expires 5 phút)
- ⚠️ Server-side validation: MIME type + file size (defense in depth, KHÔNG tin client).
```

### Prompt 14 — Frontend Upload & Attachments

```
Đọc docs/5_UI_Wireframes.md phần 3.2 (Message Bubbles — Attachment variants), 4.5 (Settings Page), 4.6 (Image Lightbox).

Implement:
- hooks/useUpload.ts: presign → upload file → confirm → return URL.
- MessageInput tích hợp file attach (📎 button + react-dropzone).
- AttachmentPreview: thumbnail inline cho ảnh, file card cho PDF/ZIP.
- Image Lightbox: full-screen overlay, pinch-to-zoom mobile, swipe to dismiss.
- Settings page: app/(app)/settings/page.tsx — update display_name, bio, avatar upload.
- Client-side validation: MIME type whitelist + file size limits (10MB image, 50MB file).
```

### Prompt 15 — Final Integration & Edge Cases

```
Đọc docs/7_Coding_Notes.md TOÀN BỘ.
Đọc docs/6_Use_Cases_Test_Cases.md phần Edge Cases.

Review và sửa:
1. Graceful shutdown: verify Hub.Close() gửi close frame đến tất cả clients.
2. Multi-tab: test mở 2 tab, gửi tin → cả 2 nhận.
3. Unfriend flow: unfriend → conversation read-only → re-friend → reuse DM cũ.
4. Group chat: admin kick → member_removed WS event → UI update.
5. Token refresh race: 3 requests 401 cùng lúc → chỉ 1 lần refresh (queue pattern).
6. Presence: close tab → 60s → OFFLINE. Reconnect trong 60s → cancel timer.
7. Soft delete message: body = NULL, UI hiện "Tin nhắn đã bị xóa".
8. Responsive: test toàn bộ flow trên mobile viewport (375px width).

Chạy: go vet ./... && go test ./... && npm run build (frontend).
```

---

## Bảng Tóm Tắt

| # | Prompt | Output chính | Verification |
|---|--------|-------------|-------------|
| 1 | Skeleton Backend | Cây thư mục + go.mod | `go build ./...` |
| 2 | Domain + Migrations | Entities + DDL files | `migrate up` on Neon |
| 3 | Shared Packages | jwt, password, response | `go build ./...` |
| 4 | sqlc + Auth Repo | SQL queries + repo impl | `sqlc generate` + build |
| 5 | Auth UseCase + API | Full auth flow | `curl register/login` |
| 6 | Frontend Base | Folder + axios + store | `npm run dev` |
| 7 | Auth UI | Login/Register forms | Browser test |
| 8 | WS Hub | pg LISTEN/NOTIFY realtime | `wscat` test |
| 9 | Message API | CRUD messages + cursor | Postman |
| 10 | Chat UI | Optimistic messages + WS | 2 tab browser test |
| 11 | Friend System | Backend full friend flow | curl test |
| 12 | Friend/Notif UI | Sidebar, search, notifs | Browser test |
| 13 | Upload Backend | Presigned URLs | Postman |
| 14 | Upload Frontend | File attach + Lightbox | Browser test |
| 15 | Integration | Edge cases + polish | Full E2E |

---

*Bộ prompt này là nguồn sự thật. Mọi thay đổi cần cập nhật lại trước khi dùng.*
