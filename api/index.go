package handler

import (
	"context"
	"encoding/json"
	//"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------
// STRUKTUR DATA (MODEL)
// ---------------------------------------------------------

// Buat nampung data Register atau Login
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Buat nampung data Task (tugas) yang lama
type Task struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
}

// Data sementara untuk task (punya kamu yang lama)
var tasks = []Task{}

// ---------------------------------------------------------
// IKLAN / UTAMA (HANDLER VERCEL)
// ---------------------------------------------------------
func Handler(w http.ResponseWriter, r *http.Request) {
	// 1. SET HEADER CORS SUPAYA TIDAK DIBLOKIR FRONTEND
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// 2. TANGANI PREFLIGHT OPTIONS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	// =========================================================
	// BAGIAN A: FITUR REGISTER & LOGIN (PAKAI DATABASE)
	// =========================================================

	// Kalau akses /api/register
	if strings.HasSuffix(path, "/register") && r.Method == http.MethodPost {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var body AuthPayload
		json.NewDecoder(r.Body).Decode(&body)

		// Encrip / Acak password biar aman
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error": "Gagal mengamankan password"}`, http.StatusInternalServerError)
			return
		}

		// Masukin ke database Neon/Vercel
		_, err = conn.Exec(ctx, "INSERT INTO users (email, password) VALUES ($1, $2)", body.Email, string(hashedPassword))
		if err != nil {
			http.Error(w, `{"error": "Email sudah terdaftar atau tidak valid"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Yeay, Berhasil Daftar!"})
		return
	}

	// Kalau akses /api/login
	if strings.HasSuffix(path, "/login") && r.Method == http.MethodPost {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var body AuthPayload
		json.NewDecoder(r.Body).Decode(&body)

		var passwordDariDatabase string
		err = conn.QueryRow(ctx, "SELECT password FROM users WHERE email = $1", body.Email).Scan(&passwordDariDatabase)
		if err != nil {
			http.Error(w, `{"error": "Login Gagal: Email tidak ditemukan"}`, http.StatusUnauthorized)
			return
		}

		// Bandingkan password ketikan user sama database
		err = bcrypt.CompareHashAndPassword([]byte(passwordDariDatabase), []byte(body.Password))
		if err != nil {
			http.Error(w, `{"error": "Login Gagal: Password salah"}`, http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login Berhasil! Selamat datang."})
		return
	}

	// =========================================================
	// BAGIAN B: FITUR TASKS (KODE KAMU YANG LAMA)
	// =========================================================

	// Helper buat ambil ID dari URL (contoh: /api/tasks/1)
	getTaskID := func() (int, error) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(pathParts) >= 3 {
				idStr = pathParts[len(pathParts)-1]
			}
		}
		return strconv.Atoi(idStr)
	}

	// GET /api/tasks (Ambil semua data task)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(tasks)
		return
	}

	// POST /api/tasks (Tambah task baru)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodPost {
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		task.ID = len(tasks) + 1
		tasks = append(tasks, task)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
		return
	}

	// PUT /api/tasks/1 (Ubah task)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodPut {
		id, err := getTaskID()
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var updatedTask Task
		json.NewDecoder(r.Body).Decode(&updatedTask)

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Task = updatedTask.Task
				json.NewEncoder(w).Encode(tasks[i])
				return
			}
		}	
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// DELETE /api/tasks/1 (Hapus task)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodDelete {
		id, err := getTaskID()
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
}

// ---------------------------------------------------------
// FUNGSI BANTUAN KONEKSI DATABASE
// ---------------------------------------------------------
func connectDB() (*pgx.Conn, context.Context, error) {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, ctx, err
	}
	return conn, ctx, nil
}
