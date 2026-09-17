// Memperbarui jam dan tanggal di header
function updateDateTime() {
    const now = new Date();
    const dateOptions = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
    
    document.getElementById('current-date').textContent = now.toLocaleDateString('id-ID', dateOptions);
    document.getElementById('current-time').textContent = now.toLocaleTimeString('id-ID');
}

setInterval(updateDateTime, 1000);
updateDateTime();

// Sesuaikan URL endpoint mengarah ke /api/tasks
const API_URL = 'https://todo-app.kishiyuusha.my.id/api/tasks';

// Ambil id_user yang disimpan di localStorage saat login berhasil
const currentUserId = localStorage.getItem('id_user');

// Fungsi Logout
function logout() {
    localStorage.removeItem('id_user');
    window.location.href = "index.html";
}

// Cek Sesi (Apakah ID masih valid di database)
async function checkSession() {
    if (!currentUserId) {
        window.location.href = "login.html";
        return;
    }

    try {
        const response = await fetch(`/api/check-session?id=${currentUserId}`);
        if (!response.ok) {
            throw new Error("Session invalid");
        }
    } catch (error) {
        console.error("Session check failed:", error);
        localStorage.removeItem('id_user');
        window.location.href = "login.html";
    }
}

// Jalankan cek sesi segera
checkSession();

// Pasang event listener logout jika tombol ada
document.getElementById('logout-btn')?.addEventListener('click', logout);

// Mengambil seluruh data tugas dari backend berdasarkan id_user
async function fetchTasks() {
    if (!currentUserId) return;

    const taskList = document.getElementById('task-list');
    const taskCount = document.getElementById('task-count');

    try {
        // Kirim id_user sebagai parameter query
        const response = await fetch(`${API_URL}?id_user=${currentUserId}`);
        if (!response.ok) throw new Error(`Status: ${response.status}`);

        const tasks = await response.json();
        taskList.innerHTML = '';

        if (!tasks || tasks.length === 0) {
            taskList.innerHTML = `
                <div class="text-center py-8 border border-dashed border-slate-700 rounded-lg">
                    <p class="text-xs text-slate-500">Belum ada tugas.</p>
                </div>`;
            taskCount.textContent = '0 Tasks';
            return;
        }

        taskCount.textContent = `${tasks.length} Task${tasks.length > 1 ? 's' : ''}`;

        // Render setiap baris data
        tasks.forEach(task => {
            const taskItem = document.createElement('div');
            taskItem.className = `flex items-center justify-between p-3 rounded-lg border transition bg-slate-800 border-slate-700/80`;

            // Sisi Kiri: Teks Task
            const leftSection = document.createElement('div');
            leftSection.className = 'flex items-center gap-3';

            const textSpan = document.createElement('span');
            textSpan.className = `text-xs font-medium text-slate-200`;
            textSpan.textContent = task.task;

            leftSection.appendChild(textSpan);

            // Sisi Kanan: Tombol Aksi
            const actionSection = document.createElement('div');
            actionSection.className = 'flex items-center gap-1.5';

            const editBtn = document.createElement('button');
            editBtn.className = 'px-2 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 text-[11px] rounded transition';
            editBtn.textContent = 'Edit';
            editBtn.addEventListener('click', () => editTask(task.id, task.task));

            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'px-2 py-1 bg-rose-500/10 hover:bg-rose-500 text-rose-400 hover:text-white text-[11px] rounded transition';
            deleteBtn.textContent = 'Hapus';
            deleteBtn.addEventListener('click', () => deleteTask(task.id));

            actionSection.appendChild(editBtn);
            actionSection.appendChild(deleteBtn);

            taskItem.appendChild(leftSection);
            taskItem.appendChild(actionSection);

            taskList.appendChild(taskItem);
        });

    } catch (error) {
        console.error('Gagal mengambil data task:', error);
        taskList.innerHTML = `
            <div class="p-3 bg-rose-500/10 border border-rose-500/20 rounded-lg text-center">
                <p class="text-xs text-rose-400">Gagal terhubung ke backend server (Go).</p>
            </div>`;
    }
}

// Menambah tugas baru (menyertakan id_user)
async function addTask() {
    const input = document.getElementById('new-task');
    const taskText = input.value.trim();

    if (!taskText) return;

    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                id_user: parseInt(currentUserId), 
                task: taskText 
            })
        });

        if (response.ok) {
            input.value = '';
            fetchTasks();
        }
    } catch (error) {
        console.error('Gagal menambah task:', error);
    }
}

// Mengedit deskripsi tugas
async function editTask(taskId, oldText) {
    const newText = prompt('Ubah deskripsi tugas:', oldText);

    if (newText === null || newText.trim() === '' || newText.trim() === oldText) return;

    try {
        const response = await fetch(`${API_URL}/${taskId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                id_user: parseInt(currentUserId),
                task: newText.trim()
            })
        });

        if (response.ok) fetchTasks();
    } catch (error) {
        console.error('Gagal mengedit task:', error);
    }
}

// Menghapus tugas
async function deleteTask(taskId) {
    try {
        const confirmDelete = confirm('Apakah Anda yakin ingin menghapus tugas ini?');
        if (!confirmDelete) return;

        const response = await fetch(`${API_URL}/${taskId}`, {
            method: 'DELETE'
        });

        if (response.ok) fetchTasks();
    } catch (error) {
        console.error('Gagal menghapus task:', error);
    }
}

// Event listener saat tombol diklik atau menekan Enter
document.getElementById('add-task').addEventListener('click', addTask);
document.getElementById('new-task').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') addTask();
});

document.addEventListener('DOMContentLoaded', fetchTasks);