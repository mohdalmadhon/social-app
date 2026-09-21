<script setup>
defineProps({
    likes: {
        type: Number,
        default: 0
    },

    dislikes: {
        type: Number,
        default: 0
    },

    commentsCount: {
        type: Number,
        default: 0
    }
});

const emit = defineEmits(['toggle-comments']);

function toggleComments() {
    emit('toggle-comments');
}
</script>

<template>
    <div class="reaction-summary">
        <div class="reaction-counts">
            <span v-if="likes" class="reaction-summary-item">
                <span class="summary-icon like-summary">
                    ↑
                </span>

                {{ likes }}
            </span>

            <span v-if="dislikes" class="reaction-summary-item">
                <span class="summary-icon dislike-summary">
                    ↓
                </span>

                {{ dislikes }}
            </span>
        </div>

        <button class="comments-count" type="button" @click="toggleComments">
            {{ commentsCount }}
            {{ commentsCount === 1 ? 'comment' : 'comments' }}
        </button>
    </div>
</template>

<style scoped>
.reaction-summary {
    min-height: 44px;

    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;

    padding: 8px 18px;

    border-bottom: 1px solid #dedede;
}

.reaction-counts {
    display: flex;
    align-items: center;
    gap: 10px;
}

.reaction-summary-item {
    display: flex;
    align-items: center;
    gap: 4px;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.summary-icon {
    width: 20px;
    height: 20px;

    display: inline-flex;
    align-items: center;
    justify-content: center;

    border: 1.5px solid var(--main-color);
    border-radius: 50%;

    color: white;

    font-size: 12px;
    font-weight: 600;
}

.like-summary {
    background: var(--input-focus);
}

.dislike-summary {
    background: var(--main-color);
}

.comments-count {
    padding: 0;

    border: none;
    background: transparent;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.comments-count:hover {
    color: var(--main-color);
    text-decoration: underline;
}

@media (max-width: 650px) {
    .reaction-summary {
        padding: 8px 12px;
    }
}
</style>
