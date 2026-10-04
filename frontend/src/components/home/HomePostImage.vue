<script setup>
defineProps({
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

function openTaggedPeople() {
    emit('open-tags');
}
</script>

<template>
    <div v-if="imagePath" class="post-image-container">
        <img :src="`/uploads/${imagePath}`" alt="Post" class="post-image">

        <button v-if="taggedPeople && taggedPeople.length != 0" class="image-tags" type="button" @click="openTaggedPeople">
            <svg viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="9" cy="8" r="3" />
                <path d="M3 19c0-3.3 2.7-6 6-6s6 2.7 6 6" />
                <circle cx="17" cy="9" r="2.2" />
                <path d="M15 14.5c.6-.3 1.3-.5 2-.5 2.8 0 5 2.2 5 5" />
            </svg>

            <span>
                {{ taggedPeople.length }}
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

.post-image {
    display: block;

    width: calc(100% - 40px);
    max-height: 450px;

    object-fit: contain;

    margin: 12px auto;
}

.image-tags {
    position: absolute;
    left: 15px;
    bottom: 15px;

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

    transition:
        transform 0.1s,
        box-shadow 0.1s;
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
}

@media (max-width: 650px) {
    .image-tags {
        left: 10px;
        bottom: 10px;
    }
}
</style>