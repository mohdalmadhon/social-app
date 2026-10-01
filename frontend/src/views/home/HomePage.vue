```vue
<script setup>

import { ref, onMounted, onUnmounted } from 'vue';

import HomePosts from '@/components/home/HomePosts.vue';

import HomeSearch from '@/components/home/HomeSearch.vue';

import SideNavigation from '@/components/layout/SideNavigation.vue';

import TopNavigation from '@/components/layout/TopNavigation.vue';

const posts = ref([]);

const loading = ref(false);

const hasMore = ref(true);

const offset = ref(0);

let throttleTimeout = null;

async function loadPosts() {
    if (loading.value || !hasMore.value) {
        return;
    }

    loading.value = true;

    try {
        const response = await fetch(
            `/api/posts?offset=${offset.value}`,
            {
                credentials: 'include'
            }
        );

        const data = await response.json();

        console.log(data);

        if (!response.ok || !data.status) {
            return;
        }

        const newPosts = data.posts || [];

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

function handleScroll() {
    if (throttleTimeout) {
        return;
    }

    throttleTimeout = setTimeout(() => {
        throttleTimeout = null;

        const scrollPosition = window.innerHeight + window.scrollY;
        const pageHeight = document.documentElement.scrollHeight;

        if (scrollPosition >= pageHeight - 500) {
            loadPosts();
        }
    }, 200);
}

onMounted(() => {
    loadPosts();
    window.addEventListener('scroll', handleScroll, { passive: true });
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
                <div class="content-container">
                    <HomeSearch />

                    <section class="posts">
                        <HomePosts
                            v-for="post in posts"
                            :key="post.id"
                            :post-id="post.id"
                            :likes="post.likeCount"
                            :dislikes="post.disLikeCount"
                            :reaction="post.ReactionValue"
                            :allowComments="post.allowComments"
                            :avatar-path="post.avatarPath"
                            v-bind="post"
                        />
                    </section>

                    <div v-if="loading" class="loading">
                        Loading posts...
                    </div>

                    <div
                        v-else-if="!hasMore && posts.length"
                        class="end-message"
                    >
                        You're all caught up.
                    </div>
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
}

</style>
```
