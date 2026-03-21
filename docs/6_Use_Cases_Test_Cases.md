# 6. Use Cases & Test Cases

> **Version:** 1.0.0
> **Role:** Senior QA/QC Automation Engineer
> **Dựa trên:** `1_MVP_Requirements.md` v1.1.0 · `3_API_WebSocket_Specs.md` · `5_UI_Wireframes.md`
> **Stack:** Golang (Fly.io) · Next.js · PostgreSQL LISTEN/NOTIFY (no Redis)

---

## 1. Test Strategy — Quy Chuẩn Chung

### 1.1 Phân Loại Test

| Loại | Ký hiệu | Định nghĩa | Mục tiêu |
|------|---------|-----------|---------|
| **Happy Path** | `HP` | Luồng đúng, đầu vào hợp lệ, hệ thống hoạt động bình thường | Xác nhận core feature hoạt động đúng spec |
| **Unhappy Path** | `UP` | Đầu vào sai, hành động không được phép, lỗi có thể dự đoán | Xác nhận error handling trả đúng mã lỗi, không lộ thông tin nhạy cảm |
| **Edge Case** | `EC` | Biên giới giá trị, race condition, network failure, đầu vào cực trị | Xác nhận hệ thống không crash, degradation graceful |

### 1.2 Cấu Trúc Chuẩn Một Test Case

Mỗi test case được trình bày theo bảng 6 cột:

| Cột | Mô tả |
|-----|-------|
| **ID** | `[UseCase]-[Loại]-[Số thứ tự]` — VD: `AUTH-HP-01` |
| **Tên Test Case** | Mô tả ngắn gọn hành vi đang kiểm thử |
| **Pre-conditions** | Trạng thái hệ thống và dữ liệu phải có trước khi chạy test |
| **Steps** | Các bước thực hiện theo thứ tự, đánh số |
| **Expected Result — UI** | Giao diện người dùng phải hiển thị gì |
| **Expected Result — DB / WS / API** | Trạng thái DB, WebSocket event, hoặc HTTP response phải như thế nào |

### 1.3 Môi Trường Test

| Layer | Tool đề xuất |
|-------|-------------|
| API (REST) | Postman Collection / `httptest` (Go) |
| WebSocket | `wscat` CLI / Playwright WebSocket intercept |
| DB assertion | `pgx` trong Go integration test / `psql` query trực tiếp |
| Frontend E2E | Playwright (Next.js) |
| Frontend Unit | Vitest + React Testing Library |
| Load / Concurrency | `k6` |

### 1.4 Môi Trường & Dữ Liệu Seed

```
Test DB: PostgreSQL Neon (branch riêng: `test-branch`)
Seed data:
  - user_alice:   email=alice@test.com, username=alice_dev,   password=Test@1234
  - user_bob:     email=bob@test.com,   username=bob_smith,   password=Test@1234
  - user_charlie: email=charlie@test.com, username=charlie_99, password=Test@1234
  - friendship(alice ↔ bob): status=ACCEPTED, conversation_id đã tồn tại
  - friendship(alice ↔ charlie): status=NONE (chưa kết bạn)
```

---

## 2. Use Case 1 — Authentication

### 2.1 Đăng Nhập Thành Công

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-HP-01` |
| **Tên** | Đăng nhập thành công với email và mật khẩu hợp lệ |
| **Pre-conditions** | `user_alice` tồn tại trong DB. Server đang chạy. Không có cookie `refresh_token` hiện tại. |
| **Steps** | 1. Mở `/login`. 2. Nhập `email = alice@test.com`. 3. Nhập `password = Test@1234`. 4. Click "Đăng nhập". |
| **Expected Result — UI** | Spinner hiển thị trên button. Sau ≤ 500ms: redirect sang `/conversations`. Sidebar hiển thị avatar + tên "Alice". Bell icon hiển thị badge (nếu có notif chưa đọc). Không có error box đỏ. |
| **Expected Result — DB / WS / API** | `POST /v1/auth/login` → HTTP 200. Response body chứa `access_token` (JWT, expires_in=900). Response header `Set-Cookie` chứa `refresh_token=<hash>; HttpOnly; Secure; SameSite=Strict`. DB: `users.status = 'ONLINE'`, `users.last_seen_at = NOW()±5s`. DB: `refresh_tokens` INSERT 1 row mới với `is_revoked=false`, `expires_at = NOW()+7days`. |

---

### 2.2 Đăng Nhập Sai Mật Khẩu (User Enumeration Prevention)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-UP-01` |
| **Tên** | Đăng nhập sai mật khẩu — không tiết lộ email có tồn tại hay không |
| **Pre-conditions** | `user_alice` tồn tại trong DB. |
| **Steps** | **Case A:** Nhập `email=alice@test.com`, `password=SaiMatKhau123`. Click "Đăng nhập". **Case B:** Nhập `email=khongtontai@test.com`, `password=BatKyMatKhau`. Click "Đăng nhập". |
| **Expected Result — UI** | **Cả 2 case:** Global error box màu đỏ xuất hiện phía trên form với **đúng cùng 1 message**: "Email hoặc mật khẩu không đúng". Không có message phân biệt "Email không tồn tại" vs "Sai mật khẩu". Input fields không bị clear. Button trở lại trạng thái bình thường. |
| **Expected Result — DB / WS / API** | `POST /v1/auth/login` → HTTP 401 cho **cả 2 case**. Response body: `{ "error": { "code": "INVALID_CREDENTIALS", "message": "Email hoặc mật khẩu không đúng" } }`. Response time của Case A và Case B phải **xấp xỉ nhau** (bcrypt dummy hash được chạy kể cả khi email không tồn tại — tránh timing attack). Không có cookie được set. DB: không có row nào thay đổi. |

