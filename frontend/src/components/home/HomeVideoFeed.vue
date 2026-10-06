<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue';
import HomeVideoItem from './HomeVideoItem.vue';
import { getHomeVideos } from '@/api/posts/posts';

defineProps({
    currentUserId: {
        type: [Number, String],
        default: null
    }
});

const BATCH_SIZE = 5;

const videos = ref([]);
const offset = ref(0);
const loading = ref(false);
const hasMore = ref(true);
const error = ref('');
const activeId = ref(null);
const muted = ref(true);
const scroller = ref(null);

const activeIndex = computed(() => {
    return videos.value.findIndex(video => video.id === activeId.value);
});

async function loadMore() {
    if (loading.value || !hasMore.value) {
        return;
    }

    loading.value = true;
    error.value = '';

    try {
        const response = await getHomeVideos(offset.value, BATCH_SIZE);
        const incoming = response?.posts || [];

        const known = new Set(videos.value.map(video => video.id));
        const fresh = incoming.filter(video => !known.has(video.id));

        videos.value.push(...fresh);
        offset.value += BATCH_SIZE;

        if (typeof response?.hasMore === 'boolean') {
            hasMore.value = response.hasMore;
        } else if (incoming.length < BATCH_SIZE) {
            hasMore.value = false;
        }

        if (activeId.value === null && videos.value.length > 0) {
            activeId.value = videos.value[0].id;
        }
    } catch (err) {
        error.value = err.message || 'Failed to load videos';
    } finally {
        loading.value = false;
    }
}

function handleVisible(id) {
    activeId.value = id;
}

function toggleMute() {
    muted.value = !muted.value;
}

function goTo(index) {
    const container = scroller.value;

    if (!container) {
        return;
    }

    const target = container.children[index];

    if (target) {
        target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
}

function handleKeydown(event) {
    const tag = event.target?.tagName;

    if (tag === 'INPUT' || tag === 'TEXTAREA' || event.target?.isContentEditable) {
        return;
    }

    if (document.querySelector('.comments-overlay')) {
        return;
    }

    if (event.key === 'ArrowDown' || event.key === 'j') {
        event.preventDefault();
        goTo(Math.min(videos.value.length - 1, activeIndex.value + 1));
    }

    if (event.key === 'ArrowUp' || event.key === 'k') {
        event.preventDefault();
        goTo(Math.max(0, activeIndex.value - 1));
    }
}

watch(activeIndex, index => {
    if (index >= 0 && index >= videos.value.length - 2) {
        loadMore();
    }
});

onMounted(async () => {
    window.addEventListener('keydown', handleKeydown);

    await loadMore();
    await nextTick();
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
    <div class="video-feed">
        <div ref="scroller" class="video-scroller">
            <HomeVideoItem v-for="video in videos" :key="video.id" :post="video" :active="video.id === activeId"
                :muted="muted" :current-user-id="currentUserId" @visible="handleVisible"
                @toggle-mute="toggleMute" />

            <div v-if="loading && videos.length > 0" class="video-status end-slide">
                Loading videos...
            </div>

            <div v-else-if="!hasMore && videos.length > 0" class="video-status end-slide">
                You're all caught up.
            </div>
        </div>

        <div v-if="loading && videos.length === 0" class="video-status centered">
            Loading videos...
        </div>

        <div v-else-if="!loading && !hasMore && videos.length === 0 && !error" class="video-empty">
            <div class="video-empty-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor" stroke-width="1.8" />
                    <path d="M10 8.5L16 12L10 15.5V8.5Z" stroke="currentColor" stroke-width="1.8"
                        stroke-linejoin="round" />
                </svg>
            </div>
            <h3>No videos yet</h3>
            <p>Videos shared by people will appear here.</p>
        </div>

        <div v-if="error" class="video-error">
            {{ error }}
            <button type="button" @click="loadMore">Retry</button>
        </div>
    </div>
</template>

<style scoped>
.video-feed {
    position: relative;
    width: 100%;
    height: 100%;
}

.video-scroller {
    width: 100%;
    height: 100%;
    overflow-y: auto;
    overscroll-behavior: contain;
    scroll-snap-type: y mandatory;
    scrollbar-width: none;
}

.video-scroller::-webkit-scrollbar {
    display: none;
}

.video-status {
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    text-align: center;
}

.video-status.end-slide {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    scroll-snap-align: start;
}

.video-status.centered {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
}

.video-empty {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
}

.video-empty-icon {
    width: 58px;
    height: 58px;
    margin-bottom: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border-color, rgba(0, 0, 0, 0.12));
    border-radius: 50%;
}

.video-empty-icon svg {
    width: 27px;
    height: 27px;
}

.video-empty h3 {
    margin: 0 0 8px;
    color: var(--font-color);
    font-size: 13px;
    font-weight: 600;
}

.video-empty p {
    margin: 0;
    font-size: 10px;
}

.video-error {
    position: absolute;
    left: 50%;
    bottom: 16px;
    z-index: 5;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    border: 1px solid var(--main-color);
    border-radius: 6px;
    background: #ffe9e9;
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.video-error button {
    padding: 4px 10px;
    border: 1px solid var(--main-color);
    border-radius: 4px;
    background: transparent;
    color: var(--main-color);
    font-family: inherit;
    font-size: 11px;
    cursor: pointer;
}
</style>
