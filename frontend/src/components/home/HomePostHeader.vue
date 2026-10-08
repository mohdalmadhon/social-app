<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';

const props = defineProps({
    userId: {
        type: [Number, String],
        default: null
    },
    groupId: {
        type: [Number, String],
        default: null
    },
    firstName: {
        type: String,
        default: ''
    },
    lastName: {
        type: String,
        default: ''
    },
    avatarPath: {
        type: String,
        default: ''
    },
    relationship: {
        type: String,
        default: 'none'
    },
    taggedPeople: {
        type: Array,
        default: () => []
    },
    formattedDate: {
        type: String,
        default: ''
    },
    locationDisplay: {
        type: String,
        default: ''
    },
    hasLocation: {
        type: Boolean,
        default: false
    },
    canDelete: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['open-tags', 'open-location', 'delete']);
const isSelectedUsersPost = computed(() => {
    if (props.groupId === null || props.groupId === undefined || props.groupId === '') {
        return false;
    }

    return Number(props.groupId) > 0;
});

const menuOpen = ref(false);
const menuRoot = ref(null);

function requestDelete() {
    menuOpen.value = false;
    emit('delete');
}

function handleOutsideClick(event) {
    if (menuOpen.value && menuRoot.value && !menuRoot.value.contains(event.target)) {
        menuOpen.value = false;
    }
}

onMounted(() => {
    document.addEventListener('click', handleOutsideClick);
});

onBeforeUnmount(() => {
    document.removeEventListener('click', handleOutsideClick);
});

function openTaggedPeople() {
    emit('open-tags');
}

function openLocationDialog() {
    emit('open-location');
}

const relativeTime = computed(() => {
    if (!props.formattedDate) {
        return '';
    }

    const date = new Date(props.formattedDate);
    const now = new Date();

    if (Number.isNaN(date.getTime())) {
        return props.formattedDate;
    }

    const difference = now.getTime() - date.getTime();

    if (difference < 0) {
        return 'just now';
    }

    const minutes = Math.floor(difference / (1000 * 60));

    if (minutes < 1) {
        return 'just now';
    }

    if (minutes < 60) {
        return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'} ago`;
    }

    const hours = Math.floor(minutes / 60);

    if (hours < 24) {
        return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
    }

    const days = Math.floor(hours / 24);

    if (days < 7) {
        return `${days} ${days === 1 ? 'day' : 'days'} ago`;
    }

    const weeks = Math.floor(days / 7);

    return `${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`;
});
</script>

<template>
    <header class="post-header">
        <div class="author">
            <a :href="`/user?id=${userId}`" class="avatar">
                <img v-if="avatarPath" :src="`/uploads/${avatarPath}`" :alt="`${firstName} ${lastName}`">
                <span v-else>
                    {{ firstName?.charAt(0) }}
                </span>
            </a>

            <div class="author-information">
                <div class="author-line">
                    <span class="author-name">
                        {{ firstName }} {{ lastName }}
                    </span>

                    <span v-if="relationship === 'friend'" class="relationship-badge">
                        Friends
                    </span>

                    <button v-else-if="relationship === 'following' || relationship === 'none'" class="follow-button"
                        type="button">
                        Follow
                    </button>

                    <span v-if="isSelectedUsersPost" class="visibility-text">
                        Visibility limited by the user
                    </span>
                </div>

                <div class="post-meta">
                    <span class="post-date">
                        {{ relativeTime }}
                    </span>

                    <template v-if="hasLocation">
                        <span class="meta-dot">•</span>

                        <button class="location" type="button" @click="openLocationDialog">
                            <svg viewBox="0 0 24 24" aria-hidden="true">
                                <path
                                    d="M12 21s7-6.1 7-13a7 7 0 1 0-14 0c0 6.9 7 13 7 13Zm0-10a3 3 0 1 1 0-6 3 3 0 0 1 0 6Z" />
                            </svg>
                            {{ locationDisplay }}
                        </button>
                    </template>
                </div>
            </div>
        </div>

        <div v-if="canDelete" ref="menuRoot" class="post-options">
            <button class="more-button" type="button" aria-label="Post options" aria-haspopup="true"
                :aria-expanded="menuOpen" @click="menuOpen = !menuOpen">
                •••
            </button>

            <div v-if="menuOpen" class="post-menu">
                <button type="button" class="post-menu-item" @click="requestDelete">
                    Delete post
                </button>
            </div>
        </div>
    </header>
</template>

<style scoped>
.post-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 15px;
    padding: 18px 20px;
}

.author {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 12px;
}

.avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    box-shadow: 3px 3px var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 20px;
    font-weight: 600;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.author-information {
    min-width: 0;
}

.author-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 5px;
    font-size: 14px;
}

.author-name {
    color: var(--main-color);
    font-family: "Liter", serif;
    font-weight: 600;
}

.relationship-badge {
    padding: 3px 7px;
    border: 1px solid var(--main-color);
    border-radius: 4px;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 600;
}

.follow-button {
    padding: 3px 7px;
    border: 1px solid var(--input-focus);
    border-radius: 4px;
    background: transparent;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 600;
}

.follow-button:hover {
    background: var(--input-focus);
    color: white;
}

.visibility-text {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
}

.with-text {
    color: var(--font-color-sub);
}

.tagged-name {
    padding: 0;
    border: none;
    background: none;
    color: var(--input-focus);
    font-weight: 600;
}

.tagged-name:hover {
    text-decoration: underline;
}

.post-meta {
    margin-top: 5px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.post-date {
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--input-focus);
    color: white;
    font-weight: 600;
}

.meta-dot {
    font-size: 8px;
}

.location {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 6px;
    border: 1px solid var(--main-color);
    border-radius: 4px;
    background: transparent;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    font-weight: 600;
    transition:
        background 0.15s,
        color 0.15s;
}

.location:hover {
    background: var(--main-color);
    color: white;
}

.location:hover svg {
    fill: white;
}

.location svg {
    width: 12px;
    height: 12px;
    fill: var(--main-color);
    transition: fill 0.15s;
}

.more-button {
    flex-shrink: 0;
    padding: 5px 8px;
    border: none;
    background: transparent;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 15px;
    font-weight: 600;
}

.more-button:hover {
    color: var(--main-color);
}

.post-options {
    position: relative;
    flex-shrink: 0;
}

.more-button {
    cursor: pointer;
}

.post-menu {
    position: absolute;
    top: 100%;
    right: 0;
    z-index: 20;
    min-width: 140px;
    margin-top: 4px;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.post-menu-item {
    display: block;
    width: 100%;
    padding: 10px 14px;
    border: 0;
    background: transparent;
    color: #d9534f;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    text-align: left;
    cursor: pointer;
}

.post-menu-item:hover {
    background: var(--page-background);
}

@media (max-width: 650px) {
    .post-header {
        padding: 14px;
    }

    .avatar {
        width: 42px;
        height: 42px;
    }
}
</style>
