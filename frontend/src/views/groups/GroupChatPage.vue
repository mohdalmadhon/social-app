<script setup>
import GroupChatTab from '@/components/groups/GroupChatTab.vue';
import GroupEventsTab from '@/components/groups/GroupEventsTab.vue';
import GroupFeedTab from '@/components/groups/GroupFeedTab.vue';
import GroupHeader from '@/components/groups/GroupHeader.vue';
import GroupMembersPanel from '@/components/groups/GroupMembersPanel.vue';
import GroupTabs from '@/components/groups/GroupTabs.vue';
import GroupFeedHeader from '@/components/groups/GroupFeedHeader.vue';

import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';

import { addNotification } from '@/data/notifications';

import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import GroupDialoge from '@/components/groups/GroupDialoge.vue';

const group = ref({});
const members = ref([]);
const route = useRoute();

const groupID = Number(route.params.id);

const posts = ref([]);
const events = ref([]);
const messages = ref([]);

const activeTab = ref('group');
const showMembers = ref(false);
const showPostDialog = ref(false);

function toggleMembers() {
    showMembers.value = !showMembers.value;
}

function closeMembers() {
    showMembers.value = false;
}

function setActiveTab(tab) {
    activeTab.value = tab;
}

function openPostDialog() {
    showPostDialog.value = true;
}

function closePostDialog() {
    showPostDialog.value = false;
}

async function handlePostCreated() {
    showPostDialog.value = false;
    await getGroupPosts();
}

async function getGroupData() {
    try {
        const resp = await fetch(
            `/api/group?groupID=${groupID}`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const result = await resp.json();

        if (!resp.ok) {
            addNotification(
                result.message ||
                result.messages ||
                'Could not get group data',
                'error'
            );
            return;
        }

        if (!result.status) {
            addNotification(
                result.message ||
                result.messages ||
                'Could not get group data',
                'error'
            );
            return;
        }

        group.value = {
            ...result.data
        };
    } catch (err) {
        addNotification(
            err.message || 'Could not get group data',
            'error'
        );
    }
}

async function getGroupPosts() {
    try {
        const resp = await fetch(
            `/api/group/posts?groupID=${groupID}&offset=0`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const result = await resp.json();

        if (!resp.ok || !result.status) {
            addNotification(
                result.message ||
                'Could not get group posts',
                'error'
            );
            return;
        }

        posts.value =
            result.data ||
            result.posts ||
            [];
    } catch (err) {
        addNotification(
            err.message ||
            'Could not get group posts',
            'error'
        );
    }
}

onMounted(async () => {
    await getGroupData();
    await getGroupPosts();
});
</script>

<template>
    <div class="group-page">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="main-content">
                <div class="content-container">

                    <div class="header-wrapper">
                        <GroupHeader :name="group.title" :description="group.description"
                            :avatar-path="group.avatarPath" :members-count="group.Count" :show-members="showMembers"
                            @toggle-members="toggleMembers" />

                        <GroupMembersPanel :show="showMembers" :members="members" :group-i-d="groupID"
                            @close="closeMembers" />
                    </div>

                    <GroupTabs :active-tab="activeTab" @change="setActiveTab" />

                    <template v-if="activeTab === 'group'">
                        <GroupFeedHeader @add-post="openPostDialog" />

                        <GroupFeedTab :posts="posts" />
                    </template>

                    <GroupEventsTab v-else-if="activeTab === 'events'" :events="events" />

                    <GroupChatTab v-else :messages="messages" :group-i-d="groupID" />
                </div>
            </main>
        </div>

        <GroupDialoge
            :show="showPostDialog"
            :group-id="groupID"
            :group-title="group.title"
            @close="closePostDialog"
            @created="handlePostCreated"
        />
    </div>
</template>

<style scoped>
.group-page {
    min-height: 100vh;
    padding-top: 64px;
}

.page-layout {
    display: flex;
    align-items: flex-start;
    min-height: calc(100vh - 64px);
}

.main-content {
    flex: 1;
    min-width: 0;
}

.content-container {
    width: 100%;
    max-width: 760px;
    margin: 0 auto;
    padding: 30px 25px 60px;
    display: flex;
    flex-direction: column;
    gap: 22px;
}

.header-wrapper {
    position: relative;
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .content-container {
        padding: 20px 15px 50px;
    }
}
</style>