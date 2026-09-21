<script setup>
defineProps({
    comment: {
        type: Object,
        required: true
    },
    isReply: {
        type: Boolean,
        default: false
    },
    isOwner: {
        type: Boolean,
        default: false
    },
    isActive: {
        type: Boolean,
        default: false
    },
    menuOpen: {
        type: Boolean,
        default: false
    },
    repliesLoading: {
        type: Boolean,
        default: false
    },
    replyError: {
        type: String,
        default: ''
    },
    formattedDate: {
        type: String,
        default: ''
    }
});

const emit = defineEmits([
    'open-profile',
    'toggle-menu',
    'reply',
    'toggle-replies',
    'like',
    'delete'
]);
</script>

<template>
    <div class="comment" :class="{ reply: isReply, 'comment-active': isActive }">
        <div class="comment-row">
            <img v-if="comment.user?.avatar" :src="`/uploads/${comment.user.avatar}`" class="comment-avatar clickable"
                @click="emit('open-profile', comment.user.ID)">

            <div v-else class="comment-avatar avatar-fallback">
                {{ comment.user?.firstName?.charAt(0) }}
            </div>

            <div class="comment-main">
                <div class="comment-header">
                    <strong>{{ comment.user?.firstName }} {{ comment.user?.lastName }}</strong>

                    <button v-if="isOwner" type="button" class="dots-button" @click="emit('toggle-menu')">
                        ⋯
                    </button>
                </div>

                <div class="comment-bubble">{{ comment.content }}</div>

                <div class="comment-actions">
                    <span>{{ formattedDate }}</span>

                    <button type="button" @click="emit('like')">Like</button>

                    <span class="vote-count">{{ comment.votes }}</span>

                    <button v-if="!isReply" type="button" @click="emit('reply')">Reply</button>

                    <button v-if="!isReply && comment.replies > 0" type="button" @click="emit('toggle-replies')">
                        {{ repliesLoading ? 'Loading...' : comment.showReplies ? 'Hide replies' : `Show ${comment.replies} replies` }}
                    </button>
                </div>

                <div v-if="menuOpen" class="comment-menu">
                    <button type="button" @click="emit('delete')">Delete</button>
                </div>

                <div v-if="replyError" class="reply-error">{{ replyError }}</div>

                <slot name="replies" />
            </div>
        </div>
    </div>
</template>

<style scoped>
.comment {
    position: relative;
    padding: 14px;
    margin-bottom: 12px;
    border: 2px solid var(--cd-border);
    border-radius: var(--cd-radius);
    box-shadow: var(--cd-shadow-sm);
    transition: box-shadow 0.12s, transform 0.12s;
}

.comment.comment-active {
    box-shadow: 3px 3px 0 var(--cd-accent);
    border-color: var(--cd-accent);
}

.comment.reply {
    box-shadow: none;
    border-style: dashed;
}

.comment-row {
    display: flex;
    gap: 10px;
}

.comment-main {
    position: relative;
    min-width: 0;
    flex: 1;
}

.comment-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.comment-header strong {
    color: var(--cd-text);
    font-size: 13px;
    font-weight: 700;
}

.dots-button {
    border: 0;
    padding: 2px 6px;
    border-radius: 6px;
    background: transparent;
    color: var(--cd-text-muted);
    font-size: 16px;
    font-weight: 700;
    line-height: 1;
    cursor: pointer;
}

.dots-button:hover {
    background: var(--cd-surface);
    color: var(--cd-text);
}

.comment-bubble {
    margin-top: 4px;
    color: var(--cd-text);
    font-size: 14px;
    font-weight: 500;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
}

.comment-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 9px;
    color: var(--cd-text-muted);
    font-size: 12px;
}

.comment-actions .vote-count {
    font-weight: 700;
    color: var(--cd-text);
}

.comment-actions button {
    border: 2px solid var(--cd-border);
    padding: 3px 10px;
    border-radius: 999px;
    background: var(--cd-bg);
    color: var(--cd-text);
    font-size: 12px;
    font-weight: 700;
    cursor: pointer;
    transition: background 0.12s, color 0.12s, transform 0.1s;
}

.comment-actions button:hover {
    background: var(--cd-text);
    color: #fff;
}

.comment-actions button:active {
    transform: translate(1px, 1px);
}

.comment-menu {
    position: absolute;
    top: 30px;
    right: 0;
    z-index: 2;
    padding: 4px;
    border: 2px solid var(--cd-border);
    border-radius: 10px;
    background: var(--cd-bg);
    box-shadow: var(--cd-shadow-sm);
}

.comment-menu button {
    border: 0;
    width: 100%;
    padding: 7px 12px;
    border-radius: 6px;
    background: transparent;
    color: var(--cd-danger);
    font-size: 13px;
    font-weight: 700;
    text-align: left;
    cursor: pointer;
}

.comment-menu button:hover {
    background: var(--cd-surface);
}

.reply-error {
    margin-top: 8px;
    padding: 9px 12px;
    border: 2px solid var(--cd-danger);
    border-radius: 10px;
    background: color-mix(in srgb, var(--cd-danger) 8%, white);
    color: var(--cd-danger);
    font-size: 12px;
    font-weight: 700;
}

.comment-avatar {
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border-radius: 50%;
    object-fit: cover;
    background: var(--cd-surface);
    border: 2px solid var(--cd-border);
}

.clickable {
    cursor: pointer;
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
</style>
