<script setup>
import { onMounted, ref } from 'vue';

import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import ProfileHeader from '@/components/personalProfile/ProfileHeader.vue';
import ProfileTabs from '@/components/personalProfile/ProfileTabs.vue';
import PrivateProfileIcon from '@/components/ProfileEdit/PrivateProfileIcon.vue';
import AboutTab from '@/components/profile/AboutTab.vue';

import { getProfileData } from '@/api/users/profiles';

import { useRoute } from 'vue-router';
import FollowersTab from '@/components/profile/FollowersTab.vue';
import { addNotification } from '@/data/notifications';
import { getFriends } from '@/api/common/friends';
import ProfilePostsTab from '@/components/profile/posts/ProfilePostsTab.vue';
import { activePage } from '@/data/chatState';

const route = useRoute();
const userID = route.query.id;
activePage.value = `profile:${userID}`;

const activeTab = ref('about');
const loading = ref(true);
const showPrivateProfile = ref(false);
const user = ref(null);
const canMessage = ref(false); 

async function getData() {
    const id = route.query.id;
    const count = 10;
    try {
        user.value = await getProfileData(id, count);
        showPrivateProfile.value = !user.value.show;
        const result = await getFriends("", id)
        user.value.Profile.friends = result.data;

        canMessage.value = user.value.canMessage;
        
    } catch (err) {
        addNotification('could not get user data', 'error')
        console.error(err);
    } finally {
        loading.value = false;
    }
}

function handleUnfollow() {
    if (user.value.isPrivate === 1) {
        showPrivateProfile.value = true;
    }
}

function handleFollow() {
    showPrivateProfile.value = false;
}

let id = route.query.id;
if (!id) {
    id = ""
}
onMounted(getData);
</script>

<template>
    <div class="facebook-layout">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="profile-page">
                <div v-if="loading">
                    Loading profile...
                </div>

                <template v-else-if="user">
                    <ProfileHeader :first-name="user.firstName"
                        :message="canMessage"
                        :last-name="user.lastName" :username="user.username"
                        :bio="user.Profile.About?.bio" :avatar-path="`/uploads/${user.Profile.avatar}`"
                        :num-of-posts="user.Profile.numOfPosts" :num-of-following="user.Profile.numOfFollowing"
                        :num-of-followers="user.Profile.numOfFollowers" :add-edit="false"
                        :is-following="user.isFollowing" @unfollow="handleUnfollow" @follow="handleFollow" />

                    <template v-if="!showPrivateProfile">
                        <section class="profile-content">
                            <ProfileTabs v-if="user.show" type="user" @change-tab="activeTab = $event" />

                            <AboutTab v-if="user.show && activeTab === 'about'" :about="user.Profile.About" />

                            <FollowersTab :target-id="id" v-if="user.show && activeTab === 'followers'"
                                :follower-list="user.Profile.followers" />

                            <FollowersTab :target-id="id" v-if="user.show && activeTab === 'following'"
                                :follower-list="user.Profile.following" />

                            <FollowersTab :target-id="id" v-if="user.show && activeTab === 'friends'"
                                :follower-list="user.Profile.friends" />

                            <ProfilePostsTab v-if="activeTab === 'posts' && user.show" :user-id="id" />
                            <PrivateProfileIcon v-else-if="!user.show" />
                        </section>
                    </template>

                    <section v-else class="profile-content">
                        <PrivateProfileIcon />
                    </section>
                </template>
            </main>
        </div>
    </div>
</template>

<style scoped>
.facebook-layout {
    min-height: 100vh;
}

.page-layout {
    display: flex;
    padding-top: 64px;
}

.profile-page {
    width: 100%;
    max-width: 1100px;
    margin: 0 auto;
    padding: 25px 30px 60px;
}

.profile-content {
    width: 100%;
}

@media (max-width: 800px) {
    .page-layout {
        display: block;
    }

    .profile-page {
        padding: 20px 15px 50px;
    }
}
</style>