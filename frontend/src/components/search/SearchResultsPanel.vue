<script setup>
defineProps({
    state: {
        type: Object,
        required: true
    },
    noun: {
        type: String,
        required: true
    },
    query: {
        type: String,
        default: ''
    }
});

defineEmits(['retry', 'load-more']);
</script>

<template>
    <div class="results-panel">
        <div
            v-if="state.loading && !state.items.length"
            class="message"
        >
            Searching...
        </div>

        <div
            v-else-if="state.error && !state.items.length"
            class="message error"
        >
            <span>{{ state.error }}</span>

            <button
                class="action-button"
                type="button"
                @click="$emit('retry')"
            >
                Try again
            </button>
        </div>

        <div
            v-else-if="state.loaded && !state.items.length"
            class="message"
        >
            No {{ noun }} found for “{{ query }}”.
        </div>

        <template v-else>
            <slot />

            <div v-if="state.error" class="message error">
                <span>{{ state.error }}</span>

                <button
                    class="action-button"
                    type="button"
                    @click="$emit('retry')"
                >
                    Try again
                </button>
            </div>

            <div v-else-if="state.loading" class="message">
                Loading more...
            </div>

            <button
                v-else-if="state.hasMore"
                class="action-button load-more"
                type="button"
                @click="$emit('load-more')"
            >
                Load more
            </button>
        </template>
    </div>
</template>

<style scoped>
.message {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 12px;

    padding: 24px 20px;

    color: var(--font-color);
    text-align: center;

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    line-height: 1.5;

    overflow-wrap: anywhere;
}

.message.error {
    color: #c62828;
}

.action-button {
    padding: 10px 20px;

    border: 2px solid var(--main-color);
    border-radius: 5px;

    background: var(--page-background);
    color: var(--font-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 600;

    transition:
        transform 0.1s,
        background 0.15s,
        color 0.15s;
}

.action-button:hover {
    background: var(--main-color);
    color: var(--bg-color);
}

.action-button:active {
    transform: translate(2px, 2px);
}

.load-more {
    display: block;
    margin: 20px auto;
}
</style>
