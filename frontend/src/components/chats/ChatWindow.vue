<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';

import { addNotification } from '@/data/notifications';
import { Message } from '@/models/chats';
import { sendWS } from '@/api/socket/socket';
import { getMessages } from '@/api/chats/chats';

const props = defineProps({
    chat: {
        type: Object,
        default: null
    },
    groupID: {
        type: Number,
        default: null
    },
    userID: {
        type: Number,
        default: null
    },
    userFirstName: {
        type: String,
        default: ''
    },
    userLastName: {
        type: String,
        default: ''
    },
    userAvatar: {
        type: String,
        default: ''
    }
});


const message = ref('');
const messages = ref([]);
const sending = ref(false);
const loading = ref(false);
const loadingMore = ref(false);
const hasMore = ref(true);
const offset = ref(0);
const messagesContainer = ref(null);
const inviteStatus = ref({});
let fetchTimer = null;
let requestID = 0;

function parseInvite(content) {
    if (typeof content !== 'string') {
        return null;
    }

    try {
        const data = JSON.parse(content);

        if (data?.type !== 'invite' || !data.group || !data.user) {
            return null;
        }

        return {
            group: {
                id: data.group.id,
                name: data.group.name,
                avatar: data.group.avatar
            },
            user: {
                id: data.user.id,
                firstName: data.user.firstName,
                lastName: data.user.lastName,
                avatar: data.user.avatar
            }
        };
    } catch {
        return null;
    }
}

function formatMessage(msg) {
    const content = msg.Content ?? msg.content;
    const invite = parseInvite(content);

    return {
        id: msg.ID ?? msg.id,
        content,
        createdAt: msg.CreatedAt ?? msg.createdAt,
        sender: {
            id:
                msg.Sender?.ID ??
                msg.Sender?.id ??
                msg.sender?.ID ??
                msg.sender?.id,
            firstName:
                msg.Sender?.FirstName ??
                msg.Sender?.firstName ??
                msg.sender?.FirstName ??
                msg.sender?.firstName,
            lastName:
                msg.Sender?.LastName ??
                msg.Sender?.lastName ??
                msg.sender?.LastName ??
                msg.sender?.lastName,
            avatar:
                msg.Sender?.Avatar ??
                msg.Sender?.avatar ??
                msg.sender?.Avatar ??
                msg.sender?.avatar
        },
        groupID: msg.GroupID ?? msg.groupID,
        invite,
        inviteStatus: invite ? inviteStatus.value[invite.group.id] ?? null : null
    };
}

async function scrollToBottom() {
    await nextTick();

    if (messagesContainer.value) {
        messagesContainer.value.scrollTop =
            messagesContainer.value.scrollHeight;
    }
}

async function fetchChatMessages(groupID) {
    if (groupID === null || groupID === undefined) {
        messages.value = [];
        return;
    }

    const currentRequestID = ++requestID;

    offset.value = 0;
    hasMore.value = true;
    loading.value = true;
    loadingMore.value = false;
    messages.value = [];

    try {
        const result = await getMessages(groupID, 0);

        if (currentRequestID !== requestID) {
            return;
        }

        const data = Array.isArray(result)
            ? result
            : result.messages || result.data || [];

        messages.value = data
            .map(formatMessage)
            .reverse();

        offset.value = data.length;

        if (data.length < 20) {
            hasMore.value = false;
        }
    } catch (err) {
        if (currentRequestID !== requestID) {
            return;
        }

        messages.value = [];
        offset.value = 0;
        hasMore.value = false;

        addNotification(
            err.message || 'Error happened while fetching messages',
            'error'
        );
    } finally {
        if (currentRequestID === requestID) {
            loading.value = false;
            await scrollToBottom();
        }
    }
}

