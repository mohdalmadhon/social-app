<script setup>
import { LIMITS } from '@/helpers/limits';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

import { postReaction } from '@/api/posts/actions';
import { searchShares } from '@/api/search/search';
import { addNotification } from '@/data/notifications';
import { Reaction } from '@/models/posts';

const props = defineProps({
    allowComments: {
        type: Boolean
    },

    reaction: {
        type: Number,
        default: 0
    },

    postId: {
        type: Number,
        required: true
    },

    likes: {
        type: Number,
        default: 0
    },

    dislikes: {
        type: Number,
        default: 0
    }
});

const emit = defineEmits(['like', 'dislike', 'toggle-comments']);

const MAX_SHARE_USERS = 30;

const SEARCH_DELAY = 300;

const likeCount = ref(props.likes);

const dislikeCount = ref(props.dislikes);

const currentReaction = ref(props.reaction);

const reactionAnimation = ref('');

const showShare = ref(false);

const loadingUsers = ref(false);

const shareUsers = ref([]);

const selectedUserIds = ref([]);

const sharing = ref(false);

const searchQuery = ref('');

const searchInput = ref(null);

const hasSelection = computed(() => selectedUserIds.value.length > 0);

let searchTimer = null;

let requestId = 0;

watch(() => props.likes, (value) => {
    likeCount.value = value;
});

watch(() => props.dislikes, (value) => {
    dislikeCount.value = value;
});

watch(() => props.reaction, (value) => {
    currentReaction.value = value;
});

watch(showShare, (open) => {
    if (open) {
        window.addEventListener('keydown', handleKeydown);
    } else {
        window.removeEventListener('keydown', handleKeydown);
    }
});

async function handleReaction(value) {
    const previousReaction = currentReaction.value;

    const newReaction = previousReaction === value ? 0 : value;

    const rect = new Reaction(value, props.postId);

    const result = await postReaction(rect.getData());

    if (!result.status) {
        addNotification(result.message);

        return;
    }

    currentReaction.value = newReaction;

    reactionAnimation.value = value === 1 ? 'like' : 'dislike';

    setTimeout(() => {
        reactionAnimation.value = '';
    }, 350);

    if (previousReaction === 1) {
        likeCount.value--;
    }

    if (previousReaction === -1) {
        dislikeCount.value--;
    }

    if (newReaction === 1) {
        likeCount.value++;
    }

    if (newReaction === -1) {
        dislikeCount.value++;
    }

    if (newReaction === 1) {
        emit('like');
    } else {
        emit('dislike');
    }
}

function toggleComments() {
    emit('toggle-comments');
}

async function loadUsers(query) {
    requestId++;

    const currentRequest = requestId;

    loadingUsers.value = true;

    const result = await searchShares(query, props.postId);

    if (currentRequest !== requestId) {
        return;
    }
    console.log(result)
    loadingUsers.value = false;

    if (!result.status) {
        addNotification(result.message || 'Could not load users');
        shareUsers.value = [];

        if (!query) {
            showShare.value = false;
        }

        return;
    }

    shareUsers.value = (result.data || []).slice(0, MAX_SHARE_USERS);
}

function isSelected(userId) {
    return selectedUserIds.value.includes(userId);
}

function toggleUser(userId) {
    if (sharing.value) {
        return;
    }

    if (isSelected(userId)) {
        selectedUserIds.value = selectedUserIds.value.filter((id) => id !== userId);

        return;
    }

    selectedUserIds.value = [...selectedUserIds.value, userId];
}

async function submitShare() {
    if (!hasSelection.value || sharing.value) {
        return;
    }

    sharing.value = true;

    try {
        const response = await fetch('/api/chats/share', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                userIds: selectedUserIds.value,
                postID: props.postId
            })
        });

        let result = null;

        try {
            result = await response.json();
        } catch {
            result = null;
        }

        if (!response.ok || (result && result.status === false)) {
            throw new Error(result?.message || 'Could not share post');
        }

        addNotification('Post shared', 'success');

        closeShare();
    } catch (err) {
        addNotification(err.message || 'Could not share post', 'error');
    } finally {
        sharing.value = false;
    }
}

