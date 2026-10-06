<script setup>
import {
    notifications,
    removeNotification,
    postDialog,
    openPostDialog,
    closePostDialog
} from '@/data/notifications';
import NotificationPostDialog from '@/components/notifications/NotificationPostDialog.vue';
import { markNotificationRead } from '@/api/common/notifications';
import { setUnreadNotificationCount } from '@/data/notificationCount';
import { useRouter } from 'vue-router';

const router = useRouter();

async function markToastRead(notification) {
    if (!notification.notificationId) {
        return;
    }

    try {
        const result = await markNotificationRead(notification.notificationId);

        if (result.status) {
            setUnreadNotificationCount(result.count);
        }
    } catch (err) {
        console.error(err);
    }
}

function openToast(notification) {
    if (notification.postId) {
        openPostDialog(notification.postId);
    } else if (notification.route) {
        router.push(notification.route);
    } else {
        return;
    }

    markToastRead(notification);
    removeNotification(notification.id);
}
</script>
    
<template>
    <div class="notification-container">
        <div
            v-for="notification in notifications"
            :key="notification.id"
            class="notification"
            :class="[notification.type, { clickable: notification.postId || notification.route }]"
            @click="openToast(notification)"
        >
            <div
                v-if="notification.avatar || notification.initial"
                class="notification-avatar"
            >
                <img
                    v-if="notification.avatar"
                    :src="notification.avatar"
                    alt=""
                >

                <span v-else>{{ notification.initial }}</span>
            </div>

            <span class="notification-text">{{ notification.message }}</span>

            <img
                v-if="notification.image"
                :src="notification.image"
                alt=""
                class="notification-image"
            >

            <button
                type="button"
                @click.stop="removeNotification(notification.id)"
            >
                ×
            </button>
        </div>
    </div>

    <NotificationPostDialog
        :show="postDialog.show"
        :post-id="postDialog.postId"
        @close="closePostDialog"
    />
</template>

<style scoped>
.notification-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 9999;

    display: flex;
    flex-direction: column;
    gap: 10px;

    width: min(350px, calc(100vw - 40px));
}

.notification {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;

    padding: 14px 16px;

    border: 2px solid var(--main-color);
    border-radius: 5px;

    background: var(--bg-color);
    color: var(--font-color);

    box-shadow: 4px 4px var(--main-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.notification.clickable {
    cursor: pointer;
}

.notification-avatar {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--input-focus);
    color: #fff;
    font-size: 14px;
    font-weight: 700;
}

.notification-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.notification-image {
    flex-shrink: 0;
    width: 44px;
    height: 44px;
    object-fit: cover;
    border: 2px solid var(--main-color);
    border-radius: 4px;
}

.notification-text {
    flex: 1;
    min-width: 0;
}

.notification.success {
    border-color: var(--input-focus);
}

.notification.error {
    border-color: #d9534f;
}

.notification button {
    border: none;
    background: none;
    color: inherit;
    font-size: 18px;
    cursor: pointer;
}
</style>