async function fetchOlderMessages() {
    if (
        props.groupID === null ||
        props.groupID === undefined ||
        loading.value ||
        loadingMore.value ||
        !hasMore.value
    ) {
        return;
    }

    const container = messagesContainer.value;

    if (!container) {
        return;
    }

    const currentRequestID = requestID;

    loadingMore.value = true;

    const oldScrollHeight = container.scrollHeight;
    const oldScrollTop = container.scrollTop;

    try {
        const result = await getMessages(
            props.groupID,
            offset.value
        );

        if (currentRequestID !== requestID) {
            return;
        }

        const data = Array.isArray(result)
            ? result
            : result.messages || result.data || [];

        if (data.length === 0) {
            hasMore.value = false;
            return;
        }

        const olderMessages = data
            .map(formatMessage)
            .reverse();

        messages.value = [
            ...olderMessages,
            ...messages.value
        ];

        offset.value += data.length;

        if (data.length < 20) {
            hasMore.value = false;
        }

        await nextTick();

        container.scrollTop =
            oldScrollTop +
            (container.scrollHeight - oldScrollHeight);
    } catch (err) {
        if (currentRequestID !== requestID) {
            return;
        }

        addNotification(
            err.message || 'Error happened while loading older messages',
            'error'
        );
    } finally {
        if (currentRequestID === requestID) {
            loadingMore.value = false;
        }
    }
}

function throttleFetchOlder() {
    if (fetchTimer) {
        return;
    }

    fetchTimer = setTimeout(() => {
        fetchTimer = null;

        if (
            messagesContainer.value &&
            messagesContainer.value.scrollTop <= 100
        ) {
            fetchOlderMessages();
        }
    }, 200);
}

function handleScroll() {
    const container = messagesContainer.value;

    if (!container) {
        return;
    }

    if (
        container.scrollTop <= 100 &&
        !loading.value &&
        !loadingMore.value &&
        hasMore.value
    ) {
        throttleFetchOlder();
    }
}

async function respondToInvite(msg, status) {
    if (!msg.invite || msg.invite.responding) return;

    msg.invite.responding = true;

    try {
        const response = await fetch('/api/groups/status', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                status,
                groupID: msg.invite.group.id,
                senderID: msg.invite.user.id,
                content: msg.rawContent
            })
        });

        const result = await response.json();

        if (!response.ok || !result.status) {
            throw new Error(result.message || 'Could not update invite');
        }

        messages.value = messages.value.filter(message => message !== msg);

        addNotification(
            status === 1 ? 'Invite accepted' : 'Invite rejected',
            'success'
        );
    } catch (err) {
        msg.invite.responding = false;
        addNotification(
            err.message || 'Could not update invite',
            'error'
        );
    }
}

function receiveMessage(event) {
    const incoming = event.detail;

    if (
        !props.chat ||
        incoming.GroupID !== props.groupID
    ) {
        return;
    }

    const formatted = formatMessage(incoming);

    const senderID = formatted.sender.id;

    messages.value.push({
        ...formatted,
        inviteStatus: formatted.invite
            ? inviteStatus.value[formatted.invite.group.id] ?? null
            : null
    });

    nextTick(() => {
        const container = messagesContainer.value;

        if (!container) {
            return;
        }

        const distanceFromBottom =
            container.scrollHeight -
            container.scrollTop -
            container.clientHeight;

        const ownMessage = senderID === props.userID;

        if (ownMessage || distanceFromBottom < 150) {
            container.scrollTop = container.scrollHeight;
        }
    });
}

async function send() {
    const content = message.value.trim();

    if (!content || sending.value || !props.chat) {
        return;
    }

    const msg = new Message(content);

    msg.userID = props.userID;
    msg.groupID = props.groupID;
    msg.private = 1;

    sending.value = true;

    try {
        sendWS({
            type: 'privateMessage',
            data: msg.getData()
        });

        messages.value.push({
            content,
            sender: {
                id: -1,
                firstName: props.userFirstName,
                lastName: props.userLastName,
                avatar: props.userAvatar
            },
            groupID: props.groupID
        });

        message.value = '';
        await scrollToBottom();
    } catch (err) {
        addNotification(
            err.message || 'Error happened while sending message',
            'error'
        );
    } finally {
        sending.value = false;
    }
}

watch(
    () => props.groupID,
    newGroupID => {
        if (fetchTimer) {
            clearTimeout(fetchTimer);
            fetchTimer = null;
        }

        requestID++;

        messages.value = [];
        offset.value = 0;
        hasMore.value = true;
        loadingMore.value = false;

        if (
            newGroupID === null ||
            newGroupID === undefined
        ) {
            loading.value = false;
            return;
        }

        fetchChatMessages(newGroupID);
    },
    { immediate: true }
);

