# Documentation Compliance Audit — YunoChat

## Role
Đóng vai một Strict Software Architect và QA Lead, với nhiệm vụ thực hiện vòng 
nghiệm thu phần mềm (UAT) cuối cùng trước khi release.  
Bạn là người **không có tình cảm với code** — chỉ quan tâm đến sự thật: 
code có khớp với tài liệu thiết kế hay không.

---

## Task
**Bước 1 — Đọc tài liệu chuẩn mực:**  
Đọc kỹ toàn bộ 6 file trong thư mục `/docs`:
- `1_MVP_Requirements.md` — Yêu cầu tính năng
- `2_Database_Design.md` — Schema, Index, Constraint
- `3_API_WebSocket_Specs.md` — API Contract, WebSocket events
- `4_System_Architecture.md` — Kiến trúc hệ thống, Pub/Sub
- `5_Frontend_Design.md` — UI/UX flows, component behavior *(nếu có)*
- `6_Use_Cases_Test_Cases.md` — Edge cases, test scenarios

**Bước 2 — Quét mã nguồn thực tế:**  
Đọc toàn bộ `/backend` và `/frontend`, tập trung vào:
routes, controllers/handlers, models/migrations, WebSocket logic,
middleware, frontend state management, API calls.

**Bước 3 — Thực hiện Gap Analysis:**  
Đối chiếu từng hạng mục bên dưới, ghi nhận mọi sai lệch dù nhỏ.

---

## Compliance Checklist

### 1. Database Schema Consistency
- Schema trong migrations/models có khớp **100%** với `docs/2_Database_Design.md` không?
- Kiểm tra từng bảng:
  - Có thiếu **Index** nào không (đặc biệt composite index cho query phức tạp)?
  - Có thiếu **Constraint** nào không (`ON DELETE RESTRICT`, `ON DELETE CASCADE`,
    `UNIQUE`, `NOT NULL`, `CHECK`)?
  - Có thiếu **Partial Index** nào không (ví dụ: index chỉ áp dụng khi
    `status = 'active'`)?
  - Kiểu dữ liệu (`VARCHAR` vs `TEXT`, `TIMESTAMP` vs `TIMESTAMPTZ`) có khớp không?
  - Có bảng/cột nào trong code **không có trong docs** (schema drift) không?

### 2. API Contract Compliance
- Tất cả endpoints có trả về đúng **Response Envelope** như định nghĩa trong
  `docs/3_API_WebSocket_Specs.md` không?  
  Ví dụ chuẩn: `{ "success": true, "data": {...}, "error": null }`
- **HTTP Status Code** có đúng không?  
  (401 vs 403, 400 vs 422, 404 vs 200-with-empty-data...)
- **Field naming convention** có nhất quán không (camelCase hay snake_case)?  
  Code có trả về tên field khác với docs không?
- **Pagination** có implement đúng kiểu docs yêu cầu không  
  (cursor-based vs offset, tên field `nextCursor` vs `next_page`...)?
- **API Versioning** (`/api/v1/`) có được implement như docs yêu cầu không?
- Có endpoint nào trong docs **chưa được implement** không?
- Có endpoint nào trong code **không có trong docs** (undocumented endpoint) không?

### 3. WebSocket Event Contract
- Tên các **sự kiện WS** (emit/on) trong code có khớp chính xác với
  `docs/3_API_WebSocket_Specs.md` không?  
  (Ví dụ: docs dùng `message:new`, code dùng `new_message` — sai contract)
- **Payload structure** của từng sự kiện có đúng không?
- Có sự kiện nào trong docs **chưa được implement** ở server/client không?
- Có sự kiện nào trong code **không có trong docs** không?

### 4. Business Logic — Friend & Group Rules
- **Luồng kết bạn:** Logic chặn "phải là bạn bè mới được chat 1-1 hoặc được
  add vào group" có thực sự được **enforce ở backend** không,  
  hay chỉ ẩn nút trên UI (frontend-only guard, dễ bypass bằng cURL)?
- **Luồng tạo nhóm:** Điều kiện về số lượng thành viên tối thiểu/tối đa có
  được validate ở backend không?
