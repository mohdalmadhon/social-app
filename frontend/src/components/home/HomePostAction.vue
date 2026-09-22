<script setup>
import { onMounted, ref, watch } from 'vue';

import { postReaction } from '@/api/posts/actions';

import { addNotification } from '@/data/notifications';

import { Reaction } from '@/models/posts';

const props = defineProps({
    allowComments: {
        type: Boolean
    },

    reaction: {

        type: Number,

        default: 0

    },

    postId: {

        type: Number,

        required: true

    },

    likes: {

        type: Number,

        default: 0

    },

    dislikes: {

        type: Number,

        default: 0

    }

});

const emit = defineEmits(['like', 'dislike', 'toggle-comments']);

const likeCount = ref(props.likes);

const dislikeCount = ref(props.dislikes);

const currentReaction = ref(props.reaction);

const reactionAnimation = ref('');

watch(() => props.likes, (value) => {
    likeCount.value = value;
});

watch(() => props.dislikes, (value) => {
    dislikeCount.value = value;
});

watch(() => props.reaction, (value) => {
    currentReaction.value = value;
});

async function handleReaction(value) {

    const previousReaction = currentReaction.value;

    const newReaction = previousReaction === value ? 0 : value;

    const rect = new Reaction(value, props.postId);

    const result = await postReaction(rect.getData());

    if (!result.status) {

        addNotification(result.message);

        return;

    }

    currentReaction.value = newReaction;

    reactionAnimation.value = value === 1 ? 'like' : 'dislike';

    setTimeout(() => {

        reactionAnimation.value = '';

    }, 350);

    if (previousReaction === 1) {

        likeCount.value--;

    }

    if (previousReaction === -1) {

        dislikeCount.value--;

    }

    if (newReaction === 1) {

        likeCount.value++;

    }

    if (newReaction === -1) {

        dislikeCount.value++;

    }

    if (newReaction === 1) {

        emit('like');

    } else {

        emit('dislike');

    }

}

function toggleComments() {

    emit('toggle-comments');

}

onMounted(() => {
    console.log(props.likes)
})
</script>

<template>

    <div class="post-actions">

        <button class="action-button" :class="{
            active: currentReaction === 1,
            'reaction-jump': reactionAnimation === 'like'
        }" type="button" @click="handleReaction(1)">

            <svg viewBox="0 0 24 24" aria-hidden="true">

                <path
                    d="M7 10v10H3V10h4Zm3 10h7.2a2 2 0 0 0 1.9-1.4l2.3-7A2 2 0 0 0 19.5 9H15l.7-3.4A2.2 2.2 0 0 0 13.5 3L9 9v11h1Z" />

            </svg>

            Like {{ likeCount }}

        </button>

        <button class="action-button dislike" :class="{
            active: currentReaction === -1,
            'reaction-jump': reactionAnimation === 'dislike'
        }" type="button" @click="handleReaction(-1)">

            <svg viewBox="0 0 24 24" aria-hidden="true">

                <path d="M7 14V4H3v10h4Zm3-10h7.2a2 2 0 0 1 1.9 1.4l.7 3.4a2.2 2.2 0 0 1-2.2 2.6L9 15V4h1Z" />

            </svg>

            Dislike {{ dislikeCount }}

        </button>

        <button v-if="props.allowComments" class="action-button" type="button" @click="toggleComments">

            <svg viewBox="0 0 24 24" aria-hidden="true">

                <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 9.6 9.6 0 0 1-4-.9L3 21l1.5-4.2A8.5 8.5 0 0 1 21 11.5Z" />

            </svg>

            Comment

        </button>

    </div>

</template>

<style scoped>
.post-actions {

    display: grid;

    grid-template-columns: repeat(3, 1fr);

    padding: 7px;

}

.action-button {

    min-height: 42px;

    display: flex;

    align-items: center;

    justify-content: center;

    gap: 7px;

    border: 2px solid transparent;

    border-radius: 5px;

    background: transparent;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;

    font-size: 9px;

    font-weight: 600;

    transition:
        background 0.15s,
        color 0.15s,
        border 0.15s;

}

.action-button:hover {

    background: var(--page-background);

    color: var(--main-color);

}

.action-button svg {

    width: 18px;

    height: 18px;

    fill: none;

    stroke: currentColor;

    stroke-width: 1.8;

    stroke-linecap: round;

    stroke-linejoin: round;

}

.action-button.active {

    border-color: var(--main-color);

    background: var(--input-focus);

    color: white;

}

.action-button.dislike.active {

    background: var(--main-color);

    color: white;

}

.reaction-jump svg {

    animation: jump 0.35s ease;

}

@keyframes jump {

    0% {

        transform: translateY(0) scale(1);

    }

    40% {

        transform: translateY(-8px) scale(1.2);

    }

    70% {

        transform: translateY(2px) scale(0.95);

    }

    100% {

        transform: translateY(0) scale(1);

    }

}

@media (max-width: 650px) {

    .post-actions {

        padding: 5px;

    }

    .action-button {

        gap: 4px;

        font-size: 8px;

    }

    .action-button svg {

        width: 16px;

        height: 16px;

    }

}
</style>