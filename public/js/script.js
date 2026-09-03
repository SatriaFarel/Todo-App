// Memperbarui jam dan tanggal di header
        function updateDateTime() {
            const now = new Date();
            const dateOptions = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
            
            document.getElementById('current-date').textContent = now.toLocaleDateString('id-ID', dateOptions);
            document.getElementById('current-time').textContent = now.toLocaleTimeString('id-ID');
        }

        setInterval(updateDateTime, 1000);
        updateDateTime();

        const API_URL = 'http://todo-app-sf/api/tasks';

        // Mengambil seluruh data tugas dari backend
        async function fetchTasks() {
            const taskList = document.getElementById('task-list');
            const taskCount = document.getElementById('task-count');

            try {
                const response = await fetch(API_URL);
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
                    taskItem.className = `flex items-center justify-between p-3 rounded-lg border transition ${
                        task.is_completed 
                            ? 'bg-slate-900/40 border-slate-800 opacity-60' 
                            : 'bg-slate-800 border-slate-700/80'
                    }`;

                    // Sisi Kiri: Checkbox & Teks
                    const leftSection = document.createElement('div');
                    leftSection.className = 'flex items-center gap-3';

                    const textSpan = document.createElement('span');
                    textSpan.className = `text-xs font-medium ${task.is_completed ? 'line-through text-slate-500' : 'text-slate-200'}`;
                    textSpan.textContent = task.task;

                    leftSection.appendChild(textSpan);

                    // Sisi Kanan: Tombol Aksi
                    const actionSection = document.createElement('div');
                    actionSection.className = 'flex items-center gap-1.5';

                    const editBtn = document.createElement('button');
                    editBtn.className = 'px-2 py-1 bg-slate-700 hover:bg-slate-600 text-slate-200 text-[11px] rounded transition';
                    editBtn.textContent = 'Edit';
                    editBtn.addEventListener('click', () => editTask(task.id, task.task, task.is_completed));

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

        // Menambah tugas baru
        async function addTask() {
            const input = document.getElementById('new-task');
            const taskText = input.value.trim();

            if (!taskText) return;

            try {
                const response = await fetch(API_URL, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ task: taskText, is_completed: false })
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
        async function editTask(taskId, oldText, isCompleted) {
            const newText = prompt('Ubah deskripsi tugas:', oldText);

            if (newText === null || newText.trim() === '' || newText.trim() === oldText) return;

            try {
                const response = await fetch(`${API_URL}/${taskId}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        task: newText.trim(),
                        is_completed: isCompleted
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