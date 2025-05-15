const logoutBtn = document.getElementById('logoutBtn');
const MAX_FILE_SIZE = 10 * 1024 * 1024;
let roleCookie = null;

logoutBtn.addEventListener('click', () => {
  fetch('/api/logout', { method: 'DELETE', credentials: 'include' })
    .catch(err => console.warn(err))
    .finally(() => { window.location.href = '/' });
});

function getCookie(name) {
  const matches = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + '=([^;]*)'));
  return matches ? decodeURIComponent(matches[1]) : undefined;
}

roleCookie = getCookie('userRole');
if (roleCookie === 'teacher') {
  document.getElementById('menuTeacher').classList.add('active');
} else if (roleCookie === 'student') {
  document.getElementById('menuStudent').classList.add('active');
}

function showSection(id) {
    document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
    const target = document.getElementById(id);
    if (target) target.classList.add('active');
  
    if (id === 'studentsSection') loadStudents();
    else if (id === 'tasksSection') loadTasks();
    else if (id === 'requestsSection') loadRequests();
    else if (id === 'profileSection') loadProfile();
    else if (id === 'allTeachersSection') loadAllTeachers();
    else if (id === 'myTeachersSection') loadMyTeachers();
    else if (id === 'studentRequestsSection') loadStudentRequests();
  }
  

async function downloadFile(taskID, type = 'task') {
  const endpoint = (type === 'solution') ? '/api/download-solution' : '/api/download-task';
  const url = `${endpoint}?taskID=${encodeURIComponent(taskID)}`;
  try {
    const res = await fetch(url);
    if (!res.ok) {
      const msg = await res.text();
      alert("Ошибка при загрузке:\n" + msg);
      return;
    }
    const fileName = getFileNameFromContentDisposition(res.headers.get("Content-Disposition")) || ((type === 'solution') ? "solution.docx" : "task.docx");
    const compressedBuffer = await res.arrayBuffer();
    const compressed = new Uint8Array(compressedBuffer);
    let uncompressed;
    try {
      uncompressed = fflate.gunzipSync(compressed);
    } catch (err) {
      alert("Ошибка распаковки файла: " + err.message);
      return;
    }
    const blob = new Blob([uncompressed]);
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = fileName;
    a.click();
  } catch (err) {
    alert("Ошибка при загрузке файла: " + err.message);
  }
}

function getFileNameFromContentDisposition(header) {
  if (!header) return null;
  const match = header.match(/filename="?([^"]+)"?/);
  return match ? match[1] : null;
}

function handleSolutionUpload(taskID) {
  const fileInput = document.createElement('input');
  fileInput.type = 'file';
  fileInput.accept = '.docx';
  fileInput.onchange = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    if (file.size > MAX_FILE_SIZE) {
      alert("Размер файла превышает 10 МБ!");
      return;
    }
    let arrayBuffer;
    try {
      arrayBuffer = await file.arrayBuffer();
    } catch (err) {
      alert("Ошибка чтения файла: " + err.message);
      return;
    }
    const uint8 = new Uint8Array(arrayBuffer);
    let compressed;
    try {
      compressed = fflate.gzipSync(uint8);
    } catch (err) {
      alert("Ошибка сжатия файла: " + err.message);
      return;
    }
    try {
      const response = await fetch("/api/upload-solution", {
        method: "POST",
        headers: {
          "Content-Type": "application/octet-stream",
          "taskID": taskID,
          "fileName": file.name
        },
        body: compressed
      });
      const text = await response.text();
      alert("Файл отправлен");
      location.reload();
    } catch (err) {
      alert("Ошибка при отправке файла: " + err.message);
    }
  };
  fileInput.click();
}

