<script setup>
defineProps({
    activeTab: {
        type: String,
        default: 'group'
    }
});

defineEmits(['change']);

const tabs = [
    { id: 'group', label: 'posts' },
    { id: 'events', label: 'Events' },
    { id: 'chat', label: 'Chat' }
];
</script>

<template>
    <nav class="group-tabs">
        <button
            v-for="tab in tabs"
            :key="tab.id"
            class="tab-button"
            type="button"
            :class="{ active: activeTab === tab.id }"
            @click="$emit('change', tab.id)"
        >
            {{ tab.label }}
        </button>
    </nav>
</template>

<style scoped>
.group-tabs {
    display: flex;

    width: 100%;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);

    overflow: hidden;
}

.tab-button {
    flex: 1;

    padding: 13px 16px;

    border: 0;
    border-right: 2px solid var(--main-color);

    background: var(--bg-color);
    color: var(--font-color);

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;

    transition:
        background 0.15s,
        color 0.15s;
}

.tab-button:last-child {
    border-right: 0;
}

.tab-button:hover {
    background: var(--page-background);
}

.tab-button.active {
    background: var(--main-color);
    color: var(--bg-color);
}

@media (max-width: 650px) {
    .tab-button {
        padding: 11px 10px;
        font-size: 11px;
    }
}
</style>
