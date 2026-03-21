# 3. API Contract & WebSocket Events Specification

> **Version:** 1.0.0
> **Dựa trên:** `2_Database_Design.md` v1.0.0
> **Base URL:** `https://api.yourdomain.com/v1`
> **Auth:** Bearer JWT (Access Token) — header `Authorization: Bearer <token>`
> **Content-Type:** `application/json` (trừ upload endpoints)
> **Timezone:** Tất cả timestamp trả về dạng **ISO 8601 UTC** — `2025-01-15T08:30:00Z`

---

## Quy ước chung

### Response Envelope

Mọi response đều bọc trong envelope thống nhất:

```json
// Success
{
  "success": true,
  "data": { ... },
  "meta": { "cursor": "...", "has_more": true }  // chỉ có ở paginated endpoints
}

// Error
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Email hoặc mật khẩu không đúng",
    "details": {}   // optional: field-level validation errors
  }
}
```

### Error Codes chuẩn

| HTTP | Code | Ý nghĩa |
|------|------|---------|
| 400 | `VALIDATION_ERROR` | Body/params không hợp lệ |
| 401 | `UNAUTHORIZED` | Thiếu hoặc token hết hạn |
| 403 | `FORBIDDEN` | Không có quyền |
| 404 | `NOT_FOUND` | Resource không tồn tại |
| 409 | `CONFLICT` | Duplicate resource |
| 413 | `FILE_TOO_LARGE` | Vượt giới hạn kích thước |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | MIME type không cho phép |
| 429 | `RATE_LIMITED` | Quá nhiều request |
| 500 | `INTERNAL_ERROR` | Lỗi server |

### Middleware Stack (Golang)

| Middleware | Mô tả |
|-----------|-------|
| `RequestID` | Gán `X-Request-ID` UUID cho mỗi request, log tracking |
| `Logger` | Log method, path, status, latency |
| `CORS` | Whitelist `vercel.app` domain + custom domain |
| `RateLimiter` | `golang.org/x/time/rate` — 100 req/min per IP (public), 300/min (authed) |
| `AuthMiddleware` | Parse & validate JWT Access Token, inject `userID` vào context |
| `RequireAuth` | Kiểm tra `userID` trong context, trả 401 nếu không có |
| `RequireConversationMember` | Verify user là active participant của conversation |
| `RequireGroupAdmin` | Verify user có `role = ADMIN` trong group |

---

## PHẦN I — REST API

---

## 1. Auth

### 1.1 Đăng ký

| Field | Value |
|-------|-------|
| **Router** | `/v1/auth` |
| **Endpoint** | `POST /v1/auth/register` |
| **Mô tả** | Tạo tài khoản mới, tự động đăng nhập sau khi đăng ký thành công |
| **Method** | `POST` |
| **Middleware** | `RequestID`, `Logger`, `RateLimiter(20/hour per IP)` |

**Input (req.body):**

```json
{
  "email": "alice@example.com",
  "username": "alice_dev",
  "password": "Secure@1234",
  "display_name": "Alice"
}
```

**Output (res.body) — `201 Created`:**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid-v4",
      "email": "alice@example.com",
      "username": "alice_dev",
      "display_name": "Alice",
      "bio": null,
      "avatar_url": null,
      "created_at": "2025-01-15T08:30:00Z"
    },
    "access_token": "eyJhbGci...",
    "expires_in": 900
  }
}
```

> `Refresh Token` được set qua **`httpOnly; Secure; SameSite=Strict` cookie** — không trả trong body.

**Errors (Unhappy path):**

```json
// 409 — Email đã tồn tại
{ "success": false, "error": { "code": "CONFLICT", "message": "Email đã được sử dụng" } }

// 409 — Username đã tồn tại
{ "success": false, "error": { "code": "CONFLICT", "message": "Username đã được sử dụng" } }

// 400 — Validation
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dữ liệu không hợp lệ",
    "details": {
      "password": "Mật khẩu tối thiểu 8 ký tự",
      "username": "Chỉ cho phép a-z, 0-9, dấu gạch dưới"
    }
  }
}
```

**Notes:**

- Hash password với `bcrypt` cost=12 trước khi INSERT
- `username` validate regex `^[a-z0-9_]{3,30}$` — lowercase enforce ở backend
- Sau register: INSERT `refresh_token`, set cookie, trả `access_token` trong body

**Packages (Go):**

- `golang.org/x/crypto/bcrypt` — hash password
- `github.com/golang-jwt/jwt/v5` — generate JWT
- `github.com/go-playground/validator/v10` — validate input

**Test:**

```
✓ POST với data hợp lệ → 201 + access_token + cookie set
✓ POST email đã tồn tại → 409
✓ POST username đã tồn tại → 409
✓ POST password < 8 ký tự → 400
✓ POST username có ký tự đặc biệt → 400
✓ POST thiếu field bắt buộc → 400
✓ 20+ requests/hour từ cùng IP → 429
```

---

### 1.2 Đăng nhập

| Field | Value |
|-------|-------|
| **Router** | `/v1/auth` |
| **Endpoint** | `POST /v1/auth/login` |
| **Mô tả** | Xác thực tài khoản, phát hành JWT Access Token + Refresh Token |
| **Method** | `POST` |
| **Middleware** | `RequestID`, `Logger`, `RateLimiter(10/min per IP)` |

**Input (req.body):**

```json
{
  "email": "alice@example.com",
  "password": "Secure@1234"
}
```

**Output (res.body) — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid-v4",
      "email": "alice@example.com",
      "username": "alice_dev",
      "display_name": "Alice",
      "bio": "Backend engineer",
      "avatar_url": "https://res.cloudinary.com/demo/image/upload/v1/avatars/alice.jpg",
      "status": "ONLINE",
      "created_at": "2025-01-15T08:30:00Z"
    },
    "access_token": "eyJhbGci...",
    "expires_in": 900
  }
}
```

**Errors (Unhappy path):**

```json
// 401 — Sai thông tin
{ "success": false, "error": { "code": "INVALID_CREDENTIALS", "message": "Email hoặc mật khẩu không đúng" } }

// 400 — Thiếu field
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "Email và mật khẩu là bắt buộc" } }
```

**Notes:**

- Dùng thông báo lỗi chung "Email hoặc mật khẩu không đúng" — không phân biệt để tránh user enumeration attack
- Sau login thành công: `UPDATE users SET status = 'ONLINE', last_seen_at = NOW()`
- Tạo Refresh Token mới, lưu hash vào DB, set `httpOnly` cookie

**Packages (Go):** `bcrypt`, `jwt/v5`

**Test:**

```
✓ Login đúng → 200 + tokens
✓ Sai password → 401 (message không nói rõ field nào sai)
✓ Email không tồn tại → 401 (cùng message)
✓ Brute force 10+ req/min → 429
```

---

### 1.3 Refresh Token

| Field | Value |
|-------|-------|
| **Router** | `/v1/auth` |
| **Endpoint** | `POST /v1/auth/refresh` |
| **Mô tả** | Đổi Refresh Token (trong cookie) lấy Access Token mới. Implement Refresh Token Rotation |
| **Method** | `POST` |
| **Middleware** | `RequestID`, `Logger` |

