<script setup>
import { computed } from 'vue';

const props = defineProps({
    imagePath: {
        type: String,
        default: ''
    }
});

const isVideo = computed(() =>
    (props.imagePath || '').split('?')[0].toLowerCase().endsWith('.mp4')
);
</script>

<template>
    <div v-if="imagePath" class="post-image-container">
        <video v-if="isVideo" :src="`/uploads/${imagePath}`" class="post-image" controls playsinline preload="metadata"></video>
        <img v-else :src="`/uploads/${imagePath}`" alt="Post" class="post-image">
    </div>
</template>

<style scoped>
.post-image-container {
    position: relative;

    width: 100%;
    padding: 0 20px;

    border-top: 2px solid var(--main-color);
    border-bottom: 2px solid var(--main-color);

    background: #dedede;
}

.post-image {
    display: block;

    width: calc(100% - 40px);
    max-height: 450px;

    object-fit: contain;

    margin: 12px auto;
}

@media (max-width: 650px) {
    .post-image-container {
        padding: 0 14px;
    }

    .post-image {
        width: calc(100% - 28px);
    }
}
</style>
