<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
    getNotifications,
    acceptFollowRequest,
    rejectFollowRequest,
    markNotificationsRead,
    respondToGroupInvite,
    respondToJoinRequest
} from '@/api/common/notifications';

import { respondGroupEvent } from '@/api/groups/events';
import { addNotification } from '@/data/notifications';
import { clearUnreadNotificationCount } from '@/data/notificationCount';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import NotificationPostDialog from '@/components/notifications/NotificationPostDialog.vue';
import { throttle } from '@/helpers/throttle';

const notifications = ref([]);
const loading = ref(true);
const loadingMore = ref(false);
const hasMore = ref(true);
const offset = ref(0);
const limit = 15;
const selectedPostId = ref(null);
const showPostDialog = ref(false);
const route = useRoute();
const router = useRouter();

const tabs = [
    { key: 'all', label: 'All' },
    { key: 'posts', label: 'Posts' },
    { key: 'comments', label: 'Comments' },
    { key: 'mentions', label: 'Mentions' },
    { key: 'follows', label: 'Follows' },
    { key: 'requests', label: 'Requests' },
    { key: 'invites', label: 'Group invites' },
    { key: 'events', label: 'Events' }
];

function normalizeTab(tab) {
    return tabs.some(item => item.key === tab) ? tab : 'all';
}

const activeTab = ref(normalizeTab(route.query.tab));

watch(
    () => route.query.tab,
    (tab) => setTab(normalizeTab(tab))
);

function isChatMention(notification) {
    return (
        !!notification.post_mention_user_id &&
        !notification.post_id &&
        !!notification.group?.id
    );
}

function openChat(notification) {
    const groupID = notification.group?.id;

    if (!groupID) {
        return;
    }

    router.push({ path: `/groups/${groupID}`, query: { tab: 'chat' } });
}

function getCategory(notification) {
    if (isChatMention(notification)) {
        return 'mentions';
    }

    if (
        notification.group_invite_user_id ||
        notification.group_join_user_id ||
        notification.group_accept_user_id
    ) {
        return 'invites';
    }

    if (
        notification.comment_mention_user_id &&
        notification.group?.id &&
        !notification.post_id
    ) {
        return 'mentions';
    }

    if (notification.follow_request_user_id) {
        return 'requests';
    }

    if (notification.follow_user_id || notification.follow_request_accept_user_id) {
        return 'follows';
    }

    if (
        notification.comment_reply_user_id ||
        notification.comment_like_user_id ||
        notification.comment_mention_user_id
    ) {
        return 'comments';
    }

    if (notification.event_invite_user_id || notification.event_response_user_id) {
        return 'events';
    }

    return 'posts';
}

const tabCounts = computed(() => {
    const counts = { all: notifications.value.length };

    for (const notification of notifications.value) {
        const category = getCategory(notification);

        counts[category] = (counts[category] || 0) + 1;
    }

    return counts;
});

const visibleNotifications = computed(() =>
    activeTab.value === 'all'
        ? notifications.value
        : notifications.value.filter(
            notification => getCategory(notification) === activeTab.value
        )
);

const emptyMessages = {
    all: 'No notifications yet',
    posts: 'No post notifications',
    comments: 'No comment notifications',
    mentions: 'No mentions',
    follows: 'No follow notifications',
    requests: 'No follow requests',
    invites: 'No group invites',
    events: 'No event notifications'
};

let liveRefreshing = false;
let liveQueued = false;

async function handleLiveNotification() {
    if (liveRefreshing) {
        liveQueued = true;
        return;
    }

    liveRefreshing = true;

    try {
        do {
            liveQueued = false;

            const result = await getNotifications(0, limit);

            const latest = Array.isArray(result.notifications)
                ? result.notifications
                : [];

            const knownIDs = new Set(notifications.value.map(item => item.id));
            const fresh = latest.filter(item => !knownIDs.has(item.id));

            if (fresh.length) {
                notifications.value.unshift(...fresh);
                offset.value += fresh.length;
            }

            await markAsRead();
        } while (liveQueued);
    } catch (err) {
        console.error(err);
    } finally {
        liveRefreshing = false;
    }
}

function handleVisibilityChange() {
    if (!document.hidden) {
        handleLiveNotification();
    }
}