**Input:** Không có body — đọc `refresh_token` từ `httpOnly` cookie.

**Output (res.body) — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGci...",
    "expires_in": 900
  }
}
```

**Errors (Unhappy path):**

```json
// 401 — Cookie không tồn tại
{ "success": false, "error": { "code": "UNAUTHORIZED", "message": "Refresh token không hợp lệ" } }

// 401 — Token đã bị revoke hoặc hết hạn
{ "success": false, "error": { "code": "UNAUTHORIZED", "message": "Phiên đăng nhập đã hết hạn" } }
```

**Notes:**

- **Rotation:** Mỗi lần refresh → revoke token cũ (`is_revoked = TRUE`) + tạo token mới → set cookie mới
- Nếu token cũ đã bị revoke → có thể là token theft → revoke **toàn bộ** refresh tokens của user đó
- SHA-256 raw token trước khi so sánh với DB

**Packages (Go):** `jwt/v5`, `crypto/sha256`

**Test:**

```
✓ Cookie hợp lệ → 200 + access_token mới + cookie mới
✓ Dùng lại token cũ sau rotation → 401 + revoke tất cả tokens của user
✓ Cookie expired → 401
✓ Cookie bị tamper → 401
```

---

### 1.4 Đăng xuất

| Field | Value |
|-------|-------|
| **Router** | `/v1/auth` |
| **Endpoint** | `POST /v1/auth/logout` |
| **Mô tả** | Revoke Refresh Token, xóa cookie, set status OFFLINE |
| **Method** | `POST` |
| **Middleware** | `RequireAuth` |

**Input:** Không có body.

**Output — `200 OK`:**

```json
{ "success": true, "data": { "message": "Đăng xuất thành công" } }
```

**Notes:**

- `UPDATE refresh_tokens SET is_revoked = TRUE WHERE token_hash = $hash`
- `UPDATE users SET status = 'OFFLINE', last_seen_at = NOW()`
- Clear cookie: `Set-Cookie: refresh_token=; Max-Age=0; HttpOnly; ...`

**Test:**

```
✓ Logout hợp lệ → 200 + cookie cleared
✓ Gọi lại endpoint với cookie cũ → 401 (token đã revoke)
✓ Không có auth header → 401
```

---

### 1.5 Lấy Profile Hiện Tại

| Field | Value |
|-------|-------|
| **Router** | `/v1/users` |
| **Endpoint** | `GET /v1/users/me` |
| **Mô tả** | Lấy thông tin profile của user đang đăng nhập. Gọi khi app load hoặc refresh page |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": "uuid-v4",
    "email": "alice@example.com",
    "username": "alice_dev",
    "display_name": "Alice",
    "bio": "Full-stack developer",
    "avatar_url": "https://res.cloudinary.com/...",
    "status": "ONLINE",
    "created_at": "2025-01-15T08:30:00Z"
  }
}
```

**Test:**

```
✓ Token hợp lệ → 200 + user data
✓ Token hết hạn → 401 (trigger silent refresh)
✓ Không có auth → 401
```

---

### 1.6 Cập nhật Profile

| Field | Value |
|-------|-------|
| **Router** | `/v1/users` |
| **Endpoint** | `PATCH /v1/users/me` |
| **Mô tả** | Cập nhật display_name, bio. Avatar update có endpoint riêng |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{
  "display_name": "Alice Updated",
  "bio": "Full-stack developer | Coffee addict ☕"
}
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": "uuid-v4",
    "username": "alice_dev",
    "display_name": "Alice Updated",
    "bio": "Full-stack developer | Coffee addict ☕",
    "avatar_url": "https://res.cloudinary.com/...",
    "updated_at": "2025-01-15T09:00:00Z"
  }
}
```

**Errors (Unhappy path):**

```json
// 400 — bio quá dài
{ "success": false, "error": { "code": "VALIDATION_ERROR", "details": { "bio": "Tối đa 160 ký tự" } } }
```

**Notes:**

- Chỉ PATCH các field được gửi lên (partial update) — dùng `sql.NullString` hoặc pointer trong Go struct
- `display_name` max 50 ký tự, `bio` max 160 ký tự

**Test:**

```
✓ PATCH display_name → 200 + updated data
✓ PATCH bio 161 ký tự → 400
✓ PATCH body rỗng → 200 (no-op, trả data hiện tại)
✓ Không có auth → 401
```

---

## 2. Upload (Presigned URL)

### 2.1 Lấy Presigned URL — Avatar (Cloudinary)

| Field | Value |
|-------|-------|
| **Router** | `/v1/upload` |
| **Endpoint** | `POST /v1/upload/avatar/presign` |
| **Mô tả** | Server tạo Cloudinary Upload Signature. Client dùng để upload thẳng lên Cloudinary mà không qua server |
| **Method** | `POST` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{
  "file_name": "my_photo.jpg",
  "mime_type": "image/jpeg",
  "file_size": 2048576
}
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "upload_url": "https://api.cloudinary.com/v1_1/{cloud_name}/image/upload",
    "upload_params": {
      "api_key": "123456789",
      "timestamp": 1705312200,
      "signature": "sha1_signature_string",
      "folder": "avatars",
      "public_id": "avatars/user_uuid_v4",
      "transformation": "c_fill,w_256,h_256,q_auto,f_auto",
      "eager": "c_fill,w_64,h_64"
    },
    "cloudinary_public_id": "avatars/user_uuid_v4",
    "expires_at": "2025-01-15T08:35:00Z"
  }
}
```

**Sau khi upload xong, client gọi tiếp:**

```
PATCH /v1/users/me/avatar
Body: { "cloudinary_public_id": "avatars/user_uuid_v4", "avatar_url": "https://res.cloudinary.com/..." }
```

**Errors (Unhappy path):**

```json
// 415 — MIME type không cho phép
{ "success": false, "error": { "code": "UNSUPPORTED_MEDIA_TYPE", "message": "Chỉ chấp nhận: image/jpeg, image/png, image/webp, image/gif" } }

// 413 — Quá lớn
{ "success": false, "error": { "code": "FILE_TOO_LARGE", "message": "Ảnh tối đa 10MB" } }
```

**Notes:**

- Server **không nhận file** — chỉ tạo signature. File đi thẳng từ client → Cloudinary
- Signature hết hạn sau **5 phút** (`timestamp + 300s`)
- Whitelist MIME: `image/jpeg`, `image/png`, `image/webp`, `image/gif`
- `public_id` = `avatars/{user_id}` — cố định theo user, tự động overwrite ảnh cũ trên Cloudinary
- Khi PATCH avatar thành công: nếu `avatar_cloudinary_id` cũ khác `public_id` mới → gọi Cloudinary Delete API để xóa ảnh cũ

**Packages (Go):** `github.com/cloudinary/cloudinary-go/v2`

**Test:**

```
✓ JPEG < 10MB → 200 + valid signature
✓ PNG > 10MB → 413
✓ MIME = video/mp4 → 415
✓ Signature hết hạn sau 5 phút → Cloudinary từ chối (kiểm tra phía Cloudinary)
```

