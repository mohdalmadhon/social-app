<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';

import HomePostHeader from './HomePostHeader.vue';
import HomePostImage from './HomePostImage.vue';
import HomePostReaction from './HomePostReaction.vue';
import HomePostAction from './HomePostAction.vue';
import HomePostComments from './HomePostComments.vue';
import LocationDialouge from './LocationDialouge.vue';
import TaggedPeopleDialoug from './TaggedPeopleDialoug.vue';
import { viewPost } from '@/api/posts/posts.js';

const props = defineProps({
    currentUserId: {
        type: [Number, String],
        default: null
    },
    allowComments: {
        type: Boolean
    },
    reaction: {
        type: Number,
        required: true
    },
    userId: {
        type: [Number, String],
        default: null
    },
    postId: {
        type: Number,
        required: true
    },
    groupId: {
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
    username: {
        type: String,
        default: ''
    },
    avatarPath: {
        type: String,
        default: ''
    },
    createdAt: {
        type: String,
        default: ''
    },
    content: {
        type: String,
        default: ''
    },
    imagePath: {
        type: String,
        default: ''
    },
    location: {
        type: String,
        default: ''
    },
    taggedPeople: {
        type: Array,
        default: () => []
    },
    likes: {
        type: Number,
        default: 0
    },
    dislikes: {
        type: Number,
        default: 0
    },
    comments: {
        type: Array,
        default: () => []
    },
    userReaction: {
        type: String,
        default: ''
    },
    relationship: {
        type: String,
        default: 'none'
    },
    visibility: {
        type: String,
        default: 'public'
    },
    visibilityUser: {
        type: String,
        default: ''
    }
});

const emit = defineEmits([
    'like',
    'dislike',
    'comment',
    'open-comments',
    'open-tags'
]);

const postElement = ref(null);

const showComments = ref(false);
const showLocationDialog = ref(false);
const showTaggedDialog = ref(false);

let observer;
let seen = false;

const locationParts = computed(() => {
    if (!props.location) {
        return null;
    }
    
    const parts = props.location.split(':').map(p => p.trim());

    if (parts.length < 3) {
        return {
            display: props.location,
            lat: null,
            lon: null
        };
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
    if (
        locationParts.value?.lat == null ||
        locationParts.value?.lon == null
    ) {
        return '';
    }

    return `https://www.google.com/maps?q=${locationParts.value.lat},${locationParts.value.lon}&output=embed`;
});

const mapExternalUrl = computed(() => {
    if (
        locationParts.value?.lat == null ||
        locationParts.value?.lon == null
    ) {
        return '';
    }

    return `https://www.google.com/maps?q=${locationParts.value.lat},${locationParts.value.lon}`;
});

const formattedDate = computed(() => {
    return formatRelativeDate(props.createdAt);
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

function handleLike() {
    emit('like');
}

function handleDislike() {
    emit('dislike');
}

function toggleComments() {
    showComments.value = !showComments.value;
    emit('open-comments');
}

function handleSubmitComment(text) {
    emit('comment', text);
}

function openTaggedPeople() {
    if (!props.taggedPeople.length) {
        return;
    }

    showTaggedDialog.value = true;
    emit('open-tags', props.taggedPeople);
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

async function markPostAsSeen() {
    if (seen) {
        return;
    }

    seen = true;

    try {
        await viewPost(props.postId);
    } catch (err) {
        console.error('Could not mark post as seen:', err);
    }
}

onMounted(() => {
    if (!postElement.value) {
        return;
    }

    if (!('IntersectionObserver' in window)) {
        markPostAsSeen();
        return;
    }

    observer = new IntersectionObserver(
        entries => {
            const entry = entries[0];

            if (entry.isIntersecting) {
                observer.disconnect();
                markPostAsSeen();
            }
        },
        {
            threshold: 0.5
        }
    );

    observer.observe(postElement.value);
});

onBeforeUnmount(() => {
    observer?.disconnect();
});
</script>

<template>
    <article ref="postElement" class="post-card">
        <HomePostHeader
            :user-id="userId"
            :group-id="groupId"
            :first-name="firstName"
            :last-name="lastName"
            :avatar-path="avatarPath"
            :relationship="relationship"
            :tagged-people="taggedPeople"
            :formatted-date="formattedDate"
            :location-display="locationParts?.display"
            :has-location="!!location"
            @open-tags="openTaggedPeople"
            @open-location="openLocationDialog"
        />

        <div v-if="content" class="post-content">
            {{ content }}
        </div>

        <HomePostImage
            :image-path="imagePath"
            :tagged-people="taggedPeople"
            @open-tags="openTaggedPeople"
        />

        <HomePostReaction
            :likes="likes"
            :dislikes="dislikes"
            :comments-count="comments.length"
            @toggle-comments="toggleComments"
        />

        <HomePostAction
            :reaction="props.reaction"
            :user-reaction="userReaction"
            :post-id="postId"
            :likes="props.likes"
            :allowComments="props.allowComments"
            :dislikes="props.dislikes"
            @like="handleLike"
            @dislike="handleDislike"
            @toggle-comments="toggleComments"
        />

        <HomePostComments
            :show="showComments"
            :post-id="postId"
            :current-user-id="currentUserId"
            :first-name="firstName"
            :last-name="lastName"
            :avatar-path="avatarPath"
            :created-at="createdAt"
            :content="content"
            :image-path="imagePath"
            @close="showComments = false"
        />

        <LocationDialouge
            :show="showLocationDialog"
            :display="locationParts?.display"
            :embed-url="mapEmbedUrl"
            :external-url="mapExternalUrl"
            @close="closeLocationDialog"
        />

        <TaggedPeopleDialoug
            :show="showTaggedDialog"
            :people="taggedPeople"
            @close="closeTaggedDialog"
        />
    </article>
</template>

<style scoped>
.post-card {
    width: 100%;
    max-width: 720px;
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.post-content {
    padding: 0 20px 18px;
    color: var(--font-color);
    font-size: 14px;
    line-height: 1.55;
    white-space: pre-wrap;
    word-break: break-word;
}

@media (max-width: 650px) {
    .post-card {
        box-shadow: 4px 4px var(--main-color);
    }

    .post-content {
        padding: 0 14px 15px;
        font-size: 13px;
    }
}
</style>