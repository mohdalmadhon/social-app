<script setup>
import ProfilePostCard from './ProfilePostCard.vue';

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

const emit = defineEmits(['deleted']);
</script>

<template>
    <div class="profile-post-grid">
        <ProfilePostCard v-for="post in posts" :key="post.id" :post="post" :current-user-id="currentUserId"
            @deleted="emit('deleted', $event)" />
    </div>
</template>

<style scoped>
.profile-post-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 22px;

    width: 100%;
}

@media (max-width: 1000px) {
    .profile-post-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 650px) {
    .profile-post-grid {
        grid-template-columns: 1fr;
        gap: 16px;
    }
}
</style>