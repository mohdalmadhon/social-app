<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { addNotification } from '@/data/notifications';
import { sendWS } from '@/api/socket/socket';
import { getMessages, sendChatMedia } from '@/api/chats/chats';
import { CHAT_MEDIA_ACCEPT, parseChatMedia, validateChatMedia } from '@/helpers/chatMedia';
import { getGroupPost, insertPostReaction } from '@/api/posts/groups';
import { activePage } from '@/data/chatState';
import HomePosts from '@/components/home/HomePosts.vue';
import GroupEventDialog from '@/components/groups/GroupEventDialog.vue';

const props = defineProps({
    groupID: {
        type: Number,
        required: true
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
    },
    group: {
        type: Object,
        default: null
    }
});

const emit = defineEmits(['event-created']);

const message = ref('');
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
const selectedPost = ref(null);
const showPostDialog = ref(false);
const loadingPost = ref(false);
const reacting = ref(false);
const showEventDialog = ref(false);

let fetchTimer = null;
let requestID = 0;
let postRequestID = 0;

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

function parsePostContent(content) {
    if (typeof content !== 'string') {
        return null;
    }

    try {
        const parsed = JSON.parse(content);

        if (
            parsed &&
            typeof parsed === 'object' &&
            parsed.postID !== undefined
        ) {
            return parsed;
        }
    } catch (error) {
        return null;
    }

    return null;
}

function truncatePostContent(content) {
    if (!content) {
        return '';
    }

    const words = String(content).trim().split(/\s+/);

    if (words.length <= 200) {
        return String(content).trim();
    }

    return words.slice(0, 200).join(' ') + '......';
}

function formatPost(post) {
    if (!post) {
        return null;
    }

    const user = post.user || {};

    return {
        postId: Number(post.postID),
        userId: Number(user.ID ?? user.id ?? 0),
        firstName: user.firstName ?? user.FirstName ?? '',
        lastName: user.lastName ?? user.LastName ?? '',
        avatarPath: user.avatar ?? user.Avatar ?? '',
        content: truncatePostContent(post.content),
        imagePath: post.imagePath ?? '',
        groupId: Number(post.groupID ?? 0),
        createdAt: post.createdAt ?? post.CreatedAt ?? '',
        reaction: 0,
        likes: 0,
        dislikes: 0,
        comments: [],
        taggedPeople: [],
        userReaction: '',
        relationship: 'none',
        visibility: 'public',
        visibilityUser: ''
    };
}

function reactionValueToLabel(value) {
    if (value === 1) {
        return 'like';
    }

    if (value === -1) {
        return 'dislike';
    }

    return '';
}

function formatFullPost(data, fallbackGroupId) {
    if (!data) {
        return null;
    }

    const reactionValue = data.ReactionValue ?? 0;

    return {
        postId: Number(data.id),
        userId: Number(data.userId),
        firstName: data.firstName ?? '',
        lastName: data.lastName ?? '',
        username: data.username ?? '',
        avatarPath: data.avatarPath ?? '',
        content: data.content ?? '',
        imagePath: data.imagePath ?? '',
        location: data.location ?? '',
        groupId: data.groupId ?? fallbackGroupId,
        createdAt: data.createdAt ?? '',
        relationship: data.relationship ?? 'none',
        visibility: data.visibility ?? 'public',
        visibilityUser: data.visibilityUser ?? '',
        taggedPeople: data.taggedPeople ?? [],
        likes: data.likeCount ?? 0,
        dislikes: data.disLikeCount ?? 0,
        comments: Array.from({ length: data.commentCount ?? 0 }),
        reaction: reactionValue,
        userReaction: reactionValueToLabel(reactionValue)
    };
}

function formatMessage(msg) {
    const rawContent =
        msg.Content ??
        msg.content ??
        '';

    const postData = parsePostContent(rawContent);

    const sender = {
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
    };

    if (postData) {
        const post = formatPost(postData);

        return {
            id: msg.ID ?? msg.id,
            content: rawContent,
            createdAt:
                msg.CreatedAt ??
                msg.createdAt,
            groupID:
                msg.GroupID ??
                msg.groupID,
            sender: {
                id: post.userId || sender.id,
                firstName:
                    post.firstName ||
                    sender.firstName,
                lastName:
                    post.lastName ||
                    sender.lastName,
                avatar:
                    post.avatarPath ||
                    sender.avatar
            },
            isPost: true,
            post
        };
    }

    return {
        id: msg.ID ?? msg.id,
        content: rawContent,
        createdAt:
            msg.CreatedAt ??
            msg.createdAt,
        groupID:
            msg.GroupID ??
            msg.groupID,
        sender,
        media: parseChatMedia(rawContent),
        isPost: false,
        post: null
    };
}