onMounted(async () => {
    window.addEventListener('notification-received', handleLiveNotification);
    document.addEventListener('visibilitychange', handleVisibilityChange);
    await loadNotifications();
    window.addEventListener('scroll', handleScroll);
    markAsRead();
    fillPage();
});

onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll);
    window.removeEventListener('notification-received', handleLiveNotification);
    document.removeEventListener('visibilitychange', handleVisibilityChange);
});

async function markAsRead() {
    if (document.hidden) {
        return;
    }

    try {
        await markNotificationsRead();
        clearUnreadNotificationCount();
    } catch (err) {
        console.error(err);
    }
}

async function loadNotifications() {
    try {
        const result = await getNotifications(offset.value, limit);

        const newNotifications = Array.isArray(result.notifications)
            ? result.notifications
            : Array.isArray(result)
                ? result
                : [];

        notifications.value.push(...newNotifications);

        hasMore.value = result.hasMore ?? newNotifications.length === limit;

        offset.value += newNotifications.length;
    } catch (err) {
        console.error(err);
        addNotification('could not fetch notifications');
    } finally {
        loading.value = false;
    }
}

async function loadMore() {
    if (loadingMore.value || !hasMore.value) {
        return 0;
    }

    loadingMore.value = true;

    let loaded = 0;

    try {
        const result = await getNotifications(offset.value, limit);

        const newNotifications = Array.isArray(result.notifications)
            ? result.notifications
            : Array.isArray(result)
                ? result
                : [];

        notifications.value.push(...newNotifications);

        hasMore.value = result.hasMore ?? newNotifications.length === limit;

        offset.value += newNotifications.length;

        loaded = newNotifications.length;
    } catch (err) {
        console.error(err);
        addNotification('could not fetch notifications');
    } finally {
        loadingMore.value = false;
    }

    return loaded;
}

async function fillPage() {
    await nextTick();

    while (
        hasMore.value &&
        !loadingMore.value &&
        document.documentElement.scrollHeight <= window.innerHeight + 100
    ) {
        const loaded = await loadMore();

        if (!loaded) {
            break;
        }

        await nextTick();
    }
}

function setTab(tab) {
    if (activeTab.value === tab) {
        return;
    }

    activeTab.value = tab;
    window.scrollTo({ top: 0 });
    fillPage();
}

const handleScroll = throttle(() => {
    const scrollPosition = window.innerHeight + window.scrollY;
    const pageHeight = document.documentElement.scrollHeight;

    if (scrollPosition >= pageHeight - 300) {
        loadMore();
    }
}, 300);

function getActor(notification) {
    return notification.actor || {};
}

function getActorName(notification) {
    const actor = getActor(notification);

    return `${actor.firstName || ''} ${actor.lastName || ''}`.trim() || 'Someone';
}

function getPostImage(notification) {
    const post = notification.post;

    if (!post?.imagePath) {
        return '';
    }

    const path = post.imagePath.toLowerCase();

    if (
        path.endsWith('.mp4') ||
        path.endsWith('.webm') ||
        path.endsWith('.mov') ||
        path.endsWith('.avi')
    ) {
        return '';
    }

    return `/uploads/${post.imagePath}`;
}

function openProfile(userID) {
    if (!userID) {
        return;
    }

    window.location.href = `/user?id=${userID}`;
}

function openPost(postID) {
    if (!postID) {
        return;
    }

    selectedPostId.value = Number(postID);
    showPostDialog.value = true;
}

function closePostDialog() {
    showPostDialog.value = false;
    selectedPostId.value = null;
}

function getMessage(notification) {
    if (notification.message) {
        return notification.message;
    }

    return 'You have a new notification';
}

function isFollowRequest(notification) {
    return !!notification.follow_request_user_id;
}

function isFollow(notification) {
    return !!notification.follow_user_id;
}

function isPostNotification(notification) {
    return !!notification.post_id;
}

function isComment(notification) {
    return !!(
        notification.comment_reply_user_id ||
        notification.comment_like_user_id ||
        notification.comment_mention_user_id
    );
}

