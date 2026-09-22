<script setup>
import { ref, computed } from 'vue';
import ProfilePostGroupBadge from './Profilepostgroupbadge.vue';
import HomePostHeader from '@/components/home/HomePostHeader.vue';
import HomePostImage from '@/components/home/HomePostImage.vue';
import HomePostReaction from '@/components/home/HomePostReaction.vue';
import HomePostAction from '@/components/home/HomePostAction.vue';
import HomePostComments from '@/components/home/HomePostComments.vue';
import LocationDialouge from '@/components/home/LocationDialouge.vue';
import TaggedPeopleDialoug from '@/components/home/TaggedPeopleDialoug.vue';
import ConfirmModal from '@/components/personalProfile/group/ConfirmModal.vue';
import { deletePost } from '@/api/posts/posts';
import { addNotification } from '@/data/notifications';

const props = defineProps({
    post: {
        type: Object,
        required: true
    },

    currentUserId: {
        type: [Number, String],
        default: null
    }
});

const emit = defineEmits(['deleted']);

const showComments = ref(false);
const showLocationDialog = ref(false);
const showTaggedDialog = ref(false);
const showDeleteConfirm = ref(false);
const deleting = ref(false);

const canDelete = computed(() => {
    return (
        props.currentUserId !== null &&
        props.currentUserId !== undefined &&
        String(props.post.userId) === String(props.currentUserId)
    );
});

const locationParts = computed(() => {
    if (!props.post.location) {
        return null;
    }

    const parts = props.post.location.split(':').map(p => p.trim());

    if (parts.length < 3) {
        return { display: props.post.location, lat: null, lon: null };
    }

    const lat = parseFloat(parts[1]);
    const lon = parseFloat(parts[2]);

    return {
        display: parts[0],
        lat: Number.isNaN(lat) ? null : lat,
        lon: Number.isNaN(lon) ? null : lon
    };
});

const mapEmbedUrl = computed(() => {
    if (!locationParts.value?.lat || !locationParts.value?.lon) {
        return '';
    }

    return `https://www.google.com/maps?q=${locationParts.value.lat},${locationParts.value.lon}&output=embed`;
});

const mapExternalUrl = computed(() => {
    if (!locationParts.value?.lat || !locationParts.value?.lon) {
        return '';
    }

    return `https://www.google.com/maps?q=${locationParts.value.lat},${locationParts.value.lon}`;
});

function formatRelativeDate(dateString) {
    if (!dateString) {
        return '';
    }

    const created = new Date(dateString);

    if (Number.isNaN(created.getTime())) {
        return dateString;
    }

    const diffMs = Date.now() - created.getTime();
    const diffHours = diffMs / (1000 * 60 * 60);
    const diffDays = diffHours / 24;

    if (diffHours < 24) {
        const hours = Math.max(1, Math.floor(diffHours));
        return `${hours}h ago`;
    }

    if (diffDays <= 6) {
        const days = Math.floor(diffDays);
        return `${days}d ago`;
    }

    const weeks = Math.floor(diffDays / 7);
    return `${weeks}w ago`;
}

const formattedDate = computed(() => formatRelativeDate(props.post.createdAt));

function toggleComments() {
    showComments.value = !showComments.value;
}

function openTaggedPeople() {
    if (!props.post.taggedPeople?.length) {
        return;
    }

    showTaggedDialog.value = true;
}

function closeTaggedDialog() {
    showTaggedDialog.value = false;
}

function openLocationDialog() {
    if (mapEmbedUrl.value) {
        showLocationDialog.value = true;
    }
}

function closeLocationDialog() {
    showLocationDialog.value = false;
}

function openDeleteConfirm() {
    showDeleteConfirm.value = true;
}

function closeDeleteConfirm() {
    showDeleteConfirm.value = false;
}