async function handleShare() {
    showShare.value = true;
    searchQuery.value = '';
    shareUsers.value = [];
    selectedUserIds.value = [];

    await nextTick();

    if (searchInput.value) {
        searchInput.value.focus();
    }

    await loadUsers('');
}

function onSearchInput() {
    clearTimeout(searchTimer);

    searchTimer = setTimeout(() => {
        loadUsers(searchQuery.value.trim());
    }, SEARCH_DELAY);
}

function closeShare() {
    clearTimeout(searchTimer);
    requestId++;
    loadingUsers.value = false;
    showShare.value = false;
    selectedUserIds.value = [];
}

function handleKeydown(event) {
    if (event.key === 'Escape') {
        closeShare();
    }
}

onMounted(() => {
    console.log(props.likes)
})

onBeforeUnmount(() => {
    clearTimeout(searchTimer);
    window.removeEventListener('keydown', handleKeydown);
})
</script>

<template>
    <div class="post-actions">
        <button class="action-button" :class="{
            active: currentReaction === 1,
            'reaction-jump': reactionAnimation === 'like'
        }" type="button" @click="handleReaction(1)">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                    d="M7 10v10H3V10h4Zm3 10h7.2a2 2 0 0 0 1.9-1.4l2.3-7A2 2 0 0 0 19.5 9H15l.7-3.4A2.2 2.2 0 0 0 13.5 3L9 9v11h1Z" />
            </svg>

            Like {{ likeCount }}
        </button>

        <button class="action-button dislike" :class="{
            active: currentReaction === -1,
            'reaction-jump': reactionAnimation === 'dislike'
        }" type="button" @click="handleReaction(-1)">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M7 14V4H3v10h4Zm3-10h7.2a2 2 0 0 1 1.9 1.4l.7 3.4a2.2 2.2 0 0 1-2.2 2.6L9 15V4h1Z" />
            </svg>

            Dislike {{ dislikeCount }}
        </button>

        <button v-if="props.allowComments" class="action-button" type="button" @click="toggleComments">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 9.6 9.6 0 0 1-4-.9L3 21l1.5-4.2A8.5 8.5 0 0 1 21 11.5Z" />
            </svg>

            Comment
        </button>

        <button class="action-button" type="button" @click="handleShare">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7M16 6l-4-4-4 4M12 2v13" />
            </svg>

            Share
        </button>
    </div>

    <Teleport to="body">
        <div v-if="showShare" class="share-overlay" @click.self="closeShare">
            <div class="share-dialog" role="dialog" aria-modal="true" aria-label="Share post">
                <div class="share-header">
                    <h3>Share with</h3>

                    <button class="share-close" type="button" aria-label="Close" @click="closeShare">
                        <svg viewBox="0 0 24 24" aria-hidden="true">
                            <path d="M6 6l12 12M18 6L6 18" />
                        </svg>
                    </button>
                </div>

                <div class="share-search">
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14ZM21 21l-4.3-4.3" />
                    </svg>

                    <input
                        ref="searchInput"
                        v-model="searchQuery"
                        :maxlength="LIMITS.search"
                        type="text"
                        placeholder="Search users"
                        autocomplete="off"
                        @input="onSearchInput"
                    />
                </div>

                <p v-if="loadingUsers" class="share-state">Loading...</p>

                <p v-else-if="!shareUsers.length" class="share-state">No users found</p>

                <div v-else class="share-users">
                    <button
                        v-for="user in shareUsers"
                        :key="user.ID"
                        type="button"
                        class="share-user"
                        :class="{ selected: isSelected(user.ID) }"
                        :aria-pressed="isSelected(user.ID)"
                        @click="toggleUser(user.ID)"
                    >
                        <span class="share-avatar-wrap">
                            <img class="share-avatar" :src="`/uploads/${user.avatar}`" :alt="`${user.firstName} ${user.lastName}`" />

                            <span v-if="isSelected(user.ID)" class="share-check" aria-hidden="true">
                                <svg viewBox="0 0 24 24">
                                    <path d="M5 12.5l4.5 4.5L19 7.5" />
                                </svg>
                            </span>
                        </span>

                        <span class="share-name">{{ user.firstName }} {{ user.lastName }}</span>
                    </button>
                </div>

                <div v-if="hasSelection" class="share-footer">
                    <span class="share-count">{{ selectedUserIds.length }} selected</span>

                    <button class="share-submit" type="button" :disabled="sharing" @click="submitShare">
                        {{ sharing ? 'Sharing...' : 'Share' }}
                    </button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.post-actions {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(0, 1fr));
    padding: 7px;
}

