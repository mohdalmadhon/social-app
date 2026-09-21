<script setup>
import { computed, ref } from 'vue';
import GroupPostHeader from './post/GroupPostHeader.vue';
import GroupPostImage from './post/GroupPostImage.vue';
import GroupPostReaction from './post/GroupPostReaction.vue';
import GroupPostActions from './post/GroupPostActions.vue';
import GroupPostComments from './post/GroupPostComments.vue';

const props = defineProps({
    post: {
        type: Object,
        required: true
    },
    currentUserId: {
        type: [Number, String],
        default: null
    }
});

const showComments = ref(false);
const commentCount = ref(Number(props.post.commentCount ?? 0));
const likeCount = ref(Number(props.post.likeCount ?? 0));
const dislikeCount = ref(Number(props.post.disLikeCount ?? 0));
const reaction = ref(Number(props.post.ReactionValue ?? props.post.reactionValue ?? 0));

const formattedDate = computed(() => {
    const created = new Date(props.post.createdAt);
    if (Number.isNaN(created.getTime())) return props.post.createdAt ?? '';
    return created.toLocaleString();
});

function toggleComments() {
    showComments.value = !showComments.value;
}

function closeComments() {
    showComments.value = false;
}

function changeCommentCount(delta) {
    commentCount.value = Math.max(0, commentCount.value + delta);
}

function applyReactionCount(value) {
    const previousReaction = reaction.value;
    const newReaction = previousReaction === value ? 0 : value;

    if (previousReaction === 1) likeCount.value = Math.max(0, likeCount.value - 1);
    if (previousReaction === -1) dislikeCount.value = Math.max(0, dislikeCount.value - 1);
    if (newReaction === 1) likeCount.value += 1;
    if (newReaction === -1) dislikeCount.value += 1;

    reaction.value = newReaction;
}

function handleLike() {
    applyReactionCount(1);
}

function handleDislike() {
    applyReactionCount(-1);
}
</script>

<template>
    <article class="post-card">
        <GroupPostHeader
            :user-id="post.userId ?? post.userID"
            :first-name="post.firstName"
            :last-name="post.lastName"
            :avatar-path="post.avatarPath"
            :formatted-date="formattedDate"
        />

        <div v-if="post.content" class="post-content">
            {{ post.content }}
        </div>

        <GroupPostImage :image-path="post.imagePath" />

        <GroupPostReaction
            :likes="likeCount"
            :dislikes="dislikeCount"
            :comments-count="commentCount"
            @toggle-comments="toggleComments"
        />

        <GroupPostActions
            :reaction="reaction"
            :post-id="Number(post.id)"
            :likes="likeCount"
            :dislikes="dislikeCount"
            @like="handleLike"
            @dislike="handleDislike"
            @toggle-comments="toggleComments"
        />

        <GroupPostComments
            :show="showComments"
            :post-id="Number(post.id)"
            :current-user-id="currentUserId"
            :first-name="post.firstName || ''"
            :last-name="post.lastName || ''"
            :avatar-path="post.avatarPath || ''"
            :created-at="post.createdAt || ''"
            :content="post.content || ''"
            :image-path="post.imagePath || ''"
            @close="closeComments"
            @count-changed="changeCommentCount"
        />
    </article>
</template>

<style scoped>
.post-card {
    width: 100%;
    max-width: 720px;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.post-content {
    padding: 0 20px 18px;
    color: var(--font-color);
    font-size: 14px;
    line-height: 1.55;
    white-space: pre-wrap;
    word-break: break-word;
}

@media (max-width: 650px) {
    .post-card {
        box-shadow: 4px 4px var(--main-color);
    }

    .post-content {
        padding: 0 14px 15px;
        font-size: 13px;
    }
}
</style>
