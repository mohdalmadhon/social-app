<script setup>
import { computed } from 'vue';

const props = defineProps({
    user: {
        type: Object,
        required: true
    }
});

const fullName = computed(() => {
    return `${props.user.firstName} ${props.user.lastName}`.trim();
});

const profileLocation = computed(() => {
    if (props.user.isMe) {
        return '/me';
    }

    return {
        path: '/user',
        query: { id: props.user.id }
    };
});

const badge = computed(() => {
    if (props.user.isMe) {
        return 'You';
    }

    if (props.user.isFriend) {
        return 'Friends';
    }

    if (props.user.followStatus === 1) {
        return 'Following';
    }

    if (props.user.followStatus === 0) {
        return 'Requested';
    }

    return '';
});
</script>

<template>
    <RouterLink :to="profileLocation" class="user-card">
        <div class="user-avatar">
            <img
                v-if="user.avatar"
                :src="`/uploads/${user.avatar}`"
                :alt="fullName"
            >

            <span v-else>
                {{ user.firstName?.charAt(0) }}
            </span>
        </div>

        <div class="user-info">
            <div class="user-line">
                <h3 class="user-name">
                    {{ fullName }}
                </h3>

                <span v-if="badge" class="user-badge">
                    {{ badge }}
                </span>

                <span v-if="user.isPrivate" class="user-badge muted">
                    Private
                </span>
            </div>

            <p v-if="user.username" class="user-username">
                @{{ user.username }}
            </p>
        </div>

        <span class="user-action">
            View profile
        </span>
    </RouterLink>
</template>

<style scoped>
.user-card {
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

.user-card:hover {
    transform: translate(-1px, -1px);
    box-shadow: 6px 6px var(--main-color);
}

.user-avatar {
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

    color: #fff;

    font-family: "Liter", serif;
    font-size: 22px;
    font-weight: 600;
}

.user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.user-info {
    flex: 1;
    min-width: 0;
}

.user-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
}

.user-name {
    margin: 0;

    min-width: 0;

    color: var(--font-color);

    font-family: "Liter", serif;
    font-size: 16px;
    font-weight: 700;

    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.user-badge {
    padding: 3px 7px;

    border: 1px solid var(--main-color);
    border-radius: 4px;

    color: var(--main-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 8px;
    font-weight: 600;
}

.user-badge.muted {
    border-color: var(--font-color-sub);
    color: var(--font-color-sub);
}

.user-username {
    margin: 4px 0 0;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 11px;

    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.user-action {
    flex-shrink: 0;

    padding: 10px 20px;

    border: 2px solid var(--main-color);
    border-radius: 5px;

    background: var(--bg-color);
    color: var(--main-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
}

@media (max-width: 650px) {
    .user-card {
        gap: 12px;
        padding: 12px 14px;
        box-shadow: 4px 4px var(--main-color);
    }

    .user-avatar {
        width: 46px;
        height: 46px;
        font-size: 18px;
    }

    .user-name {
        font-size: 14px;
    }

    .user-action {
        display: none;
    }
}
</style>
