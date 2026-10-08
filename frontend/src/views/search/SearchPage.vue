<script setup>
import { computed, onBeforeUnmount, reactive, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { searchGroups, searchPosts, searchUsers } from '@/api/search/search';
import HomePosts from '@/components/home/HomePosts.vue';
import HomeSearch from '@/components/home/HomeSearch.vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import SearchGroupCard from '@/components/search/SearchGroupCard.vue';
import SearchResultsPanel from '@/components/search/SearchResultsPanel.vue';
import SearchTabs from '@/components/search/SearchTabs.vue';
import SearchUserCard from '@/components/search/SearchUserCard.vue';

const TABS = [
    { id: 'groups', label: 'Groups' },
    { id: 'users', label: 'Users' },
    { id: 'posts', label: 'Posts' }
];

const fetchers = {
    groups: searchGroups,
    users: searchUsers,
    posts: searchPosts
};

const route = useRoute();
const router = useRouter();

const query = computed(() => {
    return typeof route.query.q === 'string' ? route.query.q.trim() : '';
});

function createTabState() {
    return {
        items: [],
        offset: 0,
        loading: false,
        loaded: false,
        hasMore: false,
        error: ''
    };
}

const results = reactive({
    groups: createTabState(),
    users: createTabState(),
    posts: createTabState()
});

const defaultTab = reactive({ id: 'groups' });

const activeTab = computed(() => {
    const requested = route.query.tab;

    if (TABS.some(tab => tab.id === requested)) {
        return requested;
    }

    return defaultTab.id;
});

const tabItems = computed(() => {
    return TABS.map(tab => {
        const state = results[tab.id];
        let count = '';

        if (state.loaded && state.items.length) {
            count = `${state.items.length}${state.hasMore ? '+' : ''}`;
        }

        return { ...tab, count };
    });
});

let searchToken = 0;

async function loadTab(id, token) {
    const state = results[id];

    if (state.loading) {
        return;
    }

    state.loading = true;
    state.error = '';

    try {
        const { data, hasMore } = await fetchers[id](
            query.value,
            state.offset
        );

        if (token !== searchToken) {
            return;
        }

        const knownIds = new Set(state.items.map(item => item.id));

        state.items.push(...data.filter(item => !knownIds.has(item.id)));
        state.offset += data.length;
        state.hasMore = hasMore;
        state.loaded = true;
    } catch (error) {
        if (token !== searchToken) {
            return;
        }

        state.error = error.message || 'could not search, please try again';
    } finally {
        if (token === searchToken) {
            state.loading = false;
        }
    }
}

async function runSearch() {
    const token = ++searchToken;

    for (const tab of TABS) {
        Object.assign(results[tab.id], createTabState());
    }

    defaultTab.id = TABS[0].id;

    if (!query.value) {
        return;
    }

    await Promise.all(TABS.map(tab => loadTab(tab.id, token)));

    if (token !== searchToken) {
        return;
    }

    if (results[defaultTab.id].items.length === 0) {
        const firstWithResults = TABS.find(tab => results[tab.id].items.length);

        if (firstWithResults) {
            defaultTab.id = firstWithResults.id;
        }
    }
}

function selectTab(id) {
    router.replace({
        path: '/search',
        query: { q: query.value, tab: id }
    });
}

function loadMore(id) {
    loadTab(id, searchToken);
}

watch(query, runSearch, { immediate: true });

onBeforeUnmount(() => {
    searchToken++;
});
</script>

<template>
    <div class="search-page">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="main-content">
                <div class="content-container">
                    <BackToHome />

                    <HomeSearch />

                    <p v-if="!query" class="hint">
                        Type something in the search box to look for groups,
                        users and posts.
                    </p>

                    <template v-else>
                        <h1 class="results-heading">
                            Results for “{{ query }}”
                        </h1>

                        <SearchTabs
                            :tabs="tabItems"
                            :active-tab="activeTab"
                            @change="selectTab"
                        />

                        <section
                            v-show="activeTab === 'groups'"
                            id="search-panel-groups"
                            class="panel"
                            role="tabpanel"
                            aria-labelledby="search-tab-groups"
                        >
                            <SearchResultsPanel
                                :state="results.groups"
                                noun="groups"
                                :query="query"
                                @retry="loadMore('groups')"
                                @load-more="loadMore('groups')"
                            >
                                <div class="cards">
                                    <SearchGroupCard
                                        v-for="group in results.groups.items"
                                        :key="group.id"
                                        :group="group"
                                    />
                                </div>
                            </SearchResultsPanel>
                        </section>

                        <section
                            v-show="activeTab === 'users'"
                            id="search-panel-users"
                            class="panel"
                            role="tabpanel"
                            aria-labelledby="search-tab-users"
                        >
                            <SearchResultsPanel
                                :state="results.users"
                                noun="users"
                                :query="query"
                                @retry="loadMore('users')"
                                @load-more="loadMore('users')"
                            >
                                <div class="cards">
                                    <SearchUserCard
                                        v-for="user in results.users.items"
                                        :key="user.id"
                                        :user="user"
                                    />
                                </div>
                            </SearchResultsPanel>
                        </section>

                        <section
                            v-show="activeTab === 'posts'"
                            id="search-panel-posts"
                            class="panel"
                            role="tabpanel"
                            aria-labelledby="search-tab-posts"
                        >
                            <SearchResultsPanel
                                :state="results.posts"
                                noun="posts"
                                :query="query"
                                @retry="loadMore('posts')"
                                @load-more="loadMore('posts')"
                            >
                                <div class="cards posts">
                                    <HomePosts
                                        v-for="post in results.posts.items"
                                        :key="post.id"
                                        :post-id="post.id"
                                        :likes="post.likeCount"
                                        :dislikes="post.disLikeCount"
                                        :reaction="post.ReactionValue"
                                        v-bind="post"
                                    />
                                </div>
                            </SearchResultsPanel>
                        </section>
                    </template>
                </div>
            </main>
        </div>
    </div>
</template>

<style scoped>
.search-page {
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
    margin: 0 auto;
    padding: 30px 25px 60px;
}

.results-heading {
    margin: 30px 0 20px;

    color: var(--font-color);

    font-family: "Liter", serif;
    font-size: 20px;
    font-weight: 700;

    overflow-wrap: anywhere;
}

.hint {
    margin: 40px 0 0;

    color: var(--font-color-sub);
    text-align: center;

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    line-height: 1.6;
}

.panel {
    margin-top: 28px;
}

.cards {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.cards.posts {
    gap: 28px;
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

    .results-heading {
        font-size: 17px;
    }
}
</style>