function getActorID(notification) {
    return (
        notification.message_user_id ||
        notification.comment_reply_user_id ||
        notification.follow_request_user_id ||
        notification.follow_request_accept_user_id ||
        notification.follow_user_id ||
        notification.post_like_user_id ||
        notification.post_dislike_user_id ||
        notification.comment_like_user_id ||
        notification.comment_mention_user_id ||
        notification.post_mention_user_id ||
        notification.group_invite_user_id ||
        notification.group_join_user_id ||
        notification.group_accept_user_id ||
        notification.event_invite_user_id ||
        notification.event_response_user_id ||
        null
    );
}

const respondingInvites = ref(new Set());

function isGroupInvite(notification) {
    return !!notification.group_invite_user_id;
}

async function answerInvite(notification, status) {
    const groupID = notification.group?.id;

    if (!groupID || respondingInvites.value.has(notification.id)) {
        return;
    }

    respondingInvites.value.add(notification.id);

    try {
        await respondToGroupInvite(groupID, status);

        notifications.value = notifications.value.filter(
            item => !(isGroupInvite(item) && item.group?.id === groupID)
        );

        offset.value = Math.max(0, offset.value - 1);

        addNotification(
            status === 1 ? 'group invite accepted' : 'group invite rejected'
        );
    } catch (err) {
        console.error(err);
        addNotification(err.message || 'could not update invite');
    } finally {
        respondingInvites.value.delete(notification.id);
    }
}

function isGroupActivity(notification) {
    return (
        !!notification.group?.id &&
        !notification.post_id &&
        !isChatMention(notification) &&
        !isGroupInvite(notification) &&
        !isGroupJoin(notification) &&
        !isEventInvite(notification)
    );
}

function openGroup(notification) {
    const groupID = notification.group?.id;

    if (!groupID) {
        return;
    }

    router.push({
        path: `/groups/${groupID}`,
        query: notification.event ? { tab: 'events' } : {}
    });
}

function isEventInvite(notification) {
    return !!notification.event_invite_user_id && !!notification.event;
}

