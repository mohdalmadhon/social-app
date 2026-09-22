<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { getFriends } from '@/api/common/friends';
import { searchFollowing, searchFollows, removeFollower } from '@/api/users/profiles';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    type: {
        type: String,
        default: 'followers'
    },
    targetId: {
        type: [String, Number],
        required: true
    },
    followerList: {
        type: Object,
        default: () => ({})
    },
    ownProfile: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['close']);

const router = useRouter();
const route = useRoute();

const PAGE_SIZE = 20;
const SCROLL_THROTTLE_MS = 250;
const PREVIEW_SIZE = 5;

const showDialog = ref(false);
const list = ref([]);
const offset = ref(0);
const loading = ref(false);
const searching = ref(false);
const hasMore = ref(true);
const error = ref(null);
const searchResults = ref([]);
const searchQuery = ref('');
const scrollBox = ref(null);
const removingIds = ref(new Set());
const localFollowerList = ref({ ...(props.followerList || {}) });

let searchDebounce;
let targetID = props.targetId || route.query.id;

const canRemove = computed(() => {
    return props.type === 'followers' && props.ownProfile;
});

watch(
    () => props.followerList,
    value => {
        localFollowerList.value = { ...(value || {}) };
    }
);

const dialogTitle = computed(() => {
    if (props.type === 'following') return 'Following';
    if (props.type === 'friends') return 'Friends';
    return 'Followers';
});

const sectionTitle = computed(() => {
    if (props.type === 'following') return 'Following';
    if (props.type === 'friends') return 'Friends';
    return 'Followers';
});

const emptyText = computed(() => {
    if (props.type === 'following') return 'No following yet.';
    if (props.type === 'friends') return 'No friends yet.';
    return 'No followers yet.';
});

const previewList = computed(() => {
    return Object.entries(localFollowerList.value || {})
        .slice(0, PREVIEW_SIZE)
        .map(([id, user]) => ({
            ...user,
            ID: Number(id)
        }));
});

const totalCount = computed(() => {
    return Object.keys(localFollowerList.value || {}).length;
});

function throttle(fn, wait) {
    let lastCall = 0;
    let timeoutId = null;

    return function throttled(...args) {
        const now = Date.now();
        const remaining = wait - (now - lastCall);

        if (remaining <= 0) {
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }

            lastCall = now;
            fn.apply(this, args);
        } else if (!timeoutId) {
            timeoutId = setTimeout(() => {
                lastCall = Date.now();
                timeoutId = null;
                fn.apply(this, args);
            }, remaining);
        }
    };
}

function openDialog() {
    showDialog.value = true;
    list.value = [];
    offset.value = 0;
    hasMore.value = true;
    error.value = null;
    searchQuery.value = '';

    fetchPage();
}

function closeDialog() {
    showDialog.value = false;
    searchQuery.value = '';
    searchResults.value = [];
    list.value = [];
    offset.value = 0;
    hasMore.value = true;
    error.value = null;
    emit('close');
}