---

### 2.2 Cập nhật Avatar sau khi upload

| Field | Value |
|-------|-------|
| **Router** | `/v1/users` |
| **Endpoint** | `PATCH /v1/users/me/avatar` |
| **Mô tả** | Lưu avatar_url và avatar_cloudinary_id vào DB sau khi client upload thành công lên Cloudinary |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{
  "avatar_url": "https://res.cloudinary.com/demo/image/upload/v1/avatars/user_uuid.jpg",
  "cloudinary_public_id": "avatars/user_uuid_v4"
}
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "avatar_url": "https://res.cloudinary.com/demo/image/upload/v1/avatars/user_uuid.jpg",
    "updated_at": "2025-01-15T09:00:00Z"
  }
}
```

**Notes:**

- Validate `avatar_url` phải là Cloudinary domain (`res.cloudinary.com`) — tránh user nhét URL tùy ý
- Nếu user đã có `avatar_cloudinary_id` cũ → gọi Cloudinary Destroy API xóa ảnh cũ (fire-and-forget, không block response)

**Test:**

```
✓ URL hợp lệ từ Cloudinary → 200
✓ URL từ domain khác → 400
✓ Không có auth → 401
```

---

### 2.3 Lấy Presigned URL — File/Video (Cloudflare R2)

| Field | Value |
|-------|-------|
| **Router** | `/v1/upload` |
| **Endpoint** | `POST /v1/upload/file/presign` |
| **Mô tả** | Server tạo R2 Presigned PUT URL. Client upload trực tiếp lên R2 |
| **Method** | `POST` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{
  "conversation_id": "uuid-v4",
  "file_name": "project_report.pdf",
  "mime_type": "application/pdf",
  "file_size": 5242880,
  "file_type": "FILE"
}
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "upload_url": "https://{account}.r2.cloudflarestorage.com/bucket/files/uuid_filename.pdf?X-Amz-Signature=...",
    "r2_key": "files/conversation_uuid/2025/01/uuid_filename.pdf",
    "expires_at": "2025-01-15T08:35:00Z",
    "method": "PUT",
    "headers": {
      "Content-Type": "application/pdf",
      "Content-Length": "5242880"
    }
  }
}
```

**Sau khi upload xong, client gọi tiếp:**

```
POST /v1/conversations/:id/messages
Body: {
  "type": "ATTACHMENT",
  "attachment": {
    "r2_key": "files/conversation_uuid/2025/01/uuid_filename.pdf",
    "file_name": "project_report.pdf",
    "mime_type": "application/pdf",
    "file_size": 5242880,
    "file_type": "FILE"
  }
}
```

**Errors (Unhappy path):**

```json
// 415 — MIME không trong whitelist
{ "success": false, "error": { "code": "UNSUPPORTED_MEDIA_TYPE", "message": "Loại file không được hỗ trợ" } }

// 413 — File > 50MB (FILE), > 100MB (VIDEO)
{ "success": false, "error": { "code": "FILE_TOO_LARGE", "message": "File tối đa 50MB" } }

// 403 — User không trong conversation
{ "success": false, "error": { "code": "FORBIDDEN", "message": "Bạn không phải thành viên của cuộc trò chuyện này" } }
```

**Notes:**

- R2 Key format: `{file_type}/{conversation_id}/{YYYY/MM}/{uuid}_{original_name}`
- Presigned URL PUT, expire 5 phút
- Whitelist FILE: `application/pdf`, `application/msword`, `application/vnd.openxmlformats-officedocument.*`, `application/zip`, `text/plain`
- Whitelist VIDEO: `video/mp4`, `video/quicktime`
- Giới hạn: FILE ≤ 50MB, VIDEO ≤ 100MB, IMAGE ≤ 10MB (qua Cloudinary)
- R2 bucket cần CORS policy cho phép PUT từ Vercel domain

**Packages (Go):** `github.com/aws/aws-sdk-go-v2/service/s3` (S3-compatible R2)

**Test:**

```
✓ PDF < 50MB, user trong conversation → 200 + presigned URL
✓ ZIP > 50MB → 413
✓ MIME = image/jpeg (dùng endpoint sai) → 415
✓ User không trong conversation → 403
✓ Presigned URL dùng sau 5 phút → R2 từ chối (403 từ R2)
```

---

## 3. Users & Friends

### 3.1 Tìm kiếm User

| Field | Value |
|-------|-------|
| **Router** | `/v1/users` |
| **Endpoint** | `GET /v1/users/search` |
| **Mô tả** | Tìm user theo username hoặc display_name. Kết quả kèm relationship_status với người đang đăng nhập |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Input (query params):**

```
GET /v1/users/search?q=alice&limit=10
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-v4",
      "username": "alice_dev",
      "display_name": "Alice",
      "avatar_url": "https://res.cloudinary.com/...",
      "status": "ONLINE",
      "relationship": "NONE"
    },
    {
      "id": "uuid-v4-2",
      "username": "alice_design",
      "display_name": "Alice D.",
      "avatar_url": null,
      "status": "OFFLINE",
      "relationship": "PENDING_SENT"
    }
  ]
}
```

> **`relationship` values:** `NONE` | `PENDING_SENT` | `PENDING_RECEIVED` | `ACCEPTED`

**Notes:**

- Tìm kiếm với `pg_trgm` ILIKE trên cả `username` và `display_name`
- **Không trả về chính user đang đăng nhập** (`WHERE id <> $current_user_id`)
- Minimum query length: 2 ký tự
- Default limit: 10, max: 20

**Test:**

```
✓ q="alice" → list users kèm relationship status đúng
✓ q="a" (1 ký tự) → 400
✓ q="" → 400
✓ Tự tìm chính mình → không có trong kết quả
```

---

### 3.2 Gửi Lời Mời Kết Bạn

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `POST /v1/friends/requests` |
| **Mô tả** | Gửi lời mời kết bạn đến user khác |
| **Method** | `POST` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{ "to_user_id": "uuid-v4" }
```

**Output — `201 Created`:**

```json
{
  "success": true,
  "data": {
    "friendship_id": "uuid-v4",
    "status": "PENDING",
    "created_at": "2025-01-15T08:30:00Z"
  }
}
```

**Errors (Unhappy path):**

```json
// 409 — Đã gửi request hoặc đã là bạn
{ "success": false, "error": { "code": "CONFLICT", "message": "Lời mời đã được gửi hoặc đã là bạn bè" } }

// 404 — User không tồn tại
{ "success": false, "error": { "code": "NOT_FOUND", "message": "Người dùng không tồn tại" } }

// 400 — Tự gửi cho mình
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "Không thể gửi lời mời cho chính mình" } }
```

**Notes:**

- Sau khi INSERT friendship: INSERT notification cho `to_user_id` với `type = FRIEND_REQUEST`
- Nếu có WebSocket connection của recipient → push WS event `notification_new`
- Kiểm tra cả 2 chiều trong `friendships` table trước khi INSERT

**Test:**

```
✓ Gửi đến user hợp lệ → 201 + notification tạo ra
✓ Gửi lại khi đã PENDING → 409
✓ Gửi khi đã ACCEPTED → 409
✓ to_user_id = current_user_id → 400
✓ to_user_id không tồn tại → 404
```

---

### 3.3 Chấp Nhận / Từ Chối Lời Mời

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `PATCH /v1/friends/requests/:request_id` |
| **Mô tả** | Accept hoặc Decline lời mời kết bạn đang PENDING |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{ "action": "ACCEPT" }
// hoặc
{ "action": "DECLINE" }
```

