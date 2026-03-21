# 1. MVP Requirements — Real-time Chat Application

> **Version:** 1.1.0
> **Status:** Draft — Updated: Thêm Friend System & mở rộng Notifications
> **Stack:** Next.js (Vercel) · Golang (fly.io) · PostgreSQL/Neon · Cloudinary (ảnh) · R2 (video/file)

---

## 1. Core Features (In-Scope)

### 1.1 Authentication & Account

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| AUTH-01 | Đăng ký tài khoản | Email + Username + Password + Display Name. Validate: định dạng email, password tối thiểu 8 ký tự, username chỉ `a-z0-9_` (3-30 ký tự, unique) | P0 |
| AUTH-02 | Đăng nhập | Email + Password, trả về JWT Access Token + Refresh Token | P0 |
| AUTH-03 | Refresh Token | Tự động làm mới access token (silent refresh) | P0 |
| AUTH-04 | Đăng xuất | Revoke refresh token phía server, xóa token phía client, đóng WS connection | P0 |
| AUTH-05 | Avatar cơ bản | Upload ảnh đại diện lên Cloudinary khi đăng ký hoặc sau đó | P1 |
| AUTH-06 | Cập nhật profile | Sửa display name, bio, avatar | P1 |
| AUTH-07 | Xem profile user | Xem display_name, username, bio, avatar, trạng thái online của user khác | P1 |

> **Không dùng OAuth (Google/GitHub) ở MVP này.**

---

### 1.2 Friend System

> **Design decision:** Hai user lạ **không thể chat trực tiếp** với nhau. Phải kết bạn thành công trước. Đây là gate chống spam thay thế cho open-search model ban đầu.

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| FRD-01 | Tìm kiếm user | Tìm theo email hoặc display name — trả về kết quả kèm trạng thái quan hệ (stranger / pending_sent / pending_received / friend) | P0 |
| FRD-02 | Gửi lời mời kết bạn | `POST /api/friends/request` — tạo bản ghi trạng thái `PENDING`, không thể gửi 2 lần | P0 |
| FRD-03 | Hủy lời mời đã gửi | Người gửi hủy request khi còn `PENDING` | P0 |
| FRD-04 | Chấp nhận lời mời | Người nhận Accept → trạng thái chuyển thành `ACCEPTED`, tự động tạo DM conversation | P0 |
| FRD-05 | Từ chối lời mời | Người nhận Decline → xóa record, không thông báo cho người gửi | P0 |
| FRD-06 | Danh sách bạn bè | Xem toàn bộ danh sách bạn, sắp xếp theo tên | P0 |
| FRD-07 | Danh sách lời mời nhận được | Tab "Lời mời kết bạn" — chờ Accept/Decline | P0 |
| FRD-08 | Hủy kết bạn (Unfriend) | Xóa quan hệ bạn bè, **không** xóa lịch sử chat | P1 |

> **Luồng quan hệ:** `NONE` → *(gửi request)* → `PENDING` → *(accept)* → `ACCEPTED`
> Decline hoặc Cancel đều trả về `NONE`, người gửi có thể gửi lại sau.

---

### 1.3 Chat 1-1 (Direct Message)

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| DM-01 | Mở chat với bạn bè | Chỉ có thể bắt đầu DM với user có trạng thái `ACCEPTED` — server reject nếu chưa là bạn | P0 |
| DM-02 | Tạo conversation | Tự động tạo khi Accept lời mời kết bạn; mở lại nếu đã tồn tại | P0 |
| DM-03 | Gửi text message | Realtime qua WebSocket, hỗ trợ Unicode/Emoji | P0 |
| DM-04 | Nhận message realtime | Nhận message ngay lập tức không cần reload | P0 |
| DM-05 | Load lịch sử chat | Phân trang (cursor-based), load 30 tin nhắn mỗi lần | P0 |
| DM-06 | Trạng thái tin nhắn | Sent → Delivered → Read (ticks) | P1 |
| DM-07 | Danh sách conversations | Hiển thị danh sách DM, sắp xếp theo tin nhắn mới nhất | P0 |

