<script setup>
import ChatsSideBar from '@/components/chats/ChatsSideBar.vue';
import ChatWindow from '@/components/chats/ChatWindow.vue';
import SideNavigation from '@/components/layout/SideNavigation.vue';
import TopNavigation from '@/components/layout/TopNavigation.vue';
import { activePage } from '@/data/chatState';
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

activePage.value = 'chat:';

const activeChat = ref(null);

const targetUserId = computed(() => route.query.userId || null);

function handleChatResolved(groupID) {
    if (activeChat.value && !activeChat.value.GroupID) {
        activeChat.value = {
            ...activeChat.value,
            GroupID: groupID
        };
    }
}

function handleSelectChat(chat) {
    activeChat.value = chat;
    activePage.value = 'chat:' + chat.UserID;
}

watch(
    targetUserId,
    (userId) => {
        if (!userId) {
            return;
        }

        activeChat.value = {
            UserID: Number(userId),
            GroupID: null,
            FirstName: route.query.firstName || '',
            LastName: route.query.lastName || '',
            Avatar: route.query.avatar || ''
        };

        activePage.value = 'chat:' + userId;
    },
    { immediate: true }
);
</script>

<template>
    <div class="app-shell">
        <TopNavigation />

        <div class="body-layout">
            <SideNavigation />
            
            <main class="chats-page">
                <ChatsSideBar :target-user-id="targetUserId" @select-chat="handleSelectChat" />

                <ChatWindow
                    :chat="activeChat"
                    :userID="activeChat?.UserID"
                    :groupID="activeChat?.GroupID"
                    :userFirstName="activeChat?.FirstName"
                    :userLastName="activeChat?.LastName"
                    :userAvatar="activeChat?.Avatar"
                    @chat-resolved="handleChatResolved"
                />
            </main>
        </div>
    </div>
</template>

<style scoped>
.app-shell {
    min-height: 100vh;
}

.body-layout {
    display: flex;
    padding-top: 64px;
}

.chats-page {
    flex: 1;
    display: flex;
    gap: 20px;
    padding: 20px clamp(16px, 3vw, 30px);
    box-sizing: border-box;
    min-width: 0;
}

@media (max-width: 800px) {
    .chats-page {
        flex-direction: column;
    }
}
</style>