package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Task struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
}

var tasks = []Task{
	
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// 1. SET HEADER CORS DI PALING ATAS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// 2. TANGANI PREFLIGHT OPTIONS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Helper mengekstrak ID dari URL path (/api/tasks/1) atau query (?id=1)
	getTaskID := func() (int, error) {
		idStr := r.URL.Query().Get("id")

		if idStr == "" {
			// Mengambil segment terakhir dari URL path
			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(pathParts) >= 3 {
				idStr = pathParts[len(pathParts)-1]
			}
		}
		return strconv.Atoi(idStr)
	}

	// GET /api/tasks
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(tasks)
		return
	}

	// POST /api/tasks
	if r.Method == http.MethodPost {
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

	// PUT /api/tasks/1 atau /api/tasks?id=1
	if r.Method == http.MethodPut {
		id, err := getTaskID()
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var task Task
		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Task = task.Task
				json.NewEncoder(w).Encode(tasks[i])
				return
			}
		}

		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// DELETE /api/tasks/1 atau /api/tasks?id=1
	if r.Method == http.MethodDelete {
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

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	// PENTING: Tambahkan slash di akhir ("/api/tasks/") agar menangkap /api/tasks/1, /api/tasks/2, dst.
	http.HandleFunc("/api/tasks/", tasksHandler)
	http.HandleFunc("/api/tasks", tasksHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}