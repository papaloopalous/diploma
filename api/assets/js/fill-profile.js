function normalizeResponse(resp) {
    return {
      success: resp.success,
      statusCode: resp.code,
      message: resp.message,
      data: resp.data
    };
  }
  function getCookie(name) {
    const matches = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + '=([^;]*)'));
    return matches ? decodeURIComponent(matches[1]) : undefined;
  }
  const role = getCookie('userRole');
  if (role === 'teacher') {
    document.getElementById('teacherFields').style.display = 'block';
  }
  document.getElementById('profileForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const alertError = document.getElementById('alertError');
    const alertSuccess = document.getElementById('alertSuccess');
    alertError.style.display = 'none';
    alertSuccess.style.display = 'none';
    const lastName = document.getElementById('lastName').value.trim();
    const firstName = document.getElementById('firstName').value.trim();
    const middleName = document.getElementById('middleName').value.trim();
    const age = document.getElementById('age').value.trim();
    let fio = `${lastName} ${firstName}`;
    if (middleName) fio += ` ${middleName}`;
    const payload = { fio: fio, age: Number(age) };
    if (role === 'teacher') {
      payload.price = Number(document.getElementById('price').value.trim());
      payload.specialty = document.getElementById('specialty').value.trim();
    }
    try {
      const resp = await fetch('/api/fill-profile', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(payload)
      });
      const raw = await resp.json();
      const result = normalizeResponse(raw);
      if (result.success) {
        alertSuccess.innerText = 'Profile saved';
        alertSuccess.style.display = 'block';
        window.location.href = 'main';
      } else {
        alertError.innerText = 'Error: ' + result.message + ' (code ' + result.statusCode + ')';
        alertError.style.display = 'block';
      }
    } catch (err) {
      alertError.innerText = 'Network error: ' + err.message;
      alertError.style.display = 'block';
    }
  });
  