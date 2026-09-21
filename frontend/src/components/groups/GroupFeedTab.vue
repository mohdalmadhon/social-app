<script setup>
import GroupPostCard from './GroupPostCard.vue';

defineProps({
    posts: {
        type: Array,
        default: () => []
    },
    currentUserId: {
        type: [Number, String],
        default: null
    }
});
</script>

<template>
    <section class="tab-panel">
        <div v-if="!posts.length" class="empty-state">
            No posts in this group yet.
        </div>

        <div v-else class="posts-list">
            <GroupPostCard
                v-for="post in posts"
                :key="post.id"
                :post="post"
                :current-user-id="currentUserId"
            />
        </div>
    </section>
</template>

<style scoped>
.tab-panel {
    width: 100%;
}

.empty-state {
    padding: 40px 20px;
    text-align: center;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 5px 5px var(--main-color);

    color: var(--font-color-sub);

    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
}

.posts-list {
    display: flex;
    flex-direction: column;
    gap: 18px;
}
</style>
