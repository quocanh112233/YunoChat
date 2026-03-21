# 7. Coding Notes — Lưu Ý Quan Trọng Khi Implement

> **Version:** 1.0.0
> **Role:** Senior Backend Engineer & Frontend Architect
> **Dựa trên:** Tất cả design docs `1_MVP_Requirements.md` → `6_Use_Cases_Test_Cases.md`
> **Mục đích:** Tổng hợp các gotchas, traps, và best practices cần nhớ khi code — tránh bugs khó debug.

---

## 1. Concurrency & Race Conditions

### 1.1 Accept Friend Request — Transaction Boundary

**❌ SAI:**

```go
// Mỗi operation là 1 query riêng → race condition
repo.UpdateFriendship(ctx, id, "ACCEPTED")
repo.InsertConversation(ctx, convID, "DM")
repo.InsertParticipants(ctx, convID, userA, userB)
repo.InsertNotification(ctx, userA, "FRIEND_ACCEPTED")
```

**✅ ĐÚNG:**

```go
// Tất cả trong 1 DB transaction
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)

queries := sqlc.New(tx)
queries.UpdateFriendshipStatus(ctx, id, "ACCEPTED")
existingConv := queries.FindDMConversation(ctx, userA, userB)  // check reuse
if existingConv == nil {
    queries.InsertConversation(ctx, convID, "DM")
    queries.InsertParticipant(ctx, convID, userA, "MEMBER")
    queries.InsertParticipant(ctx, convID, userB, "MEMBER")
}
queries.InsertNotification(ctx, userA, "FRIEND_ACCEPTED")

tx.Commit(ctx)
```

### 1.2 Hai User Gửi Friend Request Cho Nhau Đồng Thời

- `idx_friendships_canonical` unique index sử dụng `LEAST(requester_id, addressee_id), GREATEST(...)` → đảm bảo chỉ 1 row tồn tại cho mỗi cặp user
- Khi cả A và B gửi request cùng lúc: 1 INSERT thành công, 1 bị unique constraint violation → trả 409
- ⚠️ Backend phải handle `duplicate key value violates unique constraint` error gracefully

### 1.3 Unfriend → Gửi Message Race Condition

- User A unfriend B ở tab 1, nhưng tab 2 vẫn đang mở conversation
- `send_message` handler PHẢI check `friendships.status = ACCEPTED` trước khi INSERT message
- Check này PHẢI ở cùng transaction với INSERT message (hoặc dùng `SELECT FOR UPDATE`)

---

## 2. PostgreSQL NOTIFY Payload Limit

```
⚠️ pg_notify() payload bị giới hạn 8000 bytes
```

**Xử lý khi payload lớn:**

```go
payload := buildPayload(event)
if len(payload) > 7500 { // safety margin
    // Chỉ gửi message ID, Hub tự query DB
    payload = fmt.Sprintf(`{"type":"new_message","message_id":"%s","conversation_id":"%s"}`, msgID, convID)
}
_, err := pool.Exec(ctx, "SELECT pg_notify('chat_events', $1)", payload)
```

**Khi nào payload có thể vượt 8KB?**
- Message body rất dài (> 4000 ký tự Unicode)
- Attachment metadata chứa URL dài
- Group chat có nhiều recipient_ids

---

## 3. Optimistic UI — Message Sending Pattern

### Frontend Flow (MessageInput.tsx → MessageList.tsx)

```
1. User gõ + nhấn Enter
2. Frontend tạo client_temp_id = uuid()
3. Render message ngay lập tức (opacity-60, spinner "◌ Đang gửi...")
4. Gửi WS: { event: "send_message", payload: { ..., client_temp_id } }
5. Nhận WS ack: { event: "message_sent", payload: { client_temp_id, message_id, status: "SENT" } }
6. Match client_temp_id → replace optimistic message với real data
7. Nếu timeout 10s không có ack → hiện retry button
```

### Lưu ý:
- `client_temp_id` PHẢI là UUID v4, tạo ở client
- Zustand store: dùng `Map<client_temp_id, OptimisticMessage>` để lookup O(1)
- Khi nhận `new_message` từ WS mà sender === current user: **bỏ qua** (đã có message từ optimistic)
- Clear input NGAY khi user nhấn Enter, không đợi ack

---

## 4. Token Refresh Interceptor — Anti-Patterns

### Queue Pattern (Prevent Parallel Refresh)

```typescript
// ❌ SAI: Mỗi 401 gọi refresh riêng → 3 requests, 3 lần refresh
axios.interceptors.response.use(null, async (error) => {
    if (error.response.status === 401) {
        await refreshToken();  // Race condition!
        return axios(error.config);
    }
});

// ✅ ĐÚNG: Chỉ 1 refresh, các request khác đợi
let refreshPromise: Promise<string> | null = null;

axios.interceptors.response.use(null, async (error) => {
    if (error.response.status === 401) {
        if (!refreshPromise) {
            refreshPromise = refreshToken().finally(() => { refreshPromise = null; });
        }
        const newToken = await refreshPromise;
        error.config.headers.Authorization = `Bearer ${newToken}`;
        return axios(error.config);
    }
});
```