**Output — `200 OK` (ACCEPT):**

```json
{
  "success": true,
  "data": {
    "friendship_id": "uuid-v4",
    "status": "ACCEPTED",
    "conversation_id": "uuid-conv",
    "friend": {
      "id": "uuid-v4",
      "username": "bob_smith",
      "display_name": "Bob",
      "avatar_url": "https://res.cloudinary.com/..."
    }
  }
}
```

**Output — `200 OK` (DECLINE):**

```json
{ "success": true, "data": { "message": "Đã từ chối lời mời kết bạn" } }
```

**Errors (Unhappy path):**

```json
// 403 — Không phải người nhận lời mời
{ "success": false, "error": { "code": "FORBIDDEN", "message": "Bạn không có quyền xử lý lời mời này" } }

// 404 — Request không tồn tại
{ "success": false, "error": { "code": "NOT_FOUND", "message": "Lời mời không tồn tại" } }

// 409 — Đã xử lý rồi
{ "success": false, "error": { "code": "CONFLICT", "message": "Lời mời này đã được xử lý" } }
```

**Notes:**

- Chỉ `addressee_id` mới được PATCH request này
- **ACCEPT:** `UPDATE friendships SET status = 'ACCEPTED'` + `INSERT INTO conversations (type='DM')` + `INSERT INTO conversation_participants` (cả 2 user) + tạo notification `FRIEND_ACCEPTED` cho requester
- **DECLINE:** `DELETE FROM friendships WHERE id = $1` (hard delete — record không còn giá trị)
- Toàn bộ ACCEPT operation phải trong **1 DB transaction**

**Test:**

```
✓ ACCEPT hợp lệ → 200 + conversation_id + notification cho requester
✓ DECLINE hợp lệ → 200 + friendship record bị xóa
✓ ACCEPT khi không phải addressee → 403
✓ ACCEPT request đã ACCEPTED → 409
✓ request_id không tồn tại → 404
```

---

### 3.4 Hủy Lời Mời Đã Gửi

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `DELETE /v1/friends/requests/:request_id` |
| **Mô tả** | Người gửi hủy lời mời đang PENDING |
| **Method** | `DELETE` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{ "success": true, "data": { "message": "Đã hủy lời mời kết bạn" } }
```

**Errors:** 403 nếu không phải `requester_id`, 404 nếu không tồn tại, 409 nếu đã ACCEPTED.

**Notes:** Hard delete record. Xóa notification tương ứng của recipient.

**Test:**

```
✓ Cancel khi PENDING, đúng requester → 200
✓ Cancel khi đã ACCEPTED → 409
✓ Cancel của người khác → 403
```

---

### 3.5 Danh Sách Bạn Bè

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `GET /v1/friends` |
| **Mô tả** | Lấy danh sách bạn bè của user hiện tại |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "friendship_id": "uuid-v4",
      "user": {
        "id": "uuid-v4",
        "username": "bob_smith",
        "display_name": "Bob",
        "avatar_url": "https://res.cloudinary.com/...",
        "status": "ONLINE",
        "last_seen_at": "2025-01-15T08:25:00Z"
      },
      "conversation_id": "uuid-conv",
      "became_friends_at": "2025-01-10T12:00:00Z"
    }
  ]
}
```

**Notes:** `conversation_id` dùng để navigate thẳng vào DM khi click vào bạn bè.

---

### 3.6 Danh Sách Lời Mời Nhận Được

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `GET /v1/friends/requests/received` |
| **Mô tả** | Lấy danh sách lời mời kết bạn chờ xử lý |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "request_id": "uuid-v4",
      "from_user": {
        "id": "uuid-v4",
        "username": "charlie_99",
        "display_name": "Charlie",
        "avatar_url": null
      },
      "requested_at": "2025-01-15T07:00:00Z"
    }
  ]
}
```

---

### 3.7 Hủy Kết Bạn

| Field | Value |
|-------|-------|
| **Router** | `/v1/friends` |
| **Endpoint** | `DELETE /v1/friends/:friendship_id` |
| **Mô tả** | Xóa quan hệ bạn bè. Lịch sử chat không bị xóa |
| **Method** | `DELETE` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{ "success": true, "data": { "message": "Đã hủy kết bạn" } }
```

**Notes:** Hard delete friendship record. Conversation và messages **giữ nguyên** — user vẫn thấy lịch sử chat nhưng không thể gửi tin mới (vì gate kết bạn ở `POST /messages`).

---

## 4. Conversations

### 4.1 Danh Sách Conversations

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `GET /v1/conversations` |
| **Mô tả** | Lấy toàn bộ conversations của user, kèm unread count và tin nhắn cuối. Sorted theo `last_activity_at DESC` |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-conv-1",
      "type": "DM",
      "name": null,
      "avatar_url": null,
      "other_user": {
        "id": "uuid-v4",
        "username": "bob_smith",
        "display_name": "Bob",
        "avatar_url": "https://res.cloudinary.com/...",
        "status": "ONLINE"
      },
      "last_message": {
        "id": "uuid-msg",
        "body": "Ok nhé, gặp lúc 3h",
        "type": "TEXT",
        "sender_id": "uuid-v4",
        "created_at": "2025-01-15T08:28:00Z"
      },
      "unread_count": 3,
      "last_activity_at": "2025-01-15T08:28:00Z"
    },
    {
      "id": "uuid-conv-2",
      "type": "GROUP",
      "name": "Team Alpha",
      "avatar_url": "https://res.cloudinary.com/...",
      "other_user": null,
      "last_message": {
        "id": "uuid-msg-2",
        "body": null,
        "type": "ATTACHMENT",
        "sender_id": "uuid-v4-3",
        "sender_name": "Charlie",
        "created_at": "2025-01-15T08:00:00Z"
      },
      "unread_count": 0,
      "last_activity_at": "2025-01-15T08:00:00Z"
    }
  ]
}
```

**Notes:**

- `other_user` chỉ có cho DM, `null` cho GROUP
- `last_message.body = null` nếu là ATTACHMENT — frontend render "📎 File đính kèm"
- Query đơn theo thiết kế mục 3.1 trong `2_Database_Design.md`

**Test:**

```
✓ User có conversations → 200 + list sorted đúng
✓ User mới (chưa có conversations) → 200 + data: []
✓ unread_count tính đúng (messages sau last_read_at, sender ≠ me)
```

---

### 4.2 Chi Tiết Conversation

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `GET /v1/conversations/:id` |
| **Mô tả** | Lấy thông tin chi tiết 1 conversation (dùng khi mở chat) |
| **Method** | `GET` |
| **Middleware** | `RequireAuth`, `RequireConversationMember` |

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": "uuid-conv",
    "type": "GROUP",
    "name": "Team Alpha",
    "avatar_url": "https://res.cloudinary.com/...",
    "participants": [
      {
        "user_id": "uuid-v4",
        "username": "alice_dev",
        "display_name": "Alice",
        "avatar_url": "https://res.cloudinary.com/...",
        "status": "ONLINE",
        "role": "ADMIN",
        "joined_at": "2025-01-10T12:00:00Z"
      }
    ],
    "created_at": "2025-01-10T12:00:00Z"
  }
}
```