---

### 2.3 Đăng Nhập Sai Mật Khẩu Nhiều Lần (Rate Limit)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-UP-02` |
| **Tên** | Brute force login bị chặn bởi rate limiter |
| **Pre-conditions** | Server đang chạy. IP hiện tại chưa bị rate-limit. |
| **Steps** | Gửi `POST /v1/auth/login` với sai mật khẩu liên tục **11 lần trong vòng 60 giây** từ cùng 1 IP. |
| **Expected Result — UI** | Lần thứ 11: global error box hiển thị "Quá nhiều yêu cầu, vui lòng thử lại sau". Button bị disable tạm thời. |
| **Expected Result — DB / WS / API** | Request thứ 11: HTTP 429. Response header `Retry-After: <seconds>`. DB: không có thay đổi. Middleware `RateLimiter(10/min per IP)` trigger. |

---

### 2.4 Silent Refresh Token qua Axios Interceptor (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-EC-01` |
| **Tên** | Access token hết hạn — Axios interceptor tự động refresh và retry request gốc |
| **Pre-conditions** | `user_alice` đã đăng nhập. `access_token` hiện tại đã **hết hạn** (expired JWT, có thể mock bằng cách set `JWT_ACCESS_TTL_SECONDS=1` và đợi). Cookie `refresh_token` hợp lệ và chưa hết hạn. |
| **Steps** | 1. Alice đang ở trang `/conversations`. 2. Access token hết hạn (sau 15 phút hoặc mock). 3. Alice click vào một conversation → app gọi `GET /v1/conversations/:id/messages`. |
| **Expected Result — UI** | **Không có gì thay đổi trên UI**. Danh sách tin nhắn load bình thường sau ≤ 1s. Không có redirect về `/login`. Không có error toast. Không có màn hình "blink" hay loading toàn trang. |
| **Expected Result — DB / WS / API** | **Luồng ngầm (không hiển thị trên UI):** 1. Request `GET /conversations/:id/messages` → HTTP 401 (token expired). 2. Axios interceptor bắt lỗi 401 → gọi `POST /v1/auth/refresh` (dùng cookie). 3. `/refresh` → HTTP 200 + `new_access_token` + cookie mới (Rotation). 4. Interceptor retry request gốc với `new_access_token` → HTTP 200 + data. DB: `refresh_tokens`: row cũ `is_revoked=true`, row mới INSERT với `is_revoked=false`. **Nếu `/refresh` cũng thất bại:** redirect về `/login`. |

---

### 2.5 Sử Dụng Lại Refresh Token Đã Bị Revoke (Token Theft Detection)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-EC-02` |
| **Tên** | Phát hiện token reuse — revoke toàn bộ session |
| **Pre-conditions** | `user_alice` đã đăng nhập. Lưu lại `refresh_token_v1` (cookie cũ). Alice đã refresh thành công → `refresh_token_v1` bị revoke, `refresh_token_v2` active. |
| **Steps** | 1. Giả lập attacker dùng lại `refresh_token_v1` (đã revoke): gửi `POST /v1/auth/refresh` với cookie chứa `refresh_token_v1`. |
| **Expected Result — UI** | Tất cả session của Alice bị đăng xuất. Nếu Alice đang mở tab → redirect về `/login` với message "Phiên đăng nhập đã hết hạn vì lý do bảo mật". |
| **Expected Result — DB / WS / API** | `POST /v1/auth/refresh` → HTTP 401. DB: **TẤT CẢ** `refresh_tokens` của `user_alice` bị UPDATE `is_revoked=true` (không chỉ token hiện tại). WebSocket connection của Alice bị server chủ động đóng. |

---

## 3. Use Case 2 — Friend Request & Notifications

### 3.1 Gửi Lời Mời Kết Bạn → Nhận Notification Realtime (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-HP-01` |
| **Tên** | A gửi lời mời kết bạn → B nhận notification realtime qua WebSocket |
| **Pre-conditions** | `user_alice` đã login, đang ở trang `/friends`. `user_charlie` đã login trên tab/browser khác, WebSocket đã kết nối. Quan hệ alice ↔ charlie = `NONE`. |
| **Steps** | 1. Alice mở modal "Tìm kiếm người dùng". 2. Gõ `@charlie_99`. 3. Kết quả hiển thị Charlie với button "+ Kết bạn". 4. Alice click "+ Kết bạn". |
| **Expected Result — UI (Alice)** | Button chuyển thành "Đã gửi lời mời ↩" (bg-slate-600, disabled). Toast success ngắn: "Đã gửi lời mời kết bạn". |
| **Expected Result — UI (Charlie)** | Trong ≤ 200ms: Bell icon badge tăng lên 1 (bg-indigo-500). Nếu Charlie mở Notification Dropdown: xuất hiện item mới "Alice đã gửi lời mời kết bạn" với 2 button "✓ Chấp nhận" và "✗ Từ chối". |
| **Expected Result — DB / WS / API** | `POST /v1/friends/requests` → HTTP 201. DB `friendships`: INSERT 1 row `(requester=alice, addressee=charlie, status=PENDING)`. DB `notifications`: INSERT 1 row `(recipient=charlie, actor=alice, type=FRIEND_REQUEST, is_read=false)`. WS Hub: `pg_notify('chat_events', {type:"notification_new", recipient_ids:["charlie_id"], ...})`. Charlie's WS client nhận event `notification_new` với `unread_count` tăng. |

---

