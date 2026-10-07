<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';

defineProps({
    userId: {
        type: [Number, String],
        default: null
    },

    firstName: {
        type: String,
        default: ''
    },

    lastName: {
        type: String,
        default: ''
    },

    avatarPath: {
        type: String,
        default: ''
    },

    formattedDate: {
        type: String,
        default: ''
    },

    canDelete: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['delete']);

const menuOpen = ref(false);
const menuRoot = ref(null);

function requestDelete() {
    menuOpen.value = false;
    emit('delete');
}

function handleOutsideClick(event) {
    if (menuOpen.value && menuRoot.value && !menuRoot.value.contains(event.target)) {
        menuOpen.value = false;
    }
}

onMounted(() => {
    document.addEventListener('click', handleOutsideClick);
});

onBeforeUnmount(() => {
    document.removeEventListener('click', handleOutsideClick);
});

</script>

<template>
    <header class="post-header">
        <div class="author">
            <a :href="`/user?id=${userId}`" class="avatar">
                <img v-if="avatarPath" :src="`/uploads/${avatarPath}`" :alt="`${firstName} ${lastName}`">

                <span v-else>
                    {{ firstName?.charAt(0) }}
                </span>
            </a>

            <div class="author-information">
                <div class="author-line">
                    <span class="author-name">
                        {{ firstName }} {{ lastName }}
                    </span>
                </div>

                <div class="post-meta">
                    <span class="post-date">{{ formattedDate }}</span>
                </div>
            </div>
        </div>

        <div v-if="canDelete" ref="menuRoot" class="post-options">
            <button class="more-button" type="button" aria-label="Post options" aria-haspopup="true"
                :aria-expanded="menuOpen" @click="menuOpen = !menuOpen">
                •••
            </button>

            <div v-if="menuOpen" class="post-menu">
                <button type="button" class="post-menu-item" @click="requestDelete">
                    Delete post
                </button>
            </div>
        </div>
    </header>
</template>

<style scoped>
.post-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 15px;

    padding: 18px 20px;
}

.author {
    min-width: 0;

    display: flex;
    align-items: center;
    gap: 12px;
}

.avatar {
    flex-shrink: 0;

    width: 48px;
    height: 48px;

    display: flex;
    align-items: center;
    justify-content: center;

    overflow: hidden;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--input-focus);
    box-shadow: 3px 3px var(--main-color);

    color: white;

    font-family: "Liter", serif;
    font-size: 20px;
    font-weight: 600;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.author-information {
    min-width: 0;
}

.author-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 5px;

    font-size: 14px;
}

.author-name {
    color: var(--main-color);

    font-family: "Liter", serif;
    font-weight: 600;
}

.post-meta {
    margin-top: 5px;

    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
}

.post-date {
    padding: 2px 6px;

    border-radius: 4px;

    background: var(--input-focus);

    color: white;

    font-weight: 600;
}

.more-button {
    flex-shrink: 0;

    padding: 5px 8px;

    border: none;
    background: transparent;

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 15px;
    font-weight: 600;
}

.more-button:hover {
    color: var(--main-color);
}

@media (max-width: 650px) {
    .post-header {
        padding: 14px;
    }

    .avatar {
        width: 42px;
        height: 42px;
    }
}

.post-options {
    position: relative;
    flex-shrink: 0;
}

.more-button {
    cursor: pointer;
}

.post-menu {
    position: absolute;
    top: 100%;
    right: 0;
    z-index: 20;
    min-width: 140px;
    margin-top: 4px;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 3px 3px var(--main-color);
}

.post-menu-item {
    display: block;
    width: 100%;
    padding: 10px 14px;
    border: 0;
    background: transparent;
    color: #d9534f;
    font-family: "JetBrains Mono", monospace;
    font-size: 11px;
    font-weight: 700;
    text-align: left;
    cursor: pointer;
}

.post-menu-item:hover {
    background: var(--page-background);
}
</style>