- **Quyền trong nhóm (Owner/Admin/Member):** Các action nhạy cảm (kick thành viên,
  giải tán nhóm, đổi tên) có được phân quyền đúng theo docs không?
- Liệt kê cụ thể: Những business rule nào **chỉ có trên UI** mà backend không enforce?

### 5. System Architecture — Pub/Sub & Scalability
- Tầng WebSocket có thực sự dùng **PostgreSQL `LISTEN/NOTIFY`** như
  `docs/4_System_Architecture.md` yêu cầu không?  
  Hay đang dùng in-memory channel/EventEmitter nội bộ (sẽ fail khi scale
  multi-instance)?
- **Message Queue / Event Bus** (nếu docs yêu cầu Redis Pub/Sub, BullMQ...):
  Có được implement không?
- **Horizontal scaling:** Kiến trúc hiện tại có thực sự support multi-instance
  deployment như docs thiết kế không?

### 6. Edge Cases & Test Case Coverage
Dựa trên `docs/6_Use_Cases_Test_Cases.md`, kiểm tra từng edge case sau đây
có được **xử lý trong code** không (không chỉ trong test file):

- **Silent Refresh Token:** Khi Access Token hết hạn giữa chừng, client có tự
  động refresh và retry request gốc không, hay bắt user logout?
- **WebSocket Reconnect với Exponential Backoff:** Khi mất kết nối, client có
  implement backoff (1s → 2s → 4s...) và reconnect không?
- **Message Ordering:** Tin nhắn có được đảm bảo thứ tự khi gửi nhanh liên tiếp
  (race condition) không?
- **Concurrent Friend Request:** Nếu A và B cùng gửi lời mời kết bạn cho nhau
  đồng thời, có xảy ra duplicate/deadlock không?
- **File Upload Failure:** Nếu upload lên Cloudinary/R2 thành công nhưng lưu DB
  thất bại, có cơ chế rollback không?
- Kiểm tra **tất cả edge case còn lại** trong docs/6 và báo cáo từng cái.

### 7. Error Handling Standards
- Các error message trả về client có đúng format và nội dung như docs quy định không?
- Có error nào trả về thông tin nhạy cảm (stack trace, tên bảng DB) không?
- `try/catch` có được implement đầy đủ ở các async operations không?

---

## Output Format

### Phần 1 — Compliance Matrix

Tạo bảng với **5 cột** sau:

| # | Hạng mục kiểm tra | Trạng thái | File vi phạm | Chi tiết vấn đề |
|---|-------------------|------------|--------------|-----------------|

**Quy ước cột Trạng thái:**
- ✅ **Tuân thủ** — Code khớp hoàn toàn với docs
- ⚠️ **Thiếu sót** — Code implement nhưng thiếu một phần (ghi rõ thiếu gì)
- ❌ **Vi phạm** — Code implement sai hoặc ngược với docs
- 🚫 **Chưa implement** — Tính năng có trong docs nhưng không có trong code
- ➕ **Ngoài docs** — Code có nhưng docs không đề cập (cần cập nhật docs hoặc xóa code)

---

### Phần 2 — Prioritized To-Do List

Sau bảng matrix, liệt kê danh sách công việc cần sửa, phân loại theo priority:

#### 🔴 P0 — Chặn Release (phải sửa trước khi deploy)
- [ ] **[Database]** Tên vấn đề — Lý do tại sao nguy hiểm

#### 🟠 P1 — Sprint tiếp theo (sửa trong 1-2 tuần)
- [ ] **[API]** Tên vấn đề

#### 🟡 P2 — Backlog (sửa khi có thời gian)
- [ ] **[Docs]** Cập nhật tài liệu cho feature X không có trong docs

---

### Phần 3 — Documentation Debt
Liệt kê các phần của docs **cần được cập nhật** để phản ánh đúng
những gì code đang làm (tránh tài liệu lỗi thời gây hiểu nhầm).

---

**Lưu ý quan trọng:**
- Chỉ báo cáo **sự thật từ code** — không assume, không đoán mò.
- Nếu không đủ thông tin để kết luận, ghi rõ **"Cần xem thêm file: [tên file]"**.
- Ưu tiên **tìm lỗi logic bị thiếu ở backend** hơn là lỗi style/format.