---

### 4.3 Tạo Group Chat

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `POST /v1/conversations/groups` |
| **Mô tả** | Tạo group chat mới. Chỉ thêm được bạn bè vào nhóm |
| **Method** | `POST` |
| **Middleware** | `RequireAuth` |

**Input (req.body):**

```json
{
  "name": "Team Alpha",
  "member_ids": ["uuid-bob", "uuid-charlie", "uuid-dave"]
}
```

**Output — `201 Created`:**

```json
{
  "success": true,
  "data": {
    "id": "uuid-conv-new",
    "type": "GROUP",
    "name": "Team Alpha",
    "avatar_url": null,
    "participant_count": 4,
    "created_at": "2025-01-15T09:00:00Z"
  }
}
```

**Errors (Unhappy path):**

```json
// 400 — Thêm < 2 người khác
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "Nhóm cần ít nhất 3 thành viên (bao gồm bạn)" } }

// 400 — member_id không phải bạn bè
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "Một số người dùng không phải bạn bè của bạn", "details": { "invalid_ids": ["uuid-unknown"] } } }
```

**Notes:**

- Validate: tất cả `member_ids` phải là bạn bè (`ACCEPTED`) của creator
- Creator tự động là ADMIN
- Sau tạo nhóm: INSERT notifications `GROUP_ADDED` cho tất cả member_ids, push WS event

**Test:**

```
✓ 3 member_ids hợp lệ (đều là bạn) → 201
✓ 1 member_id không phải bạn → 400 + danh sách id lỗi
✓ member_ids < 2 → 400
✓ member_ids trùng nhau → deduplicate hoặc 400
```

---

### 4.4 Lịch Sử Tin Nhắn (Cursor-Based Pagination)

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `GET /v1/conversations/:id/messages` |
| **Mô tả** | Load lịch sử tin nhắn, cursor-based pagination. Scroll lên để load thêm |
| **Method** | `GET` |
| **Middleware** | `RequireAuth`, `RequireConversationMember` |

**Input (query params):**

```
GET /v1/conversations/uuid-conv/messages?limit=30&before_id=uuid-msg&before_time=2025-01-15T08:00:00Z
```

| Param | Bắt buộc | Mô tả |
|-------|---------|-------|
| `limit` | Không | Default 30, max 50 |
| `before_id` | Không | UUID của message làm cursor (dùng với `before_time`) |
| `before_time` | Không | Timestamp của cursor message |

> Lần đầu load: không truyền cursor — server trả 30 messages mới nhất.

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-msg-1",
      "conversation_id": "uuid-conv",
      "sender": {
        "id": "uuid-v4",
        "username": "alice_dev",
        "display_name": "Alice",
        "avatar_url": "https://res.cloudinary.com/..."
      },
      "body": "Mọi người ơi, họp lúc 3h nha",
      "type": "TEXT",
      "status": "READ",
      "attachment": null,
      "created_at": "2025-01-15T08:28:00Z",
      "deleted_at": null
    },
    {
      "id": "uuid-msg-2",
      "conversation_id": "uuid-conv",
      "sender": {
        "id": "uuid-bob",
        "username": "bob_smith",
        "display_name": "Bob",
        "avatar_url": null
      },
      "body": null,
      "type": "ATTACHMENT",
      "status": "SENT",
      "attachment": {
        "id": "uuid-att",
        "file_type": "IMAGE",
        "url": "https://res.cloudinary.com/demo/image/upload/v1/chat/img.jpg",
        "thumbnail_url": "https://res.cloudinary.com/demo/image/upload/c_fill,w_300/v1/chat/img.jpg",
        "original_name": "screenshot.jpg",
        "mime_type": "image/jpeg",
        "size_bytes": 1048576,
        "width": 1920,
        "height": 1080
      },
      "created_at": "2025-01-15T08:20:00Z",
      "deleted_at": null
    }
  ],
  "meta": {
    "has_more": true,
    "next_cursor": {
      "before_id": "uuid-oldest-in-batch",
      "before_time": "2025-01-15T07:00:00Z"
    }
  }
}
```

**Notes:**

- Messages trả về theo thứ tự `ASC` (cũ → mới) để render đúng chiều chat — server query `DESC` rồi reverse
- `deleted_at != null` → render "Tin nhắn đã bị xóa" thay vì body (body sẽ là null sau soft delete)
- Sau khi load messages: tự động trigger `PATCH /conversations/:id/read` để mark-as-read

**Test:**

```
✓ Lần đầu (không cursor) → 30 messages mới nhất, has_more=true nếu còn
✓ Với cursor hợp lệ → 30 messages cũ hơn cursor
✓ Hết messages → has_more=false, data=[]
✓ limit=51 → clamp về 50
✓ User không trong conversation → 403
```

---

### 4.5 Gửi Tin Nhắn (REST fallback)

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `POST /v1/conversations/:id/messages` |
| **Mô tả** | Gửi tin nhắn TEXT hoặc ATTACHMENT qua REST. Dùng cho ATTACHMENT sau khi upload R2/Cloudinary xong. Text message nên dùng WebSocket |
| **Method** | `POST` |
| **Middleware** | `RequireAuth`, `RequireConversationMember` |

**Input — Text:**

```json
{ "type": "TEXT", "body": "Xin chào mọi người!" }
```

**Input — Attachment:**

```json
{
  "type": "ATTACHMENT",
  "attachment": {
    "storage_type": "R2",
    "file_type": "FILE",
    "r2_key": "files/conv_uuid/2025/01/uuid_report.pdf",
    "original_name": "Q4_Report.pdf",
    "mime_type": "application/pdf",
    "size_bytes": 5242880
  }
}
```

**Output — `201 Created`:**

```json
{
  "success": true,
  "data": {
    "id": "uuid-new-msg",
    "conversation_id": "uuid-conv",
    "sender": { "id": "uuid-v4", "display_name": "Alice" },
    "body": null,
    "type": "ATTACHMENT",
    "status": "SENT",
    "attachment": {
      "id": "uuid-att",
      "file_type": "FILE",
      "url": "https://r2.example.com/files/...",
      "original_name": "Q4_Report.pdf",
      "mime_type": "application/pdf",
      "size_bytes": 5242880
    },
    "created_at": "2025-01-15T09:05:00Z"
  }
}
```

**Notes:**

- Sau INSERT message: `UPDATE conversations SET last_message_id, last_activity_at`
- Broadcast WS event `new_message` đến tất cả participants
- Kiểm tra DM gate: nếu `type='DM'` và 2 user không còn là bạn → 403

**Test:**

```
✓ TEXT message hợp lệ → 201 + WS broadcast
✓ ATTACHMENT với r2_key hợp lệ → 201
✓ TEXT không có body → 400
✓ DM khi không còn là bạn → 403
✓ User không trong conversation → 403
```

---

### 4.6 Mark Conversation As Read

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `PATCH /v1/conversations/:id/read` |
| **Mô tả** | Cập nhật last_read_message_id và last_read_at cho participant hiện tại |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth`, `RequireConversationMember` |