function isOwnMessage(msg) {
    const userID = Number(props.userID);

    if (msg.isPost && msg.post) {
        return Number(msg.post.userId) === userID;
    }

    const senderID = Number(msg.sender?.id);

    return senderID === -1 || senderID === userID;
}

async function openPost(post) {
    if (!post) {
        return;
    }

    const currentPostRequestID = ++postRequestID;

    selectedPost.value = post;
    showPostDialog.value = true;
    loadingPost.value = true;

    try {
        const data = await getGroupPost(
            post.groupId || props.groupID,
            post.postId
        );

        if (currentPostRequestID !== postRequestID) {
            return;
        }

        selectedPost.value = formatFullPost(data, props.groupID);
    } catch (err) {
        if (currentPostRequestID !== postRequestID) {
            return;
        }

        addNotification(
            err.message ||
            'Could not load post',
            'error'
        );

        closePost();
    } finally {
        if (currentPostRequestID === postRequestID) {
            loadingPost.value = false;
        }
    }
}

function closePost() {
    postRequestID++;
    showPostDialog.value = false;
    selectedPost.value = null;
    loadingPost.value = false;
}

async function handleReaction(value) {
    if (!selectedPost.value || reacting.value) {
        return;
    }

    const post = selectedPost.value;
    const previousReaction = post.reaction;
    const previousLikes = post.likes;
    const previousDislikes = post.dislikes;

    let nextReaction = value;
    let likes = post.likes;
    let dislikes = post.dislikes;

    if (previousReaction === 1) {
        likes -= 1;
    } else if (previousReaction === -1) {
        dislikes -= 1;
    }

    if (previousReaction === value) {
        nextReaction = 0;
    } else if (value === 1) {
        likes += 1;
    } else if (value === -1) {
        dislikes += 1;
    }

    selectedPost.value = {
        ...post,
        reaction: nextReaction,
        userReaction: reactionValueToLabel(nextReaction),
        likes,
        dislikes
    };

    reacting.value = true;

    try {
        await insertPostReaction(post.postId, value);
    } catch (err) {
        selectedPost.value = {
            ...selectedPost.value,
            reaction: previousReaction,
            userReaction: reactionValueToLabel(previousReaction),
            likes: previousLikes,
            dislikes: previousDislikes
        };

        addNotification(
            err.message ||
            'Could not update reaction',
            'error'
        );
    } finally {
        reacting.value = false;
    }
}

function handleLike() {
    handleReaction(1);
}

function handleDislike() {
    handleReaction(-1);
}

async function scrollToBottom() {
    await nextTick();

    if (messagesContainer.value) {
        messagesContainer.value.scrollTop =
            messagesContainer.value.scrollHeight;
    }
}

async function fetchChatMessages(groupID) {
    if (
        groupID === null ||
        groupID === undefined
    ) {
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
        const result = await getMessages(
            groupID,
            0
        );

        if (currentRequestID !== requestID) {
            return;
        }

        const data = Array.isArray(result)
            ? result
            : result.messages ||
            result.data ||
            [];

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
            offset.value
        );

        if (currentRequestID !== requestID) {
            return;
        }

        const data = Array.isArray(result)
            ? result
            : result.messages ||
            result.data ||
            [];

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
            (
                container.scrollHeight -
                oldScrollHeight
            );
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

function receiveMessage(event) {
    const incoming = event.detail;

    if (!incoming) {
        return;
    }

    const incomingGroupID =
        incoming.GroupID ??
        incoming.groupID;

    if (
        Number(incomingGroupID) !==
        Number(props.groupID)
    ) {
        return;
    }

    const formatted =
        formatMessage(incoming);

    if (
        formatted.id &&
        messages.value.some(
            msg => msg.id === formatted.id
        )
    ) {
        return;
    }

    messages.value.push(formatted);

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

        const ownMessage = isOwnMessage(formatted);

        if (
            ownMessage ||
            distanceFromBottom < 150
        ) {
            container.scrollTop =
                container.scrollHeight;
        }
    });
}