---

### 1.4 Chat Nhóm (Group Chat)

> **Constraint:** Chỉ có thể thêm người đã là bạn bè vào nhóm.

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| GRP-01 | Tạo nhóm | Đặt tên nhóm, chọn thành viên (tối thiểu 3 người) | P0 |
| GRP-02 | Gửi/nhận message nhóm | Realtime qua WebSocket, hiển thị tên người gửi | P0 |
| GRP-03 | Thêm thành viên | Admin nhóm thêm thành viên mới | P1 |
| GRP-04 | Xóa thành viên | Admin nhóm kick thành viên | P1 |
| GRP-05 | Rời nhóm | Thành viên tự rời, admin tự động chuyển nếu là người cuối | P1 |
| GRP-06 | Đổi tên nhóm | Chỉ admin thực hiện | P1 |
| GRP-07 | Upload ảnh nhóm | Upload avatar nhóm lên Cloudinary | P2 |
| GRP-08 | Danh sách thành viên | Xem danh sách member trong nhóm | P0 |
| GRP-09 | Load lịch sử chat nhóm | Cursor-based pagination, 30 tin nhắn mỗi lần | P0 |

---

### 1.5 Gửi File & Ảnh

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| FILE-01 | Gửi ảnh (image) | JPEG, PNG, GIF, WebP — upload Cloudinary, giới hạn **10 MB** | P0 |
| FILE-02 | Gửi file | PDF, DOCX, XLSX, ZIP — upload R2, giới hạn **50 MB** | P0 |
| FILE-03 | Gửi video | MP4, MOV — upload R2, giới hạn **100 MB** | P1 |
| FILE-04 | Preview ảnh | Hiển thị thumbnail inline trong chat, click để xem full-size | P0 |
| FILE-05 | Download file | Nút download cho file/video | P0 |
| FILE-06 | Progress upload | Hiển thị thanh tiến trình khi upload | P1 |
| FILE-07 | Presigned URL | Backend cấp presigned URL, client upload trực tiếp lên R2/Cloudinary | P0 |

> **Giới hạn loại file:** Chỉ chấp nhận whitelist MIME type, reject phía server.

---

### 1.6 Trạng Thái Online / Offline

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| PRES-01 | Chỉ báo Online | Chấm xanh hiển thị khi user đang kết nối WebSocket | P0 |
| PRES-02 | Trạng thái Offline | Hiển thị "Last seen HH:MM" sau khi ngắt kết nối | P0 |
| PRES-03 | Broadcast trạng thái | Server broadcast đến tất cả user trong cùng conversation | P0 |
| PRES-04 | Heartbeat / Ping-Pong | Client ping mỗi 30s, server mark offline sau 60s không nhận | P0 |

---

### 1.7 Thông Báo (Notifications)

> **In-app notification** là tính năng bắt buộc ở MVP — dùng bell icon trên navbar, kết hợp WebSocket để push realtime.

#### 1.7.1 In-App Notifications (Bell Icon)

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| NOTIF-01 | Bell icon + unread count | Badge đỏ hiển thị tổng số thông báo chưa đọc trên navbar | P0 |
| NOTIF-02 | Notification dropdown/panel | Click bell → danh sách thông báo, sắp xếp mới nhất trước | P0 |
| NOTIF-03 | Thông báo lời mời kết bạn | "**[Tên]** đã gửi lời mời kết bạn" — kèm nút Accept / Decline inline | P0 |
| NOTIF-04 | Thông báo được chấp nhận kết bạn | "**[Tên]** đã chấp nhận lời mời kết bạn của bạn" | P0 |
| NOTIF-05 | Thông báo được thêm vào nhóm | "**[Tên admin]** đã thêm bạn vào nhóm **[Tên nhóm]**" | P0 |
| NOTIF-06 | Mark as read (từng thông báo) | Click vào thông báo → mark read, navigate đến context | P0 |
| NOTIF-07 | Mark all as read | Nút "Đánh dấu tất cả đã đọc" | P1 |
| NOTIF-08 | Realtime push qua WebSocket | Server push WS event `notification_new` ngay khi có thông báo mới | P0 |
| NOTIF-09 | Persist notifications | Thông báo lưu DB, không mất khi reload trang | P0 |
| NOTIF-10 | Load lịch sử thông báo | Pagination, load 20 thông báo mỗi lần | P1 |

