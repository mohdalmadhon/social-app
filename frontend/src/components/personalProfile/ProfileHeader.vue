```vue
<script setup>
import { ref } from 'vue';
import { requestFollow } from '@/api/users/profiles';
import { useRoute, useRouter } from 'vue-router';
import { addNotification } from '@/data/notifications';

const route = useRoute();
const router = useRouter();

const props = defineProps({
    addEdit: Boolean,
    firstName: String,
    lastName: String,
    username: String,
    bio: String,
    avatarPath: String,
    numOfPosts: Number,
    numOfFollowing: Number,
    numOfFollowers: Number,
    isFollowing: Number
});

const emit = defineEmits([
    'follow',
    'unfollow',
    'cancel-request'
]);

const followingStatus = ref(props.isFollowing);

function handleMessage() {
    const id = route.query.id;

    if (!id) {
        addNotification("could not open chat", 'error');
        return;
    }

    router.push({
        path: '/chats',
        query: {
            userId: id,
            firstName: props.firstName || '',
            lastName: props.lastName || '',
            avatar: (props.avatarPath || '').replace(/^\/uploads\//, '')
        }
    });
}



async function handleFollow() {
    const id = route.query.id;

    try {
        const result = await requestFollow(id, "POST");

        if (result.status) {
            followingStatus.value = result.followStatus;
            emit('follow');
        } else {
            addNotification("could not follow user", 'error');
        }
    } catch (err) {
        addNotification("could not follow user", 'error');
        console.error(err);
    }
}

async function handleRemoveFollow() {
    const id = route.query.id;
    const oldStatus = followingStatus.value;

    try {
        const result = await requestFollow(id, "DELETE");

        if (result.status) {
            followingStatus.value = result.followStatus;

            if (oldStatus === 0) {
                emit('cancel-request');
            } else if (oldStatus === 1) {
                emit('unfollow');
            }
        }
    } catch (err) {
        addNotification("could not unfollow user", 'error');
        console.error(err);
    }
}
</script>

<template>
    <section class="profile-header">
        <div class="cover">
            <div class="cover-grid"></div>
        </div>

        <div class="profile-information">
            <div class="avatar">
                <img
                    v-if="props.avatarPath"
                    :src="props.avatarPath"
                    alt="Profile avatar"
                >
            </div>

            <div class="profile-details">
                <div class="name-row">
                    <div>
                        <h1>
                            {{ props.firstName }} {{ props.lastName }}
                        </h1>

                        <p class="username">
                            {{ props.username || '' }}
                        </p>
                    </div>

                    <div class="profile-actions">
                        <a
                            v-if="props.addEdit"
                            href="/me/edit"
                            class="edit-button"
                        >
                            Edit profile
                        </a>

                        <template v-else>
                            <button
                                class="relationship-button message"
                                @click="handleMessage"
                            >
                                Message
                            </button>

                            <button
                                v-if="followingStatus === -1"
                                class="relationship-button follow"
                                @click="handleFollow"
                            >
                                Follow
                            </button>

                            <button
                                v-else-if="followingStatus === 0"
                                class="relationship-button requested"
                                @click="handleRemoveFollow"
                            >
                                Requested
                            </button>

                            <button
                                v-else-if="followingStatus === 1"
                                class="relationship-button following"
                                @click="handleRemoveFollow"
                            >
                                Following
                            </button>
                        </template>
                    </div>
                </div>

                <p class="about">
                    {{ props.bio }}
                </p>

                <div class="profile-stats">
                    <span>
                        <strong>{{ props.numOfPosts }}</strong> Posts
                    </span>

                    <span>
                        <strong>{{ props.numOfFollowing }}</strong> Following
                    </span>

                    <span>
                        <strong>{{ props.numOfFollowers }}</strong> Followers
                    </span>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.profile-header {
    overflow: hidden;
    border: 2px solid var(--main-color);
    border-radius: 8px;
    background: var(--bg-color);
    box-shadow: 7px 7px var(--main-color);
}

.cover {
    position: relative;
    height: 150px;
    z-index: -1;
    overflow: hidden;
    background: var(--input-focus);
    border-bottom: 2px solid var(--main-color);
}

.cover::before,
.cover::after {
    position: absolute;
    content: "";
    border: 2px solid var(--main-color);
    background: var(--bg-color);
    box-shadow: 6px 6px var(--main-color);
}

.cover::before {
    width: 150px;
    height: 150px;
    right: 100px;
    top: -70px;
    transform: rotate(25deg);
}

.cover::after {
    width: 80px;
    height: 80px;
    left: 120px;
    bottom: -40px;
    transform: rotate(45deg);
}

.cover-grid {
    position: absolute;
    inset: 0;
    opacity: 0.15;
    background-image:
        linear-gradient(var(--main-color) 1px, transparent 1px),
        linear-gradient(90deg, var(--main-color) 1px, transparent 1px);
    background-size: 25px 25px;
}

.profile-information {
    display: flex;
    gap: 28px;
    padding: 0 35px 30px;
}

.avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
}

.avatar {
    flex-shrink: 0;
    width: 150px;
    height: 150px;
    margin-top: -75px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 4px solid var(--bg-color);
    outline: 2px solid var(--main-color);
    border-radius: 50%;
    background: var(--main-color);
    color: white;
    font-family: "Liter", serif;
    font-size: 65px;
    box-shadow: 5px 5px var(--main-color);
}

.profile-details {
    width: 100%;
    padding-top: 22px;
}

.name-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 20px;
}

.profile-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
}

h1 {
    margin: 0;
    color: var(--main-color);
    font-family: "Liter", serif;
    font-size: 36px;
    line-height: 1;
}

.username {
    margin: 7px 0 0;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.edit-button,
.relationship-button {
    flex-shrink: 0;
    padding: 11px 18px;
    border: 2px solid var(--main-color);
    border-radius: 5px;
    background: var(--input-focus);
    box-shadow: 4px 4px var(--main-color);
    color: white;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.1s, box-shadow 0.1s, background 0.15s;
}

.edit-button:hover,
.relationship-button:hover {
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px var(--main-color);
}

.relationship-button.follow {
    background: var(--input-focus);
    color: white;
}

.relationship-button.requested {
    background: var(--bg-color);
    color: var(--main-color);
}

.relationship-button.following {
    background: var(--main-color);
    color: var(--bg-color);
}

.relationship-button.message {
    background: var(--bg-color);
    color: var(--main-color);
}

.relationship-button:active {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px var(--main-color);
}

.about {
    max-width: 650px;
    margin: 18px 0;
    color: var(--font-color-sub);
    font-size: 14px;
    line-height: 1.6;
}

.profile-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 22px;
    color: var(--font-color-sub);
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
}

.profile-stats strong {
    color: var(--main-color);
    font-size: 12px;
}

@media (max-width: 650px) {
    .cover {
        height: 150px;
    }

    .profile-information {
        display: block;
        padding: 0 20px 25px;
    }

    .avatar {
        width: 115px;
        height: 115px;
        margin-top: -58px;
        font-size: 48px;
    }

    .profile-details {
        padding-top: 20px;
    }

    h1 {
        font-size: 29px;
    }

    .name-row {
        align-items: center;
    }

    .profile-actions {
        gap: 8px;
    }

    .profile-actions .relationship-button,
    .profile-actions .edit-button {
        padding: 9px 12px;
    }

    .profile-stats {
        gap: 12px;
    }
}
</style>