### 3.2 Charlie Accept Lời Mời → Tự Động Tạo DM (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-HP-02` |
| **Tên** | B accept lời mời → tự động tạo DM conversation, Alice nhận notification |
| **Pre-conditions** | `FRD-HP-01` đã chạy thành công. Charlie đang xem Notification Dropdown với item lời mời của Alice. |
| **Steps** | 1. Charlie click "✓ Chấp nhận" trên notification item. |
| **Expected Result — UI (Charlie)** | Notification item biến mất (hoặc chuyển sang trạng thái "Đã chấp nhận"). Bell badge giảm 1. Sidebar của Charlie xuất hiện conversation DM mới với Alice ở đầu danh sách (`last_activity_at` mới nhất). Toast: "Bạn và Alice đã trở thành bạn bè". |
| **Expected Result — UI (Alice)** | Bell badge tăng 1. Notification mới: "Charlie đã chấp nhận lời mời kết bạn của bạn". Sidebar của Alice xuất hiện conversation DM với Charlie. |
| **Expected Result — DB / WS / API** | `PATCH /v1/friends/requests/:id` body `{action:"ACCEPT"}` → HTTP 200 (trong **1 DB transaction**): DB `friendships`: UPDATE `status=ACCEPTED`, `updated_at=NOW()`. DB `conversations`: INSERT 1 row `(type='DM')`. DB `conversation_participants`: INSERT 2 rows (alice + charlie, role='MEMBER'). DB `notifications`: INSERT 1 row `(recipient=alice, actor=charlie, type=FRIEND_ACCEPTED)`. WS: `notification_new` push đến Alice. `GET /v1/conversations` của cả 2 user phải trả về conversation mới. |

---

### 3.3 Gửi Lời Mời Cho Người Đã Là Bạn Bè (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-UP-01` |
| **Tên** | Gửi lời mời kết bạn tới user đã có quan hệ ACCEPTED — kiểm tra DB constraint |
| **Pre-conditions** | alice ↔ bob đã có `friendships.status = ACCEPTED`. Alice đã login. |
| **Steps** | 1. Alice gọi `POST /v1/friends/requests` với body `{ "to_user_id": "<bob_id>" }` (bypass UI hoặc qua Postman). |
| **Expected Result — UI** | Nếu qua UI: button "Nhắn tin →" hiển thị (không có button "+ Kết bạn"). Nếu API bị gọi trực tiếp: toast error ngắn "Bạn đã là bạn bè với người này". |
| **Expected Result — DB / WS / API** | `POST /v1/friends/requests` → HTTP 409. Response: `{ "error": { "code": "CONFLICT", "message": "Lời mời đã được gửi hoặc đã là bạn bè" } }`. DB `friendships`: **không có row nào thay đổi**. DB `notifications`: **không có row nào được INSERT**. `idx_friendships_canonical` unique index ngăn duplicate. |

---

### 3.4 Gửi Lời Mời Khi Đang Có Pending Request (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-UP-02` |
| **Tên** | Gửi lời mời kết bạn khi đã có PENDING request tồn tại — không tạo duplicate |
| **Pre-conditions** | alice đã gửi lời mời cho charlie (`PENDING`). Charlie chưa Accept/Decline. |
| **Steps** | 1. Alice gọi lại `POST /v1/friends/requests` với `to_user_id = charlie_id` lần thứ 2. |
| **Expected Result — UI** | Button đã ở trạng thái "Đã gửi lời mời ↩" (disabled). Nếu API gọi trực tiếp: toast error. |
| **Expected Result — DB / WS / API** | HTTP 409. `{ "error": { "code": "CONFLICT" } }`. DB: chỉ có 1 row duy nhất trong `friendships` cho cặp alice-charlie. `idx_friendships_canonical` unique index enforce. |

---

### 3.5 Tự Gửi Lời Mời Cho Chính Mình (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-UP-03` |
| **Tên** | User cố gửi lời mời kết bạn cho chính mình |
| **Pre-conditions** | Alice đã login. |
| **Steps** | 1. Gọi `POST /v1/friends/requests` với body `{ "to_user_id": "<alice_own_id>" }`. |
| **Expected Result — UI** | Trong search results: user hiện tại không xuất hiện (server filter `WHERE id <> $current_user_id`). Nếu API gọi trực tiếp: error message rõ ràng. |
| **Expected Result — DB / WS / API** | HTTP 400. `{ "error": { "code": "VALIDATION_ERROR", "message": "Không thể gửi lời mời cho chính mình" } }`. DB `CHECK (requester_id <> addressee_id)` constraint ngăn INSERT. Không có notification được tạo. |

---

### 3.6 Chat với User Chưa Kết Bạn — Bị Từ Chối (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `FRD-UP-04` |
| **Tên** | Cố gắng gửi tin nhắn cho user chưa phải bạn bè — server từ chối |
| **Pre-conditions** | alice ↔ charlie: `NONE`. Alice biết `conversation_id` của Charlie (hoặc tạo request thẳng). |
| **Steps** | 1. Gọi `POST /v1/conversations/:charlie_conv_id/messages` với body hợp lệ. |
| **Expected Result — UI** | Toast error: "Bạn không thể gửi tin nhắn cho người này". UI không render optimistic message. |
| **Expected Result — DB / WS / API** | HTTP 403. `{ "error": { "code": "FORBIDDEN" } }`. Server kiểm tra `friendships.status = ACCEPTED` trước khi INSERT message. DB `messages`: không có row nào được INSERT. Không có `pg_notify` được gọi. |

---

## 4. Use Case 3 — Real-time Messaging (Chat 1-1)