async function confirmDelete() {
    if (deleting.value) {
        return;
    }

    deleting.value = true;

    try {
        const result = await deletePost(props.post.id);

        if (!result || !result.status) {
            throw new Error(result?.message || 'Could not delete post');
        }

        addNotification('Post deleted', 'success');
        showDeleteConfirm.value = false;
        emit('deleted', props.post.id);
    } catch (err) {
        console.error(err);
        addNotification(err.message || 'Could not delete post', 'error');
    } finally {
        deleting.value = false;
    }
}
</script>

<template>
    <article class="profile-post-card">
        <ProfilePostGroupBadge :group-id="post.groupId" :group-name="post.groupName" />

        <button
            v-if="canDelete"
            type="button"
            class="delete-post-btn"
            title="Delete post"
            @click="openDeleteConfirm"
        >
            🗑
        </button>

        <HomePostHeader :user-id="post.userId" :group-id="post.groupId" :first-name="post.firstName"
            :last-name="post.lastName" :avatar-path="post.avatarPath" :relationship="post.relationship"
            :tagged-people="post.taggedPeople" :formatted-date="formattedDate"
            :location-display="locationParts?.display" :has-location="!!post.location"
            @open-tags="openTaggedPeople" @open-location="openLocationDialog" />

        <div v-if="post.content" class="post-content">
            {{ post.content }}
        </div>

        <HomePostImage :image-path="post.imagePath" :tagged-people="post.taggedPeople"
            @open-tags="openTaggedPeople" />

        <HomePostReaction :likes="post.likeCount" :dislikes="post.dislikeCount" :comments-count="post.commentCount"
            @toggle-comments="toggleComments" />

        <HomePostAction :reaction="post.reactionValue" :post-id="post.id" :likes="post.likeCount"
            :dislikes="post.dislikeCount" @toggle-comments="toggleComments" />

        <HomePostComments :show="showComments" :post-id="post.id" :current-user-id="currentUserId"
            :first-name="post.firstName" :last-name="post.lastName" :avatar-path="post.avatarPath"
            :created-at="post.createdAt" :content="post.content" :image-path="post.imagePath"
            @close="showComments = false" />

        <LocationDialouge :show="showLocationDialog" :display="locationParts?.display" :embed-url="mapEmbedUrl"
            :external-url="mapExternalUrl" @close="closeLocationDialog" />

        <TaggedPeopleDialoug :show="showTaggedDialog" :people="post.taggedPeople" @close="closeTaggedDialog" />

        <ConfirmModal
            v-if="showDeleteConfirm"
            title="Delete this post?"
            confirm-label="Delete"
            cancel-label="Cancel"
            :danger="true"
            @confirm="confirmDelete"
            @cancel="closeDeleteConfirm"
        >
            <p>This will permanently delete the post, its comments, and reactions. This cannot be undone.</p>
        </ConfirmModal>
    </article>
</template>

<style scoped>
.profile-post-card {
    position: relative;
    display: flex;
    flex-direction: column;

    width: 100%;
    overflow: hidden;

    border: 2px solid var(--main-color);
    border-radius: 8px;

    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.delete-post-btn {
    position: absolute;
    top: 10px;
    right: 10px;
    z-index: 2;

    display: flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;

    border: 2px solid var(--main-color);
    border-radius: 50%;

    background: var(--bg-color);
    color: var(--font-color);

    font-size: 14px;
    line-height: 1;
    cursor: pointer;

    box-shadow: 3px 3px var(--main-color);
    transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
}

.delete-post-btn:hover {
    border-color: #c0392b;
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px var(--main-color);
}

.post-content {
    padding: 14px 20px 18px;

    color: var(--font-color);

    font-size: 14px;
    line-height: 1.55;

    white-space: pre-wrap;
    word-break: break-word;
}

@media (max-width: 650px) {
    .profile-post-card {
        box-shadow: 4px 4px var(--main-color);
    }

    .post-content {
        padding: 10px 14px 15px;

        font-size: 13px;
    }
}
</style>