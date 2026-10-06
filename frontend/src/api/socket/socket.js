import { activePage, openGroupPage } from '@/data/chatState';
import { addNotification, postImageUrl, avatarUrl } from '@/data/notifications';
import { setTyping } from '@/data/typingState';
import {
    handleIncomingNotification,
    incrementUnreadNotificationCount,
    refreshUnreadNotificationCount,
    setUnreadNotificationCount
} from '@/data/notificationCount';

let ws = null;
let reconnectTimer = null;
const notificationDebounce = new Map();
const toastDebounce = new Map();
const TOAST_COOLDOWN = 10 * 1000;
const MESSAGE_COOLDOWN = 5 * 1000;

function isDuplicateToast(payload) {
    const key = `${payload.actor?.id || 0}:${payload.kind || ''}:${payload.post_id || 0}:${payload.message}`;
    const now = Date.now();
    const last = toastDebounce.get(key);

    toastDebounce.set(key, now);

    for (const [storedKey, time] of toastDebounce) {
        if (now - time > TOAST_COOLDOWN) {
            toastDebounce.delete(storedKey);
        }
    }

    return !!last && now - last < TOAST_COOLDOWN;
}

function viewingNotifications() {
    return window.location.pathname === '/notifications' && !document.hidden;
}

function initialOf(firstName) {
    return (firstName || '').trim().charAt(0).toUpperCase();
}

function notificationOptions(payload) {
    const options = {
        avatar: avatarUrl(payload.actor?.avatarPath),
        initial: initialOf(payload.actor?.firstName),
        notificationId: payload.id
    };

    if (payload.post_id) {
        options.postId = payload.post_id;
        options.image = postImageUrl(payload.image_path);
    } else if (payload.kind === 'follow_request') {
        options.route = { path: '/notifications', query: { tab: 'requests' } };
    } else if (payload.kind === 'post mention' && payload.group_id) {
        options.route = { path: `/groups/${payload.group_id}`, query: { tab: 'chat' } };
    } else if (payload.kind === 'event invite') {
        options.route = { path: '/notifications', query: { tab: 'events' } };
    } else if (payload.kind === 'event response' && payload.group_id) {
        options.route = { path: `/groups/${payload.group_id}`, query: { tab: 'events' } };
    } else if (payload.kind === 'group invite' || payload.kind === 'group join') {
        options.route = { path: '/notifications', query: { tab: 'invites' } };
    } else if (payload.group_id) {
        options.route = { path: `/groups/${payload.group_id}` };
    } else {
        options.route = { path: '/notifications' };
    }

    return options;
}

function messageOptions(payload) {
    const message = payload.data;
    const sender = message.Sender || {};

    const options = {
        avatar: avatarUrl(sender.avatar),
        initial: initialOf(sender.firstName)
    };

    if (payload.isPrivate) {
        options.route = {
            path: '/chats',
            query: {
                userId: sender.ID,
                firstName: sender.firstName || '',
                lastName: sender.lastName || '',
                avatar: sender.avatar || ''
            }
        };
    } else {
        options.route = {
            path: `/groups/${message.GroupID}`,
            query: { tab: 'chat' }
        };
    }

    return options;
}

export function connectToWS() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        return ws;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

    ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`);

    let opened = false;

    ws.onopen = () => {
        opened = true;
        console.log('websocket connected');
        refreshUnreadNotificationCount();
    };

    ws.onmessage = (event) => {
        const payload = JSON.parse(event.data);
        switch (payload.type) {
            case 'notification':
                console.log('Notification:', payload.data);

                if (payload.error == true) {
                    addNotification(payload.message || 'could not send notification', 'error')
                    return
                }

                if (payload.message) {
                    if (!viewingNotifications()) {
                        if (typeof payload.unread === 'number') {
                            setUnreadNotificationCount(payload.unread);
                        } else {
                            incrementUnreadNotificationCount();
                        }
                    }

                    window.dispatchEvent(
                        new CustomEvent('notification-received', {
                            detail: payload
                        })
                    );

                    if (!isDuplicateToast(payload)) {
                        addNotification(
                            payload.message,
                            'success',
                            notificationOptions(payload)
                        );
                    }

                    return;
                }

                if (payload.data?.error) {
                    window.dispatchEvent(
                        new CustomEvent('message-send-error', {
                            detail: payload.data
                        })
                    );
                }

                handleIncomingNotification(payload.data);
                break;

            case 'typing': {
                const typingData = payload.data || {};

                setTyping(typingData.userID, Boolean(typingData.typing));

                break;
            }

            case 'message': {
                const message = payload.data;

                const groupID = message.GroupID;

                if (payload.isPrivate && message.Sender) {
                    setTyping(message.Sender.ID, false);

                    window.dispatchEvent(
                        new CustomEvent('private-chat-activity', {
                            detail: {
                                userID: message.Sender.ID,
                                groupID: message.GroupID,
                                firstName: message.Sender.firstName || '',
                                lastName: message.Sender.lastName || '',
                                avatar: message.Sender.avatar || '',
                                own: false
                            }
                        })
                    );
                }
                const privateChat =
                    activePage.value === 'chat:' + message.Sender.ID;

                const groupChat =
                    activePage.value === 'group:' + groupID;

                const onGroupPage =
                    !payload.isPrivate &&
                    Number(openGroupPage.value) === Number(groupID);

                if (privateChat || groupChat) {
                    window.dispatchEvent(
                        new CustomEvent('chat-message', {
                            detail: message
                        })
                    );
                } else if (onGroupPage) {
                    window.dispatchEvent(
                        new CustomEvent('chat-message', {
                            detail: message
                        })
                    );
                } else if (!payload.silent) {
                    const now = Date.now();
                    const lastNotification =
                        notificationDebounce.get(groupID);

                    if (
                        !lastNotification ||
                        now - lastNotification >= MESSAGE_COOLDOWN
                    ) {
                        const sender = message.Sender.firstName || 'user';

                        addNotification(
                            !payload.isPrivate && payload.groupName
                                ? `New message from ${sender} in ${payload.groupName}`
                                : `New message from ${sender}`,
                            'message',
                            messageOptions(payload)
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

        if (opened && !reconnectTimer) {
            reconnectTimer = setTimeout(() => {
                reconnectTimer = null;
                connectToWS();
            }, 3000);
        }
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