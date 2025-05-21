const messageTemplate = (message) => `
    <div class="message-wrapper">
        <div class="message ${message.isSender ? 'message-sent' : 'message-received'}" data-message-id="${message.id}">
            <div class="message-text">${message.text}</div>
            <div class="message-info">
                <div class="message-time">
                    ${new Date(message.sentAt).toLocaleTimeString()}
                </div>
                ${message.isSender ? `
                    <div class="message-status" data-status="${message.status}">
                        ${getStatusIcon(message.status)}
                    </div>
                ` : ''}
            </div>
        </div>
    </div>
`;

const getStatusIcon = (status) => {
    switch (status) {
        case 'sent': return '✓';
        case 'delivered': return '✓✓';
        default: return '✓';
    }
};

let ws;
const connectWebSocket = () => {
    const roomId = new URLSearchParams(window.location.search).get('room');
    if (!roomId) {
        console.error('No room ID provided');
        return;
    }

    ws = new WebSocket(`ws://localhost:8081/ws?room=${roomId}`);
    
    ws.onopen = () => {
        console.log("Connected to chat");
        document.querySelector('.status-indicator').classList.replace('status-offline', 'status-online');
    };

    ws.onclose = () => {
        console.log("Disconnected from chat");
        document.querySelector('.status-indicator').classList.replace('status-online', 'status-offline');
        setTimeout(connectWebSocket, 3000);
    };

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            const chatMessages = document.getElementById('chat-messages');
            
            if (message.type === 'message') {
                const messageElement = document.createElement('div');
                messageElement.innerHTML = messageTemplate(message);
                chatMessages.appendChild(messageElement);
                chatMessages.scrollTop = chatMessages.scrollHeight;
            } else if (message.type === 'status') {
                const messageEl = document.querySelector(`[data-message-id="${message.id}"]`);
                if (messageEl) {
                    const statusEl = messageEl.querySelector('.message-status');
                    if (statusEl) {
                        statusEl.setAttribute('data-status', message.status);
                        statusEl.innerHTML = getStatusIcon(message.status);
                    }
                }
            }
        } catch (err) {
            console.error('Failed to parse message:', err);
        }
    };
};

document.getElementById('send-button').addEventListener('click', () => {
    const input = document.getElementById('message-input');
    const text = input.value.trim();
    
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;

    const message = {
        type: 'message',
        text: text.replace(/\n/g, '<br>'),
        sentAt: new Date().toISOString()
    };

    ws.send(JSON.stringify(message));
    input.value = '';
    input.style.height = 'auto';
});

const messageInput = document.getElementById('message-input');

function autoResize() {
    messageInput.style.height = 'auto';
    messageInput.style.height = messageInput.scrollHeight + 'px';
}

messageInput.addEventListener('input', autoResize);
messageInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        document.getElementById('send-button').click();
        messageInput.style.height = 'auto';
    }
});

try {
    connectWebSocket();
} catch (err) {
    console.error('Failed to connect:', err);
}