### 4.1 Gửi Tin Nhắn Text → B Nhận Realtime (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-HP-01` |
| **Tên** | A gửi tin nhắn text → Optimistic UI phía A, B nhận qua WebSocket `new_message` |
| **Pre-conditions** | alice ↔ bob: `ACCEPTED`. Cả 2 đã login. Cả 2 đang mở conversation DM (đã `join_conversation`). WebSocket của cả 2 đang kết nối. |
| **Steps** | 1. Alice nhập "Mọi người ơi, họp lúc 3h nha!" vào MessageInput. 2. Alice nhấn Enter (hoặc click Send). |
| **Expected Result — UI (Alice)** | **Ngay lập tức (< 16ms):** Message xuất hiện ở cuối MessageList, align phải, bg-indigo-600, trạng thái spinner "◌ Đang gửi..." (opacity-60). Input field được clear. **Sau khi nhận `message_sent` ack:** Spinner biến mất → hiển thị tick đơn "✓" (text-slate-400, SENT). Sau khi Bob's WS ack → tick đôi "✓✓" (DELIVERED). |
| **Expected Result — UI (Bob)** | Trong ≤ 200ms: Message xuất hiện ở cuối MessageList. Nếu Bob không đang focus tab này: notification badge tăng trên conversation item trong sidebar. |
| **Expected Result — DB / WS / API** | WS Client→Server: `{event:"send_message", payload:{conversation_id, body, type:"TEXT", client_temp_id:"temp-uuid"}}`. Server: INSERT `messages` → `pg_notify('chat_events', {type:"new_message", ...})`. Hub `ListenLoop` nhận notify → `dispatch()`. WS Server→Alice: `{event:"message_sent", payload:{client_temp_id:"temp-uuid", message_id:"real-uuid", status:"SENT"}}`. WS Server→Bob: `{event:"new_message", payload:{message:{...real data...}}}`. DB: `messages` có 1 row mới. `conversations.last_message_id` và `last_activity_at` được UPDATE trong cùng transaction. |

---

### 4.2 Kiểm Tra Read Receipt — Tick Xanh (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-HP-02` |
| **Tên** | Bob đọc tin nhắn → Alice thấy tick đôi chuyển màu xanh (READ) |
| **Pre-conditions** | `MSG-HP-01` đã chạy. Message đang ở trạng thái DELIVERED. Bob đang mở conversation nhưng tab không active. |
| **Steps** | 1. Bob click vào conversation tab → tab trở thành active. 2. MessageList scroll xuống cuối (tin nhắn mới nhất trong viewport). |
| **Expected Result — UI (Bob)** | Không có thay đổi visual đặc biệt. Unread badge trên conversation item trong sidebar về 0. |
| **Expected Result — UI (Alice)** | Tick đôi "✓✓" chuyển từ `text-slate-400` sang `text-indigo-400` (READ state) trong ≤ 200ms. |
| **Expected Result — DB / WS / API** | Client Bob gửi WS `{event:"mark_read", payload:{conversation_id, last_message_id}}`. Server: UPDATE `conversation_participants.last_read_message_id` và `last_read_at`. UPDATE `messages.status = 'READ'` cho tất cả messages chưa đọc trong DM. `pg_notify` → WS Server→Alice: `{event:"read_receipt", payload:{reader_id:"bob_id", last_read_message_id, read_at}}`. |

---

### 4.3 Upload File PDF Vượt Giới Hạn 51MB (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-EC-01` |
| **Tên** | Upload file PDF 51MB — bị chặn trước khi gọi presign API |
| **Pre-conditions** | Alice đang ở conversation view với Bob. File `large_file.pdf` có dung lượng 51MB. |
| **Steps** | 1. Alice click icon đính kèm 📎. 2. Chọn file `large_file.pdf` (51MB). |
| **Expected Result — UI** | **Client-side validation chạy trước (không gọi API).** Toast error xuất hiện: "File tối đa 50MB. File của bạn: 51.0 MB". File không được chọn (input reset). Không có progress bar. Không có optimistic message. |
| **Expected Result — DB / WS / API** | `POST /v1/upload/file/presign` **không được gọi** (validation ở `useUpload` hook phía client). Không có row nào trong DB `messages` hoặc `attachments`. Không có WS event. Nếu bypass client validation và gọi API trực tiếp: HTTP 413, `{ "error": { "code": "FILE_TOO_LARGE", "message": "File tối đa 50MB" } }`. |

---

### 4.4 Upload File PDF Đúng 50MB (Boundary Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-EC-02` |
| **Tên** | Upload file đúng bằng giới hạn 50MB — phải thành công |
| **Pre-conditions** | Alice đang ở conversation view. File `exact_limit.pdf` = 52,428,800 bytes (50 × 1024 × 1024). |
| **Steps** | 1. Alice chọn file `exact_limit.pdf`. 2. Xác nhận upload. |
| **Expected Result — UI** | Progress bar xuất hiện. Upload hoàn tất. File card hiển thị trong chat. |
| **Expected Result — DB / WS / API** | `POST /v1/upload/file/presign` → HTTP 200 + presigned PUT URL. R2 PUT upload thành công. `POST /v1/conversations/:id/messages` → HTTP 201. DB `attachments.size_bytes = 52428800`. WS `new_message` broadcast tới Bob. |

---

