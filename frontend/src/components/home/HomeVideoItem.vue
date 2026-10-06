<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import HomePostComments from './HomePostComments.vue';
import { postReaction } from '@/api/posts/actions';
import { viewPost } from '@/api/posts/posts';
import { addNotification } from '@/data/notifications';
import { Reaction } from '@/models/posts';

const props = defineProps({
    post: {
        type: Object,
        required: true
    },
    active: {
        type: Boolean,
        default: false
    },
    muted: {
        type: Boolean,
        default: true
    },
    currentUserId: {
        type: [Number, String],
        default: null
    }
});

const emit = defineEmits(['toggle-mute', 'visible']);

const itemElement = ref(null);
const videoElement = ref(null);
const playing = ref(false);
const progress = ref(0);
const flash = ref('');
const expanded = ref(false);
const showComments = ref(false);
const failed = ref(false);

const likeCount = ref(props.post.likeCount ?? 0);
const dislikeCount = ref(props.post.disLikeCount ?? 0);
const commentCount = ref(props.post.commentCount ?? 0);
const currentReaction = ref(props.post.ReactionValue ?? 0);

let observer = null;
let flashTimer = null;
let seen = false;

const videoUrl = computed(() => `/uploads/${props.post.imagePath}`);

const authorName = computed(() => {
    return `${props.post.firstName || ''} ${props.post.lastName || ''}`.trim();
});

const hasLongContent = computed(() => (props.post.content || '').length > 90);

function showFlash(kind) {
    flash.value = kind;

    clearTimeout(flashTimer);

    flashTimer = setTimeout(() => {
        flash.value = '';
    }, 600);
}

async function play() {
    const video = videoElement.value;

    if (!video) {
        return;
    }

    try {
        await video.play();
    } catch (err) {
        playing.value = false;
    }
}

function pause() {
    const video = videoElement.value;

    if (!video) {
        return;
    }

    video.pause();
}

function togglePlay() {
    const video = videoElement.value;

    if (!video) {
        return;
    }

    if (video.paused) {
        play();
        showFlash('play');
    } else {
        pause();
        showFlash('pause');
    }
}

function handleTimeUpdate() {
    const video = videoElement.value;

    if (!video || !video.duration) {
        return;
    }

    progress.value = (video.currentTime / video.duration) * 100;
}

function seek(event) {
    const video = videoElement.value;

    if (!video || !video.duration) {
        return;
    }

    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width));

    video.currentTime = ratio * video.duration;
}

async function markSeen() {
    if (seen) {
        return;
    }

    seen = true;

    try {
        await viewPost(props.post.id);
    } catch (err) {
        console.error('Could not mark video as seen:', err);
    }
}

async function react(value) {
    const previous = currentReaction.value;
    const next = previous === value ? 0 : value;

    try {
        const result = await postReaction(new Reaction(value, props.post.id).getData());

        if (!result.status) {
            addNotification(result.message);
            return;
        }
    } catch (err) {
        addNotification(err.message || 'could not react');
        return;
    }

    if (previous === 1) {
        likeCount.value--;
    }

    if (previous === -1) {
        dislikeCount.value--;
    }

    if (next === 1) {
        likeCount.value++;
    }

    if (next === -1) {
        dislikeCount.value++;
    }

    currentReaction.value = next;
}

function formatCount(value) {
    if (value >= 1000000) {
        return `${(value / 1000000).toFixed(1).replace(/\.0$/, '')}M`;
    }

    if (value >= 1000) {
        return `${(value / 1000).toFixed(1).replace(/\.0$/, '')}K`;
    }

    return String(value);
}

watch(() => props.active, isActive => {
    const video = videoElement.value;

    if (isActive) {
        if (video) {
            video.currentTime = 0;
        }

        play();
        markSeen();
    } else {
        pause();
        expanded.value = false;
    }
});

watch(() => props.muted, value => {
    if (videoElement.value) {
        videoElement.value.muted = value;
    }
});