async function fetchPage() {
    if (
        loading.value ||
        !hasMore.value ||
        searchQuery.value.trim()
    ) {
        return;
    }

    loading.value = true;
    error.value = null;

    const endpoint =
        props.type === 'following'
            ? '/api/profile/following'
            : props.type === 'friends'
                ? '/api/friends'
                : '/api/profile/follow';

    try {
        const res = await fetch(
            `${endpoint}?targetid=${targetID}&offset=${offset.value}`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const body = await res.json();

        if (!res.ok || !body.status) {
            throw new Error(body.message || 'Failed to load');
        }

        const page = Object.entries(body.data || {}).map(([key, value]) => ({
            ID: Number(key),
            firstName: value.firstName,
            lastName: value.lastName,
            username: value.username,
            avatar: value.avatar
        }));

        list.value.push(...page);
        offset.value += PAGE_SIZE;

        if (page.length < PAGE_SIZE) {
            hasMore.value = false;
        }
    } catch (err) {
        console.error(err);
        error.value = 'Could not load more.';
    } finally {
        loading.value = false;
    }
}

function checkAndFetch() {
    if (searchQuery.value.trim()) {
        return;
    }

    const el = scrollBox.value;

    if (!el) {
        return;
    }

    const distanceFromBottom =
        el.scrollHeight - el.scrollTop - el.clientHeight;

    if (distanceFromBottom < 120) {
        fetchPage();
    }
}

const throttledScroll = throttle(
    checkAndFetch,
    SCROLL_THROTTLE_MS
);

async function runSearch(query) {
    if (!query) {
        searchResults.value = [];
        return;
    }

    searching.value = true;
    error.value = null;

    try {
        let result;

        if (props.type === 'following') {
            result = await searchFollowing(query, props.targetId);
        } else if (props.type === 'friends') {
            result = await getFriends(query, props.targetId);
        } else {
            result = await searchFollows(query, props.targetId);
        }

        searchResults.value = Object.entries(result.data || {}).map(
            ([id, value]) => ({
                ID: Number(id),
                ...value
            })
        );
    } catch (err) {
        console.error(err);
        searchResults.value = [];
    } finally {
        searching.value = false;
    }
}

watch(searchQuery, value => {
    clearTimeout(searchDebounce);

    const query = value.trim();

    if (!query) {
        searchResults.value = [];
        searching.value = false;
        error.value = null;
        return;
    }

    searchDebounce = setTimeout(() => {
        runSearch(query);
    }, 300);
});

function handleKeydown(e) {
    if (e.key === 'Escape' && showDialog.value) {
        closeDialog();
    }
}

async function goToProfile(id) {
    closeDialog();
    await router.replace(`/user?id=${id}`);
    window.location.reload();
}

async function removeFollowerHandler(id, event) {
    event?.stopPropagation();

    if (removingIds.value.has(id)) {
        return;
    }

    removingIds.value = new Set(removingIds.value).add(id);

    try {
        const result = await removeFollower(id);

        if (!result || !result.status) {
            throw new Error(result?.message || 'Could not remove follower');
        }

        const updated = { ...localFollowerList.value };
        delete updated[id];
        localFollowerList.value = updated;

        list.value = list.value.filter(follower => follower.ID !== id);
        searchResults.value = searchResults.value.filter(
            user => user.ID !== id
        );

        addNotification('Follower removed', 'success');
    } catch (err) {
        console.error(err);
        addNotification(err.message || 'Could not remove follower', 'error');
    } finally {
        const next = new Set(removingIds.value);
        next.delete(id);
        removingIds.value = next;
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
    clearTimeout(searchDebounce);
});
</script>

<template>
    <section class="followers-section">
        <div class="section-heading">
            <p class="eyebrow">SOCIAL</p>

            <div class="heading-row">
                <h2>{{ sectionTitle }}</h2>

                <button
                    v-if="totalCount"
                    type="button"
                    class="show-all-btn"
                    @click="openDialog"
                >
                    Show all
                </button>
            </div>
        </div>

        <div class="followers-card">
            <div
                v-if="previewList.length"
                class="followers-grid"
            >
                <article
                    v-for="follower in previewList"
                    :key="follower.ID"
                    class="follower-card"
                    @click="goToProfile(follower.ID)"
                >
                    <img
                        :src="
                            follower.avatar
                                ? `/uploads/${follower.avatar}`
                                : '/default-avatar.png'
                        "
                        :alt="`${follower.firstName} ${follower.lastName}`"
                        class="follower-avatar"
                    >

                    <div class="follower-info">
                        <p class="follower-name">
                            {{ follower.firstName }}
                            {{ follower.lastName }}
                        </p>
                    </div>

                    <button
                        v-if="canRemove"
                        type="button"
                        class="remove-follower-btn"
                        :disabled="removingIds.has(follower.ID)"
                        title="Remove follower"
                        @click="removeFollowerHandler(follower.ID, $event)"
                    >
                        ×
                    </button>
                </article>
            </div>

            <p v-else class="empty">
                {{ emptyText }}
            </p>
        </div>
    </section>

    <Teleport to="body">
        <div
            v-if="showDialog"
            class="dialog-overlay"
            @click="closeDialog"
        >
            <aside
                class="dialog-panel"
                @click.stop
            >
                <header class="dialog-header">
                    <div class="header-title">
                        <span class="header-accent"></span>

                        <div>
                            <p class="eyebrow">SOCIAL</p>
                            <h2>{{ dialogTitle }}</h2>
                        </div>
                    </div>

                    <button
                        type="button"
                        class="close-btn"
                        aria-label="Close"
                        @click="closeDialog"
                    >
                        ×
                    </button>
                </header>

                <div class="search-wrap">
                    <div class="search-box">
                        <span class="search-icon">⌕</span>

                        <input
                            v-model="searchQuery"
                            type="text"
                            placeholder="Search by name..."
                            class="group-search-input"
                        >
                    </div>
                </div>

                <div
                    ref="scrollBox"
                    class="dialog-body"
                    @scroll="throttledScroll"
                >
                    <template v-if="searchQuery.trim()">
                        <article
                            v-for="user in searchResults"
                            :key="user.ID"
                            class="follower-row"
                            @click="goToProfile(user.ID)"
                        >
                            <div class="row-avatar-wrap">
                                <img
                                    :src="
                                        user.avatar
                                            ? `/uploads/${user.avatar}`
                                            : '/default-avatar.png'
                                    "
                                    :alt="`${user.firstName} ${user.lastName}`"
                                    class="follower-avatar"
                                >
                            </div>

                            <div class="row-info">
                                <p class="follower-name">
                                    {{ user.firstName }}
                                    {{ user.lastName }}
                                </p>

                                <span class="row-label">
                                    VIEW PROFILE
                                </span>
                            </div>

                            <span class="row-arrow">↗</span>
                        </article>

                        <p
                            v-if="searching"
                            class="status-text"
                        >
                            Searching…
                        </p>

                        <p
                            v-if="
                                !searching &&
                                !searchResults.length
                            "
                            class="status-text empty-status"
                        >
                            No users found.
                        </p>
                    </template>

                    <template v-else>
                        <article
                            v-for="follower in list"
                            :key="follower.ID"
                            class="follower-row"
                            @click="goToProfile(follower.ID)"
                        >
                            <div class="row-avatar-wrap">
                                <img
                                    :src="
                                        follower.avatar
                                            ? `/uploads/${follower.avatar}`
                                            : '/default-avatar.png'
                                    "
                                    :alt="`${follower.firstName} ${follower.lastName}`"
                                    class="follower-avatar"
                                >
                            </div>

                            <div class="row-info">
                                <p class="follower-name">
                                    {{ follower.firstName }}
                                    {{ follower.lastName }}
                                </p>

                                <span class="row-label">
                                    VIEW PROFILE
                                </span>
                            </div>

                            <button
                                v-if="canRemove"
                                type="button"
                                class="remove-follower-row-btn"
                                :disabled="removingIds.has(follower.ID)"
                                title="Remove follower"
                                @click="removeFollowerHandler(follower.ID, $event)"
                            >
                                {{ removingIds.has(follower.ID) ? '…' : 'Remove' }}
                            </button>

                            <span v-else class="row-arrow">↗</span>
                        </article>

                        <p
                            v-if="loading"
                            class="status-text"
                        >
                            Loading…
                        </p>

                        <p
                            v-if="error"
                            class="status-text error"
                        >
                            {{ error }}
                        </p>

                        <p
                            v-if="
                                !hasMore &&
                                !list.length &&
                                !loading
                            "
                            class="status-text empty-status"
                        >
                            No
                            {{
                                type === 'following'
                                    ? 'following'
                                    : type === 'friends'
                                        ? 'friends'
                                        : 'followers'
                            }}
                            yet.
                        </p>

                        <p
                            v-if="
                                !hasMore &&
                                list.length &&
                                !loading
                            "
                            class="status-text end-status"
                        >
                            — That's everyone —
                        </p>
                    </template>
                </div>
            </aside>
        </div>
    </Teleport>
</template>

<style scoped>
.followers-section {
    width: 100%;
}

.section-heading {
    margin-bottom: 18px;
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 2px;
}

.heading-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
}

