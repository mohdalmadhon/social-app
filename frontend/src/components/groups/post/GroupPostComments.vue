<script setup>
import { ref, computed, watch, nextTick } from 'vue';
import {
    getGroupComments,
    addGroupComment,
    deleteGroupComment,
    voteGroupComment
} from '@/api/posts/groupComments';
import { GroupComment, createGroupComment } from '@/models/groupPosts';
import { router } from '@/router/router';
import { useCommentMedia, commentImageSrc, COMMENT_MEDIA_ACCEPT } from '@/helpers/commentMedia';
import GroupCommentItem from './GroupCommentItem.vue';

const props = defineProps({
    show: { type: Boolean, default: false },
    postId: { type: Number, required: true },
    currentUserId: { type: [Number, String], default: null },
    firstName: { type: String, default: '' },
    lastName: { type: String, default: '' },
    avatarPath: { type: String, default: '' },
    createdAt: { type: String, default: '' },
    content: { type: String, default: '' },
    imagePath: { type: String, default: '' }
});

const emit = defineEmits(['close', 'count-changed']);

const comments = ref([]);
const loading = ref(false);
const submitting = ref(false);
const error = ref('');
const newComment = ref('');
const replyingTo = ref(null);
const loadingReplies = ref({});
const replyErrors = ref({});
const menuComment = ref(null);
const commentInput = ref(null);
const {
    file: mediaFile,
    previewUrl: mediaPreview,
    error: mediaError,
    fileInput,
    clearMedia,
    openPicker,
    onSelect,
    takeMedia,
    restoreMedia,
    revokeMedia
} = useCommentMedia();
const resolvedUserId = ref(null);

const displayName = computed(() => `${props.firstName} ${props.lastName}`.trim());
const replyingToName = computed(() => {
    const user = replyingTo.value?.user;
    return user ? `${user.firstName} ${user.lastName}`.trim() : '';
});
const inputPlaceholder = computed(() =>
    replyingTo.value ? `Reply to ${replyingToName.value}...` : 'Write a comment...'
);
const activeUserId = computed(() => props.currentUserId ?? resolvedUserId.value);

function convertComments(data) {
    return data.map(comment => createGroupComment(comment));
}

function formatDate(date) {
    const created = new Date(date);
    if (Number.isNaN(created.getTime())) return date;

    const diff = Date.now() - created.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (minutes < 1) return 'now';
    if (minutes < 60) return `${minutes}m`;
    if (hours < 24) return `${hours}h`;
    if (days < 7) return `${days}d`;
    return created.toLocaleDateString();
}

async function loadComments() {
    loading.value = true;
    error.value = '';

    try {
        const result = await getGroupComments(props.postId);
        comments.value = convertComments(result.comments);
        resolvedUserId.value = result.userId;
    } catch (err) {
        error.value = err.message || 'Failed to load comments';
    } finally {
        loading.value = false;
    }
}

function createTemporaryComment(content, replyTo = null, imagePath = '') {
    const comment = new GroupComment(
        `temporary-${Date.now()}`, content, props.postId, replyTo, 0,
        new Date().toISOString(),
        { ID: Number(activeUserId.value), firstName: 'You', lastName: '', avatar: '' },
        0
    );
    comment.pending = true;
    comment.imagePath = imagePath;
    return comment;
}

async function submitTopLevel(content, media = null) {
    submitting.value = true;
    error.value = '';
    const temporaryComment = createTemporaryComment(content, null, media?.previewUrl ?? '');
    comments.value.unshift(temporaryComment);
    emit('count-changed', 1);

    try {
        const data = await addGroupComment(props.postId, content, null, media?.file ?? null);
        const realComment = createGroupComment(data);
        const index = comments.value.findIndex(comment => comment.ID === temporaryComment.ID);
        if (index !== -1) comments.value[index] = realComment;
        revokeMedia(media);
    } catch (err) {
        comments.value = comments.value.filter(comment => comment.ID !== temporaryComment.ID);
        emit('count-changed', -1);
        newComment.value = content;
        restoreMedia(media);
        error.value = err.message || 'Failed to add comment';
    } finally {
        submitting.value = false;
    }
}