watch(showComments, open => {
    if (open) {
        pause();
    } else if (props.active) {
        play();
    }
});

onMounted(() => {
    if (!itemElement.value) {
        return;
    }

    observer = new IntersectionObserver(
        entries => {
            const entry = entries[0];

            if (entry.isIntersecting && entry.intersectionRatio >= 0.6) {
                emit('visible', props.post.id);
            }
        },
        {
            threshold: [0.6]
        }
    );

    observer.observe(itemElement.value);

    if (props.active) {
        play();
        markSeen();
    }
});

onBeforeUnmount(() => {
    observer?.disconnect();
    clearTimeout(flashTimer);
    pause();
});
</script>

<template>
    <article ref="itemElement" class="reel">
        <div class="reel-frame">
            <video ref="videoElement" class="reel-video" :src="videoUrl" :muted="muted" loop playsinline
                preload="metadata" @play="playing = true" @pause="playing = false" @timeupdate="handleTimeUpdate"
                @error="failed = true" @click="togglePlay"></video>

            <div v-if="failed" class="reel-failed">
                Could not load this video.
            </div>

            <div v-if="flash" class="reel-flash" :class="flash">
                <svg v-if="flash === 'pause'" viewBox="0 0 24 24" fill="currentColor">
                    <rect x="6" y="5" width="4" height="14" rx="1" />
                    <rect x="14" y="5" width="4" height="14" rx="1" />
                </svg>

                <svg v-else viewBox="0 0 24 24" fill="currentColor">
                    <path d="M8 5.5V18.5L19 12L8 5.5Z" />
                </svg>
            </div>

            <button class="reel-mute" type="button" :aria-label="muted ? 'Unmute' : 'Mute'"
                @click.stop="emit('toggle-mute')">
                <svg v-if="muted" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round" stroke-linejoin="round">
                    <path d="M11 5L6 9H3V15H6L11 19V5Z" />
                    <path d="M22 9L16 15" />
                    <path d="M16 9L22 15" />
                </svg>

                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round" stroke-linejoin="round">
                    <path d="M11 5L6 9H3V15H6L11 19V5Z" />
                    <path d="M15.5 8.5C17.2 10.2 17.2 13.8 15.5 15.5" />
                    <path d="M18.5 5.5C22 9 22 15 18.5 18.5" />
                </svg>
            </button>

            <div class="reel-gradient"></div>

            <div class="reel-info">
                <a :href="`/user?id=${post.userId}`" class="reel-author">
                    <span class="reel-avatar">
                        <img v-if="post.avatarPath" :src="`/uploads/${post.avatarPath}`" :alt="authorName">
                        <span v-else>{{ (post.firstName || '?').charAt(0) }}</span>
                    </span>

                    <span class="reel-name">{{ authorName }}</span>
                </a>

                <p v-if="post.content" class="reel-caption" :class="{ expanded }"
                    @click="hasLongContent && (expanded = !expanded)">
                    {{ post.content }}
                </p>
            </div>

            <div class="reel-actions">
                <button class="reel-action" :class="{ on: currentReaction === 1 }" type="button" aria-label="Like"
                    @click.stop="react(1)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round">
                        <path d="M12 19V5" />
                        <path d="M5 12L12 5L19 12" />
                    </svg>
                    <span>{{ formatCount(likeCount) }}</span>
                </button>

                <button class="reel-action" :class="{ on: currentReaction === -1 }" type="button"
                    aria-label="Dislike" @click.stop="react(-1)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round">
                        <path d="M12 5V19" />
                        <path d="M5 12L12 19L19 12" />
                    </svg>
                    <span>{{ formatCount(dislikeCount) }}</span>
                </button>

                <button v-if="post.allowComments" class="reel-action" type="button" aria-label="Comments"
                    @click.stop="showComments = true">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21 12C21 16.4 16.97 20 12 20C10.8 20 9.65 19.8 8.6 19.4L3 21L4.6 16.6C3.6 15.3 3 13.7 3 12C3 7.6 7.03 4 12 4C16.97 4 21 7.6 21 12Z" />
                    </svg>
                    <span>{{ formatCount(commentCount) }}</span>
                </button>
            </div>

            <div class="reel-progress" @click.stop="seek">
                <div class="reel-progress-bar" :style="{ width: `${progress}%` }"></div>
            </div>
        </div>

        <HomePostComments :show="showComments" :post-id="post.id" :current-user-id="currentUserId"
            :first-name="post.firstName || ''" :last-name="post.lastName || ''" :avatar-path="post.avatarPath || ''"
            :created-at="post.createdAt || ''" :content="post.content || ''" image-path=""
            @close="showComments = false" />
    </article>
