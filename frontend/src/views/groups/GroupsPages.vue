<script setup>
import { getGroupChats } from '@/api/chats/chats';
import GroupCard from '@/components/groups/GroupCard.vue';
import GroupCreateDialog from '@/components/groups/GroupCreateDialog.vue';
import GroupsSearch from '@/components/groups/Groupssearch.vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import { groupTypingCount } from '@/data/typingState';
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

const currentGroups = ref([]);
const discoverGroups = ref([]);
const searchQuery = ref('');

const loadingCurrent = ref(false);
const loadingDiscover = ref(false);

const hasMoreCurrent = ref(true);
const hasMoreDiscover = ref(true);

const pendingGroupIds = ref(new Set());

let currentOffset = 0;
let discoverOffset = 0;

let scrollTimeout = null;
let searchTimeout = null;

const scrollThrottle = 200;
const searchDebounce = 400;

async function loadCurrentGroups() {
    if (loadingCurrent.value || !hasMoreCurrent.value) {
        return;
    }

    loadingCurrent.value = true;

    try {
        const result = await getGroupChats(currentOffset);
        console.log(result)
        if (!result.status) {
            throw new Error(result.message || 'Could not get groups');
        }

        const groups = result.data || [];

        const existingIds = new Set(
            currentGroups.value.map(group => Number(group.ID))
        );

        const newGroups = groups.filter(
            group => !existingIds.has(Number(group.ID))
        );

        currentGroups.value.push(...newGroups);

        currentOffset += groups.length;

        if (groups.length < 10) {
            hasMoreCurrent.value = false;
        }
    } catch (error) {
        console.error(error);
    } finally {
        loadingCurrent.value = false;
    }
}

async function loadDiscoverGroups() {
    if (loadingDiscover.value || !hasMoreDiscover.value) {
        return;
    }

    loadingDiscover.value = true;

    try {
        const params = new URLSearchParams();

        params.set('offset', discoverOffset.toString());

        const search = searchQuery.value.trim();

        if (search) {
            params.set('search', search);
        }

        const response = await fetch(
            `/api/groups/discover?${params.toString()}`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const result = await response.json();

        if (!response.ok || !result.status) {
            throw new Error(
                result.message || 'Could not get discover groups'
            );
        }

        const groups = result.data || [];

        const existingIds = new Set(
            discoverGroups.value.map(group => Number(group.ID))
        );

        const newGroups = groups.filter(
            group => !existingIds.has(Number(group.ID))
        );

        discoverGroups.value.push(...newGroups);

        discoverOffset += groups.length;

        if (groups.length < 12) {
            hasMoreDiscover.value = false;
        }
    } catch (error) {
        console.error(error);
    } finally {
        loadingDiscover.value = false;
    }
}

function searchGroups() {
    discoverGroups.value = [];
    discoverOffset = 0;
    hasMoreDiscover.value = true;

    loadDiscoverGroups();
}

watch(searchQuery, () => {
    if (searchTimeout !== null) {
        clearTimeout(searchTimeout);
    }

    searchTimeout = setTimeout(() => {
        searchGroups();
    }, searchDebounce);
});

function handleScroll() {
    if (scrollTimeout !== null) {
        return;
    }

    scrollTimeout = setTimeout(() => {
        scrollTimeout = null;

        const scrollPosition =
            window.scrollY + window.innerHeight;

        const pageHeight =
            document.documentElement.scrollHeight;

        if (scrollPosition >= pageHeight - 300) {
            loadCurrentGroups();
            loadDiscoverGroups();
        }
    }, scrollThrottle);
}

async function joinGroup(groupId) {
    if (pendingGroupIds.value.has(groupId)) {
        return;
    }

    pendingGroupIds.value.add(groupId);

    try {
        const response = await fetch(
            `/api/groups/${groupId}/join`,
            {
                method: 'POST',
                credentials: 'include'
            }
        );

        const data = await response.json();

        if (!response.ok || !data.status) {
            return;
        }

        const index = discoverGroups.value.findIndex(
            group => Number(group.ID) === Number(groupId)
        );

        if (index !== -1) {
            const [joinedGroup] =
                discoverGroups.value.splice(index, 1);

            joinedGroup.Count =
                (joinedGroup.Count || 0) + 1;

            const alreadyJoined =
                currentGroups.value.some(
                    group =>
                        Number(group.ID) === Number(groupId)
                );

            if (!alreadyJoined) {
                currentGroups.value.unshift(joinedGroup);
            }
        }
    } catch (error) {
        console.error(error);
    } finally {
        pendingGroupIds.value.delete(groupId);
    }
}

function openGroup(groupId) {
    router.push(`/groups/${groupId}`);
}

function openCreateDialog() {
    showCreateDialog.value = true;
}

function closeCreateDialog() {
    showCreateDialog.value = false;
}

function handleGroupCreated() {
    currentGroups.value = [];
    currentOffset = 0;
    hasMoreCurrent.value = true;

    loadCurrentGroups();
}

const showCreateDialog = ref(false);

onMounted(() => {
    loadCurrentGroups();
    loadDiscoverGroups();

    window.addEventListener('scroll', handleScroll, {
        passive: true
    });
});

onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll);

    if (scrollTimeout !== null) {
        clearTimeout(scrollTimeout);
    }

    if (searchTimeout !== null) {
        clearTimeout(searchTimeout);
    }
});
</script>

