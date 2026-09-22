<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { getPrivateChatsLists, searchChats } from '@/api/chats/chats';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    targetUserId: {
        type: [Number, String],
        default: null
    }
});

const emit = defineEmits(['select-chat']);

const chats = ref([]);
const offset = ref(0);
const loading = ref(false);
const hasMore = ref(true);
const searchValue = ref('');
const activeChatId = ref(null);

const listEl = ref(null);

function throttle(fn, wait = 300) {
    let lastCallTime = 0;
    let pendingTimeout = null;

    return function throttled(...args) {
        const now = Date.now();
        const remaining = wait - (now - lastCallTime);

        if (remaining <= 0) {
            if (pendingTimeout) {
                clearTimeout(pendingTimeout);
                pendingTimeout = null;
            }
            lastCallTime = now;
            fn.apply(this, args);
        } else if (!pendingTimeout) {
            pendingTimeout = setTimeout(() => {
                lastCallTime = Date.now();
                pendingTimeout = null;
                fn.apply(this, args);
            }, remaining);
        }
    };
}

function debounce(fn, wait = 300) {
    let pendingTimeout = null;

    return function debounced(...args) {
        if (pendingTimeout) {
            clearTimeout(pendingTimeout);
        }

        pendingTimeout = setTimeout(() => {
            pendingTimeout = null;
            fn.apply(this, args);
        }, wait);
    };
}

function normalizeList(result) {
    if (Array.isArray(result)) return result;
    return result.data || result.chats || result.groups || [];
}

function normalizeHasMore(result, list) {
    if (typeof result?.hasMore === 'boolean') return result.hasMore;
    return list.length > 0;
}

function mergeByUserId(existing, incoming) {
    const merged = new Map(existing.map((chat) => [chat.UserID, chat]));

    for (const chat of incoming) {
        if (!merged.has(chat.UserID)) {
            merged.set(chat.UserID, chat);
        }
    }

    return Array.from(merged.values());
}

async function loadChats({ reset = false } = {}) {
    if (loading.value) return;
    if (!reset && !hasMore.value) return;

    loading.value = true;
    const nextOffset = reset ? 0 : offset.value;

    try {
        const result = searchValue.value
            ? await searchChats(nextOffset, searchValue.value)
            : await getPrivateChatsLists(nextOffset);
``
        const list = normalizeList(result);
        console.log(list)
        chats.value = mergeByUserId(reset ? [] : chats.value, list);


        offset.value = nextOffset + list.length;
        hasMore.value = normalizeHasMore(result, list);

        if (reset && chats.value.length) {
            if (props.targetUserId) {
                const match = chats.value.find(
                    (chat) => String(chat.UserID) === String(props.targetUserId)
                );

                if (match) {
                    selectChat(match);
                }
            } else if (!activeChatId.value) {
                selectChat(chats.value[0]);
            }
        }
    } catch (err) {
        addNotification(err.message || "could not load chats", 'error');
        console.error(err);
    } finally {
        loading.value = false;
    }
}

function selectChat(chat) {
    console.log(chat)
    activeChatId.value = chat.UserID;
    emit('select-chat', chat);
}

const handleScroll = throttle(() => {
    const el = listEl.value;
    if (!el) return;

    const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 120;
    if (nearBottom) {
        loadChats();
    }
}, 250);

const handleSearchInput = debounce(() => {
    hasMore.value = true;
    loadChats({ reset: true });
}, 500);

onMounted(() => {
    loadChats({ reset: true });
    listEl.value?.addEventListener('scroll', handleScroll);
});

onBeforeUnmount(() => {
    listEl.value?.removeEventListener('scroll', handleScroll);
});
</script>

<template>
    <aside class="chat-sidebar">

        <div class="sidebar-heading">
            <div>
                <p class="eyebrow">MESSAGES</p>
                <h2>Chats</h2>
            </div>
        </div>

        <div class="search">
            <span class="search-icon">⌕</span>

            <input v-model="searchValue" type="text" placeholder="Search chats..." @input="handleSearchInput" />

            <button v-if="searchValue" type="button" class="clear-search"
                @click="searchValue = ''; handleSearchInput()">
                ×
            </button>
        </div>

        <div ref="listEl" class="chat-list">

            <button v-for="chat in chats" :key="chat.UserID" type="button" class="chat-item"
                :class="{ active: activeChatId === chat.UserID }" @click="selectChat(chat)">

                <div class="avatar">

                    <img v-if="chat.Avatar" :src="`/uploads/${chat.Avatar}`" alt="" />

                    <span v-else>
                        {{ (chat.FirstName || '?').charAt(0).toUpperCase() }}
                    </span>


                </div>

                <div class="chat-info">

                    <div class="chat-info-top">

                        <strong>
                            {{ chat.FirstName }} {{ chat.LastName }}
                        </strong>

                        <span v-if="chat.UnreadCount" class="badge">
                            {{ chat.UnreadCount > 99 ? '99+' : chat.UnreadCount }}
                        </span>

                    </div>



                </div>

                <span v-if="chat.LastMessageAt" class="time">
                    {{ chat.LastMessageAt }}
                </span>

                <span v-if="activeChatId === chat.UserID" class="active-arrow">
                    ›
                </span>

            </button>

            <p v-if="loading" class="status-text">
                Loading chats...
            </p>

            <p v-else-if="!chats.length" class="status-text">
                No chats found
            </p>

            <p v-else-if="!hasMore" class="status-text">
                No more chats
            </p>

        </div>

    </aside>