### 4.5 Cursor-Based Pagination — Scroll Lên Xem Lịch Sử (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-EC-03` |
| **Tên** | User scroll lên load thêm lịch sử — kiểm tra composite cursor `(created_at, id)` |
| **Pre-conditions** | Conversation alice ↔ bob có 75 messages trong DB. Alice mở conversation → 30 messages mới nhất đã load. |
| **Steps** | 1. Alice scroll lên đỉnh của MessageList. 2. `useMessages` hook detect scroll position ≤ threshold → trigger load more. 3. Lặp lại đến khi `has_more = false`. |
| **Expected Result — UI** | Loading skeleton xuất hiện phía trên message list trong khi fetch. 30 messages cũ hơn được prepend vào đầu list (không gây scroll jump — scroll position được giữ nguyên). Lần 3 (sau 90 messages): không còn skeleton, không còn "load more" trigger. |
| **Expected Result — DB / WS / API** | **Lần 1:** `GET /conversations/:id/messages?limit=30` → 30 messages mới nhất, `meta.has_more=true`, `meta.next_cursor={before_id:"msg-30th", before_time:"..."}`. **Lần 2:** `GET /conversations/:id/messages?limit=30&before_id=msg-30th&before_time=...` → 30 messages tiếp theo, `meta.has_more=true`. **Lần 3:** `GET /conversations/:id/messages?limit=30&before_id=msg-60th&before_time=...` → 15 messages còn lại, `meta.has_more=false`. DB query dùng `WHERE (created_at, id) < ($cursor_time, $cursor_id)` — không có page bị duplicate hay bị skip kể cả khi 2 messages có cùng `created_at`. |

---

### 4.6 Gửi Tin Nhắn Khi Không Còn Là Bạn Bè (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-UP-01` |
| **Tên** | Alice unfriend Bob nhưng vẫn cố gửi tin nhắn qua conversation cũ |
| **Pre-conditions** | alice ↔ bob trước đó là `ACCEPTED`. Alice vừa unfriend Bob (DELETE friendship). Conversation DM vẫn tồn tại trong DB. Alice vẫn đang ở màn hình conversation. |
| **Steps** | 1. Alice nhập text vào MessageInput. 2. Nhấn Send. |
| **Expected Result — UI** | Optimistic message **không được** render (hoặc render rồi bị remove ngay). Toast error: "Bạn không thể gửi tin nhắn cho người này". MessageInput text bị restore (user không mất nội dung đã gõ). |
| **Expected Result — DB / WS / API** | WS `send_message` event → server validate `friendships.status = ACCEPTED` → thất bại → WS Server→Alice: `{event:"error", payload:{code:"FORBIDDEN", ref_event:"send_message"}}`. DB `messages`: không có row mới. Không có `pg_notify`. |

---

### 4.7 Gửi Tin Nhắn Attachment Ảnh — Preview Thumbnail (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `MSG-HP-03` |
| **Tên** | Alice gửi ảnh PNG 2MB — Bob thấy thumbnail inline, click để xem full-size |
| **Pre-conditions** | alice ↔ bob: `ACCEPTED`. Cả 2 đang mở conversation. File `photo.png` = 2MB. |
| **Steps** | 1. Alice chọn `photo.png` qua file picker. 2. Client gọi presign Cloudinary. 3. Upload trực tiếp lên Cloudinary. 4. Client gọi `POST /conversations/:id/messages` với attachment metadata. 5. Bob nhìn vào màn hình. |
| **Expected Result — UI (Alice)** | Progress bar hiển thị % upload lên Cloudinary. Sau khi xong: thumbnail ảnh hiển thị trong bubble (280×200px, object-cover, rounded-xl). |
| **Expected Result — UI (Bob)** | WS `new_message` → thumbnail ảnh render trong MessageList. Click thumbnail → Lightbox modal mở full-size image. |
| **Expected Result — DB / WS / API** | `POST /v1/upload/avatar/presign` (Cloudinary) → 200 + signature. Client upload → Cloudinary trả về `public_id` + `secure_url`. `POST /v1/conversations/:id/messages` body `{type:"ATTACHMENT", attachment:{storage_type:"CLOUDINARY", file_type:"IMAGE", url:"https://res.cloudinary.com/...", thumbnail_url:"...", ...}}` → 201. DB `attachments`: 1 row với `width`, `height` được lưu. `thumbnail_url` dùng Cloudinary transformation `c_fill,w_300`. |

---

## 5. Use Case 4 — Trạng Thái & WebSocket Resilience

### 5.1 User Đóng Tab → Grace Period → Broadcast Offline (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-HP-01` |
| **Tên** | Bob đóng tab → sau 60s grace period → hệ thống broadcast OFFLINE + "Last seen" |
| **Pre-conditions** | Bob đã login, WebSocket đang kết nối. Alice đang xem conversation với Bob. Bob hiển thị chấm xanh online trên header conversation của Alice. |
| **Steps** | 1. Bob đóng tab trình duyệt (hoặc đóng browser). 2. Đợi **60 giây**. 3. Quan sát UI của Alice. |
| **Expected Result — UI (Alice)** | **Trong vòng 60s:** Chấm xanh vẫn hiển thị (grace period). **Sau đúng 60s ± 5s:** Chấm xanh biến mất. Header conversation hiển thị "Last seen HH:MM" (thời điểm Bob disconnect). Trong sidebar, conversation item của Bob không còn chấm online. |
| **Expected Result — DB / WS / API** | Server detect `WebSocket disconnect` → khởi động 60s timer. Sau 60s không có ping từ Bob: `UPDATE users SET status='OFFLINE', last_seen_at=NOW()`. `pg_notify('chat_events', {type:"presence_update", user_id:"bob_id", status:"OFFLINE", last_seen_at:"..."})`. Hub dispatch → WS Server→Alice: `{event:"presence_update", payload:{user_id:"bob", status:"OFFLINE", last_seen_at:"..."}}`. |

---

