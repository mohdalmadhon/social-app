<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
    groupId: {
        type: [Number, String],
        required: true
    },
    name: {
        type: String,
        default: ''
    },
    description: {
        type: String,
        default: ''
    },
    avatarPath: {
        type: String,
        default: ''
    },
    membersCount: {
        type: Number,
        default: 0
    },
    isMember: {
        type: Boolean,
        default: false
    },
    isPending: {
        type: Boolean,
        default: false
    },
    typingCount: {
        type: Number,
        default: 0
    },
    unreadCount: {
        type: Number,
        default: 0
    }
});

const typingLabel = computed(() => {
    if (props.typingCount <= 0) {
        return '';
    }

    if (props.typingCount === 1) {
        return 'Someone is typing';
    }

    if (props.typingCount === 2) {
        return '2 people are typing';
    }

    return 'More than 2 people are typing';
});

const showRequests = ref(false);

function openRequests() {
    showRequests.value = true;
}

function closeRequests() {
    showRequests.value = false;
}

const emit = defineEmits(['join', 'open']);

const requestPending = ref(props.isPending);
const requestLoading = ref(false);

function handleAction() {
    if (props.isMember) {
        emit('open', props.groupId);
        return;
    }

    requestToJoin(props.groupId);
}

function initials(name) {
    if (!name) {
        return '?';
    }

    return name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map(word => word[0]?.toUpperCase())
        .join('');
}

async function requestToJoin(groupID) {
    if (requestLoading.value) {
        return;
    }

    requestLoading.value = true;

    const code = requestPending.value ? -1 : 0;

    try {
        const response = await fetch('/api/groups/request', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({
                groupID: Number(groupID),
                code
            })
        });

        const data = await response.json();

        if (!response.ok || !data.status) {
            throw new Error(
                data.message || 'Could not process group request'
            );
        }

        requestPending.value = !requestPending.value;
    } catch (error) {
        console.error(error);
    } finally {
        requestLoading.value = false;
    }
}
</script>

<template>
    <article class="group-card" :class="{ 'has-unread': unreadCount > 0 }">
        <div class="group-avatar">
            <img v-if="avatarPath" :src="`/uploads/${avatarPath}`" :alt="name" class="group-avatar-img">
            <span v-else class="group-avatar-fallback">
                {{ initials(name) }}
            </span>
        </div>

        <div class="group-info">
            <h3 class="group-name">
                {{ name }}
                <span v-if="unreadCount > 0" class="unread-badge"
                    :aria-label="`${unreadCount} unread messages`">
                    {{ unreadCount > 99 ? '99+' : unreadCount }}
                </span>
            </h3>

            <p v-if="description" class="group-description user-text">
                {{ description }}
            </p>

            <div v-if="typingLabel" class="group-typing">
                <span class="group-typing-text">{{ typingLabel }}</span>
                <span class="typing-dots"><i></i><i></i><i></i></span>
            </div>

            <div v-else class="group-members">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                    <circle cx="9" cy="7" r="4"></circle>
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                    <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                </svg>
                <span>{{ membersCount }} {{ membersCount === 1 ? 'member' : 'members' }}</span>
            </div>
        </div>

        <button class="group-action" :class="{ 'is-member': isMember }" type="button"
            :disabled="requestLoading || isPending" @click="handleAction">
            {{
                requestLoading
                    ? '...'
                    : isMember
                        ? 'Open'
                        : requestPending
                            ? 'Requested'
                            : 'Join'
            }}
        </button>
    </article>
</template>

<style scoped>
.group-card {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: 12px 14px;

    width: 100%;
    min-width: 0;
    max-width: 100%;
    padding: 16px;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);

    transition:
        transform 0.15s,
        box-shadow 0.15s;
}

.group-card:hover {
    transform: translate(-1px, -1px);
    box-shadow: 6px 6px var(--main-color);
}

.group-avatar {
    flex-shrink: 0;

    width: 56px;
    height: 56px;

    display: flex;
    align-items: center;
    justify-content: center;

    overflow: hidden;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--input-focus);
}

.group-avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.group-avatar-fallback {
    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 16px;
    font-weight: 600;
}

.group-info {
    min-width: 0;
}

.unread-badge {
    display: inline-block;
    min-width: 20px;
    margin-left: 8px;
    padding: 1px 6px;

    border-radius: 999px;
    background: var(--input-focus);

    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    line-height: 18px;
    text-align: center;
    vertical-align: middle;
}

.group-card.has-unread {
    border-color: var(--input-focus);
}

.group-name {
    margin: 0 0 4px;

    color: var(--font-color);

    font-family: "Liter", serif;
    font-size: 16px;
    font-weight: 700;

    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.group-description {
    margin: 0 0 8px;

    color: var(--font-color-sub);

    font-size: 13px;
    line-height: 1.4;

    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    overflow-wrap: anywhere;
}

.group-members {
    display: flex;
    align-items: center;
    gap: 6px;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.group-members svg {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
}

.group-members span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.group-typing {
    display: flex;
    align-items: center;
    gap: 5px;

    min-width: 0;

    color: var(--input-focus);

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-style: italic;
}

.group-typing-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.typing-dots {
    flex-shrink: 0;
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

.group-action {
    grid-column: 1 / -1;

    width: 100%;
    min-width: 0;

    padding: 10px 20px;

    border: 2px solid var(--main-color);
    border-radius: 5px;

    background: var(--main-color);
    color: var(--bg-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;

    transition:
        transform 0.1s,
        background 0.15s,
        color 0.15s;
}

.group-action:hover:not(:disabled) {
    transform: translate(1px, 1px);
}

.group-action:disabled {
    opacity: 0.6;
    cursor: default;
}

.group-action.is-member {
    background: var(--bg-color);
    color: var(--main-color);
}

@media (max-width: 650px) {
    .group-card {
        gap: 10px 12px;
        padding: 12px 14px;
        box-shadow: 4px 4px var(--main-color);
    }

    .group-avatar {
        width: 46px;
        height: 46px;
    }

    .group-name {
        font-size: 14px;
    }

    .group-description {
        font-size: 12px;
    }

    .group-action {
        padding: 8px 14px;
        font-size: 12px;
    }
}
</style>