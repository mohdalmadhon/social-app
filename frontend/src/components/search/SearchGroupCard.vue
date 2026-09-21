<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
    group: {
        type: Object,
        required: true
    }
});

const avatarFailed = ref(false);

const initials = computed(() => {
    const name = props.group.name;

    if (!name) {
        return '?';
    }

    return name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map(word => word[0]?.toUpperCase())
        .join('');
});

const members = computed(() => {
    const count = props.group.membersCount || 0;

    return `${count} ${count === 1 ? 'member' : 'members'}`;
});
</script>

<template>
    <article class="group-card">
        <div class="group-avatar">
            <img
                v-if="group.avatar && !avatarFailed"
                :src="`/uploads/${group.avatar}`"
                :alt="group.name"
                class="group-avatar-img"
                @error="avatarFailed = true"
            >

            <span v-else class="group-avatar-fallback">
                {{ initials }}
            </span>
        </div>

        <div class="group-info">
            <h3 class="group-name">
                {{ group.name }}
            </h3>

            <p v-if="group.description" class="group-description">
                {{ group.description }}
            </p>

            <div class="group-members">
                <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    aria-hidden="true"
                >
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                    <circle cx="9" cy="7" r="4"></circle>
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                    <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                </svg>

                <span>{{ members }}</span>
            </div>
        </div>

        <RouterLink
            v-if="group.isMember"
            :to="`/groups/${group.id}`"
            class="group-action"
        >
            Open
        </RouterLink>

        <span v-else class="group-status">
            Not joined
        </span>
    </article>
</template>

<style scoped>
.group-card {
    display: flex;
    align-items: center;
    gap: 16px;

    width: 100%;
    padding: 16px 18px;

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
    flex: 1;
    min-width: 0;
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

.group-action {
    flex-shrink: 0;

    padding: 10px 20px;

    border: 2px solid var(--main-color);
    border-radius: 5px;

    background: var(--bg-color);
    color: var(--main-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;

    transition:
        transform 0.1s,
        background 0.15s,
        color 0.15s;
}

.group-action:hover {
    transform: translate(1px, 1px);
    background: var(--main-color);
    color: var(--bg-color);
}

.group-status {
    flex-shrink: 0;

    padding: 6px 10px;

    border: 1px solid var(--font-color-sub);
    border-radius: 4px;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
}

@media (max-width: 650px) {
    .group-card {
        gap: 12px;
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
