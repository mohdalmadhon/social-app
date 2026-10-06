<script setup>
import GroupChatTab from '@/components/groups/GroupChatTab.vue';
import GroupEventsTab from '@/components/groups/GroupEventsTab.vue';
import GroupFeedTab from '@/components/groups/GroupFeedTab.vue';
import GroupHeader from '@/components/groups/GroupHeader.vue';
import GroupMembersPanel from '@/components/groups/GroupMembersPanel.vue';
import GroupTabs from '@/components/groups/GroupTabs.vue';
import GroupFeedHeader from '@/components/groups/GroupFeedHeader.vue';
import GroupDialoge from '@/components/groups/GroupDialoge.vue';

import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';

import { addNotification } from '@/data/notifications';
import { openGroupPage } from '@/data/chatState';
import { getGroupPosts as fetchGroupPosts } from '@/api/posts/groupComments';
import { throttle } from '@/helpers/throttle';

import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

const group = ref({});
const members = ref([]);
const route = useRoute();

const groupID = Number(route.params.id);

const postsLimit = 10;

const posts = ref([]);
const postsOffset = ref(0);
const postsHasMore = ref(true);
const loadingPosts = ref(false);
const currentUserId = ref(null);

const events = ref([]);
const eventsTab = ref(null);

const validTabs = ['group', 'events', 'chat'];

function tabFromRoute() {
    return validTabs.includes(route.query.tab) ? route.query.tab : 'group';
}

const activeTab = ref(tabFromRoute());

watch(
    () => route.query.tab,
    () => {
        activeTab.value = tabFromRoute();
    }
);
const showMembers = ref(false);
const showPostDialog = ref(false);

let postsRequestID = 0;

function toggleMembers() {
    showMembers.value = !showMembers.value;
}

function closeMembers() {
    showMembers.value = false;
}

function setActiveTab(tab) {
    activeTab.value = tab;
    window.scrollTo({ top: 0 });
}

function openPostDialog() {
    showPostDialog.value = true;
}

function closePostDialog() {
    showPostDialog.value = false;
}

async function handlePostCreated() {
    showPostDialog.value = false;
    await getGroupPosts(true);
}

const isOwner = ref(false);
async function getGroupData() {
    console.log(groupID)
    try {
        const resp = await fetch(`/api/group?groupID=${groupID}`, {
            method: 'GET',
            credentials: 'include'
        });

        const result = await resp.json();
        console.log(result)
        if (result.isOwner) {
            isOwner.value = true;
        }
        
        if (!resp.ok || !result.status) {
            addNotification(
                result.message || result.messages || 'Could not get group data',
                'error'
            );
            return;
        }

        group.value = { ...result.data };
    } catch (err) {
        addNotification(err.message || 'Could not get group data', 'error');
    }
}

async function getGroupPosts(reset = false) {
    if (loadingPosts.value) return;
    if (!reset && !postsHasMore.value) return;

    const requestID = ++postsRequestID;

    loadingPosts.value = true;

    if (reset) {
        postsOffset.value = 0;
        postsHasMore.value = true;
        posts.value = [];
    }

    try {
        const result = await fetchGroupPosts(groupID, postsOffset.value);

        if (requestID !== postsRequestID) return;

        posts.value = [...posts.value, ...result.posts];
        postsOffset.value += result.posts.length;
        currentUserId.value = result.userId;

        
        if (result.posts.length < postsLimit) {
            postsHasMore.value = false;
        }
    } catch (err) {
        if (requestID !== postsRequestID) return;

        postsHasMore.value = false;

        addNotification(err.message || 'Could not get group posts', 'error');
    } finally {
        if (requestID === postsRequestID) {
            loadingPosts.value = false;
        }
    }
}

const handlePostsScroll = throttle(() => {
    if (activeTab.value !== 'group') return;

    const distance =
        document.documentElement.scrollHeight -
        window.scrollY -
        window.innerHeight;

    if (distance < 300) {
        getGroupPosts();
    }
}, 300);

function handleEventCreated(event) {
    activeTab.value = 'events';

    if (eventsTab.value) {
        eventsTab.value.addCreatedEvent(event);
    } else {
        events.value = [event, ...events.value];
    }
}

onMounted(async () => {
    openGroupPage.value = groupID;

    window.addEventListener('scroll', handlePostsScroll, { passive: true });

    await getGroupData();
    await getGroupPosts(true);
});

onUnmounted(() => {
    if (openGroupPage.value === groupID) {
        openGroupPage.value = null;
    }

    postsRequestID++;

    window.removeEventListener('scroll', handlePostsScroll);
});
</script>

<template>
    <div class="group-page">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="main-content">
                <div class="content-container">
                    <BackToHome />


                    <div class="sticky-top">
                        <div class="header-wrapper">
                            <GroupHeader :name="group.title" :description="group.description"
                                :avatar-path="group.avatar" :members-count="group.Count" :show-members="showMembers"
                                @toggle-members="toggleMembers" />

                            <GroupMembersPanel :show="showMembers" :isOwner="isOwner" :members="members"
                                :group-i-d="groupID" @close="closeMembers" />
                        </div>

                        <GroupTabs :active-tab="activeTab" @change="setActiveTab" />
                    </div>

                    <template v-if="activeTab === 'group'">
                        <GroupFeedHeader @add-post="openPostDialog" />

                        <div class="posts-feed">
                            <GroupFeedTab :posts="posts" :current-user-id="currentUserId" />

                            <div v-if="loadingPosts" class="loading-more">Loading posts...</div>
                        </div>
                    </template>

                    <GroupEventsTab v-else-if="activeTab === 'events'" ref="eventsTab" :group-id="groupID" />

                    <GroupChatTab v-else :group-i-d="groupID" :user-i-d="currentUserId" :group="group"
                        @event-created="handleEventCreated" />
                </div>
            </main>
        </div>

        <GroupDialoge :show="showPostDialog" :group-id="groupID" :group-title="group.title" @close="closePostDialog"
            @created="handlePostCreated" />
    </div>
</template>

<style scoped>
.group-page {
    min-height: 100vh;
    min-height: 100dvh;
    padding-top: 64px;
}

.page-layout {
    display: flex;
    align-items: flex-start;
    min-height: calc(100vh - 64px);
    min-height: calc(100dvh - 64px);
}

.main-content {
    flex: 1;
    min-width: 0;
}

.content-container {
    display: flex;
    width: 100%;
    max-width: 760px;
    flex-direction: column;
    gap: 22px;
    margin: 0 auto;
    padding: 30px 25px 60px;
}

.sticky-top {
    position: sticky;
    top: 64px;
    z-index: 100;

    display: flex;
    flex-direction: column;
    gap: 22px;

    margin: -12px -10px -10px 0;
    padding: 12px 10px 10px 0;

    background: var(--page-background);
}

.header-wrapper {
    position: relative;
}

.posts-feed {
    display: flex;
    flex-direction: column;
    gap: 18px;
}

.loading-more {
    padding: 14px;
    color: var(--font-color-sub);
    text-align: center;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

@media (max-width: 1024px) {
    .content-container {
        max-width: 100%;
        padding: 26px 20px 50px;
    }
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .content-container {
        gap: 18px;
        padding: 20px 15px 50px;
    }

    .sticky-top {
        gap: 18px;
    }
}

@media (max-width: 520px) {
    .group-page {
        padding-top: 64px;
    }

    .content-container {
        gap: 14px;
        padding: 14px 10px 44px;
    }

    .sticky-top {
        gap: 14px;
    }
}
</style>