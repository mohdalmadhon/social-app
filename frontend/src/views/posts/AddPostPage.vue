<script setup>
import { reactive, ref, watch } from 'vue';

import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import HideCommentsToggle from '@/components/addPost/HideCommentsToggle.vue';
import LocationPicker from '@/components/addPost/LocationPicker.vue';
import PostContentForm from '@/components/addPost/PostContentForm.vue';
import TagPeople from '@/components/addPost/TagPeople.vue';
import PostPrivacy from '@/components/addPost/PostPrivacy.vue';
import ImageContainer from '@/components/addPost/ImageContainer.vue';

import { addPost } from '@/api/posts/posts';
import { addNotification } from '@/data/notifications';
import { validatePost } from '@/helpers/validators/posts';
import { router } from '@/router/router';

const post = reactive({
    title: '',
    content: '',
    hideComments: false,
    privacy: 'public',
    groupID: 0,
    location: '',
    taggedPeople: [],
    image: null
});

const currentSlide = ref(1);

const validation = ref({
    field: null,
    message: null
});

const totalSlides = 4;

watch(
    post,
    () => {
        const data = {
            content: post.content,
            allowComments: post.hideComments ? 0 : 1,
            privatePost: 0,
            groupID: post.groupID,
            location: post.location || '',
            taggedPeople: post.taggedPeople.map(user => user.id),
            image: post.image
        };

        validation.value = validatePost(data);
    },
    { deep: true }
);

function nextSlide() {
    if (validation.value.field) {
        addNotification(validation.value.message, 'error');
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

async function handleSubmit() {
    const data = {
        content: post.content,
        allowComments: post.hideComments ? 0 : 1,
        privatePost: 0,
        groupID: post.groupID,
        location: post.location || null,
        taggedPeople: post.taggedPeople.map(user => user.id),
        image: post.image
    };

    const validationResult = validatePost(data);

    if (validationResult.field) {
        validation.value = validationResult;
        addNotification(validationResult.message, 'error');
        return;
    }

    try {
        const result = await addPost(data, post.image);

        if (!result.status) {
            addNotification('could not send post', 'error');
            return;
        }

        router.push("/me?tab=posts")
        console.log(result);
    } catch (err) {
        addNotification(
            err.message || 'could not send post',
            'error'
        );
    }
}
</script>

<template>
    <div class="facebook-layout">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="add-post-page">
                <BackToHome />

                <div class="page-heading">
                    <p class="eyebrow">SHARE</p>
                    <h1>Create post</h1>
                </div>

                <form class="add-post-card" @submit.prevent="handleSubmit">
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

                        <h2 v-else-if="currentSlide === 3">
                            Location
                        </h2>

                        <h2 v-else>
                            Privacy & Comments
                        </h2>
                    </div>

                    <div v-if="currentSlide === 1" class="slide">
                        <PostContentForm v-model:description="post.content" />

                        <p v-if="validation.field === 'content'" class="validation-error">
                            {{ validation.message }}
                        </p>

                        <ImageContainer v-model="post.image" />

                        <p v-if="validation.field === 'image'" class="validation-error">
                            {{ validation.message }}
                        </p>
                    </div>

                    <div v-else-if="currentSlide === 2" class="slide">
                        <TagPeople v-model="post.taggedPeople" />

                        <p v-if="validation.field === 'tags'" class="validation-error">
                            {{ validation.message }}
                        </p>
                    </div>

                    <div v-else-if="currentSlide === 3" class="slide">
                        <LocationPicker v-model="post.location" />

                        <p v-if="validation.field === 'location'" class="validation-error">
                            {{ validation.message }}
                        </p>
                    </div>

                    <div v-else class="slide">
                        <HideCommentsToggle v-model="post.hideComments" />

                        <p v-if="validation.field === 'allowComments'" class="validation-error">
                            {{ validation.message }}
                        </p>

                        <PostPrivacy v-model="post.privacy" v-model:group-i-d-value="post.groupID" />

                        <p v-if="validation.field === 'groupID'" class="validation-error">
                            {{ validation.message }}
                        </p>
                    </div>

                    <div class="navigation-buttons">
                        <button v-if="currentSlide > 1" class="previous-button" type="button" @click="previousSlide">
                            Previous
                        </button>

                        <button v-if="currentSlide < totalSlides" class="next-button" type="button" @click="nextSlide">
                            Next
                        </button>

                        <button v-else class="submit-button" type="submit">
                            Post
                        </button>
                    </div>
                </form>
            </main>
        </div>
    </div>
</template>

<style scoped>
.facebook-layout {
    min-height: 100vh;
    min-height: 100dvh;
}

.page-layout {
    display: flex;
    padding-top: 64px;
}

.add-post-page {
    width: 100%;
    max-width: 1100px;
    margin: 0 auto;
    padding: 25px 30px 60px;
    box-sizing: border-box;
}

.page-heading .eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.page-heading h1 {
    margin: 0 0 25px;
    font-family: "Liter", serif;
    font-size: 36px;
}

.add-post-card {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 25px;
    padding: 35px 40px;
    border: 2px solid var(--main-color);
    border-radius: 7px;
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
    box-sizing: border-box;
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
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
    margin-top: 10px;
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
    transition: transform 0.1s ease, box-shadow 0.1s ease;
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

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .add-post-page {
        padding: 20px 15px 50px;
    }
}

@media (max-width: 550px) {
    .add-post-card {
        padding: 28px 24px;
        border-radius: 12px;
    }

    .page-heading h1 {
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