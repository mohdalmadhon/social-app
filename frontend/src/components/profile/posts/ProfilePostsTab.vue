<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue';
import ProfilePostGrid from './ProfilePostGrid.vue';
import { getUserPosts } from '@/api/posts/posts';
import { normalizeProfilePost } from '@/helpers/common/locationHelpers';
import { useRoute } from 'vue-router';

const props = defineProps({
    userId: {
        type: [Number, String],
    },
    currentUserId: {
        type: [Number, String],
        default: null
    }
});

const BATCH_SIZE = 9;

const route = useRoute();
const activeTab = ref('posts');

const posts = ref([]);
const videos = ref([]);
const taggedPosts = ref([]);

const offset = ref(0);
const loading = ref(false);
const hasMore = ref(true);
const error = ref('');

const videosOffset = ref(0);
const videosLoading = ref(false);
const videosHasMore = ref(true);
const videosError = ref('');

const taggedOffset = ref(0);
const taggedLoading = ref(false);
const taggedHasMore = ref(true);
const taggedError = ref('');

const sentinel = ref(null);

let observer = null;

const tabs = [
    {
        id: 'posts',
        label: 'Posts'
    },
    {
        id: 'videos',
        label: 'Videos'
    },
    {
        id: 'tagged',
        label: 'Tagged'
    }
];

const currentItems = computed(() => {
    if (activeTab.value === 'videos') {
        return videos.value;
    }

    if (activeTab.value === 'tagged') {
        return taggedPosts.value;
    }

    return posts.value;
});

function getTargetID() {
    if (route.path === '/me') {
        return '';
    }

    if (route.path === '/user') {
        return route.query.id || '';
    }

    return '';
}

async function loadMorePosts() {
    if (loading.value || !hasMore.value || activeTab.value !== 'posts') {
        return;
    }

    loading.value = true;
    error.value = '';

    try {
        const targetID = getTargetID();
        const response = await getUserPosts(targetID, offset.value, BATCH_SIZE);
        const rawPosts = response?.data || [];
        const normalized = rawPosts.map(normalizeProfilePost);

        posts.value.push(...normalized);
        offset.value += BATCH_SIZE;

        if (typeof response?.hasMore === 'boolean') {
            hasMore.value = response.hasMore;
        } else if (normalized.length < BATCH_SIZE) {
            hasMore.value = false;
        }
    } catch (err) {
        error.value = err.message || 'Failed to load posts';
    } finally {
        loading.value = false;
    }
}

async function loadMoreVideos() {
    if (videosLoading.value || !videosHasMore.value || activeTab.value !== 'videos') {
        return;
    }

    videosLoading.value = true;
    videosError.value = '';

    try {
        const targetID = getTargetID();
        const response = await getUserPosts(targetID, videosOffset.value, BATCH_SIZE, 'videos');
        const rawVideos = response?.data || [];
        const normalized = rawVideos.map(normalizeProfilePost);

        const knownIds = new Set(videos.value.map(video => video.id));
        videos.value.push(...normalized.filter(video => !knownIds.has(video.id)));
        videosOffset.value += BATCH_SIZE;

        if (typeof response?.hasMore === 'boolean') {
            videosHasMore.value = response.hasMore;
        } else if (normalized.length < BATCH_SIZE) {
            videosHasMore.value = false;
        }
    } catch (err) {
        videosError.value = err.message || 'Failed to load videos';
    } finally {
        videosLoading.value = false;
    }
}

async function loadMoreTagged() {
    if (taggedLoading.value || !taggedHasMore.value || activeTab.value !== 'tagged') {
        return;
    }

    taggedLoading.value = true;
    taggedError.value = '';

    try {
        const targetID = getTargetID();
        const response = await getUserPosts(targetID, taggedOffset.value, BATCH_SIZE, 'tagged');
        const rawPosts = response?.data || [];
        const normalized = rawPosts.map(normalizeProfilePost);

        const knownIds = new Set(taggedPosts.value.map(post => post.id));
        taggedPosts.value.push(...normalized.filter(post => !knownIds.has(post.id)));
        taggedOffset.value += BATCH_SIZE;

        if (typeof response?.hasMore === 'boolean') {
            taggedHasMore.value = response.hasMore;
        } else if (rawPosts.length < BATCH_SIZE) {
            taggedHasMore.value = false;
        }
    } catch (err) {
        taggedError.value = err.message || 'Failed to load tagged posts';
    } finally {
        taggedLoading.value = false;
    }
}

