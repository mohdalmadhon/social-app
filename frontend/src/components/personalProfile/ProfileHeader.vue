<script setup>
import { ref, computed, nextTick, watch, onBeforeUnmount } from 'vue';

import { requestFollow, shareProfile } from '@/api/users/profiles';
import { searchShareProfile } from '@/api/search/search';

import { useRoute, useRouter } from 'vue-router';

import { addNotification } from '@/data/notifications';

const route = useRoute();

const router = useRouter();

const props = defineProps({
    message: Boolean,
    addEdit: Boolean,
    userId: [Number, String],
    firstName: String,
    lastName: String,
    username: String,
    bio: String,
    avatarPath: String,
    numOfPosts: Number,
    numOfFollowing: Number,
    numOfFollowers: Number,
    isFollowing: Number,
    email: String,
    dob: String
});

const emit = defineEmits([
    'follow',
    'unfollow',
    'cancel-request'
]);

const MAX_SHARE_USERS = 30;

const SEARCH_DELAY = 300;

const followingStatus = ref(props.isFollowing);

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

watch(() => props.isFollowing, (value) => {
    followingStatus.value = value;
});

watch(showShare, (open) => {
    if (open) {
        window.addEventListener('keydown', handleKeydown);
    } else {
        window.removeEventListener('keydown', handleKeydown);
    }
});

function formatDOB(dob) {
    if (!dob) {
        return '';
    }

    const date = new Date(dob);

    if (Number.isNaN(date.getTime())) {
        return dob;
    }

    return date.toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        timeZone: 'UTC'
    });
}

function getProfileId() {
    const id = Number(route.query.id);

    if (!Number.isInteger(id) || id <= 0) {
        return -99;
    }

    return id;
}