#### 1.7.2 Chat Notifications

| ID | Feature | Mô tả | Priority |
|----|---------|-------|----------|
| NOTIF-11 | Unread message badge | Số tin nhắn chưa đọc trên mỗi conversation trong sidebar | P0 |
| NOTIF-12 | Browser notification | Push Notification khi tab không active (Web Notification API) — chỉ cho tin nhắn mới | P1 |
| NOTIF-13 | Notification sound | Âm thanh khi có tin nhắn mới | P2 |

> **Notification data model:** Mỗi notification có `type` enum: `FRIEND_REQUEST` · `FRIEND_ACCEPTED` · `GROUP_ADDED`. Frontend render nội dung dựa trên `type` + `actor` + `target_id`.

---

## 2. Out-of-Scope (Không làm ở MVP)

> Các tính năng dưới đây được ghi nhận cho roadmap tương lai nhưng **bị đóng băng** ở phase này để tránh scope creep.

| Tính năng | Lý do defer |
|-----------|-------------|
| OAuth (Google, GitHub, Facebook) | Phức tạp hóa auth flow, không cần thiết ở MVP |
| Block user | Friend system đã là gate chống spam; Block là layer bổ sung, defer phase 2 |
| Privacy setting "Ai có thể tìm thấy tôi" | Phụ thuộc vào Block feature, defer cùng |
| Voice / Video Call (WebRTC) | Tăng độ phức tạp cơ sở hạ tầng đáng kể |
| Message Reactions (Emoji react) | Nice-to-have, không ảnh hưởng core value |
| Reply / Quote message | UX phức tạp hơn, có thể làm sau |
| Forward message | Phụ thuộc vào reply feature |
| Pin message | Không quan trọng ở MVP |
| Tìm kiếm nội dung tin nhắn (Full-text search) | Cần index riêng (Elasticsearch/pg_trgm), defer |
| Message edit / delete | Phức tạp về consistency và audit log |
| Disappearing messages | Không cần thiết ở MVP |
| End-to-end Encryption (E2EE) | Phức tạp về key management |
| Bot / Webhook integration | Thuộc về platform features |
| Screen sharing | Phụ thuộc WebRTC |
| Threads / Sub-threads | Complexity cao (như Slack threads) |
| Push Notification mobile (FCM/APNs) | Không có mobile app ở MVP |
| Admin dashboard | Dành cho internal operations, phase sau |
| Multi-workspace / Org management | Phase 2 |
| Import/Export chat history | Phase 2 |
| Message translate | Third-party API cost, phase sau |

---

## 3. User Flows

### Flow 1: Đăng ký & Onboarding

```
[Landing Page]
    │
    ▼
[Form Đăng Ký] ──── Nhập Email, Password, Display Name
    │
    ▼
[POST /api/auth/register]
    │
    ├─ Lỗi: Email đã tồn tại ──► Hiển thị lỗi inline
    │
    └─ Thành công
         │
         ▼
    [Tự động đăng nhập] ──► Nhận JWT + Refresh Token
         │
         ▼
    [Upload Avatar] (optional, có thể skip)
         │
         ▼
    [Home / Conversation List] (trống, gợi ý tìm kiếm user)
```

---

### Flow 2: Đăng nhập

```
[Login Page]
    │
    ▼
[Form] ──── Email + Password
    │
    ▼
[POST /api/auth/login]
    │
    ├─ Lỗi: Sai thông tin ──► "Email hoặc mật khẩu không đúng"
    │
    └─ Thành công
         │
         ▼
    [Lưu Access Token (memory) + Refresh Token (httpOnly cookie)]
         │
         ▼
    [Home / Conversation List]
```

