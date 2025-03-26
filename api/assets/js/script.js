document.addEventListener("DOMContentLoaded", () => {
    const loginForm = document.getElementById("loginForm");
    const registerForm = document.getElementById("registerForm");
    const showLogin = document.getElementById("showLogin");
    const showRegister = document.getElementById("showRegister");
    let encryptionKey = null;

    showLogin.addEventListener("click", () => {
        loginForm.classList.remove("hidden");
        registerForm.classList.add("hidden");
    });

    showRegister.addEventListener("click", () => {
        registerForm.classList.remove("hidden");
        loginForm.classList.add("hidden");
    });

    async function getEncryptionKey() {
        if (encryptionKey) return encryptionKey;
        try {
            const response = await fetch('/api/encryption-key');
            encryptionKey = await response.text();
            return encryptionKey;
        } catch (error) {
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

    async function handleSubmit(event, endpoint, usernameId, passwordId, roleId = null) {
        event.preventDefault();
        
        const key = await getEncryptionKey();
        if (!key) return;

        const username = document.getElementById(usernameId).value;
        const password = document.getElementById(passwordId).value;

        const encryptedUsername = await encryptData(username, key);
        const encryptedPassword = await encryptData(password, key);

        const payload = {
            username: encryptedUsername,
            password: encryptedPassword
        };

        if (roleId) {
            const role = document.getElementById(roleId).value;
            payload.role = role;
        }

        fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
    }

    document.querySelector("#loginForm form").addEventListener("submit", function(event) {
        handleSubmit(event, '/api/login', 'loginUsername', 'loginPassword');
    });

    document.querySelector("#registerForm form").addEventListener("submit", function(event) {
        handleSubmit(event, '/api/register', 'registerUsername', 'registerPassword', 'registerRole');
    });
});
