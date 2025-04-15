function normalizeResponse(resp) {
    return {
      success: resp.success,
      statusCode: resp.code,
      message: resp.message,
      data: resp.data
    };
  }
  let encryptionKey = null;
  async function getEncryptionKey() {
    if (encryptionKey) return encryptionKey;
    try {
      const res = await fetch('/api/encryption-key', { credentials: 'include' });
      if (!res.ok) throw new Error('Failed to get encryption key');
      encryptionKey = await res.text();
      return encryptionKey;
    } catch (err) {
      return null;
    }
  }
  async function encryptData(data, key) {
    if (!key) return null;
    const keyUtf8 = CryptoJS.enc.Base64.parse(key);
    return CryptoJS.AES.encrypt(data, keyUtf8, {
      mode: CryptoJS.mode.ECB,
      padding: CryptoJS.pad.Pkcs7
    }).toString();
  }
  document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const alertError = document.getElementById('alertError');
    const alertSuccess = document.getElementById('alertSuccess');
    alertError.style.display = 'none';
    alertSuccess.style.display = 'none';
    const username = document.getElementById('loginUsername').value.trim();
    const password = document.getElementById('loginPassword').value.trim();
    const key = await getEncryptionKey();
    if (!key) {
      alertError.innerText = 'Ошибка: не удалось получить ключ шифрования';
      alertError.style.display = 'block';
      return;
    }
    const encryptedUsername = await encryptData(username, key);
    const encryptedPassword = await encryptData(password, key);
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          username: encryptedUsername,
          password: encryptedPassword
        })
      });
      const raw = await response.json();
      const result = normalizeResponse(raw);
      if (result.success) {
        alertSuccess.innerText = 'Успешный вход';
        alertSuccess.style.display = 'block';
        window.location.href = 'main';
      } else {
        alertError.innerText = 'Ошибка: ' + result.message + ' (код ' + result.statusCode + ')';
        alertError.style.display = 'block';
      }
    } catch (err) {
      alertError.innerText = 'Сетевая ошибка: ' + err.message;
      alertError.style.display = 'block';
    }
  });
  