**Input (req.body):**

```json
{ "last_message_id": "uuid-msg-latest" }
```

**Output — `200 OK`:**

```json
{ "success": true, "data": { "message": "Đã cập nhật trạng thái đọc" } }
```

**Notes:**

- Sau UPDATE `conversation_participants`: UPDATE messages status = 'READ' cho DM
- Broadcast WS event `read_receipt` đến sender

---

### 4.7 Thêm Thành Viên Vào Nhóm

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `POST /v1/conversations/:id/members` |
| **Mô tả** | Admin thêm thành viên mới vào group. Chỉ được thêm bạn bè |
| **Method** | `POST` |
| **Middleware** | `RequireAuth`, `RequireGroupAdmin` |

**Input (req.body):**

```json
{ "user_ids": ["uuid-new-member"] }
```

**Output — `200 OK`:**

```json
{ "success": true, "data": { "added_count": 1 } }
```

**Notes:** INSERT `conversation_participants` + INSERT notifications `GROUP_ADDED` + WS push.

---

### 4.8 Xóa Thành Viên / Rời Nhóm

| Field | Value |
|-------|-------|
| **Router** | `/v1/conversations` |
| **Endpoint** | `DELETE /v1/conversations/:id/members/:user_id` |
| **Mô tả** | Admin kick thành viên hoặc thành viên tự rời nhóm |
| **Method** | `DELETE` |
| **Middleware** | `RequireAuth`, `RequireConversationMember` |

**Notes:**

- Nếu `:user_id = me` → rời nhóm (ai cũng làm được)
- Nếu `:user_id ≠ me` → phải là ADMIN mới kick được
- Nếu ADMIN rời nhóm và còn thành viên → promote thành viên lâu nhất lên ADMIN
- `UPDATE conversation_participants SET left_at = NOW()` (soft remove)

---

## 5. Notifications

### 5.1 Lấy Danh Sách Thông Báo

| Field | Value |
|-------|-------|
| **Router** | `/v1/notifications` |
| **Endpoint** | `GET /v1/notifications` |
| **Mô tả** | Lấy danh sách in-app notifications, mới nhất trước |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Input (query params):**

```
GET /v1/notifications?limit=20&offset=0
```

**Output — `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-notif-1",
      "type": "FRIEND_REQUEST",
      "is_read": false,
      "actor": {
        "id": "uuid-charlie",
        "username": "charlie_99",
        "display_name": "Charlie",
        "avatar_url": null
      },
      "reference_id": "uuid-friendship",
      "reference_type": "friendship",
      "preview_text": "Charlie đã gửi lời mời kết bạn",
      "created_at": "2025-01-15T07:00:00Z"
    },
    {
      "id": "uuid-notif-2",
      "type": "GROUP_ADDED",
      "is_read": true,
      "actor": {
        "id": "uuid-alice",
        "username": "alice_dev",
        "display_name": "Alice",
        "avatar_url": "https://res.cloudinary.com/..."
      },
      "reference_id": "uuid-conv-group",
      "reference_type": "conversation",
      "preview_text": "Alice đã thêm bạn vào nhóm Team Alpha",
      "created_at": "2025-01-14T15:00:00Z",
      "read_at": "2025-01-14T16:00:00Z"
    }
  ],
  "meta": {
    "unread_count": 1,
    "total": 15,
    "has_more": false
  }
}
```

**Notes:** `preview_text` được generate ở server, không lưu DB — tránh migration khi đổi copy.

---

### 5.2 Số Thông Báo Chưa Đọc

| Field | Value |
|-------|-------|
| **Router** | `/v1/notifications` |
| **Endpoint** | `GET /v1/notifications/unread-count` |
| **Mô tả** | Lấy số lượng thông báo chưa đọc cho bell badge |
| **Method** | `GET` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{ "success": true, "data": { "count": 3 } }
```

---

### 5.3 Mark Notification As Read

| Field | Value |
|-------|-------|
| **Router** | `/v1/notifications` |
| **Endpoint** | `PATCH /v1/notifications/:id/read` |
| **Mô tả** | Đánh dấu 1 thông báo đã đọc |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{ "success": true, "data": { "message": "Đã đánh dấu đã đọc" } }
```

**Notes:** Verify `recipient_id = current_user_id` trước khi UPDATE.

---

### 5.4 Mark All Notifications As Read

| Field | Value |
|-------|-------|
| **Router** | `/v1/notifications` |
| **Endpoint** | `PATCH /v1/notifications/read-all` |
| **Mô tả** | Đánh dấu tất cả thông báo chưa đọc là đã đọc |
| **Method** | `PATCH` |
| **Middleware** | `RequireAuth` |

**Output — `200 OK`:**

```json
{ "success": true, "data": { "updated_count": 3 } }
```

---

## PHẦN II — WebSocket Events

---

## 6. Kiến Trúc Nội Bộ — PostgreSQL LISTEN/NOTIFY (Thay thế Redis Pub/Sub)

> **⚠️ Lưu ý quan trọng khi implement WebSocket Hub trong Golang.**

Thay vì dựng thêm Redis chỉ để làm pub/sub, MVP tận dụng cơ chế **`LISTEN/NOTIFY` native của PostgreSQL** thông qua thư viện `pgx/v5`. Đây là lựa chọn phù hợp khi chạy **single instance** trên fly.io.

### 6.1 Luồng hoạt động

```
REST Handler (Goroutine A)              WebSocket Hub (Goroutine B)
        │                                         │
        │  INSERT message vào DB                  │  conn.WaitForNotification(ctx)
        │                                         │  (blocking, chạy vòng lặp liên tục)
        ▼                                         │
  NOTIFY chat_events, '<json>'  ──────────────►  │  Nhận Notification{}
                                                  │       │
                                                  │       ▼
                                                  │  Parse JSON payload
                                                  │       │
                                                  │       ▼
                                                  │  Tìm clients đang kết nối
                                                  │  theo conversation_id / user_id
                                                  │       │
                                                  │       ▼
                                                  │  Push WS frame đến từng client
```

### 6.2 NOTIFY payload chuẩn

Mọi lệnh `NOTIFY` đều dùng chung channel `chat_events`, phân biệt nhau qua field `type`:

```go
// Golang — gọi sau INSERT trong REST handler
payload, _ := json.Marshal(map[string]any{
    "type":            "new_message",          // khớp với WS event name
    "conversation_id": msg.ConversationID,
    "recipient_ids":   []string{"uuid-bob"},   // nil = broadcast tất cả participants
    "data":            msg,                    // struct đã marshal
})
conn.Exec(ctx, "SELECT pg_notify('chat_events', $1)", string(payload))
```

**Các `type` được NOTIFY:**

