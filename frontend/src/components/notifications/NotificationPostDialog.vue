<script setup>
import { ref, watch } from 'vue';
import HomePosts from '@/components/home/HomePosts.vue';
import { getSinglePost } from '@/api/posts/single';
import { postReaction } from '@/api/posts/actions';
import { Reaction } from '@/models/posts';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    postId: {
        type: [Number, String],
        default: null
    }
});

const emit = defineEmits(['close']);

const post = ref(null);
const currentUserId = ref(null);
const loadingPost = ref(false);
const loadError = ref('');
const reacting = ref(false);

let loadRequestID = 0;

function reactionLabel(value) {
    if (value === 1) return 'like';
    if (value === -1) return 'dislike';
    return '';
}

function normalizePost(data) {
    if (!data) return null;

    const reactionValue = data.ReactionValue ?? data.reactionValue ?? 0;

    return {
        postId: Number(data.id),
        userId: Number(data.userId),
        firstName: data.firstName ?? '',
        lastName: data.lastName ?? '',
        username: data.username ?? '',
        avatarPath: data.avatarPath ?? '',
        content: data.content ?? '',
        imagePath: data.imagePath ?? '',
        location: data.location ?? '',
        groupId: data.groupId ?? null,
        createdAt: data.createdAt ?? '',
        relationship: data.relationship || 'none',
        visibility: data.visibility || 'public',
        visibilityUser: data.visibilityUser ?? '',
        taggedPeople: data.taggedPeople ?? [],
        likes: data.likeCount ?? 0,
        dislikes: data.disLikeCount ?? 0,
        comments: Array.from({ length: data.commentCount ?? 0 }),
        reaction: reactionValue,
        userReaction: reactionLabel(reactionValue)
    };
}

async function loadPost() {
    if (!props.postId) return;

    const requestID = ++loadRequestID;

    loadingPost.value = true;
    loadError.value = '';
    post.value = null;

    try {
        const result = await getSinglePost(props.postId);

        if (requestID !== loadRequestID) return;

        post.value = normalizePost(result.post);
        currentUserId.value = result.userId;
    } catch (err) {
        if (requestID !== loadRequestID) return;

        loadError.value = err.message || 'Could not load post';
    } finally {
        if (requestID === loadRequestID) {
            loadingPost.value = false;
        }
    }
}

async function applyReaction(value) {
    if (!post.value || reacting.value) return;

    const current = post.value;
    const previousReaction = current.reaction;
    const previousLikes = current.likes;
    const previousDislikes = current.dislikes;

    let nextReaction = value;
    let likes = current.likes;
    let dislikes = current.dislikes;

    if (previousReaction === 1) {
        likes -= 1;
    } else if (previousReaction === -1) {
        dislikes -= 1;
    }

    if (previousReaction === value) {
        nextReaction = 0;
    } else if (value === 1) {
        likes += 1;
    } else if (value === -1) {
        dislikes += 1;
    }

    post.value = {
        ...current,
        reaction: nextReaction,
        userReaction: reactionLabel(nextReaction),
        likes,
        dislikes
    };

    reacting.value = true;

    try {
        const reaction = new Reaction(value, current.postId);
        await postReaction(reaction.getData());
    } catch (err) {
        post.value = {
            ...post.value,
            reaction: previousReaction,
            userReaction: reactionLabel(previousReaction),
            likes: previousLikes,
            dislikes: previousDislikes
        };

        addNotification(err.message || 'Could not update reaction', 'error');
    } finally {
        reacting.value = false;
    }
}

function likePost() {
    applyReaction(1);
}

function dislikePost() {
    applyReaction(-1);
}

function closeDialog() {
    loadRequestID++;
    emit('close');
}

watch(
    () => [props.show, props.postId],
    ([visible]) => {
        if (visible) {
            loadPost();
        } else {
            post.value = null;
            loadError.value = '';
            loadingPost.value = false;
        }
    },
    { immediate: true }
);
</script>

<template>
    <Teleport to="body">
        <div v-if="show" class="notification-post-overlay" @click.self="closeDialog">
            <div class="notification-post-dialog">
                <button type="button" class="notification-post-close" @click="closeDialog">×</button>

                <div v-if="loadingPost" class="notification-post-state">Loading post...</div>

                <div v-else-if="loadError" class="notification-post-state error">{{ loadError }}</div>

                <HomePosts
                    v-else-if="post"
                    :key="post.postId"
                    :current-user-id="currentUserId"
                    :post-id="post.postId"
                    :user-id="post.userId"
                    :first-name="post.firstName"
                    :last-name="post.lastName"
                    :username="post.username"
                    :avatar-path="post.avatarPath"
                    :group-id="post.groupId"
                    :content="post.content"
                    :image-path="post.imagePath"
                    :location="post.location"
                    :created-at="post.createdAt"
                    :reaction="post.reaction"
                    :likes="post.likes"
                    :dislikes="post.dislikes"
                    :comments="post.comments"
                    :tagged-people="post.taggedPeople"
                    :user-reaction="post.userReaction"
                    :relationship="post.relationship"
                    :visibility="post.visibility"
                    :visibility-user="post.visibilityUser"
                    @like="likePost"
                    @dislike="dislikePost"
                />
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.notification-post-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 30px;
    overflow-y: auto;
    background: rgba(0, 0, 0, 0.7);
}

.notification-post-dialog {
    position: relative;
    width: 100%;
    max-width: 720px;
    max-height: calc(100vh - 60px);
    overflow-y: auto;
}

.notification-post-close {
    position: absolute;
    top: -14px;
    right: -14px;
    z-index: 10;
    width: 34px;
    height: 34px;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--font-color);
    font-size: 22px;
    line-height: 26px;
    cursor: pointer;
    box-shadow: 3px 3px var(--main-color);
}

.notification-post-state {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 200px;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
}

.notification-post-state.error {
    color: #e5484d;
}

@media (max-width: 800px) {
    .notification-post-overlay {
        padding: 16px;
    }

    .notification-post-dialog {
        max-height: calc(100vh - 32px);
    }

    .notification-post-close {
        top: -6px;
        right: -6px;
    }
}

@media (max-width: 520px) {
    .notification-post-overlay {
        padding: 10px;
    }

    .notification-post-close {
        top: 4px;
        right: 4px;
        width: 30px;
        height: 30px;
        font-size: 18px;
        line-height: 22px;
    }
}
</style>