### Edge Cases:
- **Không retry `/auth/refresh`**: Nếu refresh cũng 401 → redirect `/login`, clear Zustand store
- **WebSocket reconnect**: Sau khi refresh thành công, WS cần reconnect với new token
- **Multiple tabs**: Mỗi tab gọi refresh riêng — OK vì refresh token rotation chỉ revoke token cũ

---

## 5. File Upload Safety Checklist

### Client-Side (`useUpload` hook):

```typescript
const ALLOWED_IMAGES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
const ALLOWED_FILES = ['application/pdf', 'video/mp4', 'application/zip', /* ... */];
const MAX_IMAGE_SIZE = 10 * 1024 * 1024;   // 10MB (Cloudinary)
const MAX_FILE_SIZE = 50 * 1024 * 1024;    // 50MB (R2)

// Validate TRƯỚC khi gọi presign API
if (!ALLOWED_IMAGES.includes(file.type) && !ALLOWED_FILES.includes(file.type)) {
    toast.error('Định dạng file không được hỗ trợ');
    return;
}
if (file.size > MAX_FILE_SIZE) {
    toast.error(`File tối đa 50MB. File của bạn: ${(file.size / 1024 / 1024).toFixed(1)} MB`);
    return;
}
```

### Server-Side (Defense in Depth):

- Server CŨNG validate MIME type + size — **không tin client**
- Presigned URL hết hạn **5 phút** → client cần xin URL mới nếu upload chậm
- Avatar upload: `public_id = avatars/{user_id}` → Cloudinary tự overwrite ảnh cũ (không cần cleanup)

---

## 6. WebSocket Connection Management

### Server Hub — `listenConn`

```
⚠️ listenConn là DEDICATED connection, TÁCH BIỆT khỏi pgxpool
   LISTEN sẽ block connection → không trả về pool
```

```go
// ❌ SAI: Dùng pool connection cho LISTEN
conn, _ := pool.Acquire(ctx)
conn.Conn().Exec(ctx, "LISTEN chat_events")  // Connection bị giữ vĩnh viễn!

// ✅ ĐÚNG: Tạo connection riêng
listenConn, _ := pgx.Connect(ctx, databaseURL)
listenConn.Exec(ctx, "LISTEN chat_events")
for {
    notification, _ := listenConn.WaitForNotification(ctx)
    hub.dispatch(notification.Payload)
}
```

### Multi-tab Support

- Server PHẢI support nhiều WS connections từ cùng `user_id`
- `hub.clients` = `map[userID][]ClientConn` (slice, không phải single)
- **KHÔNG kick connection cũ** khi có connection mới

### Client Reconnect — Exponential Backoff

```
Attempt 1: 1s
Attempt 2: 2s
Attempt 3: 4s
Attempt 4: 8s
...
Max: 30s
```

- Sau reconnect: gửi lại `join_conversation` cho tất cả conversations đang active
- Invalidate TanStack Query cache → refetch conversations + messages mới nhất

---

## 7. Soft Delete — Messages

```sql
-- ⚠️ KHÔNG ĐƯỢC dùng DELETE FROM messages
-- ✅ Chỉ dùng UPDATE
UPDATE messages
SET deleted_at = NOW(), body = NULL
WHERE id = $1 AND sender_id = $current_user_id;
```

### Tại sao set `body = NULL`?
- Tránh lưu nội dung nhạy cảm sau khi user "xóa"
- UI hiển thị: "Tin nhắn đã bị xóa" (giống Telegram/WhatsApp)
- Attachment metadata vẫn giữ trong bảng `attachments` → có thể xóa file storage riêng

### Query pattern — Luôn filter deleted:

```sql
SELECT * FROM messages
WHERE conversation_id = $1
  AND deleted_at IS NULL  -- ← BẮT BUỘC
  AND (created_at, id) < ($cursor_time, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT 30;
```

---

## 8. Cursor Pagination — Gotchas

### Tại sao dùng `(created_at, id)` thay vì chỉ `created_at`?

- 2 messages gửi trong cùng 1ms → `created_at` trùng nhau
- Chỉ dùng `created_at` → page boundary không deterministic → có thể skip hoặc duplicate messages
- Composite cursor `(created_at, id)` → 100% deterministic

### Frontend Gotchas:

```typescript
// Server trả data DESC order (mới → cũ)
// Frontend cần render ASC order (cũ → mới, cuộn xuống = mới nhất)

const newMessages = response.data.reverse();  // ← ĐẢO CHIỀU

// Scroll position preservation khi prepend (load more):
const scrollHeight = scrollRef.current.scrollHeight;
setMessages(prev => [...newMessages, ...prev]);
// Sau render:
scrollRef.current.scrollTop = scrollRef.current.scrollHeight - scrollHeight;
```

---

## 9. GDPR-Ready User Deletion

### Vấn đề: `messages.sender_id` dùng `ON DELETE RESTRICT`

