<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { addNotification } from '@/data/notifications';
import { chatsSidebarOpen, toggleChatsSidebar } from '@/data/chatState';
import { Message } from '@/models/chats';
import { sendWS } from '@/api/socket/socket';
import { isUserTyping } from '@/data/typingState';
import { getMessages, sendChatMedia } from '@/api/chats/chats';
import { router } from '@/router/router';
import { CHAT_MEDIA_ACCEPT, parseChatMedia, validateChatMedia } from '@/helpers/chatMedia';
import HomePosts from '../home/HomePosts.vue';
import EmojiPicker from './EmojiPicker.vue';

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

const emit = defineEmits(['chat-resolved']);

const message = ref('');
const messageInput = ref(null);
const messages = ref([]);
const fileInput = ref(null);
const pendingFile = ref(null);
const pendingPreview = ref('');
const lightboxSrc = ref('');
const sending = ref(false);
const loading = ref(false);
const loadingMore = ref(false);
const hasMore = ref(true);
const offset = ref(0);
const messagesContainer = ref(null);
const inviteStatus = ref({});
const canMessage = ref(false);

let fetchTimer = null;
let requestID = 0;
let typingTarget = null;
let typingStopTimer = null;
let lastTypingSent = 0;

const TYPING_RESEND = 2000;
const TYPING_IDLE = 3000;

const partnerTyping = computed(
    () => Boolean(props.userID) && isUserTyping(props.userID)
);

function emitTyping(target, typing) {
    sendWS({
        type: 'typing',
        data: {
            userID: target.userID,
            groupID: target.groupID || -1,
            typing
        }
    });
}

function stopTyping() {
    if (typingStopTimer) {
        clearTimeout(typingStopTimer);
        typingStopTimer = null;
    }

    if (!typingTarget) {
        return;
    }

    emitTyping(typingTarget, false);

    typingTarget = null;
    lastTypingSent = 0;
}

function handleTypingInput(value) {
    if (!props.chat || !props.userID || !canMessage.value) {
        return;
    }

    if (!value || !value.trim()) {
        stopTyping();
        return;
    }

    if (typingTarget && typingTarget.userID !== props.userID) {
        stopTyping();
    }

    if (!typingTarget) {
        typingTarget = {
            userID: props.userID,
            groupID: props.groupID
        };

        lastTypingSent = 0;
    }

    const now = Date.now();

    if (now - lastTypingSent >= TYPING_RESEND) {
        emitTyping(typingTarget, true);
        lastTypingSent = now;
    }

    if (typingStopTimer) {
        clearTimeout(typingStopTimer);
    }

    typingStopTimer = setTimeout(stopTyping, TYPING_IDLE);
}

function announceActivity(groupID) {
    window.dispatchEvent(
        new CustomEvent('private-chat-activity', {
            detail: {
                userID: props.userID,
                groupID,
                firstName: props.userFirstName,
                lastName: props.userLastName,
                avatar: props.userAvatar,
                own: true
            }
        })
    );
}

const postCache = new Map();

function getSidebarCanMessage() {
    return Boolean(props.chat?.canMessage);
}

function updateCanMessage(messagePermission) {
    canMessage.value = Boolean(
        canMessage.value ||
        getSidebarCanMessage() ||
        messagePermission
    );
}

function pickFile() {
    if (fileInput.value) {
        fileInput.value.click();
    }
}

function clearPending() {
    if (pendingPreview.value) {
        URL.revokeObjectURL(pendingPreview.value);
    }

    pendingFile.value = null;
    pendingPreview.value = '';

    if (fileInput.value) {
        fileInput.value.value = '';
    }
}

function onFileChange(event) {
    const file = event.target.files?.[0];

    if (!file) {
        return;
    }

    const error = validateChatMedia(file);

    if (error) {
        addNotification(error, 'error');
        event.target.value = '';
        return;
    }

    if (pendingPreview.value) {
        URL.revokeObjectURL(pendingPreview.value);
    }

    pendingFile.value = file;
    pendingPreview.value = URL.createObjectURL(file);
}

function openImage(path) {
    lightboxSrc.value = `/uploads/${path}`;
}

function closeImage() {
    lightboxSrc.value = '';
}