function formatEventDate(value) {
    const date = new Date(value);

    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

async function answerEvent(notification, value) {
    const eventID = notification.event?.id;

    if (!eventID || respondingInvites.value.has(notification.id)) {
        return;
    }

    respondingInvites.value.add(notification.id);

    try {
        await respondGroupEvent(eventID, value);

        notifications.value = notifications.value.filter(
            item => item.id !== notification.id
        );

        offset.value = Math.max(0, offset.value - 1);

        addNotification(value === 1 ? 'you are going' : 'you are not going');
    } catch (err) {
        console.error(err);
        addNotification(err.message || 'could not save response');
    } finally {
        respondingInvites.value.delete(notification.id);
    }
}

function isGroupJoin(notification) {
    return !!notification.group_join_user_id;
}

async function answerJoin(notification, code) {
    const groupID = notification.group?.id;

    if (!groupID || respondingInvites.value.has(notification.id)) {
        return;
    }

    respondingInvites.value.add(notification.id);

    try {
        await respondToJoinRequest(
            groupID,
            notification.group_join_user_id,
            code
        );

        notifications.value = notifications.value.filter(
            item => item.id !== notification.id
        );

        offset.value = Math.max(0, offset.value - 1);

        addNotification(
            code === 1 ? 'join request accepted' : 'join request rejected'
        );
    } catch (err) {
        console.error(err);
        addNotification(err.message || 'could not update request');
    } finally {
        respondingInvites.value.delete(notification.id);
    }
}

async function acceptRequest(notification) {
    if (respondingInvites.value.has(notification.id)) {
        return;
    }

    respondingInvites.value.add(notification.id);

    try {
        const result = await acceptFollowRequest(
            notification.follow_request_user_id
        );

        if (!result.status) {
            addNotification(result.message || 'could not accept follow request');
            return;
        }

        notifications.value = notifications.value.filter(
            item => item.id !== notification.id
        );

        addNotification('follow request accepted');
    } catch (err) {
        console.error(err);
        addNotification('could not accept follow request');
    } finally {
        respondingInvites.value.delete(notification.id);
    }
}

async function rejectRequest(notification) {
    if (respondingInvites.value.has(notification.id)) {
        return;
    }

    respondingInvites.value.add(notification.id);

    try {
        const result = await rejectFollowRequest(
            notification.follow_request_user_id
        );

        if (!result.status) {
            addNotification(result.message || 'could not reject follow request');
            return;
        }

        notifications.value = notifications.value.filter(
            item => item.id !== notification.id
        );

        offset.value = Math.max(0, offset.value - 1);

        addNotification('follow request rejected');
    } catch (err) {
        console.error(err);
        addNotification('could not reject follow request');
    } finally {
        respondingInvites.value.delete(notification.id);
    }
}
</script>

<template>
    <div class="notifications-page-wrap">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="notifications-page">
                <div class="page-header">
                    <span>ACTIVITY</span>
                    <h1>Notifications</h1>
                </div>

                <div class="notification-tabs">
                    <button
                        v-for="tab in tabs"
                        :key="tab.key"
                        type="button"
                        class="tab-button"
                        :class="{ active: activeTab === tab.key }"
                        @click="setTab(tab.key)"
                    >
                        {{ tab.label }}
                        
                    </button>
                </div>

                <section class="notifications-card">
                    <div v-if="loading" class="empty-state">
                        Loading notifications...
                    </div>

                    <div
                        v-else-if="visibleNotifications.length === 0 && loadingMore"
                        class="empty-state"
                    >
                        Loading...
                    </div>

                    <div
                        v-else-if="visibleNotifications.length === 0"
                        class="empty-state"
                    >
                        {{ emptyMessages[activeTab] }}
                    </div>

                    <div
                        v-else
                        v-for="notification in visibleNotifications"
                        :key="notification.id"
                        class="notification"
                    >
                        <div
                            class="notification-avatar"
                            @click="openProfile(getActorID(notification))"
                        >
                            <img
                                v-if="getActor(notification).avatarPath"
                                :src="`/uploads/${getActor(notification).avatarPath}`"
                                alt=""
                            >

                            <span v-else>
                                {{ getActorName(notification).charAt(0).toUpperCase() }}
                            </span>
                        </div>

                        <div class="notification-content">
                            <div class="notification-top">
                                <button
                                    class="actor-name"
                                    @click="openProfile(getActorID(notification))"
                                >
                                    {{ getActorName(notification) }}
                                </button>

                                <span class="notification-time">
                                    {{ notification.created_at }}
                                </span>
                            </div>

                            <p class="notification-message">
                                {{ getMessage(notification) }}
                            </p>

                            <div
                                v-if="isChatMention(notification)"
                                class="notification-actions"
                            >
                                <button
                                    class="accept-button"
                                    @click="openChat(notification)"
                                >
                                    Open chat
                                </button>
                            </div>

                            <div
                                v-if="isComment(notification) && notification.comment"
                                class="comment-preview"
                                @click="openPost(notification.post?.id)"
                            >
                                <p>{{ notification.comment }}</p>

                                <span class="comment-likes">
                                    ♥ {{ notification.comment_likes || 0 }}
                                </span>
                            </div>

                            <div
                                v-if="isFollowRequest(notification)"
                                class="notification-actions invite-actions"
                            >
                                <button
                                    class="accept-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="acceptRequest(notification)"
                                >
                                    Accept
                                </button>

                                <button
                                    class="reject-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="rejectRequest(notification)"
                                >
                                    Reject
                                </button>
                            </div>

                            <div
                                v-if="(isGroupInvite(notification) || isGroupJoin(notification) || isGroupActivity(notification)) && notification.group"
                                class="group-preview"
                            >
                                <div class="group-preview-avatar">
                                    <img
                                        v-if="notification.group.avatar"
                                        :src="`/uploads/${notification.group.avatar}`"
                                        alt=""
                                    >

                                    <span v-else>
                                        {{ (notification.group.name || '?').charAt(0).toUpperCase() }}
                                    </span>
                                </div>

                                <div class="group-preview-text">
                                    <span>GROUP</span>
                                    <p>{{ notification.group.name }}</p>
                                </div>
                            </div>

                            <div
                                v-if="isGroupActivity(notification)"
                                class="notification-actions"
                            >
                                <button
                                    class="accept-button"
                                    @click="openGroup(notification)"
                                >
                                    {{ notification.event ? 'Open events' : 'Open group' }}
                                </button>
                            </div>

                            <div
                                v-if="isGroupInvite(notification)"
                                class="notification-actions invite-actions"
                            >
                                <button
                                    class="accept-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerInvite(notification, 1)"
                                >
                                    Accept
                                </button>

                                <button
                                    class="reject-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerInvite(notification, -1)"
                                >
                                    Reject
                                </button>
                            </div>

                            <div
                                v-if="isEventInvite(notification)"
                                class="group-preview"
                            >
                                <div class="group-preview-text">
                                    <span>EVENT · {{ formatEventDate(notification.event.eventTime) }}</span>
                                    <p>{{ notification.event.title }}</p>
                                </div>
                            </div>

                            <div
                                v-if="isEventInvite(notification)"
                                class="notification-actions invite-actions"
                            >
                                <button
                                    class="accept-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerEvent(notification, 1)"
                                >
                                    Going
                                </button>

                                <button
                                    class="reject-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerEvent(notification, 0)"
                                >
                                    Not going
                                </button>
                            </div>

                            <div
                                v-if="isGroupJoin(notification)"
                                class="notification-actions invite-actions"
                            >
                                <button
                                    class="accept-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerJoin(notification, 1)"
                                >
                                    Accept
                                </button>

                                <button
                                    class="reject-button"
                                    :disabled="respondingInvites.has(notification.id)"
                                    @click="answerJoin(notification, -1)"
                                >
                                    Reject
                                </button>
                            </div>

                            <div
                                v-if="isPostNotification(notification)"
                                class="post-preview"
                                @click="openPost(notification.post?.id || notification.post_id)"
                            >
                                <div class="post-preview-text">
                                    <span>POST</span>

                                    <p>
                                        {{ notification.post?.content || 'Open post' }}
                                    </p>
                                </div>

                                <img
                                    v-if="getPostImage(notification)"
                                    :src="getPostImage(notification)"
                                    alt=""
                                    class="post-image"
                                >
                            </div>
                        </div>
                    </div>

                    <div
                        v-if="loadingMore"
                        class="loading-more"
                    >
                        Loading more...
                    </div>
                </section>
            </main>
        </div>

        <NotificationPostDialog
            :show="showPostDialog"
            :post-id="selectedPostId"
            @close="closePostDialog"
        />
    </div>
</template>

<style scoped>
.notifications-page-wrap {
    min-height: 100vh;
    padding-top: 64px;
}

.page-layout {
    display: flex;
    align-items: flex-start;
    gap: 24px;
    min-height: calc(100vh - 64px);
}

.notifications-page {
    width: min(900px, calc(100% - 48px));
    flex: 1;
    margin: 0 auto;
    padding: 32px 0 60px;
}

.page-header {
    margin-bottom: 28px;
}

.page-header span {
    display: block;
    margin-bottom: 8px;
    color: #2f8ff0;
    font-size: 9px;
    letter-spacing: 2px;
}

.page-header h1 {
    margin: 0;
    color: #202b38;
    font-size: 32px;
    font-weight: 700;
}

.notification-tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 20px;
}