</template>

<style scoped>
.chat-sidebar {
    display: flex;
    flex-direction: column;
    width: 330px;
    flex-shrink: 0;
    height: calc(100vh - 64px - 40px);
    border: 2px solid var(--main-color);
    border-radius: 10px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    overflow: hidden;
}

.sidebar-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 20px 14px;
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    letter-spacing: 2px;
}

h2 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 25px;
}

.search {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 40px;
    margin: 0 15px 15px;
    padding: 0 11px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--page-background);
    transition: box-shadow 0.15s ease, transform 0.15s ease;
}

.search:focus-within {
    box-shadow: 3px 3px var(--main-color);
    transform: translate(-1px, -1px);
}

.search-icon {
    flex-shrink: 0;
    color: var(--font-color-sub);
    font-size: 20px;
    line-height: 1;
}

.search input {
    width: 100%;
    min-width: 0;
    border: 0;
    outline: 0;
    background: transparent;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.search input::placeholder {
    color: var(--font-color-sub);
}

.clear-search {
    flex-shrink: 0;
    width: 22px;
    height: 22px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--font-color-sub);
    font-size: 17px;
    line-height: 1;
    cursor: pointer;
}

.clear-search:hover {
    color: var(--font-color);
}

.chat-list {
    flex: 1;
    overflow-y: auto;
    border-top: 2px solid var(--page-background);
    padding: 5px 0;
}

.chat-list::-webkit-scrollbar {
    width: 5px;
}

.chat-list::-webkit-scrollbar-thumb {
    border-radius: 10px;
    background: var(--main-color);
}

.chat-item {
    position: relative;
    width: calc(100% - 12px);
    min-height: 72px;
    display: flex;
    align-items: center;
    gap: 12px;
    margin: 2px 6px;
    padding: 10px 11px;
    border: 2px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--font-color);
    text-align: left;
    cursor: pointer;
    transition:
        background 0.15s ease,
        border-color 0.15s ease,
        transform 0.15s ease;
}

.chat-item:hover {
    border-color: var(--main-color);
    background: var(--page-background);
    transform: translateX(2px);
}

.chat-item.active {
    border-color: var(--main-color);
    background: var(--input-focus);
    color: white;
    box-shadow: 3px 3px var(--main-color);
    transform: translate(-1px, -1px);
}

.avatar {
    position: relative;
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 18px;
    overflow: visible;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
}

.online-dot {
    position: absolute;
    right: -2px;
    bottom: -2px;
    width: 11px;
    height: 11px;
    border: 2px solid var(--bg-color);
    border-radius: 50%;
    background: #6bcb77;
}

.active .online-dot {
    border-color: var(--input-focus);
}

.chat-info {
    flex: 1;
    min-width: 0;
    overflow: hidden;
}

.chat-info-top {
    display: flex;
    align-items: center;
    gap: 7px;
    min-width: 0;
}

.chat-info-top strong {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
    font-weight: 700;
}

.preview {
    margin: 5px 0 0;
    overflow: hidden;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    line-height: 1.4;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.active .preview {
    color: rgba(255, 255, 255, 0.8);
}

.time {
    align-self: flex-start;
    flex-shrink: 0;
    margin-top: 3px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
}

.active .time {
    color: white;
}

.badge {
    flex-shrink: 0;
    min-width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 5px;
    border: 2px solid var(--main-color);
    border-radius: 999px;
    background: #d9534f;
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 700;
}

.active .badge {
    border-color: white;
    background: white;
    color: var(--input-focus);
}

.active-arrow {
    flex-shrink: 0;
    color: white;
    font-family: "Liter", serif;
    font-size: 20px;
    line-height: 1;
}

.status-text {
    margin: 0;
    padding: 22px 16px;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

@media (max-width: 800px) {

    .chat-sidebar {
        width: 100%;
        height: 320px;
    }

    .chat-list {
        padding: 3px 0;
    }

    .chat-item {
        min-height: 66px;
    }

    .avatar {
        width: 44px;
        height: 44px;
    }

}
</style>