function loadMoreActive() {
    if (activeTab.value === 'posts') {
        loadMorePosts();
    } else if (activeTab.value === 'videos') {
        loadMoreVideos();
    } else if (activeTab.value === 'tagged') {
        loadMoreTagged();
    }
}

function handlePostDeleted(postId) {
    posts.value = posts.value.filter(post => post.id !== postId);
    taggedPosts.value = taggedPosts.value.filter(post => post.id !== postId);
    videos.value = videos.value.filter(post => post.id !== postId);
}

function switchTab(tab) {
    if (activeTab.value === tab) {
        return;
    }

    activeTab.value = tab;

    if (tab === 'posts' && posts.value.length === 0 && hasMore.value) {
        loadMorePosts();
    }

    if (tab === 'videos' && videos.value.length === 0 && videosHasMore.value) {
        loadMoreVideos();
    }

    if (tab === 'tagged' && taggedPosts.value.length === 0 && taggedHasMore.value) {
        loadMoreTagged();
    }
}

function setupObserver() {
    observer = new IntersectionObserver(
        entries => {
            if (entries[0].isIntersecting) {
                loadMoreActive();
            }
        },
        {
            rootMargin: '200px'
        }
    );

    if (sentinel.value) {
        observer.observe(sentinel.value);
    }
}

onMounted(() => {
    loadMorePosts();
    setupObserver();
});

onBeforeUnmount(() => {
    observer?.disconnect();
});
</script>

<template>
    <div class="profile-posts-page">
        <div class="profile-posts-body">
            <main class="profile-posts-content">
                <nav class="profile-tabs">
                    <button v-for="tab in tabs" :key="tab.id" type="button" class="profile-tab"
                        :class="{ active: activeTab === tab.id }" @click="switchTab(tab.id)">
                        <svg v-if="tab.id === 'posts'" class="tab-icon" viewBox="0 0 24 24" fill="none"
                            xmlns="http://www.w3.org/2000/svg">
                            <rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.8" />
                            <rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.8" />
                            <rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor" stroke-width="1.8" />
                            <rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor"
                                stroke-width="1.8" />
                        </svg>

                        <svg v-else-if="tab.id === 'videos'" class="tab-icon" viewBox="0 0 24 24" fill="none"
                            xmlns="http://www.w3.org/2000/svg">
                            <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor"
                                stroke-width="1.8" />
                            <path d="M10 8.5L16 12L10 15.5V8.5Z" stroke="currentColor" stroke-width="1.8"
                                stroke-linejoin="round" />
                        </svg>

                        <svg v-else class="tab-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path
                                d="M20 12C20 16.4183 16.4183 20 12 20C7.58172 20 4 16.4183 4 12C4 7.58172 7.58172 4 12 4C16.4183 4 20 7.58172 20 12Z"
                                stroke="currentColor" stroke-width="1.8" />
                            <path d="M8 12.5L10.5 15L16 9.5" stroke="currentColor" stroke-width="1.8"
                                stroke-linecap="round" stroke-linejoin="round" />
                        </svg>

                        <span>{{ tab.label }}</span>
                    </button>
                </nav>

                <section v-if="activeTab === 'posts'" class="profile-tab-content">
                    <ProfilePostGrid :posts="posts" :current-user-id="currentUserId" @deleted="handlePostDeleted" />

                    <div v-if="error" class="profile-posts-error">
                        {{ error }}
                    </div>

                    <div v-if="loading" class="profile-posts-loading">
                        Loading posts...
                    </div>

                    <div v-else-if="!hasMore && posts.length === 0" class="profile-posts-empty">
                        No posts yet.
                    </div>

                </section>

                <section v-else-if="activeTab === 'videos'" class="profile-tab-content">
                    <ProfilePostGrid v-if="videos.length > 0" :posts="videos" :current-user-id="currentUserId"
                        @deleted="handlePostDeleted" />

                    <div v-if="videosError" class="profile-posts-error">
                        {{ videosError }}
                    </div>

                    <div v-if="videosLoading" class="profile-posts-loading">
                        Loading videos...
                    </div>

                    <div v-else-if="!videosHasMore && videos.length === 0" class="empty-tab">
                        <div class="empty-icon">
                            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor"
                                    stroke-width="1.8" />
                                <path d="M10 8.5L16 12L10 15.5V8.5Z" stroke="currentColor" stroke-width="1.8"
                                    stroke-linejoin="round" />
                            </svg>
                        </div>
                        <h3>No videos yet</h3>
                        <p>Videos shared by this user will appear here.</p>
                    </div>
                </section>

                <section v-else class="profile-empty-section">
                    <div v-if="taggedPosts.length > 0">
                        <ProfilePostGrid :posts="taggedPosts" :current-user-id="currentUserId"
                            @deleted="handlePostDeleted" />
                    </div>

                    <div v-if="taggedError" class="profile-posts-error">
                        {{ taggedError }}
                    </div>

                    <div v-if="taggedLoading" class="profile-posts-loading">
                        Loading tagged posts...
                    </div>

                    <div v-else-if="!taggedHasMore && taggedPosts.length === 0 && !taggedError" class="empty-tab">
                        <div class="empty-icon">
                            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path
                                    d="M20 12C20 16.4183 16.4183 20 12 20C7.58172 20 4 16.4183 4 12C4 7.58172 7.58172 4 12 4C16.4183 4 20 7.58172 20 12Z"
                                    stroke="currentColor" stroke-width="1.8" />
                                <path d="M8 12.5L10.5 15L16 9.5" stroke="currentColor" stroke-width="1.8"
                                    stroke-linecap="round" stroke-linejoin="round" />
                            </svg>
                        </div>
                        <h3>No tagged posts</h3>
                        <p>Posts that tag this user will appear here.</p>
                    </div>
                </section>

                <div ref="sentinel" class="profile-posts-sentinel"></div>
            </main>
        </div>
    </div>