.tab-count {
    margin-left: 6px;
    padding: 1px 6px;
    border-radius: 9px;
    background: #2f8ff0;
    color: #fff;
    font-size: 9px;
}

.tab-button {
    padding: 10px 18px;
    border: 2px solid #292929;
    border-radius: 5px;
    background: #fff;
    color: #292929;
    cursor: pointer;
    box-shadow: 3px 3px 0 #292929;
    font-size: 11px;
    font-weight: 700;
}

.tab-button.active {
    background: #292929;
    color: #fff;
}

.tab-button:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px 0 #292929;
}

.notifications-card {
    overflow: hidden;
    border: 2px solid #292929;
    border-radius: 7px;
    background: #fff;
    box-shadow: 6px 6px 0 #292929;
}

.notification {
    display: flex;
    gap: 16px;
    padding: 22px;
    border-bottom: 1px solid #d5d5d5;
}

.notification:last-child {
    border-bottom: none;
}

.notification-avatar {
    display: flex;
    flex: 0 0 48px;
    width: 48px;
    height: 48px;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid #292929;
    border-radius: 50%;
    background: #2f8ff0;
    color: #fff;
    cursor: pointer;
    font-size: 18px;
    font-weight: 700;
}

.notification-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.notification-content {
    flex: 1;
    min-width: 0;
}

.notification-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}

.actor-name {
    padding: 0;
    border: none;
    background: none;
    color: #202b38;
    cursor: pointer;
    font-size: 15px;
    font-weight: 700;
}

.actor-name:hover {
    text-decoration: underline;
}

