<script setup>
import { LIMITS } from '@/helpers/limits';
const props = defineProps({
    modelValue: {
        type: String,
        default: ''
    },
    placeholder: {
        type: String,
        default: 'Search groups...'
    }
});

const emit = defineEmits(['update:modelValue']);

function updateValue(event) {
    emit('update:modelValue', event.target.value);
}

function clear() {
    emit('update:modelValue', '');
}
</script>

<template>
    <div class="search-container">
        <div class="search-icon">
            <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
            >
                <circle cx="11" cy="11" r="7"></circle>
                <path d="m20 20-4-4"></path>
            </svg>
        </div>

        <input
            :value="modelValue"
            :maxlength="LIMITS.search"
            type="text"
            :placeholder="placeholder"
            class="search-input"
            @input="updateValue"
        >

        <button
            v-if="modelValue"
            class="clear-button"
            type="button"
            @click="clear"
        >
            ×
        </button>
    </div>
</template>

<style scoped>
.search-container {
    width: 100%;
    max-width: 680px;
    min-height: 62px;

    display: flex;
    align-items: center;
    gap: 12px;

    padding: 10px 16px;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);

    box-shadow: 5px 5px var(--main-color);

    transition:
        transform 0.15s,
        box-shadow 0.15s;
}

.search-container:focus-within {
    transform: translate(-1px, -1px);
    box-shadow: 6px 6px var(--main-color);
}

.search-icon {
    flex-shrink: 0;

    width: 38px;
    height: 38px;

    display: flex;
    align-items: center;
    justify-content: center;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--input-focus);
    color: white;
}

.search-icon svg {
    width: 19px;
    height: 19px;
}

.search-input {
    width: 100%;
    min-width: 0;

    padding: 11px 14px;

    border: 2px solid transparent;
    border-radius: 25px;

    outline: none;

    background: #f0f1f3;
    color: var(--font-color);

    font-family: "Hedvig Letters Sans", sans-serif;
    font-size: 15px;

    transition:
        border-color 0.15s,
        background 0.15s;
}

.search-input::placeholder {
    color: var(--font-color-sub);
}

.search-input:focus {
    border-color: var(--input-focus);
    background: var(--bg-color);
}

.clear-button {
    flex-shrink: 0;

    width: 32px;
    height: 32px;

    display: flex;
    align-items: center;
    justify-content: center;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--bg-color);
    color: var(--main-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 18px;
    line-height: 1;

    transition:
        transform 0.1s,
        background 0.15s;
}

.clear-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.clear-button:active {
    transform: translate(2px, 2px);
}

@media (max-width: 650px) {
    .search-container {
        min-height: 56px;
        padding: 8px 12px;
    }

    .search-icon {
        width: 34px;
        height: 34px;
    }

    .search-input {
        padding: 9px 12px;
        font-size: 14px;
    }
}
</style>