<template>
    <div class="groups-page">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="main-content">
                <div class="content-container">
                    <BackToHome />

                    <div class="page-header">
                        <h1 class="page-title">
                            Groups
                        </h1>

                        <button class="create-group-button" type="button" @click="openCreateDialog">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M12 5v14"></path>
                                <path d="M5 12h14"></path>
                            </svg>

                            Create group
                        </button>
                    </div>

                    <section class="groups-section">
                        <h2 class="section-title">
                            Current groups
                        </h2>

                        <div v-if="loadingCurrent && !currentGroups.length" class="section-message">
                            Loading your groups...
                        </div>

                        <div v-else-if="!currentGroups.length" class="section-message">
                            You haven't joined any groups yet.
                        </div>

                        <div v-else class="groups-list">
                            <GroupCard v-for="group in currentGroups" :key="group.ID" :group-id="group.ID"
                                :name="group.title" :description="group.description" :avatar-path="group.avatar"
                                :members-count="group.Count || 0" :is-member="true"
                                :typing-count="groupTypingCount(group.ID)" :unread-count="group.unreadCount || 0"
                                @open="openGroup" />
                        </div>

                        <div v-if="loadingCurrent && currentGroups.length" class="section-message">
                            Loading more groups...
                        </div>

                        <div v-if="!hasMoreCurrent && currentGroups.length" class="section-message">
                            You've reached the end of your groups.
                        </div>
                    </section>

                    <section class="groups-section">
                        <h2 class="section-title">
                            Discover groups
                        </h2>

                        <GroupsSearch v-model="searchQuery" />

                        <div v-if="loadingDiscover && !discoverGroups.length" class="section-message">
                            Loading groups...
                        </div>

                        <div v-else-if="!discoverGroups.length" class="section-message">
                            No groups found.
                        </div>

                        <div v-else class="groups-list">
                            <GroupCard v-for="group in discoverGroups" :key="group.ID" :group-id="group.ID"
                                :name="group.title" :description="group.Description" :avatar-path="group.avatar"
                                :members-count="group.Count || 0" :is-member="false"
                                :is-pending="pendingGroupIds.has(group.ID)" @join="joinGroup" @open="openGroup" />
                        </div>

                        <div v-if="loadingDiscover && discoverGroups.length" class="section-message">
                            Loading more groups...
                        </div>

                        <div v-if="!hasMoreDiscover && discoverGroups.length" class="section-message">
                            You've reached the end of discover groups.
                        </div>
                    </section>
                </div>
            </main>
        </div>

        <GroupCreateDialog :show="showCreateDialog" @close="closeCreateDialog" @created="handleGroupCreated" />
    </div>
</template>

<style scoped>
.groups-page {
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
    width: 100%;
    max-width: 760px;
    min-width: 0;
    margin: 0 auto;
    padding: 30px 25px 60px;
    overflow-x: clip;
}

.page-header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 28px;
}

.page-title {
    margin: 0;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 28px;
    font-weight: 700;
}

.create-group-button {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
    box-shadow: 4px 4px var(--main-color);
    transition:
        transform 0.1s,
        box-shadow 0.1s,
        background 0.15s,
        color 0.15s;
}

.create-group-button svg {
    width: 16px;
    height: 16px;
}

.create-group-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.create-group-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.groups-section {
    margin-bottom: 36px;
}

.groups-section:last-child {
    margin-bottom: 0;
}

.section-title {
    margin: 0 0 16px;
    color: var(--font-color);
    font-family: "Liter", serif;
    font-size: 19px;
    font-weight: 700;
}

.groups-section .search-container {
    margin-bottom: 18px;
}

.groups-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
    padding-right: 6px;
    padding-bottom: 6px;
}

.section-message {
    padding: 20px;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .groups-list {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .content-container {
        padding: 20px 15px 50px;
    }
}

@media (max-width: 650px) {
    .content-container {
        padding-left: 10px;
        padding-right: 10px;
    }

    .page-header {
        flex-wrap: wrap;
    }

    .page-title {
        font-size: 22px;
    }

    .groups-list {
        grid-template-columns: minmax(0, 1fr);
    }

    .create-group-button {
        padding: 8px 14px;
        font-size: 12px;
    }
}
</style>