| `type` | Trigger tại REST endpoint | Recipient |
|--------|--------------------------|-----------|
| `new_message` | `POST /conversations/:id/messages` | Tất cả participants |
| `notification_new` | `POST /friends/requests`, Accept, `POST /groups` | `recipient_ids` cụ thể |
| `read_receipt` | `PATCH /conversations/:id/read` | sender của DM |
| `presence_update` | Login / Logout / WS disconnect | Users có conversation chung |
| `conversation_updated` | Đổi tên/avatar nhóm, thêm/kick member | Tất cả participants |

### 6.3 WebSocket Hub skeleton (Go)

```go
// internal/hub/hub.go
type Hub struct {
    clients    map[string][]*Client   // userID → danh sách WS connections
    mu         sync.RWMutex
    listenConn *pgx.Conn              // Connection riêng dành cho LISTEN — KHÔNG dùng pool
}

func (h *Hub) ListenLoop(ctx context.Context) {
    h.listenConn.Exec(ctx, "LISTEN chat_events")

    for {
        notification, err := h.listenConn.WaitForNotification(ctx)
        if err != nil {
            // reconnect với backoff nếu connection bị drop
            h.reconnect(ctx)
            continue
        }
        go h.dispatch(notification.Payload)   // xử lý bất đồng bộ, không block loop
    }
}

func (h *Hub) dispatch(rawPayload string) {
    var event ChatEvent
    json.Unmarshal([]byte(rawPayload), &event)

    h.mu.RLock()
    defer h.mu.RUnlock()

    for _, recipientID := range event.RecipientIDs {
        for _, client := range h.clients[recipientID] {
            client.send <- event.Data   // non-blocking channel send
        }
    }
}
```

> **Lưu ý quan trọng:**
>
> - `listenConn` là **dedicated connection**, tách biệt hoàn toàn với connection pool (`pgxpool`) dùng cho query thông thường — LISTEN sẽ block connection đó.
> - Payload của `pg_notify` bị giới hạn **8000 bytes**. Nếu `data` lớn hơn: chỉ NOTIFY `message_id`, Hub tự query lại DB để lấy full data.
> - Khi scale lên multi-instance sau này: thay `pg_notify` bằng Redis Pub/Sub mà **không cần đổi WS Hub interface** — chỉ swap phần `ListenLoop`.

**Packages (Go):**

- `github.com/jackc/pgx/v5` — `WaitForNotification`, `pg_notify`
- `github.com/jackc/pgx/v5/pgxpool` — connection pool riêng cho REST handlers

---

## 7. WebSocket Connection

### 7.1 Kết nối

```
WSS wss://api.yourdomain.com/v1/ws?token=<access_token>
```

- **Auth:** Access Token truyền qua query param `?token=` (WebSocket không hỗ trợ custom header)
- **Timeout:** Server đóng connection sau 60s không có ping
- **Reconnect:** Client tự reconnect với exponential backoff (1s, 2s, 4s, 8s, max 30s)
- **Protocol:** JSON text frames

### 7.2 Base Event Shape

Mọi WS message (cả client→server và server→client) đều theo cấu trúc:

```json
{
  "event": "event_name",
  "payload": { ... },
  "id": "client-generated-uuid"  // optional, dùng để ack
}
```

---

## 8. Client → Server Events

### 8.1 `ping`

```json
{ "event": "ping", "payload": {} }
```

> Gửi mỗi 30s. Server phản hồi bằng `pong`. Dùng để giữ connection và heartbeat presence.

---

### 8.2 `join_conversation`

```json
{
  "event": "join_conversation",
  "payload": {
    "conversation_id": "uuid-conv"
  }
}
```

> Đăng ký nhận events của conversation này. Server verify user là participant trước khi join room. Nên gọi khi mở conversation view.

---

### 8.3 `leave_conversation`

```json
{
  "event": "leave_conversation",
  "payload": {
    "conversation_id": "uuid-conv"
  }
}
```

> Hủy đăng ký room. Gọi khi đóng conversation hoặc navigate away.

---

### 8.4 `send_message`

```json
{
  "event": "send_message",
  "id": "client-uuid-for-ack",
  "payload": {
    "conversation_id": "uuid-conv",
    "body": "Mọi người ơi, họp lúc 3h nha!",
    "type": "TEXT",
    "client_temp_id": "temp-uuid-for-optimistic-ui"
  }
}
```

> **`client_temp_id`:** UUID do client tạo để render optimistic UI ngay lập tức. Khi nhận lại `new_message` từ server, client thay thế temp message bằng message thực (với `id` từ DB).

**Server response (ack):**

```json
{
  "event": "message_sent",
  "payload": {
    "client_temp_id": "temp-uuid-for-optimistic-ui",
    "message_id": "uuid-real-msg",
    "created_at": "2025-01-15T09:05:00Z",
    "status": "SENT"
  }
}
```

---

### 7.5 `typing_start`

```json
{
  "event": "typing_start",
  "payload": {
    "conversation_id": "uuid-conv"
  }
}
```

> Server broadcast `user_typing` đến các participant khác. Throttle phía client: chỉ gửi mỗi 3s nếu vẫn đang gõ.

---

### 7.6 `typing_stop`

```json
{
  "event": "typing_stop",
  "payload": {
    "conversation_id": "uuid-conv"
  }
}
```

> Gửi khi user dừng gõ hoặc gửi tin nhắn.

---

### 8.7 `mark_read`

```json
{
  "event": "mark_read",
  "payload": {
    "conversation_id": "uuid-conv",
    "last_message_id": "uuid-latest-msg"
  }
}
```

> Server UPDATE `last_read_message_id`, broadcast `read_receipt` đến sender.

---

## 9. Server → Client Events

### 9.1 `pong`

```json
{ "event": "pong", "payload": { "server_time": "2025-01-15T09:05:30Z" } }
```

---

### 9.2 `new_message`

Broadcast đến tất cả active participants của conversation khi có tin nhắn mới.

```json
{
  "event": "new_message",
  "payload": {
    "message": {
      "id": "uuid-real-msg",
      "conversation_id": "uuid-conv",
      "sender": {
        "id": "uuid-alice",
        "username": "alice_dev",
        "display_name": "Alice",
        "avatar_url": "https://res.cloudinary.com/..."
      },
      "body": "Mọi người ơi, họp lúc 3h nha!",
      "type": "TEXT",
      "status": "DELIVERED",
      "attachment": null,
      "created_at": "2025-01-15T09:05:00Z"
    },
    "client_temp_id": "temp-uuid-for-optimistic-ui"
  }
}
```

> **`client_temp_id`:** Chỉ có trong payload gửi về **chính sender** — dùng để replace optimistic message. Các recipient khác không nhận field này.

---

### 9.3 `message_sent` (ack về sender)

Xem mục 7.4.

---

### 9.4 `user_typing`

```json
{
  "event": "user_typing",
  "payload": {
    "conversation_id": "uuid-conv",
    "user": {
      "id": "uuid-bob",
      "display_name": "Bob"
    },
    "is_typing": true
  }
}
```

> Server auto-clear typing indicator sau 5s nếu không nhận `typing_stop`.

