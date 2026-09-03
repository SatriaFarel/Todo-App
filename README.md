# 📝 Todo App (Go & JavaScript)

Aplikasi manajemen tugas harian (*Todo List*) yang dibangun menggunakan **Golang** sebagai RESTful API Backend dan **JavaScript (Vanilla)** + **Tailwind CSS** untuk Frontend.

---

## 🚀 Fitur Utama

- **CRUD Lengkap (Create, Read, Update, Delete):**
  - **Tambah Task:** Menambahkan tugas harian baru.
  - **Checklist Selesai:** Menandai tugas yang sudah selesai diselesaikan.
  - **Edit Task:** Mengubah deskripsi tugas yang sudah ada secara instan.
  - **Hapus Task:** Menghapus tugas yang tidak lagi diperlukan.
- **UI Modern & Responsive:** Menggunakan Tailwind CSS v4 dengan pendekatan *Clean Dark Mode*.
- **Widget Tanggal & Jam Realtime:** Menampilkan waktu lokal terkini yang diperbarui setiap detik.
- **Validasi & Proteksi Dasar:** Mencegah *input* kosong serta proteksi dasar dari celah XSS.

---

## 🛠️ Tech Stack

- **Backend:** Golang (`net/http` standard library)
- **Frontend:** HTML5, Vanilla JavaScript (ES6+), Tailwind CSS v4 (CDN)
- **Format Data:** REST API / JSON

---

## 📁 Struktur Folder Project

```text
todo-app/
├── api/
│   └── main.go         # Server HTTP & REST API Golang
├── web/
│   ├── js/
│   │   └── script.js   # Logic Fetch API & DOM Manipulation
│   └── index.html      # Interface Frontend (HTML & Tailwind CSS)
├── go.mod              # Inisialisasi Modul Go
└── README.md