.notification-time {
    color: #777;
    font-size: 10px;
}

.notification-message {
    margin: 7px 0 0;
    color: #333;
    font-size: 13px;
    line-height: 1.5;
}

.notification-actions {
    margin-top: 14px;
}

.accept-button {
    padding: 9px 18px;
    border: 2px solid #292929;
    border-radius: 5px;
    background: #2f8ff0;
    color: #fff;
    cursor: pointer;
    box-shadow: 3px 3px 0 #292929;
    font-size: 11px;
    font-weight: 700;
}

.accept-button:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px 0 #292929;
}

.invite-actions {
    display: flex;
    gap: 10px;
}

.reject-button {
    padding: 9px 18px;
    border: 2px solid #292929;
    border-radius: 5px;
    background: #fff;
    color: #292929;
    cursor: pointer;
    box-shadow: 3px 3px 0 #292929;
    font-size: 11px;
    font-weight: 700;
}

.reject-button:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px 0 #292929;
}

.accept-button:disabled,
.reject-button:disabled {
    cursor: not-allowed;
    opacity: 0.6;
}

.group-preview {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 14px;
    padding: 10px 14px;
    border: 2px solid #292929;
    border-radius: 5px;
    background: #fafafa;
}

.group-preview-avatar {
    display: flex;
    flex: 0 0 40px;
    width: 40px;
    height: 40px;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid #292929;
    border-radius: 8px;
    background: #2f8ff0;
    color: #fff;
    font-weight: 700;
}

.group-preview-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.group-preview-text span {
    color: #2f8ff0;
    font-size: 8px;
    letter-spacing: 1.5px;
}

.group-preview-text p {
    margin: 4px 0 0;
    color: #333;
    font-size: 13px;
    font-weight: 700;
}

.post-preview {
    display: flex;
    min-height: 78px;
    margin-top: 14px;
    overflow: hidden;
    border: 2px solid #292929;
    border-radius: 5px;
    cursor: pointer;
    background: #fafafa;
}

.post-preview-text {
    flex: 1;
    padding: 12px 14px;
}

.post-preview-text span {
    color: #2f8ff0;
    font-size: 8px;
    letter-spacing: 1.5px;
}

.post-preview-text p {
    display: -webkit-box;
    margin: 7px 0 0;
    overflow: hidden;
    color: #333;
    font-size: 12px;
    line-height: 1.4;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
}

.post-image {
    width: 100px;
    height: 100px;
    flex-shrink: 0;
    object-fit: cover;
    border-left: 2px solid #292929;
}

.comment-preview {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-top: 12px;
    padding: 11px 14px;
    border-left: 4px solid #2f8ff0;
    background: #f1f1f1;
    cursor: pointer;
}

.comment-preview p {
    margin: 0;
    color: #333;
    font-size: 12px;
    line-height: 1.5;
}

.comment-likes {
    flex-shrink: 0;
    color: #d9534f;
    font-size: 11px;
    font-weight: 700;
}

.loading-more {
    padding: 18px;
    color: #777;
    text-align: center;
    font-size: 11px;
}

.empty-state {
    padding: 60px 20px;
    color: #777;
    text-align: center;
    font-size: 12px;
}

@media (max-width: 1024px) {
    .notifications-page {
        width: min(100%, calc(100% - 32px));
    }
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
        gap: 0;
    }

    .notifications-page {
        width: calc(100% - 28px);
        padding: 24px 0 50px;
    }

    .page-header h1 {
        font-size: 26px;
    }
}

@media (max-width: 420px) {
    .notifications-page {
        width: calc(100% - 16px);
    }

    .notification {
        gap: 10px;
        padding: 13px;
    }

    .notification-avatar {
        flex: 0 0 38px;
        width: 38px;
        height: 38px;
        font-size: 15px;
    }

    .post-preview {
        flex-direction: column;
    }

    .post-image {
        width: 100%;
        height: 140px;
        border-left: none;
        border-top: 2px solid #292929;
    }
}

@media (max-width: 650px) {
    .notifications-page {
        width: calc(100% - 24px);
        padding-top: 20px;
    }

    .notification {
        padding: 16px;
    }

    .notification-top {
        align-items: flex-start;
        flex-direction: column;
        gap: 4px;
    }

    .post-image {
        width: 80px;
        height: 80px;
    }
}
</style>