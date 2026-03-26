# Security Audit Request — YunoChat Application

## Role
Đóng vai một Senior Application Security Engineer và Penetration Tester với kinh nghiệm 
kiểm thử các hệ thống chat/messaging và ứng dụng Next.js + Node.js.

## Task
Quét toàn bộ mã nguồn trong thư mục `/backend` và `/frontend` của dự án Chat App 
này để tìm kiếm, phân tích và báo cáo các lỗ hổng bảo mật. 
Đọc kỹ từng file — đừng bỏ qua logic phức tạp hay helper functions.

---

## Focus Areas

### 1. Authentication & Session Management
- Refresh Token có được lưu trong `httpOnly + Secure + SameSite` cookie không?
- Access Token có bị lưu vào `localStorage` (dễ bị XSS đánh cắp) không?
- Middleware xác thực JWT có bao phủ **đầy đủ** các route nhạy cảm không?  
  Liệt kê cụ thể các route bị bỏ sót (nếu có).
- JWT secret có đủ mạnh và không bị hard-code trong source không?
- Cơ chế Refresh Token Rotation và Token Revocation (blacklist) có được triển khai không?

### 2. Authorization & IDOR (Insecure Direct Object Reference)
- Endpoint `/conversations/:id` và `/conversations/:id/messages`:  
  Backend có xác minh **người gọi API là participant** của conversation không?  
  (Chống User A thay đổi `:id` để đọc trộm tin nhắn của nhóm B)
- Các thao tác nhạy cảm (xóa tin nhắn, thêm/xóa thành viên, đổi tên nhóm) có kiểm tra  
  quyền (owner/admin) hay chỉ kiểm tra "đã đăng nhập"?

### 3. WebSocket Security
- Cơ chế xác thực khi upgrade HTTP → WS có an toàn không?  
  Token có bị truyền qua query string (dễ bị log) không?
- Sau khi kết nối WS, server có re-validate quyền truy cập theo từng sự kiện không,  
  hay chỉ validate một lần lúc connect?
- Có Rate Limiting / Throttling trên WS để chống spam message flood không?

### 4. File Upload Security (Cloudinary / R2 / Presigned URL)
- Backend có validate `mime_type` và `file_size` **trước khi** cấp Presigned URL không?
- Chỉ kiểm tra extension (dễ bypass) hay kiểm tra Magic Bytes thực sự?
- Có giới hạn loại file được phép upload không (chặn `.exe`, `.sh`, `.php`...)?
- Presigned URL có bị cấp với thời hạn quá dài hoặc không giới hạn không?

### 5. Injection & XSS
- Giao diện Next.js có sử dụng `dangerouslySetInnerHTML` hoặc render HTML thô không?  
  Nội dung tin nhắn có bị sanitize trước khi render không?
- Backend có sanitize/validate input trước khi lưu vào DB không?  
  Kiểm tra nguy cơ **NoSQL Injection** (nếu dùng MongoDB).
- Markdown/rich-text renderer (nếu có) có bị bypass để chèn script không?

### 6. CORS & HTTP Security Headers
- CORS `origin` có được whitelist cứng, hay đang dùng wildcard `*` hoặc  
  reflect `Origin` header không an toàn?
- Các HTTP Security Headers có được cấu hình đủ không:  
  `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options`,  
  `Strict-Transport-Security (HSTS)`, `Permissions-Policy`?

### 7. Rate Limiting & Brute Force Protection
- Endpoint `/auth/login`, `/auth/register`, `/auth/forgot-password` có Rate Limiting không?
- Có cơ chế chống brute force (lockout, CAPTCHA, exponential backoff) không?

### 8. Secrets & Configuration Management
- Có API Key, DB connection string, JWT secret nào bị hard-code trong source không?
- File `.env.example` hoặc `.env` có vô tình bị commit lên repo không?
- Các biến môi trường phía client (Next.js `NEXT_PUBLIC_*`) có chứa thông tin nhạy cảm không?

### 9. Error Handling & Information Disclosure
- Response lỗi có trả về stack trace, tên DB, cấu trúc thư mục, hoặc thông tin  
  nội bộ nhạy cảm không?
- Production environment có tắt verbose logging không?

### 10. Dependency Vulnerabilities
- Kiểm tra `package.json` (cả backend và frontend) xem có dependencies nào chứa  
  CVE đã biết (mức HIGH/CRITICAL) không?

---

## Output Format

Lập báo cáo bảo mật theo chuẩn Markdown với cấu trúc sau:
```
# Security Audit Report — YunoChat
**Date:** ...  
**Auditor:** Senior AppSec Engineer  
**Scope:** /backend, /frontend

---

## Executive Summary
(Tổng quan: X lỗi CRITICAL, Y lỗi HIGH, Z lỗi MEDIUM, W lỗi LOW)

---

## Findings

### [VULN-001] Tên lỗ hổng
| Attribute       | Detail                              |
|-----------------|-------------------------------------|
| Severity        | CRITICAL / HIGH / MEDIUM / LOW      |
| CVSS Score      | e.g., 9.1                           |
| File(s)         | `backend/src/routes/auth.js:L42`    |
| Attack Vector   | Remote / Local                      |
| Authentication  | Required / Not Required             |

**Description:** Giải thích rủi ro và tác động thực tế.

**Proof of Concept:**
\```
// Đoạn code/request minh họa cách khai thác
\```

**Remediation:**
\```
// Đoạn code đã được sửa
\```

---
(Lặp lại cho từng finding)

---

## Passed Checks
- ✅ Mục X: Không tìm thấy lỗ hổng.

## Remediation Roadmap
(Ưu tiên xử lý theo thứ tự: CRITICAL → HIGH → MEDIUM → LOW)
```

**Lưu ý quan trọng:**
- Nếu một mục hoàn toàn an toàn, ghi rõ **"Passed ✅"** kèm lý do.
- Mỗi finding phải có file cụ thể, số dòng (nếu có), và code sửa thực tế — không chỉ mô tả chung chung.
- Ưu tiên tìm các lỗi có thể khai thác được trong thực tế (exploitable), không chỉ  
  lý thuyết.