### 5.2 Grace Period Bị Hủy Khi User Reconnect (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-EC-01` |
| **Tên** | Bob đóng tab rồi mở lại trong vòng 60s — grace period bị hủy, không broadcast offline |
| **Pre-conditions** | Bob đã login. WebSocket đang kết nối. |
| **Steps** | 1. Bob đóng tab. 2. **Sau 30 giây** (còn trong grace period), Bob mở lại tab và kết nối WS lại. |
| **Expected Result — UI (Alice)** | Không có thay đổi gì — Bob vẫn hiển thị online trong suốt quá trình. Không có "Last seen" nào xuất hiện. |
| **Expected Result — DB / WS / API** | Server: timer 60s bị **cancel** khi nhận WS reconnect từ Bob. DB `users.status` **không bị UPDATE** thành `OFFLINE`. Không có `pg_notify` presence_update. Sau reconnect: Bob gửi ping → server reset heartbeat timer. |

---

### 5.3 Heartbeat Ping/Pong — Phát Hiện Zombie Connection (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-EC-02` |
| **Tên** | Client không gửi ping trong 60s (zombie connection) → server đóng connection và mark offline |
| **Pre-conditions** | Bob đang kết nối WS. Giả lập bằng cách: mock `useWebSocket` hook để không gửi ping (hoặc dùng network throttle để block outgoing WS frames). |
| **Steps** | 1. Block ping của Bob. 2. Đợi 60 giây. |
| **Expected Result — UI (Alice)** | Bob chuyển sang trạng thái "Last seen" sau ≤ 65s. |
| **Expected Result — DB / WS / API** | Server: sau 60s không nhận ping → chủ động đóng WS connection của Bob với close code `1001` (Going Away). Server logic tương tự `WS-HP-01`: UPDATE status=OFFLINE, pg_notify. |

---

### 5.4 Mất Mạng Đột Ngột → Reconnect Với Exponential Backoff (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-EC-03` |
| **Tên** | Client mất mạng và có mạng lại — `useWebSocket` tự reconnect với exponential backoff |
| **Pre-conditions** | Alice đang chat với Bob, WS đang kết nối. DevTools Network throttling sẵn sàng. |
| **Steps** | 1. Simulate mất mạng (DevTools → Offline). 2. Đợi 10 giây (observe retry attempts). 3. Restore network (DevTools → Online). 4. Quan sát reconnect. |
| **Expected Result — UI (Alice)** | **Khi mất mạng:** Subtle indicator xuất hiện (VD: banner nhỏ "Đang kết nối lại..." hoặc icon WS màu amber). MessageInput bị disable. Các lần retry: không có popup spam — chỉ 1 indicator duy nhất. **Sau khi có mạng lại:** Indicator biến mất. MessageInput enable lại. WS kết nối lại. Nếu có tin nhắn mới trong lúc offline: tự động load lại conversation (invalidate TanStack Query cache). |
| **Expected Result — DB / WS / API** | `useWebSocket` hook implement backoff: Attempt 1: 1s, Attempt 2: 2s, Attempt 3: 4s, Attempt 4: 8s, ..., max: 30s. Mỗi reconnect gửi `GET /v1/ws?token=<current_access_token>`. Nếu access token hết hạn trong lúc offline: interceptor refresh trước khi reconnect. Sau reconnect thành công: client gửi `join_conversation` cho tất cả conversations đang active. |

---

### 5.5 Nhiều Tab Cùng Lúc — WS Chỉ Kết Nối 1 Lần (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-EC-04` |
| **Tên** | User mở 2 tab trình duyệt cùng lúc — cả 2 nhận được realtime events |
| **Pre-conditions** | Alice đã login ở Tab 1. Alice mở Tab 2 (cùng origin, share cookie). |
| **Steps** | 1. Alice mở conversation với Bob trên Tab 1. 2. Alice mở `/conversations` trên Tab 2. 3. Bob gửi tin nhắn. |
| **Expected Result — UI** | **Tab 1:** Message xuất hiện realtime. **Tab 2:** Unread badge tăng trên conversation item trong sidebar. Cả 2 tab đều nhận được WS event. |
| **Expected Result — DB / WS / API** | Server Hub: `clients["alice_id"]` là một slice chứa **2 client entries** (2 WS connections). `dispatch()` gửi `new_message` đến **cả 2 connections**. DB: chỉ 1 `refresh_token` (share cookie). Server xử lý 2 concurrent connections từ cùng user_id bình thường (không kick connection cũ). |

---

### 5.6 Typing Indicator — Throttle và Auto-Clear (Edge Case)

| Trường | Nội dung |
|--------|---------|
| **ID** | `WS-EC-05` |
| **Tên** | Typing indicator chỉ gửi 1 lần mỗi 3s và tự xóa sau 5s khi không có `typing_stop` |
| **Pre-conditions** | Alice và Bob đang mở cùng conversation. |
| **Steps** | **Case A:** Alice gõ liên tục 10 giây (nhấn phím mỗi 500ms). **Case B:** Alice gõ rồi dừng đột ngột (không gửi tin, không `typing_stop`). |
| **Expected Result — UI (Bob) — Case A** | "Alice đang gõ..." xuất hiện. Trong 10 giây, indicator **không nhấp nháy** hay biến mất (throttle hoạt động đúng — không spam events). |
| **Expected Result — UI (Bob) — Case B** | Typing indicator xuất hiện. Sau **đúng 5s** kể từ lần `typing_start` cuối cùng: indicator tự biến mất (server auto-clear). |
| **Expected Result — DB / WS / API** | **Case A:** Client chỉ gửi `typing_start` event **1 lần mỗi 3s** (throttle trong `MessageInput`). Server broadcast `user_typing {is_typing:true}` đến Bob. **Case B:** Server có timer 5s cho mỗi typing state. Sau 5s không có `typing_start` hay `typing_stop`: server tự broadcast `user_typing {is_typing:false}` đến Bob. DB: không có gì lưu vào DB cho typing events (ephemeral). |

