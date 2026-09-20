<script setup>
defineEmits(['add-post']);
</script>

<template>
    <div class="feed-heading">
        <div>
            <p class="eyebrow">
                GROUP FEED
            </p>

            <h2>
                Posts
            </h2>
        </div>

        <button type="button" class="add-post-button" @click="$emit('add-post')">
            + Add Post
        </button>
    </div>
</template>

<style scoped>
.feed-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    margin-top: 5px;
}

.feed-heading h2 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 27px;
}

.eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.add-post-button {
    flex-shrink: 0;
    padding: 12px 20px;
    border: 2px solid var(--main-color);
    border-radius: 6px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 700;
    cursor: pointer;
    transition:
        transform 0.1s ease,
        box-shadow 0.1s ease;
}

.add-post-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.add-post-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

@media (max-width: 550px) {
    .feed-heading {
        align-items: flex-start;
        flex-direction: column;
    }

    .add-post-button {
        width: 100%;
    }
}
</style>