</template>

<style scoped>
.profile-posts-page {
    width: 100%;
    min-height: 100vh;
    min-height: 100dvh;
    background: var(--page-background);
}

.profile-posts-body {
    display: flex;
    padding-top: 64px;
}

.profile-posts-content {
    flex: 1;
    min-width: 0;
    padding: 0 clamp(14px, 3vw, 32px) 24px;
}

.profile-tabs {
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: stretch;
    gap: 8px;
    border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
    margin-bottom: 24px;
}

.profile-tab {
    position: relative;
    min-width: 120px;
    padding: 16px 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 9px;
    border: none;
    background: transparent;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition:
        color 0.2s ease,
        background 0.2s ease;
}

.profile-tab:hover {
    color: var(--font-color);
    background: rgba(0, 0, 0, 0.025);
}

.profile-tab.active {
    color: var(--font-color);
}

.profile-tab.active::after {
    content: "";
    position: absolute;
    left: 15%;
    right: 15%;
    bottom: -1px;
    height: 2px;
    background: var(--main-color);
    border-radius: 2px 2px 0 0;
}

.tab-icon {
    width: 17px;
    height: 17px;
    flex-shrink: 0;
}

.profile-tab-content {
    width: 100%;
}

.profile-posts-sentinel {
    height: 1px;
}

.profile-posts-loading,
.profile-posts-empty {
    padding: 24px 0;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.profile-posts-error {
    margin-top: 16px;
    padding: 10px 14px;
    border: 1px solid var(--main-color);
    border-radius: 6px;
    background: #ffe9e9;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.profile-empty-section {
    width: 100%;
}

.empty-tab {
    min-height: 360px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
}

.empty-icon {
    width: 58px;
    height: 58px;
    margin-bottom: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border-color, rgba(0, 0, 0, 0.12));
    border-radius: 50%;
    color: var(--font-color-sub);
}

.empty-icon svg {
    width: 27px;
    height: 27px;
}

.empty-tab h3 {
    margin: 0 0 8px;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 13px;
    font-weight: 600;
}

.empty-tab p {
    max-width: 320px;
    margin: 0;
    line-height: 1.7;
    font-size: 10px;
}

@media (max-width: 800px) {
    .profile-posts-body {
        flex-direction: column;
    }

    .profile-posts-content {
        padding-left: 10px;
        padding-right: 10px;
    }

    .profile-tabs {
        gap: 0;
        margin-bottom: 18px;
    }

    .profile-tab {
        flex: 1;
        min-width: 0;
        padding: 14px 8px;
        font-size: 9px;
    }

    .profile-tab.active::after {
        left: 20%;
        right: 20%;
    }

    .tab-icon {
        width: 15px;
        height: 15px;
    }
}

@media (max-width: 450px) {
    .profile-tab {
        gap: 6px;
        padding: 13px 4px;
        font-size: 8px;
    }

    .tab-icon {
        width: 14px;
        height: 14px;
    }
}
</style>
