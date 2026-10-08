<script setup>
defineProps({
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
    showMembers: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['toggle-members']);

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
</script>

<template>
    <header class="group-header">
        <div class="group-avatar">
            <img
                v-if="avatarPath"
                :src="`/uploads/${avatarPath}`"
                :alt="name"
                class="group-avatar-img"
            >
            <span v-else class="group-avatar-fallback">
                {{ initials(name) }}
            </span>
        </div>

        <div class="group-info">
            <h1 class="group-name">
                {{ name }}
            </h1>

            <p v-if="description" class="group-description user-text">
                {{ description }}
            </p>
        </div>

        <div class="group-meta">
            <span class="members-count">
                {{ membersCount }} {{ membersCount === 1 ? 'member' : 'members' }}
            </span>

            <button
                class="dots-button"
                type="button"
                :class="{ active: showMembers }"
                @click="emit('toggle-members')"
            >
                <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                >
                    <circle cx="12" cy="5" r="2"></circle>
                    <circle cx="12" cy="12" r="2"></circle>
                    <circle cx="12" cy="19" r="2"></circle>
                </svg>
            </button>
        </div>
    </header>
</template>

<style scoped>
.group-header {
    position: relative;

    display: flex;
    align-items: center;
    gap: 18px;

    width: 100%;
    padding: 24px;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.group-avatar {
    flex-shrink: 0;

    width: 72px;
    height: 72px;

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
    font-size: 20px;
    font-weight: 600;
}

.group-info {
    flex: 1;
    min-width: 0;
}

.group-name {
    margin: 0 0 6px;

    color: var(--font-color);

    font-family: "Liter", serif;
    font-size: 24px;
    font-weight: 700;
}

.group-description {
    margin: 0;

    color: var(--font-color-sub);

    font-size: 13px;
    line-height: 1.5;
}

.group-meta {
    flex-shrink: 0;

    display: flex;
    align-items: center;
    gap: 12px;
}

.members-count {
    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    white-space: nowrap;
}

.dots-button {
    flex-shrink: 0;

    width: 36px;
    height: 36px;

    display: flex;
    align-items: center;
    justify-content: center;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--bg-color);
    color: var(--main-color);

    transition:
        transform 0.1s,
        background 0.15s,
        color 0.15s;
}

.dots-button svg {
    width: 16px;
    height: 16px;
}

.dots-button:hover,
.dots-button.active {
    background: var(--main-color);
    color: var(--bg-color);
}

.dots-button:active {
    transform: translate(1px, 1px);
}

@media (max-width: 650px) {
    .group-header {
        gap: 14px;
        padding: 18px;
        box-shadow: 4px 4px var(--main-color);
    }

    .group-avatar {
        width: 56px;
        height: 56px;
    }

    .group-name {
        font-size: 19px;
    }

    .group-description {
        font-size: 12px;
    }

    .group-meta {
        flex-direction: column;
        align-items: flex-end;
        gap: 8px;
    }

    .members-count {
        font-size: 10px;
    }
}
</style>
