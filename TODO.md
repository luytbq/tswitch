# TODO — tswitch improvement ideas

## 1. Global fuzzy search — tìm kiếm xuyên session + window

**Vấn đề:** Hiện phải drill in thủ công qua từng level để tìm window. Không thể tìm xuyên suốt cùng lúc.

**Ý tưởng:** Thêm chế độ search flat list tổng hợp tất cả `session > window`. Gõ `/` (hoặc phím riêng) hiện flat list, fuzzy match, chọn → switch thẳng. Tương tự Telescope của Neovim.

**Priority:** Cao | **Độ khó:** Trung bình

---

## 2. Tag filtering — lọc session theo tag ngay trên grid

**Vấn đề:** Tags đã được lưu vào `state.yaml` nhưng không có UI nào để xem hay lọc.

**Ý tưởng:** Phím `t` mở tag picker: hiện danh sách tags hiện có, chọn một → grid chỉ hiển thị sessions có tag đó. Thêm tag indicator nhỏ trên card. Hữu ích khi có 20+ sessions.

**Priority:** Cao | **Độ khó:** Thấp

---

## 3. Session grouping — nhóm sessions theo tag trên grid

**Ý tưởng:** Thay vì flat grid, sessions được nhóm thành sections theo tag:

```
[ work ]          [ personal ]      [ infra ]
  ● proj-alpha      dotfiles          k8s-prod
  proj-beta         music             monitoring
```

Hỗ trợ collapse/expand từng group bằng phím.

**Priority:** Trung bình | **Độ khó:** Cao

---

## 4. Quick-command palette (`:`)

**Ý tưởng:** Phím `:` mở command palette kiểu vim, gõ lệnh ngắn:
- `:new api` → tạo session "api"
- `:kill` → kill session đang focus
- `:tag dev` → gắn tag "dev" cho session đang focus
- `:sort alpha` → sort sessions theo alphabet

Giúp power user thao tác nhanh mà không cần nhớ hết keybinding.

**Priority:** Trung bình | **Độ khó:** Cao

---

## 5. Recent sessions — sort theo lịch sử dùng trong tswitch

**Ý tưởng:** tswitch tự lưu timestamp mỗi lần visit session vào `state.yaml`. Thêm `sort_by: recent` trong settings — hiển thị session mới dùng nhất lên đầu. Khác với `sort_by: activity` hiện tại vốn dựa vào tmux timestamp.

**Priority:** Trung bình | **Độ khó:** Rất thấp

---

## 6. Per-session color coding

**Ý tưởng:** Mỗi session được gán một màu accent (tự động hoặc thủ công). Card border đổi màu theo session — giúp nhận diện nhanh khi nhiều session tên tương tự.

Cấu hình trong `tswitch-config.json`:
```json
"session_colors": {
  "work": "blue",
  "personal": "green",
  "infra": "red"
}
```

**Priority:** Thấp (thẩm mỹ) | **Độ khó:** Thấp

---

## 7. Live refresh — tự động cập nhật danh sách sessions

**Vấn đề:** Sessions/windows được load một lần khi mở. Session tạo bởi script bên ngoài không hiện lên.

**Ý tưởng:** Background tick mỗi 2–5 giây gọi `tmux list-sessions`, so sánh diff với state hiện tại, tự động thêm/xóa cards mà không reset focus. Có thể bật/tắt qua config.

**Priority:** Trung bình | **Độ khó:** Trung bình

---

## 8. Pane layout diagram — sơ đồ ASCII trong preview panel

**Ý tưởng:** Ở Pane view, thay vì chỉ hiện text metadata trong preview panel, vẽ sơ đồ ASCII layout thực tế của các panes trong window. Pane đang focus được highlight.

```
┌───────────┬──────────┐
│  pane 0   │  pane 1  │
│  (vim)    │  (zsh) ◀ │
├───────────┴──────────┤
│       pane 2         │
│       (htop)         │
└──────────────────────┘
```

Tương tự `tmux display-panes` nhưng tích hợp thẳng vào preview.

**Priority:** Thấp | **Độ khó:** Cao

---

## Bảng tóm tắt

| # | Tính năng | Độ khó | Priority |
|---|-----------|--------|----------|
| 1 | Global fuzzy search | Trung bình | Cao |
| 2 | Tag filtering UI | Thấp | Cao |
| 5 | Recent sessions sort | Rất thấp | Trung bình |
| 7 | Live refresh | Trung bình | Trung bình |
| 3 | Session grouping theo tag | Cao | Trung bình |
| 4 | Command palette | Cao | Trung bình |
| 6 | Per-session colors | Thấp | Thấp |
| 8 | Pane layout diagram | Cao | Thấp |
