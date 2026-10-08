<script setup>
import { LIMITS, charCount } from '@/helpers/limits';
import { reactive, ref, watch } from 'vue';

import PostContentForm from '@/components/addPost/PostContentForm.vue';
import ImageContainer from '@/components/addPost/ImageContainer.vue';
import TagPeople from '@/components/addPost/TagPeople.vue';
import LocationPicker from '@/components/addPost/LocationPicker.vue';

import { addNotification } from '@/data/notifications';
import { addGroupPost } from '@/api/posts/groups';
import { sendWS } from '@/api/socket/socket';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    groupId: {
        type: Number,
        required: true
    },
    groupTitle: {
        type: String,
        default: ''
    }
});

const emit = defineEmits(['close', 'created']);

const totalSlides = 3;

const post = reactive({
    content: '',
    location: '',
    taggedPeople: [],
    image: null
});

const currentSlide = ref(1);
const posting = ref(false);

const validation = ref({
    field: null,
    message: null
});

function resetPost() {
    post.content = '';
    post.location = '';
    post.taggedPeople = [];
    post.image = null;

    currentSlide.value = 1;

    validation.value = {
        field: null,
        message: null
    };
}

watch(
    post,
    () => {
        validation.value = validatePost();
    },
    { deep: true }
);

function validatePost() {
    if (!post.content.trim()) {
        return {
            field: 'content',
            message: 'Post content is required'
        };
    }

    if (charCount(post.content) > LIMITS.postContent) {
        return {
            field: 'content',
            message: `Post content cannot exceed ${LIMITS.postContent} characters`
        };
    }

    if (post.taggedPeople.length > LIMITS.postTags) {
        return {
            field: 'tags',
            message: `You can tag at most ${LIMITS.postTags} people`
        };
    }

    if (post.image && post.image.size > LIMITS.postMediaSize) {
        return {
            field: 'image',
            message: 'File must be smaller than 50MB'
        };
    }

    if (post.taggedPeople.length > 0) {
        const invalidTag = post.taggedPeople.some(
            person => !person.id
        );

        if (invalidTag) {
            return {
                field: 'tags',
                message: 'Invalid tagged person'
            };
        }
    }

    return {
        field: null,
        message: null
    };
}

function nextSlide() {
    const result = validatePost();

    if (result.field) {
        validation.value = result;
        addNotification(result.message, 'error');
        return;
    }

    if (currentSlide.value < totalSlides) {
        currentSlide.value++;

        validation.value = {
            field: null,
            message: null
        };
    }
}

function previousSlide() {
    if (currentSlide.value > 1) {
        currentSlide.value--;

        validation.value = {
            field: null,
            message: null
        };
    }
}

function closeDialog() {
    if (posting.value) {
        return;
    }

    emit('close');
    resetPost();
}

async function handleSubmit() {
    const result = validatePost();

    if (result.field) {
        validation.value = result;
        addNotification(result.message, 'error');
        return;
    }

    if (!props.groupId) {
        addNotification('Invalid group', 'error');
        return;
    }

    posting.value = true;

    try {
        const data = {
            content: post.content,
            allowComments: 1,
            groupID: props.groupId,
            location: post.location || '',
            taggedPeople: post.taggedPeople.map(
                person => person.id
            ),
            image: post.image
        };

        const result = await addGroupPost(data);

        if (!result.status) {
            addNotification(
                result.message || 'Could not send post',
                'error'
            );
            return;
        }

        addNotification('Post created!', 'success');

        sendWS({
            type: 'postGroup',
            data: result.data.data
        });

        emit('created');
        closeDialog();
    } catch (err) {
        addNotification(
            err.message || 'Could not send post',
            'error'
        );
    } finally {
        posting.value = false;
    }
}
</script>

