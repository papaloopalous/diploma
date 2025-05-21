function normalizeResponse(resp) {
  return {
    success:    resp.success,
    statusCode: resp.code,
    message:    resp.message,
    data:       resp.data
  };
}

document.getElementById('registerForm').addEventListener('submit', async e => {
  e.preventDefault();

  const $err = document.getElementById('alertError');
  const $ok = document.getElementById('alertSuccess');
  $err.style.display = 'none'; $ok.style.display = 'none';

  try {
    const username = document.getElementById('registerUsername').value.trim();
    const password = document.getElementById('registerPassword').value.trim();
    const role = document.getElementById('registerRole').value;

    const [encUser, encPass] = await Promise.all([
      encryptData(username),
      encryptData(password)
    ]);

    const res = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: encUser, password: encPass, role })
    });

    const result = normalizeResponse(await res.json());

    if (result.success) {
      $ok.textContent = 'Registered successfully';
      $ok.style.display = 'block';
      location.href = 'fill-profile';
    } else {
      throw new Error(`${result.message}`);
    }
  } catch (err) {
    $err.textContent = err.message || 'Unexpected error';
    $err.style.display = 'block';
    console.error(err);
  }
});