---

## 6. Use Case 5 — Registration

### 6.1 Đăng Ký Thành Công (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-HP-02` |
| **Tên** | Đăng ký tài khoản thành công → tự động đăng nhập |
| **Pre-conditions** | Email `newuser@test.com` và username `new_user` chưa tồn tại trong DB. |
| **Steps** | 1. Mở `/register`. 2. Nhập `email=newuser@test.com`, `username=new_user`, `password=Secure@1234`, `display_name=New User`. 3. Click "Đăng ký". |
| **Expected Result — UI** | Spinner trên button. Sau ≤ 500ms: redirect sang `/conversations`. Sidebar hiển thị avatar mặc định và tên "New User". |
| **Expected Result — DB / WS / API** | `POST /v1/auth/register` → HTTP 201. Response chứa `access_token` + cookie `refresh_token`. DB: INSERT `users` 1 row (`status='ONLINE'`, `password_hash` là bcrypt). |

---

### 6.2 Đăng Ký Email Trùng (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-UP-03` |
| **Tên** | Đăng ký với email đã tồn tại → 409 Conflict |
| **Pre-conditions** | `user_alice` đã tồn tại với `alice@test.com`. |
| **Steps** | 1. Mở `/register`. 2. Nhập `email=alice@test.com`, `username=another_user`, `password=Test@1234`. 3. Click "Đăng ký". |
| **Expected Result — UI** | Global error box: "Email đã được sử dụng". |
| **Expected Result — DB / WS / API** | `POST /v1/auth/register` → HTTP 409. `{ "error": { "code": "CONFLICT" } }`. DB: không có row mới. |

---

### 6.3 Đăng Ký Username Trùng (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-UP-04` |
| **Tên** | Đăng ký với username đã tồn tại → 409 Conflict |
| **Pre-conditions** | `user_alice` đã tồn tại với `alice_dev`. |
| **Steps** | 1. Nhập `email=new@test.com`, `username=alice_dev`, `password=Test@1234`. 2. Click "Đăng ký". |
| **Expected Result — UI** | Global error box: "Username đã được sử dụng". |
| **Expected Result — DB / WS / API** | HTTP 409. DB: không có row mới. |

---

### 6.4 Username Có Ký Tự Đặc Biệt (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-UP-05` |
| **Tên** | Username chứa ký tự không hợp lệ → 400 |
| **Steps** | Nhập `username=alice@dev!` (chứa @, !) → click "Đăng ký". |
| **Expected Result — UI** | Inline error dưới field username: "Chỉ chấp nhận a-z, 0-9 và dấu gạch dưới". |
| **Expected Result — DB / WS / API** | HTTP 400 `VALIDATION_ERROR`. |

---

### 6.5 Logout Thành Công (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `AUTH-HP-03` |
| **Tên** | Logout → token revoke + WS disconnect + status OFFLINE |
| **Pre-conditions** | `user_alice` đã login, WS connected, đang ở `/conversations`. |
| **Steps** | 1. Alice vào Settings. 2. Click "Đăng xuất". |
| **Expected Result — UI** | Redirect về `/login`. Cookie bị clear. |
| **Expected Result — DB / WS / API** | `POST /v1/auth/logout` → 200. DB: `refresh_token` bị revoke, `users.status='OFFLINE'`, `last_seen_at=NOW()`. WS connection bị đóng. Presence broadcast `OFFLINE` đến bạn bè. |

---

## 7. Use Case 6 — Group Chat

### 7.1 Tạo Nhóm 3 Người (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `GRP-HP-01` |
| **Tên** | Tạo nhóm 3 người → notifications realtime |
| **Pre-conditions** | Alice là bạn của Bob và Charlie. Tất cả đang online. |
| **Steps** | 1. Alice mở "Tạo nhóm mới". 2. Đặt tên "Team Alpha". 3. Chọn Bob và Charlie. 4. Click "Tạo nhóm". |
| **Expected Result — UI** | Group xuất hiện đầu sidebar của cả 3 user. Bob và Charlie nhận notification "Alice đã tạo nhóm Team Alpha". |
| **Expected Result — DB / WS / API** | `POST /v1/conversations/groups` → 201. DB: INSERT `conversations` (type='GROUP'), INSERT 3 `conversation_participants` (Alice=ADMIN, Bob=MEMBER, Charlie=MEMBER). WS: `notification_new` push đến Bob, Charlie. |

---

### 7.2 Tạo Nhóm Dưới 3 Người (Unhappy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `GRP-UP-01` |
| **Tên** | Tạo nhóm chỉ có 2 người → 400 |
| **Steps** | Alice chỉ chọn 1 thành viên khác (Bob) → click "Tạo nhóm". |
| **Expected Result — UI** | Button "Tạo nhóm" bị disabled. Warning: "Cần ít nhất 2 thành viên khác". |
| **Expected Result — DB / WS / API** | Client-side validation chặn trước. Nếu bypass: HTTP 400 `VALIDATION_ERROR`. |

---

### 7.3 Admin Kick Thành Viên (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `GRP-HP-02` |
| **Tên** | Admin kick thành viên → WS event `member_removed` |
| **Pre-conditions** | Group "Team Alpha" tồn tại. Alice là ADMIN. Bob là MEMBER. |
| **Steps** | 1. Alice mở group settings. 2. Click "Xóa" bên cạnh Bob. 3. Xác nhận. |
| **Expected Result — UI** | Bob biến mất khỏi danh sách thành viên. Bob: group biến mất khỏi sidebar. |
| **Expected Result — DB / WS / API** | `DELETE /v1/conversations/:id/members/:bob_id` → 200. DB: `conversation_participants.left_at = NOW()`. WS: `member_removed` event broadcast. |