.heading-row h2 {
    margin: 0;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 25px;
}

.show-all-btn {
    padding: 8px 13px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
    cursor: pointer;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
}

.show-all-btn:hover {
    border-color: var(--input-focus);
    transform: translate(-2px, -2px);
    box-shadow: 5px 5px var(--input-focus);
}

.show-all-btn:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.followers-card {
    width: 100%;
    padding: 20px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);
    box-sizing: border-box;
}

.followers-grid {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 14px;
}

.follower-card {
    position: relative;
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    padding: 11px;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    cursor: pointer;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
}

.follower-card:hover {
    border-color: var(--input-focus);
    transform: translate(-2px, -2px);
    box-shadow: 5px 5px var(--input-focus);
}

.follower-avatar {
    display: block;
    width: 44px;
    height: 44px;
    flex: 0 0 44px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    object-fit: cover;
    box-sizing: border-box;
}

.follower-info {
    min-width: 0;
    flex: 1;
}

.follower-name {
    margin: 0;
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: 13px;
    font-weight: 700;
    line-height: 1.3;
    overflow-wrap: anywhere;
}

.remove-follower-btn {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    padding: 0;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--main-color);
    font-family: Arial, sans-serif;
    font-size: 14px;
    line-height: 1;
    cursor: pointer;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease,
        color 0.15s ease;
}