function sendToProfile() {
    router.push(`/user?id=${props.userID}`);
    window.location.reload();
}

function parseInvite(content) {
    if (typeof content !== 'string') {
        return null;
    }

    try {
        const data = JSON.parse(content);

        if (
            !data?.type ||
            data.type !== 'invite' ||
            !data.group ||
            !data.user
        ) {
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

function parseSharedPost(content) {
    if (typeof content !== 'string') {
        return null;
    }

    try {
        const data = JSON.parse(content);

        if (
            data?.type !== 'message' ||
            !Number.isInteger(Number(data?.postID)) ||
            Number(data.postID) <= 0
        ) {
            return null;
        }

        return {
            postID: Number(data.postID)
        };
    } catch {
        return null;
    }
}

function parseSharedProfile(content) {
    if (
        typeof content !== 'string' ||
        !content.startsWith('{')
    ) {
        return null;
    }

    try {
        const data = JSON.parse(content);

        if (
            data?.type !== 'profile' ||
            !Number.isInteger(Number(data?.profileID)) ||
            Number(data.profileID) <= 0
        ) {
            return null;
        }

        return {
            id: Number(data.profileID),
            firstName: data.firstName || '',
            lastName: data.lastName || '',
            username: data.username || '',
            avatar: data.avatar || ''
        };
    } catch {
        return null;
    }
}

function openSharedProfile(profile) {
    router.push({
        path: '/user',
        query: {
            id: profile.id
        }
    });
}

async function getSharedPost(postID) {
    if (postCache.has(postID)) {
        return postCache.get(postID);
    }

    const request = fetch(
        `/api/post/single?postID=${encodeURIComponent(postID)}`,
        {
            method: 'GET',
            credentials: 'include'
        }
    )
        .then(async response => {
            let result = null;

            try {
                result = await response.json();
            } catch {
                result = null;
            }

            if (!response.ok || !result?.status) {
                throw new Error(
                    result?.message ||
                    'Could not load shared post'
                );
            }

            return result.data;
        })
        .catch(error => {
            postCache.delete(postID);
            throw error;
        });

    postCache.set(postID, request);

    return request;
}

function getPostValue(post, ...keys) {
    for (const key of keys) {
        if (
            post &&
            post[key] !== undefined &&
            post[key] !== null
        ) {
            return post[key];
        }
    }

    return null;
}

function formatPost(post) {
    if (!post) {
        return null;
    }

    const reaction = Number(
        getPostValue(
            post,
            'ReactionValue',
            'reactionValue',
            'Reaction',
            'reaction'
        ) ?? 0
    );

    return {
        currentUserId: props.userID,
        allowComments: Boolean(
            getPostValue(
                post,
                'AllowComments',
                'allowComments'
            )
        ),
        reaction,
        userId: Number(
            getPostValue(
                post,
                'UserId',
                'userId',
                'UserID',
                'userID'
            ) ?? 0
        ),
        postId: Number(
            getPostValue(
                post,
                'Id',
                'id',
                'PostId',
                'postId'
            ) ?? 0
        ),
        groupId: Number(
            getPostValue(
                post,
                'GroupId',
                'groupId',
                'GroupID',
                'groupID'
            ) ?? 0
        ),
        firstName: getPostValue(
            post,
            'FirstName',
            'firstName'
        ) ?? '',
        lastName: getPostValue(
            post,
            'LastName',
            'lastName'
        ) ?? '',
        username: getPostValue(
            post,
            'Username',
            'username'
        ) ?? '',
        avatarPath: getPostValue(
            post,
            'AvatarPath',
            'avatarPath',
            'Avatar',
            'avatar'
        ) ?? '',
        createdAt: getPostValue(
            post,
            'CreatedAt',
            'createdAt'
        ),
        content: getPostValue(
            post,
            'Content',
            'content'
        ) ?? '',
        imagePath: getPostValue(
            post,
            'ImagePath',
            'imagePath'
        ),
        location: getPostValue(
            post,
            'Location',
            'location'
        ),
        taggedPeople: getPostValue(
            post,
            'TaggedPeople',
            'taggedPeople'
        ) ?? [],
        likes: Number(
            getPostValue(
                post,
                'LikeCount',
                'likeCount',
                'Likes',
                'likes'
            ) ?? 0
        ),
        dislikes: Number(
            getPostValue(
                post,
                'DisLikeCount',
                'disLikeCount',
                'DislikeCount',
                'dislikeCount',
                'Dislikes',
                'dislikes'
            ) ?? 0
        ),
        comments: [],
        userReaction:
            reaction === 1
                ? 'like'
                : reaction === -1
                    ? 'dislike'
                    : '',
        relationship: getPostValue(
            post,
            'Relationship',
            'relationship'
        ),
        visibility: getPostValue(
            post,
            'Visibility',
            'visibility'
        ),
        visibilityUser: getPostValue(
            post,
            'VisibilityUser',
            'visibilityUser'
        ),
        commentCount: Number(
            getPostValue(
                post,
                'CommentCount',
                'commentCount',
                'CommentsCount',
                'commentsCount'
            ) ?? 0
        )
    };
}

async function parsePostMessage(content) {
    const sharedPost = parseSharedPost(content);

    if (!sharedPost) {
        return {
            post: null,
            postError: null
        };
    }

    try {
        const post = await getSharedPost(
            sharedPost.postID
        );

        return {
            post: formatPost(post),
            postError: null
        };
    } catch (error) {
        return {
            post: null,
            postError:
                error.message ||
                'Could not load shared post'
        };
    }
}

async function formatMessage(msg) {
    const content = msg.Content ?? msg.content;

    const invite = parseInvite(content);

    const sharedProfile = invite
        ? null
        : parseSharedProfile(content);

    const sharedPostInfo =
        invite || sharedProfile
            ? {
                post: null,
                postError: null
            }
            : await parsePostMessage(content);

    const media =
        invite ||
        sharedProfile ||
        sharedPostInfo.post
            ? null
            : parseChatMedia(content);

    return {
        id: msg.ID ?? msg.id,
        clientID: msg.ClientID ?? msg.clientID,
        content,
        rawContent: content,
        media,
        post: sharedPostInfo.post,
        postError: sharedPostInfo.postError,
        profile: sharedProfile,
        createdAt:
            msg.CreatedAt ?? msg.createdAt,
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
        groupID:
            msg.GroupID ?? msg.groupID,
        invite,
        inviteStatus: invite
            ? inviteStatus.value[
                invite.group.id
            ] ?? null
            : null
    };
}

async function formatMessages(data) {
    return Promise.all(
        data.map(message =>
            formatMessage(message)
        )
    );
}

async function scrollToBottom() {
    await nextTick();

    if (messagesContainer.value) {
        messagesContainer.value.scrollTop =
            messagesContainer.value.scrollHeight;
    }
}

async function fetchChatMessages(groupID) {
    const currentRequestID = ++requestID;

    offset.value = 0;
    hasMore.value = true;
    loading.value = true;
    loadingMore.value = false;
    messages.value = [];

    canMessage.value = getSidebarCanMessage();

    try {
        const result = await getMessages(
            groupID,
            0,
            props.userID
        );
        console.log(result)
        if (currentRequestID !== requestID) {
            return;
        }

        updateCanMessage(
            Array.isArray(result)
                ? false
                : result?.canMessage
        );

        const data = Array.isArray(result)
            ? result
            : result.messages ||
              result.data ||
              [];

        const formattedMessages =
            await formatMessages(data);

        if (currentRequestID !== requestID) {
            return;
        }

        messages.value =
            formattedMessages.reverse();

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

        canMessage.value =
            getSidebarCanMessage();

        addNotification(
            err.message ||
            'Error happened while fetching messages',
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

    const container =
        messagesContainer.value;

    if (!container) {
        return;
    }

    const currentRequestID = requestID;

    loadingMore.value = true;

    const oldScrollHeight =
        container.scrollHeight;

    const oldScrollTop =
        container.scrollTop;

    try {
        const result = await getMessages(
            props.groupID,
            offset.value,
            props.userID
        );

        if (currentRequestID !== requestID) {
            return;
        }

        updateCanMessage(
            Array.isArray(result)
                ? false
                : result?.canMessage
        );

        const data = Array.isArray(result)
            ? result
            : result.messages ||
              result.data ||
              [];

        if (data.length === 0) {
            hasMore.value = false;
            return;
        }

        const olderMessages =
            await formatMessages(data);

        if (currentRequestID !== requestID) {
            return;
        }

        messages.value = [
            ...olderMessages.reverse(),
            ...messages.value
        ];

        offset.value += data.length;

        if (data.length < 20) {
            hasMore.value = false;
        }

        await nextTick();

        container.scrollTop =
            oldScrollTop +
            (container.scrollHeight -
                oldScrollHeight);
    } catch (err) {
        if (currentRequestID !== requestID) {
            return;
        }

        addNotification(
            err.message ||
            'Error happened while loading older messages',
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
    const container =
        messagesContainer.value;

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
    if (
        !msg.invite ||
        msg.invite.responding
    ) {
        return;
    }

    msg.invite.responding = true;

    try {
        const response = await fetch(
            '/api/groups/status',
            {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type':
                        'application/json'
                },
                body: JSON.stringify({
                    status,
                    groupID:
                        msg.invite.group.id,
                    senderID:
                        msg.invite.user.id,
                    content:
                        msg.rawContent
                })
            }
        );

        const result =
            await response.json();

        if (
            !response.ok ||
            !result.status
        ) {
            throw new Error(
                result.message ||
                'Could not update invite'
            );
        }

        messages.value =
            messages.value.filter(
                message => message !== msg
            );

        addNotification(
            status === 1
                ? 'Invite accepted'
                : 'Invite rejected',
            'success'
        );
    } catch (err) {
        msg.invite.responding = false;

        addNotification(
            err.message ||
            'Could not update invite',
            'error'
        );
    }
}

async function receiveMessage(event) {
    const incoming = event.detail;

    if (
        !props.chat ||
        incoming.GroupID !== props.groupID
    ) {
        return;
    }

    const formatted =
        await formatMessage(incoming);

    const senderID =
        formatted.sender.id;

    messages.value.push({
        ...formatted,
        inviteStatus: formatted.invite
            ? inviteStatus.value[
                formatted.invite.group.id
            ] ?? null
            : null
    });

    nextTick(() => {
        const container =
            messagesContainer.value;

        if (!container) {
            return;
        }

        const distanceFromBottom =
            container.scrollHeight -
            container.scrollTop -
            container.clientHeight;

        const ownMessage =
            senderID === props.userID;

        if (
            ownMessage ||
            distanceFromBottom < 150
        ) {
            container.scrollTop =
                container.scrollHeight;
        }
    });
}

function insertEmoji(emoji) {
    const input = messageInput.value;
    const current = message.value;

    const start =
        input?.selectionStart ??
        current.length;

    const end =
        input?.selectionEnd ??
        start;

    message.value =
        current.slice(0, start) +
        emoji +
        current.slice(end);

    const position =
        start + emoji.length;

    nextTick(() => {
        if (!input) {
            return;
        }

        input.focus();
        input.setSelectionRange(
            position,
            position
        );
    });
}

async function send() {
    const content =
        message.value.trim();

    const file =
        pendingFile.value;

    if (
        (!content && !file) ||
        sending.value ||
        !props.chat ||
        !canMessage.value
    ) {
        return;
    }

    sending.value = true;

    let resolvedGroupID =
        props.groupID;

    try {
        if (file) {
            const result =
                await sendChatMedia(file, {
                    userID: props.userID,
                    groupID: props.groupID
                });

            resolvedGroupID =
                result.groupID ??
                resolvedGroupID;

            messages.value.push({
                content:
                    result.content,
                rawContent:
                    result.content,
                media:
                    parseChatMedia(
                        result.content
                    ),
                post: null,
                postError: null,
                createdAt:
                    new Date().toISOString(),
                sender: {
                    id: -1,
                    firstName:
                        props.userFirstName,
                    lastName:
                        props.userLastName,
                    avatar:
                        props.userAvatar
                },
                groupID:
                    resolvedGroupID,
                invite: null
            });

            clearPending();
            announceActivity(resolvedGroupID);
            await scrollToBottom();
        }

        if (content) {
            const clientID =
                crypto.randomUUID();

            const msg =
                new Message(content);

            msg.userID =
                props.userID;

            msg.groupID =
                resolvedGroupID;

            msg.private = 1;

            msg.clientID =
                clientID;

            sendWS({
                type: 'privateMessage',
                data: msg.getData()
            });

            announceActivity(resolvedGroupID);

            const sharedPost =
                parseSharedPost(content);

            let localPost = null;

            if (sharedPost) {
                try {
                    const post =
                        await getSharedPost(
                            sharedPost.postID
                        );

                    localPost =
                        formatPost(post);
                } catch {
                    localPost = null;
                }
            }

            messages.value.push({
                clientID,
                content,
                rawContent: content,
                media: null,
                post: localPost,
                postError: null,
                sender: {
                    id: -1,
                    firstName:
                        props.userFirstName,
                    lastName:
                        props.userLastName,
                    avatar:
                        props.userAvatar
                },
                groupID:
                    resolvedGroupID,
                sending: true,
                failed: false,
                error: null
            });

            message.value = '';

            await scrollToBottom();
        }

        if (
            resolvedGroupID !== null &&
            resolvedGroupID !== undefined &&
            resolvedGroupID !== props.groupID
        ) {
            emit(
                'chat-resolved',
                resolvedGroupID
            );
        }
    } catch (err) {
        addNotification(
            err.message ||
            'Error happened while sending message',
            'error'
        );
    } finally {
        sending.value = false;
    }
}

watch(message, handleTypingInput);

watch(
    () => props.userID,
    (newID, oldID) => {
        if (newID !== oldID) {
            stopTyping();
        }
    }
);

watch(
    () => props.chat?.canMessage,
    value => {
        if (value) {
            canMessage.value = true;
        }
    },
    {
        immediate: true
    }
);

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

        canMessage.value =
            getSidebarCanMessage();

        if (
            newGroupID === null ||
            newGroupID === undefined
        ) {
            loading.value = false;
            return;
        }

        fetchChatMessages(newGroupID);
    },
    {
        immediate: true
    }
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
    {
        immediate: true
    }
);

function handleMessageSendError(event) {
    const error = event.detail;

    const msg =
        messages.value.find(
            message =>
                message.clientID ===
                error.clientID
        );

    if (!msg) {
        return;
    }

    msg.sending = false;
    msg.failed = true;
    msg.error = error.message;
}

onMounted(() => {
    window.addEventListener(
        'chat-message',
        receiveMessage
    );

    window.addEventListener(
        'message-send-error',
        handleMessageSendError
    );
});

onUnmounted(() => {
    stopTyping();

    window.removeEventListener(
        'chat-message',
        receiveMessage
    );

    window.removeEventListener(
        'message-send-error',
        handleMessageSendError
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

                <button
                    type="button"
                    class="chats-burger"
                    :class="{ open: chatsSidebarOpen }"
                    :aria-expanded="chatsSidebarOpen"
                    aria-controls="chat-drawer"
                    aria-label="Toggle chats list"
                    @click="toggleChatsSidebar"
                >
                    <span></span>
                    <span></span>
                    <span></span>
                </button>

                <div class="avatar">

                    <img
                        style="cursor: pointer;"
                        v-if="chat.Avatar"
                        :src="`/uploads/${chat.Avatar}`"
                        alt=""
                        @click="sendToProfile"
                    />

                </div>

                <div class="chat-user-info">

                    <strong>
                        {{ chat.FirstName }}
                        {{ chat.LastName }}
                    </strong>

                    <span
                        v-if="partnerTyping"
                        class="typing-status"
                    >
                        typing<span class="typing-dots"><i></i><i></i><i></i></span>
                    </span>

                </div>

            </header>

            <div
                ref="messagesContainer"
                class="messages"
            >

                <div
                    v-if="loadingMore"
                    class="loading-more"
                >

                    <div class="small-loader"></div>

                    <span>
                        Loading older messages...
                    </span>

                </div>

                <div
                    v-if="loading"
                    class="loading-state"
                >

                    <div class="loader"></div>

                    <p>
                        Loading messages...
                    </p>

                </div>

                <template v-else>

                    <div
                        v-for="(msg, index) in messages"
                        :key="msg.id ?? msg.clientID ?? index"
                        class="message"
                        :class="[
                            msg.sender?.id === -1
                                ? 'sent'
                                : 'received',

                            msg.invite
                                ? 'invite-message'
                                : '',

                            msg.post
                                ? 'post-message'
                                : '',

                            msg.profile
                                ? 'profile-message'
                                : ''
                        ]"
                    >

                        <div
                            v-if="msg.sender?.id !== -1"
                            class="message-avatar"
                        >

                            <img
                                v-if="msg.sender?.avatar"
                                :src="`/uploads/${msg.sender.avatar}`"
                                alt=""
                            />

                            <span v-else>
                                {{ msg.sender?.firstName?.[0] }}
                                {{ msg.sender?.lastName?.[0] }}
                            </span>

                        </div>

                        <div
                            v-if="msg.invite"
                            class="invite-card"
                        >

                            <div class="invite-group">

                                <div class="invite-group-avatar">

                                    <img
                                        v-if="msg.invite.group.avatar"
                                        :src="`/uploads/${msg.invite.group.avatar}`"
                                        alt=""
                                    />

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

                                    <img
                                        v-if="msg.invite.user.avatar"
                                        :src="`/uploads/${msg.invite.user.avatar}`"
                                        alt=""
                                    />

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

                            <div
                                v-if="
                                    msg.inviteStatus === null ||
                                    msg.inviteStatus === undefined
                                "
                                class="invite-actions"
                            >

                                <button
                                    type="button"
                                    class="invite-accept"
                                    @click="respondToInvite(msg, 1)"
                                >
                                    Accept
                                </button>

                                <button
                                    type="button"
                                    class="invite-reject"
                                    @click="respondToInvite(msg, -1)"
                                >
                                    Reject
                                </button>

                            </div>

                            <div
                                v-else
                                class="invite-result"
                            >

                                <span v-if="msg.inviteStatus === 1">
                                    Invite accepted
                                </span>

                                <span v-else>
                                    Invite rejected
                                </span>

                            </div>

                        </div>

                        <div
                            v-else-if="msg.profile"
                            class="shared-profile-wrapper"
                        >

                            <div class="shared-post-label">
                                Shared profile
                            </div>

                            <button
                                type="button"
                                class="shared-profile-card"
                                @click="openSharedProfile(msg.profile)"
                            >

                                <span class="shared-profile-avatar">

                                    <img
                                        v-if="msg.profile.avatar"
                                        :src="`/uploads/${msg.profile.avatar}`"
                                        alt=""
                                    />

                                    <span v-else>
                                        {{ msg.profile.firstName?.[0] }}
                                        {{ msg.profile.lastName?.[0] }}
                                    </span>

                                </span>

                                <span class="shared-profile-info">

                                    <span class="shared-profile-name">
                                        {{ msg.profile.firstName }}
                                        {{ msg.profile.lastName }}
                                    </span>

                                    <span
                                        v-if="msg.profile.username"
                                        class="shared-profile-username"
                                    >
                                        {{ msg.profile.username }}
                                    </span>

                                </span>

                                <span class="shared-profile-action">
                                    View profile
                                </span>

                            </button>

                        </div>

                        <div
                            v-else-if="msg.post"
                            class="shared-post-wrapper"
                        >

                            <div class="shared-post-label">
                                Shared post
                            </div>

                            <HomePosts
                                :auto="false"
                                :current-user-id="msg.post.currentUserId"
                                :allow-comments="msg.post.allowComments"
                                :reaction="msg.post.reaction"
                                :user-id="msg.post.userId"
                                :post-id="msg.post.postId"
                                :group-id="msg.post.groupId"
                                :first-name="msg.post.firstName"
                                :last-name="msg.post.lastName"
                                :username="msg.post.username"
                                :avatar-path="msg.post.avatarPath"
                                :created-at="msg.post.createdAt"
                                :content="msg.post.content"
                                :image-path="msg.post.imagePath"
                                :location="msg.post.location"
                                :tagged-people="msg.post.taggedPeople"
                                :likes="msg.post.likes"
                                :dislikes="msg.post.dislikes"
                                :comments="msg.post.comments"
                                :user-reaction="msg.post.userReaction"
                                :relationship="msg.post.relationship"
                                :visibility="msg.post.visibility"
                                :visibility-user="msg.post.visibilityUser"
                            />

                            <span
                                v-if="msg.failed"
                                class="message-error"
                            >
                                {{ msg.error }}
                            </span>

                        </div>

                        <div
                            v-else
                            class="message-content"
                        >

                            <div
                                class="message-body"
                                :class="{
                                    'media-body': msg.media
                                }"
                            >

                                <span
                                    v-if="msg.sender?.id !== -1"
                                    class="message-sender-name"
                                >
                                    {{ msg.sender?.firstName }}
                                    {{ msg.sender?.lastName }}
                                </span>

                                <img
                                    v-if="msg.media"
                                    class="message-image"
                                    :src="`/uploads/${msg.media.path}`"
                                    alt=""
                                    @click="openImage(msg.media.path)"
                                />

                                <p v-else>
                                    {{ msg.content }}
                                </p>

                            </div>

                            <span
                                v-if="msg.failed"
                                class="message-error"
                            >
                                {{ msg.error }}
                            </span>

                        </div>

                    </div>

                    <div
                        v-if="messages.length === 0"
                        class="no-messages"
                    >
                        <p>
                            No messages yet
                        </p>
                    </div>

                </template>

            </div>

            <div
                v-if="pendingPreview"
                class="pending-media"
            >

                <div class="pending-media-item">

                    <img
                        :src="pendingPreview"
                        alt=""
                    />

                    <button
                        type="button"
                        class="pending-media-remove"
                        title="Remove image"
                        @click="clearPending"
                    >
                        ×
                    </button>

                </div>

            </div>

            <form
                class="composer"
                @submit.prevent="send"
            >

                <input
                    ref="fileInput"
                    type="file"
                    :accept="CHAT_MEDIA_ACCEPT"
                    hidden
                    @change="onFileChange"
                />

                <button
                    type="button"
                    class="attach-trigger"
                    title="Send image"
                    :disabled="
                        sending ||
                        loading ||
                        !canMessage
                    "
                    @click="pickFile"
                >
                    + Image
                </button>

                <EmojiPicker
                    :disabled="
                        sending ||
                        loading ||
                        !canMessage
                    "
                    @select="insertEmoji"
                />

                <input
                    ref="messageInput"
                    v-model="message"
                    type="text"
                    :placeholder="
                        canMessage
                            ? 'Type a message...'
                            : 'You cannot text this user because of user preferences'
                    "
                    :disabled="
                        loading ||
                        !canMessage
                    "
                />

                <button
                    type="submit"
                    :disabled="
                        sending ||
                        loading ||
                        !canMessage
                    "
                >
                    {{ sending ? 'Sending...' : 'Send' }}
                </button>

            </form>

        </template>

        <div
            v-else
            class="empty-state"
        >

            <button
                type="button"
                class="chats-burger empty-burger"
                aria-controls="chat-drawer"
                aria-label="Open chats list"
                @click="toggleChatsSidebar"
            >
                <span></span>
                <span></span>
                <span></span>
            </button>

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

        <div
            v-if="lightboxSrc"
            class="lightbox"
            @click="closeImage"
        >
            <img
                :src="lightboxSrc"
                alt=""
            />
        </div>

    </section>

</template>

<style scoped>

.message-image {
    display: block;
    max-width: 260px;
    max-height: 300px;
    width: auto;
    height: auto;
    border-radius: 6px;
    object-fit: cover;
    cursor: zoom-in;
}

.message-body.media-body {
    padding: 5px;
}

.attach-trigger {
    flex-shrink: 0;
    padding: 0 14px;
    height: 42px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    box-shadow: 4px 4px var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    white-space: nowrap;
    cursor: pointer;
}

.attach-trigger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.attach-trigger:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.pending-media {
    flex-shrink: 0;
    padding: 12px 20px 0;
    border-top: 2px solid var(--page-background);
}

.pending-media-item {
    position: relative;
    display: inline-block;
}

.pending-media-item img {
    display: block;
    max-width: 120px;
    max-height: 120px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    object-fit: cover;
}

.pending-media-remove {
    position: absolute;
    top: -8px;
    right: -8px;
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    color: white;
    font-size: 14px;
    line-height: 1;
    cursor: pointer;
}

.lightbox {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(0, 0, 0, 0.8);
    cursor: zoom-out;
}

.lightbox img {
    max-width: 100%;
    max-height: 100%;
    border-radius: 6px;
    object-fit: contain;
}

.message-content {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
}

.message-error {
    margin-top: 4px;
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    color: var(--input-focus);
}

.chat-window {
    position: relative;
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    height: var(--chat-height, calc(100dvh - 64px - 40px));
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    overflow: hidden;
}

.chat-window-header {
    position: sticky;
    top: 0;
    z-index: 5;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 13px;
    padding: 16px 20px;
    border-bottom: 2px solid var(--page-background);
    background: var(--bg-color);
}

.chats-burger {
    display: none;
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    padding: 0;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 5px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    box-shadow: 3px 3px var(--main-color);
    cursor: pointer;
}

.chats-burger span {
    display: block;
    width: 18px;
    height: 2px;
    border-radius: 2px;
    background: var(--main-color);
    transition: transform 0.25s ease, opacity 0.2s ease;
}

.chats-burger.open span:nth-child(1) {
    transform: translateY(7px) rotate(45deg);
}

.chats-burger.open span:nth-child(2) {
    opacity: 0;
}

.chats-burger.open span:nth-child(3) {
    transform: translateY(-7px) rotate(-45deg);
}

.chats-burger:active {
    transform: translate(3px, 3px);
    box-shadow: none;
}

.empty-burger {
    margin-bottom: 20px;
}

.chat-user-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
}

.chat-user-info strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.chat-window-header strong {
    display: block;
    font-size: 14px;
}

.typing-status {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 2px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-style: italic;
}

.typing-dots {
    display: inline-flex;
    gap: 2px;
}

.typing-dots i {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: currentColor;
    animation: typing-bounce 1s infinite ease-in-out;
}

.typing-dots i:nth-child(2) {
    animation-delay: 0.15s;
}

.typing-dots i:nth-child(3) {
    animation-delay: 0.3s;
}

@keyframes typing-bounce {
    0%, 60%, 100% {
        opacity: 0.3;
        transform: translateY(0);
    }

    30% {
        opacity: 1;
        transform: translateY(-3px);
    }
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

.invite-from > div:last-child {
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
    transition:
        transform 0.1s ease,
        box-shadow 0.1s ease;
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

.shared-post-wrapper {
    width: min(520px, 100%);
    box-sizing: border-box;
}

.message.sent .shared-post-wrapper {
    margin-left: auto;
}

.message.received .shared-post-wrapper {
    margin-right: auto;
}

.shared-post-label {
    margin-bottom: 7px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.post-message {
    max-width: 70%;
}

.profile-message {
    max-width: 70%;
}

.shared-profile-wrapper {
    width: min(360px, 100%);
    box-sizing: border-box;
}

.message.sent .shared-profile-wrapper {
    margin-left: auto;
}

.message.received .shared-profile-wrapper {
    margin-right: auto;
}

.shared-profile-card {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
    box-sizing: border-box;
    padding: 12px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--font-color);
    font: inherit;
    text-align: left;
    cursor: pointer;
    transition: transform 0.1s, box-shadow 0.1s;
}

.shared-profile-card:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.shared-profile-card:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.shared-profile-avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 700;
}

.shared-profile-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.shared-profile-info {
    min-width: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 3px;
}

.shared-profile-name {
    overflow-wrap: anywhere;
    font-size: 13px;
    font-weight: 700;
}

.shared-profile-username {
    overflow-wrap: anywhere;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.shared-profile-action {
    flex-shrink: 0;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
}

.post-message .shared-post-wrapper {
    min-width: 0;
}

.post-message :deep(.home-post) {
    width: 100%;
}

.post-message :deep(.post-card) {
    max-width: 100%;
}

.composer {
    position: relative;
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

@media (max-width: 1024px) {

    .chats-burger {
        display: flex;
    }

    .chat-window-header {
        padding: 12px 14px;
    }
}

@media (max-width: 800px) {

    .message-image {
        max-width: 100%;
    }

    .composer {
        flex-wrap: wrap;
        padding: 12px 14px;
    }

    .composer input[type="text"] {
        flex: 1 1 100%;
        order: -1;
    }

    .messages {
        padding: 14px;
    }

    .message {
        max-width: 80%;
    }

    .invite-message {
        max-width: 90%;
    }

    .post-message {
        max-width: 90%;
    }

}

</style>