async function addGrade(taskID) {
  const grade = prompt("Введите оценку:", "5");
  if (!grade) return;
  try {
    const url = `/api/add-grade?taskID=${encodeURIComponent(taskID)}&grade=${encodeURIComponent(grade)}`;
    const response = await fetch(url, { method: "POST", credentials: 'include' });
    if (!response.ok) {
      const msg = await response.text();
      alert("Ошибка при добавлении оценки:\n" + msg);
      return;
    }
    alert("Оценка успешно добавлена");
    loadTasks();
  } catch (err) {
    alert("Ошибка при добавлении оценки: " + err.message);
  }
}

async function loadStudents() {
  try {
    const res = await fetch('/api/get-students', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let users = responseData.data;
    if (!Array.isArray(users)) users = [];
    renderStudents(Array.isArray(users) ? users : []);
  } catch (err) {
    alert(err.message);
  }
}

async function startChat(otherUserId) {
  try {
    const res = await fetch('/api/create-chat-room', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ otherUserId })
    });

    if (!res.ok) {
      const text = await res.text();
      alert('Ошибка при создании чата: ' + text);
      return;
    }

    const data = await res.json();
    if (!data.success) {
      alert('Не удалось создать чат: ' + data.message);
      return;
    }

    window.location.href = `/chat?room=${data.data.roomId}`;
  } catch (err) {
    alert('Ошибка при создании чата: ' + err.message);
  }
}

