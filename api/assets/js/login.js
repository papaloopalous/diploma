function normalizeResponse(resp) {
  return {
    success:    resp.success,
    statusCode: resp.code,
    message:    resp.message,
    data:       resp.data
  };
}

document.getElementById('loginForm').addEventListener('submit', async e => {
  e.preventDefault();

  const $err = document.getElementById('alertError');
  const $ok = document.getElementById('alertSuccess');
  $err.style.display = 'none'; $ok.style.display = 'none';

  try {
    const username = document.getElementById('loginUsername').value.trim();
    const password = document.getElementById('loginPassword').value.trim();

    const [encUser, encPass] = await Promise.all([
      encryptData(username),
      encryptData(password)
    ]);

    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: encUser, password: encPass })
    });

    const result = normalizeResponse(await res.json());

    if (result.success) {
      $ok.textContent = 'Login successful';
      $ok.style.display = 'block';
      location.href = 'main';
    } else {
      throw new Error(`Error: ${result.message} (code ${result.statusCode})`);
    }
  } catch (err) {
    $err.textContent = err.message || 'Unexpected error';
    $err.style.display = 'block';
    console.error(err);
  }
});