<template>
    <div v-if="show" class="post-modal-overlay" @click.self="closeDialog">
        <form class="post-modal" @submit.prevent="handleSubmit">
            <div class="modal-scroll">
                <div class="modal-header">
                    <div>
                        <p class="eyebrow">
                            GROUP POST
                        </p>

                        <h1>
                            Create post
                        </h1>

                        <p v-if="groupTitle" class="posting-to">
                            Posting to <strong>{{ groupTitle }}</strong>
                        </p>
                    </div>

                    <button type="button" class="close-button" :disabled="posting" @click="closeDialog">
                        ×
                    </button>
                </div>

                <div class="progress">
                    <div v-for="slide in totalSlides" :key="slide" class="progress-step" :class="{
                        active: slide === currentSlide,
                        completed: slide < currentSlide
                    }">
                        {{ slide }}
                    </div>
                </div>

                <div class="slide-title">
                    <p class="slide-number">
                        STEP {{ currentSlide }} / {{ totalSlides }}
                    </p>

                    <h2 v-if="currentSlide === 1">
                        Content & Image
                    </h2>

                    <h2 v-else-if="currentSlide === 2">
                        Tag People
                    </h2>

                    <h2 v-else>
                        Location
                    </h2>
                </div>

                <div v-if="currentSlide === 1" class="slide">
                    <PostContentForm v-model:description="post.content" />

                    <p v-if="validation.field === 'content'" class="validation-error">
                        {{ validation.message }}
                    </p>

                    <ImageContainer v-model="post.image" />
                </div>

                <div v-else-if="currentSlide === 2" class="slide">
                    <TagPeople v-model="post.taggedPeople" />

                    <p v-if="validation.field === 'tags'" class="validation-error">
                        {{ validation.message }}
                    </p>
                </div>

                <div v-else class="slide">
                    <LocationPicker v-model="post.location" />
                </div>
            </div>

            <div class="navigation-buttons">
                <button v-if="currentSlide > 1" class="previous-button" type="button" @click="previousSlide">
                    Previous
                </button>

                <button v-if="currentSlide < totalSlides" class="next-button" type="button" @click="nextSlide">
                    Next
                </button>

                <button v-else class="submit-button" type="submit" :disabled="posting">
                    {{ posting ? 'Posting...' : 'Post' }}
                </button>
            </div>
        </form>
    </div>
</template>

<style scoped>
.post-modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 30px;
    background: rgba(0, 0, 0, 0.55);
    overflow-y: auto;
}

.post-modal {
    width: 100%;
    max-width: 700px;
    max-height: calc(100vh - 60px);
    max-height: calc(100dvh - 60px);
    display: flex;
    flex-direction: column;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    box-sizing: border-box;
    overflow: hidden;
}

.modal-scroll {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 25px;
    padding: 35px 40px;
}

.modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 20px;
}

.modal-header h1 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 30px;
}

.posting-to {
    margin: 6px 0 0;
    color: var(--font-color-sub);
    font-size: 12px;
}

.posting-to strong {
    color: var(--font-color);
}

.close-button {
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--bg-color);
    color: var(--main-color);
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
}

.close-button:hover:not(:disabled) {
    background: var(--input-focus);
    color: white;
}

.close-button:disabled {
    opacity: 0.5;
    cursor: default;
}

.progress {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
}

.progress-step {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--main-color);
    border-radius: 50%;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    background: var(--bg-color);
    color: var(--main-color);
}

.progress-step.active {
    background: var(--input-focus);
    color: white;
}

.progress-step.completed {
    background: var(--main-color);
    color: white;
}

.slide-title {
    border-bottom: 2px solid var(--main-color);
    padding-bottom: 18px;
}

.slide-number {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.slide-title h2 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 27px;
}

.slide {
    min-height: 250px;
    display: flex;
    flex-direction: column;
    gap: 22px;
}

.validation-error {
    margin: -12px 0 0;
    color: #d93025;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
}

.navigation-buttons {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
    padding: 18px 40px;
    border-top: 2px solid var(--main-color);
    background: var(--bg-color);
}

.previous-button,
.next-button,
.submit-button {
    padding: 13px 30px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    cursor: pointer;
    transition:
        transform 0.1s ease,
        box-shadow 0.1s ease;
}

.previous-button {
    background: var(--bg-color);
    color: var(--main-color);
    box-shadow: 4px 4px var(--main-color);
}

.next-button,
.submit-button {
    margin-left: auto;
    background: var(--input-focus);
    color: white;
    box-shadow: 4px 4px var(--main-color);
}

.previous-button:hover,
.next-button:hover,
.submit-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.previous-button:active,
.next-button:active,
.submit-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.submit-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

@media (max-width: 800px) {
    .post-modal-overlay {
        padding: 15px;
    }

    .post-modal {
        max-height: calc(100vh - 30px);
        max-height: calc(100dvh - 30px);
    }

    .modal-scroll {
        padding: 28px 24px;
    }

    .navigation-buttons {
        padding: 16px 24px;
    }
}

@media (max-width: 550px) {
    .post-modal {
        border-radius: 12px;
    }

    .modal-scroll {
        padding: 25px 20px;
    }

    .navigation-buttons {
        padding: 14px 20px;
    }

    .modal-header h1 {
        font-size: 26px;
    }

    .slide-title h2 {
        font-size: 23px;
    }

    .navigation-buttons {
        flex-direction: column-reverse;
    }

    .previous-button,
    .next-button,
    .submit-button {
        width: 100%;
        margin-left: 0;
        text-align: center;
    }
}
</style>