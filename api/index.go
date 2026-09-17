package handler

import (
	"context"
	"encoding/json"
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

type AuthPayload struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Task struct {
	ID     int    `json:"id"`
	IDUser int    `json:"id_user"`
	Task   string `json:"task"`
}

type User struct {
	ID int `json:"id"`
	Name string	`json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

// ---------------------------------------------------------
// HANDLER UTAMA VERCEL
// ---------------------------------------------------------
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	// =========================================================
	// BAGIAN A: REGISTER & LOGIN
	// =========================================================

	if strings.HasSuffix(path, "/register") && r.Method == http.MethodPost {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var body AuthPayload
		json.NewDecoder(r.Body).Decode(&body)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error": "Gagal mengamankan password"}`, http.StatusInternalServerError)
			return
		}

		_, err = conn.Exec(ctx, "INSERT INTO users (email, password) VALUES ($1, $2)", body.Email, string(hashedPassword))
		if err != nil {
			http.Error(w, `{"error": "Email sudah terdaftar atau tidak valid"}`, http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Yeay, Berhasil Daftar!"})
		return
	}

	if strings.HasSuffix(path, "/login") && r.Method == http.MethodPost {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var body AuthPayload
		json.NewDecoder(r.Body).Decode(&body)

		var userID int
		var passwordDariDatabase string
		err = conn.QueryRow(ctx, "SELECT id, password FROM users WHERE email = $1", body.Email).Scan(&userID, &passwordDariDatabase)
		if err != nil {
			http.Error(w, `{"error": "Login Gagal: Email tidak ditemukan"}`, http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(passwordDariDatabase), []byte(body.Password))
		if err != nil {
			http.Error(w, `{"error": "Login Gagal: Password salah"}`, http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Login Berhasil!",
			"id":      userID,
		})
		return
	}

	if strings.HasSuffix(path, "/check-session") && r.Method == http.MethodGet {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		userID := r.URL.Query().Get("id")
		var exists bool
		err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, `{"error": "Sesi tidak valid"}`, http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"valid": true})
		return
	}

	// =========================================================
	// BAGIAN B: TASKS (TERHUBUNG KE DATABASE)
	// =========================================================

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

	// GET /api/tasks (Ambil semua task berdasarkan id_user, contoh: /api/tasks?id_user=1)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodGet {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		userID := r.URL.Query().Get("id_user")
		var rows pgx.Rows

		if userID != "" {
			rows, err = conn.Query(ctx, "SELECT id, id_user, task FROM tasks WHERE id_user = $1 ORDER BY id DESC", userID)
		} else {
			rows, err = conn.Query(ctx, "SELECT id, id_user, task FROM tasks ORDER BY id DESC")
		}

		if err != nil {
			http.Error(w, `{"error": "Gagal mengambil data task"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var t Task
			rows.Scan(&t.ID, &t.IDUser, &t.Task)
			tasks = append(tasks, t)
		}

		// Jika kosong, pastikan mengembalikan array kosong `[]` bukan null
		if tasks == nil {
			tasks = []Task{}
		}

		json.NewEncoder(w).Encode(tasks)
		return
	}

	// POST /api/tasks (Tambah task baru ke database)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodPost {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var task Task
		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil || task.Task == "" || task.IDUser == 0 {
			http.Error(w, `{"error": "Data JSON atau id_user/task tidak valid"}`, http.StatusBadRequest)
			return
		}

		err = conn.QueryRow(ctx, "INSERT INTO tasks (id_user, task) VALUES ($1, $2) RETURNING id", task.IDUser, task.Task).Scan(&task.ID)
		if err != nil {
			http.Error(w, `{"error": "Gagal menyimpan task ke database"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
		return
	}

	// PUT /api/tasks/1 (Ubah task di database)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodPut {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		id, err := getTaskID()
		if err != nil {
			http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
			return
		}

		var updatedTask Task
		json.NewDecoder(r.Body).Decode(&updatedTask)

		result, err := conn.Exec(ctx, "UPDATE tasks SET task = $1 WHERE id = $2", updatedTask.Task, id)
		if err != nil || result.RowsAffected() == 0 {
			http.Error(w, `{"error": "Task tidak ditemukan atau gagal diubah"}`, http.StatusNotFound)
			return
		}

		updatedTask.ID = id
		json.NewEncoder(w).Encode(updatedTask)
		return
	}

	// DELETE /api/tasks/1 (Hapus task dari database)
	if strings.Contains(path, "/tasks") && r.Method == http.MethodDelete {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		id, err := getTaskID()
		if err != nil {
			http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
			return
		}

		result, err := conn.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
		if err != nil || result.RowsAffected() == 0 {
			http.Error(w, `{"error": "Task tidak ditemukan"}`, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET /api/users (Ambil semua users)
	if strings.Contains(path, "/users") && r.Method == http.MethodGet {
		conn, ctx, err := connectDB()
		if err != nil {
			http.Error(w, `{"error": "Gagal konek database"}`, http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		var rows pgx.Rows
		rows, err = conn.Query(ctx, "SELECT * FROM userss ORDER BY id DESC")

		if err != nil {
			http.Error(w, `{"error": "Gagal mengambil data task"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []User
		for rows.Next() {
			var t User
			rows.Scan(&t.ID, &t.Name, &t.Email)
			tasks = append(tasks, t)
		}

		// Jika kosong, pastikan mengembalikan array kosong `[]` bukan null
		if tasks == nil {
			tasks = []User{}
		}

		json.NewEncoder(w).Encode(tasks)
		return
	}

	http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
}

// ---------------------------------------------------------
// FUNGSI KONEKSI DATABASE
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
