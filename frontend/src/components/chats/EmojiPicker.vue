<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { EMOJI_CATEGORIES } from '@/helpers/emojis';

const props = defineProps({
    disabled: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['select']);

const open = ref(false);
const activeIndex = ref(0);
const root = ref(null);

function toggle() {
    if (props.disabled) {
        return;
    }

    open.value = !open.value;
}

function close() {
    open.value = false;
}

function pick(emoji) {
    emit('select', emoji);
}

function handleOutsideClick(event) {
    if (open.value && root.value && !root.value.contains(event.target)) {
        close();
    }
}

function handleKeydown(event) {
    if (open.value && event.key === 'Escape') {
        close();
    }
}

watch(
    () => props.disabled,
    disabled => {
        if (disabled) {
            close();
        }
    }
);

onMounted(() => {
    document.addEventListener('mousedown', handleOutsideClick);
    document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
    document.removeEventListener('mousedown', handleOutsideClick);
    document.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
    <div ref="root" class="emoji-picker">
        <button type="button" class="emoji-trigger" title="Add emoji" :disabled="disabled" @click="toggle">
            😊 Emoji
        </button>

        <div v-if="open" class="emoji-panel">
            <div class="emoji-tabs">
                <button v-for="(category, index) in EMOJI_CATEGORIES" :key="category.name" type="button"
                    class="emoji-tab" :class="{ active: index === activeIndex }" :title="category.name"
                    @mousedown.prevent @click="activeIndex = index">
                    {{ category.icon }}
                </button>
            </div>

            <div class="emoji-grid">
                <button v-for="emoji in EMOJI_CATEGORIES[activeIndex].emojis" :key="emoji" type="button"
                    class="emoji-item" @mousedown.prevent @click="pick(emoji)">
                    {{ emoji }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.emoji-picker {
    flex-shrink: 0;
    display: flex;
}

.emoji-trigger {
    flex: 1;
    flex-shrink: 0;
    padding: 0 14px;
    height: 42px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--bg-color);
    color: var(--font-color);
    box-shadow: 4px 4px var(--main-color);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    white-space: nowrap;
    cursor: pointer;
}

.emoji-trigger:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.emoji-trigger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.emoji-panel {
    position: absolute;
    left: 20px;
    bottom: 100%;
    z-index: 30;
    width: min(340px, calc(100% - 40px));
    margin-bottom: 8px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--bg-color);
    box-shadow: 4px 4px var(--main-color);
    box-sizing: border-box;
    overflow: hidden;
}

.emoji-tabs {
    display: flex;
    gap: 2px;
    padding: 4px;
    border-bottom: 2px solid var(--page-background);
    overflow-x: auto;
}

.emoji-tab {
    flex: 1 0 auto;
    min-width: 34px;
    height: 32px;
    padding: 0;
    border: 2px solid transparent;
    border-radius: 5px;
    background: transparent;
    font-family: "Apple Color Emoji", "Segoe UI Emoji", "Noto Color Emoji", sans-serif;
    font-size: 16px;
    line-height: 1;
    cursor: pointer;
}

.emoji-tab:hover {
    background: var(--page-background);
}

.emoji-tab.active {
    border-color: var(--main-color);
    background: var(--page-background);
}

.emoji-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(34px, 1fr));
    gap: 2px;
    max-height: 200px;
    padding: 6px;
    overflow-y: auto;
}

.emoji-item {
    height: 34px;
    padding: 0;
    border: none;
    border-radius: 5px;
    background: transparent;
    font-family: "Apple Color Emoji", "Segoe UI Emoji", "Noto Color Emoji", sans-serif;
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
}

.emoji-item:hover {
    background: var(--page-background);
}

@media (max-width: 800px) {
    .emoji-panel {
        left: 14px;
        width: calc(100% - 28px);
    }
}
</style>