async function send() {
    const content =
        message.value.trim();
    const file = pendingFile.value;

    if (
        (!content && !file) ||
        sending.value ||
        !props.groupID
    ) {
        return;
    }

    sending.value = true;

    try {
        if (file) {
            const result = await sendChatMedia(file, {
                userID: -1,
                groupID: props.groupID
            });

            messages.value.push({
                id: `local-${Date.now()}-media`,
                content: result.content,
                createdAt:
                    new Date().toISOString(),
                groupID: props.groupID,
                sender: {
                    id: -1,
                    firstName:
                        props.userFirstName,
                    lastName:
                        props.userLastName,
                    avatar:
                        props.userAvatar
                },
                media: parseChatMedia(result.content),
                isPost: false,
                post: null
            });

            clearPending();

            scrollToBottom();
        }

        if (content) {
            sendWS({
                type: 'privateMessage',
                data: {
                    userID: props.userID,
                    groupID: props.groupID,
                    content,
                    private: 1
                }
            });

            messages.value.push({
                id: `local-${Date.now()}`,
                content,
                createdAt:
                    new Date().toISOString(),
                groupID: props.groupID,
                sender: {
                    id: -1,
                    firstName:
                        props.userFirstName,
                    lastName:
                        props.userLastName,
                    avatar:
                        props.userAvatar
                },
                media: null,
                isPost: false,
                post: null
            });

            message.value = '';

            scrollToBottom();
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

function openEventDialog() {
    showEventDialog.value = true;
}

function closeEventDialog() {
    showEventDialog.value = false;
}

function eventCreated(event) {
    showEventDialog.value = false;
    emit('event-created', event);
}

function handleKeydown(event) {
    if (
        event.key === 'Enter' &&
        !event.shiftKey
    ) {
        event.preventDefault();
        send();
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

        closePost();

        if (
            newGroupID === null ||
            newGroupID === undefined
        ) {
            loading.value = false;
            return;
        }

        activePage.value =
            'group:' + newGroupID;

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

onMounted(() => {
    activePage.value =
        'group:' + props.groupID;

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
    postRequestID++;
});
</script>

<template>
    <section class="chat-window">
        <header v-if="group" class="chat-window-header">
            <div class="avatar">
                <img v-if="
                    group.avatar ||
                    group.Avatar
                " :src="`/uploads/${group.avatar ||
                    group.Avatar
                    }`
                    " alt="" />
            </div>

            <div class="chat-user-info">
                <strong>
                    {{
                        group.name ||
                        group.Name
                    }}
                </strong>
            </div>
        </header>

        <div ref="messagesContainer" class="messages">
            <div v-if="loadingMore" class="loading-more">
                <div class="small-loader"></div>

                <span>
                    Loading older messages...
                </span>
            </div>

            <div v-if="loading" class="loading-state">
                <div class="loader"></div>

                <p>
                    Loading messages...
                </p>
            </div>

            <template v-else>
                <div v-for="(msg, index) in messages" :key="msg.id ?? index" class="message" :class="msg.isPost
                    ? 'post-centered'
                    : (isOwnMessage(msg) ? 'sent' : 'received')
                    ">
                    <div v-if="
                        !msg.isPost && !isOwnMessage(msg)
                    " class="message-avatar">
                        <img v-if="
                            msg.sender?.avatar
                        " :src="`/uploads/${msg.sender.avatar}`
            " alt="" />

                        <span v-else>
                            {{
                                msg.sender
                                    ?.firstName?.[0]
                            }}
                            {{
                                msg.sender
                                    ?.lastName?.[0]
                            }}
                        </span>
                    </div>

                    <div class="message-body" :class="{ 'media-body': msg.media }">
                        <span v-if="
                            !msg.isPost && !isOwnMessage(msg)
                        " class="message-sender-name">
                            {{
                                msg.sender?.firstName
                            }}
                            {{
                                msg.sender?.lastName
                            }}
                        </span>

                        <template v-if="msg.isPost && msg.post">
                            <button type="button" class="post-message" :class="{ 'no-image': !msg.post.imagePath }"
                                @click="openPost(msg.post)">
                                <div class="post-message-header">
                                    <span class="post-badge">
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                            <rect x="3" y="4" width="18" height="14" rx="2" />
                                            <path d="M3 8h18M7 12h6" />
                                        </svg>
                                    </span>

                                    <div class="post-header-text">
                                        <strong>
                                            {{
                                                msg.post
                                                    .firstName
                                            }}
                                            {{
                                                msg.post
                                                    .lastName
                                            }}
                                        </strong>

                                        <span>
                                            shared a post
                                        </span>
                                    </div>
                                </div>

                                <div v-if="
                                    msg.post.content
                                " class="post-message-content">
                                    {{
                                        msg.post.content
                                    }}
                                </div>

                                <div v-if="
                                    msg.post.imagePath
                                " class="post-message-image">
                                    <img :src="`/uploads/${msg.post.imagePath}`
                                        " alt="" />
                                </div>

                                <div class="post-message-footer">
                                    <span>Click to view post</span>
                                    <span>→</span>
                                </div>
                            </button>
                        </template>

                        <img v-else-if="msg.media" class="message-image" :src="`/uploads/${msg.media.path}`"
                            alt="" @click="openImage(msg.media.path)" />

                        <p v-else>
                            {{ msg.content }}
                        </p>
                    </div>
                </div>

                <div v-if="messages.length === 0" class="no-messages">
                    <p>
                        No messages yet
                    </p>
                </div>
            </template>
        </div>

        <div v-if="pendingPreview" class="pending-media">
            <div class="pending-media-item">
                <img :src="pendingPreview" alt="" />

                <button type="button" class="pending-media-remove" title="Remove image" @click="clearPending">
                    ×
                </button>
            </div>
        </div>

        <form class="composer" @submit.prevent="send">
            <input ref="fileInput" type="file" :accept="CHAT_MEDIA_ACCEPT" hidden @change="onFileChange" />

            <button type="button" class="event-trigger" title="Create event" @click="openEventDialog">
                + Event
            </button>

            <button type="button" class="event-trigger" title="Send image" :disabled="sending || loading"
                @click="pickFile">
                + Image
            </button>

            <input v-model="message" type="text" placeholder="Type a message..." :disabled="loading"
                @keydown="handleKeydown" />

            <button type="submit" :disabled="sending ||
                loading ||
                (!message.trim() && !pendingFile)
                ">
                {{
                    sending
                        ? 'Sending...'
                        : 'Send'
                }}
            </button>
        </form>

        <div v-if="lightboxSrc" class="lightbox" @click="closeImage">
            <img :src="lightboxSrc" alt="" />
        </div>

        <GroupEventDialog :show="showEventDialog" :group-id="groupID" @close="closeEventDialog"
            @created="eventCreated" />

        <div v-if="showPostDialog" class="post-dialog-overlay" @click.self="closePost">
            <div class="post-dialog">
                <button type="button" class="post-dialog-close" @click="closePost">
                    ×
                </button>

                <div v-if="loadingPost" class="post-dialog-loading">
                    <div class="loader"></div>
                    <p>Loading post...</p>
                </div>

                <HomePosts v-else-if="selectedPost" :key="selectedPost.postId"
                    :current-user-id="userID"
                    :post-id="selectedPost.postId"
                    :user-id="selectedPost.userId"
                    :first-name="selectedPost.firstName"
                    :last-name="selectedPost.lastName"
                    :username="selectedPost.username"
                    :avatar-path="selectedPost.avatarPath"
                    :group-id="selectedPost.groupId"
                    :content="selectedPost.content"
                    :image-path="selectedPost.imagePath"
                    :location="selectedPost.location"
                    :created-at="selectedPost.createdAt"
                    :reaction="selectedPost.reaction"
                    :likes="selectedPost.likes"
                    :dislikes="selectedPost.dislikes"
                    :comments="selectedPost.comments"
                    :tagged-people="selectedPost.taggedPeople"
                    :user-reaction="selectedPost.userReaction"
                    :relationship="selectedPost.relationship"
                    :visibility="selectedPost.visibility"
                    :visibility-user="selectedPost.visibilityUser"
                    @like="handleLike"
                    @dislike="handleDislike"
                />
            </div>
        </div>
    </section>
</template>

<style scoped>
.chat-window {
    width: 100%;
    height: calc(100vh - 100px);
    min-height: 0;
    display: flex;
    flex-direction: column;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    overflow: hidden;
    box-sizing: border-box;
}

.message.post-centered {
    align-self: center;
    max-width: 100%;
    justify-content: center;
}

.message.post-centered .message-body {
    padding: 0;
    border: none;
    box-shadow: none;
    background: transparent;
}

.chat-window-header {
    flex: 0 0 auto;
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
    flex: 1 1 0;
    min-height: 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    overflow-y: auto;
    overflow-x: hidden;
    background: var(--page-background);
    box-sizing: border-box;
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
    flex-shrink: 0;
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
    flex-shrink: 0;
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
    min-width: 0;
    padding: 11px 15px;
    border: 2px solid var(--main-color);
    border-radius: 10px;
    overflow-wrap: anywhere;
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

.message.received {
    align-self: flex-start;
}

.message.received .message-body {
    background: var(--input-focus);
    color: white;
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

.post-message {
    display: block;
    width: 320px;
    max-width: 100%;
    padding: 0;
    border: 1px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
    overflow: hidden;
    box-shadow: 3px 3px var(--main-color);
    transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.post-message:hover {
    transform: translateY(-2px);
    box-shadow: 5px 5px var(--main-color);
}

.post-message:active {
    transform: translateY(0);
    box-shadow: 2px 2px var(--main-color);
}

.post-message-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px 8px;
    border-bottom: 1px solid var(--main-color);
}

.post-message-header .post-badge {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--input-focus);
    color: white;
}

.post-message-header .post-badge svg {
    width: 12px;
    height: 12px;
}

.post-message-header .post-header-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.3;
}

.post-message-header strong {
    font-size: 11px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.post-message-header span {
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    opacity: 0.65;
}

.post-message-content {
    position: relative;
    padding: 10px 12px 6px;
    font-size: 11.5px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 100px;
    overflow: hidden;
}

.post-message:not(.no-image) .post-message-content::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 28px;
    background: linear-gradient(to bottom, transparent, var(--bg-color));
    pointer-events: none;
}

.post-message-image {
    width: 100%;
    max-height: 160px;
    overflow: hidden;
    border-top: 1px solid var(--main-color);
}

.post-message-image img {
    display: block;
    width: 100%;
    max-height: 160px;
    object-fit: cover;
    transition: transform 0.25s ease;
}

.post-message:hover .post-message-image img {
    transform: scale(1.03);
}

.post-message-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-top: 1px solid var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    opacity: 0.75;
    transition: opacity 0.15s ease;
}

.post-message:hover .post-message-footer {
    opacity: 1;
    color: var(--input-focus);
}

.post-dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 30px;
    background: rgba(0, 0, 0, 0.7);
    overflow-y: auto;
}

.post-dialog {
    position: relative;
    width: 100%;
    max-width: 720px;
    max-height: calc(100vh - 60px);
    overflow-y: auto;
}

.post-dialog-close {
    position: absolute;
    top: -14px;
    right: -14px;
    z-index: 10;
    width: 34px;
    height: 34px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--font-color);
    font-size: 22px;
    line-height: 26px;
    cursor: pointer;
    box-shadow: 3px 3px var(--main-color);
}

.post-dialog-close:hover {
    transform: translateY(-1px);
}

.post-dialog-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    min-height: 200px;
    background: var(--bg-color);
    border-radius: 8px;
}

.post-dialog-loading p {
    margin: 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.composer {
    flex: 0 0 auto;
    display: flex;
    gap: 10px;
    padding: 14px 20px;
    border-top: 2px solid var(--page-background);
}

.composer input {
    flex: 1;
    min-width: 0;
    height: 42px;
    padding: 0 14px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    box-sizing: border-box;
}

.composer input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.composer button {
    flex-shrink: 0;
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

.event-trigger {
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
}

.event-trigger:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.event-trigger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

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

.pending-media {
    flex: 0 0 auto;
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

@media (max-width: 800px) {
    .chat-window {
        height: calc(100vh - 220px);
        min-height: 380px;
        max-height: none;
    }

    .message {
        max-width: 80%;
    }

    .post-message {
        width: 280px;
    }

    .post-dialog-overlay {
        padding: 15px;
    }

    .post-dialog {
        max-height: calc(100vh - 30px);
    }

    .post-dialog-close {
        top: -8px;
        right: -8px;
    }
}

@media (max-width: 560px) {
    .chat-window {
        height: calc(100vh - 200px);
        min-height: 340px;
        box-shadow: 4px 4px var(--main-color);
    }

    .messages {
        padding: 14px;
    }

    .message {
        max-width: 92%;
    }

    .post-message {
        width: 100%;
    }

    .composer {
        flex-wrap: wrap;
        padding: 12px 14px;
    }

    .composer input {
        flex: 1 1 100%;
        order: 1;
    }

    .event-trigger {
        order: 2;
        flex: 1 1 auto;
    }

    .message-image {
        max-width: 100%;
    }

    .composer button[type="submit"] {
        order: 3;
        flex: 1 1 auto;
    }

    .post-dialog-overlay {
        padding: 10px;
    }

    .post-dialog-close {
        top: 4px;
        right: 4px;
    }
}
</style>