.remove-follower-btn:hover {
    border-color: #c0392b;
    color: #c0392b;
    transform: translate(-1px, -1px);
    box-shadow: 2px 2px var(--main-color);
}

.remove-follower-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
}

.remove-follower-row-btn {
    flex: 0 0 auto;
    padding: 6px 10px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.4px;
    cursor: pointer;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease,
        color 0.15s ease;
}

.remove-follower-row-btn:hover {
    border-color: #c0392b;
    color: #c0392b;
    transform: translate(-1px, -1px);
    box-shadow: 2px 2px var(--main-color);
}

.remove-follower-row-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
}

.empty {
    margin: 0;
    padding: 20px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    text-align: center;
}

.dialog-overlay {
    position: fixed;
    inset: 0;
    display: flex;
    justify-content: flex-end;
    background: rgba(0, 0, 0, 0.48);
    backdrop-filter: blur(3px);
    z-index: 1000;
}

.dialog-panel {
    position: relative;
    display: flex;
    flex-direction: column;
    width: min(410px, 94vw);
    height: 100%;
    background: var(--bg-color);
    border-left: 3px solid var(--main-color);
    box-shadow: -8px 0 var(--main-color);
    animation: slide-in 0.2s ease-out;
}

@keyframes slide-in {
    from {
        transform: translateX(100%);
    }

    to {
        transform: translateX(0);
    }
}

.dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
    padding: 20px;
    border-bottom: 2px solid var(--main-color);
    background: var(--bg-color);
    box-shadow: 0 4px 0 var(--input-focus);
    flex: 0 0 auto;
}

.header-title {
    display: flex;
    align-items: center;
    gap: 11px;
}

.header-accent {
    width: 5px;
    height: 40px;
    border: 1px solid var(--input-focus);
    border-radius: 2px;
    background: var(--input-focus);
    box-shadow: 2px 2px var(--main-color);
}

.dialog-header h2 {
    margin: 0;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 24px;
    line-height: 1;
}

.close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    flex: 0 0 auto;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    color: var(--font-color);
    font-family: Arial, sans-serif;
    font-size: 21px;
    line-height: 1;
    cursor: pointer;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
}

.close-btn:hover {
    border-color: var(--input-focus);
    transform: translate(-2px, -2px);
    box-shadow: 6px 6px var(--input-focus);
}

.close-btn:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.search-wrap {
    padding: 18px 20px 14px;
    flex: 0 0 auto;
}

.search-box {
    position: relative;
}