---

### 9.5 `read_receipt`

Gửi về sender khi recipient mark-as-read. Chỉ áp dụng cho DM.

```json
{
  "event": "read_receipt",
  "payload": {
    "conversation_id": "uuid-conv",
    "reader_id": "uuid-bob",
    "last_read_message_id": "uuid-msg",
    "read_at": "2025-01-15T09:06:00Z"
  }
}
```

---

### 9.6 `presence_update`

Broadcast khi user online/offline. Gửi đến tất cả user có conversation chung.

```json
{
  "event": "presence_update",
  "payload": {
    "user_id": "uuid-bob",
    "status": "OFFLINE",
    "last_seen_at": "2025-01-15T09:10:00Z"
  }
}
```

---

### 9.7 `notification_new`

Push thông báo realtime khi user đang online.

```json
{
  "event": "notification_new",
  "payload": {
    "notification": {
      "id": "uuid-notif",
      "type": "FRIEND_REQUEST",
      "actor": {
        "id": "uuid-charlie",
        "display_name": "Charlie",
        "avatar_url": null
      },
      "reference_id": "uuid-friendship",
      "reference_type": "friendship",
      "preview_text": "Charlie đã gửi lời mời kết bạn",
      "created_at": "2025-01-15T09:00:00Z"
    },
    "unread_count": 4
  }
}
```

> **`unread_count`:** Tổng thông báo chưa đọc sau khi thêm cái mới — dùng cập nhật bell badge ngay lập tức.

---

### 9.8 `conversation_updated`

Broadcast khi metadata group thay đổi (tên, avatar, thêm/xóa thành viên).

```json
{
  "event": "conversation_updated",
  "payload": {
    "conversation_id": "uuid-conv",
    "changes": {
      "name": "Team Alpha v2",
      "avatar_url": "https://res.cloudinary.com/..."
    },
    "updated_by": {
      "id": "uuid-alice",
      "display_name": "Alice"
    }
  }
}
```

---

### 9.9 `member_added`

```json
{
  "event": "member_added",
  "payload": {
    "conversation_id": "uuid-conv",
    "new_member": {
      "id": "uuid-dave",
      "username": "dave_k",
      "display_name": "Dave",
      "avatar_url": null,
      "role": "MEMBER"
    },
    "added_by": { "id": "uuid-alice", "display_name": "Alice" }
  }
}
```

---

### 9.10 `member_removed`

```json
{
  "event": "member_removed",
  "payload": {
    "conversation_id": "uuid-conv",
    "removed_user_id": "uuid-dave",
    "removed_by": { "id": "uuid-alice", "display_name": "Alice" },
    "reason": "KICKED"
  }
}
```

> **`reason`:** `KICKED` | `LEFT`

---

### 9.11 `error`

Server gửi về client khi WS event có lỗi (auth, validation, ...).

```json
{
  "event": "error",
  "payload": {
    "code": "FORBIDDEN",
    "message": "Bạn không phải thành viên của cuộc trò chuyện này",
    "ref_event": "join_conversation"
  }
}
```

---

## 10. Tổng Hợp Sơ Đồ WebSocket Events

```
CLIENT                              SERVER
  │                                   │
  │──── connect (/ws?token=...) ─────►│  Validate JWT, mark ONLINE
  │                                   │  ◄── presence_update (broadcast)
  │                                   │
  │──── join_conversation ───────────►│  Verify member, join room
  │                                   │
  │──── send_message ────────────────►│  INSERT message
  │◄─── message_sent (ack) ──────────│  (chỉ về sender, kèm client_temp_id)
  │                                   │
  │                                   │──── new_message ──────────────────►│ (broadcast to all participants)
  │                                   │
  │──── typing_start ────────────────►│
  │                                   │──── user_typing ──────────────────►│ (broadcast to others)
  │──── typing_stop ─────────────────►│
  │                                   │
  │──── mark_read ───────────────────►│  UPDATE last_read_message_id
  │                                   │──── read_receipt ─────────────────►│ (về sender)
  │                                   │
  │──── ping (every 30s) ────────────►│
  │◄─── pong ────────────────────────│
  │                                   │
  │──── disconnect ──────────────────►│  60s grace period
  │                                   │  mark OFFLINE, update last_seen_at
  │                                   │  ◄── presence_update (broadcast)
```

---

## 11. Tổng Hợp Endpoints

| Method | Endpoint | Mô tả | Auth |
|--------|----------|-------|------|
| POST | `/v1/auth/register` | Đăng ký | ✗ |
| POST | `/v1/auth/login` | Đăng nhập | ✗ |
| POST | `/v1/auth/refresh` | Refresh token | Cookie |
| POST | `/v1/auth/logout` | Đăng xuất | ✓ |
| GET | `/v1/users/me` | Lấy profile hiện tại | ✓ |
| PATCH | `/v1/users/me` | Cập nhật profile | ✓ |
| PATCH | `/v1/users/me/avatar` | Cập nhật avatar URL | ✓ |
| GET | `/v1/users/search` | Tìm kiếm user | ✓ |
| POST | `/v1/upload/avatar/presign` | Lấy Cloudinary signature | ✓ |
| POST | `/v1/upload/file/presign` | Lấy R2 presigned URL | ✓ |
| GET | `/v1/friends` | Danh sách bạn bè | ✓ |
| GET | `/v1/friends/requests/received` | Lời mời nhận được | ✓ |
| GET | `/v1/friends/requests/sent` | Lời mời đã gửi | ✓ |
| POST | `/v1/friends/requests` | Gửi lời mời | ✓ |
| PATCH | `/v1/friends/requests/:id` | Accept/Decline | ✓ |
| DELETE | `/v1/friends/requests/:id` | Hủy lời mời | ✓ |
| DELETE | `/v1/friends/:id` | Hủy kết bạn | ✓ |
| GET | `/v1/conversations` | Danh sách conversations | ✓ |
| GET | `/v1/conversations/:id` | Chi tiết conversation | ✓ |
| POST | `/v1/conversations/groups` | Tạo group | ✓ |
| PATCH | `/v1/conversations/:id` | Cập nhật tên/avatar nhóm | Admin |
| GET | `/v1/conversations/:id/messages` | Lịch sử tin nhắn | ✓ |
| POST | `/v1/conversations/:id/messages` | Gửi tin nhắn (REST) | ✓ |
| PATCH | `/v1/conversations/:id/read` | Mark as read | ✓ |
| POST | `/v1/conversations/:id/members` | Thêm thành viên | Admin |
| DELETE | `/v1/conversations/:id/members/:uid` | Kick/Rời nhóm | ✓ |
| GET | `/v1/notifications` | Danh sách thông báo | ✓ |
| GET | `/v1/notifications/unread-count` | Số chưa đọc | ✓ |
| PATCH | `/v1/notifications/:id/read` | Mark 1 read | ✓ |
| PATCH | `/v1/notifications/read-all` | Mark all read | ✓ |
| WSS | `/v1/ws?token=` | WebSocket connection | Token |

---

*Mọi thay đổi API Contract phải versioned (`/v2/...`) — không breaking change trên endpoint hiện có.*
