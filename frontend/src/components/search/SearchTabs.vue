<script setup>
defineProps({
    tabs: {
        type: Array,
        required: true
    },
    activeTab: {
        type: String,
        required: true
    }
});

defineEmits(['change']);
</script>

<template>
    <nav class="search-tabs" role="tablist" aria-label="Search results">
        <button
            v-for="tab in tabs"
            :id="`search-tab-${tab.id}`"
            :key="tab.id"
            class="tab-button"
            type="button"
            role="tab"
            :aria-selected="activeTab === tab.id"
            :aria-controls="`search-panel-${tab.id}`"
            :class="{ active: activeTab === tab.id }"
            @click="$emit('change', tab.id)"
        >
            <span>{{ tab.label }}</span>

            <span v-if="tab.count" class="tab-count">
                {{ tab.count }}
            </span>
        </button>
    </nav>
</template>

<style scoped>
.search-tabs {
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

    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;

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

.tab-count {
    min-width: 20px;

    padding: 2px 6px;

    border: 1px solid currentColor;
    border-radius: 10px;

    font-size: 10px;
    line-height: 1.3;
    letter-spacing: 0;
}

@media (max-width: 650px) {
    .tab-button {
        gap: 5px;
        padding: 11px 6px;
        font-size: 11px;
    }

    .tab-count {
        padding: 1px 5px;
        font-size: 9px;
    }
}
</style>
