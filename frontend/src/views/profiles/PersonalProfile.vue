<script setup>
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { getUserData } from '@/api/users/personalProfile';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import ProfileHeader from '@/components/personalProfile/ProfileHeader.vue';
import ProfileTabs from '@/components/personalProfile/ProfileTabs.vue';
import AboutTab from '@/components/profile/AboutTab.vue';
import FollowersTab from '@/components/profile/FollowersTab.vue';
import GroupTab from '@/components/personalProfile/group/GroupTab.vue';
import { addNotification } from '@/data/notifications';
import ProfilePostsTab from '@/components/profile/posts/ProfilePostsTab.vue';
import { activePage } from '@/data/chatState';

const route = useRoute();

const loading = ref(true);
const user = ref(null);
const activeTab = ref('personal');
activePage.value = 'personalProfile';

async function getData() {
    try {
        user.value = await getUserData();
    } catch (err) {
        addNotification('Could not get user data', 'error');
        console.error(err);
    } finally {
        loading.value = false;
    }
}

onMounted(() => {
    getData();
    activeTab.value = ref(
        route.query.tab === 'posts' ? 'posts' : 'personal'
    );

});
</script>

<template>
    <div class="facebook-layout">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="profile-page">
                <BackToHome />

                <div v-if="loading">
                    Loading profile...
                </div>

                <template v-else-if="user">
                    <ProfileHeader :user-id="user.ID" :first-name="user.firstName" :last-name="user.lastName" :username="user.username"
                        :add-edit="true" :bio="user.Profile.About?.bio" :avatar-path="`/uploads/${user.Profile.avatar}`"
                        :num-of-posts="user.Profile.numOfPosts" :num-of-following="user.Profile.numOfFollowing"
                        :num-of-followers="user.Profile.numOfFollowers" />

                    <ProfileTabs type="personal" @change-tab="activeTab = $event" />

                    <AboutTab v-if="activeTab === 'about'" :about="user.Profile.About" :email="user.email" :dob="user.DOB" />

                    <FollowersTab v-if="activeTab === 'followers'" type="followers" :target-id="user.ID"
                        :follower-list="user.Profile.followers" :own-profile="true" />

                    <FollowersTab v-if="activeTab === 'following'" type="following" :target-id="user.ID"
                        :follower-list="user.Profile.following" />

                    <FollowersTab v-if="activeTab === 'friends'" type="friends" :target-id="user.ID"
                        :follower-list="user.Profile.friends" />

                    <ProfilePostsTab v-if="activeTab === 'posts'" :current-user-id="user.ID" />
                    <GroupTab v-if="activeTab === 'groups'" />
                </template>
            </main>
        </div>
    </div>
</template>

<style scoped>
.facebook-layout {
    min-height: 100vh;
    min-height: 100dvh;
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