---

### Flow 3: Kết Bạn (Friend Request)

```
[User A — Tìm kiếm User B]
    │
    ▼
[GET /api/users/search?q=...] ──► Kết quả kèm relationship_status = "NONE"
    │
    ▼
[User A nhấn "Kết bạn"]
    │
    ▼
[POST /api/friends/request { to_user_id: B }]
    │
    ├─ Lỗi: Đã gửi rồi / Đã là bạn ──► Toast lỗi
    │
    └─ Thành công
         │
         ├─ [DB] Tạo record friendship status = PENDING
         │
         ├─ [DB] Tạo notification record cho User B (type: FRIEND_REQUEST)
         │
         └─ [WS] Server push event `notification_new` đến User B (nếu online)
                   │
                   ▼
              [User B thấy badge đỏ trên bell icon]
              [Mở notification panel]
                   │
                   ├─ Nhấn "Accept"
                   │     │
                   │     ▼
                   │  [POST /api/friends/accept { request_id }]
                   │     │
                   │     ├─ [DB] friendship status = ACCEPTED
                   │     ├─ [DB] Tự động tạo DM conversation A ↔ B
                   │     ├─ [DB] Tạo notification cho User A (type: FRIEND_ACCEPTED)
                   │     └─ [WS] Push `notification_new` đến User A
                   │
                   └─ Nhấn "Decline"
                         │
                         ▼
                      [POST /api/friends/decline { request_id }]
                         │
                         └─ [DB] Xóa record, không thông báo cho User A
```

---

### Flow 4: Chat 1-1 (sau khi đã kết bạn)

```
[Home / Danh sách bạn bè hoặc Conversations]
    │
    ▼
[Chọn bạn / conversation DM]
    │
    ▼
[GET /api/conversations/:id] ──── Server verify: 2 user là bạn bè
    │
    ├─ Lỗi: Chưa là bạn ──► 403 Forbidden
    │
    └─ Thành công
         │
         ▼
    [Mở Conversation View]
    │
    ├── [WebSocket Connect] ──── JOIN room conversation_id
    │
    ├── [Load lịch sử] ──── GET /api/conversations/:id/messages?cursor=...
    │
    └── [Gửi tin nhắn]
              │
              ▼
         [WS Event: send_message]
              │
              ▼
         [Server broadcast đến tất cả participant]
              │
              ▼
         [Cập nhật UI realtime]
```

---

### Flow 5: Tạo & Chat Nhóm

```
[Home]
    │
    ▼
[Nút "Tạo nhóm mới"]
    │
    ▼
[Modal tạo nhóm]
    ├── Nhập tên nhóm
    └── Chọn thành viên từ danh sách bạn bè (multi-select, tối thiểu 2 người khác)
    │         ⚠️ Chỉ hiển thị bạn bè — không thể thêm người lạ
    ▼
[POST /api/groups]
    │
    ▼
[Server tạo group + thêm creator làm admin]
    │
    ├─ [DB] Tạo notification cho từng thành viên được thêm (type: GROUP_ADDED)
    └─ [WS] Push `notification_new` đến từng thành viên online
    │
    ▼
[Mở Group Chat View]
    │
    └── (tương tự Flow 4 từ bước WebSocket)
```

---

### Flow 6: Gửi File / Ảnh

```
[Trong Conversation View]
    │
    ▼
[Click icon đính kèm / paste ảnh]
    │
    ▼
[Client validate: loại file, kích thước]
    │
    ├─ Lỗi: Sai loại / quá lớn ──► Toast thông báo lỗi
    │
    └─ Hợp lệ
         │
         ▼
    [POST /api/upload/presign] ──── Xin presigned URL từ server
         │
         ▼
    [Client upload trực tiếp lên Cloudinary (ảnh) hoặc R2 (file/video)]
         │     (Hiển thị progress bar)
         ▼
    [Upload xong] ──── Nhận public URL / key
         │
         ▼
    [WS Event: send_message với attachment metadata]
         │
         ▼
    [Hiển thị ảnh thumbnail hoặc file card trong chat]
```

