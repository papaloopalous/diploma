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
      const resp = await fetch('/api/encryption-key', { credentials: 'include' });
      if (!resp.ok) throw new Error("Key not received");
      encryptionKey = await resp.text();
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
  document.getElementById('registerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('registerUsername').value.trim();
    const password = document.getElementById('registerPassword').value.trim();
    const role = document.getElementById('registerRole').value;
    const alertError = document.getElementById('alertError');
    const alertSuccess = document.getElementById('alertSuccess');
    alertError.style.display = 'none';
    alertSuccess.style.display = 'none';
    const key = await getEncryptionKey();
    if (!key) {
      alertError.innerText = 'Error: failed to get encryption key';
      alertError.style.display = 'block';
      return;
    }
    const encryptedUsername = await encryptData(username, key);
    const encryptedPassword = await encryptData(password, key);
    try {
      const response = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          username: encryptedUsername,
          password: encryptedPassword,
          role: role
        })
      });
      const raw = await response.json();
      const result = normalizeResponse(raw);
      if (result.success) {
        alertSuccess.innerText = 'Registered successfully';
        alertSuccess.style.display = 'block';
        window.location.href = 'fill-profile';
      } else {
        alertError.innerText = 'Error: ' + result.message + ' (code ' + result.statusCode + ')';
        alertError.style.display = 'block';
      }
    } catch (err) {
      alertError.innerText = 'Network error: ' + err.message;
      alertError.style.display = 'block';
    }
  });
  