async function submitReplyTo(parentComment, content, media = null) {
    submitting.value = true;
    replyErrors.value[parentComment.ID] = '';
    const temporaryReply = createTemporaryComment(content, parentComment.ID, media?.previewUrl ?? '');

    if (!parentComment.loadedReplies) parentComment.loadedReplies = [];
    parentComment.loadedReplies.unshift(temporaryReply);
    parentComment.replies++;
    parentComment.showReplies = true;
    emit('count-changed', 1);

    try {
        const data = await addGroupComment(props.postId, content, parentComment.ID, media?.file ?? null);
        const realReply = createGroupComment(data);
        const index = parentComment.loadedReplies.findIndex(reply => reply.ID === temporaryReply.ID);
        if (index !== -1) parentComment.loadedReplies[index] = realReply;
        revokeMedia(media);
        replyingTo.value = null;
    } catch (err) {
        parentComment.loadedReplies = parentComment.loadedReplies.filter(reply => reply.ID !== temporaryReply.ID);
        parentComment.replies--;
        emit('count-changed', -1);
        newComment.value = content;
        restoreMedia(media);
        replyErrors.value[parentComment.ID] = err.message || 'Failed to add reply';
    } finally {
        submitting.value = false;
    }
}

async function submitComment() {
    const content = newComment.value.trim();
    if ((!content && !mediaFile.value) || submitting.value) return;

    const media = takeMedia();
    newComment.value = '';

    if (replyingTo.value) {
        await submitReplyTo(replyingTo.value, content, media);
    } else {
        await submitTopLevel(content, media);
    }
}

async function showReplies(comment) {
    if (comment.loadedReplies !== null) {
        comment.showReplies = !comment.showReplies;
        return;
    }

    loadingReplies.value[comment.ID] = true;
    replyErrors.value[comment.ID] = '';

    try {
        const result = await getGroupComments(props.postId, comment.ID);
        comment.loadedReplies = convertComments(result.comments);
        comment.showReplies = true;
    } catch (err) {
        replyErrors.value[comment.ID] = err.message || 'Failed to load replies';
    } finally {
        loadingReplies.value[comment.ID] = false;
    }
}

function startReply(comment) {
    replyingTo.value = comment;
    newComment.value = '';
    clearMedia();
    nextTick(() => commentInput.value?.focus());
}

function cancelReply() {
    replyingTo.value = null;
    newComment.value = '';
    clearMedia();
}

function isOwner(comment) {
    const userId = activeUserId.value;
    if (userId === null || userId === undefined) return false;
    return Number(comment.user?.ID) === Number(userId);
}

async function removeComment(comment) {
    menuComment.value = null;
    const index = comments.value.findIndex(item => item.ID === comment.ID);
    if (index === -1) return;

    const removedComment = comments.value[index];
    comments.value.splice(index, 1);
    emit('count-changed', -1);

    try {
        await deleteGroupComment(comment.ID);
    } catch (err) {
        comments.value.splice(index, 0, removedComment);
        emit('count-changed', 1);
        error.value = err.message || 'Failed to delete comment';
    }
}

async function removeReply(parentComment, reply) {
    menuComment.value = null;
    const index = parentComment.loadedReplies.findIndex(item => item.ID === reply.ID);
    if (index === -1) return;

    const removedReply = parentComment.loadedReplies[index];
    parentComment.loadedReplies.splice(index, 1);
    parentComment.replies--;
    emit('count-changed', -1);

    try {
        await deleteGroupComment(reply.ID);
    } catch (err) {
        parentComment.loadedReplies.splice(index, 0, removedReply);
        parentComment.replies++;
        emit('count-changed', 1);
        replyErrors.value[parentComment.ID] = err.message || 'Failed to delete reply';
    }
}

async function likeComment(comment) {
    const oldVotes = comment.votes;
    comment.votes++;

    try {
        await voteGroupComment(comment.ID, 1);
    } catch (err) {
        comment.votes = oldVotes;
        error.value = err.message || 'Failed to like comment';
    }
}

