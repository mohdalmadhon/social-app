import { activePage } from '@/data/chatState';
import { addNotification } from '@/data/notifications';
import { handleIncomingNotification } from '@/data/notificationCount';

let ws = null;
const notificationDebounce = new Map();

export function connectToWS() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        return ws;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

    ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`);

    ws.onopen = () => {
        console.log('websocket connected');
    };

    ws.onmessage = (event) => {
        const payload = JSON.parse(event.data);
        switch (payload.type) {
            case 'notification':
                console.log('Notification:', payload.data);

                if (payload.data?.error) {
                    window.dispatchEvent(
                        new CustomEvent('message-send-error', {
                            detail: payload.data
                        })
                    );
                }

                handleIncomingNotification(payload.data);
                break;

            case 'message': {
                const message = payload.data;

                const groupID = message.GroupID;
                const privateChat =
                    activePage.value === 'chat:' + message.Sender.ID;

                const groupChat =
                    activePage.value === 'group:' + groupID;

                if (privateChat || groupChat) {
                    window.dispatchEvent(
                        new CustomEvent('chat-message', {
                            detail: message
                        })
                    );
                } else {
                    const now = Date.now();
                    const lastNotification =
                        notificationDebounce.get(groupID);

                    if (
                        !lastNotification ||
                        now - lastNotification >= 30 * 60 * 1000
                    ) {
                        addNotification(
                            `New message from ${message.Sender.firstName || 'user'}`,
                            'message'
                        );

                        notificationDebounce.set(groupID, now);
                    }
                }

                break;
            }
        }
    };

    ws.onclose = () => {
        console.log('websocket disconnected');
        ws = null;
    };

    ws.onerror = (error) => {
        console.error('websocket error:', error);
    };

    return ws;
}

export function sendWS(payload) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        console.error('websocket is disconnected');
        try {
            connectToWS()
        } catch (err) {
            addNotification('could not connect to websocket')
            return;
        }
        return;
    }

    console.log(payload)
    ws.send(JSON.stringify(payload));
}