.search-icon {
    position: absolute;
    top: 50%;
    left: 12px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 19px;
    line-height: 1;
    transform: translateY(-50%);
    pointer-events: none;
}

.group-search-input {
    width: 100%;
    padding: 11px 14px 11px 38px;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    outline: none;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    color: var(--font-color);
    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: 14px;
    box-sizing: border-box;
    transition:
        border-color 0.15s ease,
        box-shadow 0.15s ease,
        transform 0.15s ease;
}

.group-search-input:focus {
    border-color: var(--input-focus);
    box-shadow: 4px 4px var(--input-focus);
    transform: translate(-1px, -1px);
}

.group-search-input::placeholder {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.dialog-body {
    flex: 1 1 auto;
    overflow-y: auto;
    padding: 6px 20px 24px;
    display: flex;
    flex-direction: column;
    gap: 11px;
}

.dialog-body::-webkit-scrollbar {
    width: 7px;
}

.dialog-body::-webkit-scrollbar-track {
    background: var(--bg-color);
    border-left: 1px solid var(--main-color);
}

.dialog-body::-webkit-scrollbar-thumb {
    border: 1px solid var(--main-color);
    background: var(--input-focus);
}

.follower-row {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 68px;
    padding: 10px 12px;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    box-sizing: border-box;
    cursor: pointer;
    overflow: hidden;
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
}

.follower-row::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 4px;
    height: 100%;
    background: var(--input-focus);
    transform: scaleY(0);
    transform-origin: bottom;
    transition: transform 0.15s ease;
}

.follower-row:hover {
    border-color: var(--input-focus);
    transform: translate(-3px, -3px);
    box-shadow: 7px 7px var(--input-focus);
}

.follower-row:hover::before {
    transform: scaleY(1);
}

.follower-row:active {
    transform: translate(0, 0);
    box-shadow: 1px 1px var(--main-color);
}

.row-avatar-wrap {
    flex: 0 0 auto;
    padding: 2px;
    border: 2px solid var(--input-focus);
    border-radius: 50%;
    box-shadow: 2px 2px var(--main-color);
}

.row-info {
    min-width: 0;
    flex: 1;
}

.row-label {
    display: block;
    margin-top: 4px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.6px;
    opacity: 0;
    transform: translateY(3px);
    transition:
        opacity 0.15s ease,
        transform 0.15s ease;
}

.follower-row:hover .row-label {
    opacity: 1;
    transform: translateY(0);
}

.row-arrow {
    flex: 0 0 auto;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 16px;
    font-weight: 700;
    opacity: 0;
    transform: translate(-3px, 3px);
    transition:
        opacity 0.15s ease,
        transform 0.15s ease;
}

.follower-row:hover .row-arrow {
    opacity: 1;
    transform: translate(0, 0);
}

.status-text {
    margin: 4px 0;
    padding: 13px;
    border: 1px dashed var(--main-color);
    border-radius: 6px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    letter-spacing: 0.4px;
    text-align: center;
}

.empty-status {
    color: var(--input-focus);
}

.end-status {
    border-style: solid;
    color: var(--input-focus);
}

.status-text.error {
    border-color: #c0392b;
    color: #c0392b;
}

@media (max-width: 800px) {
    .followers-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

@media (max-width: 500px) {
    .followers-grid {
        grid-template-columns: 1fr;
    }

    .dialog-panel {
        width: 100%;
        border-left-width: 2px;
        box-shadow: -5px 0 var(--main-color);
    }

    .dialog-header {
        padding: 17px;
    }

    .search-wrap {
        padding: 15px 17px 12px;
    }

    .dialog-body {
        padding: 6px 17px 20px;
    }

    .follower-row {
        min-height: 64px;
    }

    .row-label,
    .row-arrow {
        display: none;
    }
}

@media (max-width: 350px) {
    .dialog-header {
        padding: 14px;
    }

    .search-wrap {
        padding: 13px 14px 10px;
    }

    .dialog-body {
        padding: 5px 14px 18px;
    }
}
</style>