---

### Flow 7: Online / Offline Presence

```
[User mở ứng dụng / tab active]
    │
    ▼
[WebSocket Connect thành công]
    │
    ▼
[Server cập nhật user status = ONLINE]
[Server broadcast PRESENCE_UPDATE đến các conversation liên quan]
    │
    ▼
[Client khác nhận event ──► Hiển thị chấm xanh]

    │ (mỗi 30 giây)
    ▼
[Client gửi WS Ping]
[Server reset timeout counter]

[User đóng tab / mất mạng]
    │
    ▼
[WebSocket Disconnect]
    │
    ▼
[Server đợi 60s (grace period)]
    │
    ▼
[Server cập nhật last_seen = NOW(), status = OFFLINE]
[Server broadcast PRESENCE_UPDATE]
    │
    ▼
[Client khác nhận event ──► Hiển thị "Last seen HH:MM"]
```

---

### Flow 8: In-App Notification (Bell Icon)

```
[Server có event mới: FRIEND_REQUEST / FRIEND_ACCEPTED / GROUP_ADDED]
    │
    ▼
[DB] INSERT INTO notifications (user_id, type, actor_id, target_id, is_read=false)
    │
    ▼
[WS] Push event `notification_new` { id, type, actor, preview_text, created_at }
    │           đến user đích (nếu đang online)
    ▼
[Client nhận WS event]
    │
    ├─ Tăng badge count trên bell icon
    └─ Hiển thị toast popup (góc phải màn hình)

[User click vào Bell Icon]
    │
    ▼
[GET /api/notifications?limit=20] ──── Load danh sách
    │
    ▼
[Notification Panel hiển thị]
    │
    ├── Item FRIEND_REQUEST ──► Hiển thị nút Accept / Decline inline
    ├── Item FRIEND_ACCEPTED ──► Click → navigate đến DM conversation
    └── Item GROUP_ADDED ──► Click → navigate đến Group chat

[User click vào 1 notification]
    │
    ▼
[PATCH /api/notifications/:id/read] ──── Mark read
    │
    ▼
[Badge count giảm, item highlight tắt, navigate đến context]
```

---

## 4. Non-Functional Requirements (MVP)

| Hạng mục | Target |
|----------|--------|
| Latency message delivery | < 200ms (p95, cùng region) |
| Concurrent WebSocket connections | 1,000 connections / instance |
| File upload timeout | 5 phút |
| API response time | < 500ms (p95) cho các endpoint thông thường |
| Availability | 99.5% uptime (single fly.io instance) |
| Auth token TTL | Access: 15 phút · Refresh: 7 ngày |
| Message history retention | Không giới hạn (MVP, xem xét lại khi scale) |

---

## 5. Tech Decisions Liên Quan đến MVP

| Quyết định | Lựa chọn | Lý do |
|-----------|----------|-------|
| Realtime transport | WebSocket (native) | Đơn giản, phù hợp chat. Không dùng SSE hay polling |
| Message queue | Không dùng ở MVP | Một instance Golang đủ cho MVP, thêm Redis Pub/Sub khi scale |
| Session management | JWT stateless + Refresh Token rotation | Phù hợp kiến trúc stateless Golang trên fly.io |
| Image storage | Cloudinary | CDN tích hợp, transformation tự động (resize thumbnail) |
| File/Video storage | Cloudflare R2 | Bandwidth miễn phí, presigned URL, S3-compatible |
| Notification delivery | WebSocket (realtime) + DB persist | Realtime qua WS khi online; DB làm source of truth khi offline/reload |
| DB migration | golang-migrate | Versioned migrations, CI/CD friendly |
| ORM | sqlc | Type-safe queries từ raw SQL, phù hợp Go |

---

*Document này được duyệt trước khi bắt đầu sprint. Mọi thay đổi scope cần có sign-off từ PM.*
