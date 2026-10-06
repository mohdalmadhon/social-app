<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import HomePosts from '@/components/home/HomePosts.vue';
import HomeSearch from '@/components/home/HomeSearch.vue';
import HomeVideoFeed from '@/components/home/HomeVideoFeed.vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';

const activeTab = ref('feed');
const posts = ref([]);
const loading = ref(false);
const hasMore = ref(true);
const offset = ref(0);

let throttleTimeout = null;

function getTaggedPeople(post) {
    if (Array.isArray(post?.taggedPeople)) {
        return post.taggedPeople;
    }

    if (Array.isArray(post?.TaggedPeople)) {
        return post.TaggedPeople;
    }

    if (Array.isArray(post?.tagged_people)) {
        return post.tagged_people;
    }

    return [];
}

async function loadPosts() {
    if (loading.value || !hasMore.value) {
        return;
    }

    loading.value = true;

    try {
        const response = await fetch(
            `/api/posts?offset=${offset.value}`,
            {
                method: 'GET',
                credentials: 'include'
            }
        );

        const data = await response.json();

        console.log('API response:', data);

        if (!response.ok || !data.status) {
            return;
        }

        const newPosts = data.posts || [];

        console.log('Posts:', newPosts);

        newPosts.forEach(post => {
            console.log(
                'Post:',
                post.id,
                'tagged people:',
                getTaggedPeople(post)
            );
        });

        posts.value.push(...newPosts);
        offset.value += newPosts.length;

        if (newPosts.length < 13) {
            hasMore.value = false;
        }
    } catch (error) {
        console.error(error);
    } finally {
        loading.value = false;
    }
}

function switchTab(tab) {
    activeTab.value = tab;

    if (tab === 'feed') {
        window.scrollTo({ top: 0 });
    }
}

function handleScroll() {
    if (activeTab.value !== 'feed' || throttleTimeout) {
        return;
    }

    throttleTimeout = setTimeout(() => {
        throttleTimeout = null;

        const scrollPosition =
            window.innerHeight + window.scrollY;

        const pageHeight =
            document.documentElement.scrollHeight;

        if (scrollPosition >= pageHeight - 500) {
            loadPosts();
        }
    }, 200);
}

onMounted(() => {
    loadPosts();

    window.addEventListener(
        'scroll',
        handleScroll,
        { passive: true }
    );
});

onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll);

    if (throttleTimeout) {
        clearTimeout(throttleTimeout);
        throttleTimeout = null;
    }
});
</script>

<template>
    <div class="home-page">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="main-content">
                <div class="content-container" :class="{ 'videos-mode': activeTab === 'videos' }">
                    <nav class="home-tabs">
                        <button type="button" class="home-tab" :class="{ active: activeTab === 'feed' }"
                            @click="switchTab('feed')">
                            <svg class="tab-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <rect x="3" y="3" width="7" height="7" rx="1.5" stroke="currentColor"
                                    stroke-width="1.8" />
                                <rect x="14" y="3" width="7" height="7" rx="1.5" stroke="currentColor"
                                    stroke-width="1.8" />
                                <rect x="3" y="14" width="7" height="7" rx="1.5" stroke="currentColor"
                                    stroke-width="1.8" />
                                <rect x="14" y="14" width="7" height="7" rx="1.5" stroke="currentColor"
                                    stroke-width="1.8" />
                            </svg>
                            <span>Feed</span>
                        </button>

                        <button type="button" class="home-tab" :class="{ active: activeTab === 'videos' }"
                            @click="switchTab('videos')">
                            <svg class="tab-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor"
                                    stroke-width="1.8" />
                                <path d="M10 8.5L16 12L10 15.5V8.5Z" stroke="currentColor" stroke-width="1.8"
                                    stroke-linejoin="round" />
                            </svg>
                            <span>Videos</span>
                        </button>
                    </nav>

                    <template v-if="activeTab === 'feed'">
                        <HomeSearch />

                        <section class="posts">
                            <HomePosts
                                :auto="true"
                                v-for="post in posts"
                                :key="post.id"
                                v-bind="post"
                                :post-id="post.id"
                                :likes="post.likeCount"
                                :dislikes="post.disLikeCount"
                                :reaction="post.ReactionValue"
                                :allow-comments="post.allowComments"
                                :avatar-path="post.avatarPath"
                                :tagged-people="getTaggedPeople(post)"
                            />
                        </section>

                        <div
                            v-if="loading"
                            class="loading"
                        >
                            Loading posts...
                        </div>

                        <div
                            v-else-if="!hasMore && posts.length"
                            class="end-message"
                        >
                            You're all caught up.
                        </div>
                    </template>

                    <section v-else class="videos-wrap">
                        <HomeVideoFeed />
                    </section>
                </div>
            </main>
        </div>
    </div>
</template>

<style scoped>
.home-page {
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
}

.home-tabs {
    height: 48px;
    display: flex;
    justify-content: center;
    align-items: stretch;
    gap: 8px;
    border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
    margin-bottom: 24px;
}

.home-tab {
    position: relative;
    min-width: 120px;
    padding: 0 22px;
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

.home-tab:hover {
    color: var(--font-color);
    background: rgba(0, 0, 0, 0.025);
}

.home-tab.active {
    color: var(--font-color);
}

.home-tab.active::after {
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

.content-container.videos-mode {
    max-width: 520px;
    padding-top: 0;
    padding-bottom: 0;
}

.content-container.videos-mode .home-tabs {
    margin-bottom: 0;
}

.videos-wrap {
    height: calc(100vh - 64px - 48px);
    height: calc(100dvh - 64px - 48px);
}

.posts {
    display: flex;
    flex-direction: column;
    gap: 28px;
    margin-top: 30px;
}

.loading,
.end-message {
    padding: 20px;
    text-align: center;
    color: var(--font-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
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

    .content-container.videos-mode {
        padding-left: 0;
        padding-right: 0;
    }

    .home-tab {
        flex: 1;
        min-width: 0;
        padding: 0 8px;
        font-size: 9px;
    }
}
</style>