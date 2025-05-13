function normalizeResponse(resp) {
  return {
    success:    resp.success,
    statusCode: resp.code,
    message:    resp.message,
    data:       resp.data
  };
}

const PRIME_HEX = 'FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD1';
const GENERATOR = 2n;
const PRIVATE   = BigInt('0x9876543210FEDCBA9876543210FEDCBA98765432');

class KeyExchange {
  constructor () {
    this.p    = BigInt('0x' + PRIME_HEX);
    this.g    = GENERATOR;
    this.priv = PRIVATE;
    this.pub  = this.modPow(this.g, this.priv, this.p);
  }
  modPow(b, e, m) {
    let r = 1n;
    b %= m;
    while (e) { if (e & 1n) r = (r * b) % m; b = (b * b) % m; e >>= 1n; }
    return r;
  }
  shared(serverPub) { return this.modPow(BigInt(serverPub), this.priv, this.p); }
}

const kex = new KeyExchange();
let sharedKeyHex = null;

async function getSharedKey () {
  if (sharedKeyHex) return sharedKeyHex;

  const r  = await fetch('/api/key-exchange', {
    method : 'POST',
    headers: { 'Content-Type': 'application/json' },
    body   : JSON.stringify({ clientPublic: kex.pub.toString() })
  });
  const { success, data, message } = await r.json();
  if (!success) throw new Error(message);

  sharedKeyHex = CryptoJS.SHA256(
                  kex.shared(data.serverPublic).toString()
                ).toString(CryptoJS.enc.Hex);

  if (sharedKeyHex.length !== 64)
    throw new Error('Shared key must be 32 bytes (64 hex chars)');

  return sharedKeyHex;
}

function deriveKeyAndIV(keyHexWA, saltWA) {
  let acc  = CryptoJS.lib.WordArray.create();
  let prev = keyHexWA;

  while (acc.sigBytes < 48) {
    prev = CryptoJS.MD5(prev.concat(saltWA));
    acc  = acc.concat(prev);
  }
  return {
    key: CryptoJS.lib.WordArray.create(acc.words.slice(0, 8)),
    iv : CryptoJS.lib.WordArray.create(acc.words.slice(8, 12))
  };
}

async function encryptData (plain) {
  if (typeof plain !== 'string')
    throw new TypeError('encryptData expects a string');

  const keyHex = await getSharedKey();
  const salt   = CryptoJS.lib.WordArray.random(8);
  const { key, iv } = deriveKeyAndIV(CryptoJS.enc.Hex.parse(keyHex), salt);

  const cipher = CryptoJS.AES.encrypt(
    CryptoJS.enc.Utf8.parse(plain),
    key,
    { iv, mode: CryptoJS.mode.CBC, padding: CryptoJS.pad.Pkcs7 }
  );

  const salted = CryptoJS.enc.Utf8.parse('Salted__').concat(salt)
                  .concat(cipher.ciphertext);
  return CryptoJS.enc.Base64.stringify(salted);
}

document.getElementById('registerForm').addEventListener('submit', async e => {
  e.preventDefault();

  const $err = document.getElementById('alertError');
  const $ok  = document.getElementById('alertSuccess');
  $err.style.display = 'none'; $ok.style.display = 'none';

  try {
    const username = document.getElementById('registerUsername').value.trim();
    const password = document.getElementById('registerPassword').value.trim();
    const role     = document.getElementById('registerRole').value;

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
      throw new Error(`Error: ${result.message} (code ${result.statusCode})`);
    }
  } catch (err) {
    $err.textContent = err.message || 'Unexpected error';
    $err.style.display = 'block';
    console.error(err);
  }
});