function renderStudents(users) {
  const container = document.getElementById('studentsList');
  container.innerHTML = '';
  if (!users.length) {
    container.textContent = 'Нет студентов';
    return;
  }
  users.forEach(u => {
    const div = document.createElement('div');
    div.textContent = `ФИО: ${u.fio}, Возраст: ${u.age}`;
    const btnTask = document.createElement('button');
    btnTask.textContent = 'Добавить задание';
    btnTask.onclick = () => goToTaskPage(u);
    const btnChat = document.createElement('button');
    btnChat.textContent = 'Начать чат';
    btnChat.onclick = () => startChat(u.id);
    div.appendChild(document.createElement('br'));
    div.appendChild(btnTask);
    div.appendChild(btnChat);
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

function goToTaskPage(student) {
  const studentID = encodeURIComponent(student.id);
  window.location.href = `/task?studentID=${studentID}`;
}

async function loadRequests() {
  try {
    const res = await fetch('/api/get-teacher-requests', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let reqs = responseData.data;
    if (!Array.isArray(reqs)) reqs = [];
    renderRequests(Array.isArray(reqs) ? reqs : []);
  } catch (err) {
    alert(err.message);
  }
}

function renderRequests(reqs) {
  const container = document.getElementById('requestsList');
  container.innerHTML = '';
  if (!reqs.length) {
    container.textContent = 'Нет заявок';
    return;
  }
  reqs.forEach(r => {
    const div = document.createElement('div');
    div.textContent = `ФИО: ${r.fio}, Возраст: ${r.age}`;
    const btnConfirm = document.createElement('button');
    btnConfirm.textContent = 'Принять';
    btnConfirm.onclick = () => confirmRequest(r, div);
    const btnDeny = document.createElement('button');
    btnDeny.textContent = 'Отклонить';
    btnDeny.onclick = () => denyRequest(r, div);
    div.appendChild(document.createElement('br'));
    div.appendChild(btnConfirm);
    div.appendChild(btnDeny);
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

async function confirmRequest(r, divEl) {
  try {
    const url = '/api/confirm?studentID=' + encodeURIComponent(r.id);
    const res = await fetch(url, { method: 'POST', credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    alert('Заявка принята');
    location.reload();
  } catch (err) {
    alert(err.message);
  }
}

async function denyRequest(r, divEl) {
  try {
    const url = '/api/deny?studentID=' + encodeURIComponent(r.id);
    const res = await fetch(url, { method: 'POST', credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    alert('Заявка отклонена');
    location.reload();
  } catch (err) {
    alert(err.message);
  }
}

async function loadTasks() {
  try {
    const res = await fetch('/api/get-tasks', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let tasks = responseData.data;
    if (!Array.isArray(tasks)) tasks = [];
    renderTasks(Array.isArray(tasks) ? tasks : []);
  } catch (err) {
    alert(err.message);
  }
}

function renderTasks(tasks) {
  const container = document.getElementById('tasksList');
  container.innerHTML = '';
  if (!tasks.length) {
    container.textContent = 'Заданий нет';
    return;
  }
  tasks.forEach(t => {
    let info = `Название: ${t.taskName}, Статус: ${t.status}`;
    if (typeof t.grade !== 'undefined') info += `, Оценка: ${t.grade}`;
    if (roleCookie === 'student') {
      info += `, Преподаватель: ${t.teacher || ''}`;
    } else if (roleCookie === 'teacher') {
      info += `, Студент: ${t.student || ''}`;
    }
    const div = document.createElement('div');
    div.textContent = info;
    div.appendChild(document.createElement('br'));
    if (roleCookie === 'student') {
      const btnDownloadTask = document.createElement('button');
      btnDownloadTask.textContent = 'Скачать задание';
      btnDownloadTask.onclick = () => downloadFile(t.taskID, 'task');
      div.appendChild(btnDownloadTask);
      const btnUploadSolution = document.createElement('button');
      btnUploadSolution.textContent = 'Загрузить решение';
      btnUploadSolution.onclick = () => handleSolutionUpload(t.taskID);
      div.appendChild(btnUploadSolution);
    } else if (roleCookie === 'teacher') {
      if (t.status === 'ready to grade') {
        const btnDownloadSol = document.createElement('button');
        btnDownloadSol.textContent = 'Скачать решение';
        btnDownloadSol.onclick = () => downloadFile(t.taskID, 'solution');
        div.appendChild(btnDownloadSol);
        const btnAddGrade = document.createElement('button');
        btnAddGrade.textContent = 'Добавить оценку';
        btnAddGrade.onclick = () => addGrade(t.taskID);
        div.appendChild(btnAddGrade);
      }
    }
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

async function loadProfile() {
  try {
    const res = await fetch('/api/get-profile', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let user = responseData.data;
    renderProfile(user);
  } catch (err) {
    alert(err.message);
  }
}

function renderProfile(user) {
  const container = document.getElementById('profileInfo');
  container.innerHTML = '';
  if (!user || typeof user.fio !== 'string' || user.fio.trim() === '') {
    container.textContent = 'Профиль не заполнен';
    return;
  }
  let html = `<p><strong>ФИО:</strong> ${user.fio}</p><p><strong>Возраст:</strong> ${user.age}</p>`;
  if (user.specialty) html += `<p><strong>Специализация:</strong> ${user.specialty}</p>`;
  if (typeof user.price === 'number') html += `<p><strong>Цена:</strong> ${user.price}</p>`;
  if (typeof user.rating === 'number') html += `<p><strong>Рейтинг:</strong> ${user.rating}</p>`;
  container.innerHTML = html;
}

async function loadAllTeachers() {
  const sortField = document.getElementById('sortField').value;
  const sortOrder = document.getElementById('sortOrder').value;
  try {
    const res = await fetch('/api/get-teachers', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let teachers = responseData.data;
    if (!Array.isArray(teachers)) teachers = [];
    teachers.sort((a, b) => {
      if (sortField === 'price' || sortField === 'rating') {
        if (sortOrder === 'asc') return a[sortField] - b[sortField];
        return b[sortField] - a[sortField];
      }
      return 0;
    });
    renderAllTeachers(teachers);
  } catch (err) {
    alert(err.message);
  }
}

function renderAllTeachers(teachers) {
  const container = document.getElementById('allTeachersList');
  container.innerHTML = '';
  if (!teachers.length) {
    container.textContent = 'Нет преподавателей';
    return;
  }
  teachers.forEach(t => {
    const div = document.createElement('div');
    div.textContent = `ФИО: ${t.fio}, Возраст: ${t.age}, Специальность: ${t.specialty || '-'}, Цена: ${t.price}, Рейтинг: ${t.rating}`;
    const btnReq = document.createElement('button');
    btnReq.textContent = 'Отправить заявку';
    btnReq.onclick = () => sendRequestToTeacher(t, btnReq);
    div.appendChild(document.createElement('br'));
    div.appendChild(btnReq);
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

async function sendRequestToTeacher(teacher, btnElement) {
  const url = '/api/send-request?teacherID=' + encodeURIComponent(teacher.id);
  try {
    const res = await fetch(url, { method: 'POST', credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    const reply = await res.json();
    alert('Заявка отправлена');
    location.reload();
  } catch (err) {
    alert(err.message);
  }
}

async function loadMyTeachers() {
  try {
    const res = await fetch('/api/get-my-teachers', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let teachers = responseData.data;
    if (!Array.isArray(teachers)) teachers = [];
    renderMyTeachers(teachers);
  } catch (err) {
    alert(err.message);
  }
}

function renderMyTeachers(teachers) {
  const container = document.getElementById('myTeachersList');
  container.innerHTML = '';
  if (!teachers.length) {
    container.textContent = 'Преподавателей нет';
    return;
  }
  teachers.forEach(t => {
    const div = document.createElement('div');
    div.textContent = `ФИО: ${t.fio}, Возраст: ${t.age}, Специальность: ${t.specialty || '-'}, Цена: ${t.price}, Рейтинг: ${t.rating}`;
    const btnChat = document.createElement('button');
    btnChat.textContent = 'Начать чат';
    btnChat.onclick = () => startChat(t.id);
    div.appendChild(document.createElement('br'));
    div.appendChild(btnChat);
    const btnAddRating = document.createElement('button');
    btnAddRating.textContent = 'Добавить рейтинг';
    btnAddRating.onclick = () => addRatingToTeacher(t.id);
    div.appendChild(btnAddRating);
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

async function addRatingToTeacher(teacherID) {
  const rating = prompt("Введите рейтинг (1-5):", "5");
  if (!rating) return;
  try {
    const url = `/api/add-rating?teacherID=${encodeURIComponent(teacherID)}&rating=${encodeURIComponent(rating)}`;
    const res = await fetch(url, { method: 'POST', credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert("Ошибка при добавлении рейтинга:\n" + text);
      return;
    }
    alert('Рейтинг успешно добавлен');
    location.reload();
  } catch (err) {
    alert("Ошибка при добавлении рейтинга: " + err.message);
  }
}

async function loadStudentRequests() {
  try {
    const res = await fetch('/api/get-student-requests', { credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    let responseData = await res.json();
    let reqs = responseData.data;
    if (!Array.isArray(reqs)) reqs = [];
    renderStudentRequests(reqs);
  } catch (err) {
    alert(err.message);
  }
}

function renderStudentRequests(reqs) {
  const container = document.getElementById('studentRequestsList');
  container.innerHTML = '';
  if (!reqs.length) {
    container.textContent = 'Нет заявок';
    return;
  }
  reqs.forEach(r => {
    const div = document.createElement('div');
    div.textContent = `ФИО: ${r.fio}, Возраст: ${r.age}, Специальность: ${r.specialty}, Цена: ${r.price}, Рейтинг: ${r.rating}`;
    const btnCancel = document.createElement('button');
    btnCancel.textContent = 'Отменить заявку';
    btnCancel.onclick = () => cancelStudentRequest(r, div);
    div.appendChild(document.createElement('br'));
    div.appendChild(btnCancel);
    container.appendChild(div);
    container.appendChild(document.createElement('hr'));
  });
}

async function cancelStudentRequest(r, divEl) {
  try {
    const url = '/api/cancel-request?teacherID=' + encodeURIComponent(r.id);
    const res = await fetch(url, { method: 'POST', credentials: 'include' });
    if (!res.ok) {
      const text = await res.text();
      alert(text);
      return;
    }
    alert('Заявка отменена');
    location.reload();
  } catch (err) {
    alert(err.message);
  }
}