watch(
    messagesContainer,
    (newEl, oldEl) => {
        if (oldEl) {
            oldEl.removeEventListener(
                'scroll',
                handleScroll
            );
        }

        if (newEl) {
            newEl.addEventListener(
                'scroll',
                handleScroll
            );
        }
    },
    { immediate: true }
);

onMounted(() => {
    window.addEventListener(
        'chat-message',
        receiveMessage
    );
});

onUnmounted(() => {
    window.removeEventListener(
        'chat-message',
        receiveMessage
    );

    if (messagesContainer.value) {
        messagesContainer.value.removeEventListener(
            'scroll',
            handleScroll
        );
    }

    if (fetchTimer) {
        clearTimeout(fetchTimer);
        fetchTimer = null;
    }

    requestID++;
});
</script>

<template>
    <section class="chat-window">
        <template v-if="chat">
            <header class="chat-window-header">
                <div class="avatar">
                    <img v-if="chat.Avatar" :src="`/uploads/${chat.Avatar}`" alt="" />
                </div>

                <div class="chat-user-info">
                    <strong>
                        {{ chat.FirstName }} {{ chat.LastName }}
                    </strong>
                </div>
            </header>

            <div ref="messagesContainer" class="messages">
                <div v-if="loadingMore" class="loading-more">
                    <div class="small-loader"></div>
                    <span>Loading older messages...</span>
                </div>

                <div v-if="loading" class="loading-state">
                    <div class="loader"></div>
                    <p>Loading messages...</p>
                </div>

                <template v-else>
                    <div v-for="(msg, index) in messages" :key="msg.id ?? index" class="message" :class="[
                        msg.sender?.id === -1
                            ? 'sent'
                            : 'received',
                        msg.invite
                            ? 'invite-message'
                            : ''
                    ]">
                        <div v-if="msg.sender?.id !== -1" class="message-avatar">
                            <img v-if="msg.sender?.avatar" :src="`/uploads/${msg.sender.avatar}`" alt="" />

                            <span v-else>
                                {{ msg.sender?.firstName?.[0] }}
                                {{ msg.sender?.lastName?.[0] }}
                            </span>
                        </div>

                        <div v-if="msg.invite" class="invite-card">
                            <div class="invite-group">
                                <div class="invite-group-avatar">
                                    <img v-if="msg.invite.group.avatar" :src="`/uploads/${msg.invite.group.avatar}`"
                                        alt="" />

                                    <span v-else>
                                        {{ msg.invite.group.name?.[0] }}
                                    </span>
                                </div>

                                <div class="invite-group-info">
                                    <span class="invite-label">
                                        Group invite
                                    </span>

                                    <strong>
                                        {{ msg.invite.group.name }}
                                    </strong>
                                </div>
                            </div>

                            <div class="invite-from">
                                <div class="invite-user-avatar">
                                    <img v-if="msg.invite.user.avatar" :src="`/uploads/${msg.invite.user.avatar}`"
                                        alt="" />

                                    <span v-else>
                                        {{ msg.invite.user.firstName?.[0] }}
                                        {{ msg.invite.user.lastName?.[0] }}
                                    </span>
                                </div>

                                <div>
                                    <span class="invite-label">
                                        Invite from
                                    </span>

                                    <strong>
                                        {{ msg.invite.user.firstName }}
                                        {{ msg.invite.user.lastName }}
                                    </strong>
                                </div>
                            </div>

                            <div v-if="msg.inviteStatus === null || msg.inviteStatus === undefined"
                                class="invite-actions">
                                <button type="button" class="invite-accept" @click="respondToInvite(msg, 1)">
                                    Accept
                                </button>
                                
                                <button type="button" class="invite-reject" @click="respondToInvite(msg, -1)">
                                    Reject
                                </button>
                            </div>

                            <div v-else class="invite-result">
                                <span v-if="msg.inviteStatus === 1">
                                    Invite accepted
                                </span>

                                <span v-else>
                                    Invite rejected
                                </span>
                            </div>
                        </div>

                        <div v-else class="message-body">
                            <span v-if="msg.sender?.id !== -1" class="message-sender-name">
                                {{ msg.sender?.firstName }}
                                {{ msg.sender?.lastName }}
                            </span>

                            <p>{{ msg.content }}</p>
                        </div>
                    </div>

                    <div v-if="messages.length === 0" class="no-messages">
                        <p>No messages yet</p>
                    </div>
                </template>
            </div>

            <form class="composer" @submit.prevent="send">
                <input v-model="message" type="text" placeholder="Type a message..." :disabled="loading" />

                <button type="submit" :disabled="sending || loading">
                    {{ sending ? 'Sending...' : 'Send' }}
                </button>
            </form>
        </template>

        <div v-else class="empty-state">
            <p class="eyebrow">
                NO CHAT SELECTED
            </p>

            <h2>
                Pick a conversation
            </h2>

            <p class="hint">
                Choose a chat from the list to start messaging.
            </p>
        </div>
    </section>