.action-button {
    min-height: 42px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    border: 2px solid transparent;
    border-radius: 5px;
    background: transparent;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    transition:
        background 0.15s,
        color 0.15s,
        border 0.15s;
}

.action-button:hover {
    background: var(--page-background);
    color: var(--main-color);
}

.action-button svg {
    width: 18px;
    height: 18px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.8;
    stroke-linecap: round;
    stroke-linejoin: round;
}

.action-button.active {
    border-color: var(--main-color);
    background: var(--input-focus);
    color: white;
}

.action-button.dislike.active {
    background: var(--main-color);
    color: white;
}

.reaction-jump svg {
    animation: jump 0.35s ease;
}

@keyframes jump {
    0% {
        transform: translateY(0) scale(1);
    }

    40% {
        transform: translateY(-8px) scale(1.2);
    }

    70% {
        transform: translateY(2px) scale(0.95);
    }

    100% {
        transform: translateY(0) scale(1);
    }
}

.share-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgba(0, 0, 0, 0.45);
}

.share-dialog {
    width: 100%;
    max-width: 520px;
    padding: 18px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--font-color);
    box-shadow: 3px 3px var(--main-color);
}

.share-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
}

.share-header h3 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 20px;
    color: var(--font-color);
}

.share-close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    cursor: pointer;
}

.share-close svg {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
}

.share-search {
    position: relative;
    margin-bottom: 12px;
}

.share-search svg {
    position: absolute;
    top: 50%;
    left: 12px;
    width: 16px;
    height: 16px;
    transform: translateY(-50%);
    fill: none;
    stroke: var(--font-color-sub);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    pointer-events: none;
}

.share-search input {
    width: 100%;
    box-sizing: border-box;
    padding: 10px 12px 10px 36px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--page-background);
    color: var(--font-color);
    font-size: 13px;
    outline: none;
}

.share-search input:focus {
    border-color: var(--input-focus);
}

.share-state {
    margin: 0;
    padding: 24px 0;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.share-users {
    display: flex;
    gap: 14px;
    overflow-x: auto;
    padding: 4px 2px 12px;
    scroll-snap-type: x proximity;
}

.share-user {
    flex: 0 0 84px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 6px 0;
    border: 2px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font: inherit;
    cursor: pointer;
    scroll-snap-align: start;
    transition:
        background 0.15s,
        border-color 0.15s;
}

.share-user:hover {
    background: var(--page-background);
}

.share-user.selected {
    border-color: var(--input-focus);
    background: var(--page-background);
}

.share-avatar-wrap {
    position: relative;
    display: block;
    width: 60px;
    height: 60px;
}

.share-avatar {
    width: 60px;
    height: 60px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    object-fit: cover;
    background: var(--page-background);
}

.share-user.selected .share-avatar {
    border-color: var(--input-focus);
}

.share-check {
    position: absolute;
    right: -4px;
    bottom: -4px;
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    color: white;
}

.share-check svg {
    width: 12px;
    height: 12px;
    fill: none;
    stroke: currentColor;
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
}

.share-name {
    width: 100%;
    text-align: center;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.25;
    overflow-wrap: anywhere;
}

.share-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 6px;
    padding-top: 14px;
    border-top: 2px solid var(--page-background);
}

.share-count {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.share-submit {
    height: 40px;
    padding: 0 22px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
}

.share-submit:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.share-submit:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

@media (max-width: 650px) {
    .post-actions {
        padding: 5px;
    }

    .action-button {
        gap: 4px;
        font-size: 8px;
    }

    .action-button svg {
        width: 16px;
        height: 16px;
    }
}
</style>