function close() {
    clearMedia();
    emit('close');
}

function takeToProfile(id) {
    if (!id) return;
    router.push(`/user?id=${id}`);
}

watch(
    () => props.show,
    value => {
        if (value) {
            replyingTo.value = null;
            newComment.value = '';
            loadComments();
        }
    }
);
</script>

<template>
    <Teleport to="body">
        <div v-if="show" class="comments-overlay" @click.self="close">
            <section class="comments-dialog">

                <header class="comments-header">
                    <h2>Comments</h2>
                    <button type="button" class="close-button" @click="close">×</button>
                </header>

                <div class="comments-content">
                    <article class="dialog-post">
                        <div class="post-user">
                            <img v-if="avatarPath" :src="`/uploads/${avatarPath}`" class="post-avatar">
                            <div v-else class="post-avatar avatar-fallback">{{ firstName.charAt(0) }}</div>
                            <div class="post-user-info">
                                <strong>{{ displayName }}</strong>
                                <span>{{ formatDate(createdAt) }}</span>
                            </div>
                        </div>
                        <div v-if="content" class="post-text">{{ content }}</div>
                        <img v-if="imagePath" :src="`/uploads/${imagePath}`" class="post-image">
                    </article>

                    <div v-if="error" class="comments-error">{{ error }}</div>
                    <div v-if="loading" class="comments-loading">Loading comments...</div>
                    <div v-else-if="!comments.length" class="no-comments">No comments yet.</div>

                    <GroupCommentItem
                        v-for="comment in comments"
                        :key="comment.ID"
                        :comment="comment"
                        :is-owner="isOwner(comment)"
                        :is-active="replyingTo && replyingTo.ID === comment.ID"
                        :menu-open="menuComment === comment.ID"
                        :replies-loading="!!loadingReplies[comment.ID]"
                        :reply-error="replyErrors[comment.ID] || ''"
                        :formatted-date="formatDate(comment.createdAt)"
                        @open-profile="takeToProfile"
                        @toggle-menu="menuComment = menuComment === comment.ID ? null : comment.ID"
                        @reply="startReply(comment)"
                        @toggle-replies="showReplies(comment)"
                        @like="likeComment(comment)"
                        @delete="removeComment(comment)"
                    >
                        <template #replies>
                            <div v-if="comment.showReplies" class="replies">
                                <GroupCommentItem
                                    v-for="reply in comment.loadedReplies"
                                    :key="reply.ID"
                                    :comment="reply"
                                    is-reply
                                    :is-owner="isOwner(reply)"
                                    :menu-open="menuComment === reply.ID"
                                    :formatted-date="formatDate(reply.createdAt)"
                                    @open-profile="takeToProfile"
                                    @toggle-menu="menuComment = menuComment === reply.ID ? null : reply.ID"
                                    @like="likeComment(reply)"
                                    @delete="removeReply(comment, reply)"
                                />
                            </div>
                        </template>
                    </GroupCommentItem>
                </div>

                <div v-if="replyingTo" class="reply-banner">
                    <span>Replying to <strong>{{ replyingToName }}</strong></span>
                    <button type="button" @click="cancelReply">✕</button>
                </div>

                <div v-if="mediaError" class="comments-error media-error">{{ mediaError }}</div>

                <div v-if="mediaPreview" class="comment-media-preview">
                    <img :src="mediaPreview" alt="Selected image">
                    <button type="button" aria-label="Remove image" @click="clearMedia">✕</button>
                </div>

                <form class="add-comment" @submit.prevent="submitComment">
                    <input ref="fileInput" type="file" class="media-file-input" :accept="COMMENT_MEDIA_ACCEPT"
                        @change="onSelect">
                    <button type="button" class="media-button" title="Add an image or GIF" aria-label="Add an image or GIF"
                        @click="openPicker">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <rect x="3" y="3" width="18" height="18" rx="3" />
                            <circle cx="9" cy="9" r="1.8" />
                            <path d="m21 15-5-5L5 21" />
                        </svg>
                    </button>
                    <input ref="commentInput" v-model="newComment" maxlength="200" :placeholder="inputPlaceholder">
                    <button type="submit" :disabled="submitting || (!newComment.trim() && !mediaFile)">{{ replyingTo ? 'Reply' : 'Post'
                        }}</button>
                </form>

            </section>
        </div>
    </Teleport>