</template>

<style scoped>
.chat-window {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: calc(100vh - 64px - 40px);
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    overflow: hidden;
}

.chat-window-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 13px;
    padding: 16px 20px;
    border-bottom: 2px solid var(--page-background);
}

.chat-user-info {
    display: flex;
    flex-direction: column;
}

.chat-window-header strong {
    display: block;
    font-size: 14px;
}

.avatar {
    flex-shrink: 0;
    width: 44px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 18px;
    overflow: hidden;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.messages {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    overflow-y: auto;
    background: var(--page-background);
}

.loading-state {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    background: var(--page-background);
    z-index: 2;
}

.loading-state p {
    margin: 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.loader {
    width: 28px;
    height: 28px;
    border: 3px solid var(--main-color);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
}

.loading-more {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 24px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.small-loader {
    width: 14px;
    height: 14px;
    border: 2px solid var(--main-color);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.message {
    display: flex;
    align-items: flex-end;
    gap: 8px;
    max-width: 60%;
    font-size: 13px;
    line-height: 1.5;
}

.message-avatar {
    flex-shrink: 0;
    width: 26px;
    height: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-size: 10px;
    overflow: hidden;
}

.message-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.message-body {
    padding: 11px 15px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
}

.message-body p {
    margin: 0;
}

.message-sender-name {
    display: block;
    margin-bottom: 3px;
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    opacity: 0.8;
}

.message.received .message-body {
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.message.sent {
    align-self: flex-end;
    flex-direction: row-reverse;
}

.message.sent .message-body {
    background: var(--input-focus);
    color: white;
    box-shadow: 3px 3px var(--main-color);
}

.invite-message {
    max-width: 360px;
}

.invite-card {
    width: 100%;
    box-sizing: border-box;
    padding: 15px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.invite-group {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--main-color);
}

.invite-group-avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 18px;
    overflow: hidden;
}

.invite-group-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.invite-group-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
}

.invite-group-info strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 15px;
}

.invite-label {
    display: block;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 600;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.invite-from {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 13px;
}

.invite-from>div:last-child {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.invite-from strong {
    font-size: 11px;
}

.invite-user-avatar {
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-size: 10px;
    overflow: hidden;
}

.invite-user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.invite-actions {
    display: flex;
    gap: 8px;
    margin-top: 15px;
}

.invite-actions button {
    flex: 1;
    padding: 9px 12px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.invite-accept {
    background: var(--input-focus);
    color: white;
    box-shadow: 3px 3px var(--main-color);
}

.invite-reject {
    background: var(--bg-color);
    color: var(--main-color);
    box-shadow: 3px 3px var(--main-color);
}

.invite-actions button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.invite-actions button:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.invite-result {
    margin-top: 15px;
    padding: 9px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
}

.composer {
    flex-shrink: 0;
    display: flex;
    gap: 10px;
    padding: 14px 20px;
    border-top: 2px solid var(--page-background);
}

.composer input {
    flex: 1;
    height: 42px;
    padding: 0 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.composer input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.composer button {
    padding: 0 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
}

.composer button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.composer button:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.no-messages {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
}

.no-messages p {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 20px;
}

.empty-state .eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    letter-spacing: 2px;
}

.empty-state h2 {
    margin: 0 0 8px;
    font-family: "Liter", serif;
    font-size: 26px;
}

.empty-state .hint {
    margin: 0;
    color: var(--font-color-sub);
    font-size: 12px;
}

@media (max-width: 800px) {
    .chat-window {
        height: 480px;
    }

    .message {
        max-width: 80%;
    }

    .invite-message {
        max-width: 90%;
    }
}
</style>