function handleMessage() {
    const id = route.query.id;

    if (!id) {
        addNotification("could not open chat", 'error');
        return;
    }

    router.push({
        path: '/chats',
        query: {
            userId: id,
            firstName: props.firstName || '',
            lastName: props.lastName || '',
            avatar: (props.avatarPath || '').replace(/^\/uploads\//, '')
        }
    });
}

async function loadUsers(query) {
    requestId++;

    const currentRequest = requestId;

    loadingUsers.value = true;

    let result = null;

    try {
        result = await searchShareProfile(query);
    } catch (err) {
        if (currentRequest !== requestId) {
            return;
        }

        loadingUsers.value = false;
        shareUsers.value = [];
        addNotification(err.message || 'Could not load users', 'error');

        if (!query) {
            showShare.value = false;
        }

        return;
    }

    if (currentRequest !== requestId) {
        return;
    }

    loadingUsers.value = false;

    if (!result.status) {
        addNotification(result.message || 'Could not load users', 'error');
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

async function submitShare() {
    if (!hasSelection.value || sharing.value) {
        return;
    }

    sharing.value = true;

    try {
        await shareProfile(getProfileId(), selectedUserIds.value);
        addNotification('Profile shared', 'success');
        closeShare();
    } catch (err) {
        addNotification(err.message || 'Could not share profile', 'error');
    } finally {
        sharing.value = false;
    }
}

async function handleFollow() {
    const id = route.query.id;

    try {
        const result = await requestFollow(id, "POST");

        if (result.status) {
            followingStatus.value = result.followStatus;
            emit('follow');
        } else {
            addNotification("could not follow user", 'error');
        }
    } catch (err) {
        addNotification("could not follow user", 'error');
        console.error(err);
    }
}

async function handleRemoveFollow() {
    const id = route.query.id;
    const oldStatus = followingStatus.value;

    try {
        const result = await requestFollow(id, "DELETE");

        if (result.status) {
            followingStatus.value = result.followStatus;

            if (oldStatus === 0) {
                emit('cancel-request');
            } else if (oldStatus === 1) {
                emit('unfollow');
            }
        }
    } catch (err) {
        addNotification("could not unfollow user", 'error');
        console.error(err);
    }
}

onBeforeUnmount(() => {
    clearTimeout(searchTimer);
    window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
    <section class="profile-header">
        <div class="cover">
            <div class="cover-grid"></div>
        </div>

        <div class="profile-information">
            <div class="avatar">
                <img
                    v-if="props.avatarPath"
                    :src="props.avatarPath"
                    alt="Profile avatar"
                >
            </div>

            <div class="profile-details">
                <div class="name-row">
                    <div class="identity">
                        <h1>
                            {{ props.firstName }} {{ props.lastName }}
                        </h1>

                        <p class="username">
                            {{ props.username || '' }}
                        </p>

                        <div
                            v-if="props.email || props.dob"
                            class="contact-details"
                        >
                            <span
                                v-if="props.email"
                                class="contact-item"
                            >
                                <span class="contact-label">Email</span>
                                <span class="contact-value">
                                    {{ props.email }}
                                </span>
                            </span>

                            <span
                                v-if="props.dob && formatDOB(props.dob) != '1 January 1'"
                                class="contact-item"
                            >
                                <span class="contact-label">DOB</span>
                                <span class="contact-value">
                                    {{ formatDOB(props.dob) }}
                                </span>
                            </span>
                        </div>
                    </div>

                    <div class="profile-actions">
                        <a
                            v-if="props.addEdit"
                            href="/me/edit"
                            class="edit-button"
                        >
                            Edit profile
                        </a>

                        <template v-else>
                            <button
                                v-if="message"
                                class="relationship-button message"
                                @click="handleMessage"
                            >
                                Message
                            </button>

                            <button
                                v-if="followingStatus === -1"
                                class="relationship-button follow"
                                @click="handleFollow"
                            >
                                Follow
                            </button>

                            <button
                                v-else-if="followingStatus === 0"
                                class="relationship-button requested"
                                @click="handleRemoveFollow"
                            >
                                Requested
                            </button>

                            <button
                                v-else-if="followingStatus === 1"
                                class="relationship-button following"
                                @click="handleRemoveFollow"
                            >
                                Following
                            </button>
                        </template>

                        <button
                            type="button"
                            class="relationship-button share"
                            @click="handleShare"
                        >
                            Share profile
                        </button>
                    </div>
                </div>

                <p class="about">
                    {{ props.bio }}
                </p>

                <div class="profile-stats">
                    <span>
                        <strong>{{ props.numOfPosts }}</strong> Posts
                    </span>

                    <span>
                        <strong>{{ props.numOfFollowing }}</strong> Following
                    </span>

                    <span>
                        <strong>{{ props.numOfFollowers }}</strong> Followers
                    </span>
                </div>
            </div>
        </div>
    </section>

    <Teleport to="body">
        <div v-if="showShare" class="share-overlay" @click.self="closeShare">
            <div class="share-dialog" role="dialog" aria-modal="true" aria-label="Share profile">
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
.profile-header {
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.cover {
    position: relative;
    height: 150px;
    z-index: 0;
    overflow: hidden;
    background: var(--input-focus);
    border-bottom: 2px solid var(--main-color);
}

.cover::before,
.cover::after {
    position: absolute;
    content: "";
    border: 2px solid var(--main-color);
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.cover::before {
    width: 150px;
    height: 150px;
    right: 100px;
    top: -70px;
    transform: rotate(25deg);
}

.cover::after {
    width: 80px;
    height: 80px;
    left: 120px;
    bottom: -40px;
    transform: rotate(45deg);
}

.cover-grid {
    position: absolute;
    inset: 0;
    opacity: 0.15;
    background-image:
        linear-gradient(var(--main-color) 1px, transparent 1px),
        linear-gradient(90deg, var(--main-color) 1px, transparent 1px);
    background-size: 25px 25px;
}

.profile-information {
    display: flex;
    gap: 28px;
    padding: 0 35px 30px;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
}

.avatar {
    flex-shrink: 0;
    width: 150px;
    height: 150px;
    margin-top: -75px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 4px solid var(--bg-color);
    outline: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 65px;
    box-shadow: 5px 5px var(--main-color);
    position: relative;
    z-index: 2;
}

.profile-details {
    width: 100%;
    padding-top: 22px;
}

.name-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 20px;
}

.identity {
    min-width: 0;
}

.profile-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
}

h1 {
    margin: 0;
    color: var(--main-color);
    font-family: "Liter", serif;
    font-size: 36px;
    line-height: 1;
}

.username {
    margin: 7px 0 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.contact-details {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 18px;
    margin-top: 9px;
}

.contact-item {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.contact-label {
    color: var(--main-color);
    font-weight: 700;
    text-transform: uppercase;
}

.contact-value {
    overflow-wrap: anywhere;
}

.edit-button,
.relationship-button {
    flex-shrink: 0;
    padding: 11px 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.1s, box-shadow 0.1s, background 0.15s;
}

.edit-button:hover,
.relationship-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.relationship-button.follow {
    background: var(--input-focus);
    color: white;
}

.relationship-button.requested {
    background: var(--bg-color);
    color: var(--main-color);
}

.relationship-button.following {
    background: var(--main-color);
    color: var(--bg-color);
}

.relationship-button.message {
    background: var(--bg-color);
    color: var(--main-color);
}

.relationship-button.share {
    background: var(--bg-color);
    color: var(--main-color);
}

.relationship-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.about {
    max-width: 650px;
    margin: 18px 0;
    color: var(--font-color-sub);
    font-size: 14px;
    line-height: 1.6;
}

.profile-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 22px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.profile-stats strong {
    color: var(--main-color);
    font-size: 12px;
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

@media (max-width: 800px) {
    .profile-information {
        gap: 20px;
        padding-left: 25px;
        padding-right: 25px;
    }

    .name-row {
        flex-direction: column;
        gap: 15px;
    }

    .profile-actions {
        width: 100%;
        flex-wrap: wrap;
    }
}

@media (max-width: 650px) {
    .cover {
        height: 150px;
    }

    .profile-information {
        display: block;
        padding: 0 20px 25px;
    }

    .avatar {
        width: 115px;
        height: 115px;
        margin-top: -58px;
        font-size: 48px;
    }

    .profile-details {
        padding-top: 20px;
    }

    h1 {
        font-size: 29px;
    }

    .name-row {
        align-items: flex-start;
    }

    .profile-actions {
        gap: 8px;
    }

    .profile-actions .relationship-button,
    .profile-actions .edit-button {
        padding: 9px 12px;
    }

    .contact-details {
        gap: 7px 14px;
    }

    .contact-item {
        font-size: 8px;
    }

    .profile-stats {
        gap: 12px;
    }
}

@media (max-width: 420px) {
    .contact-details {
        display: block;
    }

    .contact-item {
        margin-bottom: 6px;
    }

    .profile-actions {
        width: 100%;
    }
}
</style>