---

## 8. Use Case 7 — Notifications Management

### 8.1 Mark Notification Read (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `NOTIF-HP-01` |
| **Tên** | Mark 1 notification read → badge giảm |
| **Pre-conditions** | Alice có 3 thông báo chưa đọc. |
| **Steps** | 1. Alice mở notification panel. 2. Click vào 1 notification. |
| **Expected Result — UI** | Bell badge giảm từ 3 → 2. Notification item chuyển sang trạng thái "read" (nhạt màu). |
| **Expected Result — DB / WS / API** | `PATCH /v1/notifications/:id/read` → 200. DB: `notifications.is_read=true`, `read_at=NOW()`. |

---

### 8.2 Mark All Notifications Read (Happy Path)

| Trường | Nội dung |
|--------|---------|
| **ID** | `NOTIF-HP-02` |
| **Tên** | Mark all read → badge = 0 |
| **Pre-conditions** | Alice có nhiều thông báo chưa đọc. |
| **Steps** | 1. Alice mở notification panel. 2. Click "Đánh dấu tất cả đã đọc". |
| **Expected Result — UI** | Bell badge biến mất (count = 0). Tất cả items chuyển sang trạng thái read. |
| **Expected Result — DB / WS / API** | `PATCH /v1/notifications/read-all` → 200. DB: tất cả `is_read=true`. Response: `{ "updated_count": N }`. |

---

## 9. Bảng Tổng Hợp Test Cases

| ID | Tên | Loại | Priority | Use Case |
|----|-----|------|---------|---------|
| `AUTH-HP-01` | Đăng nhập thành công | Happy Path | P0 | Authentication |
| `AUTH-HP-02` | Đăng ký thành công → tự động login | Happy Path | P0 | Registration |
| `AUTH-HP-03` | Logout → revoke + WS disconnect | Happy Path | P0 | Authentication |
| `AUTH-UP-01` | Sai mật khẩu — no enumeration | Unhappy Path | P0 | Authentication |
| `AUTH-UP-02` | Rate limit brute force | Unhappy Path | P0 | Authentication |
| `AUTH-UP-03` | Đăng ký email trùng → 409 | Unhappy Path | P0 | Registration |
| `AUTH-UP-04` | Đăng ký username trùng → 409 | Unhappy Path | P0 | Registration |
| `AUTH-UP-05` | Username ký tự đặc biệt → 400 | Unhappy Path | P0 | Registration |
| `AUTH-EC-01` | Silent refresh qua Axios interceptor | Edge Case | P0 | Authentication |
| `AUTH-EC-02` | Token reuse detection | Edge Case | P1 | Authentication |
| `FRD-HP-01` | Gửi lời mời → notification realtime | Happy Path | P0 | Friend System |
| `FRD-HP-02` | Accept lời mời → tạo DM tự động | Happy Path | P0 | Friend System |
| `FRD-UP-01` | Gửi lời mời cho người đã là bạn | Unhappy Path | P0 | Friend System |
| `FRD-UP-02` | Duplicate pending request | Unhappy Path | P0 | Friend System |
| `FRD-UP-03` | Tự gửi lời mời cho mình | Unhappy Path | P0 | Friend System |
| `FRD-UP-04` | Chat với người chưa kết bạn | Unhappy Path | P0 | Friend System |
| `MSG-HP-01` | Gửi text → optimistic UI + realtime | Happy Path | P0 | Messaging |
| `MSG-HP-02` | Read receipt → tick xanh | Happy Path | P1 | Messaging |
| `MSG-HP-03` | Upload ảnh → thumbnail inline | Happy Path | P0 | Messaging |
| `MSG-EC-01` | Upload file 51MB — bị chặn client | Edge Case | P0 | Messaging |
| `MSG-EC-02` | Upload file đúng 50MB — phải pass | Edge Case | P0 | Messaging |
| `MSG-EC-03` | Cursor pagination — scroll lên | Edge Case | P0 | Messaging |
| `MSG-UP-01` | Gửi tin khi đã unfriend | Unhappy Path | P0 | Messaging |
| `GRP-HP-01` | Tạo nhóm 3 người → notifications | Happy Path | P0 | Group Chat |
| `GRP-HP-02` | Admin kick thành viên → WS event | Happy Path | P0 | Group Chat |
| `GRP-UP-01` | Tạo nhóm dưới 3 người → 400 | Unhappy Path | P0 | Group Chat |
| `NOTIF-HP-01` | Mark 1 notification read → badge giảm | Happy Path | P0 | Notifications |
| `NOTIF-HP-02` | Mark all read → badge = 0 | Happy Path | P0 | Notifications |
| `WS-HP-01` | Đóng tab → 60s grace → offline | Happy Path | P0 | WS Resilience |
| `WS-EC-01` | Reconnect trong grace period | Edge Case | P1 | WS Resilience |
| `WS-EC-02` | Zombie connection — no ping | Edge Case | P1 | WS Resilience |
| `WS-EC-03` | Mất mạng → exponential backoff | Edge Case | P0 | WS Resilience |
| `WS-EC-04` | Multi-tab — cả 2 nhận events | Edge Case | P1 | WS Resilience |
| `WS-EC-05` | Typing throttle + auto-clear | Edge Case | P2 | WS Resilience |

---

*Tổng: **33 test cases** — 10 Happy Path · 11 Unhappy Path · 12 Edge Cases. P0: 22 · P1: 8 · P2: 3.*
