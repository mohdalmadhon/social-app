
<script setup>
import { computed } from 'vue';

const props = defineProps({
    imagePath: {
        type: String,
        default: ''
    },

    taggedPeople: {
        type: Array,
        default: () => []
    }
});

const emit = defineEmits(['open-tags']);

const validTaggedPeople = computed(() => {
    if (!Array.isArray(props.taggedPeople)) {
        return [];
    }

    return props.taggedPeople.filter(person => {
        return person && (person.id != null || person.ID != null);
    });
});

const hasImage = computed(() => {
    return !!props.imagePath;
});

const isVideo = computed(() => {
    if (!props.imagePath) {
        return false;
    }

    return props.imagePath
        .split('?')[0]
        .toLowerCase()
        .endsWith('.mp4');
});

const mediaUrl = computed(() => {
    if (!props.imagePath) {
        return '';
    }

    return `/uploads/${props.imagePath}`;
});

function openTaggedPeople() {
    emit('open-tags');
}
</script>

<template>
    <div
        v-if="hasImage || validTaggedPeople.length"
        class="post-image-container"
        :class="{ 'no-image': !hasImage }"
    >
        <video
            v-if="isVideo"
            :src="mediaUrl"
            class="post-video"
            controls
            playsinline
            preload="metadata"
        ></video>

        <img
            v-else-if="hasImage"
            :src="mediaUrl"
            alt="Post"
            class="post-image"
        >

        <button
            v-if="validTaggedPeople.length"
            class="image-tags"
            :class="{ 'no-image-tags': !hasImage }"
            type="button"
            aria-label="Show tagged people"
            title="Tagged people"
            @click.stop="openTaggedPeople"
        >
            <svg
                viewBox="0 0 24 24"
                aria-hidden="true"
            >
                <path
                    d="M20.6 13.4 13.4 20.6a2 2 0 0 1-2.8 0L3 13V3h10l7.6 7.6a2 2 0 0 1 0 2.8Z"
                />

                <circle
                    cx="7.5"
                    cy="7.5"
                    r="1.5"
                />
            </svg>

            <span>
                {{ validTaggedPeople.length }}
            </span>
        </button>
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

.post-image-container.no-image {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 12px 20px;
    background: transparent;
}

.post-image {
    display: block;
    width: calc(100% - 40px);
    max-height: 450px;
    object-fit: contain;
    margin: 12px auto;
}

.post-video {
    display: block;
    width: calc(100% - 40px);
    max-height: 450px;
    object-fit: contain;
    margin: 12px auto;
    border-radius: 4px;
    background: #000;
}

.image-tags {
    position: absolute;
    left: 15px;
    bottom: 15px;
    z-index: 2;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 10px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
    color: var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
    transition:
        transform 0.1s,
        box-shadow 0.1s;
}

.image-tags.no-image-tags {
    position: static;
}

.image-tags:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.image-tags:active {
    transform: translate(2px, 2px);
    box-shadow: 1px 1px var(--main-color);
}

.image-tags svg {
    width: 17px;
    height: 17px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.8;
    stroke-linecap: round;
    stroke-linejoin: round;
}

.image-tags svg circle {
    fill: currentColor;
    stroke: none;
}

@media (max-width: 650px) {
    .post-image-container.no-image {
        padding: 10px 14px;
    }

    .image-tags {
        left: 10px;
        bottom: 10px;
    }

    .post-image,
    .post-video {
        width: calc(100% - 20px);
    }
}
</style>