</template>

<style scoped>
.comments-dialog {
    --cd-bg: #ffffff;
    --cd-surface: #f4f4f4;
    --cd-border: #000000;
    --cd-text: #111114;
    --cd-text-muted: #6b6b70;
    --cd-accent: #2f6fed;
    --cd-accent-ink: #ffffff;
    --cd-danger: #e5484d;
    --cd-radius: 14px;
    --cd-shadow: 4px 4px 0 var(--cd-border);
    --cd-shadow-sm: 2px 2px 0 var(--cd-border);
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.comments-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgb(0 0 0 / 55%);
}

.comments-dialog {
    width: 100%;
    max-width: 620px;
    height: min(760px, 90vh);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 2px solid var(--cd-border);
    border-radius: var(--cd-radius);
    background: var(--cd-bg);
    box-shadow: 6px 6px 0 var(--cd-border);
}

.comments-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 22px;
    border-bottom: 2px solid var(--cd-border);
}

.comments-header h2 {
    margin: 0;
    color: var(--cd-text);
    font-size: 16px;
    font-weight: 800;
    letter-spacing: -0.01em;
}

.close-button {
    border: 2px solid var(--cd-border);
    width: 30px;
    height: 30px;
    border-radius: 8px;
    background: var(--cd-bg);
    color: var(--cd-text);
    font-size: 18px;
    line-height: 1;
    font-weight: 700;
    cursor: pointer;
    box-shadow: var(--cd-shadow-sm);
    transition: transform 0.1s, box-shadow 0.1s, background 0.1s, color 0.1s;
}

.close-button:hover {
    background: var(--cd-text);
    color: #fff;
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0 var(--cd-border);
}

.comments-content {
    flex: 1;
    overflow-y: auto;
    padding: 20px 22px;
}

.dialog-post {
    margin-bottom: 18px;
    padding: 14px;
    border: 2px solid var(--cd-border);
    border-radius: var(--cd-radius);
    box-shadow: var(--cd-shadow-sm);
}

.post-user {
    display: flex;
    align-items: center;
    gap: 10px;
}

.post-user-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.post-user-info strong {
    color: var(--cd-text);
    font-size: 13px;
    font-weight: 700;
}

.post-user-info span {
    color: var(--cd-text-muted);
    font-size: 12px;
    font-weight: 600;
}

.post-avatar {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border-radius: 50%;
    object-fit: cover;
    background: var(--cd-surface);
    border: 2px solid var(--cd-border);
}

.avatar-fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--cd-accent);
    color: #fff;
    font-weight: 800;
    font-size: 14px;
    text-transform: uppercase;
}

.post-text {
    padding-top: 10px;
    color: var(--cd-text);
    font-size: 14px;
    font-weight: 600;
    line-height: 1.55;
    white-space: pre-wrap;
    word-break: break-word;
}

.post-image {
    width: 100%;
    max-height: 340px;
    margin-top: 12px;
    border: 2px solid var(--cd-border);
    border-radius: 10px;
    object-fit: contain;
    background: var(--cd-surface);
}

.replies {
    margin-top: 12px;
    margin-left: 14px;
    padding-left: 14px;
    border-left: 2px solid var(--cd-border);
}

.reply-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 0 22px;
    padding: 8px 14px;
    border: 2px solid var(--cd-accent);
    border-bottom: 0;
    border-radius: 12px 12px 0 0;
    background: color-mix(in srgb, var(--cd-accent) 10%, white);
    color: var(--cd-text);
    font-size: 12px;
    font-weight: 600;
}

.reply-banner strong {
    font-weight: 800;
}

