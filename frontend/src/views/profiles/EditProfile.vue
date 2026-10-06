<script setup>
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import BackToHome from '@/components/layout/BackToHome.vue';
import EditProfileTabs from '@/components/ProfileEdit/EditProfileTabs.vue';
import EditPersonalInfo from '@/components/ProfileEdit/EditPersonalInfo.vue';
import EditAdditionalInfo from '@/components/ProfileEdit/EditAdditionalInfo.vue';

import { onMounted, ref } from 'vue';
import { getUserData } from '@/api/users/personalProfile';
import { addNotification } from '@/data/notifications';
import { activePage } from '@/data/chatState';

const activeTab = ref('personal');
const loading = ref(true);
const user = ref(null);

activePage.value = 'editProfile';

async function getData() {
    try {
        user.value = await getUserData();
    } catch (err) {
        addNotification('could not get user data', 'error')
        console.error(err);
    } finally {
        loading.value = false;
    }
}

onMounted(getData);

</script>

<template>
    <div class="facebook-layout">
        <TopNavigation />

        <div class="page-layout">
            <SideNavigation />

            <main class="profile-page">
                <BackToHome />

                <div class="page-heading">
                    <p class="eyebrow">SETTINGS</p>
                    <h1>Edit Profile</h1>
                </div>

                <EditProfileTabs v-model:activeTab="activeTab" />

                <section class="profile-content">
                    <template v-if="!loading && user">
                        <EditPersonalInfo v-if="activeTab === 'personal'" :first-name="user.firstName"
                            :last-name="user.lastName" :username="user.username" :email="user.email"
                            :bio="user.Profile.About?.bio" :avatar_path="`/uploads/${user.Profile.avatar}`"
                            :is-private="user.isPrivate" />

                        <EditAdditionalInfo v-else-if="activeTab === 'additional'" :about="user.Profile.About" />
                    </template>
                </section>
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

.page-heading .eyebrow {
    margin: 0 0 5px;
    color: var(--input-focus);
    font-family: "JetBrains Mono", monospace;
    font-size: 9px;
    letter-spacing: 2px;
}

.page-heading h1 {
    margin: 0;
    font-family: "Liter", serif;
    font-size: 36px;
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