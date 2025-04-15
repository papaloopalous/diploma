let globalStudentID = "";
const MAX_FILE_SIZE = 10 * 1024 * 1024;

function fillFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const sid = params.get("studentID");
  if (sid) {
    globalStudentID = sid.trim();
  }
}

function showStatus(message, isError = false) {
  const box = document.getElementById('statusBox');
  box.textContent = message;
  box.className = 'alert ' + (isError ? 'alert-error' : 'alert-info');
  box.style.display = 'block';
  setTimeout(() => {
    box.style.display = 'none';
  }, 4000);
}

async function uploadFile() {
  const taskName = document.getElementById("taskName").value.trim();
  const fileInput = document.getElementById("fileInput");
  if (!globalStudentID || !taskName || fileInput.files.length === 0) {
    showStatus("Не указаны все данные: studentID, имя задания или файл", true);
    return;
  }

  const file = fileInput.files[0];
  if (file.size > MAX_FILE_SIZE) {
    showStatus("Размер файла больше 10 МБ", true);
    return;
  }

  let arrayBuffer;
  try {
    arrayBuffer = await file.arrayBuffer();
  } catch (err) {
    showStatus("Ошибка чтения файла: " + err.message, true);
    return;
  }

  const uint8 = new Uint8Array(arrayBuffer);
  let compressed;
  try {
    compressed = fflate.gzipSync(uint8);
  } catch (err) {
    showStatus("Ошибка сжатия файла: " + err.message, true);
    return;
  }

  try {
    const response = await fetch("/api/upload-task", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        "studentID": globalStudentID,
        "taskName": taskName,
        "fileName": file.name
      },
      body: compressed
    });
    if (!response.ok) {
      const msg = await response.text();
      showStatus("Ошибка сервера: " + msg, true);
      return;
    }
    showStatus("Файл успешно прикреплён");
    setTimeout(() => window.location.href = "/main", 1000);
  } catch (err) {
    showStatus("Ошибка при отправке файла: " + err.message, true);
  }
}