```
Không thể DELETE FROM users WHERE id = $1
→ ERROR: update or delete on table "users" violates foreign key constraint
```

### Giải pháp: System Account Pattern

```sql
-- Seed 1 lần (migration hoặc seed script)
INSERT INTO users (id, email, username, display_name, status, password_hash)
VALUES (
    '00000000-0000-0000-0000-000000000000',  -- UUID cố định
    'system@deleted.local',
    'deleted_user',
    '[Deleted User]',
    'OFFLINE',
    '$2a$12$invalid_hash_that_never_matches'  -- Không thể login
);
```

```go
// Application layer: trước khi xóa user
func (uc *UserUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
    tx, _ := uc.pool.Begin(ctx)
    defer tx.Rollback(ctx)

    // 1. Reassign messages
    tx.Exec(ctx, "UPDATE messages SET sender_id = $1 WHERE sender_id = $2",
        systemDeletedUserID, userID)

    // 2. Xóa related data
    tx.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)
    tx.Exec(ctx, "DELETE FROM notifications WHERE recipient_id = $1 OR actor_id = $1", userID)
    tx.Exec(ctx, "DELETE FROM friendships WHERE requester_id = $1 OR addressee_id = $1", userID)
    tx.Exec(ctx, "DELETE FROM conversation_participants WHERE user_id = $1", userID)

    // 3. Xóa user
    tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

    return tx.Commit(ctx)
}
```

---

## 10. Password Security

### Timing Attack Prevention

```go
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*TokenPair, error) {
    user, err := uc.userRepo.FindByEmail(ctx, email)
    if err != nil {
        // ⚠️ User không tồn tại → VẪN chạy bcrypt compare với dummy hash
        // Tránh timing attack: attacker đo thời gian response để đoán email có tồn tại không
        bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
        return nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
        return nil, ErrInvalidCredentials  // CÙNG error message với case trên
    }
    // ...
}
```

### Error Messages — User Enumeration Prevention

```
⚠️ PHẢI dùng cùng 1 message cho cả "email không tồn tại" và "sai password":
   "Email hoặc mật khẩu không đúng"

❌ KHÔNG ĐƯỢC trả: "Email không tồn tại" hoặc "Sai mật khẩu"
```

### Bcrypt Config

```go
const bcryptCost = 12  // ~250ms trên server trung bình
// Tương lai: nâng lên 14 khi hardware nhanh hơn
```

---

## 11. Presence — Grace Period Pattern

### Vấn đề: User refresh page = WS disconnect + reconnect → blink offline/online

### Giải pháp: 60 giây grace period

```go
func (h *Hub) onClientDisconnect(client *Client) {
    // Xóa client khỏi hub ngay
    h.removeClient(client)

    // Kiểm tra: user còn connection nào khác không? (multi-tab)
    if h.hasActiveConnections(client.UserID) {
        return  // Còn tab khác → không làm gì
    }

    // Bắt đầu grace period
    timer := time.AfterFunc(60*time.Second, func() {
        // Sau 60s mà user không reconnect → broadcast OFFLINE
        h.db.Exec(ctx, "UPDATE users SET status='OFFLINE', last_seen_at=NOW() WHERE id=$1", client.UserID)
        h.broadcastPresence(client.UserID, "OFFLINE")
    })

    // Lưu timer để cancel nếu user reconnect
    h.gracePeriods[client.UserID] = timer
}

func (h *Hub) onClientConnect(client *Client) {
    // Cancel grace period nếu có
    if timer, ok := h.gracePeriods[client.UserID]; ok {
        timer.Stop()
        delete(h.gracePeriods, client.UserID)
    }
    // ...
}
```

---

## 12. Typing Indicator — Throttle & Auto-Clear

### Client Side (MessageInput.tsx)

```typescript
// Throttle: chỉ gửi typing_start 1 lần mỗi 3s
const lastTypingSent = useRef(0);

const handleInput = () => {
    const now = Date.now();
    if (now - lastTypingSent.current > 3000) {
        ws.send({ event: 'typing_start', payload: { conversation_id } });
        lastTypingSent.current = now;
    }
};

// Khi user dừng gõ hoặc gửi message → typing_stop
const handleSubmit = () => {
    ws.send({ event: 'typing_stop', payload: { conversation_id } });
    // ... send message
};
```

### Server Side — Auto-Clear Timer

```go
// Nếu không nhận typing_stop sau 5s → tự broadcast is_typing=false
func (h *Hub) onTypingStart(client *Client, convID string) {
    // Cancel timer cũ nếu có
    key := client.UserID + ":" + convID
    if timer, ok := h.typingTimers[key]; ok {
        timer.Stop()
    }

    // Broadcast typing=true
    h.broadcastTyping(convID, client.UserID, true)

    // Set auto-clear timer
    h.typingTimers[key] = time.AfterFunc(5*time.Second, func() {
        h.broadcastTyping(convID, client.UserID, false)
        delete(h.typingTimers, key)
    })
}
```

---

*File này là nguồn tham chiếu khi code. Mọi deviation cần ghi chú lại tại sao.*