</template>

<style scoped>
.reel {
    width: 100%;
    height: 100%;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 10px 0;
    box-sizing: border-box;
    scroll-snap-align: start;
    scroll-snap-stop: always;
}

.reel-frame {
    position: relative;
    width: 100%;
    max-width: 440px;
    height: 100%;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: #000;
    box-shadow: 6px 6px var(--main-color);
}

.reel-video {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: contain;
    background: #000;
    cursor: pointer;
}

.reel-failed {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    text-align: center;
    pointer-events: none;
}

.reel-flash {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 76px;
    height: 76px;
    margin: -38px 0 0 -38px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.55);
    color: #fff;
    pointer-events: none;
    animation: reel-flash 0.6s ease forwards;
}

.reel-flash svg {
    width: 36px;
    height: 36px;
}

@keyframes reel-flash {
    0% {
        opacity: 0;
        transform: scale(0.8);
    }

    25% {
        opacity: 1;
        transform: scale(1);
    }

    100% {
        opacity: 0;
        transform: scale(1.15);
    }
}

.reel-mute {
    position: absolute;
    top: 12px;
    right: 12px;
    z-index: 3;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.55);
    color: #fff;
    cursor: pointer;
}

.reel-mute svg {
    width: 18px;
    height: 18px;
}

.reel-gradient {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 45%;
    background: linear-gradient(to top, rgba(0, 0, 0, 0.78), rgba(0, 0, 0, 0));
    pointer-events: none;
}

.reel-info {
    position: absolute;
    left: 0;
    right: 76px;
    bottom: 14px;
    z-index: 2;
    padding: 0 16px;
    color: #fff;
    font-family: "JetBrains Mono", monospace;
}

.reel-author {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    max-width: 100%;
    color: #fff;
    text-decoration: none;
}

.reel-avatar {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 2px solid #fff;
    border-radius: 50%;
    background: #555;
    font-size: 14px;
    font-weight: 700;
    text-transform: uppercase;
}

.reel-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.reel-name {
    overflow: hidden;
    font-size: 12px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.reel-caption {
    margin: 10px 0 0;
    font-size: 12px;
    line-height: 1.5;
    word-break: break-word;
    white-space: pre-wrap;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    cursor: pointer;
}

.reel-caption.expanded {
    display: block;
    max-height: 38vh;
    overflow-y: auto;
}

.reel-actions {
    position: absolute;
    right: 10px;
    bottom: 18px;
    z-index: 3;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 18px;
}

.reel-action {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 0;
    border: none;
    background: transparent;
    color: #fff;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
}

.reel-action svg {
    width: 38px;
    height: 38px;
    padding: 8px;
    box-sizing: border-box;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.5);
    transition: transform 0.15s ease, background 0.15s ease;
}

.reel-action:hover svg {
    transform: scale(1.08);
}

.reel-action.on svg {
    background: var(--main-color);
    color: var(--bg-color);
}

.reel-progress {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 4;
    height: 10px;
    display: flex;
    align-items: flex-end;
    cursor: pointer;
}

.reel-progress-bar {
    height: 3px;
    background: #fff;
    transition: width 0.15s linear;
}

@media (max-width: 650px) {
    .reel {
        padding: 0;
    }

    .reel-frame {
        max-width: none;
        border-left: none;
        border-right: none;
        border-radius: 0;
        box-shadow: none;
    }
}
</style>
