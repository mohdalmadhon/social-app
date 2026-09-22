<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
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

const posts = ref([]);
const offset = ref(0);
const loading = ref(false);
const hasMore = ref(true);
const error = ref('');
const route = useRoute();
const sentinel = ref(null);
let observer = null;

async function loadMorePosts() {
    if (loading.value || !hasMore.value) {
        return;
    }

    loading.value = true;
    error.value = '';

    try {
        const targetID = route.path === '/me'
            ? ''
            : route.path === '/user'
                ? route.query.id || ''
                : '';

        const response = await getUserPosts(targetID, offset.value);
        const rawPosts = response?.data || [];
        const normalized = rawPosts.map(normalizeProfilePost);

        console.log(response);

        posts.value.push(...normalized);
        offset.value += BATCH_SIZE;

        if (normalized.length < BATCH_SIZE) {
            hasMore.value = false;
        }
    } catch (err) {
        error.value = err.message || 'Failed to load posts';
    } finally {
        loading.value = false;
    }
}

function handlePostDeleted(postId) {
    posts.value = posts.value.filter(post => post.id !== postId);
}

function setupObserver() {
    observer = new IntersectionObserver(entries => {
        if (entries[0].isIntersecting) {
            loadMorePosts();
        }
    }, { rootMargin: '200px' });

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
                <ProfilePostGrid :posts="posts" :current-user-id="currentUserId" @deleted="handlePostDeleted" />

                <div v-if="error" class="profile-posts-error">{{ error }}</div>
                <div v-if="loading" class="profile-posts-loading">Loading posts...</div>
                <div v-else-if="!hasMore && posts.length === 0" class="profile-posts-empty">No posts yet.</div>

                <div ref="sentinel" class="profile-posts-sentinel"></div>
            </main>
        </div>
    </div>
</template>

<style scoped>
.profile-posts-page {
    width: 100%;
    min-height: 100vh;

    background: var(--page-background);
}

.profile-posts-body {
    display: flex;

    padding-top: 64px;
}

.profile-posts-content {
    flex: 1;
    min-width: 0;

    padding: 24px clamp(14px, 3vw, 32px);
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

@media (max-width: 800px) {
    .profile-posts-body {
        flex-direction: column;
    }
}
</style>