.reply-banner button {
    border: 0;
    background: transparent;
    color: var(--cd-text-muted);
    font-size: 14px;
    font-weight: 800;
    cursor: pointer;
}

.reply-banner button:hover {
    color: var(--cd-danger);
}

.add-comment {
    display: flex;
    gap: 10px;
    padding: 16px 22px;
    border-top: 2px solid var(--cd-border);
}

.add-comment input {
    min-width: 0;
    flex: 1;
    border: 2px solid var(--cd-border);
    border-radius: 999px;
    padding: 10px 16px;
    outline: none;
    background: var(--cd-surface);
    color: var(--cd-text);
    font-size: 13px;
    font-weight: 600;
    box-shadow: var(--cd-shadow-sm);
    transition: background 0.12s, box-shadow 0.12s;
}

.add-comment input:focus {
    background: var(--cd-bg);
    box-shadow: 3px 3px 0 var(--cd-accent);
}

.add-comment button {
    border: 2px solid var(--cd-border);
    border-radius: 999px;
    padding: 10px 18px;
    background: var(--cd-text);
    color: #fff;
    font-size: 13px;
    font-weight: 800;
    cursor: pointer;
    box-shadow: var(--cd-shadow-sm);
    transition: transform 0.1s, box-shadow 0.1s, background 0.1s;
}

.add-comment button:hover:not(:disabled) {
    background: var(--cd-accent);
}

.add-comment button:active:not(:disabled) {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0 var(--cd-border);
}

.add-comment button:disabled {
    cursor: not-allowed;
    opacity: 0.4;
}

.comments-error {
    margin-bottom: 14px;
    padding: 9px 12px;
    border: 2px solid var(--cd-danger);
    border-radius: 10px;
    background: color-mix(in srgb, var(--cd-danger) 8%, white);
    color: var(--cd-danger);
    font-size: 13px;
    font-weight: 700;
}

.comments-loading,
.no-comments {
    padding: 40px 0;
    text-align: center;
    color: var(--cd-text-muted);
    font-size: 13px;
    font-weight: 600;
}

.add-comment .media-file-input {
    display: none;
}

.add-comment .media-button {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 42px;
    padding: 0;
    background: var(--cd-bg);
    color: var(--cd-text);
}

.add-comment .media-button:hover:not(:disabled) {
    background: var(--cd-text);
    color: #fff;
}

.comment-media-preview {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 22px 0;
    border-top: 2px solid var(--cd-border);
}

.comment-media-preview + .add-comment {
    border-top: 0;
}

.comment-media-preview img {
    max-width: 160px;
    max-height: 120px;
    border: 2px solid var(--cd-border);
    border-radius: 10px;
    object-fit: contain;
    background: var(--cd-surface);
}

.comment-media-preview button {
    width: 24px;
    height: 24px;
    border: 2px solid var(--cd-border);
    border-radius: 50%;
    background: var(--cd-bg);
    color: var(--cd-text);
    font-size: 11px;
    font-weight: 800;
    line-height: 1;
    cursor: pointer;
}

.comment-media-preview button:hover {
    background: var(--cd-danger);
    color: #fff;
}

.comments-error.media-error {
    margin: 0 22px 10px;
}

.comment-image {
    display: block;
    max-width: 100%;
    max-height: 260px;
    margin-top: 8px;
    border: 2px solid var(--cd-border);
    border-radius: 10px;
    object-fit: contain;
    background: var(--cd-surface);
}

@media (max-width: 650px) {
    .comments-overlay {
        padding: 0;
    }

    .comments-dialog {
        width: 100%;
        height: 100%;
        max-width: none;
        border-radius: 0;
        border-width: 0 0 2px 0;
        box-shadow: none;
    }

    .comments-content {
        padding: 16px 16px;
    }

    .reply-banner {
        margin: 0 16px;
    }

    .comment-media-preview {
        padding: 12px 16px 0;
    }

    .comments-error.media-error {
        margin: 0 16px 10px;
    }

    .add-comment {
        padding: 12px 16px;
    }
}
</style>
